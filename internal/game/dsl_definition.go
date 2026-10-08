package game

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"go-tcg/internal/constants"
)

// wonderlandsReignDSL 是唯一進入 production content loading 的 v1 DSL 卡牌定義。
//
//go:embed card_definitions/wonderlands_reign.json
var wonderlandsReignDSL []byte

type dslCardDefinition struct {
	SchemaVersion          int               `json:"schema_version"`
	CompiledContentVersion string            `json:"compiled_content_version"`
	Definition             CardID            `json:"definition"`
	Face                   CardFaceID        `json:"face"`
	Abilities              []json.RawMessage `json:"abilities"`
}

type dslAbilityHeader struct {
	Slot AbilitySlotID `json:"slot"`
}

type dslAbilityDefinition struct {
	Slot          AbilitySlotID         `json:"slot"`
	Kind          abilityDefinitionKind `json:"kind"`
	Timing        abilityTiming         `json:"timing"`
	Usage         abilityUsage          `json:"usage"`
	CostReduction abilityCostReduction  `json:"cost_reduction"`
	BaseCost      int                   `json:"base_cost"`
	Selector      selectorKind          `json:"selector,omitempty"`
	Cost          actionCostKind        `json:"cost,omitempty"`
	Effects       []dslEffectDefinition `json:"effects"`
}

type dslEffectDefinition struct {
	Kind       authoredEffectKind `json:"kind"`
	Amount     *int               `json:"amount,omitempty"`
	Recipient  abilityReference   `json:"recipient,omitempty"`
	Selector   selectorKind       `json:"selector,omitempty"`
	Expression *valueExpression   `json:"expression,omitempty"`
	Zone       cardZone           `json:"zone,omitempty"`
	Duration   modifierDuration   `json:"duration,omitempty"`
}

// compileContentAbilityDefinitions 在 content loading 階段選取唯一 DSL card，其餘卡牌沿用 Go 編寫資料，並全部交給相同 compiler。
// 輸入為已載入的 Card Definition；輸出為一次完成驗證的 immutable abilities，副作用為零且不在 resolution loop 解析內容。
func compileContentAbilityDefinitions(definition CardDefinition) ([]compiledAbilityDefinition, error) {
	if definition.id == wonderlandsReignCardID {
		return compileDSLCardDefinition(
			definition,
			wonderlandsReignDSL,
		)
	}
	return compileAbilityDefinitions(
		definition,
		authoredAbilitiesForCard(definition.id),
	)
}

// compileDSLCardDefinition 嚴格解碼單一版本化 DSL 文件，驗證 Definition、Face、Slot 與欄位後交給既有 compiler。
// 輸入為目標 Card Definition 與 JSON bytes；輸出為既有 compiled IR 或具完整內容脈絡的錯誤，副作用為零且失敗時不 fallback。
func compileDSLCardDefinition(definition CardDefinition, contents []byte) ([]compiledAbilityDefinition, error) {
	var document dslCardDefinition
	if err := decodeStrictDSLJSON(
		contents,
		&document,
	); err != nil {
		return nil, fmt.Errorf("decode DSL definition: %w", err)
	}
	if document.SchemaVersion != constants.CardDefinitionDSLSchemaVersion {
		return nil, fmt.Errorf(
			"definition %q face %q schema_version unsupported %d",
			definition.id,
			definition.face.id,
			document.SchemaVersion,
		)
	}
	if document.CompiledContentVersion != constants.FixedCompiledContentVersion {
		return nil, fmt.Errorf(
			"definition %q face %q compiled_content_version %q does not match %q",
			definition.id,
			definition.face.id,
			document.CompiledContentVersion,
			constants.FixedCompiledContentVersion,
		)
	}
	if document.Definition != definition.id {
		return nil, fmt.Errorf(
			"definition %q face %q field definition is %q",
			definition.id,
			definition.face.id,
			document.Definition,
		)
	}
	if document.Face != definition.face.id {
		return nil, fmt.Errorf(
			"definition %q face %q field face is %q",
			definition.id,
			definition.face.id,
			document.Face,
		)
	}

	abilityCount := len(document.Abilities)
	authored := make([]authoredAbilityDefinition, 0, abilityCount)
	for _, rawAbility := range document.Abilities {
		var header dslAbilityHeader
		if err := json.Unmarshal(
			rawAbility,
			&header,
		); err != nil {
			return nil, fmt.Errorf(
				"definition %q face %q slot %q: decode slot: %w",
				definition.id,
				definition.face.id,
				header.Slot,
				err,
			)
		}
		context := fmt.Sprintf(
			"definition %q face %q slot %q",
			definition.id,
			definition.face.id,
			header.Slot,
		)
		var ability dslAbilityDefinition
		if err := decodeStrictDSLJSON(
			rawAbility,
			&ability,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", context, err)
		}
		converted, err := convertDSLAbility(ability)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", context, err)
		}
		authored = append(
			authored,
			converted,
		)
	}

	compiled, err := compileAbilityDefinitions(
		definition,
		authored,
	)
	if err != nil {
		return nil, fmt.Errorf("compile DSL: %w", err)
	}
	return compiled, nil
}

// convertDSLAbility 將 v1 JSON 欄位複製到既有 Go authoring shape，並拒絕此 vertical slice 未使用的合法原語。
// 輸入為已嚴格解碼的 DSL ability；輸出為 compiler 可接受的 authored ability 或精確欄位錯誤，副作用為零。
func convertDSLAbility(ability dslAbilityDefinition) (authoredAbilityDefinition, error) {
	if ability.Selector != "" {
		if !isKnownDSLSelector(ability.Selector) {
			return authoredAbilityDefinition{}, fmt.Errorf("selector unknown %q", ability.Selector)
		}
		return authoredAbilityDefinition{}, fmt.Errorf("selector is unused for Wonderland's Reign")
	}
	if ability.Cost != "" {
		if ability.Cost != actionCostSacrificeControlledWeapon {
			return authoredAbilityDefinition{}, fmt.Errorf("cost unknown %q", ability.Cost)
		}
		return authoredAbilityDefinition{}, fmt.Errorf("cost is unused for Wonderland's Reign")
	}
	effectCount := len(ability.Effects)
	converted := authoredAbilityDefinition{
		slot:          ability.Slot,
		kind:          ability.Kind,
		timing:        ability.Timing,
		usage:         ability.Usage,
		reduction:     ability.CostReduction,
		baseCost:      ability.BaseCost,
		effects:       make([]authoredEffectDefinition, 0, effectCount),
		supportStatus: supportStatusSupported,
	}
	for index, effect := range ability.Effects {
		convertedEffect, err := convertDSLEffect(effect)
		if err != nil {
			return authoredAbilityDefinition{}, fmt.Errorf("effects[%d].%w", index, err)
		}
		converted.effects = append(
			converted.effects,
			convertedEffect,
		)
	}
	return converted, nil
}

// convertDSLEffect 驗證 v1 effect enum 與欄位使用方式，並只將既有 draw primitive 映射到 authoring shape。
// 輸入為單一 DSL effect；輸出為既有 authored effect 或欄位錯誤，副作用為零且不建立新 operation kind。
func convertDSLEffect(effect dslEffectDefinition) (authoredEffectDefinition, error) {
	if effect.Selector != "" {
		if !isKnownDSLSelector(effect.Selector) {
			return authoredEffectDefinition{}, fmt.Errorf("selector unknown %q", effect.Selector)
		}
		return authoredEffectDefinition{}, fmt.Errorf("selector is unused for draw")
	}
	if effect.Expression != nil {
		if err := validateValueExpression(*effect.Expression); err != nil {
			return authoredEffectDefinition{}, fmt.Errorf("expression: %w", err)
		}
		return authoredEffectDefinition{}, fmt.Errorf("expression is unused for draw")
	}
	if effect.Zone != "" {
		if effect.Zone != cardZoneHand && effect.Zone != cardZoneGraveyard && effect.Zone != cardZoneMemory {
			return authoredEffectDefinition{}, fmt.Errorf("zone unknown %q", effect.Zone)
		}
		return authoredEffectDefinition{}, fmt.Errorf("zone is unused for draw")
	}
	if effect.Duration != "" {
		if effect.Duration != modifierDurationEndOfTurn && effect.Duration != modifierDurationEndOfNextTurn {
			return authoredEffectDefinition{}, fmt.Errorf("duration unknown %q", effect.Duration)
		}
		return authoredEffectDefinition{}, fmt.Errorf("duration is unused for draw")
	}
	if effect.Kind != effectKindDraw {
		return authoredEffectDefinition{}, fmt.Errorf("kind unknown %q", effect.Kind)
	}
	if effect.Amount == nil || *effect.Amount <= 0 {
		return authoredEffectDefinition{}, fmt.Errorf("amount must be positive")
	}
	return authoredEffectDefinition{
		kind: effectKindDraw,
		draw: &drawEffectDefinition{
			amount:    *effect.Amount,
			recipient: effect.Recipient,
		},
	}, nil
}

// isKnownDSLSelector 判定 selector 是否已存在於目前 Go compiler 的封閉原語集合。
// 輸入為 selector kind；輸出為是否為既有原語，副作用為零。
func isKnownDSLSelector(selector selectorKind) bool {
	switch selector {
	case selectorUnits,
		selectorControlledSuited,
		selectorControlledSuitedAlly,
		selectorAllAllies,
		selectorRetargetableControlledSuitedAlly,
		selectorOtherControlledAllies:
		return true
	default:
		return false
	}
}

// decodeStrictDSLJSON 解碼單一 JSON 值並拒絕未知欄位、第二個值及尾端垃圾。
// 輸入為 JSON bytes 與目標指標；成功時填入目標，失敗時回傳欄位或格式錯誤，副作用僅修改呼叫者提供的目標。
func decodeStrictDSLJSON(contents []byte, target any) error {
	reader := bytes.NewReader(contents)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("contains more than one JSON value")
		}
		return fmt.Errorf("decode trailing data: %w", err)
	}
	return nil
}
