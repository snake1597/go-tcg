package game

import (
	"fmt"

	"go-tcg/internal/model"
)

type abilityInstanceID uint64

type abilityInstance struct {
	ID         abilityInstanceID `json:"id"`
	Controller *model.Player     `json:"controller"`
	Source     cardInstanceID    `json:"source"`
	SourceLKI  cardInstanceID    `json:"source_lki"`
	Target     objectID          `json:"target,omitempty"`
	Operations []effectOperation `json:"operations"`
}

type effectOperationKind string

const (
	effectOperationChoose             effectOperationKind = "choose"
	effectOperationMove               effectOperationKind = "move"
	effectOperationDraw               effectOperationKind = "draw"
	effectOperationCounter            effectOperationKind = "counter"
	effectOperationDamage             effectOperationKind = "damage"
	effectOperationContinuousModifier effectOperationKind = "continuous_modifier"
)

// effectOperation contains only replayable values. A card ability is compiled
// into this sequence at declaration, after its legal target is fixed.
type effectOperation struct {
	Kind                      effectOperationKind `json:"kind"`
	Target                    objectID            `json:"target,omitempty"`
	Amount                    int                 `json:"amount,omitempty"`
	Counter                   string              `json:"counter,omitempty"`
	MoveSourceToGraveyard     bool                `json:"move_source_to_graveyard,omitempty"`
	DistinctSuitedCostsDamage bool                `json:"distinct_suited_costs_damage,omitempty"`
	ContinuousEffect          continuousEffect    `json:"continuous_effect,omitempty"`
	Options                   []objectID          `json:"options,omitempty"`
}

type abilityChoice struct {
	Instance   abilityInstance   `json:"instance"`
	Operations []effectOperation `json:"operations"`
}

func (g *Game) newAbilityInstance(controller *model.Player, source cardInstanceID, target objectID, operations []effectOperation) abilityInstance {
	g.state.NextAbility++
	return abilityInstance{
		ID:         abilityInstanceID(g.state.NextAbility),
		Controller: controller,
		Source:     source,
		SourceLKI:  source,
		Target:     target,
		Operations: operations,
	}
}

func (g *Game) pushAbility(instance abilityInstance) {
	g.state.EffectsStack = append(g.state.EffectsStack, effectStackItem{
		Kind:       effectStackAbility,
		Controller: instance.Controller,
		Source:     instance.Source,
		SourceLKI:  instance.SourceLKI,
		Target:     instance.Target,
		Ability:    &instance,
	})
}

func (g *Game) resolveAbility(instance abilityInstance) {
	for operationIndex, operation := range instance.Operations {
		target := operation.Target
		if target == "" {
			target = instance.Target
		}
		switch operation.Kind {
		case effectOperationDamage:
			amount := operation.Amount
			if operation.DistinctSuitedCostsDamage {
				amount += g.distinctSuitedPrintedReserveCosts(instance.Controller)
			}
			if g.isLegalTarget(target) {
				g.damageUnit(target, amount)
				g.recordPublicEvent(instance.Controller, "ability", "damage", instance.SourceLKI)
			}
		case effectOperationContinuousModifier:
			effect := operation.ContinuousEffect
			effect.Controller = instance.Controller
			effect.Source = objectID(instance.SourceLKI)
			effect.Target = target
			g.addContinuousEffect(effect)
			g.recordPublicEvent(instance.Controller, "ability", "continuous-effect", instance.SourceLKI)
		case effectOperationDraw:
			g.drawCards(instance.Controller, operation.Amount)
		case effectOperationCounter:
			g.addCounter(target, operation.Counter, operation.Amount)
		case effectOperationMove:
			if operation.MoveSourceToGraveyard {
				g.putInGraveyard(instance.Source)
			}
		case effectOperationChoose:
			if len(operation.Options) == 0 {
				return
			}
			g.state.AbilityChoice = &abilityChoice{
				Instance:   instance,
				Operations: append([]effectOperation(nil), instance.Operations[operationIndex+1:]...),
			}
			g.setDeclarationChoice(instance.Controller, operation.Options)
			return
		default:
			panic(fmt.Sprintf("unknown effect operation %q", operation.Kind))
		}
	}
}

func (g *Game) drawCards(player *model.Player, amount int) {
	zones := g.state.Zones[player.UID]
	for draw := 0; draw < amount && len(zones.MainDeck) > 0; draw++ {
		card := zones.MainDeck[0]
		zones.MainDeck = removeCardAt(zones.MainDeck, 0)
		zones.Hand = append(zones.Hand, card)
		g.recordPublicEvent(player, "ability", "draw", card)
	}
	g.state.Zones[player.UID] = zones
}

func (g *Game) addCounter(target objectID, counter string, amount int) {
	object, exists := g.state.Objects[target]
	if !exists {
		return
	}
	if object.Counters == nil {
		object.Counters = make(map[string]int)
	}
	object.Counters[counter] += amount
	g.state.Objects[target] = object
}
