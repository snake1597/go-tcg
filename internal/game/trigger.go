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
			ability := g.newAbilityInstance(
				player,
				weapon,
				unit,
				[]effectOperation{
					{
						Kind:   effectOperationDamage,
						Amount: 3,
					},
				},
			)
			triggers = append(triggers, effectStackItem{
				Kind:       effectStackAbility,
				Controller: player,
				Source:     weapon,
				SourceLKI:  weapon,
				Target:     unit,
				Ability:    &ability,
			})
		}
	}
	g.state.Events = append(g.state.Events, batch)
	g.flushTriggers(triggers)
	return nil
}

// flushTriggers 將本批觸發送入效果堆疊；空批次不改變狀態。
// 單一觸發直接入堆疊並授予控制者行動機會；多個觸發建立排序選擇並暫停一般行動機會。
// 控制者取自首項，此函式不驗證整批是否屬於同一控制者。
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

// submitTriggerOrderChoice 將玩家選中的觸發入堆疊，移出待排序集合後提供剩餘選項。
// 選擇順序是入堆疊順序，後選的觸發會先結算；全數入堆疊後才清除選擇並授予行動機會。
// 此函式依賴既有 TriggerOrder 對應的 Choice，呼叫端須維持兩者一致。
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
