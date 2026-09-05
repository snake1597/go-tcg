package game

import (
	"encoding/json"
	"errors"
	"go-tcg/internal/constants"
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
		Kind:     constants.InputConcede,
	}

	if err := first.Submit(constants.PlayerOne, input); err != nil {
		t.Fatalf("first Submit() error = %v", err)
	}
	if err := second.Submit(constants.PlayerOne, input); err != nil {
		t.Fatalf("second Submit() error = %v", err)
	}

	firstHash := first.StateHash()
	secondHash := second.StateHash()
	if firstHash != secondHash {
		t.Fatalf("state hashes differ: %q != %q", firstHash, secondHash)
	}
	firstView := first.PlayerView(constants.PlayerTwo)
	secondView := second.PlayerView(constants.PlayerTwo)
	if firstView != secondView || !firstView.Finished || firstView.Winner != constants.PlayerTwo {
		t.Fatalf("final views = %#v and %#v, want player two to win", firstView, secondView)
	}
	firstReplay := first.Replay()
	secondReplay := second.Replay()
	firstStepCount := len(firstReplay.Steps)
	secondStepCount := len(secondReplay.Steps)
	if firstStepCount != 1 || secondStepCount != 1 {
		t.Fatalf("replay step counts = %d and %d, want 1", firstStepCount, secondStepCount)
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
	otherGame := NewGame(43)
	if other := otherGame.StateHash(); other == want {
		t.Fatalf("StateHash() ignored the seed: seed 43 also produced %q", other)
	}
}

func TestRejectedInputDoesNotChangeGame(t *testing.T) {
	testCases := []struct {
		name       string
		player     constants.PlayerID
		input      Input
		wantReason string
	}{
		{
			name:   "stale revision",
			player: constants.PlayerOne,
			input: Input{
				Revision: 0,
				Kind:     constants.InputConcede,
			},
			wantReason: "stale revision",
		},
		{
			name:   "unknown player",
			player: "intruder",
			input: Input{
				Revision: 1,
				Kind:     constants.InputConcede,
			},
			wantReason: "unknown player",
		},
		{
			name:   "unknown input kind",
			player: constants.PlayerOne,
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
				beforeView := game.PlayerView(constants.PlayerOne)
				beforeHash := game.StateHash()
				replayBeforeInput := game.Replay()
				beforeReplay, err := json.Marshal(replayBeforeInput)
				if err != nil {
					t.Fatalf("marshal replay before input: %v", err)
				}

				err = game.Submit(testCase.player, testCase.input)
				if err == nil {
					t.Fatalf("Submit() error = nil, want reason %q", testCase.wantReason)
				}
				errorMessage := err.Error()
				if !strings.Contains(errorMessage, testCase.wantReason) {
					t.Fatalf("Submit() error = %v, want reason %q", err, testCase.wantReason)
				}

				replayAfterInput := game.Replay()
				afterReplay, err := json.Marshal(replayAfterInput)
				if err != nil {
					t.Fatalf("marshal replay after input: %v", err)
				}
				if game.PlayerView(constants.PlayerOne) != beforeView {
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
		Kind:     constants.InputConcede,
	}
	if err := game.Submit(constants.PlayerTwo, input); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	replay := game.Replay()
	if err := replay.Verify(); err != nil {
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
				game := NewGame(42)
				replay := game.Replay()
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
				var diagnostic *ReplayError
				if !errors.As(err, &diagnostic) {
					t.Fatalf("Verify() error = %v, want *ReplayError", err)
				}
				if diagnostic.InputIndex != -1 {
					t.Fatalf("ReplayError.InputIndex = %d, want -1", diagnostic.InputIndex)
				}
				if diagnostic.Failure != constants.ReplayVersionMismatch {
					t.Fatalf("ReplayError.Failure = %q, want %q", diagnostic.Failure, constants.ReplayVersionMismatch)
				}
				errorMessage := diagnostic.Error()
				if !strings.Contains(errorMessage, testCase.wantReason) {
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
		Kind:     constants.InputConcede,
	}
	if err := game.Submit(constants.PlayerOne, input); err != nil {
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
	if diagnostic.Failure != constants.ReplayStateHashMismatch {
		t.Fatalf("ReplayError.Failure = %q, want %q", diagnostic.Failure, constants.ReplayStateHashMismatch)
	}
	errorMessage := diagnostic.Error()
	if !strings.Contains(errorMessage, "state hash mismatch") {
		t.Fatalf("ReplayError.Error() = %q, want readable hash mismatch reason", errorMessage)
	}
}
