package game

import (
	"encoding/json"
	"testing"

	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

// TestBanishCardsFromGraveyardDoesNotDependOnSelection 驗證墓地放逐只依輸入卡牌付款，且錯誤不改動原區域。
func TestBanishCardsFromGraveyardDoesNotDependOnSelection(t *testing.T) {
	zones := playerZones{
		Graveyard: []cardInstanceID{
			"first",
			"second",
		},
	}
	if _, err := banishCardsFromGraveyard(
		zones,
		[]cardInstanceID{
			"first",
			"first",
		},
	); err == nil {
		t.Fatal("duplicate banishment was accepted")
	}
	updated, err := banishCardsFromGraveyard(
		zones,
		[]cardInstanceID{
			"first",
		},
	)
	if err != nil {
		t.Fatalf("banishCardsFromGraveyard() error = %v", err)
	}
	if len(zones.Graveyard) != 2 || len(zones.Banishment) != 0 {
		t.Fatalf("original zones changed: %#v", zones)
	}
	if len(updated.Graveyard) != 1 || updated.Graveyard[0] != "second" || len(updated.Banishment) != 1 || updated.Banishment[0] != "first" {
		t.Fatalf("updated zones = %#v", updated)
	}
}

// TestCommitObjectAbilityBanishmentRejectsForeignCard 驗證場上來源的卡牌不屬於控制者時，放逐提交不會留下部分狀態變更。
// 輸入為 Player One 控制但 Card Instance 屬於 Player Two 的 Object；輸出為錯誤，副作用為零且 StateHash 保持不變。
func TestCommitObjectAbilityBanishmentRejectsForeignCard(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	card := findCard(t, game, player, grandCrusadersRingCardID)
	source := objectID("regalia:foreign-card")
	game.state.Objects[source] = fieldObject{
		ID:    source,
		Card:  card,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	instance := game.state.Cards[card]
	instance.Owner = model.PlayerTwo
	game.state.Cards[card] = instance
	wantStateHash := game.StateHash()
	if _, err := game.commitObjectAbilityBanishment(player, source); err == nil {
		t.Fatal("commitObjectAbilityBanishment() error = nil, want foreign card rejection")
	}
	if got := game.StateHash(); got != wantStateHash {
		t.Fatalf("StateHash() = %q, want unchanged %q", got, wantStateHash)
	}
}

// TestCommitObjectSacrificeDispatchesSacrificeTrigger 驗證犧牲提交會以獨立 sacrifice event 派發已編譯 trigger，而不誤用 destroy 原因。
// 輸入為帶有 sacrifice trigger 的受控 Ally；輸出為 sacrifice batch 與 LKI trigger stack 項目；副作用為移除 Object、移動卡牌並建立觸發。
func TestCommitObjectSacrificeDispatchesSacrificeTrigger(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	card := findCard(t, game, player, twoOfHeartsCardID)
	definitionID := game.state.Cards[card].Definition
	definition := game.definitions[definitionID]
	definition.abilities = append(
		definition.abilities,
		compiledAbilityDefinition{
			slot: "ability:e8ygl32jef:front:on-sacrifice",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event: eventKindSacrifice,
			},
		},
	)
	game.definitions[definitionID] = definition
	object := objectID("ally:sacrifice-trigger")
	game.state.Objects[object] = fieldObject{
		ID:    object,
		Card:  card,
		Owner: player,
		Types: []string{
			"ALLY",
		},
	}
	_, err := game.commitObjectSacrifice(
		objectSacrificeRequest{
			Controller:   player,
			Object:       object,
			RequiredType: "ALLY",
			Cause:        "ability:sacrifice",
		},
	)
	if err != nil {
		t.Fatalf("commitObjectSacrifice() error = %v", err)
	}
	batch := game.state.Events[len(game.state.Events)-1]
	if batch.Cause != "ability:sacrifice" || len(batch.Events) != 1 || batch.Events[0].Kind != string(eventKindSacrifice) {
		t.Fatalf("sacrifice batch = %#v, want independent sacrifice event", batch)
	}
	if len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Ability == nil || game.state.EffectsStack[0].Ability.Slot != "ability:e8ygl32jef:front:on-sacrifice" || game.state.EffectsStack[0].SourceLKI != card {
		t.Fatalf("EffectsStack = %#v, want sacrifice trigger with source LKI", game.state.EffectsStack)
	}
}

// TestCommitPlayerCardMovesMovesOwnedCardAndRejectsExpiredSource 驗證玩家卡牌移動保留同一張 Card Instance、記錄公開事件，且過期來源被拒絕時不改變權威狀態。
// 輸入為正式單局中的手牌與後續失效移動請求；輸出為移至墓地的同一張牌與錯誤；副作用為成功時更新區域、追蹤權及事件。
func TestCommitPlayerCardMovesMovesOwnedCardAndRejectsExpiredSource(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	card := game.state.Zones[player.UID].Hand[0]

	err := game.commitPlayerCardMoves(
		playerCardMoveRequest{
			Player: player,
			Moves: []playerCardMove{
				{
					Card: card,
					From: cardZoneHand,
					To:   cardZoneGraveyard,
				},
			},
			Cause:        "ability",
			EventKind:    "discard",
			VisibleToAll: true,
		},
	)
	if err != nil {
		t.Fatalf("commitPlayerCardMoves() error = %v", err)
	}
	zones := game.state.Zones[player.UID]
	if cardIndex(zones.Hand, card) >= 0 || cardIndex(zones.Graveyard, card) < 0 {
		t.Fatalf("zones = %#v, want card %q moved from hand to graveyard", zones, card)
	}
	if game.state.Cards[card].Owner != player {
		t.Fatalf("card owner = %q, want %q", game.state.Cards[card].Owner, player)
	}
	opponentView, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() opponent error = %v", err)
	}
	cardName := game.state.Entities[entityID(card)].Name
	if !hasVisibleEvent(opponentView.VisibleEvents, "discard", cardName) {
		t.Fatalf("opponent visible events = %#v, want discard for %q", opponentView.VisibleEvents, cardName)
	}
	beforeState, err := json.Marshal(game.state)
	if err != nil {
		t.Fatalf("marshal state before expired move: %v", err)
	}
	beforeHash := game.StateHash()
	err = game.commitPlayerCardMoves(
		playerCardMoveRequest{
			Player: player,
			Moves: []playerCardMove{
				{
					Card: card,
					From: cardZoneHand,
					To:   cardZoneGraveyard,
				},
			},
			Cause:        "ability",
			EventKind:    "discard",
			VisibleToAll: true,
		},
	)
	if err == nil {
		t.Fatal("commitPlayerCardMoves() accepted expired source")
	}
	afterState, err := json.Marshal(game.state)
	if err != nil {
		t.Fatalf("marshal state after expired move: %v", err)
	}
	if game.StateHash() != beforeHash || string(afterState) != string(beforeState) {
		t.Fatal("expired move changed authoritative state")
	}
}

// TestRecollectionMovesMemoryToHandInOrderAndKeepsKnowledgePrivate 驗證正式回合流程會按 Memory 順序回收至 Hand，保留 owner 追蹤且不向對手洩漏私有事件。
// 輸入為停在 Recollection Phase 的正式單局與兩張 Memory；輸出為回收後的 Player View 與可驗證 replay；副作用為透過 Submit 推進回合、移動區域及記錄私有事件。
func TestRecollectionMovesMemoryToHandInOrderAndKeepsKnowledgePrivate(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	zones := game.state.Zones[player.UID]
	first := zones.Hand[0]
	second := zones.Hand[1]
	zones.Hand = removeCardAt(zones.Hand, 1)
	zones.Hand = removeCardAt(zones.Hand, 0)
	zones.Memory = append(
		zones.Memory,
		first,
		second,
	)
	game.state.Zones[player.UID] = zones
	game.state.Scheduler = schedulerFrame{
		Kind:              schedulerStable,
		TurnPlayer:        player,
		OpportunityHolder: player,
		Phase:             constants.PhaseRecollection,
		TurnNumber:        1,
	}
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()

	submitActionKind(t, game, player, constants.ActionPass)
	submitActionKind(t, game, model.PlayerTwo, constants.ActionPass)

	zones = game.state.Zones[player.UID]
	if len(zones.Memory) != 0 {
		t.Fatalf("Memory = %#v, want empty after recollection", zones.Memory)
	}
	if len(zones.Hand) < 2 || zones.Hand[len(zones.Hand)-2] != first || zones.Hand[len(zones.Hand)-1] != second {
		t.Fatalf("Hand = %#v, want recollection order ending in %q, %q", zones.Hand, first, second)
	}
	ownerView, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() owner error = %v", err)
	}
	firstEntity := entityID(first)
	firstName := game.state.Entities[firstEntity].Name
	if !hasVisibleEvent(ownerView.VisibleEvents, "recollect", firstName) {
		t.Fatalf("owner visible events = %#v, want recollect for %q", ownerView.VisibleEvents, first)
	}
	handCapacity := len(ownerView.Hand)
	ownerHandles := make(map[ViewHandle]bool, handCapacity)
	for _, card := range ownerView.Hand {
		ownerHandles[card.Handle] = true
	}
	ownerKnowledge := game.getPlayerKnowledge(player)
	firstHandle := ownerKnowledge.Cards[firstEntity]
	secondEntity := entityID(second)
	secondHandle := ownerKnowledge.Cards[secondEntity]
	if !ownerHandles[firstHandle] || !ownerHandles[secondHandle] {
		t.Fatalf("owner Hand = %#v, want handles for recollected cards", ownerView.Hand)
	}
	opponentView, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() opponent error = %v", err)
	}
	if hasVisibleEvent(opponentView.VisibleEvents, "recollect", firstName) {
		t.Fatalf("opponent visible events leaked recollect: %#v", opponentView.VisibleEvents)
	}
	replay := game.Replay()
	if err := replay.Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}
