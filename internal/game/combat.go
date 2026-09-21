package game

import (
	"fmt"
	"go-tcg/internal/constants"
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

// legalAttackers 分別評估回合玩家目前可宣告攻擊的 Champion 與支援攻擊的 Ally。
// 輸入為持有行動機會的玩家；輸出為合法攻擊者的內部識別，無副作用。
func (g *Game) legalAttackers(player *model.Player) []objectID {
	scheduler := g.state.Scheduler
	if scheduler.TurnNumber == 1 || !samePlayer(scheduler.TurnPlayer, player) || scheduler.Phase != constants.PhaseMain || len(g.state.EffectsStack) != 0 {
		return nil
	}
	attackers := make([]objectID, 0, 1)
	champion, exists := g.state.Champions[player.UID]
	if exists && !champion.Rested {
		championCharacteristics := g.characteristicsFor(champion.ID)
		championTargets := g.attackTargets(player, champion.ID)
		if championCharacteristics.Power > 0 && len(championTargets) > 0 {
			attackers = append(attackers, champion.ID)
		}
	}
	for id, object := range g.state.Objects {
		if !samePlayer(object.Owner, player) || !g.canAttackWith(player, id) {
			continue
		}
		targets := g.attackTargets(player, id)
		if len(targets) > 0 {
			attackers = append(attackers, id)
		}
	}
	sort.Slice(
		attackers,
		func(first, second int) bool {
			return attackers[first] < attackers[second]
		},
	)
	return attackers
}

// canAttackWith 是所有 Ally 攻擊的中央 permission query；卡牌特例只透過 derived characteristics 改變 Pride 等限制。
// 輸入為控制者與候選 object；輸出為其是否為受控、awake、正 power 且 obey 的 Ally，無副作用。
func (g *Game) canAttackWith(player *model.Player, attacker objectID) bool {
	object, exists := g.state.Objects[attacker]
	if !exists || !samePlayer(object.Owner, player) || object.Rested || !containsString(object.Types, "ALLY") {
		return false
	}
	return g.characteristicsFor(attacker).Power > 0 && g.obeys(player, attacker)
}

func (g *Game) obeys(player *model.Player, ally objectID) bool {
	pride := g.characteristicsFor(ally).Pride
	if pride == 0 {
		return true
	}
	champion, exists := g.state.Champions[player.UID]
	return exists && g.state.Cards[champion.Card].Level >= int64(pride)
}

func (g *Game) legalAttackTargets(player *model.Player) []objectID {
	champion, exists := g.state.Champions[player.UID]
	if !exists {
		return nil
	}
	return g.attackTargets(player, champion.ID)
}

// attackTargets 以 attacker 的目前特性計算 player 可宣告或保留的攻擊目標。
// 回傳值已套用 stealth、true sight 與醒著 Taunt 的限制；查詢本身不改變遊戲狀態。
func (g *Game) attackTargets(player *model.Player, attacker objectID) []objectID {
	targets := make([]objectID, 0, len(g.state.Champions))
	if _, exists := g.cardForObject(attacker); !exists {
		return nil
	}
	attackerHasTrueSight := g.characteristicsFor(attacker).TrueSight
	tauntTargets := make([]objectID, 0, len(g.players))
	for _, opponent := range g.players {
		if samePlayer(opponent, player) {
			continue
		}
		champion, exists := g.state.Champions[opponent.UID]
		if exists {
			targets = append(targets, champion.ID)
			if !champion.Rested && champion.TauntUntilTurn > g.state.Scheduler.TurnNumber {
				tauntTargets = append(tauntTargets, champion.ID)
			}
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
	if len(tauntTargets) > 0 {
		sort.Slice(
			tauntTargets,
			func(first, second int) bool {
				return tauntTargets[first] < tauntTargets[second]
			},
		)
		return tauntTargets
	}
	return targets
}

// isLegalAttackTarget 回傳 target 是否仍在 attacker 的目前合法攻擊目標集合中。
// 它不改變遊戲狀態，供宣告、重導與 combat resolution 共用相同規則。
func (g *Game) isLegalAttackTarget(player *model.Player, attacker, target objectID) bool {
	return containsObject(g.attackTargets(player, attacker), target)
}

func (g *Game) beginAttack(player *model.Player, attacker objectID) error {
	if !containsObject(g.legalAttackers(player), attacker) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, attacker)
	}
	g.state.Knowledge.Attack = &attackDeclaration{Controller: player, Attacker: attacker}
	g.setDeclarationChoice(player, g.attackTargets(player, attacker))
	return nil
}

func (g *Game) canWield(player *model.Player, weapon objectID) bool {
	scheduler := g.state.Scheduler
	if !samePlayer(scheduler.OpportunityHolder, player) ||
		!samePlayer(scheduler.TurnPlayer, player) ||
		scheduler.Phase != constants.PhaseMain ||
		len(g.state.EffectsStack) != 0 ||
		!g.isLegalWeapon(player, weapon) {
		return false
	}
	legalTargets := g.legalTargets()
	if len(legalTargets) == 0 {
		return false
	}
	reserveCost := g.wieldReserveCost(weapon)
	return len(g.state.Zones[player.UID].Memory) >= reserveCost
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
	if attack == nil || !samePlayer(attack.Controller, player) || !g.isLegalAttackTarget(player, attack.Attacker, target) || !containsObject(g.legalAttackers(player), attack.Attacker) {
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
	g.state.Knowledge.Attack = nil
	g.state.Knowledge.Choice = nil
	g.flushTriggers(g.onAttackTriggers(player, attack.Attacker))
	if len(g.state.EffectsStack) == 1 {
		g.grantOpportunity(player)
	}
	return nil
}

func (g *Game) onAttackTriggers(player *model.Player, attacker objectID) []effectStackItem {
	card, exists := g.cardForObject(attacker)
	if !exists {
		return nil
	}
	if card.Definition == heatedVengeanceCardID && g.championHasClass(player, card.Classes) {
		champion, championExists := g.state.Champions[player.UID]
		if !championExists {
			return nil
		}
		ability := g.newAbilityInstance(
			player,
			card.ID,
			champion.ID,
			[]effectOperation{
				{
					Kind:    effectOperationChoose,
					Options: []objectID{champion.ID},
					CanPass: true,
				},
				{
					Kind:   effectOperationDamage,
					Amount: 3,
				},
			},
		)
		return []effectStackItem{
			{
				Kind:       effectStackAbility,
				Controller: player,
				Source:     card.ID,
				SourceLKI:  card.ID,
				Target:     champion.ID,
				Ability:    &ability,
			},
		}
	}
	if card.Definition != redHareCardID || !g.characteristicsFor(attacker).GrantedOnAttack || len(g.state.Zones[player.UID].Hand) == 0 {
		return nil
	}
	ability := g.newAbilityInstance(
		player,
		card.ID,
		"",
		[]effectOperation{
			{
				Kind:    effectOperationChooseHandCard,
				CanPass: true,
			},
			{
				Kind: effectOperationDiscard,
			},
			{
				Kind:   effectOperationDraw,
				Amount: 1,
			},
		},
	)
	return []effectStackItem{
		{
			Kind:       effectStackAbility,
			Controller: player,
			Source:     card.ID,
			SourceLKI:  card.ID,
			Ability:    &ability,
		},
	}
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
	if !attackerExists || !targetExists || attackerID == item.Target || !g.isLegalAttackTarget(item.Controller, attackerID, item.Target) {
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
