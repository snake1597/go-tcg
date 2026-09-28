package game

import (
	"fmt"
	"regexp"
	"strings"
)

type abilityDefinitionKind string

const (
	abilityKindCardistry   abilityDefinitionKind = "cardistry"
	abilityKindAction      abilityDefinitionKind = "action"
	abilityKindTriggered   abilityDefinitionKind = "triggered"
	abilityKindStatic      abilityDefinitionKind = "static"
	abilityKindReplacement abilityDefinitionKind = "replacement"
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
	effectKindDraw       authoredEffectKind = "draw"
	effectKindDamage     authoredEffectKind = "damage"
	effectKindChooseCard authoredEffectKind = "choose-card"
	effectKindDiscard    authoredEffectKind = "discard"
)

type abilityReference string

const (
	referenceController     abilityReference = "controller"
	referenceDeclaredTarget abilityReference = "declared-target"
	referenceEventTarget    abilityReference = "event-target"
)

type eventKind string

const eventKindWield eventKind = "wield"

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

type chooseCardEffectDefinition struct {
	zone    cardZone
	binding resolutionBinding
}

type discardEffectDefinition struct {
	binding resolutionBinding
}

type triggerDefinition struct {
	event eventKind
}

type staticPredicateKind string

const (
	staticPredicateSelf                       staticPredicateKind = "self"
	staticPredicateControlsQualifiedHumanAlly staticPredicateKind = "controls-qualified-human-ally"
)

type effectSubLayer string

const effectSubLayerNone effectSubLayer = "none"

type staticDuration string

const staticDurationWhilePredicate staticDuration = "while-predicate"

type sourcePresence string

const sourcePresenceRequired sourcePresence = "required"

type staticEffectDefinition struct {
	predicate      staticPredicateKind
	layer          effectLayer
	sublayer       effectSubLayer
	modifier       continuousModifier
	duration       staticDuration
	sourcePresence sourcePresence
}

type replacementEvent string

const replacementEventRecover replacementEvent = "recover"

type replacementTransformation string

const replacementTransformReduce replacementTransformation = "recover-reduce"

type replacementDefinition struct {
	event          replacementEvent
	transformation replacementTransformation
	amount         int
}

type authoredEffectDefinition struct {
	kind       authoredEffectKind
	draw       *drawEffectDefinition
	damage     *damageEffectDefinition
	chooseCard *chooseCardEffectDefinition
	discard    *discardEffectDefinition
}

type authoredAbilityDefinition struct {
	slot        AbilitySlotID
	kind        abilityDefinitionKind
	timing      abilityTiming
	usage       abilityUsage
	reduction   abilityCostReduction
	baseCost    int
	target      *targetSelector
	trigger     *triggerDefinition
	static      *staticEffectDefinition
	replacement *replacementDefinition
	effects     []authoredEffectDefinition
}

// compiledEffect 保存已驗證的具體效果 payload，不讓 authoring tag 進入對局狀態。
type compiledEffect struct {
	kind       authoredEffectKind
	draw       drawEffectDefinition
	damage     damageEffectDefinition
	chooseCard chooseCardEffectDefinition
	discard    discardEffectDefinition
}

// compiledAbilityDefinition 是建局前完成驗證的能力中介表示。
type compiledAbilityDefinition struct {
	slot        AbilitySlotID
	kind        abilityDefinitionKind
	timing      abilityTiming
	usage       abilityUsage
	reduction   abilityCostReduction
	baseCost    int
	target      *targetSelector
	trigger     *triggerDefinition
	static      *staticEffectDefinition
	replacement *replacementDefinition
	effects     []compiledEffect
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

// threeOfHeartsAbilities 集中宣告 Three of Hearts 的抽牌、具名選牌及棄牌 Cardistry Slot。
// 輸入為零；輸出為獨立的 Go 編寫資料，副作用為零。
func threeOfHeartsAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot:      "ability:1db8hz4prm:front:cardistry-draw-discard",
			kind:      abilityKindCardistry,
			timing:    timingMainPhase,
			usage:     usageOncePerObject,
			reduction: reductionDistinctSuitedCosts,
			baseCost:  3,
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDraw,
					draw: &drawEffectDefinition{
						amount:    1,
						recipient: referenceController,
					},
				},
				{
					kind: effectKindChooseCard,
					chooseCard: &chooseCardEffectDefinition{
						zone:    cardZoneHand,
						binding: "discard-card",
					},
				},
				{
					kind: effectKindDiscard,
					discard: &discardEffectDefinition{
						binding: "discard-card",
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

// impactHammerAbilities 宣告 Impact Hammer 觀察 wield event 後對該事件 unit 造成傷害的 triggered Slot。
// 輸入為零；輸出為新的 Go 編寫資料，副作用為零。
func impactHammerAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:chsbalegbs:front:on-wield-self-damage",
			kind: abilityKindTriggered,
			trigger: &triggerDefinition{
				event: eventKindWield,
			},
			effects: []authoredEffectDefinition{
				{
					kind: effectKindDamage,
					damage: &damageEffectDefinition{
						target: referenceEventTarget,
						amount: valueExpression{
							Kind:     valueConstant,
							Constant: 3,
						},
					},
				},
			},
		},
	}
}

// redHareAbilities 宣告 Red Hare 的 Pride 與條件式靜態修正，讓中央 evaluator 套用攻擊限制與能力。
// 輸入為零；輸出為不共享可變狀態的 Go 編寫資料，副作用為零。
func redHareAbilities() []authoredAbilityDefinition {
	pride := 3
	return []authoredAbilityDefinition{
		{
			slot: "ability:5du8f077ua:front:pride",
			kind: abilityKindStatic,
			static: &staticEffectDefinition{
				predicate: staticPredicateSelf,
				layer:     effectLayerAbility,
				sublayer:  effectSubLayerNone,
				modifier: continuousModifier{
					SetPride: &pride,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
			},
		},
		{
			slot: "ability:5du8f077ua:front:qualified-human",
			kind: abilityKindStatic,
			static: &staticEffectDefinition{
				predicate: staticPredicateControlsQualifiedHumanAlly,
				layer:     effectLayerAbility,
				sublayer:  effectSubLayerNone,
				modifier: continuousModifier{
					RemovePride:          true,
					GrantRedHareOnAttack: true,
				},
				duration:       staticDurationWhilePredicate,
				sourcePresence: sourcePresenceRequired,
			},
		},
	}
}

// infernalVesselAbilities 宣告 Infernal Vessel 對 recover event 的 replacement 與減少數量 transformation。
// 輸入為零；輸出為不可共用的 Go 編寫資料，副作用為零。
func infernalVesselAbilities() []authoredAbilityDefinition {
	return []authoredAbilityDefinition{
		{
			slot: "ability:vgWgu1DUYv:front:recover-reduce",
			kind: abilityKindReplacement,
			replacement: &replacementDefinition{
				event:          replacementEventRecover,
				transformation: replacementTransformReduce,
				amount:         3,
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
	case threeOfHeartsCardID:
		return threeOfHeartsAbilities()
	case impactHammerCardID:
		return impactHammerAbilities()
	case redHareCardID:
		return redHareAbilities()
	case infernalVesselCardID:
		return infernalVesselAbilities()
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
		switch ability.kind {
		case abilityKindCardistry:
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
		case abilityKindAction:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 {
				return nil, fmt.Errorf("%s: action cost or timing payload unsupported", context)
			}
			if ability.target == nil || ability.target.kind != selectorUnits {
				return nil, fmt.Errorf("%s: target selector must be units", context)
			}
			if ability.trigger != nil {
				return nil, fmt.Errorf("%s: trigger unsupported for Action", context)
			}
		case abilityKindTriggered:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.target != nil {
				return nil, fmt.Errorf("%s: triggered declaration payload unsupported", context)
			}
			if ability.trigger == nil || ability.trigger.event != eventKindWield {
				return nil, fmt.Errorf("%s: trigger.event must be wield", context)
			}
		case abilityKindStatic:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.target != nil || ability.trigger != nil || len(ability.effects) != 0 {
				return nil, fmt.Errorf("%s: static declaration payload unsupported", context)
			}
			if err := validateStaticEffectDefinition(ability.static); err != nil {
				return nil, fmt.Errorf("%s: static: %w", context, err)
			}
		case abilityKindReplacement:
			if ability.timing != "" || ability.usage != "" || ability.reduction != "" || ability.baseCost != 0 || ability.target != nil || ability.trigger != nil || ability.static != nil || len(ability.effects) != 0 {
				return nil, fmt.Errorf("%s: replacement declaration payload unsupported", context)
			}
			if err := validateReplacementDefinition(ability.replacement); err != nil {
				return nil, fmt.Errorf("%s: replacement: %w", context, err)
			}
		default:
			return nil, fmt.Errorf("%s: unknown kind %q", context, ability.kind)
		}
		if ability.kind != abilityKindStatic && ability.kind != abilityKindReplacement && len(ability.effects) == 0 {
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
		if ability.trigger != nil {
			trigger := *ability.trigger
			result.trigger = &trigger
		}
		if ability.static != nil {
			static := cloneStaticEffectDefinition(*ability.static)
			result.static = &static
		}
		if ability.replacement != nil {
			replacement := *ability.replacement
			result.replacement = &replacement
		}
		bindings := make(map[resolutionBinding]struct{})
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
				if (ability.kind != abilityKindAction && ability.kind != abilityKindTriggered) || effect.damage == nil || effect.draw != nil {
					return nil, fmt.Errorf("%s: %s.damage payload is required only for Action or triggered", context, field)
				}
				expectedTarget := referenceDeclaredTarget
				if ability.kind == abilityKindTriggered {
					expectedTarget = referenceEventTarget
				}
				if effect.damage.target != expectedTarget {
					return nil, fmt.Errorf("%s: %s.damage.target has wrong reference %q", context, field, effect.damage.target)
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
			case effectKindChooseCard:
				if ability.kind != abilityKindCardistry || effect.chooseCard == nil || effect.draw != nil || effect.damage != nil || effect.discard != nil {
					return nil, fmt.Errorf("%s: %s.choose_card payload is required only for Cardistry", context, field)
				}
				if effect.chooseCard.zone != cardZoneHand || effect.chooseCard.binding == "" {
					return nil, fmt.Errorf("%s: %s.choose_card requires hand zone and binding", context, field)
				}
				if _, exists := bindings[effect.chooseCard.binding]; exists {
					return nil, fmt.Errorf("%s: %s.choose_card.binding %q is duplicated", context, field, effect.chooseCard.binding)
				}
				bindings[effect.chooseCard.binding] = struct{}{}
				result.effects = append(result.effects, compiledEffect{
					kind:       effectKindChooseCard,
					chooseCard: *effect.chooseCard,
				})
			case effectKindDiscard:
				if ability.kind != abilityKindCardistry || effect.discard == nil || effect.draw != nil || effect.damage != nil || effect.chooseCard != nil {
					return nil, fmt.Errorf("%s: %s.discard payload is required only for Cardistry", context, field)
				}
				if effect.discard.binding == "" {
					return nil, fmt.Errorf("%s: %s.discard.binding must not be empty", context, field)
				}
				if _, exists := bindings[effect.discard.binding]; !exists {
					return nil, fmt.Errorf("%s: %s.discard.binding %q has no preceding choice", context, field, effect.discard.binding)
				}
				result.effects = append(result.effects, compiledEffect{
					kind:    effectKindDiscard,
					discard: *effect.discard,
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

// validateReplacementDefinition 限制 replacement event filter 與 transformation 為 Infernal Vessel 已證明的 recover 規則。
// 輸入為編寫的 replacement 定義；輸出為驗證錯誤或 nil，副作用為零。
func validateReplacementDefinition(definition *replacementDefinition) error {
	if definition == nil {
		return fmt.Errorf("definition is required")
	}
	if definition.event != replacementEventRecover {
		return fmt.Errorf("event unsupported %q", definition.event)
	}
	if definition.transformation != replacementTransformReduce {
		return fmt.Errorf("transformation unsupported %q", definition.transformation)
	}
	if definition.amount != 3 {
		return fmt.Errorf("amount must be 3")
	}
	return nil
}

// validateStaticEffectDefinition 限制目前支援集合可使用的 static predicate、layer、duration 與 modifier。
// 輸入為編寫的靜態效果；輸出為驗證錯誤或 nil，副作用為零。
func validateStaticEffectDefinition(effect *staticEffectDefinition) error {
	if effect == nil {
		return fmt.Errorf("definition is required")
	}
	if effect.layer != effectLayerAbility || effect.sublayer != effectSubLayerNone || effect.duration != staticDurationWhilePredicate || effect.sourcePresence != sourcePresenceRequired {
		return fmt.Errorf("layer, sublayer, duration, or source presence unsupported")
	}
	switch effect.predicate {
	case staticPredicateSelf:
		if !isPrideStaticModifier(effect.modifier) {
			return fmt.Errorf("self predicate requires nonnegative Pride modifier")
		}
	case staticPredicateControlsQualifiedHumanAlly:
		if !isQualifiedHumanStaticModifier(effect.modifier) {
			return fmt.Errorf("qualified Human predicate requires Pride removal and granted On Attack")
		}
	default:
		return fmt.Errorf("predicate unsupported %q", effect.predicate)
	}
	return nil
}

// isPrideStaticModifier 驗證 Pride static 只設定非負 Pride，不混入未支援的特徵修正。
// 輸入為靜態能力的 modifier；輸出為是否符合 Pride 定義，副作用為零。
func isPrideStaticModifier(modifier continuousModifier) bool {
	return modifier.SetPride != nil &&
		*modifier.SetPride >= 0 &&
		modifier.SetPower == nil &&
		modifier.SetLife == nil &&
		modifier.PowerDelta == 0 &&
		modifier.LifeDelta == 0 &&
		modifier.ReserveCostDelta == 0 &&
		!modifier.GrantImmortality &&
		!modifier.ProhibitRecover &&
		!modifier.SwitchPowerLife &&
		!modifier.GrantStealth &&
		!modifier.GrantTrueSight &&
		!modifier.RemovePride &&
		!modifier.GrantRedHareOnAttack
}

// isQualifiedHumanStaticModifier 驗證條件式 static 只移除 Pride 並授予對應的 On Attack ability。
// 輸入為靜態能力的 modifier；輸出為是否符合 qualified Human 定義，副作用為零。
func isQualifiedHumanStaticModifier(modifier continuousModifier) bool {
	return modifier.SetPride == nil &&
		modifier.SetPower == nil &&
		modifier.SetLife == nil &&
		modifier.PowerDelta == 0 &&
		modifier.LifeDelta == 0 &&
		modifier.ReserveCostDelta == 0 &&
		!modifier.GrantImmortality &&
		!modifier.ProhibitRecover &&
		!modifier.SwitchPowerLife &&
		!modifier.GrantStealth &&
		!modifier.GrantTrueSight &&
		modifier.RemovePride &&
		modifier.GrantRedHareOnAttack
}

// cloneStaticEffectDefinition 複製 static 定義及其指標欄位，隔離編寫資料和編譯結果。
// 輸入為已驗證的靜態效果；輸出為獨立的靜態效果，副作用為零。
func cloneStaticEffectDefinition(effect staticEffectDefinition) staticEffectDefinition {
	clone := effect
	if effect.modifier.SetPride != nil {
		pride := *effect.modifier.SetPride
		clone.modifier.SetPride = &pride
	}
	return clone
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
		case effectKindChooseCard:
			operations = append(operations, effectOperation{
				Kind:     effectOperationChooseZoneCard,
				CardZone: effect.chooseCard.zone,
				Binding:  effect.chooseCard.binding,
			})
		case effectKindDiscard:
			operations = append(operations, effectOperation{
				Kind:    effectOperationDiscard,
				Binding: effect.discard.binding,
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
	for _, id := range []CardID{wonderlandsReignCardID, straightFlareCardID, threeOfHeartsCardID, impactHammerCardID, redHareCardID, infernalVesselCardID} {
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
