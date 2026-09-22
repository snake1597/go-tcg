package game

import (
	"fmt"
	"go-tcg/internal/constants"
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
	Controller *model.Player  `json:"controller"`
	Source     cardInstanceID `json:"source"`
	// Reserved 保存宣告時由玩家明確選定、提交時才會移至 Memory 的手牌付款。
	Reserved []cardInstanceID `json:"reserved"`
	Target   objectID         `json:"target,omitempty"`
	Stage    declarationStage `json:"stage"`
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
	if !exists || !samePlayer(candidate.Owner, player) || (!containsString(candidate.Types, "ACTION") && !containsString(candidate.Types, "ALLY")) {
		return false
	}
	alternativeCostCards := g.veritaAlternativeCostCards(player)
	canUseAlternativeCost := candidate.Definition == veritaCardID && len(alternativeCostCards) > 0
	visibleReserveCards := g.visibleReserveCards(player, card)
	reserveCost := g.actionReserveCost(player, card)
	if !samePlayer(scheduler.OpportunityHolder, player) || (!canUseAlternativeCost && len(visibleReserveCards) < reserveCost) {
		return false
	}
	if !candidate.Fast && (!samePlayer(scheduler.TurnPlayer, player) || scheduler.Phase != constants.PhaseMain || len(g.state.EffectsStack) != 0) {
		return false
	}
	if containsString(candidate.Types, "ALLY") {
		return true
	}
	if candidate.Definition == blazingThrowCardID && len(g.legalWeapons(player)) == 0 {
		return false
	}
	if candidate.Definition == veritaCardID {
		return true
	}
	if candidate.Definition == trumpSetCardID {
		targets := g.trumpSetTargets(player)
		return len(targets) > 0
	}
	return candidate.Definition == blazingThrowCardID || candidate.Definition == fieryInterferenceCardID || candidate.Definition == straightFlareCardID
}

// beginActionDeclaration 為行動建立選目標的宣告，尚不移走來源牌或支付費用。
// Blazing Throw 選完目標後還須選擇犧牲武器；Trump Set 僅能選受控的 Suited ally。
// Verita 會先建立可取消的替代費用選擇；其他 action 才建立目標宣告。
func (g *Game) beginActionDeclaration(player *model.Player, card cardInstanceID, reserve []ViewHandle) error {
	if g.state.Cards[card].Definition == veritaCardID {
		if err := g.beginVeritaAlternativeCostDeclaration(player, card, reserve); err != nil {
			return fmt.Errorf("begin Verita alternative cost: %w", err)
		}
		return nil
	}
	targets := g.legalTargets()
	if g.state.Cards[card].Definition == trumpSetCardID {
		targets = g.trumpSetTargets(player)
	}
	if !g.canActivateAction(player, card) || len(targets) == 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	reserved, err := g.reserveCardsForHandles(player, card, reserve)
	if err != nil {
		return fmt.Errorf("reserve action cards: %w", err)
	}
	g.state.Knowledge.Declaration = &actionDeclaration{
		Controller: player,
		Source:     card,
		Reserved:   reserved,
		Stage:      declarationTarget,
	}
	g.setDeclarationChoice(player, targets)
	return nil
}

// reserveCardsForHandles 驗證玩家選定的手牌正好可支付來源卡目前的 reserve cost。
// 輸入為玩家、來源卡與玩家視圖 handles；輸出為內部卡牌識別或驗證錯誤，無副作用。
func (g *Game) reserveCardsForHandles(player *model.Player, source cardInstanceID, handles []ViewHandle) ([]cardInstanceID, error) {
	want := g.actionReserveCost(player, source)
	got := len(handles)
	if got != want {
		return nil, fmt.Errorf("%w: reserve cards = %d, want %d", tcgErrors.ErrInvalidViewHandle, got, want)
	}
	reserved := make([]cardInstanceID, 0, want)
	seen := make(map[cardInstanceID]bool, want)
	for _, handle := range handles {
		card, exists := g.reserveCardForHandle(player, handle)
		if !exists || card == source || cardIndex(g.state.Zones[player.UID].Hand, card) < 0 {
			return nil, fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, handle)
		}
		if seen[card] {
			return nil, fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, handle)
		}
		seen[card] = true
		reserved = append(reserved, card)
	}
	return reserved, nil
}

// reserveCardForHandle 依玩家目前的追蹤映射反查 reserve 選擇對應的卡牌。
// 輸入為玩家與不透明 handle；輸出為卡牌識別與是否存在，無副作用。
func (g *Game) reserveCardForHandle(player *model.Player, handle ViewHandle) (cardInstanceID, bool) {
	for entity, candidate := range g.state.Knowledge.Players[player.UID].Cards {
		if candidate == handle {
			return cardInstanceID(entity), true
		}
	}
	return "", false
}

// commitAllyActivation 將手牌中的 Ally 與明確選定的 Reserve 手牌原子地移到 Effects Stack 與 Memory。
// 輸入為控制者、Ally 卡牌與 Player View Reserve handles；輸出為提交錯誤，副作用為成功時建立可回應的 activation 並授予控制者 Opportunity。
func (g *Game) commitAllyActivation(player *model.Player, card cardInstanceID, reserve []ViewHandle) error {
	if !g.canActivateAction(player, card) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	if g.state.Cards[card].Definition == veritaCardID {
		if len(g.veritaAlternativeCostCards(player)) > 0 {
			if err := g.beginVeritaAlternativeCostDeclaration(player, card, reserve); err != nil {
				return fmt.Errorf("begin Verita alternative cost: %w", err)
			}
			return nil
		}
	}
	reserved, err := g.reserveCardsForHandles(player, card, reserve)
	if err != nil {
		return fmt.Errorf("reserve Ally cards: %w", err)
	}
	zones := g.state.Zones[player.UID]
	index := cardIndex(zones.Hand, card)
	if index < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	zones.Hand = removeCardAt(zones.Hand, index)
	for _, payment := range reserved {
		paymentIndex := cardIndex(zones.Hand, payment)
		if paymentIndex < 0 {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, payment)
		}
		zones.Hand = removeCardAt(zones.Hand, paymentIndex)
		zones.Memory = append(zones.Memory, payment)
	}
	g.state.Zones[player.UID] = zones
	g.queueAllyActivation(player, card)
	return nil
}

// queueAllyActivation 將已支付費用且離開手牌的 Ally 登記為 Source Card 與 activation Stack item。
// 輸入為控制者與 Ally CardInstance；輸出為零值，副作用為追加 EffectSources／EffectsStack 並將 Opportunity 授予控制者。
func (g *Game) queueAllyActivation(player *model.Player, card cardInstanceID) {
	g.state.EffectSources = append(g.state.EffectSources, card)
	operations := []effectOperation{
		{
			Kind: effectOperationPutAllyOnField,
		},
	}
	instance := g.newAbilityInstance(
		player,
		card,
		"",
		operations,
	)
	g.pushAbility(instance)
	g.grantOpportunity(player)
}

// beginVeritaAlternativeCostDeclaration 開始 Verita 的逐張替代費用選擇，或在無替代費用時以 Reserve 建立 activation。
// 輸入為控制者、手牌中的 Verita 與 Reserve handles；輸出為待選 handle 或提交錯誤，副作用為建立費用宣告或 Effects Stack item。
func (g *Game) beginVeritaAlternativeCostDeclaration(player *model.Player, card cardInstanceID, reserve []ViewHandle) error {
	if !g.canActivateAction(player, card) || cardIndex(g.state.Zones[player.UID].Hand, card) < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	if len(g.veritaAlternativeCostCards(player)) == 0 {
		if err := g.commitAllyActivation(player, card, reserve); err != nil {
			return fmt.Errorf("commit Ally activation: %w", err)
		}
		return nil
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
// 當選到至少三張且總和十時，原子地放逐付款並建立 Verita activation；否則只更新宣告，無區域副作用。
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
// 輸入為已完成的宣告；成功時將付款放逐並把 Verita activation 放上 Effects Stack，失敗時完全不改變遊戲狀態。
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
	g.queueAllyActivation(declaration.Controller, declaration.Source)
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
		handle := g.newViewHandle(player, constants.ViewHandleSubjectChoicePrefix+string(subject))
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

// commitActionDeclaration 在目標、來源手牌與 reserve 選擇仍有效時，將來源與保留牌移出手牌。
// 保留牌移入 Memory，來源與能力入堆疊，並清除宣告及待選項目後授予控制者行動機會。
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
	for _, reserved := range declaration.Reserved {
		reservedIndex := cardIndex(zones.Hand, reserved)
		if reservedIndex < 0 {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, reserved)
		}
		zones.Hand = removeCardAt(zones.Hand, reservedIndex)
		zones.Memory = append(zones.Memory, reserved)
	}
	g.state.Zones[declaration.Controller.UID] = zones
	g.state.EffectSources = append(g.state.EffectSources, declaration.Source)
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
	if cardIndex(zones.Hand, declaration.Source) < 0 || len(declaration.Reserved) != g.actionReserveCost(declaration.Controller, declaration.Source) {
		return false
	}
	for _, reserved := range declaration.Reserved {
		if reserved == declaration.Source || cardIndex(zones.Hand, reserved) < 0 {
			return false
		}
	}
	return true
}

func (g *Game) actionReserveCost(player *model.Player, card cardInstanceID) int {
	cost := g.characteristicsForCard(card).ReserveCost
	if g.state.Cards[card].Definition == trumpSetCardID && g.championHasClass(player, g.state.Cards[card].Classes) && cost > 0 {
		cost--
	}
	if g.viridianProtectiveTrinketTaxApplies(player, card) {
		return cost + 2
	}
	return cost
}

// viridianProtectiveTrinketTaxApplies 判定啟動者是否須支付 Viridian Protective Trinket 的額外費用。
// 輸入為啟動行動的玩家與卡牌；輸出為是否加稅，副作用為零。
func (g *Game) viridianProtectiveTrinketTaxApplies(player *model.Player, card cardInstanceID) bool {
	candidate, exists := g.state.Cards[card]
	if !exists || !containsString(candidate.Elements, "WATER") {
		return false
	}
	for _, object := range g.state.Objects {
		if g.state.Cards[object.Card].Definition != viridianProtectiveTrinketCardID || !samePlayer(object.Owner, g.state.Scheduler.TurnPlayer) || samePlayer(object.Owner, player) {
			continue
		}
		return true
	}
	return false
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

// removeEffectSource 從 Effects Stack 的來源區移除指定卡牌實例。
// 輸入為 CardInstance ID；輸出為零值，副作用為來源存在時更新 EffectSources，不存在時保持狀態不變。
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
