package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
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
	Owner    *model.Player `json:"owner"`
	CardName string        `json:"card_name"`
	Power    int           `json:"power"`
	Life     int           `json:"life"`
	Rested   bool          `json:"rested"`
	Taunt    bool          `json:"taunt"`
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
	Revision          uint64            `json:"revision"`
	Finished          bool              `json:"finished"`
	Winner            *model.Player     `json:"winner,omitempty"`
	Diagnostic        string            `json:"diagnostic,omitempty"`
	TurnPlayer        *model.Player     `json:"turn_player,omitempty"`
	Phase             Phase             `json:"phase,omitempty"`
	OpportunityHolder *model.Player     `json:"opportunity_holder,omitempty"`
	Champions         []VisibleChampion `json:"champions,omitempty"`
	Cards             []VisibleCard     `json:"cards"`
	VisibleEvents     []VisibleEvent    `json:"visible_events"`
	LegalActions      []LegalAction     `json:"legal_actions"`
	PendingChoice     *PendingChoice    `json:"pending_choice,omitempty"`
}

type Game struct {
	versions Versions
	players  []*model.Player
	state    gameState
	replay   Replay
}

type gameState struct {
	Revision          uint64
	Finished          bool
	Winner            *model.Player
	Diagnostic        string
	PRNG              prngState
	Knowledge         knowledgeState
	Entities          map[entityID]knowledgeEntity
	Cards             map[cardInstanceID]cardInstance
	Zones             map[string]playerZones
	Champions         map[string]championObject
	Objects           map[objectID]fieldObject
	EffectSources     []cardInstanceID
	EffectsStack      []effectStackItem
	ContinuousEffects []continuousEffect
	AbilityChoice     *abilityChoice
	CardistryUsed     map[objectID]bool
	Scheduler         schedulerFrame
	Events            []eventBatch
	NextHandle        uint64
	NextEvent         uint64
	NextEffect        uint64
	NextAbility       uint64
	NextObject        uint64
}

type prngState struct {
	Seed   uint64 `json:"seed"`
	Cursor uint64 `json:"cursor"`
}

type canonicalState struct {
	SchemaVersion     int                             `json:"schema_version"`
	Versions          Versions                        `json:"versions"`
	Players           []*model.Player                 `json:"players"`
	Revision          uint64                          `json:"revision"`
	Finished          bool                            `json:"finished"`
	Winner            *model.Player                   `json:"winner"`
	Diagnostic        string                          `json:"diagnostic,omitempty"`
	PRNG              prngState                       `json:"prng"`
	Knowledge         knowledgeState                  `json:"knowledge"`
	Entities          map[entityID]knowledgeEntity    `json:"entities"`
	Cards             map[cardInstanceID]cardInstance `json:"cards"`
	Zones             map[string]playerZones          `json:"zones"`
	Champions         map[string]championObject       `json:"champions"`
	Objects           map[objectID]fieldObject        `json:"objects"`
	EffectSources     []cardInstanceID                `json:"effect_sources,omitempty"`
	EffectsStack      []effectStackItem               `json:"effects_stack,omitempty"`
	ContinuousEffects []continuousEffect              `json:"continuous_effects,omitempty"`
	AbilityChoice     *abilityChoice                  `json:"ability_choice,omitempty"`
	CardistryUsed     map[objectID]bool               `json:"cardistry_used,omitempty"`
	Scheduler         schedulerFrame                  `json:"scheduler"`
	Events            []eventBatch                    `json:"events"`
	NextHandle        uint64                          `json:"next_handle"`
	NextEvent         uint64                          `json:"next_event"`
	NextEffect        uint64                          `json:"next_effect"`
	NextAbility       uint64                          `json:"next_ability"`
	NextObject        uint64                          `json:"next_object"`
}

func NewGame(seed uint64) *Game {
	versions := currentVersions()
	game := &Game{
		versions: versions,
		players: []*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		state: gameState{
			Revision:      1,
			Entities:      make(map[entityID]knowledgeEntity),
			Cards:         make(map[cardInstanceID]cardInstance),
			Zones:         make(map[string]playerZones),
			Champions:     make(map[string]championObject),
			Objects:       make(map[objectID]fieldObject),
			CardistryUsed: make(map[objectID]bool),
			Events:        []eventBatch{},
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

func (g *Game) Submit(player *model.Player, input Input) error {
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
	kind, exists := g.state.Knowledge.Actions[player.UID][input.Action]
	if g.state.Knowledge.Choice != nil && (!exists || kind != constants.ActionConcede) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Action)
	}
	if !exists {
		card, materializeExists := g.state.Knowledge.Materializations[player.UID][input.Action]
		if materializeExists {
			if err := g.materializeChampion(
				player,
				card,
			); err != nil {
				return err
			}
		} else {
			card, activateExists := g.state.Knowledge.Activations[player.UID][input.Action]
			if activateExists {
				if err := g.beginActionDeclaration(player, card); err != nil {
					return err
				}
			} else {
				attacker, attackExists := g.state.Knowledge.Attacks[player.UID][input.Action]
				if attackExists {
					if err := g.beginAttack(player, attacker); err != nil {
						return err
					}
				} else {
					source, cardistryExists := g.state.Knowledge.Cardistries[player.UID][input.Action]
					if cardistryExists {
						if err := g.activateCardistry(player, source); err != nil {
							return err
						}
					} else {
						weapon, wieldExists := g.state.Knowledge.Wields[player.UID][input.Action]
						if !wieldExists {
							return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Action)
						}
						if err := g.beginWield(player, weapon); err != nil {
							return err
						}
					}
				}
			}
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

func (g *Game) recordReplayStep(player *model.Player, input Input) {
	g.replay.Steps = append(
		g.replay.Steps,
		ReplayStep{
			Player: player,
			Input:  input,
		},
	)
	g.replay.Steps[len(g.replay.Steps)-1].StateHash = g.StateHash()
}

func (g *Game) PlayerView(player *model.Player) (PlayerView, error) {
	if !g.hasPlayer(player) {
		return PlayerView{}, fmt.Errorf("%w %q", tcgErrors.ErrUnknownPlayer, player)
	}
	return PlayerView{
		Revision:          g.state.Revision,
		Finished:          g.state.Finished,
		Winner:            g.state.Winner,
		Diagnostic:        g.state.Diagnostic,
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
		SchemaVersion: 2,
		Versions:      g.versions,
		Players:       g.players,
		Revision:      g.state.Revision,
		Finished:      g.state.Finished,
		Winner:        g.state.Winner,
		Diagnostic:    g.state.Diagnostic,
		PRNG:          g.state.PRNG,
		Knowledge:     g.state.Knowledge,
		Entities:      g.state.Entities,
		Cards:         g.state.Cards,
		Zones:         g.state.Zones,
		Champions:     g.state.Champions,
		Objects:       g.state.Objects,
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
		[]*model.Player(nil),
		replay.InitialPlayers...,
	)
	replay.Steps = append(
		[]ReplayStep(nil),
		g.replay.Steps...,
	)
	return replay
}

func (g *Game) hasPlayer(player *model.Player) bool {
	return lo.ContainsBy(
		g.players,
		func(candidate *model.Player) bool {
			return samePlayer(candidate, player)
		},
	)
}

func (g *Game) otherPlayer(player *model.Player) *model.Player {
	for _, candidate := range g.players {
		if !samePlayer(candidate, player) {
			return candidate
		}
	}
	return nil
}

func samePlayer(first, second *model.Player) bool {
	if first == nil || second == nil {
		return first == second
	}
	return first.UID == second.UID
}
