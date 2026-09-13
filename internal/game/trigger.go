package game

import (
	"fmt"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
)

type triggerOrder struct {
	Controller *model.Player     `json:"controller"`
	Triggers   []effectStackItem `json:"triggers"`
}

func (g *Game) wieldWeapon(player *model.Player, unit, weapon objectID) error {
	if !g.isLegalWeapon(player, weapon) || !g.isLegalTarget(unit) {
		return fmt.Errorf("%w", tcgErrors.ErrInvalidViewHandle)
	}
	weaponCard := g.state.Objects[weapon].Card
	return g.wieldWeapons(player, unit, []cardInstanceID{weaponCard})
}

func (g *Game) wieldWeapons(player *model.Player, unit objectID, weapons []cardInstanceID) error {
	if !g.isLegalTarget(unit) {
		return fmt.Errorf("%w", tcgErrors.ErrInvalidViewHandle)
	}
	batch := eventBatch{
		Player:       player,
		Cause:        "wield",
		ParentFlow:   "wield",
		Simultaneous: len(weapons) > 1,
		Events:       make([]gameEvent, 0, len(weapons)),
	}
	triggers := make([]effectStackItem, 0, len(weapons))
	for _, weapon := range weapons {
		g.state.NextEvent++
		batch.Events = append(batch.Events, gameEvent{
			Sequence: g.state.NextEvent,
			Kind:     "wield",
			Card:     weapon,
		})
		if g.state.Cards[weapon].Definition == impactHammerCardID {
			triggers = append(triggers, effectStackItem{
				Kind:       effectStackImpactHammer,
				Controller: player,
				Source:     weapon,
				SourceLKI:  weapon,
				Target:     unit,
			})
		}
	}
	g.state.Events = append(g.state.Events, batch)
	g.flushTriggers(triggers)
	return nil
}

func (g *Game) flushTriggers(triggers []effectStackItem) {
	if len(triggers) == 0 {
		return
	}
	controller := triggers[0].Controller
	if len(triggers) == 1 {
		g.state.EffectsStack = append(g.state.EffectsStack, triggers[0])
		g.grantOpportunity(controller)
		return
	}
	g.state.Knowledge.TriggerOrder = &triggerOrder{
		Controller: controller,
		Triggers:   triggers,
	}
	g.state.Scheduler.OpportunityHolder = nil
	g.state.Scheduler.ConsecutivePasses = 0
	g.setTriggerOrderChoice()
}

func (g *Game) setTriggerOrderChoice() {
	order := g.state.Knowledge.TriggerOrder
	options := make(map[ViewHandle]entityID, len(order.Triggers))
	for _, trigger := range order.Triggers {
		handle := g.newViewHandle(order.Controller, "trigger-order:"+string(trigger.Source))
		options[handle] = entityID(trigger.Source)
	}
	g.state.Knowledge.Choice = &pendingChoice{
		Actor:   order.Controller,
		Options: options,
	}
}

func (g *Game) submitTriggerOrderChoice(player *model.Player, handle ViewHandle) error {
	order := g.state.Knowledge.TriggerOrder
	if order == nil || !samePlayer(order.Controller, player) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, handle)
	}
	source, exists := g.state.Knowledge.Choice.Options[handle]
	if !exists {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, handle)
	}
	for index, trigger := range order.Triggers {
		if entityID(trigger.Source) != source {
			continue
		}
		g.state.EffectsStack = append(g.state.EffectsStack, trigger)
		order.Triggers = append(order.Triggers[:index], order.Triggers[index+1:]...)
		break
	}
	if len(order.Triggers) > 0 {
		g.setTriggerOrderChoice()
		return nil
	}
	g.state.Knowledge.Choice = nil
	g.state.Knowledge.TriggerOrder = nil
	g.grantOpportunity(player)
	return nil
}
