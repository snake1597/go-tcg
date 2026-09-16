package game

import (
	"fmt"
	"sort"

	"go-tcg/internal/model"
)

type replacementKind string

const (
	replacementDamagePrevent replacementKind = "damage_prevent"
	replacementRecoverReduce replacementKind = "recover_reduce"
)

type replacementIntentKind string

const (
	replacementIntentDamage  replacementIntentKind = "damage"
	replacementIntentRecover replacementIntentKind = "recover"
)

type replacementCause struct {
	Kind   string   `json:"kind"`
	Source objectID `json:"source,omitempty"`
	Amount int      `json:"amount"`
}

type replacementIntent struct {
	Kind     replacementIntentKind `json:"kind"`
	Target   objectID              `json:"target"`
	Affected *model.Player         `json:"affected"`
	Amount   int                   `json:"amount"`
	Combat   bool                  `json:"combat,omitempty"`
	Source   cardInstanceID        `json:"source,omitempty"`
	Cause    string                `json:"cause"`
	Chain    []replacementCause    `json:"chain"`
	Applied  []objectID            `json:"applied,omitempty"`
}

type replacementEffect struct {
	Kind          replacementKind `json:"kind"`
	Source        objectID        `json:"source"`
	Controller    *model.Player   `json:"controller"`
	Target        objectID        `json:"target,omitempty"`
	ExpiresAtTurn uint64          `json:"expires_at_turn,omitempty"`
}

type replacementChoice struct {
	Intent       replacementIntent `json:"intent"`
	Continuation abilityInstance   `json:"continuation"`
}

// activateSafeguardAmulet 放逐己方場上的 Amulet，建立本回合對己方 Champion 的非戰鬥傷害 prevention。
// 輸入為控制者與 Amulet 物件；成功時移除物件、放逐卡牌並加入 delayed replacement。
func (g *Game) activateSafeguardAmulet(player *model.Player, source objectID) error {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) || g.state.Cards[object.Card].Definition != safeguardAmuletCardID {
		return fmt.Errorf("invalid Safeguard Amulet activation")
	}
	champion, exists := g.state.Champions[player.UID]
	if !exists {
		return fmt.Errorf("Safeguard Amulet controller has no champion")
	}
	delete(g.state.Objects, source)
	zones := g.state.Zones[player.UID]
	zones.Banishment = append(zones.Banishment, object.Card)
	g.state.Zones[player.UID] = zones
	effect := replacementEffect{
		Kind:          replacementDamagePrevent,
		Source:        source,
		Controller:    player,
		Target:        champion.ID,
		ExpiresAtTurn: g.state.Scheduler.TurnNumber + 1,
	}
	g.state.ReplacementEffects = append(
		g.state.ReplacementEffects,
		effect,
	)
	g.recordPublicEvent(player, "ability", "banish", object.Card)
	return nil
}

// expireReplacementEffects 移除已到期的 delayed replacement，避免它們殘留在後續 canonical state。
// 輸入使用目前回合數判定；輸出保留未到期效果，副作用是重建 ReplacementEffects 切片。
func (g *Game) expireReplacementEffects() {
	active := make([]replacementEffect, 0, len(g.state.ReplacementEffects))
	for _, effect := range g.state.ReplacementEffects {
		if effect.ExpiresAtTurn > 0 && effect.ExpiresAtTurn <= g.state.Scheduler.TurnNumber {
			continue
		}
		active = append(
			active,
			effect,
		)
	}
	g.state.ReplacementEffects = active
}

// applyReplacementIntent 重新計算候選並逐項套用；完成時提交候選，或建立 resolution-scoped PendingChoice。
// continuation 為 nil 時供同步內部呼叫使用；有 continuation 時多候選由受影響玩家透過選擇處理。
func (g *Game) applyReplacementIntent(intent replacementIntent, continuation *abilityInstance) bool {
	if intent.Amount <= 0 {
		return g.commitReplacementIntent(intent)
	}
	candidates := g.replacementCandidates(intent)
	if len(candidates) == 0 {
		return g.commitReplacementIntent(intent)
	}
	if len(candidates) == 1 || continuation == nil {
		g.applyReplacement(&intent, candidates[0])
		return g.applyReplacementIntent(intent, continuation)
	}
	options := make([]objectID, 0, len(candidates))
	for _, candidate := range candidates {
		options = append(options, candidate.Source)
	}
	g.state.ReplacementChoice = &replacementChoice{
		Intent:       intent,
		Continuation: *continuation,
	}
	g.setDeclarationChoice(intent.Affected, options)
	return false
}

// submitReplacementChoice 套用受影響玩家選定的 replacement，重算後續候選，完成時恢復被中斷的能力。
// 輸入為選擇者與 replacement source；無效來源回傳錯誤，成功時可能建立下一個 PendingChoice。
func (g *Game) submitReplacementChoice(player *model.Player, source objectID) error {
	choice := g.state.ReplacementChoice
	if choice == nil || !samePlayer(choice.Intent.Affected, player) {
		return fmt.Errorf("invalid replacement choice")
	}
	var selected *replacementEffect
	for _, candidate := range g.replacementCandidates(choice.Intent) {
		if candidate.Source == source {
			selected = &candidate
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("invalid replacement source %q", source)
	}
	intent := choice.Intent
	continuation := choice.Continuation
	g.state.ReplacementChoice = nil
	g.state.Knowledge.Choice = nil
	g.applyReplacement(&intent, *selected)
	if !g.applyReplacementIntent(intent, &continuation) {
		return nil
	}
	g.pushAbility(continuation)
	g.grantOpportunity(continuation.Controller)
	return nil
}

// replacementCandidates 從尚未到期的 Safeguard 與場上的 Infernal Vessel 重建適用候選並依 source 穩定排序。
// 輸入 intent 不會改變 state；輸出排除已套用來源，讓每次 replacement 後都會重新判定。
func (g *Game) replacementCandidates(intent replacementIntent) []replacementEffect {
	candidates := []replacementEffect{}
	if intent.Kind == replacementIntentDamage && !intent.Combat {
		for _, effect := range g.state.ReplacementEffects {
			if effect.Kind == replacementDamagePrevent && effect.Target == intent.Target && effect.ExpiresAtTurn > g.state.Scheduler.TurnNumber && !containsObject(intent.Applied, effect.Source) {
				candidates = append(candidates, effect)
			}
		}
	}
	if intent.Kind == replacementIntentRecover {
		for source, object := range g.state.Objects {
			if g.state.Cards[object.Card].Definition == infernalVesselCardID && !containsObject(intent.Applied, source) {
				effect := replacementEffect{
					Kind:       replacementRecoverReduce,
					Source:     source,
					Controller: object.Owner,
				}
				candidates = append(
					candidates,
					effect,
				)
			}
		}
	}
	sort.Slice(
		candidates,
		func(first, second int) bool {
			return candidates[first].Source < candidates[second].Source
		},
	)
	return candidates
}

// applyReplacement 改寫暫存候選數量並追加 cause chain；Safeguard 最多防止 4，Infernal 每次減少 3。
// 輸入為 intent 指標與已選候選；副作用僅更新 intent，實際狀態異動由 commitReplacementIntent 執行。
func (g *Game) applyReplacement(intent *replacementIntent, effect replacementEffect) {
	switch effect.Kind {
	case replacementDamagePrevent:
		intent.Amount -= 4
		if intent.Amount < 0 {
			intent.Amount = 0
		}
	case replacementRecoverReduce:
		intent.Amount -= 3
	}
	intent.Applied = append(intent.Applied, effect.Source)
	cause := replacementCause{
		Kind:   string(effect.Kind),
		Source: effect.Source,
		Amount: intent.Amount,
	}
	intent.Chain = append(
		intent.Chain,
		cause,
	)
}

// commitReplacementIntent 提交無候選的傷害或 recover；零或負 recover 不視為 recover，因而不改變 state。
// 輸入為已完成 replacement 判定的 intent；成功時寫入傷害與完整 cause-chain event 並回傳 true。
func (g *Game) commitReplacementIntent(intent replacementIntent) bool {
	if intent.Amount <= 0 {
		if intent.Kind == replacementIntentDamage {
			g.recordReplacementEvent(intent)
			return true
		}
		return false
	}
	switch intent.Kind {
	case replacementIntentDamage:
		if !g.isLegalTarget(intent.Target) {
			return false
		}
		g.damageUnit(intent.Target, intent.Amount)
	case replacementIntentRecover:
		if !g.recoverChampionUnreplaced(intent.Target, intent.Amount) {
			return false
		}
	default:
		return false
	}
	g.recordReplacementEvent(intent)
	return true
}

// recordReplacementEvent 保存原始意圖、每個 replacement 與最終提交事件，讓 canonical state 和 replay 都可追溯。
// 輸入為已提交 intent；副作用為新增 event batch 與所有玩家可見事件。
func (g *Game) recordReplacementEvent(intent replacementIntent) {
	chain := append(
		[]replacementCause(nil),
		intent.Chain...,
	)
	chain = append(
		chain,
		replacementCause{
			Kind:   string(intent.Kind),
			Source: objectID(intent.Source),
			Amount: intent.Amount,
		},
	)
	g.state.NextEvent++
	event := gameEvent{
		Sequence: g.state.NextEvent,
		Kind:     string(intent.Kind),
		Card:     intent.Source,
	}
	batch := eventBatch{
		Player:     intent.Affected,
		Cause:      intent.Cause,
		Events:     []gameEvent{event},
		CauseChain: chain,
	}
	g.state.Events = append(
		g.state.Events,
		batch,
	)
	for _, player := range g.players {
		g.recordVisibleEvent(
			player,
			string(intent.Kind),
			entityID(intent.Source),
		)
	}
}

// damageNonCombat 將能力傷害送進 pipeline；輸入為目標、數量、來源與續行能力，回傳是否已完成提交。
// 多個適用 replacement 時副作用是建立 resolution-scoped PendingChoice，呼叫端應停止目前結算。
func (g *Game) damageNonCombat(target objectID, amount int, source cardInstanceID, continuation *abilityInstance) bool {
	intent := replacementIntent{
		Kind:     replacementIntentDamage,
		Target:   target,
		Affected: g.controllerFor(target),
		Amount:   amount,
		Source:   source,
		Cause:    "ability",
		Chain: []replacementCause{
			{
				Kind:   "damage-intent",
				Source: objectID(source),
				Amount: amount,
			},
		},
	}
	return g.applyReplacementIntent(
		intent,
		continuation,
	)
}

// recoverChampion 將一般 recover 送進 pipeline；recover cost 必須改用 recoverChampionUnreplaced，避免 Infernal 改寫付款條件。
// 輸入為 Champion 與數量；回傳是否實際 recover，副作用為減少傷害及記錄 cause chain。
func (g *Game) recoverChampion(target objectID, amount int) bool {
	intent := replacementIntent{
		Kind:     replacementIntentRecover,
		Target:   target,
		Affected: g.controllerFor(target),
		Amount:   amount,
		Cause:    "recover",
		Chain: []replacementCause{
			{
				Kind:   "recover-intent",
				Amount: amount,
			},
		},
	}
	return g.applyReplacementIntent(
		intent,
		nil,
	)
}

// recoverChampionUnreplaced 執行完成 replacement 判定後的 recover 或 recover cost，並忽略零或負數量。
// 輸入為 Champion 與正數 amount；成功時減少傷害至最低零，RecoverProhibited 時回傳 false 且無副作用。
func (g *Game) recoverChampionUnreplaced(target objectID, amount int) bool {
	if amount <= 0 {
		return false
	}
	for playerID, champion := range g.state.Champions {
		if champion.ID != target || g.characteristicsFor(target).RecoverProhibited {
			continue
		}
		champion.Damage -= amount
		if champion.Damage < 0 {
			champion.Damage = 0
		}
		g.state.Champions[playerID] = champion
		return true
	}
	return false
}

// controllerFor 回傳 Champion 或場上物件的控制玩家；找不到目標回傳 nil 供 pipeline 拒絕提交。
// 輸入為目標 object ID；副作用為零。
func (g *Game) controllerFor(target objectID) *model.Player {
	for _, champion := range g.state.Champions {
		if champion.ID == target {
			return champion.Owner
		}
	}
	if object, exists := g.state.Objects[target]; exists {
		return object.Owner
	}
	return nil
}
