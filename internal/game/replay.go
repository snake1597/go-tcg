package game

import "fmt"

type Replay struct {
	EngineVersion   string        `json:"engine_version"`
	RulesVersion    string        `json:"rules_version"`
	CardDataVersion string        `json:"card_data_version"`
	DeckVersion     string        `json:"deck_version"`
	PRNGVersion     string        `json:"prng_version"`
	Seed            uint64        `json:"seed"`
	Inputs          []ReplayInput `json:"inputs"`
	Hashes          []string      `json:"hashes"`
}

type ReplayInput struct {
	Player   PlayerID `json:"player"`
	Kind     string   `json:"kind"`
	Revision uint64   `json:"revision,omitempty"`
	Handle   string   `json:"handle,omitempty"`
	Option   string   `json:"option,omitempty"`
}

// Verify replays the canonical input sequence against a fresh game instance.
func (r Replay) Verify(game *Game) error {
	if len(r.Inputs) != len(r.Hashes) {
		return fmt.Errorf("replay has %d inputs but %d hashes", len(r.Inputs), len(r.Hashes))
	}
	for index, input := range r.Inputs {
		var err error
		switch input.Kind {
		case "choice":
			err = game.SubmitChoice(input.Player, SubmitChoice{Revision: input.Revision, Handle: input.Handle, Option: input.Option})
		case "concede":
			err = game.Concede(input.Player)
		default:
			return fmt.Errorf("replay input %d has unknown kind %q", index, input.Kind)
		}
		if err != nil {
			return fmt.Errorf("replay input %d: %w", index, err)
		}
		if got := game.StateHash(); got != r.Hashes[index] {
			return fmt.Errorf("replay state %d hash = %s, want %s", index, got, r.Hashes[index])
		}
	}
	return nil
}
