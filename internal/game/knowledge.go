package game

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
	"sort"
)

type entityID string

type knowledgeEntity struct {
	Name string `json:"name"`
}

type knowledgeState struct {
	Actions          map[string]map[ViewHandle]constants.ActionKind `json:"actions"`
	Materializations map[string]map[ViewHandle]cardInstanceID       `json:"materializations"`
	Activations      map[string]map[ViewHandle]cardInstanceID       `json:"activations"`
	Cards            map[string]map[entityID]ViewHandle             `json:"cards"`
	Events           map[string][]VisibleEvent                      `json:"events"`
	Choice           *pendingChoice                                 `json:"choice,omitempty"`
	Declaration      *actionDeclaration                             `json:"declaration,omitempty"`
}

type pendingChoice struct {
	Actor   *model.Player           `json:"actor"`
	Options map[ViewHandle]entityID `json:"options"`
}

func (g *Game) initializeKnowledgeState() {
	knowledge := knowledgeState{
		Actions:          make(map[string]map[ViewHandle]constants.ActionKind, len(g.players)),
		Materializations: make(map[string]map[ViewHandle]cardInstanceID, len(g.players)),
		Activations:      make(map[string]map[ViewHandle]cardInstanceID, len(g.players)),
		Cards:            make(map[string]map[entityID]ViewHandle, len(g.players)),
		Events:           make(map[string][]VisibleEvent, len(g.players)),
	}
	for _, player := range g.players {
		knowledge.Actions[player.UID] = make(map[ViewHandle]constants.ActionKind)
		knowledge.Materializations[player.UID] = make(map[ViewHandle]cardInstanceID)
		knowledge.Activations[player.UID] = make(map[ViewHandle]cardInstanceID)
		knowledge.Cards[player.UID] = make(map[entityID]ViewHandle)
		knowledge.Events[player.UID] = []VisibleEvent{}
	}
	g.state.Knowledge = knowledge
	g.refreshLegalActions()
}

func (g *Game) refreshLegalActions() {
	for _, player := range g.players {
		actions := g.state.Knowledge.Actions[player.UID]
		materializations := g.state.Knowledge.Materializations[player.UID]
		activations := g.state.Knowledge.Activations[player.UID]
		clear(actions)
		clear(materializations)
		clear(activations)
		if g.state.Finished {
			continue
		}
		handle := g.newViewHandle(
			player,
			"action:concede",
		)
		actions[handle] = constants.ActionConcede
		if g.state.Knowledge.Choice == nil {
			switch {
			case samePlayer(player, g.state.Scheduler.OpportunityHolder):
				handle := g.newViewHandle(
					player,
					"action:pass",
				)
				actions[handle] = constants.ActionPass
				for _, card := range g.legalActionCards(player) {
					handle := g.newViewHandle(
						player,
						"action:activate:"+string(card),
					)
					activations[handle] = card
				}
			case samePlayer(player, g.state.Scheduler.TurnPlayer) && g.state.Scheduler.Phase == PhaseMaterialize:
				for _, card := range g.legalChampionMaterializations(player) {
					handle := g.newViewHandle(
						player,
						"action:materialize:"+string(card),
					)
					materializations[handle] = card
				}
				handle := g.newViewHandle(
					player,
					"action:skip-materialize",
				)
				actions[handle] = constants.ActionSkipMaterialize
			}
		}
	}
}

func (g *Game) hasAction(player *model.Player, kind constants.ActionKind) bool {
	for _, candidate := range g.state.Knowledge.Actions[player.UID] {
		if candidate == kind {
			return true
		}
	}
	return false
}

func (g *Game) legalActions(player *model.Player) []LegalAction {
	actions := g.state.Knowledge.Actions[player.UID]
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
	for handle, card := range g.state.Knowledge.Materializations[player.UID] {
		legalActions = append(
			legalActions,
			LegalAction{
				Handle:   handle,
				Kind:     constants.ActionMaterialize,
				CardName: g.state.Entities[entityID(card)].Name,
			},
		)
	}
	for handle, card := range g.state.Knowledge.Activations[player.UID] {
		legalActions = append(
			legalActions,
			LegalAction{
				Handle:   handle,
				Kind:     constants.ActionActivate,
				CardName: g.state.Entities[entityID(card)].Name,
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

func (g *Game) visibleChampions(_ *model.Player) []VisibleChampion {
	champions := make([]VisibleChampion, 0, len(g.state.Champions))
	for _, owner := range g.players {
		champion, exists := g.state.Champions[owner.UID]
		if !exists {
			continue
		}
		champions = append(
			champions,
			VisibleChampion{
				Owner:    owner,
				CardName: g.state.Entities[entityID(champion.Card)].Name,
				Rested:   champion.Rested,
				Taunt:    champion.TauntUntilTurn > g.state.Scheduler.TurnNumber,
			},
		)
	}
	return champions
}

func (g *Game) grantCardTracking(player *model.Player, card entityID) {
	if _, exists := g.state.Knowledge.Cards[player.UID][card]; exists {
		return
	}
	handle := g.newViewHandle(
		player,
		"card:"+string(card),
	)
	g.state.Knowledge.Cards[player.UID][card] = handle
}

func (g *Game) revokeCardTracking(player *model.Player, card entityID) {
	handle, exists := g.state.Knowledge.Cards[player.UID][card]
	if !exists {
		return
	}
	delete(g.state.Knowledge.Cards[player.UID], card)
	choice := g.state.Knowledge.Choice
	if choice == nil || !samePlayer(choice.Actor, player) {
		return
	}
	delete(choice.Options, handle)
	if len(choice.Options) == 0 {
		g.state.Knowledge.Choice = nil
	}
}

func (g *Game) visibleCards(player *model.Player) []VisibleCard {
	cards := g.state.Knowledge.Cards[player.UID]
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

func (g *Game) recordVisibleEvent(player *model.Player, kind string, card entityID) {
	event := VisibleEvent{
		Kind:     kind,
		CardName: g.state.Entities[card].Name,
	}
	g.state.Knowledge.Events[player.UID] = append(
		g.state.Knowledge.Events[player.UID],
		event,
	)
}

func (g *Game) visibleEvents(player *model.Player) []VisibleEvent {
	return append(
		[]VisibleEvent(nil),
		g.state.Knowledge.Events[player.UID]...,
	)
}

func (g *Game) setPendingCardChoice(player *model.Player, card entityID) {
	handle, exists := g.state.Knowledge.Cards[player.UID][card]
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

func (g *Game) pendingChoice(player *model.Player) *PendingChoice {
	choice := g.state.Knowledge.Choice
	if choice == nil || !samePlayer(choice.Actor, player) {
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

func (g *Game) submitChoice(player *model.Player, input Input) error {
	choice := g.state.Knowledge.Choice
	if choice == nil || !samePlayer(choice.Actor, player) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Choice)
	}
	subject, exists := choice.Options[input.Choice]
	if !exists {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, input.Choice)
	}
	if g.state.Knowledge.Declaration != nil {
		if err := g.submitActionDeclarationChoice(player, subject); err != nil {
			return err
		}
		g.advanceKnowledgeRevision()
		return nil
	}
	g.state.Knowledge.Choice = nil
	g.advanceKnowledgeRevision()
	return nil
}

func (g *Game) advanceKnowledgeRevision() {
	g.state.Revision++
	g.refreshLegalActions()
}

func (g *Game) newViewHandle(player *model.Player, subject string) ViewHandle {
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
