package game

import (
	"fmt"
	"regexp"
	"strings"
)

type abilityDefinitionKind string

const abilityKindCardistry abilityDefinitionKind = "cardistry"

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

const effectKindDraw authoredEffectKind = "draw"

type abilityReference string

const referenceController abilityReference = "controller"

type drawEffectDefinition struct {
	amount    int
	recipient abilityReference
}

type authoredEffectDefinition struct {
	kind authoredEffectKind
	draw *drawEffectDefinition
}

type authoredAbilityDefinition struct {
	slot      AbilitySlotID
	kind      abilityDefinitionKind
	timing    abilityTiming
	usage     abilityUsage
	reduction abilityCostReduction
	baseCost  int
	effects   []authoredEffectDefinition
}

// compiledEffect 保存已驗證的具體效果 payload，不讓 authoring tag 進入對局狀態。
type compiledEffect struct {
	draw drawEffectDefinition
}

// compiledAbilityDefinition 是建局前完成驗證的能力中介表示。
type compiledAbilityDefinition struct {
	slot      AbilitySlotID
	kind      abilityDefinitionKind
	timing    abilityTiming
	usage     abilityUsage
	reduction abilityCostReduction
	baseCost  int
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

// authoredAbilitiesForCard 將固定卡牌的 Go 能力編寫資料交給同一個 compiler。
// 輸入為 Card ID；輸出為該卡目前已遷移的能力資料，副作用為零。
func authoredAbilitiesForCard(id CardID) []authoredAbilityDefinition {
	if id == wonderlandsReignCardID {
		return wonderlandsReignAbilities()
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
		if ability.kind != abilityKindCardistry {
			return nil, fmt.Errorf("%s: unknown kind %q", context, ability.kind)
		}
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
		for index, effect := range ability.effects {
			field := fmt.Sprintf("effects[%d]", index)
			if effect.kind != effectKindDraw {
				return nil, fmt.Errorf("%s: %s.kind unknown %q", context, field, effect.kind)
			}
			if effect.draw == nil {
				return nil, fmt.Errorf("%s: %s.draw payload is required", context, field)
			}
			if effect.draw.amount <= 0 {
				return nil, fmt.Errorf("%s: %s.draw.amount must be positive", context, field)
			}
			if effect.draw.recipient != referenceController {
				return nil, fmt.Errorf("%s: %s.draw.recipient unknown or wrong reference %q", context, field, effect.draw.recipient)
			}
			payload := *effect.draw
			result.effects = append(result.effects, compiledEffect{
				draw: payload,
			})
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

// operations 將已編譯效果轉成目前統一 Ability Runtime 的可序列化操作。
// 輸入為已驗證的能力；輸出為每次呼叫新建的操作序列，副作用為零。
func (ability compiledAbilityDefinition) operations() []effectOperation {
	operations := make([]effectOperation, 0, len(ability.effects))
	for _, effect := range ability.effects {
		operations = append(operations, effectOperation{
			Kind:   effectOperationDraw,
			Amount: effect.draw.amount,
		})
	}
	return operations
}

// compiledCardistry 取得來源卡的已編譯 Cardistry 定義，不讀取可變對局狀態以外的編寫資料。
// 輸入為卡牌實例；輸出為能力及存在旗標，副作用為零。
func (g *Game) compiledCardistry(card cardInstanceID) (compiledAbilityDefinition, bool) {
	definition, exists := g.definitions[g.state.Cards[card].Definition]
	if !exists {
		return compiledAbilityDefinition{}, false
	}
	for _, ability := range definition.abilities {
		if ability.kind == abilityKindCardistry {
			return ability, true
		}
	}
	return compiledAbilityDefinition{}, false
}

// compileReplayDefinitions 重建重播所需的固定可執行定義，不將編寫資料寫入 Game State。
// 輸入為零；輸出為目前引擎版本釘選的定義或驗證錯誤，副作用為零。
func compileReplayDefinitions() (map[CardID]CardDefinition, error) {
	definition := CardDefinition{
		id: wonderlandsReignCardID,
		face: CardFace{
			id: "face:0mf1ug6yfi:front",
		},
	}
	authored := wonderlandsReignAbilities()
	abilities, err := compileAbilityDefinitions(definition, authored)
	if err != nil {
		return nil, fmt.Errorf("compile replay definition: %w", err)
	}
	definition.abilities = abilities
	return map[CardID]CardDefinition{
		definition.id: definition,
	}, nil
}
