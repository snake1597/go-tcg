package game

import (
	"fmt"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
	"sort"
)

type attackDeclaration struct {
	Controller *model.Player `json:"controller"`
	Attacker   objectID      `json:"attacker"`
}

type wieldDeclaration struct {
	Controller *model.Player `json:"controller"`
	Weapon     objectID      `json:"weapon"`
}

func (g *Game) legalAttackers(player *model.Player) []objectID {
	scheduler := g.state.Scheduler
	if !samePlayer(scheduler.TurnPlayer, player) || scheduler.Phase != PhaseMain || len(g.state.EffectsStack) != 0 {
		return nil
	}
	champion, exists := g.state.Champions[player.UID]
	if !exists || champion.Rested || g.characteristicsFor(champion.ID).Power <= 0 || len(g.legalAttackTargets(player)) == 0 {
		return nil
	}
	attackers := []objectID{champion.ID}
	for id, object := range g.state.Objects {
		if samePlayer(object.Owner, player) && !object.Rested && g.state.Cards[object.Card].Definition == redHareCardID && g.redHareObeys(player, id) {
			attackers = append(attackers, id)
		}
	}
	return attackers
}

func (g *Game) legalAttackTargets(player *model.Player) []objectID {
	targets := make([]objectID, 0, len(g.state.Champions))
	champion, championExists := g.state.Champions[player.UID]
	attackerHasTrueSight := championExists && g.characteristicsFor(champion.ID).TrueSight
	for _, opponent := range g.players {
		if samePlayer(opponent, player) {
			continue
		}
		champion, exists := g.state.Champions[opponent.UID]
		if exists {
			targets = append(targets, champion.ID)
		}
	}
	for id, object := range g.state.Objects {
		if samePlayer(object.Owner, player) || !containsString(object.Types, "ALLY") || (!attackerHasTrueSight && g.characteristicsFor(id).Stealth) {
			continue
		}
		targets = append(targets, id)
	}
	sort.Slice(
		targets,
		func(first, second int) bool {
			return targets[first] < targets[second]
		},
	)
	return targets
}

func (g *Game) beginAttack(player *model.Player, attacker objectID) error {
	if !containsObject(g.legalAttackers(player), attacker) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, attacker)
	}
	g.state.Knowledge.Attack = &attackDeclaration{Controller: player, Attacker: attacker}
	g.setDeclarationChoice(player, g.legalAttackTargets(player))
	return nil
}

func (g *Game) canWield(player *model.Player, weapon objectID) bool {
	return samePlayer(g.state.Scheduler.OpportunityHolder, player) && g.isLegalWeapon(player, weapon) && len(g.legalTargets()) > 0 && len(g.state.Zones[player.UID].Memory) >= g.wieldReserveCost(weapon)
}

func (g *Game) wieldReserveCost(weapon objectID) int {
	return g.characteristicsFor(weapon).ReserveCost
}

func (g *Game) beginWield(player *model.Player, weapon objectID) error {
	if !g.canWield(player, weapon) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, weapon)
	}
	g.state.Knowledge.Wield = &wieldDeclaration{
		Controller: player,
		Weapon:     weapon,
	}
	g.setDeclarationChoice(player, g.legalTargets())
	return nil
}

func (g *Game) submitAttackChoice(player *model.Player, subject entityID) error {
	attack := g.state.Knowledge.Attack
	target := objectID(subject)
	if attack == nil || !samePlayer(attack.Controller, player) || !containsObject(g.legalAttackTargets(player), target) || !containsObject(g.legalAttackers(player), attack.Attacker) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
	}
	attacker, exists := g.cardForObject(attack.Attacker)
	if !exists {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, attack.Attacker)
	}
	if champion, championExists := g.state.Champions[player.UID]; championExists && champion.ID == attack.Attacker {
		champion.Rested = true
		champion.CombatRole = "attacker"
		g.state.Champions[player.UID] = champion
	} else {
		object := g.state.Objects[attack.Attacker]
		object.Rested = true
		g.state.Objects[attack.Attacker] = object
	}
	g.state.EffectsStack = append(g.state.EffectsStack, effectStackItem{
		Kind:       effectStackCombat,
		Controller: player,
		Source:     attacker.ID,
		Target:     target,
		Attacker:   attack.Attacker,
	})
	if attacker.Definition == redHareCardID && len(g.state.Zones[player.UID].Hand) > 0 {
		g.pushAbility(g.newAbilityInstance(
			player,
			attacker.ID,
			"",
			[]effectOperation{
				{
					Kind: effectOperationChooseHandCard,
				},
				{
					Kind: effectOperationDiscard,
				},
				{
					Kind:   effectOperationDraw,
					Amount: 1,
				},
			},
		))
	}
	g.state.Knowledge.Attack = nil
	g.state.Knowledge.Choice = nil
	g.grantOpportunity(player)
	return nil
}

func (g *Game) submitWieldChoice(player *model.Player, subject entityID) error {
	declaration := g.state.Knowledge.Wield
	target := objectID(subject)
	if declaration == nil || !samePlayer(declaration.Controller, player) || !g.canWield(player, declaration.Weapon) || !g.isLegalTarget(target) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, subject)
	}
	g.payWieldReserve(player, declaration.Weapon)
	if err := g.wieldWeapon(player, target, declaration.Weapon); err != nil {
		return err
	}
	g.state.Knowledge.Wield = nil
	g.state.Knowledge.Choice = nil
	return nil
}

func (g *Game) payWieldReserve(player *model.Player, weapon objectID) {
	zones := g.state.Zones[player.UID]
	for payment := 0; payment < g.wieldReserveCost(weapon); payment++ {
		lastIndex := len(zones.Memory) - 1
		zones.Banishment = append(zones.Banishment, zones.Memory[lastIndex])
		zones.Memory = zones.Memory[:lastIndex]
	}
	g.state.Zones[player.UID] = zones
}

// resolveCombat 先取得雙方目前攻擊力，再施加雙向傷害並記錄同時發生的戰鬥事件。
// 雙方傷害完成後才執行狀態檢查，避免先受傷的一方死亡而失去反擊。
// 任一單位不存在或雙方為同一物件時不結算；未指定攻擊者時使用控制者的 champion。
func (g *Game) resolveCombat(item effectStackItem) {
	attackerID := item.Attacker
	if attackerID == "" {
		attackerID = g.state.Champions[item.Controller.UID].ID
	}
	attacker, attackerExists := g.cardForObject(attackerID)
	target, targetExists := g.cardForObject(item.Target)
	if !attackerExists || !targetExists || attackerID == item.Target {
		return
	}
	attackerPower := g.characteristicsFor(attackerID).Power
	targetPower := g.characteristicsFor(item.Target).Power
	g.damageUnit(attackerID, targetPower)
	g.damageUnit(item.Target, attackerPower)
	g.recordCombatDamage(item.Controller, attacker.ID, target.ID)
	g.resolveCombatStateBasedWithCause("combat:on-kill", "combat:damage")
}

func (g *Game) recordCombatDamage(player *model.Player, attacker, target cardInstanceID) {
	g.state.NextEvent++
	attackSequence := g.state.NextEvent
	g.state.NextEvent++
	retaliationSequence := g.state.NextEvent
	g.state.Events = append(g.state.Events, eventBatch{
		Player:       player,
		Cause:        "combat:damage",
		ParentFlow:   "attack",
		Simultaneous: true,
		Events: []gameEvent{
			{
				Sequence: attackSequence,
				Kind:     "on-hit:attack",
				Card:     attacker,
			},
			{
				Sequence: retaliationSequence,
				Kind:     "on-hit:retaliation",
				Card:     target,
			},
		},
	})
	for _, viewer := range g.players {
		g.recordVisibleEvent(viewer, "combat-damage", entityID(attacker))
		g.recordVisibleEvent(viewer, "retaliation-damage", entityID(target))
	}
}

func (g *Game) resolveCombatStateBased() {
	g.resolveCombatStateBasedWithCause("", "")
}

// resolveCombatStateBasedWithCause 重複執行死亡與勝負檢查，直到狀態不再改變或對局結束。
// 移除單位可能使其他單位失去靜態加成，因此需要再次檢查。
// 超過 32 輪仍未穩定時結束對局並設定 Diagnostic，供 replay 與狀態雜湊追查。
func (g *Game) resolveCombatStateBasedWithCause(cause, parentFlow string) {
	for pass := 0; pass < 32; pass++ {
		if !g.resolveCombatStateBasedPass(cause, parentFlow) {
			g.flushTriggers(nil)
			return
		}
		if g.state.Finished {
			g.flushTriggers(nil)
			return
		}
	}
	g.state.Finished = true
	g.state.Scheduler = schedulerFrame{
		Kind: schedulerFinished,
	}
	g.state.Diagnostic = "combat state-based checks exceeded 32 passes; inspect replay and state hash"
	g.flushTriggers(nil)
}

func (g *Game) resolveCombatStateBasedPass(cause, parentFlow string) bool {
	changed := false
	objectIDs := make([]objectID, 0, len(g.state.Objects))
	for id := range g.state.Objects {
		objectIDs = append(objectIDs, id)
	}
	sort.Slice(objectIDs, func(first, second int) bool {
		return objectIDs[first] < objectIDs[second]
	})
	for _, id := range objectIDs {
		object := g.state.Objects[id]
		if !containsString(object.Types, "ALLY") || g.isImmortal(id) || g.characteristicsFor(id).Life <= 0 || object.Damage < g.characteristicsFor(id).Life {
			continue
		}
		delete(g.state.Objects, id)
		g.putInGraveyard(object.Card)
		g.enqueueVeritaDeath(object.Owner, object.Card)
		g.recordCombatStateBasedEvent(cause, parentFlow, "destroy", object.Card)
		changed = true
	}
	for _, player := range g.players {
		playerID := player.UID
		champion, exists := g.state.Champions[playerID]
		if !exists {
			continue
		}
		if g.characteristicsFor(champion.ID).Life <= 0 || champion.Damage < g.characteristicsFor(champion.ID).Life {
			continue
		}
		g.state.Finished = true
		g.state.Winner = g.otherPlayer(champion.Owner)
		g.state.Scheduler = schedulerFrame{Kind: schedulerFinished}
		delete(g.state.Champions, playerID)
		g.recordCombatStateBasedEvent(cause, parentFlow, "on-kill", champion.Card)
		return true
	}
	return changed
}

func (g *Game) recordCombatStateBasedEvent(cause, parentFlow, kind string, card cardInstanceID) {
	if cause == "" {
		return
	}
	g.state.NextEvent++
	g.state.Events = append(g.state.Events, eventBatch{
		Cause:      cause,
		ParentFlow: parentFlow,
		Events: []gameEvent{
			{
				Sequence: g.state.NextEvent,
				Kind:     kind,
				Card:     card,
			},
		},
	})
	for _, viewer := range g.players {
		g.recordVisibleEvent(viewer, kind, entityID(card))
	}
}

func containsObject(objects []objectID, want objectID) bool {
	for _, object := range objects {
		if object == want {
			return true
		}
	}
	return false
}
