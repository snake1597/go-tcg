// Package game contains the authoritative, deterministic game-module seam.
package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

var errInvalidInput = errors.New("invalid game input")

type PlayerID string

type PendingChoice struct {
	ID      string
	Player  PlayerID
	Options []string
}

type ChoiceView struct {
	Handle  string   `json:"handle"`
	Options []string `json:"options"`
}

type PlayerView struct {
	Revision      uint64      `json:"revision"`
	PendingChoice *ChoiceView `json:"pending_choice,omitempty"`
	Finished      bool        `json:"finished"`
	Winner        PlayerID    `json:"winner,omitempty"`
}

type SubmitChoice struct {
	Revision uint64
	Handle   string
	Option   string
}

type Game struct {
	players []PlayerID
	state   gameState
	replay  Replay
}

type gameState struct {
	Revision  uint64
	Pending   *PendingChoice
	Selection string
	Finished  bool
	Winner    PlayerID
}

func newGame(players []PlayerID, pending PendingChoice) *Game {
	game := &Game{
		players: players,
		state:   gameState{Revision: 1, Pending: &pending},
		replay: Replay{
			EngineVersion:   "walking-skeleton-v1",
			RulesVersion:    "test-fixture",
			CardDataVersion: "test-fixture",
			DeckVersion:     "test-fixture",
			PRNGVersion:     "none",
			Seed:            0,
		},
	}
	return game
}

func (g *Game) PlayerView(player PlayerID) PlayerView {
	view := PlayerView{Revision: g.state.Revision, Finished: g.state.Finished, Winner: g.state.Winner}
	if g.state.Pending != nil && g.state.Pending.Player == player {
		view.PendingChoice = &ChoiceView{
			Handle:  g.handleFor(player),
			Options: append([]string(nil), g.state.Pending.Options...),
		}
	}
	return view
}

func (g *Game) SubmitChoice(player PlayerID, input SubmitChoice) error {
	if g.state.Finished || g.state.Pending == nil || input.Revision != g.state.Revision || player != g.state.Pending.Player || input.Handle != g.handleFor(player) || !contains(g.state.Pending.Options, input.Option) {
		return errInvalidInput
	}
	g.state.Selection = input.Option
	g.state.Pending = nil
	g.state.Revision++
	g.replay.Inputs = append(g.replay.Inputs, ReplayInput{Player: player, Kind: "choice", Revision: input.Revision, Handle: input.Handle, Option: input.Option})
	g.replay.Hashes = append(g.replay.Hashes, g.StateHash())
	return nil
}

func (g *Game) Concede(player PlayerID) error {
	if g.state.Finished || !g.hasPlayer(player) {
		return errInvalidInput
	}
	g.state.Finished = true
	g.state.Winner = g.otherPlayer(player)
	g.state.Revision++
	g.replay.Inputs = append(g.replay.Inputs, ReplayInput{Player: player, Kind: "concede"})
	g.replay.Hashes = append(g.replay.Hashes, g.StateHash())
	return nil
}

func (g *Game) StateHash() string {
	state, err := json.Marshal(g.state)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(state)
	return hex.EncodeToString(sum[:])
}

func (g *Game) Replay() Replay {
	return g.replay
}

func (g *Game) handleFor(player PlayerID) string {
	return fmt.Sprintf("choice-%s-r%d", player, g.state.Revision)
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
