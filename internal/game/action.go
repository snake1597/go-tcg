package game

import (
	"fmt"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
	"sort"
)

const (
	blazingThrowCardID      CardID = "iohZMWh5v5"
	fieryInterferenceCardID CardID = "gt2zqtgs42"
	straightFlareCardID     CardID = "28bjn8g50v"
)

type declarationStage string

const (
	declarationTarget declarationStage = "target"
	declarationWeapon declarationStage = "weapon"
)

type actionDeclaration struct {
	Controller *model.Player    `json:"controller"`
	Source     cardInstanceID   `json:"source"`
	Target     objectID         `json:"target,omitempty"`
	Stage      declarationStage `json:"stage"`
}

// veritaAlternativeCostDeclaration 保存尚未提交的墓地付款選擇。
// Source 是手牌中的 Verita，Selected 是玩家逐張選取的墓地牌；建立與取消都不移動任何牌。
type veritaAlternativeCostDeclaration struct {
	Controller *model.Player    `json:"controller"`
	Source     cardInstanceID   `json:"source"`
	Selected   []cardInstanceID `json:"selected"`
}

func (g *Game) legalActionCards(player *model.Player) []cardInstanceID {
	if g.state.Knowledge.Declaration != nil {
		return nil
	}
	zones := g.state.Zones[player.UID]
	cards := make([]cardInstanceID, 0, len(zones.Hand))
	for _, card := range zones.Hand {
		if g.canActivateAction(player, card) {
			cards = append(cards, card)
		}
	}
	return cards
}

// canActivateAction 檢查持有者、行動機會、費用與目前已實作的卡牌限制。
// 非 Fast 行動只允許在自己的主階段且效果堆疊為空時使用；Verita 可採替代費用。
// 此檢查不保證來源位於手牌，也不驗證最終目標；宣告與提交階段仍須檢查。
func (g *Game) canActivateAction(player *model.Player, card cardInstanceID) bool {
	scheduler := g.state.Scheduler
	candidate, exists := g.state.Cards[card]
	if !exists || !samePlayer(candidate.Owner, player) || (!containsString(candidate.Types, "ACTION") && candidate.Definition != veritaCardID) {
		return false
	}
	canUseAlternativeCost := candidate.Definition == veritaCardID && len(g.veritaAlternativeCostCards(player)) > 0
	if !samePlayer(scheduler.OpportunityHolder, player) || (!canUseAlternativeCost && len(g.state.Zones[player.UID].Memory) < g.actionReserveCost(player, card)) {
		return false
	}
	if !candidate.Fast && (!samePlayer(scheduler.TurnPlayer, player) || scheduler.Phase != PhaseMain || len(g.state.EffectsStack) != 0) {
		return false
	}
	if candidate.Definition == blazingThrowCardID && len(g.legalWeapons(player)) == 0 {
		return false
	}
	if candidate.Definition == veritaCardID {
		return true
	}
	if candidate.Definition == trumpSetCardID {
		return len(g.trumpSetTargets(player)) > 0
	}
	return candidate.Definition == blazingThrowCardID || candidate.Definition == fieryInterferenceCardID || candidate.Definition == straightFlareCardID
}

// beginActionDeclaration 為行動建立選目標的宣告，尚不移走來源牌或支付費用。
// Blazing Throw 選完目標後還須選擇犧牲武器；Trump Set 僅能選受控的 Suited ally。
// Verita 會先建立可取消的替代費用選擇；其他 action 才建立目標宣告。
func (g *Game) beginActionDeclaration(player *model.Player, card cardInstanceID) error {
	if g.state.Cards[card].Definition == veritaCardID {
		return g.beginVeritaAlternativeCostDeclaration(player, card)
	}
	targets := g.legalTargets()
	if g.state.Cards[card].Definition == trumpSetCardID {
		targets = g.trumpSetTargets(player)
	}
	if !g.canActivateAction(player, card) || len(targets) == 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	g.state.Knowledge.Declaration = &actionDeclaration{
		Controller: player,
		Source:     card,
		Stage:      declarationTarget,
	}
	g.setDeclarationChoice(player, targets)
	return nil
}

func (g *Game) commitAllyActivation(player *model.Player, card cardInstanceID) error {
	if !g.canActivateAction(player, card) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	if g.state.Cards[card].Definition == veritaCardID {
		if len(g.veritaAlternativeCostCards(player)) > 0 {
			return g.beginVeritaAlternativeCostDeclaration(player, card)
		}
	}
	zones := g.state.Zones[player.UID]
	index := cardIndex(zones.Hand, card)
	if index < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	zones.Hand = removeCardAt(zones.Hand, index)
	for payment := 0; payment < g.actionReserveCost(player, card); payment++ {
		memoryIndex := int(g.nextRandom() % uint64(len(zones.Memory)))
		zones.Banishment = append(zones.Banishment, zones.Memory[memoryIndex])
		zones.Memory = removeCardAt(zones.Memory, memoryIndex)
	}
	g.state.Zones[player.UID] = zones
	g.putAllyOnField(player, card)
	return nil
}

// beginVeritaAlternativeCostDeclaration 開始 Verita 的逐張替代費用選擇。
// 輸入為控制者與手牌中的 Verita；輸出為待選 handle，副作用僅建立可取消的宣告狀態。
func (g *Game) beginVeritaAlternativeCostDeclaration(player *model.Player, card cardInstanceID) error {
	if !g.canActivateAction(player, card) || cardIndex(g.state.Zones[player.UID].Hand, card) < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	if len(g.veritaAlternativeCostCards(player)) == 0 {
		return g.commitAllyActivation(player, card)
	}
	declaration := &veritaAlternativeCostDeclaration{
		Controller: player,
		Source:     card,
	}
	if len(g.veritaAlternativeCostChoiceCards(declaration)) == 0 {
		return fmt.Errorf("invalid Verita alternative cost")
	}
	g.state.Knowledge.VeritaCost = declaration
	g.setVeritaAlternativeCostChoice(declaration)
	return nil
}

// submitVeritaAlternativeCostChoice 接受一張仍可完成精確總和的墓地牌。
// 當選到至少三張且總和十時，原子地放逐付款並部署 Verita；否則只更新宣告，無區域副作用。
func (g *Game) submitVeritaAlternativeCostChoice(player *model.Player, card cardInstanceID) error {
	declaration := g.state.Knowledge.VeritaCost
	if declaration == nil || !samePlayer(declaration.Controller, player) || !containsCard(g.veritaAlternativeCostChoiceCards(declaration), card) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	declaration.Selected = append(declaration.Selected, card)
	if g.canUseVeritaAlternativeCost(player, declaration.Selected) {
		if err := g.commitVeritaAlternativeCost(declaration); err != nil {
			return err
		}
		g.state.Knowledge.VeritaCost = nil
		g.state.Knowledge.Choice = nil
		return nil
	}
	g.setVeritaAlternativeCostChoice(declaration)
	return nil
}

// setVeritaAlternativeCostChoice 以仍存在完整付款組合的墓地牌重建選擇 handle。
// 輸入為暫存宣告；輸出寫入 pending choice，副作用不移動卡牌且允許玩家取消。
func (g *Game) setVeritaAlternativeCostChoice(declaration *veritaAlternativeCostDeclaration) {
	options := make([]objectID, 0)
	for _, card := range g.veritaAlternativeCostChoiceCards(declaration) {
		options = append(options, objectID(card))
	}
	g.setDeclarationChoice(declaration.Controller, options)
	g.state.Knowledge.Choice.CanPass = true
}

// veritaAlternativeCostChoiceCards 回傳下一張可選且保證仍有精確付款組合的墓地牌。
// 輸入為暫存選擇；輸出依墓地順序排列，副作用為零。
func (g *Game) veritaAlternativeCostChoiceCards(declaration *veritaAlternativeCostDeclaration) []cardInstanceID {
	candidates := []cardInstanceID{}
	for _, card := range g.state.Zones[declaration.Controller.UID].Graveyard {
		if containsCard(declaration.Selected, card) {
			continue
		}
		selected := append(append([]cardInstanceID(nil), declaration.Selected...), card)
		if len(g.findVeritaAlternativeCostCards(declaration.Controller, g.state.Zones[declaration.Controller.UID].Graveyard, selected, 0)) > 0 {
			candidates = append(candidates, card)
		}
	}
	return candidates
}

// commitVeritaAlternativeCost 驗證來源與完整付款後一次提交所有區域異動。
// 輸入為已完成的宣告；成功時將付款放逐並部署 Verita，失敗時完全不改變遊戲狀態。
func (g *Game) commitVeritaAlternativeCost(declaration *veritaAlternativeCostDeclaration) error {
	if !g.canUseVeritaAlternativeCost(declaration.Controller, declaration.Selected) {
		return fmt.Errorf("invalid Verita alternative cost")
	}
	zones := g.state.Zones[declaration.Controller.UID]
	index := cardIndex(zones.Hand, declaration.Source)
	if index < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, declaration.Source)
	}
	for _, card := range declaration.Selected {
		graveyardIndex := cardIndex(zones.Graveyard, card)
		if graveyardIndex < 0 {
			return fmt.Errorf("invalid Verita alternative cost")
		}
		zones.Graveyard = removeCardAt(zones.Graveyard, graveyardIndex)
		zones.Banishment = append(zones.Banishment, card)
	}
	zones.Hand = removeCardAt(zones.Hand, index)
	g.state.Zones[declaration.Controller.UID] = zones
	g.putAllyOnField(declaration.Controller, declaration.Source)
	return nil
}

func (g *Game) submitActionDeclarationChoice(player *model.Player, subject entityID) error {
	declaration := g.state.Knowledge.Declaration
	if declaration == nil || !samePlayer(declaration.Controller, player) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
	}
	switch declaration.Stage {
	case declarationTarget:
		target := objectID(subject)
		if !g.isLegalTarget(target) || (g.state.Cards[declaration.Source].Definition == trumpSetCardID && !containsObject(g.trumpSetTargets(player), target)) {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
		}
		declaration.Target = target
		if g.state.Cards[declaration.Source].Definition == blazingThrowCardID {
			declaration.Stage = declarationWeapon
			weapons := g.legalWeapons(player)
			g.setDeclarationChoice(player, weapons)
			return nil
		}
		return g.commitActionDeclaration()
	case declarationWeapon:
		weapon := objectID(subject)
		if !g.isLegalWeapon(player, weapon) {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
		}
		return g.commitActionDeclarationWithWeapon(weapon)
	default:
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
	}
}

func (g *Game) setDeclarationChoice(player *model.Player, subjects []objectID) {
	options := make(map[ViewHandle]entityID, len(subjects))
	for _, subject := range subjects {
		handle := g.newViewHandle(player, "choice:"+string(subject))
		options[handle] = entityID(subject)
	}
	g.state.Knowledge.Choice = &pendingChoice{
		Actor:   player,
		Options: options,
	}
}

// commitActionDeclarationWithWeapon 在重新驗證武器與宣告後，先犧牲武器，再提交行動。
// 武器進入擁有者墓地；此函式沒有在後續提交失敗時還原武器的機制。
func (g *Game) commitActionDeclarationWithWeapon(weapon objectID) error {
	declaration := g.state.Knowledge.Declaration
	if declaration == nil || !g.isLegalWeapon(declaration.Controller, weapon) || !g.canCommitActionDeclaration(declaration) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, weapon)
	}
	object := g.state.Objects[weapon]
	delete(g.state.Objects, weapon)
	zones := g.state.Zones[object.Owner.UID]
	zones.Graveyard = append(zones.Graveyard, object.Card)
	g.state.Zones[object.Owner.UID] = zones
	return g.commitActionDeclaration()
}

// commitActionDeclaration 在目標、來源手牌與費用仍有效時，移走來源並隨機放逐 Memory 付款。
// 付款後將能力入堆疊、清除宣告與待選項目，並授予控制者行動機會；能力尚未結算。
func (g *Game) commitActionDeclaration() error {
	declaration := g.state.Knowledge.Declaration
	if declaration == nil || !g.canCommitActionDeclaration(declaration) {
		return fmt.Errorf("%w", tcgErrors.ErrInvalidViewHandle)
	}
	zones := g.state.Zones[declaration.Controller.UID]
	sourceIndex := cardIndex(zones.Hand, declaration.Source)
	if sourceIndex < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, declaration.Source)
	}
	zones.Hand = removeCardAt(zones.Hand, sourceIndex)
	for paymentCount := 0; paymentCount < g.actionReserveCost(declaration.Controller, declaration.Source); paymentCount++ {
		memoryIndex := int(g.nextRandom() % uint64(len(zones.Memory)))
		payment := zones.Memory[memoryIndex]
		zones.Memory = removeCardAt(zones.Memory, memoryIndex)
		zones.Banishment = append(zones.Banishment, payment)
	}
	g.state.Zones[declaration.Controller.UID] = zones
	g.pushAbility(g.actionAbilityInstance(declaration))
	g.state.Knowledge.Choice = nil
	g.state.Knowledge.Declaration = nil
	g.grantOpportunity(declaration.Controller)
	return nil
}

func (g *Game) canCommitActionDeclaration(declaration *actionDeclaration) bool {
	if !g.isLegalTarget(declaration.Target) {
		return false
	}
	candidate, exists := g.state.Cards[declaration.Source]
	if !exists || !samePlayer(candidate.Owner, declaration.Controller) || !containsString(candidate.Types, "ACTION") {
		return false
	}
	zones := g.state.Zones[declaration.Controller.UID]
	return cardIndex(zones.Hand, declaration.Source) >= 0 && len(zones.Memory) >= g.actionReserveCost(declaration.Controller, declaration.Source)
}

func (g *Game) actionReserveCost(player *model.Player, card cardInstanceID) int {
	cost := g.characteristicsForCard(card).ReserveCost
	if g.state.Cards[card].Definition == trumpSetCardID && g.championHasClass(player, g.state.Cards[card].Classes) && cost > 0 {
		return cost - 1
	}
	return cost
}

func (g *Game) legalTargets() []objectID {
	targets := make([]objectID, 0, len(g.state.Champions)+len(g.state.Objects))
	for _, player := range g.players {
		champion, exists := g.state.Champions[player.UID]
		if exists {
			targets = append(targets, champion.ID)
		}
	}
	for id, object := range g.state.Objects {
		if containsString(object.Types, "ALLY") {
			targets = append(targets, id)
		}
	}
	sort.Slice(
		targets,
		func(first, second int) bool {
			return targets[first] < targets[second]
		},
	)
	return targets
}

func (g *Game) isLegalTarget(target objectID) bool {
	for _, champion := range g.state.Champions {
		if champion.ID == target {
			return true
		}
	}
	object, exists := g.state.Objects[target]
	return exists && containsString(object.Types, "ALLY")
}

func (g *Game) legalWeapons(player *model.Player) []objectID {
	weapons := make([]objectID, 0, len(g.state.Objects))
	for id := range g.state.Objects {
		if g.isLegalWeapon(player, id) {
			weapons = append(weapons, id)
		}
	}
	sort.Slice(
		weapons,
		func(first, second int) bool {
			return weapons[first] < weapons[second]
		},
	)
	return weapons
}

func (g *Game) isLegalWeapon(player *model.Player, id objectID) bool {
	object, exists := g.state.Objects[id]
	return exists && samePlayer(object.Owner, player) && containsString(object.Types, "WEAPON")
}

func (g *Game) actionAbilityInstance(declaration *actionDeclaration) abilityInstance {
	operations := []effectOperation{}
	source := g.state.Cards[declaration.Source]
	switch source.Definition {
	case blazingThrowCardID:
		operations = append(operations, effectOperation{
			Kind:   effectOperationDamage,
			Amount: 4,
		})
	case fieryInterferenceCardID:
		operations = append(operations, effectOperation{
			Kind:   effectOperationDamage,
			Amount: 2,
		})
		operations = append(operations, effectOperation{
			Kind: effectOperationContinuousModifier,
			ContinuousEffect: continuousEffect{
				Scope:         effectScopeObject,
				Layer:         effectLayerAbility,
				ExpiresAtTurn: g.state.Scheduler.TurnNumber + 1,
				Modifier: continuousModifier{
					ProhibitRecover: true,
				},
			},
		})
	case straightFlareCardID:
		operations = append(operations, effectOperation{
			Kind:                      effectOperationDamage,
			Amount:                    1,
			DistinctSuitedCostsDamage: true,
		})
	case trumpSetCardID:
		operations = append(operations, effectOperation{Kind: effectOperationRetargetAttack})
	}
	operations = append(operations, effectOperation{
		Kind:                  effectOperationMove,
		MoveSourceToGraveyard: true,
	})
	return g.newAbilityInstance(
		declaration.Controller,
		declaration.Source,
		declaration.Target,
		operations,
	)
}

func (g *Game) removeEffectSource(source cardInstanceID) {
	index := cardIndex(g.state.EffectSources, source)
	if index >= 0 {
		g.state.EffectSources = removeCardAt(g.state.EffectSources, index)
	}
}

func (g *Game) putInGraveyard(card cardInstanceID) {
	owner := g.state.Cards[card].Owner
	zones := g.state.Zones[owner.UID]
	zones.Graveyard = append(zones.Graveyard, card)
	g.state.Zones[owner.UID] = zones
}

func (g *Game) damageUnit(target objectID, amount int) {
	for playerID, champion := range g.state.Champions {
		if champion.ID == target {
			champion.Damage += amount
			champion.DamageTurn = g.state.Scheduler.TurnNumber
			g.state.Champions[playerID] = champion
			return
		}
	}
	object, exists := g.state.Objects[target]
	if exists {
		object.Damage += amount
		g.state.Objects[target] = object
	}
}

func (g *Game) isChampion(target objectID) bool {
	for _, champion := range g.state.Champions {
		if champion.ID == target {
			return true
		}
	}
	return false
}

func (g *Game) recoverChampion(target objectID, amount int) bool {
	for playerID, champion := range g.state.Champions {
		if champion.ID != target {
			continue
		}
		if g.characteristicsFor(target).RecoverProhibited {
			return false
		}
		champion.Damage -= amount
		if champion.Damage < 0 {
			champion.Damage = 0
		}
		g.state.Champions[playerID] = champion
		return true
	}
	return false
}

func (g *Game) distinctSuitedPrintedReserveCosts(player *model.Player) int {
	costs := make(map[int]struct{})
	for _, object := range g.state.Objects {
		if !samePlayer(object.Owner, player) || !g.cardHasSubtype(object.Card, "SUITED") {
			continue
		}
		costs[g.printedReserveCost(object.Card)] = struct{}{}
	}
	return len(costs)
}

func (g *Game) cardHasSubtype(card cardInstanceID, want string) bool {
	return containsString(g.state.Cards[card].Subtypes, want)
}

func (g *Game) printedReserveCost(card cardInstanceID) int {
	return g.state.Cards[card].ReserveCost
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
