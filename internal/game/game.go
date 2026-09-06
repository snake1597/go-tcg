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
	Revision uint64     `json:"revision"`
	Action   ViewHandle `json:"action"`
	Choice   ViewHandle `json:"choice"`
}

type ViewHandle string

type LegalAction struct {
	Handle   ViewHandle           `json:"handle"`
	Kind     constants.ActionKind `json:"kind"`
	CardName string               `json:"card_name,omitempty"`
}

type VisibleChampion struct {
	Owner    constants.PlayerID `json:"owner"`
	CardName string             `json:"card_name"`
	Taunt    bool               `json:"taunt"`
}

type VisibleCard struct {
	Handle ViewHandle `json:"handle"`
	Name   string     `json:"name"`
}

type VisibleEvent struct {
	Kind     string `json:"kind"`
	CardName string `json:"card_name"`
}

type PendingChoice struct {
	Options []ViewHandle `json:"options"`
}

type PlayerView struct {
	Revision          uint64             `json:"revision"`
	Finished          bool               `json:"finished"`
	Winner            constants.PlayerID `json:"winner,omitempty"`
	TurnPlayer        constants.PlayerID `json:"turn_player,omitempty"`
	Phase             Phase              `json:"phase,omitempty"`
	OpportunityHolder constants.PlayerID `json:"opportunity_holder,omitempty"`
	Champions         []VisibleChampion  `json:"champions,omitempty"`
	Cards             []VisibleCard      `json:"cards"`
	VisibleEvents     []VisibleEvent     `json:"visible_events"`
	LegalActions      []LegalAction      `json:"legal_actions"`
	PendingChoice     *PendingChoice     `json:"pending_choice,omitempty"`
}

type Game struct {
	versions Versions
	players  []constants.PlayerID
	state    gameState
	replay   Replay
}

type gameState struct {
	Revision      uint64
	Finished      bool
	Winner        constants.PlayerID
	PRNG          prngState
	Knowledge     knowledgeState
	Entities      map[entityID]knowledgeEntity
	Cards         map[cardInstanceID]cardInstance
	Zones         map[constants.PlayerID]playerZones
	Champions     map[constants.PlayerID]championObject
	EffectSources []cardInstanceID
	EffectsStack  []effectStackItem
	Scheduler     schedulerFrame
	Events        []eventBatch
	NextHandle    uint64
	NextEvent     uint64
}

type prngState struct {
	Seed   uint64 `json:"seed"`
	Cursor uint64 `json:"cursor"`
}

type canonicalState struct {
	SchemaVersion int                                   `json:"schema_version"`
	Versions      Versions                              `json:"versions"`
	Players       []constants.PlayerID                  `json:"players"`
	Revision      uint64                                `json:"revision"`
	Finished      bool                                  `json:"finished"`
	Winner        constants.PlayerID                    `json:"winner"`
	PRNG          prngState                             `json:"prng"`
	Knowledge     knowledgeState                        `json:"knowledge"`
	Entities      map[entityID]knowledgeEntity          `json:"entities"`
	Cards         map[cardInstanceID]cardInstance       `json:"cards"`
	Zones         map[constants.PlayerID]playerZones    `json:"zones"`
	Champions     map[constants.PlayerID]championObject `json:"champions"`
	EffectSources []cardInstanceID                      `json:"effect_sources,omitempty"`
	EffectsStack  []effectStackItem                     `json:"effects_stack,omitempty"`
	Scheduler     schedulerFrame                        `json:"scheduler"`
	Events        []eventBatch                          `json:"events"`
	NextHandle    uint64                                `json:"next_handle"`
	NextEvent     uint64                                `json:"next_event"`
}

func NewGame(seed uint64) *Game {
	versions := currentVersions()
	game := &Game{
		versions: versions,
		players: []constants.PlayerID{
			constants.PlayerOne,
			constants.PlayerTwo,
		},
		state: gameState{
			Revision:  1,
			Entities:  make(map[entityID]knowledgeEntity),
			Cards:     make(map[cardInstanceID]cardInstance),
			Zones:     make(map[constants.PlayerID]playerZones),
			Champions: make(map[constants.PlayerID]championObject),
			Events:    []eventBatch{},
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
	game.initializeKnowledgeState()
	game.captureReplayInitialState()
	return game
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
	if !g.hasPlayer(player) {
		return fmt.Errorf("%w %q", tcgErrors.ErrUnknownPlayer, player)
	}
	if g.state.Finished {
		return tcgErrors.ErrGameFinished
	}
	if input.Revision != g.state.Revision {
		return fmt.Errorf("%w: got %d, current %d", tcgErrors.ErrStaleRevision, input.Revision, g.state.Revision)
	}
	if input.Action != "" && input.Choice != "" {
		return fmt.Errorf("%w: action and choice cannot be submitted together", tcgErrors.ErrInvalidViewHandle)
	}
	if input.Choice != "" {
		err := g.submitChoice(
			player,
			input,
		)
		if err != nil {
			return err
		}
		g.recordReplayStep(
			player,
			input,
		)
		return nil
	}
	kind, exists := g.state.Knowledge.Actions[player][input.Action]
	if !exists {
		card, materializeExists := g.state.Knowledge.Materializations[player][input.Action]
		if !materializeExists {
			return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Action)
		}
		if err := g.materializeChampion(
			player,
			card,
		); err != nil {
			return err
		}
		g.advanceKnowledgeRevision()
		g.recordReplayStep(
			player,
			input,
		)
		return nil
	}
	switch kind {
	case constants.ActionConcede:
		g.state.Finished = true
		g.state.Winner = g.otherPlayer(player)
		g.state.Scheduler = schedulerFrame{
			Kind: schedulerFinished,
		}
	case constants.ActionPass:
		if err := g.passOpportunity(player); err != nil {
			return err
		}
	case constants.ActionSkipMaterialize:
		if err := g.skipMaterialize(player); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Action)
	}
	g.advanceKnowledgeRevision()
	g.recordReplayStep(
		player,
		input,
	)
	return nil
}

func (g *Game) recordReplayStep(player constants.PlayerID, input Input) {
	g.replay.Steps = append(
		g.replay.Steps,
		ReplayStep{
			Player: player,
			Input:  input,
		},
	)
	g.replay.Steps[len(g.replay.Steps)-1].StateHash = g.StateHash()
}

func (g *Game) PlayerView(player constants.PlayerID) (PlayerView, error) {
	if !g.hasPlayer(player) {
		return PlayerView{}, fmt.Errorf("%w %q", tcgErrors.ErrUnknownPlayer, player)
	}
	return PlayerView{
		Revision:          g.state.Revision,
		Finished:          g.state.Finished,
		Winner:            g.state.Winner,
		TurnPlayer:        g.state.Scheduler.TurnPlayer,
		Phase:             g.state.Scheduler.Phase,
		OpportunityHolder: g.state.Scheduler.OpportunityHolder,
		Champions: g.visibleChampions(
			player,
		),
		Cards: g.visibleCards(
			player,
		),
		VisibleEvents: g.visibleEvents(
			player,
		),
		LegalActions: g.legalActions(
			player,
		),
		PendingChoice: g.pendingChoice(
			player,
		),
	}, nil
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
		Knowledge:     g.state.Knowledge,
		Entities:      g.state.Entities,
		Cards:         g.state.Cards,
		Zones:         g.state.Zones,
		Champions:     g.state.Champions,
		EffectSources: g.state.EffectSources,
		EffectsStack:  g.state.EffectsStack,
		Scheduler:     g.state.Scheduler,
		Events:        g.state.Events,
		NextHandle:    g.state.NextHandle,
		NextEvent:     g.state.NextEvent,
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
	if replay.InitialState != nil {
		state := cloneGameState(*replay.InitialState)
		replay.InitialState = &state
	}
	replay.InitialPlayers = append(
		[]constants.PlayerID(nil),
		replay.InitialPlayers...,
	)
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
