package game

import (
	"testing"

	"go-tcg/internal/model"
)

func TestGrandCrusadersRingBanishesThenDrawsThroughAbilityStack(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	card := findCard(t, game, player, grandCrusadersRingCardID)
	source := objectID("regalia:grand-crusaders-ring")
	game.state.Objects[source] = fieldObject{
		ID:    source,
		Card:  card,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	beforeHand := len(game.state.Zones[player.UID].Hand)
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := actionByCardName(t, view, "Grand Crusader's Ring")
	if action.AbilitySlot != "ability:2gv7DC0KID:front:banish-draw" {
		t.Fatalf("AbilitySlot = %q, want Grand Crusader's Ring slot", action.AbilitySlot)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   action.Handle,
		},
	); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if _, exists := game.state.Objects[source]; exists {
		t.Fatal("Grand Crusader's Ring remained on the field after paying banish cost")
	}
	if cardIndex(game.state.Zones[player.UID].Banishment, card) < 0 {
		t.Fatal("Grand Crusader's Ring was not banished")
	}
	if got := len(game.state.EffectsStack); got != 1 {
		t.Fatalf("EffectsStack = %d, want 1", got)
	}
	ability := game.state.EffectsStack[0].Ability
	if ability == nil || ability.Source != card || ability.SourceLKI != card {
		t.Fatalf("ability = %#v, want source and LKI %q", ability, card)
	}
	if !hasGameEvent(game, "banish") {
		t.Fatalf("events = %#v, want public banish event", game.state.Events)
	}
	passOpportunityRound(t, game, player)
	if got := len(game.state.Zones[player.UID].Hand); got != beforeHand+1 {
		t.Fatalf("hand size = %d, want %d after draw", got, beforeHand+1)
	}
	replay := game.Replay()
	if err := replay.Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestViridianProtectiveTrinketTaxesOnlyOpponentWaterActivationOnControllerTurn(t *testing.T) {
	game := newActionGame(t)
	controller := model.PlayerOne
	opponent := model.PlayerTwo
	viridianCard := findCard(t, game, controller, viridianProtectiveTrinketCardID)
	viridian := objectID("regalia:viridian")
	game.state.Objects[viridian] = fieldObject{
		ID:    viridian,
		Card:  viridianCard,
		Owner: controller,
		Types: []string{
			"ITEM",
		},
	}
	action := findCard(t, game, opponent, blazingThrowCardID)
	candidate := game.state.Cards[action]
	candidate.Elements = []string{
		"WATER",
	}
	game.state.Cards[action] = candidate
	baseCost := candidate.ReserveCost
	if got := game.actionReserveCost(opponent, action); got != baseCost+2 {
		t.Fatalf("opponent water activation cost = %d, want %d", got, baseCost+2)
	}
	if got := game.actionReserveCost(controller, action); got != baseCost {
		t.Fatalf("controller water activation cost = %d, want %d", got, baseCost)
	}
	game.state.Scheduler.TurnPlayer = opponent
	if got := game.actionReserveCost(opponent, action); got != baseCost {
		t.Fatalf("off-turn water activation cost = %d, want %d", got, baseCost)
	}
	game.state.Scheduler.TurnPlayer = controller
	delete(game.state.Objects, viridian)
	if got := game.actionReserveCost(opponent, action); got != baseCost {
		t.Fatalf("water activation cost after source left = %d, want %d", got, baseCost)
	}
}

func TestResonanceBaublesAreUnavailableWithoutMatchingOpponentChampion(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	for _, definition := range []CardID{
		waterResonanceBaubleCardID,
		windResonanceBaubleCardID,
	} {
		card := findCard(t, game, player, definition)
		source := objectID("regalia:" + string(definition))
		game.state.Objects[source] = fieldObject{
			ID:    source,
			Card:  card,
			Owner: player,
			Types: []string{
				"ITEM",
			},
		}
		if game.canActivateObjectAbility(player, source) {
			t.Fatalf("%q ability is legal without a matching opponent Champion", definition)
		}
	}
	game.advanceKnowledgeRevision()
	if got := len(game.getPlayerKnowledge(player).Abilities); got != 0 {
		t.Fatalf("ability handles = %#v, want none", game.getPlayerKnowledge(player).Abilities)
	}
}

// TestWindResonanceBaubleResolvesThroughPlayerView 驗證 Wind 條件成立時以穩定 Slot 完成放逐與抽牌。
// 輸入為具有 Wind Champion 的對手與場上 Bauble；輸出為抽牌結果，副作用為推進 Effects Stack 與 Banishment。
func TestWindResonanceBaubleResolvesThroughPlayerView(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	opponent := model.PlayerTwo
	champion := game.state.Champions[opponent.UID]
	championCard := game.state.Cards[champion.Card]
	championCard.Elements = []string{
		"WIND",
	}
	game.state.Cards[champion.Card] = championCard
	card := findCard(t, game, player, windResonanceBaubleCardID)
	source := objectID("regalia:wind-resonance-bauble")
	game.state.Objects[source] = fieldObject{
		ID:    source,
		Card:  card,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	beforeHand := len(game.state.Zones[player.UID].Hand)
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := actionByCardName(t, view, "Wind Resonance Bauble")
	if action.AbilitySlot != "ability:bHGUNMFLg9:front:banish-draw-wind-champion" {
		t.Fatalf("AbilitySlot = %q, want Wind Resonance Bauble slot", action.AbilitySlot)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   action.Handle,
		},
	); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	passOpportunityRound(t, game, player)
	if got := len(game.state.Zones[player.UID].Hand); got != beforeHand+1 {
		t.Fatalf("hand size = %d, want %d after draw", got, beforeHand+1)
	}
}

// TestObjectAbilityRejectsInvalidatedSourceWithoutFurtherMutation 驗證公開宣告後來源失效時不會支付費用或建立 Stack item。
// 輸入為已取得 handle 後離場的 Ring；輸出為拒絕錯誤，副作用僅包含測試預先移除的來源。
func TestObjectAbilityRejectsInvalidatedSourceWithoutFurtherMutation(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	card := findCard(t, game, player, grandCrusadersRingCardID)
	source := objectID("regalia:invalidated-ring")
	game.state.Objects[source] = fieldObject{
		ID:    source,
		Card:  card,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := actionByCardName(t, view, "Grand Crusader's Ring")
	delete(game.state.Objects, source)
	wantStateHash := game.StateHash()
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   action.Handle,
		},
	); err == nil {
		t.Fatal("Submit() error = nil, want invalidated source rejection")
	}
	if got := game.StateHash(); got != wantStateHash {
		t.Fatalf("StateHash() = %q, want unchanged %q", got, wantStateHash)
	}
}

// TestDuchessThornesRestCostRejectsRestedSource 驗證 rest cost 不可由已 rested 的來源支付。
// 輸入為場上 rested Duchess's Thornes；輸出為 Player View 不暴露能力，副作用為零。
func TestDuchessThornesRestCostRejectsRestedSource(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	card := findCard(t, game, player, duchessThornesCardID)
	source := objectID("regalia:rested-thornes")
	game.state.Objects[source] = fieldObject{
		ID:     source,
		Card:   card,
		Owner:  player,
		Rested: true,
		Types: []string{
			"ITEM",
		},
	}
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	for _, action := range view.LegalActions {
		if action.CardName == "The Duchess's Thornes" {
			t.Fatalf("rested Duchess's Thornes exposed illegal action %#v", action)
		}
	}
}

// TestDuchessThornesActivatedSlotUsesPlayerViewAndAbilityStack 驗證 rest、banish 與折扣依共用 Object 生命週期執行。
// 輸入為 awake Duchess's Thornes；輸出為穩定 Slot 與結算後折扣，副作用為放逐來源並推進 Effects Stack。
func TestDuchessThornesActivatedSlotUsesPlayerViewAndAbilityStack(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	card := findCard(t, game, player, duchessThornesCardID)
	source := objectID("regalia:awake-thornes")
	game.state.Objects[source] = fieldObject{
		ID:    source,
		Card:  card,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := actionByCardName(t, view, "The Duchess's Thornes")
	if action.AbilitySlot != "ability:bEXmm4rKOs:front:rest-banish-cardistry-discount" {
		t.Fatalf("AbilitySlot = %q, want Duchess's Thornes activated slot", action.AbilitySlot)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   action.Handle,
		},
	); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if _, exists := game.state.Objects[source]; exists {
		t.Fatal("Duchess's Thornes remained on field after paying costs")
	}
	if _, exists := game.state.CardistryDiscounts[player.UID]; exists {
		t.Fatal("Cardistry discount applied before ability resolution")
	}
	passOpportunityRound(t, game, player)
	if got := game.state.CardistryDiscounts[player.UID]; got != 6 {
		t.Fatalf("Cardistry discount = %d, want 6", got)
	}
}

// TestSmokeBombsIsUnavailableWithoutAnAlly 驗證 Smoke Bombs 的必要 Ally 選擇會在宣告前檢查。
// 輸入為只有 Smoke Bombs、沒有任何 Ally 的對局；輸出為沒有合法 ability，副作用為零。
func TestSmokeBombsIsUnavailableWithoutAnAlly(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	card := findCard(t, game, player, smokeBombsCardID)
	source := objectID("regalia:smoke-without-ally")
	game.state.Objects[source] = fieldObject{
		ID:    source,
		Card:  card,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	if game.canActivateObjectAbility(player, source) {
		t.Fatal("Smoke Bombs is legal without an Ally target")
	}
	game.advanceKnowledgeRevision()
	if got := len(game.getPlayerKnowledge(player).Abilities); got != 0 {
		t.Fatalf("ability handles = %#v, want none", game.getPlayerKnowledge(player).Abilities)
	}
}

// TestWaterResonanceBaubleUsesDefinitionConditionAndResolvesThroughPlayerView 驗證 Water Bauble 的元素條件、放逐與抽牌完全經由 Object Definition。
// 輸入為具有 Water Champion 的對手與場上的 Water Bauble；輸出為穩定 slot 的合法行動與抽牌結果，副作用為對局狀態推進。
func TestWaterResonanceBaubleUsesDefinitionConditionAndResolvesThroughPlayerView(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	opponent := model.PlayerTwo
	champion := game.state.Champions[opponent.UID]
	championCard := game.state.Cards[champion.Card]
	championCard.Elements = []string{
		"WATER",
	}
	game.state.Cards[champion.Card] = championCard
	card := findCard(t, game, player, waterResonanceBaubleCardID)
	source := objectID("regalia:water-resonance-bauble")
	game.state.Objects[source] = fieldObject{
		ID:    source,
		Card:  card,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	beforeHand := len(game.state.Zones[player.UID].Hand)
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := actionByCardName(t, view, "Water Resonance Bauble")
	if action.AbilitySlot != "ability:dSSRtNnPtw:front:banish-draw-water-champion" {
		t.Fatalf("AbilitySlot = %q, want Water Resonance Bauble slot", action.AbilitySlot)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   action.Handle,
		},
	); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if _, exists := game.state.Objects[source]; exists {
		t.Fatal("Water Resonance Bauble remained on the field after its banish cost")
	}
	if cardIndex(game.state.Zones[player.UID].Banishment, card) < 0 {
		t.Fatal("Water Resonance Bauble was not banished")
	}
	passOpportunityRound(t, game, player)
	if got := len(game.state.Zones[player.UID].Hand); got != beforeHand+1 {
		t.Fatalf("hand size = %d, want %d after draw", got, beforeHand+1)
	}
}
