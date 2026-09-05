package game

import "fmt"

const ReplayFormatVersion = 1

type ReplayFailure string

const (
	ReplayInputRejected     ReplayFailure = "input_rejected"
	ReplayStateHashMismatch ReplayFailure = "state_hash_mismatch"
)

type ReplayError struct {
	InputIndex int
	Failure    ReplayFailure
	Reason     string
	Cause      error
}

func (e *ReplayError) Error() string {
	return fmt.Sprintf("replay input %d %s: %s", e.InputIndex, e.Failure.description(), e.Reason)
}

func (e *ReplayError) Unwrap() error {
	return e.Cause
}

func (f ReplayFailure) description() string {
	switch f {
	case ReplayInputRejected:
		return "rejected"
	case ReplayStateHashMismatch:
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
	Player    PlayerID `json:"player"`
	Input     Input    `json:"input"`
	StateHash string   `json:"state_hash"`
}

// Verify replays the canonical input sequence against a fresh game instance.
func (r Replay) Verify() error {
	if r.FormatVersion != ReplayFormatVersion {
		return fmt.Errorf("incompatible replay format version %d, want %d", r.FormatVersion, ReplayFormatVersion)
	}
	if err := verifyVersions(r.Versions); err != nil {
		return err
	}

	game := NewGame(r.InitialSeed)
	for index, step := range r.Steps {
		if err := game.Submit(step.Player, step.Input); err != nil {
			return &ReplayError{
				InputIndex: index,
				Failure:    ReplayInputRejected,
				Reason:     err.Error(),
				Cause:      err,
			}
		}
		if got := game.StateHash(); got != step.StateHash {
			return &ReplayError{
				InputIndex: index,
				Failure:    ReplayStateHashMismatch,
				Reason:     fmt.Sprintf("got %s, want %s", got, step.StateHash),
			}
		}
	}
	return nil
}

func verifyVersions(got Versions) error {
	want := currentVersions()
	if got.Engine != want.Engine {
		return fmt.Errorf("incompatible engine version %q, want %q", got.Engine, want.Engine)
	}
	if got.Rules != want.Rules {
		return fmt.Errorf("incompatible rules version %q, want %q", got.Rules, want.Rules)
	}
	if got.CardData != want.CardData {
		return fmt.Errorf("incompatible card data version %q, want %q", got.CardData, want.CardData)
	}
	if got.Deck != want.Deck {
		return fmt.Errorf("incompatible deck version %q, want %q", got.Deck, want.Deck)
	}
	if got.PRNG != want.PRNG {
		return fmt.Errorf("incompatible PRNG version %q, want %q", got.PRNG, want.PRNG)
	}
	return nil
}
