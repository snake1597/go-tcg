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
	effectOperationChoose                effectOperationKind = "choose"
	effectOperationMove                  effectOperationKind = "move"
	effectOperationDraw                  effectOperationKind = "draw"
	effectOperationDrawToMemory          effectOperationKind = "draw_to_memory"
	effectOperationCounter               effectOperationKind = "counter"
	effectOperationDamage                effectOperationKind = "damage"
	effectOperationContinuousModifier    effectOperationKind = "continuous_modifier"
	effectOperationChooseHandCard        effectOperationKind = "choose_hand_card"
	effectOperationChooseMemoryAlly      effectOperationKind = "choose_memory_ally"
	effectOperationDiscard               effectOperationKind = "discard"
	effectOperationDeploy                effectOperationKind = "deploy"
	effectOperationSuitedThresholdDamage effectOperationKind = "suited_threshold_damage"
	effectOperationChooseDuchessCopy     effectOperationKind = "choose_duchess_copy"
	effectOperationCopyDuchessAction     effectOperationKind = "copy_duchess_action"
	effectOperationSacrificeForChef      effectOperationKind = "sacrifice_for_chef"
	effectOperationRetargetAttack        effectOperationKind = "retarget_attack"
)

// effectOperation contains only replayable values. A card ability is compiled
// into this sequence at declaration, after its legal target is fixed.
type effectOperation struct {
	Kind                      effectOperationKind `json:"kind"`
	Target                    objectID            `json:"target,omitempty"`
	Source                    objectID            `json:"source,omitempty"`
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
			if !g.isLegalTarget(target) {
				continue
			}
			effect := operation.ContinuousEffect
			effect.Controller = instance.Controller
			effect.Source = objectID(instance.SourceLKI)
			effect.Target = target
			g.addContinuousEffect(effect)
			g.recordPublicEvent(instance.Controller, "ability", "continuous-effect", instance.SourceLKI)
		case effectOperationDraw:
			g.drawCards(instance.Controller, operation.Amount)
		case effectOperationDrawToMemory:
			g.drawToMemory(instance.Controller, operation.Amount)
		case effectOperationCounter:
			if g.addCounter(target, operation.Counter, operation.Amount) {
				g.recordPublicEvent(instance.Controller, "ability", "counter", instance.SourceLKI)
			}
		case effectOperationChooseHandCard:
			g.beginAbilityCardChoice(
				instance,
				operationIndex,
				instance.Controller,
				g.state.Zones[instance.Controller.UID].Hand,
			)
			return
		case effectOperationChooseMemoryAlly:
			g.beginAbilityCardChoice(
				instance,
				operationIndex,
				instance.Controller,
				g.qualifiedMemoryAllies(instance.Controller),
			)
			return
		case effectOperationChooseDuchessCopy:
			g.beginAbilityCardChoice(
				instance,
				operationIndex,
				instance.Controller,
				g.eligibleDuchessCopies(instance.Controller),
			)
			return
		case effectOperationCopyDuchessAction:
			copied, err := g.copyDuchessAction(
				instance.Controller,
				cardInstanceID(target),
				"",
			)
			if err != nil {
				return
			}
			chooseTarget := effectOperation{
				Kind:    effectOperationChoose,
				Options: g.legalTargets(),
			}
			copied.Operations = append(
				[]effectOperation{
					chooseTarget,
				},
				copied.Operations...,
			)
			g.pushAbility(copied)
			return
		case effectOperationSacrificeForChef:
			if err := g.sacrificeForPepperedChef(
				instance.Controller,
				operation.Source,
				target,
			); err != nil {
				return
			}
		case effectOperationRetargetAttack:
			if err := g.retargetAttackWithTrumpSet(instance.Controller, target); err != nil {
				return
			}
		case effectOperationDiscard:
			g.discardCard(instance.Controller, cardInstanceID(target))
		case effectOperationDeploy:
			g.deployAlly(instance.Controller, cardInstanceID(target))
		case effectOperationSuitedThresholdDamage:
			amount := suitedThresholdAmount(g.suitedReserveTotal(instance.Controller)) * 2
			if amount > 0 && g.isLegalTarget(target) {
				g.damageUnit(target, amount)
				g.recordPublicEvent(instance.Controller, "ability", "damage", instance.SourceLKI)
			}
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

func (g *Game) beginAbilityCardChoice(instance abilityInstance, operationIndex int, player *model.Player, cards []cardInstanceID) {
	if len(cards) == 0 {
		return
	}
	options := make([]objectID, 0, len(cards))
	for _, card := range cards {
		options = append(options, objectID(card))
	}
	g.state.AbilityChoice = &abilityChoice{Instance: instance, Operations: append([]effectOperation(nil), instance.Operations[operationIndex+1:]...)}
	g.setDeclarationChoice(player, options)
}

func (g *Game) discardCard(player *model.Player, card cardInstanceID) {
	zones := g.state.Zones[player.UID]
	index := cardIndex(zones.Hand, card)
	if index < 0 {
		return
	}
	zones.Hand = removeCardAt(zones.Hand, index)
	zones.Graveyard = append(zones.Graveyard, card)
	g.state.Zones[player.UID] = zones
	g.recordPublicEvent(player, "ability", "discard", card)
}

func (g *Game) drawToMemory(player *model.Player, amount int) {
	zones := g.state.Zones[player.UID]
	for draw := 0; draw < amount && len(zones.MainDeck) > 0; draw++ {
		card := zones.MainDeck[0]
		zones.MainDeck = removeCardAt(zones.MainDeck, 0)
		zones.Memory = append(zones.Memory, card)
		g.recordPublicEvent(player, "ability", "draw-to-memory", card)
	}
	g.state.Zones[player.UID] = zones
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

func (g *Game) addCounter(target objectID, counter string, amount int) bool {
	object, exists := g.state.Objects[target]
	if !exists {
		return false
	}
	if object.Counters == nil {
		object.Counters = make(map[string]int)
	}
	object.Counters[counter] += amount
	g.state.Objects[target] = object
	return true
}
