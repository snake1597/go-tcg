package game

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go-tcg/internal/constants"
	tcgErrors "go-tcg/internal/tcg_errors"
	"sort"
)

type entityID string

type knowledgeEntity struct {
	Name string `json:"name"`
}

type knowledgeState struct {
	Actions map[constants.PlayerID]map[ViewHandle]constants.ActionKind `json:"actions"`
	Cards   map[constants.PlayerID]map[entityID]ViewHandle             `json:"cards"`
	Events  map[constants.PlayerID][]VisibleEvent                      `json:"events"`
	Choice  *pendingChoice                                             `json:"choice,omitempty"`
}

type pendingChoice struct {
	Actor   constants.PlayerID      `json:"actor"`
	Options map[ViewHandle]entityID `json:"options"`
}

func (g *Game) initializeKnowledgeState() {
	knowledge := knowledgeState{
		Actions: make(map[constants.PlayerID]map[ViewHandle]constants.ActionKind, len(g.players)),
		Cards:   make(map[constants.PlayerID]map[entityID]ViewHandle, len(g.players)),
		Events:  make(map[constants.PlayerID][]VisibleEvent, len(g.players)),
	}
	for _, player := range g.players {
		knowledge.Actions[player] = make(map[ViewHandle]constants.ActionKind)
		knowledge.Cards[player] = make(map[entityID]ViewHandle)
		knowledge.Events[player] = []VisibleEvent{}
	}
	g.state.Knowledge = knowledge
	g.refreshLegalActions()
}

func (g *Game) refreshLegalActions() {
	for _, player := range g.players {
		actions := g.state.Knowledge.Actions[player]
		if g.state.Finished {
			clear(actions)
			continue
		}
		if !g.hasAction(player, constants.ActionConcede) {
			handle := g.newViewHandle(
				player,
				"action:concede",
			)
			actions[handle] = constants.ActionConcede
		}
	}
}

func (g *Game) hasAction(player constants.PlayerID, kind constants.ActionKind) bool {
	for _, candidate := range g.state.Knowledge.Actions[player] {
		if candidate == kind {
			return true
		}
	}
	return false
}

func (g *Game) legalActions(player constants.PlayerID) []LegalAction {
	actions := g.state.Knowledge.Actions[player]
	legalActions := make([]LegalAction, 0, len(actions))
	for handle, kind := range actions {
		legalActions = append(
			legalActions,
			LegalAction{
				Handle: handle,
				Kind:   kind,
			},
		)
	}
	sort.Slice(
		legalActions,
		func(first, second int) bool {
			return legalActions[first].Kind < legalActions[second].Kind
		},
	)
	return legalActions
}

func (g *Game) grantCardTracking(player constants.PlayerID, card entityID) {
	if _, exists := g.state.Knowledge.Cards[player][card]; exists {
		return
	}
	handle := g.newViewHandle(
		player,
		"card:"+string(card),
	)
	g.state.Knowledge.Cards[player][card] = handle
}

func (g *Game) revokeCardTracking(player constants.PlayerID, card entityID) {
	handle, exists := g.state.Knowledge.Cards[player][card]
	if !exists {
		return
	}
	delete(g.state.Knowledge.Cards[player], card)
	choice := g.state.Knowledge.Choice
	if choice == nil || choice.Actor != player {
		return
	}
	delete(choice.Options, handle)
	if len(choice.Options) == 0 {
		g.state.Knowledge.Choice = nil
	}
}

func (g *Game) visibleCards(player constants.PlayerID) []VisibleCard {
	cards := g.state.Knowledge.Cards[player]
	visibleCards := make([]VisibleCard, 0, len(cards))
	for card, handle := range cards {
		visibleCards = append(
			visibleCards,
			VisibleCard{
				Handle: handle,
				Name:   g.state.Entities[card].Name,
			},
		)
	}
	sort.Slice(
		visibleCards,
		func(first, second int) bool {
			return visibleCards[first].Handle < visibleCards[second].Handle
		},
	)
	return visibleCards
}

func (g *Game) recordVisibleEvent(player constants.PlayerID, kind string, card entityID) {
	event := VisibleEvent{
		Kind:     kind,
		CardName: g.state.Entities[card].Name,
	}
	g.state.Knowledge.Events[player] = append(
		g.state.Knowledge.Events[player],
		event,
	)
}

func (g *Game) visibleEvents(player constants.PlayerID) []VisibleEvent {
	return append(
		[]VisibleEvent(nil),
		g.state.Knowledge.Events[player]...,
	)
}

func (g *Game) setPendingCardChoice(player constants.PlayerID, card entityID) {
	handle, exists := g.state.Knowledge.Cards[player][card]
	if !exists {
		return
	}
	g.state.Knowledge.Choice = &pendingChoice{
		Actor: player,
		Options: map[ViewHandle]entityID{
			handle: card,
		},
	}
}

func (g *Game) pendingChoice(player constants.PlayerID) *PendingChoice {
	choice := g.state.Knowledge.Choice
	if choice == nil || choice.Actor != player {
		return nil
	}
	options := make([]ViewHandle, 0, len(choice.Options))
	for handle := range choice.Options {
		options = append(options, handle)
	}
	sort.Slice(
		options,
		func(first, second int) bool {
			return options[first] < options[second]
		},
	)
	return &PendingChoice{
		Options: options,
	}
}

func (g *Game) submitChoice(player constants.PlayerID, input Input) error {
	choice := g.state.Knowledge.Choice
	if choice == nil || choice.Actor != player {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Choice)
	}
	if _, exists := choice.Options[input.Choice]; !exists {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Choice)
	}
	g.state.Knowledge.Choice = nil
	g.advanceKnowledgeRevision()
	return nil
}

func (g *Game) advanceKnowledgeRevision() {
	g.state.Revision++
	g.refreshLegalActions()
}

func (g *Game) newViewHandle(player constants.PlayerID, subject string) ViewHandle {
	g.state.NextHandle++
	value := fmt.Sprintf(
		"view-handle-v1:%d:%d:%s:%s",
		g.state.PRNG.Seed,
		g.state.NextHandle,
		player,
		subject,
	)
	valueBytes := []byte(value)
	sum := sha256.Sum256(valueBytes)
	encoded := hex.EncodeToString(sum[:])
	return ViewHandle(encoded)
}
