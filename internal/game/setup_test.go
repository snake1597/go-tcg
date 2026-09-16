package game

import (
	"encoding/json"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	"path/filepath"
	"testing"
)

// materializationActionByCardName 從玩家視圖取得指定卡名的 materialize action，避免測試依賴 handle 順序。
// 輸入為測試、PlayerView 與卡名；輸出為唯一同名 materialize action，找不到時使測試失敗且不改變遊戲狀態。
func materializationActionByCardName(t *testing.T, view PlayerView, name string) LegalAction {
	t.Helper()
	for _, action := range view.LegalActions {
		if action.Kind == constants.ActionMaterialize && action.CardName == name {
			return action
		}
	}
	t.Fatalf("materialization action %q not found in %#v", name, view.LegalActions)
	return LegalAction{}
}

// Rules: 602c917f2f8fd4df7198429a72eb596bf7f647c6,
// general-rules-starting-the-game.md § Standard Game Setup;
// turn-order-main-phase.md § General Rules.
func TestStandardSetupStartsFirstTurnAtMainAndPassesToSecondPlayersDraw(t *testing.T) {
	configuration := StandardGameConfig{
		Players: [2]*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: filepath.Clean("../.."),
		Seed:           42,
	}
	game, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
	}

	assertTurnView(
		t,
		game,
		model.PlayerOne,
		model.PlayerOne,
		PhaseMain,
		model.PlayerOne,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)
	assertTurnView(
		t,
		game,
		model.PlayerTwo,
		model.PlayerOne,
		PhaseMain,
		model.PlayerOne,
		[]constants.ActionKind{
			constants.ActionConcede,
		},
	)

	submitActionKind(
		t,
		game,
		model.PlayerOne,
		constants.ActionPass,
	)
	assertTurnView(
		t,
		game,
		model.PlayerTwo,
		model.PlayerOne,
		PhaseMain,
		model.PlayerTwo,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)

	submitActionKind(
		t,
		game,
		model.PlayerTwo,
		constants.ActionPass,
	)
	assertTurnView(
		t,
		game,
		model.PlayerOne,
		model.PlayerOne,
		PhaseEnd,
		model.PlayerOne,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)

	submitActionKind(
		t,
		game,
		model.PlayerOne,
		constants.ActionPass,
	)
	submitActionKind(
		t,
		game,
		model.PlayerTwo,
		constants.ActionPass,
	)
	secondView, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if len(secondView.VisibleEvents) != 8 {
		t.Fatalf("PlayerView().VisibleEvents = %#v, want seven opening draws and one turn draw", secondView.VisibleEvents)
	}
	assertTurnView(
		t,
		game,
		model.PlayerTwo,
		model.PlayerTwo,
		PhaseMain,
		model.PlayerTwo,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)
}

// TestStandardGameCompletesTurnsAndReplaysWakeUp 驗證完整回合流程會喚醒曾攻擊的 Ally，且 canonical replay 可重播該狀態。
// 輸入為固定 seed 的 Standard 單局、雙方各一張 Ally 與 Player Two 的一次攻擊；輸出為 wake-up event、醒著的 Ally 與可驗證 replay，副作用為依序提交三個完整回合的公開操作。
func TestStandardGameCompletesTurnsAndReplaysWakeUp(t *testing.T) {
	game, err := NewStandardGame(
		StandardGameConfig{
			Players: [2]*model.Player{
				model.PlayerOne,
				model.PlayerTwo,
			},
			RepositoryRoot: filepath.Clean("../.."),
			Seed:           42,
		},
	)
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
	}

	playTargetedActionAndResolve(
		t,
		game,
		model.PlayerOne,
		"Straight Flare",
	)
	completeMainAndEnd(
		t,
		game,
		model.PlayerOne,
	)
	secondView, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if secondView.TurnNumber != 2 || secondView.Phase != PhaseMain {
		t.Fatalf(
			"Player Two turn = %d phase %q, want turn 2 main after wake-up and draw",
			secondView.TurnNumber,
			secondView.Phase,
		)
	}
	drawEventCount := 0
	for _, event := range secondView.VisibleEvents {
		if event.Kind == "draw" {
			drawEventCount++
		}
	}
	if drawEventCount != 8 {
		t.Fatalf(
			"Player Two draw events = %d, want seven opening draws and one turn draw before main",
			drawEventCount,
		)
	}

	playAllyAndResolve(
		t,
		game,
		model.PlayerTwo,
		"Noire, Ace of Spades",
	)
	attackWithOnlyLegalAttacker(
		t,
		game,
		model.PlayerTwo,
	)
	secondView, err = game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() after attack error = %v", err)
	}
	attacker := fieldObjectByOwnerAndName(
		t,
		secondView,
		model.PlayerTwo,
		"Noire, Ace of Spades",
	)
	if !attacker.Rested {
		t.Fatalf("Noire rested = %t, want true after attack", attacker.Rested)
	}
	completeMainAndEnd(
		t,
		game,
		model.PlayerTwo,
	)
	assertTurnView(
		t,
		game,
		model.PlayerOne,
		model.PlayerOne,
		PhaseMaterialize,
		nil,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionMaterialize,
			constants.ActionSkipMaterialize,
		},
	)
	submitActionKind(
		t,
		game,
		model.PlayerOne,
		constants.ActionSkipMaterialize,
	)
	passOpportunityRound(
		t,
		game,
		model.PlayerOne,
	)
	completeMainAndEnd(
		t,
		game,
		model.PlayerOne,
	)

	wakeView, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() after wake-up error = %v", err)
	}
	woken := fieldObjectByOwnerAndName(
		t,
		wakeView,
		model.PlayerTwo,
		"Noire, Ace of Spades",
	)
	if woken.Rested {
		t.Fatalf("Noire rested = %t, want false after wake-up", woken.Rested)
	}
	if !hasVisibleEvent(wakeView.VisibleEvents, "wake-up", "Noire, Ace of Spades") {
		t.Fatalf(
			"PlayerView().VisibleEvents = %#v, want Noire wake-up",
			wakeView.VisibleEvents,
		)
	}
	replay := game.Replay()
	if len(replay.Steps) == 0 {
		t.Fatal("Replay().Steps is empty after completed turns")
	}
	lastStep := replay.Steps[len(replay.Steps)-1]
	if lastStep.StateHash != game.StateHash() {
		t.Fatalf(
			"final replay state hash = %q, want wake-up state hash %q",
			lastStep.StateHash,
			game.StateHash(),
		)
	}
	if err := replay.Verify(); err != nil {
		t.Fatalf("Replay().Verify() after wake-up error = %v", err)
	}
}

func playTargetedActionAndResolve(
	t *testing.T,
	game *Game,
	player *model.Player,
	cardName string,
) {
	t.Helper()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := actionByCardName(
		t,
		view,
		cardName,
	)
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   action.Handle,
			Reserve:  reserveHandles(action),
		},
	); err != nil {
		t.Fatalf("Submit() activation error = %v", err)
	}
	view, err = game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() target error = %v", err)
	}
	if view.PendingChoice == nil || len(view.PendingChoice.Options) == 0 {
		t.Fatalf(
			"PlayerView().PendingChoice = %#v, want at least one target",
			view.PendingChoice,
		)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Choice:   view.PendingChoice.Options[0],
		},
	); err != nil {
		t.Fatalf("Submit() target error = %v", err)
	}
	passOpportunityRound(
		t,
		game,
		player,
	)
	assertTurnView(
		t,
		game,
		player,
		player,
		PhaseMain,
		player,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)
}

func playAllyAndResolve(
	t *testing.T,
	game *Game,
	player *model.Player,
	cardName string,
) {
	t.Helper()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := actionByCardName(
		t,
		view,
		cardName,
	)
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   action.Handle,
			Reserve:  reserveHandles(action),
		},
	); err != nil {
		t.Fatalf("Submit() activation error = %v", err)
	}
	passOpportunityRound(
		t,
		game,
		player,
	)
	assertTurnView(
		t,
		game,
		player,
		player,
		PhaseMain,
		player,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)
}

func attackWithOnlyLegalAttacker(
	t *testing.T,
	game *Game,
	player *model.Player,
) {
	t.Helper()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() before attack error = %v", err)
	}
	attack := actionByKind(
		t,
		view,
		constants.ActionAttack,
	)
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   attack.Handle,
		},
	); err != nil {
		t.Fatalf("Submit() attack error = %v", err)
	}
	view, err = game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() attack target error = %v", err)
	}
	if view.PendingChoice == nil || len(view.PendingChoice.Options) != 1 {
		t.Fatalf(
			"PlayerView().PendingChoice = %#v, want one attack target",
			view.PendingChoice,
		)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Choice:   view.PendingChoice.Options[0],
		},
	); err != nil {
		t.Fatalf("Submit() attack target error = %v", err)
	}
	passOpportunityRound(
		t,
		game,
		player,
	)
}

func completeMainAndEnd(
	t *testing.T,
	game *Game,
	player *model.Player,
) {
	t.Helper()
	assertTurnView(
		t,
		game,
		player,
		player,
		PhaseMain,
		player,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)
	passOpportunityRound(
		t,
		game,
		player,
	)
	assertTurnView(
		t,
		game,
		player,
		player,
		PhaseEnd,
		player,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)
	passOpportunityRound(
		t,
		game,
		player,
	)
}

func fieldObjectByOwnerAndName(
	t *testing.T,
	view PlayerView,
	owner *model.Player,
	name string,
) VisibleFieldObject {
	t.Helper()
	for _, object := range view.Field {
		if samePlayer(object.Owner, owner) && object.CardName == name {
			return object
		}
	}
	t.Fatalf(
		"PlayerView().Field = %#v, want %s controlled by %q",
		view.Field,
		name,
		owner,
	)
	return VisibleFieldObject{}
}

// TestWakeUpPhaseWakesAllControlledRestedObjects 驗證換回合時 scheduler 同時喚醒回合玩家的 Champion 與 Ally。
// 輸入為 End Phase 的 Standard 單局及兩次 pass；輸出為醒著的受控 objects 與單一 simultaneous event batch，副作用為推進至下一位玩家的 Main Phase。
func TestWakeUpPhaseWakesAllControlledRestedObjects(t *testing.T) {
	game, err := NewStandardGame(
		StandardGameConfig{
			Players: [2]*model.Player{
				model.PlayerOne,
				model.PlayerTwo,
			},
			RepositoryRoot: filepath.Clean("../.."),
			Seed:           42,
		},
	)
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
	}
	zones := game.state.Zones[model.PlayerTwo.UID]
	allyCard := zones.MainDeck[0]
	zones.MainDeck = removeCardAt(zones.MainDeck, 0)
	game.state.Zones[model.PlayerTwo.UID] = zones
	ally := objectID("ally:wake-up-test")
	game.state.Objects[ally] = fieldObject{
		ID:    ally,
		Card:  allyCard,
		Owner: model.PlayerTwo,
		Types: []string{
			"ALLY",
		},
		Rested: true,
	}
	champion := game.state.Champions[model.PlayerTwo.UID]
	champion.Rested = true
	game.state.Champions[model.PlayerTwo.UID] = champion
	game.state.Scheduler.Phase = PhaseEnd
	game.state.Scheduler.OpportunityHolder = model.PlayerOne
	game.advanceKnowledgeRevision()

	passOpportunityRound(t, game, model.PlayerOne)

	if game.state.Champions[model.PlayerTwo.UID].Rested || game.state.Objects[ally].Rested {
		t.Fatalf("wake state = champion rested %t, ally rested %t; want both awake", game.state.Champions[model.PlayerTwo.UID].Rested, game.state.Objects[ally].Rested)
	}
	for _, batch := range game.state.Events {
		if batch.Cause != "turn:wake-up" {
			continue
		}
		if !batch.Simultaneous || len(batch.Events) != 2 {
			t.Fatalf("wake batch = %#v, want two simultaneous events", batch)
		}
		return
	}
	t.Fatal("wake-up event batch not found")
}

// Rules: 602c917f2f8fd4df7198429a72eb596bf7f647c6,
// game-mechanics-timing-and-permissions.md § Opportunity;
// turn-order-recollection-phase.md § General Rules.
func TestStandardPassesDeterministicallyReachRecollectionOnTheNextTurn(t *testing.T) {
	configuration := StandardGameConfig{
		Players: [2]*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: filepath.Clean("../.."),
		Seed:           42,
	}
	first, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("first NewStandardGame() error = %v", err)
	}
	second, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("second NewStandardGame() error = %v", err)
	}

	for step := 0; step < 8; step++ {
		firstView, err := first.PlayerView(model.PlayerOne)
		if err != nil {
			t.Fatalf("first PlayerView() error = %v", err)
		}
		submitCurrentTurnAction(t, first, firstView)
		secondView, err := second.PlayerView(model.PlayerOne)
		if err != nil {
			t.Fatalf("second PlayerView() error = %v", err)
		}
		submitCurrentTurnAction(t, second, secondView)
		if first.StateHash() != second.StateHash() {
			t.Fatalf("state hashes differ after pass %d: %q != %q", step+1, first.StateHash(), second.StateHash())
		}
	}

	assertTurnView(
		t,
		first,
		model.PlayerOne,
		model.PlayerOne,
		PhaseMaterialize,
		nil,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionMaterialize,
			constants.ActionSkipMaterialize,
		},
	)
	materializeView, err := first.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	submitCurrentTurnAction(t, first, materializeView)
	assertTurnView(
		t,
		first,
		model.PlayerOne,
		model.PlayerOne,
		PhaseRecollection,
		model.PlayerOne,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)
}

// Rules: 602c917f2f8fd4df7198429a72eb596bf7f647c6,
// turn-order-materialize-phase.md § General Rules.
func TestStandardTurnStopsAtMaterializeUntilTurnPlayerSkipsIt(t *testing.T) {
	configuration := StandardGameConfig{
		Players: [2]*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: filepath.Clean("../.."),
		Seed:           42,
	}
	game, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
	}
	for step := 0; step < 15; step++ {
		view, err := game.PlayerView(model.PlayerOne)
		if err != nil {
			t.Fatalf("PlayerView() error = %v", err)
		}
		submitCurrentTurnAction(t, game, view)
	}

	assertTurnView(
		t,
		game,
		model.PlayerTwo,
		model.PlayerTwo,
		PhaseMaterialize,
		nil,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionMaterialize,
			constants.ActionSkipMaterialize,
		},
	)
	submitActionKind(
		t,
		game,
		model.PlayerTwo,
		constants.ActionSkipMaterialize,
	)
	assertTurnView(
		t,
		game,
		model.PlayerTwo,
		model.PlayerTwo,
		PhaseRecollection,
		model.PlayerTwo,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)
}

// Rules: 602c917f2f8fd4df7198429a72eb596bf7f647c6,
// card-types-champion.md § Leveling Champions;
// playing-cards-card-materialization.md § Materialization.
func TestMaterializingTonorisLevelsUpChampionAndGrantsTaunt(t *testing.T) {
	game := newTonorisMaterializationGame(t)
	player := model.PlayerTwo
	championBefore := game.state.Champions[player.UID]
	championBefore.Rested = true
	championBefore.Counters = map[string]int{
		"enlighten": 2,
	}
	championBefore.CombatRole = "attacker"
	game.state.Champions[player.UID] = championBefore
	game.captureReplayInitialState()

	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	materialize := materializationActionByCardName(t, view, "Tonoris, Lone Mercenary")
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   materialize.Handle,
		},
	); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if len(game.state.EffectsStack) != 1 {
		t.Fatalf("EffectsStack = %#v, want one materialization", game.state.EffectsStack)
	}
	if len(game.state.EffectSources) != 1 {
		t.Fatalf("EffectSources = %#v, want one source card", game.state.EffectSources)
	}
	paymentView, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() after materialization payment error = %v", err)
	}
	hasBanishMemoryEvent := false
	for _, event := range paymentView.VisibleEvents {
		if event.Kind == "materialize-banish-memory" && event.CardName != "" {
			hasBanishMemoryEvent = true
			break
		}
	}
	if !hasBanishMemoryEvent {
		t.Fatalf("PlayerView().VisibleEvents = %#v, want named random materialization payment", paymentView.VisibleEvents)
	}

	passOpportunityRound(t, game, player)
	championAfterLevelUp := game.state.Champions[player.UID]
	if championAfterLevelUp.ID != championBefore.ID {
		t.Fatalf("Champion ID = %q, want preserved ID %q", championAfterLevelUp.ID, championBefore.ID)
	}
	if championAfterLevelUp.Card == championBefore.Card {
		t.Fatal("Champion top card did not change after Level Up")
	}
	if len(championAfterLevelUp.InnerLineage) != 1 || championAfterLevelUp.InnerLineage[0] != championBefore.Card {
		t.Fatalf("InnerLineage = %#v, want original top card %q", championAfterLevelUp.InnerLineage, championBefore.Card)
	}
	if !championAfterLevelUp.Rested || championAfterLevelUp.Counters["enlighten"] != 2 || championAfterLevelUp.CombatRole != "attacker" {
		t.Fatalf("Champion runtime state reset after Level Up: %#v", championAfterLevelUp)
	}

	passOpportunityRound(t, game, player)
	view, err = game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() after taunt error = %v", err)
	}
	champion := championByOwner(
		t,
		view,
		player,
	)
	if !champion.Taunt {
		t.Fatalf("PlayerView().Champions = %#v, want Tonoris taunt", view.Champions)
	}
	if !hasVisibleEvent(view.VisibleEvents, "level-up", "Tonoris, Lone Mercenary") || !hasVisibleEvent(view.VisibleEvents, "taunt-granted", "Tonoris, Lone Mercenary") {
		t.Fatalf("PlayerView().VisibleEvents = %#v, want level-up and taunt-granted", view.VisibleEvents)
	}
	replayData, err := json.Marshal(game.Replay())
	if err != nil {
		t.Fatalf("marshal Replay() error = %v", err)
	}
	var replay Replay
	if err := json.Unmarshal(
		replayData,
		&replay,
	); err != nil {
		t.Fatalf("unmarshal Replay() error = %v", err)
	}
	if err := replay.Verify(); err != nil {
		t.Fatalf("serialized Replay().Verify() error = %v", err)
	}
	for game.state.Scheduler.TurnPlayer != player || game.state.Scheduler.Phase != PhaseMaterialize {
		currentView, err := game.PlayerView(game.state.Scheduler.TurnPlayer)
		if err != nil {
			t.Fatalf("PlayerView() advancing taunt duration error = %v", err)
		}
		submitCurrentTurnAction(
			t,
			game,
			currentView,
		)
	}
	view, err = game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() after taunt expiry error = %v", err)
	}
	champion = championByOwner(
		t,
		view,
		player,
	)
	if champion.Taunt {
		t.Fatalf("PlayerView().Champions = %#v, want expired Tonoris taunt", view.Champions)
	}
	if !hasVisibleEvent(view.VisibleEvents, "taunt-expired", "Tonoris, Lone Mercenary") {
		t.Fatalf("PlayerView().VisibleEvents = %#v, want taunt-expired", view.VisibleEvents)
	}
}

// TestMaterializationExposesEveryEligibleFixedMaterialDeckCard 驗證固定 Material Deck 的所有非起始卡都可透過 PlayerView materialize。
// 輸入為進入第二位玩家 Materialize Phase 的正式 Standard 單局；輸出為 Tonoris 與十張 Regalia 的 materialize actions，副作用僅為建立隔離測試單局與付款用 Memory。
func TestMaterializationExposesEveryEligibleFixedMaterialDeckCard(t *testing.T) {
	game := newTonorisMaterializationGame(t)
	view, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	actualNames := make(map[string]bool)
	for _, action := range view.LegalActions {
		if action.Kind == constants.ActionMaterialize {
			actualNames[action.CardName] = true
		}
	}
	wantNames := []string{
		"Tonoris, Lone Mercenary",
		"Bulwark Sword",
		"Grand Crusader's Ring",
		"Safeguard Amulet",
		"Smoke Bombs",
		"Viridian Protective Trinket",
		"Water Resonance Bauble",
		"Wind Resonance Bauble",
		"Impact Hammer",
		"Infernal Vessel",
		"The Duchess's Thornes",
	}
	for _, wantName := range wantNames {
		if !actualNames[wantName] {
			t.Fatalf("PlayerView().LegalActions = %#v, missing materialize action %q", view.LegalActions, wantName)
		}
	}
}

// TestMaterializingHinderedRegaliaEntersRested 驗證 Hindered Regalia 經正式 materialization Stack 結算後才進場，且進場即 rested。
// 輸入為 The Duchess's Thornes 的 PlayerView materialize action 與雙方 pass；輸出為場上的 rested Regalia object，副作用為移除 Material Deck 來源、建立 Effects Stack item、記錄公開事件與 replay。
func TestMaterializingHinderedRegaliaEntersRested(t *testing.T) {
	game := newTonorisMaterializationGame(t)
	player := model.PlayerTwo
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	materialize := materializationActionByCardName(t, view, "The Duchess's Thornes")
	source := game.state.Knowledge.Materializations[player.UID][materialize.Handle]
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   materialize.Handle,
		},
	); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Source != source {
		t.Fatalf("EffectsStack = %#v, want Duchess's Thornes materialization", game.state.EffectsStack)
	}
	passOpportunityRound(t, game, player)
	objectID, exists := objectIDForCard(game, source)
	if !exists {
		t.Fatalf("Objects = %#v, want materialized Duchess's Thornes", game.state.Objects)
	}
	object := game.state.Objects[objectID]
	if !object.Rested || !containsString(object.Types, "REGALIA") || !containsString(object.Types, "ITEM") {
		t.Fatalf("materialized object = %#v, want rested Regalia Item", object)
	}
	if containsCard(game.state.EffectSources, source) {
		t.Fatalf("EffectSources = %#v, want resolved source removed", game.state.EffectSources)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestMaterializingTonorisRejectsInsufficientPaymentWithoutChangingState(t *testing.T) {
	game := newTonorisMaterializationGame(t)
	player := model.PlayerTwo
	zones := game.state.Zones[player.UID]
	zones.Memory = nil
	game.state.Zones[player.UID] = zones
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	before := game.StateHash()
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   ViewHandle("not-a-materialization-option"),
		},
	); err == nil {
		t.Fatal("Submit() insufficient payment error = nil")
	}
	if got := game.StateHash(); got != before {
		t.Fatalf("StateHash() after insufficient payment = %q, want unchanged %q", got, before)
	}
}

func TestMaterializingTonorisRejectsIllegalLineageWithoutChangingState(t *testing.T) {
	game := newTonorisMaterializationGame(t)
	player := model.PlayerTwo
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	materialize := materializationActionByCardName(t, view, "Tonoris, Lone Mercenary")
	champion := game.state.Champions[player.UID]
	card := game.state.Cards[champion.Card]
	card.Definition = tonorisCardID
	game.state.Cards[champion.Card] = card
	before := game.StateHash()
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   materialize.Handle,
		},
	); err == nil {
		t.Fatal("Submit() illegal lineage error = nil")
	}
	if got := game.StateHash(); got != before {
		t.Fatalf("StateHash() after illegal lineage = %q, want unchanged %q", got, before)
	}
}

func TestMaterializingTonorisRejectsIllegalTimingWithoutChangingState(t *testing.T) {
	game := newTonorisMaterializationGame(t)
	player := model.PlayerTwo
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	materialize := materializationActionByCardName(t, view, "Tonoris, Lone Mercenary")
	game.state.Scheduler.Phase = PhaseMain
	before := game.StateHash()
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   materialize.Handle,
		},
	); err == nil {
		t.Fatal("Submit() illegal timing error = nil")
	}
	if got := game.StateHash(); got != before {
		t.Fatalf("StateHash() after illegal timing = %q, want unchanged %q", got, before)
	}
}

func TestMaterializingTonorisFizzlesWhenLineageBecomesIllegalBeforeResolution(t *testing.T) {
	game := newTonorisMaterializationGame(t)
	player := model.PlayerTwo
	championBefore := game.state.Champions[player.UID]
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	materialize := materializationActionByCardName(t, view, "Tonoris, Lone Mercenary")
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   materialize.Handle,
		},
	); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	champion := game.state.Champions[player.UID]
	card := game.state.Cards[champion.Card]
	card.Definition = tonorisCardID
	game.state.Cards[champion.Card] = card
	materializedCard := game.state.EffectsStack[0].Source

	passOpportunityRound(t, game, player)
	championAfter := game.state.Champions[player.UID]
	if championAfter.ID != championBefore.ID || championAfter.Card != championBefore.Card || len(championAfter.InnerLineage) != 0 {
		t.Fatalf("Champion after fizzle = %#v, want unchanged lineage %#v", championAfter, championBefore)
	}
	if len(game.state.EffectsStack) != 0 {
		t.Fatalf("EffectsStack after fizzle = %#v, want empty", game.state.EffectsStack)
	}
	zones := game.state.Zones[player.UID]
	if cardIndex(zones.Banishment, materializedCard) < 0 {
		t.Fatalf("Banishment after fizzle = %#v, want materialized card %q", zones.Banishment, materializedCard)
	}
}

func TestNewStandardGameCreatesMirroredOpeningState(t *testing.T) {
	configuration := StandardGameConfig{
		Players: [2]*model.Player{
			&model.Player{
				UID: "aria",
			},
			&model.Player{
				UID: "boris",
			},
		},
		RepositoryRoot: filepath.Join(
			"..",
			"..",
		),
		Seed: 42,
	}

	game, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
	}
	if game.state.Finished {
		t.Fatal("setup game finished unexpectedly")
	}
	if game.state.Scheduler.Kind != schedulerStable {
		t.Fatalf("scheduler kind = %q, want stable", game.state.Scheduler.Kind)
	}
	if game.state.Scheduler.TurnPlayer != configuration.Players[0] {
		t.Fatalf("scheduler turn player = %q, want %q", game.state.Scheduler.TurnPlayer, configuration.Players[0])
	}
	for _, player := range configuration.Players {
		zones := game.state.Zones[player.UID]
		if len(zones.MainDeck) != 53 {
			t.Fatalf("%s main deck = %d cards, want 53", player, len(zones.MainDeck))
		}
		if len(zones.Hand) != 7 {
			t.Fatalf("%s hand = %d cards, want 7", player, len(zones.Hand))
		}
		if len(zones.MaterialDeck) != 11 {
			t.Fatalf("%s material deck = %d cards, want 11", player, len(zones.MaterialDeck))
		}
		if _, exists := game.state.Champions[player.UID]; !exists {
			t.Fatalf("%s has no starting Champion Object", player)
		}
	}
	if len(game.state.Events) != 2 {
		t.Fatalf("committed event batches = %d, want 2", len(game.state.Events))
	}
	for index, batch := range game.state.Events {
		if batch.Cause != spiritOfFireOnEnterCause {
			t.Fatalf("event batch %d cause = %q, want %q", index, batch.Cause, spiritOfFireOnEnterCause)
		}
		if batch.Player != configuration.Players[index] {
			t.Fatalf("event batch %d player = %q, want %q", index, batch.Player, configuration.Players[index])
		}
		if len(batch.Events) != 7 {
			t.Fatalf("event batch %d draw count = %d, want 7", index, len(batch.Events))
		}
	}

	firstView, err := game.PlayerView(configuration.Players[0])
	if err != nil {
		t.Fatalf("first PlayerView() error = %v", err)
	}
	secondView, err := game.PlayerView(configuration.Players[1])
	if err != nil {
		t.Fatalf("second PlayerView() error = %v", err)
	}
	if len(firstView.VisibleEvents) != 7 || len(secondView.VisibleEvents) != 7 {
		t.Fatalf("visible draw events = %d and %d, want 7 each", len(firstView.VisibleEvents), len(secondView.VisibleEvents))
	}
}

func TestStandardSetupEndsWhenStartingHandDrawDecksOut(t *testing.T) {
	configuration := StandardGameConfig{
		Players: [2]*model.Player{
			&model.Player{
				UID: "aria",
			},
			&model.Player{
				UID: "boris",
			},
		},
		RepositoryRoot: filepath.Join(
			"..",
			"..",
		),
		Seed: 42,
	}
	definitions, err := loadCardDefinitions(
		filepath.Join(
			configuration.RepositoryRoot,
			"card",
		),
		filepath.Join(
			configuration.RepositoryRoot,
			"card-data-manifest.json",
		),
	)
	if err != nil {
		t.Fatalf("loadCardDefinitions() error = %v", err)
	}
	deck := fixedStandardDeck()
	deck.MainDeck = DeckSection{
		deckEntry(
			"i9hf5lhl5f",
			3,
		),
	}

	game, err := newStandardSetup(
		configuration,
		definitions,
		deck,
		deck,
	)
	if err != nil {
		t.Fatalf("newStandardSetup() error = %v", err)
	}
	if !game.state.Finished || game.state.Winner != configuration.Players[1] {
		t.Fatalf("deckout result = finished:%t winner:%q, want player two to win", game.state.Finished, game.state.Winner)
	}
	if game.state.Scheduler.Kind != schedulerFinished {
		t.Fatalf("scheduler kind = %q, want finished", game.state.Scheduler.Kind)
	}
	if len(game.state.Events) != 1 || len(game.state.Events[0].Events) != 3 {
		t.Fatalf("committed draws = %#v, want one batch with three events", game.state.Events)
	}
	secondZones := game.state.Zones[configuration.Players[1].UID]
	if len(secondZones.Hand) != 0 {
		t.Fatalf("second player hand = %d cards, want no draws after game end", len(secondZones.Hand))
	}
}

func TestNewStandardGameIsReproducibleForTheSameSeed(t *testing.T) {
	configuration := StandardGameConfig{
		Players: [2]*model.Player{
			&model.Player{
				UID: "aria",
			},
			&model.Player{
				UID: "boris",
			},
		},
		RepositoryRoot: filepath.Join(
			"..",
			"..",
		),
		Seed: 42,
	}
	first, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("first NewStandardGame() error = %v", err)
	}
	second, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("second NewStandardGame() error = %v", err)
	}
	firstHash := first.StateHash()
	secondHash := second.StateHash()
	if firstHash != secondHash {
		t.Fatalf("same-seed setup hashes differ: %q != %q", firstHash, secondHash)
	}
	configuration.Seed = 43
	other, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("different-seed NewStandardGame() error = %v", err)
	}
	otherHash := other.StateHash()
	if firstHash == otherHash {
		t.Fatalf("different seed produced setup hash %q", otherHash)
	}
}

func assertTurnView(
	t *testing.T,
	game *Game,
	player *model.Player,
	wantTurnPlayer *model.Player,
	wantPhase Phase,
	wantOpportunity *model.Player,
	wantActions []constants.ActionKind,
) {
	t.Helper()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if view.TurnPlayer != wantTurnPlayer {
		t.Fatalf("PlayerView().TurnPlayer = %q, want %q", view.TurnPlayer, wantTurnPlayer)
	}
	if view.Phase != wantPhase {
		t.Fatalf("PlayerView().Phase = %q, want %q", view.Phase, wantPhase)
	}
	if view.OpportunityHolder != wantOpportunity {
		t.Fatalf("PlayerView().OpportunityHolder = %q, want %q", view.OpportunityHolder, wantOpportunity)
	}
	allowedActionCapacity := len(wantActions) + 3
	allowedActions := make(map[constants.ActionKind]bool, allowedActionCapacity)
	for _, wantAction := range wantActions {
		allowedActions[wantAction] = true
	}
	if samePlayer(player, wantOpportunity) {
		allowedActions[constants.ActionActivate] = true
		if game.state.Scheduler.TurnNumber > 1 && samePlayer(player, wantTurnPlayer) && wantPhase == PhaseMain && len(game.state.EffectsStack) == 0 {
			allowedActions[constants.ActionAttack] = true
			allowedActions[constants.ActionWield] = true
		}
	}
	for _, action := range view.LegalActions {
		if !allowedActions[action.Kind] {
			t.Fatalf("PlayerView().LegalActions = %#v, unexpected %q", view.LegalActions, action.Kind)
		}
	}
	for _, wantAction := range wantActions {
		found := false
		for _, action := range view.LegalActions {
			if action.Kind == wantAction {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("PlayerView().LegalActions = %#v, missing %q", view.LegalActions, wantAction)
		}
	}
}

func submitActionKind(
	t *testing.T,
	game *Game,
	player *model.Player,
	wantKind constants.ActionKind,
) {
	t.Helper()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	for _, action := range view.LegalActions {
		if action.Kind != wantKind {
			continue
		}
		input := Input{
			Revision: view.Revision,
			Action:   action.Handle,
		}
		if err := game.Submit(player, input); err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
		return
	}
	t.Fatalf("PlayerView().LegalActions = %#v, want %q", view.LegalActions, wantKind)
}

func submitCurrentTurnAction(t *testing.T, game *Game, view PlayerView) {
	t.Helper()
	if view.OpportunityHolder != nil {
		submitActionKind(
			t,
			game,
			view.OpportunityHolder,
			constants.ActionPass,
		)
		return
	}
	submitActionKind(
		t,
		game,
		view.TurnPlayer,
		constants.ActionSkipMaterialize,
	)
}

func newTonorisMaterializationGame(t *testing.T) *Game {
	t.Helper()
	repositoryRoot := filepath.Clean("../..")
	configuration := StandardGameConfig{
		Players: [2]*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: repositoryRoot,
		Seed:           42,
	}
	game, err := NewStandardGame(configuration)
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
	}
	for game.state.Scheduler.TurnPlayer != model.PlayerTwo || game.state.Scheduler.Phase != PhaseMaterialize {
		view, err := game.PlayerView(model.PlayerOne)
		if err != nil {
			t.Fatalf("PlayerView() error = %v", err)
		}
		submitCurrentTurnAction(t, game, view)
	}
	zones := game.state.Zones[model.PlayerTwo.UID]
	if len(zones.Hand) == 0 {
		t.Fatal("player two has no card to place in Memory")
	}
	payment := zones.Hand[len(zones.Hand)-1]
	zones.Hand = zones.Hand[:len(zones.Hand)-1]
	zones.Memory = append(zones.Memory, payment)
	game.state.Zones[model.PlayerTwo.UID] = zones
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	return game
}

func actionByKind(t *testing.T, view PlayerView, want constants.ActionKind) LegalAction {
	t.Helper()
	for _, action := range view.LegalActions {
		if action.Kind == want {
			return action
		}
	}
	t.Fatalf("PlayerView().LegalActions = %#v, want %q", view.LegalActions, want)
	return LegalAction{}
}

func passOpportunityRound(t *testing.T, game *Game, first *model.Player) {
	t.Helper()
	second := game.otherPlayer(first)
	submitActionKind(
		t,
		game,
		first,
		constants.ActionPass,
	)
	submitActionKind(
		t,
		game,
		second,
		constants.ActionPass,
	)
}

func hasVisibleEvent(events []VisibleEvent, kind, cardName string) bool {
	for _, event := range events {
		if event.Kind == kind && event.CardName == cardName {
			return true
		}
	}
	return false
}

func championByOwner(t *testing.T, view PlayerView, owner *model.Player) VisibleChampion {
	t.Helper()
	for _, champion := range view.Champions {
		if champion.Owner == owner {
			return champion
		}
	}
	t.Fatalf("PlayerView().Champions = %#v, want owner %q", view.Champions, owner)
	return VisibleChampion{}
}
