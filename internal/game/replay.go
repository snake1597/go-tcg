package game

import (
	"fmt"
	"go-tcg/internal/constants"
)

type ReplayError struct {
	InputIndex int // -1 when verification fails before the first input.
	Failure    constants.ReplayFailure
	Reason     string
	Cause      error
}

func (e *ReplayError) Error() string {
	description := describeReplayFailure(e.Failure)
	if e.Failure == constants.ReplayVersionMismatch {
		return fmt.Sprintf("replay header %s: %s", description, e.Reason)
	}
	return fmt.Sprintf("replay input %d %s: %s", e.InputIndex, description, e.Reason)
}

func (e *ReplayError) Unwrap() error {
	return e.Cause
}

func describeReplayFailure(f constants.ReplayFailure) string {
	switch f {
	case constants.ReplayVersionMismatch:
		return "version mismatch"
	case constants.ReplayInputRejected:
		return "rejected"
	case constants.ReplayStateHashMismatch:
		return "state hash mismatch"
	default:
		return string(f)
	}
}

type Versions struct {
	Engine   string `json:"engine"`
	Rules    string `json:"rules"`
	CardData string `json:"card_data"`
	Deck     string `json:"deck"`
	PRNG     string `json:"prng"`
}

type Replay struct {
	FormatVersion int          `json:"format_version"`
	Versions      Versions     `json:"versions"`
	InitialSeed   uint64       `json:"initial_seed"`
	Steps         []ReplayStep `json:"steps"`
}

type ReplayStep struct {
	Player    constants.PlayerID `json:"player"`
	Input     Input              `json:"input"`
	StateHash string             `json:"state_hash"`
}

// Verify replays the canonical input sequence against a fresh game instance.
func (r Replay) Verify() error {
	if r.FormatVersion != constants.ReplayFormatVersion {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible replay format version %d, want %d",
				r.FormatVersion,
				constants.ReplayFormatVersion,
			),
		)
	}
	if err := verifyVersions(r.Versions); err != nil {
		return err
	}

	game := NewGame(r.InitialSeed)
	for index, step := range r.Steps {
		if err := game.Submit(step.Player, step.Input); err != nil {
			return &ReplayError{
				InputIndex: index,
				Failure:    constants.ReplayInputRejected,
				Reason:     err.Error(),
				Cause:      err,
			}
		}
		if got := game.StateHash(); got != step.StateHash {
			return &ReplayError{
				InputIndex: index,
				Failure:    constants.ReplayStateHashMismatch,
				Reason:     fmt.Sprintf("got %s, want %s", got, step.StateHash),
			}
		}
	}
	return nil
}

func verifyVersions(got Versions) error {
	want := currentVersions()
	if got.Engine != want.Engine {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible engine version %q, want %q",
				got.Engine,
				want.Engine,
			),
		)
	}
	if got.Rules != want.Rules {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible rules version %q, want %q",
				got.Rules,
				want.Rules,
			),
		)
	}
	if got.CardData != want.CardData {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible card data version %q, want %q",
				got.CardData,
				want.CardData,
			),
		)
	}
	if got.Deck != want.Deck {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible deck version %q, want %q",
				got.Deck,
				want.Deck,
			),
		)
	}
	if got.PRNG != want.PRNG {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible PRNG version %q, want %q",
				got.PRNG,
				want.PRNG,
			),
		)
	}
	return nil
}

func newReplayVersionMismatch(reason string) *ReplayError {
	return &ReplayError{
		InputIndex: -1,
		Failure:    constants.ReplayVersionMismatch,
		Reason:     reason,
	}
}
