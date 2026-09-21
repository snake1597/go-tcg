package game

import (
	"fmt"

	"go-tcg/internal/model"
)

type effectStackItemKind string

const (
	effectStackMaterialization effectStackItemKind = "materialization"
	effectStackTonorisTaunt    effectStackItemKind = "tonoris_on_enter_taunt"
	effectStackCombat          effectStackItemKind = "combat"
	effectStackAbility         effectStackItemKind = "ability"
)

// effectStackItem 保存等待結算的 materialization、戰鬥或能力。
type effectStackItem struct {
	Kind       effectStackItemKind `json:"kind"`
	Controller *model.Player       `json:"controller"`
	Source     cardInstanceID      `json:"source"`
	Target     objectID            `json:"target,omitempty"`
	Attacker   objectID            `json:"attacker,omitempty"`
	SourceLKI  cardInstanceID      `json:"source_lki,omitempty"`
	Ability    *abilityInstance    `json:"ability,omitempty"`
}

// resolveTopEffectStack 先移除堆疊頂項目，再依種類結算，因此結算中加入的新項目會留在堆疊。
// 呼叫端須保證堆疊非空，能力項目須有 Ability；未知種類會 panic。
func (g *Game) resolveTopEffectStack() {
	lastIndex := len(g.state.EffectsStack) - 1
	item := g.state.EffectsStack[lastIndex]
	g.state.EffectsStack = g.state.EffectsStack[:lastIndex]
	switch item.Kind {
	case effectStackMaterialization:
		g.resolveMaterialization(item)
	case effectStackTonorisTaunt:
		g.resolveTonorisTaunt(item)
	case effectStackCombat:
		g.resolveCombat(item)
	case effectStackAbility:
		g.resolveAbility(*item.Ability)
	default:
		panic(fmt.Sprintf("unknown effect stack item %q", item.Kind))
	}
}
