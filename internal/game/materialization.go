package game

import (
	"fmt"
	"go-tcg/internal/constants"
	tcgErrors "go-tcg/internal/tcg_errors"
)

const (
	spiritOfFireCardID  CardID = "LMyKyVC2O9"
	tonorisCardID       CardID = "zb14m4c8lj"
	tonorisOnEnterCause        = "ability:zb14m4c8lj:front:on-enter-taunt"
)

type effectStackItemKind string

const (
	effectStackMaterialization effectStackItemKind = "materialization"
	effectStackTonorisTaunt    effectStackItemKind = "tonoris_on_enter_taunt"
)

type effectStackItem struct {
	Kind       effectStackItemKind `json:"kind"`
	Controller constants.PlayerID  `json:"controller"`
	Source     cardInstanceID      `json:"source"`
}

func (g *Game) legalChampionMaterializations(player constants.PlayerID) []cardInstanceID {
	zones := g.state.Zones[player]
	materialDeckSize := len(zones.MaterialDeck)
	cards := make([]cardInstanceID, 0, materialDeckSize)
	for _, card := range zones.MaterialDeck {
		if g.canMaterializeChampion(
			player,
			card,
		) {
			cards = append(cards, card)
		}
	}
	return cards
}

func (g *Game) canMaterializeChampion(player constants.PlayerID, card cardInstanceID) bool {
	scheduler := g.state.Scheduler
	if scheduler.Kind != schedulerStable || scheduler.Phase != PhaseMaterialize || scheduler.TurnPlayer != player || scheduler.OpportunityHolder != "" {
		return false
	}
	candidate, exists := g.state.Cards[card]
	if !exists || candidate.Owner != player || candidate.Definition != tonorisCardID {
		return false
	}
	champion, exists := g.state.Champions[player]
	if !exists {
		return false
	}
	current, exists := g.state.Cards[champion.Card]
	if !exists || current.Definition != spiritOfFireCardID || candidate.Level != current.Level+1 {
		return false
	}
	return len(g.state.Zones[player].Memory) >= candidate.MemoryCost
}

func (g *Game) materializeChampion(player constants.PlayerID, card cardInstanceID) error {
	if !g.canMaterializeChampion(
		player,
		card,
	) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	zones := g.state.Zones[player]
	materialDeckIndex := cardIndex(
		zones.MaterialDeck,
		card,
	)
	if materialDeckIndex < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	return g.payChampionMaterialization(player, card)
}

func (g *Game) payChampionMaterialization(player constants.PlayerID, card cardInstanceID) error {
	if !g.canMaterializeChampion(
		player,
		card,
	) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	zones := g.state.Zones[player]
	materialDeckIndex := cardIndex(
		zones.MaterialDeck,
		card,
	)
	if materialDeckIndex < 0 {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, card)
	}
	candidate := g.state.Cards[card]
	zones.MaterialDeck = removeCardAt(
		zones.MaterialDeck,
		materialDeckIndex,
	)
	for paymentCount := 0; paymentCount < candidate.MemoryCost; paymentCount++ {
		memoryCount := len(zones.Memory)
		randomValue := g.nextRandom()
		paymentIndex := int(randomValue % uint64(memoryCount))
		payment := zones.Memory[paymentIndex]
		zones.Memory = removeCardAt(zones.Memory, paymentIndex)
		zones.Banishment = append(zones.Banishment, payment)
	}
	g.state.Zones[player] = zones
	g.state.EffectSources = append(g.state.EffectSources, card)
	g.state.EffectsStack = append(
		g.state.EffectsStack,
		effectStackItem{
			Kind:       effectStackMaterialization,
			Controller: player,
			Source:     candidate.ID,
		},
	)
	g.grantOpportunity(player)
	return nil
}

func cardIndex(cards []cardInstanceID, want cardInstanceID) int {
	for index, card := range cards {
		if card == want {
			return index
		}
	}
	return -1
}

func removeCardAt(cards []cardInstanceID, index int) []cardInstanceID {
	return append(
		cards[:index],
		cards[index+1:]...,
	)
}

func (g *Game) resolveTopEffectStack() {
	lastIndex := len(g.state.EffectsStack) - 1
	item := g.state.EffectsStack[lastIndex]
	g.state.EffectsStack = g.state.EffectsStack[:lastIndex]
	switch item.Kind {
	case effectStackMaterialization:
		g.resolveChampionLevelUp(item)
	case effectStackTonorisTaunt:
		g.resolveTonorisTaunt(item)
	default:
		panic(fmt.Sprintf("unknown effect stack item %q", item.Kind))
	}
}

func (g *Game) resolveChampionLevelUp(item effectStackItem) {
	sourceIndex := cardIndex(g.state.EffectSources, item.Source)
	if sourceIndex < 0 {
		return
	}
	g.state.EffectSources = removeCardAt(g.state.EffectSources, sourceIndex)
	if !g.canResolveChampionLevelUp(item) {
		zones := g.state.Zones[item.Controller]
		zones.Banishment = append(zones.Banishment, item.Source)
		g.state.Zones[item.Controller] = zones
		return
	}
	champion := g.state.Champions[item.Controller]
	previousTop := champion.Card
	sourceEntity := entityID(item.Source)
	champion.Card = item.Source
	champion.InnerLineage = append(champion.InnerLineage, previousTop)
	g.state.Champions[item.Controller] = champion
	for _, player := range g.players {
		g.grantCardTracking(player, sourceEntity)
	}
	g.recordPublicEvent(
		item.Controller,
		"materialize:level-up",
		"level-up",
		item.Source,
	)
	g.state.EffectsStack = append(
		g.state.EffectsStack,
		effectStackItem{
			Kind:       effectStackTonorisTaunt,
			Controller: item.Controller,
			Source:     item.Source,
		},
	)
}

func (g *Game) canResolveChampionLevelUp(item effectStackItem) bool {
	candidate, exists := g.state.Cards[item.Source]
	if !exists || candidate.Owner != item.Controller || candidate.Definition != tonorisCardID {
		return false
	}
	champion, exists := g.state.Champions[item.Controller]
	if !exists {
		return false
	}
	current, exists := g.state.Cards[champion.Card]
	if !exists || current.Definition != spiritOfFireCardID {
		return false
	}
	return candidate.Level == current.Level+1
}

func (g *Game) resolveTonorisTaunt(item effectStackItem) {
	champion := g.state.Champions[item.Controller]
	playerCount := len(g.players)
	champion.TauntUntilTurn = g.state.Scheduler.TurnNumber + uint64(playerCount)
	g.state.Champions[item.Controller] = champion
	g.recordPublicEvent(
		item.Controller,
		tonorisOnEnterCause,
		"taunt-granted",
		item.Source,
	)
}

func (g *Game) expireTimedChampionEffects() {
	for _, player := range g.players {
		champion := g.state.Champions[player]
		if champion.TauntUntilTurn == 0 || champion.TauntUntilTurn > g.state.Scheduler.TurnNumber {
			continue
		}
		champion.TauntUntilTurn = 0
		g.state.Champions[player] = champion
		g.recordPublicEvent(
			player,
			tonorisOnEnterCause,
			"taunt-expired",
			champion.Card,
		)
	}
}

func (g *Game) recordPublicEvent(player constants.PlayerID, cause, kind string, card cardInstanceID) {
	cardEntity := entityID(card)
	g.state.NextEvent++
	g.state.Events = append(
		g.state.Events,
		eventBatch{
			Player: player,
			Cause:  cause,
			Events: []gameEvent{
				{
					Sequence: g.state.NextEvent,
					Kind:     kind,
					Card:     card,
				},
			},
		},
	)
	for _, viewer := range g.players {
		g.recordVisibleEvent(viewer, kind, cardEntity)
	}
}
