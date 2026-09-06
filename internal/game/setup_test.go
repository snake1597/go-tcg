package game

import (
	"go-tcg/internal/constants"
	"path/filepath"
	"testing"
)

// Rules: 602c917f2f8fd4df7198429a72eb596bf7f647c6,
// general-rules-starting-the-game.md § Standard Game Setup;
// turn-order-main-phase.md § General Rules.
func TestStandardSetupStartsFirstTurnAtMainAndPassesToSecondPlayersDraw(t *testing.T) {
	configuration := StandardGameConfig{
		Players: [2]constants.PlayerID{
			constants.PlayerOne,
			constants.PlayerTwo,
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
		constants.PlayerOne,
		constants.PlayerOne,
		PhaseMain,
		constants.PlayerOne,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)
	assertTurnView(
		t,
		game,
		constants.PlayerTwo,
		constants.PlayerOne,
		PhaseMain,
		constants.PlayerOne,
		[]constants.ActionKind{
			constants.ActionConcede,
		},
	)

	submitActionKind(
		t,
		game,
		constants.PlayerOne,
		constants.ActionPass,
	)
	assertTurnView(
		t,
		game,
		constants.PlayerTwo,
		constants.PlayerOne,
		PhaseMain,
		constants.PlayerTwo,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)

	submitActionKind(
		t,
		game,
		constants.PlayerTwo,
		constants.ActionPass,
	)
	assertTurnView(
		t,
		game,
		constants.PlayerOne,
		constants.PlayerOne,
		PhaseEnd,
		constants.PlayerOne,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)

	submitActionKind(
		t,
		game,
		constants.PlayerOne,
		constants.ActionPass,
	)
	submitActionKind(
		t,
		game,
		constants.PlayerTwo,
		constants.ActionPass,
	)
	secondView, err := game.PlayerView(constants.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if len(secondView.VisibleEvents) != 8 {
		t.Fatalf("PlayerView().VisibleEvents = %#v, want seven opening draws and one turn draw", secondView.VisibleEvents)
	}
	assertTurnView(
		t,
		game,
		constants.PlayerTwo,
		constants.PlayerTwo,
		PhaseMain,
		constants.PlayerTwo,
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
		Players: [2]constants.PlayerID{
			constants.PlayerOne,
			constants.PlayerTwo,
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
		firstView, err := first.PlayerView(constants.PlayerOne)
		if err != nil {
			t.Fatalf("first PlayerView() error = %v", err)
		}
		submitCurrentTurnAction(t, first, firstView)
		secondView, err := second.PlayerView(constants.PlayerOne)
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
		constants.PlayerOne,
		constants.PlayerOne,
		PhaseMaterialize,
		"",
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionSkipMaterialize,
		},
	)
	materializeView, err := first.PlayerView(constants.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	submitCurrentTurnAction(t, first, materializeView)
	assertTurnView(
		t,
		first,
		constants.PlayerOne,
		constants.PlayerOne,
		PhaseRecollection,
		constants.PlayerOne,
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
		Players: [2]constants.PlayerID{
			constants.PlayerOne,
			constants.PlayerTwo,
		},
		RepositoryRoot: filepath.Clean("../.."),
		Seed:           42,
	}
	game, err := NewStandardSetup(configuration)
	if err != nil {
		t.Fatalf("NewStandardSetup() error = %v", err)
	}
	for step := 0; step < 15; step++ {
		view, err := game.PlayerView(constants.PlayerOne)
		if err != nil {
			t.Fatalf("PlayerView() error = %v", err)
		}
		submitCurrentTurnAction(t, game, view)
	}

	assertTurnView(
		t,
		game,
		constants.PlayerTwo,
		constants.PlayerTwo,
		PhaseMaterialize,
		"",
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionSkipMaterialize,
		},
	)
	submitActionKind(
		t,
		game,
		constants.PlayerTwo,
		constants.ActionSkipMaterialize,
	)
	assertTurnView(
		t,
		game,
		constants.PlayerTwo,
		constants.PlayerTwo,
		PhaseRecollection,
		constants.PlayerTwo,
		[]constants.ActionKind{
			constants.ActionConcede,
			constants.ActionPass,
		},
	)
}

func TestNewStandardSetupCreatesMirroredOpeningState(t *testing.T) {
	configuration := StandardGameConfig{
		Players: [2]constants.PlayerID{
			"aria",
			"boris",
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
		zones := game.state.Zones[player]
		if len(zones.MainDeck) != 53 {
			t.Fatalf("%s main deck = %d cards, want 53", player, len(zones.MainDeck))
		}
		if len(zones.Hand) != 7 {
			t.Fatalf("%s hand = %d cards, want 7", player, len(zones.Hand))
		}
		if len(zones.MaterialDeck) != 11 {
			t.Fatalf("%s material deck = %d cards, want 11", player, len(zones.MaterialDeck))
		}
		if _, exists := game.state.Champions[player]; !exists {
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
		Players: [2]constants.PlayerID{
			"aria",
			"boris",
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
	secondZones := game.state.Zones[configuration.Players[1]]
	if len(secondZones.Hand) != 0 {
		t.Fatalf("second player hand = %d cards, want no draws after game end", len(secondZones.Hand))
	}
}

func TestNewStandardSetupIsReproducibleForTheSameSeed(t *testing.T) {
	configuration := StandardGameConfig{
		Players: [2]constants.PlayerID{
			"aria",
			"boris",
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
	player constants.PlayerID,
	wantTurnPlayer constants.PlayerID,
	wantPhase Phase,
	wantOpportunity constants.PlayerID,
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
	player constants.PlayerID,
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
	if view.OpportunityHolder != "" {
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
