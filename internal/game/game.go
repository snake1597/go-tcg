package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-tcg/internal/constants"
	tcgErrors "go-tcg/internal/tcg_errors"

	"github.com/samber/lo"
)

type Input struct {
	Revision uint64              `json:"revision"`
	Kind     constants.InputKind `json:"kind"`
}

type PlayerView struct {
	Revision uint64             `json:"revision"`
	Finished bool               `json:"finished"`
	Winner   constants.PlayerID `json:"winner,omitempty"`
}

type Game struct {
	versions Versions
	players  []constants.PlayerID
	state    gameState
	replay   Replay
}

type gameState struct {
	Revision uint64
	Finished bool
	Winner   constants.PlayerID
	PRNG     prngState
}

type prngState struct {
	Seed   uint64 `json:"seed"`
	Cursor uint64 `json:"cursor"`
}

type canonicalState struct {
	SchemaVersion int                  `json:"schema_version"`
	Versions      Versions             `json:"versions"`
	Players       []constants.PlayerID `json:"players"`
	Revision      uint64               `json:"revision"`
	Finished      bool                 `json:"finished"`
	Winner        constants.PlayerID   `json:"winner"`
	PRNG          prngState            `json:"prng"`
}

func NewGame(seed uint64) *Game {
	versions := currentVersions()
	return &Game{
		versions: versions,
		players: []constants.PlayerID{
			constants.PlayerOne,
			constants.PlayerTwo,
		},
		state: gameState{
			Revision: 1,
			PRNG: prngState{
				Seed: seed,
			},
		},
		replay: Replay{
			FormatVersion: constants.ReplayFormatVersion,
			Versions:      versions,
			InitialSeed:   seed,
		},
	}
}

func currentVersions() Versions {
	return Versions{
		Engine:   "grand-archive-v1",
		Rules:    "602c917f2f8fd4df7198429a72eb596bf7f647c6",
		CardData: "card-data-v3",
		Deck:     "standard-fire-v2",
		PRNG:     "splitmix64-v1",
	}
}

func (g *Game) Submit(player constants.PlayerID, input Input) error {
	if g.state.Finished {
		return tcgErrors.ErrGameFinished
	}
	if input.Revision != g.state.Revision {
		return fmt.Errorf("%w: got %d, current %d", tcgErrors.ErrStaleRevision, input.Revision, g.state.Revision)
	}
	if !g.hasPlayer(player) {
		return fmt.Errorf("%w %q", tcgErrors.ErrUnknownPlayer, player)
	}
	if input.Kind != constants.InputConcede {
		return fmt.Errorf("%w %q", tcgErrors.ErrUnknownInputKind, input.Kind)
	}

	g.state.Finished = true
	g.state.Winner = g.otherPlayer(player)
	g.state.Revision++
	g.replay.Steps = append(
		g.replay.Steps,
		ReplayStep{
			Player: player,
			Input:  input,
		},
	)
	g.replay.Steps[len(g.replay.Steps)-1].StateHash = g.StateHash()
	return nil
}

func (g *Game) PlayerView(player constants.PlayerID) PlayerView {
	return PlayerView{
		Revision: g.state.Revision,
		Finished: g.state.Finished,
		Winner:   g.state.Winner,
	}
}

func (g *Game) StateHash() string {
	canonical := canonicalState{
		SchemaVersion: 1,
		Versions:      g.versions,
		Players:       g.players,
		Revision:      g.state.Revision,
		Finished:      g.state.Finished,
		Winner:        g.state.Winner,
		PRNG:          g.state.PRNG,
	}
	state, err := json.Marshal(canonical)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(state)
	return hex.EncodeToString(sum[:])
}

func (g *Game) Replay() Replay {
	replay := g.replay
	replay.Steps = append(
		[]ReplayStep(nil),
		g.replay.Steps...,
	)
	return replay
}

func (g *Game) hasPlayer(player constants.PlayerID) bool {
	return lo.Contains(g.players, player)
}

func (g *Game) otherPlayer(player constants.PlayerID) constants.PlayerID {
	for _, candidate := range g.players {
		if candidate != player {
			return candidate
		}
	}
	return ""
}
