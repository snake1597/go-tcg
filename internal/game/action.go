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

type sourceCardOrigin string

const (
	declarationTarget declarationStage = "target"
	declarationWeapon declarationStage = "weapon"

	sourceCardOriginEffectsStack sourceCardOrigin = "effects_stack"
	sourceCardOriginMemory       sourceCardOrigin = "memory"
)

type actionDeclaration struct {
	Controller *model.Player  `json:"controller"`
	Source     cardInstanceID `json:"source"`
	// Reserved 保存宣告時由玩家明確選定、提交時才會移至 Memory 的手牌付款。
	Reserved []cardInstanceID `json:"reserved"`
	Target   objectID         `json:"target,omitempty"`
	Weapon   objectID         `json:"weapon,omitempty"`
	Stage    declarationStage `json:"stage"`
}

// actionDeclarationSnapshot 保存開始 Action 宣告前、宣告流程可能修改的權威欄位與 replay 邊界。
// 取消時只還原這些欄位，避免覆蓋宣告期間不可能改動的區域與物件狀態。
type actionDeclarationSnapshot struct {
	Revision    uint64
	Knowledge   knowledgeState
	Entities    map[entityID]knowledgeEntity
	NextHandle  uint64
	ReplaySteps int
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
	ability, exists := g.compiledAction(card)
	actionTargets := g.actionTargets(player, ability)
	if !exists || len(actionTargets) == 0 {
		return false
	}
	if ability.cost == nil {
		return true
	}
	costTargets := g.actionCostTargets(player, *ability.cost)
	return len(costTargets) > 0
}

// beginActionDeclaration 為行動建立選目標的宣告，尚不移走來源牌或支付費用。
// Blazing Throw 選完目標後還須選擇犧牲武器；Trump Set 僅能選受控的 Suited ally。
// Action 依其目標規則建立宣告；替代費用在啟動入口單獨選擇。
func (g *Game) beginActionDeclaration(player *model.Player, card cardInstanceID, reserve []ViewHandle) error {
	ability, exists := g.compiledAction(card)
	if !exists {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	targets := g.actionTargets(player, ability)
	if !g.canActivateAction(player, card) || len(targets) == 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	reserved, err := g.reserveCardsForHandles(player, card, reserve)
	if err != nil {
		return fmt.Errorf("reserve action cards: %w", err)
	}
	g.captureActionDeclarationSnapshot()
	g.state.Knowledge.Declaration = &actionDeclaration{
		Controller: player,
		Source:     card,
		Reserved:   reserved,
		Stage:      declarationTarget,
	}
	g.setDeclarationChoice(player, targets)
	g.state.Knowledge.Choice.CanPass = true
	return nil
}

// captureActionDeclarationSnapshot 保存開始宣告前的 revision、Knowledge、handle 與 replay 邊界。
// 無輸入與輸出；副作用為建立非權威取消快照，不改變 canonical Game State。
func (g *Game) captureActionDeclarationSnapshot() {
	state := cloneGameState(
		g.state,
	)
	g.actionDeclarationSnapshot = &actionDeclarationSnapshot{
		Revision:    state.Revision,
		Knowledge:   state.Knowledge,
		Entities:    state.Entities,
		NextHandle:  state.NextHandle,
		ReplaySteps: len(g.replay.Steps),
	}
}

// cancelActionDeclaration 放棄尚未提交的 Action 宣告並保留來源與 Reserve 付款卡的原區域。
// 輸入為取消玩家；輸出為錯誤或 nil；成功時清除暫存宣告與選擇，且不建立來源卡或 Stack Item。
func (g *Game) cancelActionDeclaration(player *model.Player) error {
	declaration := g.state.Knowledge.Declaration
	choice := g.state.Knowledge.Choice
	snapshot := g.actionDeclarationSnapshot
	if declaration == nil || choice == nil || snapshot == nil || !choice.CanPass || !samePlayer(declaration.Controller, player) || !samePlayer(choice.Actor, player) {
		return fmt.Errorf("%w", tcgErrors.ErrInvalidViewHandle)
	}
	g.state.Revision = snapshot.Revision
	g.state.Knowledge = snapshot.Knowledge
	g.state.Entities = snapshot.Entities
	g.state.NextHandle = snapshot.NextHandle
	g.replay.Steps = g.replay.Steps[:snapshot.ReplaySteps]
	g.actionDeclarationSnapshot = nil
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
	return g.commitDeclarationTransaction(
		func() error {
			return g.commitAllyActivationTransaction(
				player,
				card,
				reserve,
			)
		},
	)
}

// commitAllyActivationTransaction 驗證 Ally 的來源與 Reserve 付款，並在候選狀態中建立來源卡及 activation Stack Item。
// 輸入為控制者、Ally 卡牌與 Player View Reserve handles；輸出為驗證或提交錯誤；成功時只修改目前的候選權威狀態。
func (g *Game) commitAllyActivationTransaction(player *model.Player, card cardInstanceID, reserve []ViewHandle) error {
	if !g.canActivateAction(player, card) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	reserved, err := g.reserveCardsForHandles(player, card, reserve)
	if err != nil {
		return fmt.Errorf("reserve Ally cards: %w", err)
	}
	zones := g.state.Zones[player.UID]
	zones, err = applyReservePayment(
		zones,
		card,
		reserved,
	)
	if err != nil {
		return fmt.Errorf("apply Ally Reserve payment: %w", err)
	}
	return g.queueAllyActivationTransaction(
		player,
		card,
		zones,
	)
}

// queueAllyActivationTransaction 寫入已驗證的 Ally 付款與來源移動，並只建立一次 Ability Instance。
// 輸入為控制者、來源卡與已移除來源及套用付款的區域快照；輸出為提交錯誤或 nil，副作用為更新區域、Effects Stack 與 Opportunity。
func (g *Game) queueAllyActivationTransaction(player *model.Player, card cardInstanceID, zones playerZones) error {
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
	if ability, exists := g.compiledAlly(card); exists {
		instance.Slot = ability.slot
	}
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
		ability, exists := g.compiledAction(declaration.Source)
		actionTargets := g.actionTargets(player, ability)
		if !exists || !containsObject(actionTargets, target) {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
		}
		declaration.Target = target
		if ability.cost != nil {
			declaration.Stage = declarationWeapon
			costTargets := g.actionCostTargets(player, *ability.cost)
			g.setDeclarationChoice(player, costTargets)
			g.state.Knowledge.Choice.CanPass = true
			return nil
		}
		return g.commitActionDeclaration()
	case declarationWeapon:
		weapon := objectID(subject)
		ability, exists := g.compiledAction(declaration.Source)
		if !exists || ability.cost == nil {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
		}
		costTargets := g.actionCostTargets(player, *ability.cost)
		if !containsObject(costTargets, weapon) {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
		}
		_, err := g.validateObjectSacrifice(
			objectSacrificeRequest{
				Controller:   player,
				Object:       weapon,
				RequiredType: "WEAPON",
				Cause:        "cost:sacrifice",
			},
		)
		if err != nil {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
		}
		declaration.Weapon = weapon
		return g.commitActionDeclaration()
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

// commitActionDeclaration 在目標、來源手牌與 reserve 選擇仍有效時，將來源與保留牌移出手牌。
// 保留牌移入 Memory，來源與能力入堆疊，並清除宣告及待選項目後授予控制者行動機會。
func (g *Game) commitActionDeclaration() error {
	err := g.commitDeclarationTransaction(
		func() error {
			return g.commitActionDeclarationTransaction()
		},
	)
	if err != nil {
		return fmt.Errorf("commit Action declaration: %w", err)
	}
	g.actionDeclarationSnapshot = nil
	return nil
}

// commitActionDeclarationTransaction 在候選狀態中重新驗證已選目標、Reserve 與額外費用，並建立 Action 來源與 Stack Item。
// 輸入取自目前 Action 宣告；輸出為提交錯誤或 nil；成功時只修改目前的候選權威狀態。
func (g *Game) commitActionDeclarationTransaction() error {
	declaration := g.state.Knowledge.Declaration
	if declaration == nil || !g.canCommitActionDeclaration(declaration) {
		return fmt.Errorf("%w", tcgErrors.ErrInvalidViewHandle)
	}
	zones, err := applyReservePayment(
		g.state.Zones[declaration.Controller.UID],
		declaration.Source,
		declaration.Reserved,
	)
	if err != nil {
		return fmt.Errorf("apply Action Reserve payment: %w", err)
	}
	g.state.Zones[declaration.Controller.UID] = zones
	if ability, exists := g.compiledAction(declaration.Source); exists && ability.cost != nil {
		if err := g.payActionCost(declaration.Controller, declaration.Weapon, *ability.cost); err != nil {
			return fmt.Errorf("pay action cost: %w", err)
		}
	}
	g.state.EffectSources = append(g.state.EffectSources, declaration.Source)
	g.pushAbility(g.actionAbilityInstance(declaration))
	g.state.Knowledge.Choice = nil
	g.state.Knowledge.Declaration = nil
	g.grantOpportunity(declaration.Controller)
	return nil
}

// applyReservePayment 先驗證來源卡與所有付款卡皆仍在 Hand，再一次產生來源移除且付款卡移入 Memory 的區域結果。
// 輸入為原區域、來源卡及已選付款卡；輸出為新的區域值或驗證錯誤；不修改傳入切片或權威 Game State。
func applyReservePayment(zones playerZones, source cardInstanceID, payments []cardInstanceID) (playerZones, error) {
	if cardIndex(zones.Hand, source) < 0 {
		return playerZones{}, fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, source)
	}
	seen := make(map[cardInstanceID]bool, len(payments))
	for _, payment := range payments {
		if payment == source || seen[payment] || cardIndex(zones.Hand, payment) < 0 {
			return playerZones{}, fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, payment)
		}
		seen[payment] = true
	}
	zones.Hand = append(
		[]cardInstanceID(nil),
		zones.Hand...,
	)
	zones.Memory = append(
		[]cardInstanceID(nil),
		zones.Memory...,
	)
	zones.Hand = removeCardAt(
		zones.Hand,
		cardIndex(zones.Hand, source),
	)
	for _, payment := range payments {
		zones.Hand = removeCardAt(
			zones.Hand,
			cardIndex(zones.Hand, payment),
		)
		zones.Memory = append(
			zones.Memory,
			payment,
		)
	}
	return zones, nil
}

func (g *Game) canCommitActionDeclaration(declaration *actionDeclaration) bool {
	ability, exists := g.compiledAction(declaration.Source)
	actionTargets := g.actionTargets(declaration.Controller, ability)
	if !exists || !containsObject(actionTargets, declaration.Target) {
		return false
	}
	if ability.cost != nil {
		costTargets := g.actionCostTargets(declaration.Controller, *ability.cost)
		if !containsObject(costTargets, declaration.Weapon) {
			return false
		}
		_, err := g.validateObjectSacrifice(
			objectSacrificeRequest{
				Controller:   declaration.Controller,
				Object:       declaration.Weapon,
				RequiredType: "WEAPON",
				Cause:        "cost:sacrifice",
			},
		)
		if err != nil {
			return false
		}
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

// actionTargets 依已編譯 Action selector 列舉目前可公開選擇的宣告目標。
// 輸入為控制者與已驗證 Action 定義；輸出為穩定排序的合法物件，副作用為零。
func (g *Game) actionTargets(player *model.Player, ability compiledAbilityDefinition) []objectID {
	if ability.target == nil {
		return nil
	}
	switch ability.target.kind {
	case selectorUnits:
		return g.selectTargets(*ability.target)
	case selectorRetargetableControlledSuitedAlly:
		return g.retargetableAttackTargets(player)
	default:
		return nil
	}
}

// actionCostTargets 列舉可支付已編譯 Action 額外費用的物件。
// 輸入為控制者與費用定義；輸出為穩定排序的合法支付對象，副作用為零。
func (g *Game) actionCostTargets(player *model.Player, cost actionCostDefinition) []objectID {
	if cost.kind != actionCostSacrificeControlledWeapon {
		return nil
	}
	return g.legalWeapons(player)
}

// payActionCost 提交已在宣告交易內重新驗證的 Action 額外費用。
// 輸入為控制者、支付物件與費用定義；輸出為錯誤或 nil，成功時將支付物件移至擁有者墓地。
func (g *Game) payActionCost(player *model.Player, payment objectID, cost actionCostDefinition) error {
	costTargets := g.actionCostTargets(player, cost)
	if !containsObject(costTargets, payment) {
		return fmt.Errorf("invalid action cost payment %q", payment)
	}
	_, err := g.commitObjectSacrifice(
		objectSacrificeRequest{
			Controller:   player,
			Object:       payment,
			RequiredType: "WEAPON",
			Cause:        "cost:sacrifice",
		},
	)
	if err != nil {
		return fmt.Errorf("commit Weapon sacrifice: %w", err)
	}
	return nil
}

func (g *Game) actionReserveCost(player *model.Player, card cardInstanceID) int {
	cost := g.characteristicsForCard(card).ReserveCost
	if ability, exists := g.compiledAction(card); exists && ability.classCostReduction > 0 && g.championHasClass(player, g.state.Cards[card].Classes) && cost > 0 {
		cost -= ability.classCostReduction
	}
	return cost + g.activationReserveTax(player, card)
}

// activationReserveTax 匯總場上已編譯 permission 對指定 activation 增加的 Reserve 費用。
// 輸入為啟動玩家與候選卡；輸出為所有適用 permission 的費用增量，副作用為零。
func (g *Game) activationReserveTax(player *model.Player, card cardInstanceID) int {
	candidate, exists := g.state.Cards[card]
	if !exists {
		return 0
	}
	tax := 0
	for _, object := range g.state.Objects {
		definition, definitionExists := g.definitions[g.state.Cards[object.Card].Definition]
		if !definitionExists {
			continue
		}
		for _, ability := range definition.abilities {
			permission := ability.permission
			if ability.kind != abilityKindPermission || permission == nil {
				continue
			}
			if permission.sourcePresence != sourcePresenceRequired || !containsString(candidate.Elements, permission.candidateElement) {
				continue
			}
			if permission.duringControllerTurn && !samePlayer(object.Owner, g.state.Scheduler.TurnPlayer) {
				continue
			}
			if permission.opponentsOnly && samePlayer(object.Owner, player) {
				continue
			}
			tax += permission.reserveCostDelta
		}
	}
	return tax
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
	ability, exists := g.compiledAction(declaration.Source)
	if !exists {
		return abilityInstance{}
	}
	operations := ability.operations()
	instance := g.newAbilityInstance(
		declaration.Controller,
		declaration.Source,
		declaration.Target,
		operations,
	)
	instance.Slot = ability.slot
	return instance
}

// removeEffectSource 從 Effects Stack 的來源區取走指定卡牌實例。
// 輸入為 CardInstance ID；輸出為是否取走來源；副作用為來源存在時更新 EffectSources，不存在時保持狀態不變。
func (g *Game) removeEffectSource(source cardInstanceID) bool {
	index := cardIndex(g.state.EffectSources, source)
	if index < 0 {
		return false
	}
	g.state.EffectSources = removeCardAt(g.state.EffectSources, index)
	return true
}

// takeSourceCard 由指定來源區取出仍有效的 Source Card，供後續移入目的區或建立場上 Object。
// 輸入為 CardInstance 與來源區；輸出為是否成功取出；副作用為成功時更新 Effects Stack 來源區或擁有者 Memory，失效來源不改變狀態。
func (g *Game) takeSourceCard(source cardInstanceID, origin sourceCardOrigin) bool {
	candidate, exists := g.state.Cards[source]
	if !exists {
		return false
	}
	switch origin {
	case sourceCardOriginEffectsStack:
		return g.removeEffectSource(source)
	case sourceCardOriginMemory:
		zones := g.state.Zones[candidate.Owner.UID]
		index := cardIndex(zones.Memory, source)
		if index < 0 {
			return false
		}
		zones.Memory = removeCardAt(zones.Memory, index)
		g.state.Zones[candidate.Owner.UID] = zones
		return true
	default:
		return false
	}
}

// resolveEffectSourceToGraveyard 結算仍在 Effects Stack 來源區的 Source Card 至其擁有者墓地。
// 輸入為來源 CardInstance；輸出為是否完成移動；副作用為移除來源、更新墓地，且來源失效時不退回既付費用。
func (g *Game) resolveEffectSourceToGraveyard(source cardInstanceID) bool {
	if !g.takeSourceCard(source, sourceCardOriginEffectsStack) {
		return false
	}
	g.putInGraveyard(source)
	return true
}

// resolveAllyEffectSource 結算仍在 Effects Stack 來源區的 Ally，並以共通入場流程建立場上 Object。
// 輸入為控制者與來源 CardInstance；輸出為是否成功入場；副作用為移除來源、無效 Ally 進墓地，或建立 Object、公開事件與入場觸發。
func (g *Game) resolveAllyEffectSource(player *model.Player, source cardInstanceID) bool {
	candidate, exists := g.state.Cards[source]
	if !exists {
		return false
	}
	if !g.takeSourceCard(source, sourceCardOriginEffectsStack) {
		return false
	}
	if !samePlayer(candidate.Owner, player) || !containsString(candidate.Types, "ALLY") {
		g.putInGraveyard(source)
		return false
	}
	g.putAllyOnField(player, source)
	return true
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
