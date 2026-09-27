package game

import (
	"fmt"
	"regexp"
	"strings"
)

type abilityDefinitionKind string

const (
	abilityKindCardistry abilityDefinitionKind = "cardistry"
	abilityKindAction    abilityDefinitionKind = "action"
)

type abilityTiming string

const (
	timingMainPhase abilityTiming = "main-phase"
	timingFast      abilityTiming = "fast"
)

type abilityUsage string

const usageOncePerObject abilityUsage = "once-per-object"

type abilityCostReduction string

const reductionDistinctSuitedCosts abilityCostReduction = "distinct-suited-reserve-costs"

type authoredEffectKind string

const (
	effectKindDraw   authoredEffectKind = "draw"
	effectKindDamage authoredEffectKind = "damage"
)

type abilityReference string

const (
	referenceController     abilityReference = "controller"
	referenceDeclaredTarget abilityReference = "declared-target"
)

type selectorKind string

const (
	selectorUnits            selectorKind = "units"
	selectorControlledSuited selectorKind = "controlled-suited"
)

type targetSelector struct {
	kind selectorKind
}

type valueKind string

const (
	valueConstant             valueKind = "constant"
	valueAdd                  valueKind = "add"
	valueDistinctPrintedCosts valueKind = "distinct-printed-costs"
)

// valueExpression 是可序列化的整數運算樹；左右運算元僅用於加法。
type valueExpression struct {
	Kind     valueKind        `json:"kind"`
	Constant int              `json:"constant,omitempty"`
	Left     *valueExpression `json:"left,omitempty"`
	Right    *valueExpression `json:"right,omitempty"`
	Selector selectorKind     `json:"selector,omitempty"`
	Player   abilityReference `json:"player,omitempty"`
}

type drawEffectDefinition struct {
	amount    int
	recipient abilityReference
}

type damageEffectDefinition struct {
	target abilityReference
	amount valueExpression
}

type authoredEffectDefinition struct {
	kind   authoredEffectKind
	draw   *drawEffectDefinition
	damage *damageEffectDefinition
}

type authoredAbilityDefinition struct {
	slot      AbilitySlotID
	kind      abilityDefinitionKind
	timing    abilityTiming
	usage     abilityUsage
	reduction abilityCostReduction
	baseCost  int
	target    *targetSelector
	effects   []authoredEffectDefinition
}

// compiledEffect 保存已驗證的具體效果 payload，不讓 authoring tag 進入對局狀態。
type compiledEffect struct {
	kind   authoredEffectKind
	draw   drawEffectDefinition
	damage damageEffectDefinition
}

// compiledAbilityDefinition 是建局前完成驗證的能力中介表示。
type compiledAbilityDefinition struct {
	slot      AbilitySlotID
	kind      abilityDefinitionKind
	timing    abilityTiming
	usage     abilityUsage
	reduction abilityCostReduction
	baseCost  int
	target    *targetSelector
	effects   []compiledEffect
}

var abilitySlotKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

// wonderlandsReignAbilities 集中宣告 Wonderland's Reign 的 Cardistry Slot、費用與抽牌效果。
// 輸入為零；輸出為獨立的 Go 編寫資料，副作用為零。
func wonderlandsReignAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:0mf1ug6yfi:front:cardistry-draw",
			kind:      abilityKindCardistry,
			timing:    timingMainPhase,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  10,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDraw,
					draw: &drawEffectDefinition{
						amount:    1,
						recipient: referenceController,
					},
				},
			},
		},
	}
}

// straightFlareAbilities 宣告行動目標與結算時求值的傷害公式。
// 輸入為零；輸出為新的 Go 編寫資料，副作用為零。
func straightFlareAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:28bjn8g50v:front:action-damage",
			kind: abilityKindAction,
			target: &targetSelector{
				kind: selectorUnits,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDamage,
					damage: &damageEffectDefinition{
						target: referenceDeclaredTarget,
						amount: valueExpression{
							Kind: valueAdd,
							Left: &valueExpression{
								Kind:     valueConstant,
								Constant: 1,
							},
							Right: &valueExpression{
								Kind:     valueDistinctPrintedCosts,
								Selector: selectorControlledSuited,
								Player:   referenceController,
							},
						},
					},
				},
			},
		},
	}
}

// authoredAbilitiesForCard 將固定卡牌的 Go 能力編寫資料交給同一個 compiler。
// 輸入為 Card ID；輸出為該卡目前已遷移的能力資料，副作用為零。
func authoredAbilitiesForCard(id CardID) []authoredAbilityDefinition {
	switch id {
	case wonderlandsReignCardID:
		return wonderlandsReignAbilities()
	case straightFlareCardID:
		return straightFlareAbilities()
	}
	return nil
}

// compileAbilityDefinitions 驗證 Definition、Face、Slot、kind、payload 與 reference，產生 typed 中介表示。
// 輸入為卡牌定義和 Go 編寫資料；輸出為不可共用 authoring pointer 的能力資料或具完整脈絡的錯誤，副作用為零。
func compileAbilityDefinitions(definition CardDefinition, authored []authoredAbilityDefinition) ([]compiledAbilityDefinition, error) {
	compiled := make([]compiledAbilityDefinition, 0, len(authored))
	seen := make(map[AbilitySlotID]struct{}, len(authored))
	allowed := make(map[AbilitySlotID]struct{})
	registeredAbilities := authoredAbilitiesForCard(definition.id)
	for _, registered := range registeredAbilities {
		allowed[registered.slot] = struct{}{}
	}
	prefix := fmt.Sprintf("ability:%s:front:", definition.id)
	cardistryCount := 0
	for _, ability := range authored {
		context := fmt.Sprintf("definition %q face %q slot %q", definition.id, definition.face.id, ability.slot)
		slotKey, matchesFace := strings.CutPrefix(string(ability.slot), prefix)
		if !matchesFace || definition.face.id != CardFaceID("face:"+string(definition.id)+":front") || !abilitySlotKeyPattern.MatchString(slotKey) {
			return nil, fmt.Errorf("%s: slot has invalid definition or face relationship", context)
		}
		if _, exists := seen[ability.slot]; exists {
			return nil, fmt.Errorf("%s: duplicate slot", context)
		}
		seen[ability.slot] = struct{}{}
		if _, exists := allowed[ability.slot]; !exists {
			return nil, fmt.Errorf("%s: slot is not registered for definition", context)
		}
		if ability.kind != abilityKindCardistry && ability.kind != abilityKindAction {
			return nil, fmt.Errorf("%s: unknown kind %q", context, ability.kind)
		}
		if ability.kind == abilityKindCardistry {
			cardistryCount++
			if cardistryCount > 1 {
				return nil, fmt.Errorf("%s: kind permits only one Cardistry slot per face", context)
			}
			if ability.timing != timingMainPhase && ability.timing != timingFast {
				return nil, fmt.Errorf("%s: timing unsupported %q", context, ability.timing)
			}
			if ability.usage != usageOncePerObject {
				return nil, fmt.Errorf("%s: usage unsupported %q", context, ability.usage)
			}
			if ability.reduction != reductionDistinctSuitedCosts {
				return nil, fmt.Errorf("%s: cost_reduction unsupported %q", context, ability.reduction)
			}
			if ability.baseCost < 0 {
				return nil, fmt.Errorf("%s: base_cost must be nonnegative", context)
			}
			if ability.target != nil {
				return nil, fmt.Errorf("%s: target unsupported for Cardistry", context)
			}
		} else {
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 {
				return nil, fmt.Errorf("%s: action cost or timing payload unsupported", context)
			}
			if ability.target == nil || ability.target.kind != selectorUnits {
				return nil, fmt.Errorf("%s: target selector must be units", context)
			}
		}
		if len(ability.effects) == 0 {
			return nil, fmt.Errorf("%s: effects must not be empty", context)
		}
		result := compiledAbilityDefinition{
			slot:      ability.slot,
			kind:      ability.kind,
			timing:    ability.timing,
			usage:     ability.usage,
			reduction: ability.reduction,
			baseCost:  ability.baseCost,
			effects:   make([]compiledEffect, 0, len(ability.effects)),
		}
		if ability.target != nil {
			selector := *ability.target
			result.target = &selector
		}
		for index, effect := range ability.effects {
			field := fmt.Sprintf("effects[%d]", index)
			switch effect.kind {
			case effectKindDraw:
				if ability.kind != abilityKindCardistry || effect.draw == nil || effect.damage != nil {
					return nil, fmt.Errorf("%s: %s.draw payload is required only for Cardistry", context, field)
				}
				if effect.draw.amount <= 0 {
					return nil, fmt.Errorf("%s: %s.draw.amount must be positive", context, field)
				}
				if effect.draw.recipient != referenceController {
					return nil, fmt.Errorf("%s: %s.draw.recipient unknown or wrong reference %q", context, field, effect.draw.recipient)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindDraw,
					draw: *effect.draw,
				})
			case effectKindDamage:
				if ability.kind != abilityKindAction || effect.damage == nil || effect.draw != nil {
					return nil, fmt.Errorf("%s: %s.damage payload is required only for Action", context, field)
				}
				if effect.damage.target != referenceDeclaredTarget {
					return nil, fmt.Errorf("%s: %s.damage.target must reference declared target", context, field)
				}
				if err := validateValueExpression(effect.damage.amount); err != nil {
					return nil, fmt.Errorf("%s: %s.damage.amount: %w", context, field, err)
				}
				result.effects = append(result.effects, compiledEffect{
					kind: effectKindDamage,
					damage: damageEffectDefinition{
						target: effect.damage.target,
						amount: cloneValueExpression(effect.damage.amount),
					},
				})
			default:
				return nil, fmt.Errorf("%s: %s.kind unknown %q", context, field, effect.kind)
			}
		}
		compiled = append(compiled, result)
	}
	for _, registered := range registeredAbilities {
		if _, exists := seen[registered.slot]; !exists {
			return nil, fmt.Errorf("definition %q face %q slot %q: slot is missing", definition.id, definition.face.id, registered.slot)
		}
	}
	return compiled, nil
}

// validateValueExpression 驗證整數樹的 kind、運算元與 reference 型別。
// 輸入為編寫的整數樹；輸出為驗證錯誤或 nil，副作用為零。
func validateValueExpression(value valueExpression) error {
	return validateValueNode(&value, make(map[*valueExpression]bool), 0)
}

// validateValueNode 限制樹深並拒絕循環指標，避免壞定義在載入時造成 panic。
// 輸入為節點、目前路徑與深度；輸出為驗證錯誤或 nil，副作用僅更新呼叫內路徑。
func validateValueNode(value *valueExpression, path map[*valueExpression]bool, depth int) error {
	if depth > 32 || path[value] {
		return fmt.Errorf("value expression is cyclic or too deep")
	}
	path[value] = true
	defer delete(path, value)
	switch value.Kind {
	case valueConstant:
		if value.Constant < 0 || value.Left != nil || value.Right != nil || value.Selector != "" || value.Player != "" {
			return fmt.Errorf("invalid constant payload")
		}
	case valueAdd:
		if value.Left == nil || value.Right == nil || value.Constant != 0 || value.Selector != "" || value.Player != "" {
			return fmt.Errorf("add requires exactly two operands")
		}
		if err := validateValueNode(value.Left, path, depth+1); err != nil {
			return fmt.Errorf("left: %w", err)
		}
		if err := validateValueNode(value.Right, path, depth+1); err != nil {
			return fmt.Errorf("right: %w", err)
		}
	case valueDistinctPrintedCosts:
		if value.Selector != selectorControlledSuited || value.Player != referenceController || value.Constant != 0 || value.Left != nil || value.Right != nil {
			return fmt.Errorf("distinct printed costs requires controlled-suited selector and controller reference")
		}
	default:
		return fmt.Errorf("unknown value kind %q", value.Kind)
	}
	return nil
}

// cloneValueExpression 複製整數樹，隔離可變的 Go 編寫資料與編譯結果。
// 輸入為已驗證的 expression；輸出為獨立運算樹，副作用為零。
func cloneValueExpression(value valueExpression) valueExpression {
	clone := value
	if value.Left != nil {
		left := cloneValueExpression(*value.Left)
		clone.Left = &left
	}
	if value.Right != nil {
		right := cloneValueExpression(*value.Right)
		clone.Right = &right
	}
	return clone
}

// operations 將已編譯效果轉成目前統一 Ability Runtime 的可序列化操作。
// 輸入為已驗證的能力；輸出為每次呼叫新建的操作序列，副作用為零。
func (ability compiledAbilityDefinition) operations() []effectOperation {
	operations := make([]effectOperation, 0, len(ability.effects))
	for _, effect := range ability.effects {
		switch effect.kind {
		case effectKindDraw:
			operations = append(operations, effectOperation{
				Kind:   effectOperationDraw,
				Amount: effect.draw.amount,
			})
		case effectKindDamage:
			amount := cloneValueExpression(effect.damage.amount)
			operations = append(operations, effectOperation{
				Kind:            effectOperationDamage,
				TargetReference: effect.damage.target,
				Value:           &amount,
			})
		}
	}
	return operations
}

// compiledAbility 取得來源卡指定 kind 的已編譯定義，不讀取可變對局狀態以外的編寫資料。
// 輸入為卡牌實例與能力種類；輸出為能力及存在旗標，副作用為零。
func (g *Game) compiledAbility(card cardInstanceID, kind abilityDefinitionKind) (compiledAbilityDefinition, bool) {
	definition, exists := g.definitions[g.state.Cards[card].Definition]
	if !exists {
		return compiledAbilityDefinition{}, false
	}
	for _, ability := range definition.abilities {
		if ability.kind == kind {
			return ability, true
		}
	}
	return compiledAbilityDefinition{}, false
}

// compiledAction 取得來源卡已編譯的 Action 定義。
// 輸入為卡牌實例；輸出為 Action 能力與存在旗標，副作用為零。
func (g *Game) compiledAction(card cardInstanceID) (compiledAbilityDefinition, bool) {
	return g.compiledAbility(card, abilityKindAction)
}

// compiledCardistry 取得來源卡的已編譯 Cardistry 定義。
// 輸入為卡牌實例；輸出為能力及存在旗標，副作用為零。
func (g *Game) compiledCardistry(card cardInstanceID) (compiledAbilityDefinition, bool) {
	return g.compiledAbility(card, abilityKindCardistry)
}

// compileReplayDefinitions 重建重播所需的固定可執行定義，不將編寫資料寫入 Game State。
// 輸入為零；輸出為目前引擎版本釘選的定義或驗證錯誤，副作用為零。
func compileReplayDefinitions() (map[CardID]CardDefinition, error) {
	definitions := make(map[CardID]CardDefinition)
	for _, id := range []CardID{wonderlandsReignCardID, straightFlareCardID} {
		definition := CardDefinition{
			id: id,
			face: CardFace{
				id: CardFaceID("face:" + string(id) + ":front"),
			},
		}
		abilities, err := compileAbilityDefinitions(definition, authoredAbilitiesForCard(id))
		if err != nil {
			return nil, fmt.Errorf("compile replay definition: %w", err)
		}
		definition.abilities = abilities
		definitions[id] = definition
	}
	return definitions, nil
}
