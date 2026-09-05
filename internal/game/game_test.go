package game

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestNewGamePinsReplayVersionsAndSeed(t *testing.T) {
	const seed uint64 = 42

	game := NewGame(seed)
	replay := game.Replay()

	wantVersions := Versions{
		Engine:   "grand-archive-v1",
		Rules:    "602c917f2f8fd4df7198429a72eb596bf7f647c6",
		CardData: "card-data-v3",
		Deck:     "standard-fire-v2",
		PRNG:     "splitmix64-v1",
	}
	if replay.FormatVersion != 1 {
		t.Fatalf("Replay().FormatVersion = %d, want 1", replay.FormatVersion)
	}
	if replay.Versions != wantVersions {
		t.Fatalf("Replay().Versions = %#v, want %#v", replay.Versions, wantVersions)
	}
	if replay.InitialSeed != seed {
		t.Fatalf("Replay().InitialSeed = %d, want %d", replay.InitialSeed, seed)
	}
}

func TestSameSeedAndInputProduceSameStateHash(t *testing.T) {
	first := NewGame(42)
	second := NewGame(42)
	input := Input{
		Revision: 1,
		Kind:     InputConcede,
	}

	if err := first.Submit(PlayerOne, input); err != nil {
		t.Fatalf("first Submit() error = %v", err)
	}
	if err := second.Submit(PlayerOne, input); err != nil {
		t.Fatalf("second Submit() error = %v", err)
	}

	if first.StateHash() != second.StateHash() {
		t.Fatalf("state hashes differ: %q != %q", first.StateHash(), second.StateHash())
	}
	firstView := first.PlayerView(PlayerTwo)
	secondView := second.PlayerView(PlayerTwo)
	if firstView != secondView || !firstView.Finished || firstView.Winner != PlayerTwo {
		t.Fatalf("final views = %#v and %#v, want player two to win", firstView, secondView)
	}
	firstReplay := first.Replay()
	secondReplay := second.Replay()
	if len(firstReplay.Steps) != 1 || len(secondReplay.Steps) != 1 {
		t.Fatalf("replay step counts = %d and %d, want 1", len(firstReplay.Steps), len(secondReplay.Steps))
	}
	if firstReplay.Steps[0].StateHash != secondReplay.Steps[0].StateHash {
		t.Fatalf("replay hashes differ: %q != %q", firstReplay.Steps[0].StateHash, secondReplay.Steps[0].StateHash)
	}
}

func TestStateHashUsesCanonicalVersionedState(t *testing.T) {
	game := NewGame(42)
	const want = "4da77f261fec6a6d37587fc54f012b9e73aeb88d84fb39281d7b40653b4956da"

	if got := game.StateHash(); got != want {
		t.Fatalf("StateHash() = %q, want canonical digest %q", got, want)
	}
	if other := NewGame(43).StateHash(); other == want {
		t.Fatalf("StateHash() ignored the seed: seed 43 also produced %q", other)
	}
}

func TestRejectedInputDoesNotChangeGame(t *testing.T) {
	testCases := []struct {
		name       string
		player     PlayerID
		input      Input
		wantReason string
	}{
		{
			name:   "stale revision",
			player: PlayerOne,
			input: Input{
				Revision: 0,
				Kind:     InputConcede,
			},
			wantReason: "stale revision",
		},
		{
			name:   "unknown player",
			player: "intruder",
			input: Input{
				Revision: 1,
				Kind:     InputConcede,
			},
			wantReason: "unknown player",
		},
		{
			name:   "unknown input kind",
			player: PlayerOne,
			input: Input{
				Revision: 1,
				Kind:     "unsupported",
			},
			wantReason: "unknown input kind",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				game := NewGame(42)
				beforeView := game.PlayerView(PlayerOne)
				beforeHash := game.StateHash()
				beforeReplay, err := json.Marshal(game.Replay())
				if err != nil {
					t.Fatalf("marshal replay before input: %v", err)
				}

				err = game.Submit(testCase.player, testCase.input)
				if err == nil || !strings.Contains(err.Error(), testCase.wantReason) {
					t.Fatalf("Submit() error = %v, want reason %q", err, testCase.wantReason)
				}

				afterReplay, err := json.Marshal(game.Replay())
				if err != nil {
					t.Fatalf("marshal replay after input: %v", err)
				}
				if game.PlayerView(PlayerOne) != beforeView {
					t.Fatalf("PlayerView() changed after rejected input")
				}
				if game.StateHash() != beforeHash {
					t.Fatalf("StateHash() changed after rejected input")
				}
				if string(afterReplay) != string(beforeReplay) {
					t.Fatalf("Replay() changed after rejected input: before %s, after %s", beforeReplay, afterReplay)
				}
			},
		)
	}
}

func TestReplayVerifiesFromRecordedVersionsAndSeed(t *testing.T) {
	game := NewGame(42)
	input := Input{
		Revision: 1,
		Kind:     InputConcede,
	}
	if err := game.Submit(PlayerTwo, input); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestReplayRejectsEachIncompatibleVersion(t *testing.T) {
	testCases := []struct {
		name       string
		field      string
		wantReason string
	}{
		{
			name:       "format",
			field:      "format",
			wantReason: "replay format version",
		},
		{
			name:       "engine",
			field:      "engine",
			wantReason: "engine version",
		},
		{
			name:       "rules",
			field:      "rules",
			wantReason: "rules version",
		},
		{
			name:       "card data",
			field:      "card_data",
			wantReason: "card data version",
		},
		{
			name:       "deck",
			field:      "deck",
			wantReason: "deck version",
		},
		{
			name:       "PRNG",
			field:      "prng",
			wantReason: "PRNG version",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				replay := NewGame(42).Replay()
				switch testCase.field {
				case "format":
					replay.FormatVersion++
				case "engine":
					replay.Versions.Engine = "old"
				case "rules":
					replay.Versions.Rules = "old"
				case "card_data":
					replay.Versions.CardData = "old"
				case "deck":
					replay.Versions.Deck = "old"
				case "prng":
					replay.Versions.PRNG = "old"
				}

				err := replay.Verify()
				if err == nil || !strings.Contains(err.Error(), testCase.wantReason) {
					t.Fatalf("Verify() error = %v, want reason %q", err, testCase.wantReason)
				}
			},
		)
	}
}

func TestReplayReportsFirstStateHashDivergence(t *testing.T) {
	game := NewGame(42)
	input := Input{
		Revision: 1,
		Kind:     InputConcede,
	}
	if err := game.Submit(PlayerOne, input); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	replay := game.Replay()
	replay.Steps[0].StateHash = strings.Repeat("0", 64)

	err := replay.Verify()
	var diagnostic *ReplayError
	if !errors.As(err, &diagnostic) {
		t.Fatalf("Verify() error = %v, want *ReplayError", err)
	}
	if diagnostic.InputIndex != 0 {
		t.Fatalf("ReplayError.InputIndex = %d, want 0", diagnostic.InputIndex)
	}
	if diagnostic.Failure != ReplayStateHashMismatch {
		t.Fatalf("ReplayError.Failure = %q, want %q", diagnostic.Failure, ReplayStateHashMismatch)
	}
	if !strings.Contains(diagnostic.Error(), "state hash mismatch") {
		t.Fatalf("ReplayError.Error() = %q, want readable hash mismatch reason", diagnostic.Error())
	}
}
