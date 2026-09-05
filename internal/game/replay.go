package game

import "fmt"

const ReplayFormatVersion = 1

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
			return fmt.Errorf("replay input %d rejected: %w", index, err)
		}
		if got := game.StateHash(); got != step.StateHash {
			return fmt.Errorf("replay input %d state hash mismatch: got %s, want %s", index, got, step.StateHash)
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
