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
// 非 Fast 行動只允許在自己的主階段且效果堆疊為空時使用；有替代費用的卡可選擇其付款方式。
// 此檢查不保證來源位於手牌，也不驗證最終目標；宣告與提交階段仍須檢查。
func (g *Game) canActivateAction(player *model.Player, card cardInstanceID) bool {
	scheduler := g.state.Scheduler
	candidate, exists := g.state.Cards[card]
	if !exists || !samePlayer(candidate.Owner, player) || (!containsString(candidate.Types, "ACTION") && !containsString(candidate.Types, "ALLY")) {
		return false
	}
	canUseAlternativeCost := len(g.alternativeCostCards(player, card)) > 0
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
	if candidate.Definition == trumpSetCardID {
		targets := g.trumpSetTargets(player)
		return len(targets) > 0
	}
	return candidate.Definition == blazingThrowCardID || candidate.Definition == fieryInterferenceCardID || candidate.Definition == straightFlareCardID
}

// beginActionDeclaration 為行動建立選目標的宣告，尚不移走來源牌或支付費用。
// Blazing Throw 選完目標後還須選擇犧牲武器；Trump Set 僅能選受控的 Suited ally。
// Action 依其目標規則建立宣告；替代費用在啟動入口單獨選擇。
func (g *Game) beginActionDeclaration(player *model.Player, card cardInstanceID, reserve []ViewHandle) error {
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
	for entity, candidate := range g.getPlayerKnowledge(player).Cards {
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
