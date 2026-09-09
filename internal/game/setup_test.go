package game

import (
	"encoding/json"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	"path/filepath"
	"testing"
)

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
	game, err := NewStandardSetup(configuration)
	if err != nil {
		t.Fatalf("NewStandardSetup() error = %v", err)
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
	first, err := NewStandardSetup(configuration)
	if err != nil {
		t.Fatalf("first NewStandardSetup() error = %v", err)
	}
	second, err := NewStandardSetup(configuration)
	if err != nil {
		t.Fatalf("second NewStandardSetup() error = %v", err)
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
	game, err := NewStandardSetup(configuration)
	if err != nil {
		t.Fatalf("NewStandardSetup() error = %v", err)
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
	materialize := actionByKind(
		t,
		view,
		constants.ActionMaterialize,
	)
	if materialize.CardName != "Tonoris, Lone Mercenary" {
		t.Fatalf("materialize CardName = %q, want Tonoris, Lone Mercenary", materialize.CardName)
	}
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
	materialize := actionByKind(
		t,
		view,
		constants.ActionMaterialize,
	)
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
	materialize := actionByKind(
		t,
		view,
		constants.ActionMaterialize,
	)
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
	materialize := actionByKind(
		t,
		view,
		constants.ActionMaterialize,
	)
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

func TestNewStandardSetupCreatesMirroredOpeningState(t *testing.T) {
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

	game, err := NewStandardSetup(configuration)
	if err != nil {
		t.Fatalf("NewStandardSetup() error = %v", err)
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

func TestNewStandardSetupIsReproducibleForTheSameSeed(t *testing.T) {
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
	first, err := NewStandardSetup(configuration)
	if err != nil {
		t.Fatalf("first NewStandardSetup() error = %v", err)
	}
	second, err := NewStandardSetup(configuration)
	if err != nil {
		t.Fatalf("second NewStandardSetup() error = %v", err)
	}
	firstHash := first.StateHash()
	secondHash := second.StateHash()
	if firstHash != secondHash {
		t.Fatalf("same-seed setup hashes differ: %q != %q", firstHash, secondHash)
	}
	configuration.Seed = 43
	other, err := NewStandardSetup(configuration)
	if err != nil {
		t.Fatalf("different-seed NewStandardSetup() error = %v", err)
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
	if len(view.LegalActions) != len(wantActions) {
		t.Fatalf("PlayerView().LegalActions = %#v, want %d actions", view.LegalActions, len(wantActions))
	}
	for index, wantAction := range wantActions {
		if view.LegalActions[index].Kind != wantAction {
			t.Fatalf("PlayerView().LegalActions[%d].Kind = %q, want %q", index, view.LegalActions[index].Kind, wantAction)
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
	game, err := NewStandardSetup(configuration)
	if err != nil {
		t.Fatalf("NewStandardSetup() error = %v", err)
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
