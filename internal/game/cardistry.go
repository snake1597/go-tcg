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
	if !exists || !samePlayer(object.Owner, player) || g.state.CardistryUsed[source] {
		return false
	}
	baseCost, fast := g.cardistryBaseCost(object.Card)
	if baseCost < 0 || !samePlayer(g.state.Scheduler.OpportunityHolder, player) {
		return false
	}
	if !fast && (!samePlayer(g.state.Scheduler.TurnPlayer, player) || g.state.Scheduler.Phase != constants.PhaseMain || len(g.state.EffectsStack) != 0) {
		return false
	}
	return g.canPayMemoryCost(
		player,
		g.cardistryCost(player, baseCost),
	)
}

// cardistryBaseCost 回傳已實作卡牌的 Cardistry 基礎費用與 Fast 屬性。
// 未實作或找不到定義時以費用 -1 表示不可啟動。
func (g *Game) cardistryBaseCost(card cardInstanceID) (int, bool) {
	switch g.state.Cards[card].Definition {
	case fiveOfSpadesCardID:
		return 5, false
	case fourOfSpadesCardID:
		return 4, false
	case fourOfHeartsCardID:
		return 4, false
	case threeOfHeartsCardID:
		return 3, false
	case threeOfSpadesCardID:
		return 3, false
	case twoOfHeartsCardID:
		return 2, false
	case twoOfSpadesCardID:
		return 2, true
	case wonderlandsReignCardID:
		return 10, false
	case duchessCardID:
		return 6, false
	default:
		return -1, false
	}
}

// cardistryCost 按受控 Suited 物件的不同印刷 reserve cost 數量與玩家折扣減費，最低為 0。
// 相同印刷費用的多個物件只貢獻一次減費。
func (g *Game) cardistryCost(player *model.Player, baseCost int) int {
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
	if !g.canActivateCardistry(player, source) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, source)
	}
	object := g.state.Objects[source]
	baseCost, _ := g.cardistryBaseCost(object.Card)
	if err := g.payMemoryCost(
		player,
		g.cardistryCost(player, baseCost),
		memoryPayment,
		"banish-floating-memory",
		"banish-memory",
	); err != nil {
		return err
	}
	g.state.CardistryUsed[source] = true
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
	floatingCards := make(map[cardInstanceID]struct{}, len(memoryPayment))
	for _, handle := range memoryPayment {
		card, err := g.memoryPaymentCardForHandle(player, handle)
		if err != nil {
			return err
		}
		if _, exists := floatingCards[card]; exists {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, handle)
		}
		floatingCards[card] = struct{}{}
	}
	if len(floatingCards) > cost || len(zones.Memory)+len(floatingCards) < cost {
		return fmt.Errorf("%w", tcgErrors.ErrInvalidViewHandle)
	}
	for card := range floatingCards {
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

// memoryPaymentCardForHandle 反查玩家指定的非隨機 Memory 付款來源。
// 輸入為付款玩家與 PlayerView handle；輸出為墓地中的 Floating Memory 卡牌或驗證錯誤，無副作用。
func (g *Game) memoryPaymentCardForHandle(player *model.Player, handle ViewHandle) (cardInstanceID, error) {
	for card, candidate := range g.state.Knowledge.Cards[player.UID] {
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
	operations := []effectOperation{}
	switch g.state.Cards[card].Definition {
	case wonderlandsReignCardID:
		operations = append(operations, effectOperation{Kind: effectOperationDraw, Amount: 1})
	case fiveOfSpadesCardID:
		operations = append(operations, g.temporaryModifierOperation(5, 0))
	case fourOfSpadesCardID:
		operations = append(operations, effectOperation{Kind: effectOperationDrawToMemory, Amount: 1})
	case fourOfHeartsCardID:
		operations = append(operations, effectOperation{Kind: effectOperationDrawToMemory, Amount: 1})
		operations = append(operations, effectOperation{Kind: effectOperationChooseMemoryAlly})
		operations = append(operations, effectOperation{Kind: effectOperationDeploy})
	case threeOfHeartsCardID:
		operations = append(operations, effectOperation{Kind: effectOperationDraw, Amount: 1})
		operations = append(operations, effectOperation{Kind: effectOperationChooseHandCard})
		operations = append(operations, effectOperation{Kind: effectOperationDiscard})
	case threeOfSpadesCardID:
		operations = append(operations, effectOperation{Kind: effectOperationChoose, Options: g.controlledSuitedAllies(player)})
		operations = append(operations, g.temporaryModifierOperation(0, 2))
	case twoOfHeartsCardID:
		operations = append(operations, g.temporaryModifierOperation(2, 0))
	case twoOfSpadesCardID:
		operations = append(operations, effectOperation{Kind: effectOperationCounter, Counter: "BUFF", Amount: 1})
	case duchessCardID:
		operations = append(
			operations,
			effectOperation{
				Kind: effectOperationChooseDuchessCopy,
			},
		)
		operations = append(
			operations,
			effectOperation{
				Kind: effectOperationCopyDuchessAction,
			},
		)
	}
	return g.newAbilityInstance(player, card, source, operations)
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

func (g *Game) qualifiedMemoryAllies(player *model.Player) []cardInstanceID {
	zones := g.state.Zones[player.UID]
	cards := []cardInstanceID{}
	for _, card := range zones.Memory {
		candidate := g.state.Cards[card]
		if containsString(candidate.Types, "ALLY") && containsString(candidate.Subtypes, "SUITED") && candidate.ReserveCost <= 3 && (containsString(candidate.Elements, "FIRE") || containsString(candidate.Elements, "NORM")) {
			cards = append(cards, card)
		}
	}
	sort.Slice(cards, func(first, second int) bool { return cards[first] < cards[second] })
	return cards
}

func (g *Game) deployAlly(player *model.Player, card cardInstanceID) {
	zones := g.state.Zones[player.UID]
	index := cardIndex(zones.Memory, card)
	if index < 0 {
		return
	}
	zones.Memory = removeCardAt(zones.Memory, index)
	g.state.Zones[player.UID] = zones
	g.putAllyOnField(player, card)
}

func (g *Game) putAllyOnField(player *model.Player, card cardInstanceID) {
	id := g.putFieldObject(player, card, "ally")
	g.recordPublicEvent(player, "ability", "deploy", card)
	g.enqueueSuitedEnterAbility(player, id, card)
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

// hasHindered 回傳固定支援卡中具有 Hindered 關鍵字、應以 rested 狀態進場的卡牌。
// 輸入為卡牌實例；輸出為是否 Hindered；不會改變對局狀態。
func (g *Game) hasHindered(card cardInstanceID) bool {
	return g.state.Cards[card].Definition == duchessThornesCardID
}

func (g *Game) suitedReserveTotal(player *model.Player) int {
	total := 0
	for _, object := range g.state.Objects {
		if samePlayer(object.Owner, player) && g.cardHasSubtype(object.Card, "SUITED") {
			total += g.printedReserveCost(object.Card)
		}
	}
	return total
}

func suitedThresholdAmount(total int) int {
	switch {
	case total >= 21:
		return 4
	case total >= 10:
		return 2
	case total >= 6:
		return 1
	default:
		return 0
	}
}

func (g *Game) enqueueSuitedEnterAbility(player *model.Player, source objectID, card cardInstanceID) {
	switch g.state.Cards[card].Definition {
	case noireCardID:
		amount := suitedThresholdAmount(g.suitedReserveTotal(player))
		if amount > 0 {
			g.pushAbility(g.newAbilityInstance(player, card, source, []effectOperation{{Kind: effectOperationCounter, Counter: "BUFF", Amount: amount}}))
		}
	case rougeCardID:
		if suitedThresholdAmount(g.suitedReserveTotal(player)) > 0 {
			g.pushAbility(g.newAbilityInstance(player, card, "", []effectOperation{{Kind: effectOperationChoose, Options: g.legalTargets()}, {Kind: effectOperationSuitedThresholdDamage}}))
		}
	case pepperedChefCardID:
		options := []objectID{}
		for id, object := range g.state.Objects {
			if id != source && samePlayer(object.Owner, player) && containsString(object.Types, "ALLY") {
				options = append(options, id)
			}
		}
		sort.Slice(options, func(first, second int) bool { return options[first] < options[second] })
		if len(options) > 0 {
			g.pushAbility(g.newAbilityInstance(
				player,
				card,
				source,
				[]effectOperation{
					{
						Kind:    effectOperationChoose,
						Options: options,
						CanPass: true,
					},
					{
						Kind:   effectOperationSacrificeForChef,
						Source: source,
					},
				},
			))
		}
	}
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
