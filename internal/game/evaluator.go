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
	SetPower         *int `json:"set_power,omitempty"`
	SetLife          *int `json:"set_life,omitempty"`
	PowerDelta       int  `json:"power_delta,omitempty"`
	LifeDelta        int  `json:"life_delta,omitempty"`
	ReserveCostDelta int  `json:"reserve_cost_delta,omitempty"`
	GrantImmortality bool `json:"grant_immortality,omitempty"`
	ProhibitRecover  bool `json:"prohibit_recover,omitempty"`
	SwitchPowerLife  bool `json:"switch_power_life,omitempty"`
}

// continuousEffect is serialized game state: static effects are rebuilt from
// the Field for each evaluation, while instanced effects keep a target snapshot.
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
	Immortal          bool
	RecoverProhibited bool
}

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
	return characteristics{
		Power:       card.Power,
		Life:        card.Life,
		ReserveCost: card.ReserveCost,
		MemoryCost:  card.MemoryCost,
	}
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
	}
	return effects
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
