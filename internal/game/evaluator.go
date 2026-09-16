package game

import (
	"sort"

	"go-tcg/internal/model"
)

type effectLayer uint8

const (
	effectLayerBase effectLayer = iota
	effectLayerType
	effectLayerElement
	effectLayerAbility
	effectLayerModifier
)

type powerLifeSubLayer uint8

const (
	powerLifeSet powerLifeSubLayer = iota
	powerLifeModify
	powerLifeCounter
	powerLifeSwitch
)

type effectScope string

const (
	effectScopeObject                effectScope = "object"
	effectScopeOtherControlledAllies effectScope = "other_controlled_allies"
	effectScopeClassMatchedSource    effectScope = "class_matched_source"
)

type continuousModifier struct {
	SetPower             *int `json:"set_power,omitempty"`
	SetLife              *int `json:"set_life,omitempty"`
	PowerDelta           int  `json:"power_delta,omitempty"`
	LifeDelta            int  `json:"life_delta,omitempty"`
	ReserveCostDelta     int  `json:"reserve_cost_delta,omitempty"`
	GrantImmortality     bool `json:"grant_immortality,omitempty"`
	ProhibitRecover      bool `json:"prohibit_recover,omitempty"`
	SwitchPowerLife      bool `json:"switch_power_life,omitempty"`
	GrantStealth         bool `json:"grant_stealth,omitempty"`
	GrantTrueSight       bool `json:"grant_true_sight,omitempty"`
	RemovePride          bool `json:"remove_pride,omitempty"`
	GrantRedHareOnAttack bool `json:"grant_red_hare_on_attack,omitempty"`
}

// continuousEffect 保存可序列化的持續效果資料。
// 靜態效果每次查詢時由場上物件重建；實例效果保留建立時的目標，不隨來源離場重新選取。
type continuousEffect struct {
	ID            uint64             `json:"id"`
	Source        objectID           `json:"source,omitempty"`
	Controller    *model.Player      `json:"controller"`
	Target        objectID           `json:"target,omitempty"`
	Scope         effectScope        `json:"scope"`
	Layer         effectLayer        `json:"layer"`
	PowerLife     powerLifeSubLayer  `json:"power_life_sub_layer,omitempty"`
	Timestamp     uint64             `json:"timestamp"`
	DependsOn     []uint64           `json:"depends_on,omitempty"`
	ExpiresAtTurn uint64             `json:"expires_at_turn,omitempty"`
	Modifier      continuousModifier `json:"modifier"`
}

type characteristics struct {
	Power             int
	Life              int
	ReserveCost       int
	MemoryCost        int
	Pride             int
	Immortal          bool
	RecoverProhibited bool
	Stealth           bool
	TrueSight         bool
	GrantedOnAttack   bool
}

// characteristicsFor 從物件對應牌的基礎值計算目前特性，不改寫牌本身。
// 先套用排序後的一般效果，再加 POWER/LIFE 指示物，最後套用攻擊力與生命交換。
// 找不到對應牌時回傳零值特性；counter 子層由物件指示物處理，不直接套用該子層 modifier。
func (g *Game) characteristicsFor(id objectID) characteristics {
	card, exists := g.cardForObject(id)
	if !exists {
		return characteristics{}
	}
	result := characteristicsFromCard(card)
	effects := g.orderedEffectsFor(id)
	for _, effect := range effects {
		if effect.Layer == effectLayerModifier && (effect.PowerLife == powerLifeCounter || effect.PowerLife == powerLifeSwitch) {
			continue
		}
		applyContinuousModifier(
			&result,
			effect.Modifier,
		)
	}
	for counter, amount := range g.countersFor(id) {
		switch counter {
		case "POWER":
			result.Power += amount
		case "LIFE":
			result.Life += amount
		}
	}
	for _, effect := range effects {
		if effect.Layer != effectLayerModifier || effect.PowerLife != powerLifeSwitch {
			continue
		}
		applyContinuousModifier(&result, effect.Modifier)
	}
	return result
}

func (g *Game) characteristicsForCard(id cardInstanceID) characteristics {
	card, exists := g.state.Cards[id]
	if !exists {
		return characteristics{}
	}
	return characteristicsFromCard(card)
}

func characteristicsFromCard(card cardInstance) characteristics {
	result := characteristics{
		Power:       card.Power,
		Life:        card.Life,
		ReserveCost: card.ReserveCost,
		MemoryCost:  card.MemoryCost,
	}
	if card.Definition == redHareCardID {
		result.Pride = 3
	}
	return result
}

func applyContinuousModifier(result *characteristics, modifier continuousModifier) {
	if modifier.SetPower != nil {
		result.Power = *modifier.SetPower
	}
	if modifier.SetLife != nil {
		result.Life = *modifier.SetLife
	}
	result.Power += modifier.PowerDelta
	result.Life += modifier.LifeDelta
	result.ReserveCost += modifier.ReserveCostDelta
	result.Immortal = result.Immortal || modifier.GrantImmortality
	result.RecoverProhibited = result.RecoverProhibited || modifier.ProhibitRecover
	result.Stealth = result.Stealth || modifier.GrantStealth
	result.TrueSight = result.TrueSight || modifier.GrantTrueSight
	if modifier.RemovePride {
		result.Pride = 0
	}
	result.GrantedOnAttack = result.GrantedOnAttack || modifier.GrantRedHareOnAttack
	if modifier.SwitchPowerLife {
		result.Power, result.Life = result.Life, result.Power
	}
}

func (g *Game) countersFor(id objectID) map[string]int {
	for _, champion := range g.state.Champions {
		if champion.ID == id {
			return champion.Counters
		}
	}
	object, exists := g.state.Objects[id]
	if !exists {
		return nil
	}
	return object.Counters
}

// orderedEffectsFor 合併場上重建的靜態效果與尚未到期的實例效果。
// 先依 layer、攻擊力生命子層及 timestamp 穩定排序，再依 DependsOn 調整順序。
func (g *Game) orderedEffectsFor(target objectID) []continuousEffect {
	effects := append(
		g.staticEffectsFor(target),
		g.instancedEffectsFor(target)...,
	)
	sort.SliceStable(
		effects,
		func(first, second int) bool {
			if effects[first].Layer != effects[second].Layer {
				return effects[first].Layer < effects[second].Layer
			}
			if effects[first].Layer == effectLayerModifier && effects[first].PowerLife != effects[second].PowerLife {
				return effects[first].PowerLife < effects[second].PowerLife
			}
			return effects[first].Timestamp < effects[second].Timestamp
		},
	)
	return dependencyOrdered(effects)
}

// dependencyOrdered 優先選出 DependsOn 已出現在結果中的效果，並保留可選效果的原始相對順序。
// 若循環或缺失依賴使全部剩餘效果都無法選出，取剩餘首項繼續，不回傳錯誤。
func dependencyOrdered(effects []continuousEffect) []continuousEffect {
	ordered := make([]continuousEffect, 0, len(effects))
	remaining := append([]continuousEffect(nil), effects...)
	for len(remaining) > 0 {
		index := -1
		for candidate, effect := range remaining {
			if dependenciesResolved(effect, ordered) {
				index = candidate
				break
			}
		}
		if index < 0 {
			index = 0
		}
		ordered = append(ordered, remaining[index])
		remaining = append(remaining[:index], remaining[index+1:]...)
	}
	return ordered
}

func dependenciesResolved(effect continuousEffect, ordered []continuousEffect) bool {
	for _, dependency := range effect.DependsOn {
		for _, candidate := range ordered {
			if candidate.ID == dependency {
				goto resolved
			}
		}
		return false
	resolved:
	}
	return true
}

// instancedEffectsFor 只回傳目標快照相符且尚未到期的實例效果。
// 不要求來源仍在場上；ExpiresAtTurn=0 表示未設定回合期限。
func (g *Game) instancedEffectsFor(target objectID) []continuousEffect {
	effects := make([]continuousEffect, 0, len(g.state.ContinuousEffects))
	for _, effect := range g.state.ContinuousEffects {
		if effect.Target != target || (effect.ExpiresAtTurn > 0 && effect.ExpiresAtTurn <= g.state.Scheduler.TurnNumber) {
			continue
		}
		effects = append(effects, effect)
	}
	return effects
}

func (g *Game) staticEffectsFor(target objectID) []continuousEffect {
	effects := make([]continuousEffect, 0, 2)
	targetObject, targetExists := g.state.Objects[target]
	for sourceID, source := range g.state.Objects {
		card := g.state.Cards[source.Card]
		if card.Definition == arthurYoungHeirCardID && source.Rested && targetExists && sourceID != target && samePlayer(source.Owner, targetObject.Owner) && containsString(targetObject.Types, "ALLY") {
			effects = append(effects, continuousEffect{
				Source:     sourceID,
				Controller: source.Owner,
				Scope:      effectScopeOtherControlledAllies,
				Layer:      effectLayerModifier,
				PowerLife:  powerLifeModify,
				Timestamp:  uint64(len(effects)),
				Modifier: continuousModifier{
					PowerDelta: 1,
				},
			})
		}
		if sourceID == target && card.Definition == bulwarkSwordCardID && g.championHasClass(source.Owner, card.Classes) {
			effects = append(effects, continuousEffect{
				Source:     sourceID,
				Controller: source.Owner,
				Scope:      effectScopeClassMatchedSource,
				Layer:      effectLayerModifier,
				PowerLife:  powerLifeModify,
				Timestamp:  uint64(len(effects)),
				Modifier: continuousModifier{
					PowerDelta:       1,
					ReserveCostDelta: 2,
				},
			})
		}
		if sourceID == target && card.Definition == noireCardID && g.controlsAnotherSuitedAlly(source.Owner, sourceID) {
			effects = append(effects, continuousEffect{Source: sourceID, Controller: source.Owner, Scope: effectScopeObject, Layer: effectLayerAbility, Timestamp: uint64(len(effects)), Modifier: continuousModifier{GrantStealth: true}})
		}
		if sourceID == target && card.Definition == heatedVengeanceCardID && g.championDamagedThisTurn(source.Owner) {
			effects = append(
				effects,
				continuousEffect{
					Source:     sourceID,
					Controller: source.Owner,
					Scope:      effectScopeObject,
					Layer:      effectLayerModifier,
					PowerLife:  powerLifeModify,
					Timestamp:  uint64(len(effects)),
					Modifier: continuousModifier{
						PowerDelta: 3,
					},
				},
			)
		}
		if sourceID == target && card.Definition == redHareCardID && g.controlsQualifiedHumanAlly(source.Owner, sourceID) {
			effects = append(
				effects,
				continuousEffect{
					Source:     sourceID,
					Controller: source.Owner,
					Scope:      effectScopeObject,
					Layer:      effectLayerAbility,
					Timestamp:  uint64(len(effects)),
					Modifier: continuousModifier{
						GrantRedHareOnAttack: true,
						RemovePride:          true,
					},
				},
			)
		}
		if card.Definition == veritaCardID && sourceID != target && targetExists && samePlayer(source.Owner, targetObject.Owner) && containsString(targetObject.Types, "ALLY") && g.cardHasSubtype(targetObject.Card, "SUITED") {
			effects = append(
				effects,
				continuousEffect{
					Source:     sourceID,
					Controller: source.Owner,
					Scope:      effectScopeOtherControlledAllies,
					Layer:      effectLayerAbility,
					Timestamp:  uint64(len(effects)),
					Modifier: continuousModifier{
						GrantImmortality: true,
					},
				},
			)
		}
	}
	return effects
}

func (g *Game) controlsQualifiedHumanAlly(player *model.Player, exclude objectID) bool {
	for id, object := range g.state.Objects {
		if id == exclude || !samePlayer(object.Owner, player) {
			continue
		}
		if g.objectHasCharacteristics(
			id,
			[]string{
				"ALLY",
				"UNIQUE",
			},
			[]string{
				"HUMAN",
			},
			[]string{
				"FIRE",
				"TERA",
			},
		) {
			return true
		}
	}
	return false
}

// objectHasCharacteristics 是 static ability 與 permission query 共用的卡牌特性查詢。
// requiredTypes 與 requiredSubtypes 必須全部符合；elements 則只要符合其中一個即可。
func (g *Game) objectHasCharacteristics(id objectID, requiredTypes, requiredSubtypes, elements []string) bool {
	card, exists := g.cardForObject(id)
	if !exists {
		return false
	}
	for _, requiredType := range requiredTypes {
		if !containsString(card.Types, requiredType) {
			return false
		}
	}
	for _, requiredSubtype := range requiredSubtypes {
		if !containsString(card.Subtypes, requiredSubtype) {
			return false
		}
	}
	return len(elements) == 0 || containsAny(card.Elements, elements)
}

func (g *Game) championDamagedThisTurn(player *model.Player) bool {
	champion, exists := g.state.Champions[player.UID]
	return exists && champion.DamageTurn == g.state.Scheduler.TurnNumber
}

func (g *Game) controlsAnotherSuitedAlly(player *model.Player, exclude objectID) bool {
	for id, object := range g.state.Objects {
		if id != exclude && samePlayer(object.Owner, player) && containsString(object.Types, "ALLY") && g.cardHasSubtype(object.Card, "SUITED") {
			return true
		}
	}
	return false
}

func (g *Game) addContinuousEffect(effect continuousEffect) {
	g.state.NextEffect++
	effect.ID = g.state.NextEffect
	effect.Timestamp = effect.ID
	g.state.ContinuousEffects = append(g.state.ContinuousEffects, effect)
}

func (g *Game) expireContinuousEffects() {
	active := g.state.ContinuousEffects[:0]
	for _, effect := range g.state.ContinuousEffects {
		if effect.ExpiresAtTurn > 0 && effect.ExpiresAtTurn <= g.state.Scheduler.TurnNumber {
			continue
		}
		active = append(active, effect)
	}
	g.state.ContinuousEffects = active
}
