// Package game contains the authoritative, deterministic game-module seam.
package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrGameFinished     = errors.New("game is finished")
	ErrStaleRevision    = errors.New("stale revision")
	ErrUnknownPlayer    = errors.New("unknown player")
	ErrUnknownInputKind = errors.New("unknown input kind")
)

type PlayerID string

const (
	PlayerOne PlayerID = "player-1"
	PlayerTwo PlayerID = "player-2"
)

type InputKind string

const InputConcede InputKind = "concede"

type Input struct {
	Revision uint64    `json:"revision"`
	Kind     InputKind `json:"kind"`
}

type PlayerView struct {
	Revision uint64   `json:"revision"`
	Finished bool     `json:"finished"`
	Winner   PlayerID `json:"winner,omitempty"`
}

type Game struct {
	versions Versions
	players  []PlayerID
	state    gameState
	replay   Replay
}

type gameState struct {
	Revision uint64
	Finished bool
	Winner   PlayerID
	PRNG     prngState
}

type prngState struct {
	Seed   uint64 `json:"seed"`
	Cursor uint64 `json:"cursor"`
}

type canonicalState struct {
	SchemaVersion int        `json:"schema_version"`
	Versions      Versions   `json:"versions"`
	Players       []PlayerID `json:"players"`
	Revision      uint64     `json:"revision"`
	Finished      bool       `json:"finished"`
	Winner        PlayerID   `json:"winner"`
	PRNG          prngState  `json:"prng"`
}

func NewGame(seed uint64) *Game {
	versions := currentVersions()
	return &Game{
		versions: versions,
		players: []PlayerID{
			PlayerOne,
			PlayerTwo,
		},
		state: gameState{
			Revision: 1,
			PRNG: prngState{
				Seed: seed,
			},
		},
		replay: Replay{
			FormatVersion: ReplayFormatVersion,
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

func (g *Game) Submit(player PlayerID, input Input) error {
	if g.state.Finished {
		return ErrGameFinished
	}
	if input.Revision != g.state.Revision {
		return fmt.Errorf("%w: got %d, current %d", ErrStaleRevision, input.Revision, g.state.Revision)
	}
	if !g.hasPlayer(player) {
		return fmt.Errorf("%w %q", ErrUnknownPlayer, player)
	}
	if input.Kind != InputConcede {
		return fmt.Errorf("%w %q", ErrUnknownInputKind, input.Kind)
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

func (g *Game) PlayerView(player PlayerID) PlayerView {
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

func (g *Game) hasPlayer(player PlayerID) bool {
	return contains(g.players, player)
}

func (g *Game) otherPlayer(player PlayerID) PlayerID {
	for _, candidate := range g.players {
		if candidate != player {
			return candidate
		}
	}
	return ""
}

func contains[T comparable](items []T, want T) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
