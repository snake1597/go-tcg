package game

import (
	"fmt"
	"sort"

	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
)

const (
	fiveOfSpadesCardID     CardID = "i9hf5lhl5f"
	fourOfSpadesCardID     CardID = "8bolq2y5qp"
	fourOfHeartsCardID     CardID = "xgax8bbjqj"
	threeOfHeartsCardID    CardID = "1db8hz4prm"
	threeOfSpadesCardID    CardID = "o09csnorqv"
	twoOfHeartsCardID      CardID = "rufki4o41y"
	twoOfSpadesCardID      CardID = "e8ygl32jef"
	wonderlandsReignCardID CardID = "0mf1ug6yfi"
	noireCardID            CardID = "wbjc9t8ycp"
	rougeCardID            CardID = "h68dr63eo5"
)

func (g *Game) legalCardistries(player *model.Player) []objectID {
	objects := make([]objectID, 0, len(g.state.Objects))
	for id := range g.state.Objects {
		if g.canActivateCardistry(player, id) {
			objects = append(objects, id)
		}
	}
	sort.Slice(objects, func(first, second int) bool { return objects[first] < objects[second] })
	return objects
}

func (g *Game) canActivateCardistry(player *model.Player, source objectID) bool {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) {
		return false
	}
	ability, compiled := g.compiledCardistry(object.Card)
	if !compiled || ability.usage == usageOncePerObject && g.state.CardistryUsed[source] {
		return false
	}
	baseCost, fast := g.cardistryBaseCost(object.Card)
	if baseCost < 0 || !samePlayer(g.state.Scheduler.OpportunityHolder, player) {
		return false
	}
	if !fast && (!samePlayer(g.state.Scheduler.TurnPlayer, player) || g.state.Scheduler.Phase != constants.PhaseMain || len(g.state.EffectsStack) != 0) {
		return false
	}
	reduction := g.cardistryReduction(object.Card)
	cost := g.cardistryCost(player, baseCost, reduction)
	return g.canPayMemoryCost(
		player,
		cost,
	)
}

// cardistryBaseCost 回傳已編譯 Cardistry 定義的基礎費用與 Fast 屬性。
// 輸入為卡牌實例；輸出為費用與 Fast 旗標，未編譯 Cardistry 時回傳 -1 和 false，副作用為零。
func (g *Game) cardistryBaseCost(card cardInstanceID) (int, bool) {
	if ability, exists := g.compiledCardistry(card); exists {
		return ability.baseCost, ability.timing == timingFast
	}
	return -1, false
}

// cardistryReduction 回傳已編譯 Cardistry 定義的費用減免規則。
// 輸入為卡牌實例；輸出為費用減免種類，未編譯 Cardistry 時回傳空值，副作用為零。
func (g *Game) cardistryReduction(card cardInstanceID) abilityCostReduction {
	if ability, exists := g.compiledCardistry(card); exists {
		return ability.reduction
	}
	return ""
}

// cardistryCost 按定義中的減免規則及玩家折扣計算費用，最低為 0。
// 輸入為玩家、基礎費用與減免種類；輸出為應付 Memory 數量，副作用為零。
func (g *Game) cardistryCost(player *model.Player, baseCost int, reduction abilityCostReduction) int {
	if reduction != reductionDistinctSuitedCosts {
		message := fmt.Sprintf("unsupported Cardistry cost reduction %q", reduction)
		panic(message)
	}
	costs := make(map[int]struct{})
	for _, object := range g.state.Objects {
		if samePlayer(object.Owner, player) && g.cardHasSubtype(object.Card, "SUITED") {
			costs[g.printedReserveCost(object.Card)] = struct{}{}
		}
	}
	cost := baseCost - len(costs) - g.state.CardistryDiscounts[player.UID]
	if cost < 0 {
		return 0
	}
	return cost
}

// activateCardistry 宣告指定物件的 Cardistry，並以通用 Memory 付款來源支付其費用。
// 輸入為控制者、來源物件與非隨機 Memory 付款 handles；輸出為啟動錯誤或 nil，成功時推入能力並授予行動機會。
func (g *Game) activateCardistry(player *model.Player, source objectID, memoryPayment []ViewHandle) error {
	return g.commitDeclarationTransaction(
		func() error {
			return g.activateCardistryTransaction(
				player,
				source,
				memoryPayment,
			)
		},
	)
}

// activateCardistryTransaction 在候選狀態中驗證 Cardistry、支付 Memory Cost 並建立 Ability Instance。
// 輸入為控制者、來源物件與非隨機 Memory 付款 handles；輸出為啟動錯誤或 nil；成功時只修改目前的候選權威狀態。
func (g *Game) activateCardistryTransaction(player *model.Player, source objectID, memoryPayment []ViewHandle) error {
	if !g.canActivateCardistry(player, source) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, source)
	}
	object := g.state.Objects[source]
	baseCost, _ := g.cardistryBaseCost(object.Card)
	reduction := g.cardistryReduction(object.Card)
	cost := g.cardistryCost(player, baseCost, reduction)
	if err := g.payMemoryCost(
		player,
		cost,
		memoryPayment,
		"banish-floating-memory",
		"banish-memory",
	); err != nil {
		return fmt.Errorf("pay memory cost: %w", err)
	}
	ability, compiled := g.compiledCardistry(object.Card)
	if !compiled {
		return fmt.Errorf("compiled Cardistry definition is missing for %q", object.Card)
	}
	if ability.usage == usageOncePerObject {
		g.state.CardistryUsed[source] = true
	}
	delete(g.state.CardistryDiscounts, player.UID)
	g.pushAbility(g.cardistryAbility(player, source, object.Card))
	g.recordPublicEvent(player, "cardistry", "ability-activated", object.Card)
	g.flushTriggers(g.cardistryObserverTriggers(player, source))
	g.grantOpportunity(player)
	return nil
}

// canPayMemoryCost 判斷玩家的隨機 Memory 與非隨機付款來源是否足以支付指定費用。
// 輸入為付款玩家與 Memory Cost；輸出為可否完成支付，無副作用。
func (g *Game) canPayMemoryCost(player *model.Player, cost int) bool {
	zones := g.state.Zones[player.UID]
	return len(zones.Memory)+len(g.memoryPaymentCards(player)) >= cost
}

// memoryPaymentCards 回傳目前可作為任一 Memory Cost 非隨機付款來源的卡牌。
// 輸入為付款玩家；輸出為其墓地中具有 Floating Memory 的卡牌，無副作用。
func (g *Game) memoryPaymentCards(player *model.Player) []cardInstanceID {
	zones := g.state.Zones[player.UID]
	cards := make([]cardInstanceID, 0, len(zones.Graveyard))
	for _, card := range zones.Graveyard {
		instance := g.state.Cards[card]
		if samePlayer(instance.Owner, player) && instance.Definition == fiveOfSpadesCardID {
			cards = append(cards, card)
		}
	}
	return cards
}

// payMemoryCost 先驗證所有指定的非隨機 Memory 付款來源、重複牌與付款總額，再修改區域。
// 指定墓地牌先放逐，剩餘費用以對局亂數從 Memory 選牌支付，並逐張記錄付款事件。
// 輸入包含付款事件名稱；指定牌超過費用或總額不足時回傳 ErrInvalidViewHandle，尚不付款或推進亂數。
func (g *Game) payMemoryCost(player *model.Player, cost int, memoryPayment []ViewHandle, floatingEvent string, randomEvent string) error {
	zones := g.state.Zones[player.UID]
	floatingCards, err := g.memoryPaymentCardsForHandles(player, cost, memoryPayment)
	if err != nil {
		return fmt.Errorf("validate Memory payment: %w", err)
	}
	for _, card := range floatingCards {
		index := cardIndex(zones.Graveyard, card)
		zones.Graveyard = removeCardAt(zones.Graveyard, index)
		zones.Banishment = append(zones.Banishment, card)
		g.recordPublicEvent(player, "cost", floatingEvent, card)
	}
	for payment := len(floatingCards); payment < cost; payment++ {
		index := int(g.nextRandom() % uint64(len(zones.Memory)))
		card := zones.Memory[index]
		zones.Memory = removeCardAt(zones.Memory, index)
		zones.Banishment = append(zones.Banishment, card)
		g.recordPublicEvent(player, "cost", randomEvent, card)
	}
	g.state.Zones[player.UID] = zones
	return nil
}

// memoryPaymentCardsForHandles 驗證 Floating Memory handles 並以輸入順序回傳付款卡。
// 輸入為付款玩家、費用與公開 handles；輸出為不重複且足額的付款卡或驗證錯誤，無副作用。
func (g *Game) memoryPaymentCardsForHandles(player *model.Player, cost int, memoryPayment []ViewHandle) ([]cardInstanceID, error) {
	zones := g.state.Zones[player.UID]
	floatingCards := make([]cardInstanceID, 0, len(memoryPayment))
	seen := make(map[cardInstanceID]struct{}, len(memoryPayment))
	for _, handle := range memoryPayment {
		card, err := g.memoryPaymentCardForHandle(player, handle)
		if err != nil {
			return nil, fmt.Errorf("resolve Floating Memory payment handle: %w", err)
		}
		if _, exists := seen[card]; exists {
			return nil, fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, handle)
		}
		seen[card] = struct{}{}
		floatingCards = append(floatingCards, card)
	}
	if len(floatingCards) > cost || len(zones.Memory)+len(floatingCards) < cost {
		return nil, fmt.Errorf("%w", tcgErrors.ErrInvalidViewHandle)
	}
	return floatingCards, nil
}

// memoryPaymentCardForHandle 反查玩家指定的非隨機 Memory 付款來源。
// 輸入為付款玩家與 PlayerView handle；輸出為墓地中的 Floating Memory 卡牌或驗證錯誤，無副作用。
func (g *Game) memoryPaymentCardForHandle(player *model.Player, handle ViewHandle) (cardInstanceID, error) {
	for card, candidate := range g.getPlayerKnowledge(player).Cards {
		if candidate != handle {
			continue
		}
		instance := cardInstanceID(card)
		if cardIndex(g.state.Zones[player.UID].Graveyard, instance) >= 0 && g.state.Cards[instance].Definition == fiveOfSpadesCardID {
			return instance, nil
		}
	}
	return "", fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, handle)
}

func (g *Game) cardistryAbility(player *model.Player, source objectID, card cardInstanceID) abilityInstance {
	ability, exists := g.compiledCardistry(card)
	if !exists {
		message := fmt.Sprintf(
			"compiled Cardistry definition is missing for %q",
			card,
		)
		panic(message)
	}
	operations := ability.operations()
	instance := g.newAbilityInstance(
		player,
		card,
		source,
		operations,
	)
	instance.Slot = ability.slot
	return instance
}

func (g *Game) controlledSuitedAllies(player *model.Player) []objectID {
	objects := []objectID{}
	for id, object := range g.state.Objects {
		if samePlayer(object.Owner, player) && containsString(object.Types, "ALLY") && g.cardHasSubtype(object.Card, "SUITED") {
			objects = append(objects, id)
		}
	}
	sort.Slice(objects, func(first, second int) bool { return objects[first] < objects[second] })
	return objects
}

func (g *Game) deployAlly(player *model.Player, card cardInstanceID) {
	candidate, exists := g.state.Cards[card]
	if !exists || !samePlayer(candidate.Owner, player) || !containsString(candidate.Types, "ALLY") {
		return
	}
	if !g.takeSourceCard(card, sourceCardOriginMemory) {
		return
	}
	g.putAllyOnField(player, card)
}

func (g *Game) putAllyOnField(player *model.Player, card cardInstanceID) {
	id := g.putFieldObject(player, card, "ally")
	g.recordPublicEvent(player, "ability", "deploy", card)
	g.enqueueEnterAbilities(player, id, card)
}

// putMaterialRegaliaOnField 將已從 Material Deck 付款完成的 Regalia 放入控制者戰場。
// 輸入為控制者與 Regalia 實例；無輸出；會建立戰場物件、套用 Hindered 並記錄公開事件。
func (g *Game) putMaterialRegaliaOnField(player *model.Player, card cardInstanceID) {
	g.putFieldObject(player, card, "regalia")
	g.recordPublicEvent(player, "materialize", "regalia-entered", card)
}

// putFieldObject 建立由指定玩家控制的戰場物件，並依卡牌的 Hindered 關鍵字決定初始休息狀態。
// 輸入為控制者、卡牌實例與穩定的物件種類前綴；輸出為新物件 ID；會遞增 NextObject 並更新 Objects。
func (g *Game) putFieldObject(player *model.Player, card cardInstanceID, objectKind string) objectID {
	g.state.NextObject++
	id := objectID(fmt.Sprintf("%s:%d", objectKind, g.state.NextObject))
	candidate := g.state.Cards[card]
	g.state.Objects[id] = fieldObject{
		ID:     id,
		Card:   card,
		Owner:  player,
		Types:  candidate.Types,
		Rested: g.hasHindered(card),
	}
	return id
}

// hasHindered 由卡牌的已編譯 keyword Slot 判斷是否應以 rested 狀態進場。
// 輸入為卡牌實例；輸出為是否 Hindered；不會改變對局狀態。
func (g *Game) hasHindered(card cardInstanceID) bool {
	definition, exists := g.definitions[g.state.Cards[card].Definition]
	if !exists {
		return false
	}
	for _, ability := range definition.abilities {
		if ability.kind == abilityKindKeyword && ability.keyword == keywordHindered {
			return true
		}
	}
	return false
}

// enqueueEnterAbilities 從來源卡已編譯的 enter triggers 建立 Ability Instances。
// 輸入為控制者、場上來源與卡牌實例；無輸出；副作用為依定義順序把可執行能力加入 stack。
func (g *Game) enqueueEnterAbilities(player *model.Player, source objectID, card cardInstanceID) {
	for _, ability := range g.compiledTriggeredAbilities(card, eventKindEnter) {
		operations := ability.operations()
		if len(operations) == 0 || !g.canEnqueueEnterOperations(player, source, operations) {
			continue
		}
		if ability.trigger.optional && len(operations) > 0 && operations[0].Kind == effectOperationChoose {
			operations[0].CanPass = true
		}
		for index := range operations {
			if operations[index].Selector == selectorOtherControlledAllies {
				operations[index].Source = source
			}
		}
		instance := g.newAbilityInstance(
			player,
			card,
			source,
			operations,
		)
		instance.Slot = ability.slot
		g.pushAbility(instance)
	}
}

// canEnqueueEnterOperations 在建立 trigger 前檢查動態門檻與必要選項。
// 輸入為控制者、來源與已編譯 operations；輸出為是否有可結算內容，副作用為零。
func (g *Game) canEnqueueEnterOperations(player *model.Player, source objectID, operations []effectOperation) bool {
	for _, operation := range operations {
		switch operation.Kind {
		case effectOperationCounter, effectOperationDamage:
			if operation.Value != nil && g.evaluateValue(*operation.Value, player) == 0 {
				return false
			}
		case effectOperationChoose:
			if operation.Selector == selectorOtherControlledAllies && len(g.otherControlledAllies(player, source)) == 0 {
				return false
			}
		}
	}
	return true
}

// otherControlledAllies 回傳控制者除來源外的 Ally，並以 object ID 排序以維持 replay 決定性。
// 輸入為控制者與要排除的來源；輸出為可選場上物件，副作用為零。
func (g *Game) otherControlledAllies(player *model.Player, source objectID) []objectID {
	options := make([]objectID, 0)
	for id, object := range g.state.Objects {
		if id != source && samePlayer(object.Owner, player) && containsString(object.Types, "ALLY") {
			options = append(options, id)
		}
	}
	sort.Slice(options, func(first, second int) bool {
		return options[first] < options[second]
	})
	return options
}

func (g *Game) temporaryModifierOperation(power, life int) effectOperation {
	return effectOperation{
		Kind: effectOperationContinuousModifier,
		ContinuousEffect: continuousEffect{
			Scope:         effectScopeObject,
			Layer:         effectLayerModifier,
			PowerLife:     powerLifeModify,
			ExpiresAtTurn: g.state.Scheduler.TurnNumber + 1,
			Modifier: continuousModifier{
				PowerDelta: power,
				LifeDelta:  life,
			},
		},
	}
}
