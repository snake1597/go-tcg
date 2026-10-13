package game

import (
	"reflect"
	"testing"

	"go-tcg/internal/model"
)

// TestDiscardUsesSpecifiedZone 驗證棄牌預設來自手牌，且指定 Memory 時從該區移至墓地。
func TestDiscardUsesSpecifiedZone(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	zones := game.state.Zones[player.UID]
	handCard := zones.Hand[0]
	memoryCard := zones.Hand[1]
	zones.Hand = removeCardAt(zones.Hand, 1)
	zones.Memory = append(zones.Memory, memoryCard)
	game.state.Zones[player.UID] = zones

	if err := game.discardCard(player, "", handCard); err != nil {
		t.Fatalf("discard hand card: %v", err)
	}
	if err := game.discardCard(player, cardZoneMemory, memoryCard); err != nil {
		t.Fatalf("discard Memory card: %v", err)
	}
	zones = game.state.Zones[player.UID]
	if cardIndex(zones.Hand, handCard) >= 0 || cardIndex(zones.Memory, memoryCard) >= 0 {
		t.Fatal("discarded cards remained in their source zones")
	}
	if cardIndex(zones.Graveyard, handCard) < 0 || cardIndex(zones.Graveyard, memoryCard) < 0 {
		t.Fatal("discarded cards did not reach the graveyard")
	}
}

// TestDiscardFromExpiredSourceFizzesRemainingAbility 驗證指定卡牌已離開來源區域時，棄牌失敗不會產生卡牌或繼續結算後續操作。
// 輸入為含失效棄牌與後續 counter 的能力；輸出為未變動的區域與 counter；副作用為建立隔離能力結算。
func TestDiscardFromExpiredSourceFizzesRemainingAbility(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	source := findCard(t, game, player, blazingThrowCardID)
	target := objectID("discard-fizzle-target")
	game.state.Objects[target] = fieldObject{
		ID:    target,
		Card:  source,
		Owner: player,
	}
	beforeZones := game.state.Zones[player.UID]
	instance := game.newAbilityInstance(
		player,
		source,
		target,
		[]effectOperation{
			{
				Kind:     effectOperationDiscard,
				CardZone: cardZoneHand,
				Target:   objectID("expired-card"),
			},
			{
				Kind:    effectOperationCounter,
				Counter: "POWER",
				Amount:  1,
			},
		},
	)
	game.pushAbility(instance)
	game.resolveTopEffectStack()
	if got := game.state.Objects[target].Counters["POWER"]; got != 0 {
		t.Fatalf("POWER counters = %d, want 0 after expired discard", got)
	}
	if got := game.state.Zones[player.UID]; !reflect.DeepEqual(got, beforeZones) {
		t.Fatalf("zones = %#v, want unchanged %#v", got, beforeZones)
	}
}

// TestAbilityChoiceUsesSpecifiedZone 驗證效果可從指定區域選牌，並由同一區域執行棄牌。
func TestAbilityChoiceUsesSpecifiedZone(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	zones := game.state.Zones[player.UID]
	card := zones.Hand[0]
	zones.Hand = removeCardAt(zones.Hand, 0)
	zones.Memory = append(zones.Memory, card)
	game.state.Zones[player.UID] = zones
	source := findCard(t, game, player, blazingThrowCardID)
	game.pushAbility(game.newAbilityInstance(
		player,
		source,
		"",
		[]effectOperation{
			{
				Kind: effectOperationChooseZoneCard,
				CardSelection: cardSelectionSpec{
					Zone: cardZoneMemory,
				},
			},
			{
				Kind:     effectOperationDiscard,
				CardZone: cardZoneMemory,
			},
		},
	))
	game.resolveTopEffectStack()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if view.PendingChoice == nil {
		t.Fatal("specified zone did not produce a pending choice")
	}
	selectPendingChoiceSubject(t, game, player, entityID(card))
	game.resolveTopEffectStack()
	zones = game.state.Zones[player.UID]
	if cardIndex(zones.Memory, card) >= 0 || cardIndex(zones.Graveyard, card) < 0 {
		t.Fatal("selected Memory card was not discarded to graveyard")
	}
}

func TestAbilityChoiceUsesVisibleHandleAndResumesTypedOperations(t *testing.T) {
	game := newActionGame(t)
	source := findCard(t, game, model.PlayerOne, blazingThrowCardID)
	target := objectID("champion:" + model.PlayerTwo.UID)
	instance := game.newAbilityInstance(
		model.PlayerOne,
		source,
		"",
		[]effectOperation{
			{
				Kind:    effectOperationChoose,
				Options: []objectID{target},
			},
			{
				Kind:   effectOperationDamage,
				Amount: 3,
			},
		},
	)
	game.pushAbility(instance)
	game.resolveTopEffectStack()
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if view.PendingChoice == nil || len(view.PendingChoice.Options) != 1 {
		t.Fatalf("PendingChoice = %#v, want one visible option", view.PendingChoice)
	}
	if otherView, err := game.PlayerView(model.PlayerTwo); err != nil || otherView.PendingChoice != nil {
		t.Fatalf("other player choice = %#v, %v; want nil, no error", otherView.PendingChoice, err)
	}
	if err := game.Submit(model.PlayerOne, Input{
		Revision: game.state.Revision,
		Choice:   view.PendingChoice.Options[0],
	}); err != nil {
		t.Fatalf("Submit() choice error = %v", err)
	}
	if len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Ability == nil {
		t.Fatalf("EffectsStack = %#v, want resumed Ability Instance", game.state.EffectsStack)
	}
	game.resolveTopEffectStack()
	if got := game.state.Champions[model.PlayerTwo.UID].Damage; got != 3 {
		t.Fatalf("target damage = %d, want 3", got)
	}
}

// TestAbilityDrawsInOnePrivateBatch 驗證能力抽多張牌會依主牌組頂端順序移至手牌，並保留單一私有 Event Batch。
// 輸入為含兩張已知牌組頂的能力；輸出為依序可見的手牌與兩個 draw Game Event；副作用為更新抽牌玩家的追蹤權及私有可見事件。
func TestAbilityDrawsInOnePrivateBatch(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	zones := game.state.Zones[player.UID]
	firstDraw := zones.MainDeck[len(zones.MainDeck)-1]
	secondDraw := zones.MainDeck[len(zones.MainDeck)-2]
	opponentView, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("opponent PlayerView() error = %v", err)
	}

	instance := game.newAbilityInstance(
		player,
		firstDraw,
		"",
		[]effectOperation{
			{
				Kind:   effectOperationDraw,
				Amount: 2,
			},
		},
	)
	game.pushAbility(instance)
	game.resolveTopEffectStack()

	zones = game.state.Zones[player.UID]
	if zones.Hand[len(zones.Hand)-2] != firstDraw || zones.Hand[len(zones.Hand)-1] != secondDraw {
		t.Fatalf("drawn hand order = %#v, want %q then %q", zones.Hand, firstDraw, secondDraw)
	}
	batch := game.state.Events[len(game.state.Events)-1]
	if batch.Cause != "ability" || len(batch.Events) != 2 {
		t.Fatalf("draw batch = %#v, want one ability batch with two events", batch)
	}
	if batch.Events[0].Kind != "draw" || batch.Events[1].Kind != "draw" {
		t.Fatalf("draw events = %#v, want two draw events", batch.Events)
	}
	opponentAfter, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("opponent PlayerView() after draw error = %v", err)
	}
	if len(opponentAfter.VisibleEvents) != len(opponentView.VisibleEvents) {
		t.Fatalf("opponent VisibleEvents = %#v, want no private draws", opponentAfter.VisibleEvents)
	}
}

// TestAbilityDrawsToMemoryInOnePrivateBatch 驗證能力抽至 Memory 會維持牌組頂端順序與私有事件批次。
// 輸入為含兩張已知牌組頂的抽至 Memory 能力；輸出為依序存入 Memory 的兩張牌與兩個 draw-to-memory Game Event；副作用為更新抽牌玩家的追蹤權及私有可見事件。
func TestAbilityDrawsToMemoryInOnePrivateBatch(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	zones := game.state.Zones[player.UID]
	firstDraw := zones.MainDeck[len(zones.MainDeck)-1]
	secondDraw := zones.MainDeck[len(zones.MainDeck)-2]
	opponentView, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("opponent PlayerView() error = %v", err)
	}

	instance := game.newAbilityInstance(
		player,
		firstDraw,
		"",
		[]effectOperation{
			{
				Kind:   effectOperationDrawToMemory,
				Amount: 2,
			},
		},
	)
	game.pushAbility(instance)
	game.resolveTopEffectStack()

	zones = game.state.Zones[player.UID]
	if zones.Memory[len(zones.Memory)-2] != firstDraw || zones.Memory[len(zones.Memory)-1] != secondDraw {
		t.Fatalf("drawn Memory order = %#v, want %q then %q", zones.Memory, firstDraw, secondDraw)
	}
	batch := game.state.Events[len(game.state.Events)-1]
	if batch.Cause != "ability" || len(batch.Events) != 2 {
		t.Fatalf("draw-to-memory batch = %#v, want one ability batch with two events", batch)
	}
	if batch.Events[0].Kind != "draw-to-memory" || batch.Events[1].Kind != "draw-to-memory" {
		t.Fatalf("draw-to-memory events = %#v, want two draw-to-memory events", batch.Events)
	}
	opponentAfter, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("opponent PlayerView() after Memory draw error = %v", err)
	}
	if len(opponentAfter.VisibleEvents) != len(opponentView.VisibleEvents) {
		t.Fatalf("opponent VisibleEvents = %#v, want no private Memory draws", opponentAfter.VisibleEvents)
	}
}

// TestTypedOperationsDrawAndCounterProduceStateChanges 驗證抽牌與 Counter 操作可在同一能力中更新對應遊戲狀態。
// 輸入為含抽牌與 Counter 操作的能力；輸出為多一張手牌與兩個 POWER counter；副作用為更新私有區域、追蹤權、事件及目標特徵。
func TestTypedOperationsDrawAndCounterProduceStateChanges(t *testing.T) {
	game := newActionGame(t)
	source := findCard(t, game, model.PlayerOne, blazingThrowCardID)
	target := objectID("counter-target")
	game.state.Objects[target] = fieldObject{
		ID:    target,
		Card:  source,
		Owner: model.PlayerOne,
	}
	beforeDraw := len(game.state.Zones[model.PlayerOne.UID].Hand)
	game.pushAbility(game.newAbilityInstance(
		model.PlayerOne,
		source,
		target,
		[]effectOperation{
			{
				Kind:   effectOperationDraw,
				Amount: 1,
			},
			{
				Kind:    effectOperationCounter,
				Counter: "POWER",
				Amount:  2,
			},
		},
	))
	game.resolveTopEffectStack()
	if got := len(game.state.Zones[model.PlayerOne.UID].Hand); got != beforeDraw+1 {
		t.Fatalf("hand size = %d, want %d", got, beforeDraw+1)
	}
	if got := game.state.Objects[target].Counters["POWER"]; got != 2 {
		t.Fatalf("POWER counters = %d, want 2", got)
	}
	if got := game.characteristicsFor(target).Power; got != game.state.Cards[source].Power+2 {
		t.Fatalf("derived power = %d, want printed power plus counters", got)
	}
}
