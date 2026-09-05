package game

import (
	"go-tcg/internal/constants"
	"path/filepath"
	"testing"
)

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
