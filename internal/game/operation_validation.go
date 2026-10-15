package game

import (
	"fmt"
	"reflect"
)

// validateAbilityOperations 在載入 Card Definition 時驗證可序列化操作的種類、參照與必要參數。
// 輸入為編譯後的有序操作；輸出為第一個無效操作的錯誤或 nil，無副作用。
func validateAbilityOperations(operations []effectOperation) error {
	cardBindings := make(map[resolutionBinding]cardSelectionSpec)
	objectBindings := make(map[resolutionBinding]struct{})
	for index, operation := range operations {
		field := fmt.Sprintf("operations[%d]", index)
		if !isSupportedOperation(operation.Kind) {
			return fmt.Errorf("%s.kind unknown %q", field, operation.Kind)
		}
		if err := validateOperationPayload(operation); err != nil {
			return fmt.Errorf("%s: %w", field, err)
		}
		if operation.TargetReference != "" && operation.TargetReference != referenceDeclaredTarget && operation.TargetReference != referenceEventTarget {
			return fmt.Errorf("%s.target_reference unknown %q", field, operation.TargetReference)
		}
		if operation.TargetBinding != "" {
			if _, exists := objectBindings[operation.TargetBinding]; !exists {
				return fmt.Errorf("%s.target_binding %q has no preceding Object choice", field, operation.TargetBinding)
			}
		}
		switch operation.Kind {
		case effectOperationChoose:
			if operation.CardSelection.Zone != "" {
				if operation.Binding == "" || operation.Selector != "" || !validCardChoiceZone(operation.CardSelection.Zone) {
					return fmt.Errorf("%s requires a supported Card zone and binding", field)
				}
				if err := validateOperationCardSelection(operation.CardSelection); err != nil {
					return fmt.Errorf("%s.card_selection: %w", field, err)
				}
				if _, exists := cardBindings[operation.Binding]; exists {
					return fmt.Errorf("%s.binding %q is duplicated", field, operation.Binding)
				}
				cardBindings[operation.Binding] = operation.CardSelection
			} else {
				if !validObjectChoiceSelector(operation.Selector) {
					return fmt.Errorf("%s.selector unsupported %q", field, operation.Selector)
				}
				if operation.Binding != "" {
					if _, exists := objectBindings[operation.Binding]; exists {
						return fmt.Errorf("%s.binding %q is duplicated", field, operation.Binding)
					}
					objectBindings[operation.Binding] = struct{}{}
				}
			}
		case effectOperationDraw, effectOperationDrawToMemory, effectOperationCardistryDiscount:
			if operation.Amount <= 0 {
				return fmt.Errorf("%s.amount must be positive", field)
			}
		case effectOperationDamage:
			if operation.Value == nil || (operation.TargetReference == "" && operation.TargetBinding == "") || (operation.TargetReference != "" && operation.TargetBinding != "") {
				return fmt.Errorf("%s requires a value and exactly one target reference", field)
			}
			if err := validateValueExpression(*operation.Value); err != nil {
				return fmt.Errorf("%s.value: %w", field, err)
			}
		case effectOperationCounter:
			if operation.Counter != "BUFF" || operation.Value == nil {
				return fmt.Errorf("%s requires a BUFF counter and value", field)
			}
			if err := validateValueExpression(*operation.Value); err != nil {
				return fmt.Errorf("%s.value: %w", field, err)
			}
		case effectOperationDiscard, effectOperationDeploy, effectOperationCopyAction:
			selection, exists := cardBindings[operation.Binding]
			if operation.Binding == "" || !exists {
				return fmt.Errorf("%s.binding %q has no preceding Card choice", field, operation.Binding)
			}
			if operation.Kind == effectOperationDeploy && (selection.Zone != cardZoneMemory || selection.RequiredType != "ALLY") {
				return fmt.Errorf("%s requires a Memory Ally choice", field)
			}
			if operation.Kind == effectOperationCopyAction && (selection.Zone != cardZoneGraveyard || selection.RequiredType != "ACTION" || !reflect.DeepEqual(operation.CardSelection, selection)) {
				return fmt.Errorf("%s requires a graveyard Action choice", field)
			}
		case effectOperationContinuousModifier:
			if operation.Duration != modifierDurationEndOfTurn && operation.Duration != modifierDurationEndOfNextTurn {
				return fmt.Errorf("%s.duration unsupported %q", field, operation.Duration)
			}
			if err := validateOperationContinuousEffect(operation); err != nil {
				return fmt.Errorf("%s.continuous_effect: %w", field, err)
			}
		case effectOperationSacrifice:
			if operation.RequiredType != "ALLY" && operation.RequiredType != "WEAPON" {
				return fmt.Errorf("%s.required_type unsupported %q", field, operation.RequiredType)
			}
		}
	}
	return nil
}

// validateOperationCardSelection 驗證目前可執行選牌操作的區域、篩選列舉與數值界限。
// 輸入為 Card selection；輸出為無效欄位錯誤或 nil，無副作用。
func validateOperationCardSelection(selection cardSelectionSpec) error {
	if !validCardChoiceZone(selection.Zone) {
		return fmt.Errorf("zone unsupported %q", selection.Zone)
	}
	if selection.RequiredType != "" && selection.RequiredType != "ALLY" && selection.RequiredType != "ACTION" {
		return fmt.Errorf("required_type unsupported %q", selection.RequiredType)
	}
	if selection.RequiredSubtype != "" && selection.RequiredSubtype != "SUITED" {
		return fmt.Errorf("required_subtype unsupported %q", selection.RequiredSubtype)
	}
	if selection.RequiredElement != "" && !validSelectionElement(selection.RequiredElement) {
		return fmt.Errorf("required_element unsupported %q", selection.RequiredElement)
	}
	if selection.RequiredElement != "" && len(selection.RequiredElements) != 0 {
		return fmt.Errorf("required_element and required_elements are mutually exclusive")
	}
	seen := make(map[string]struct{}, len(selection.RequiredElements))
	for _, element := range selection.RequiredElements {
		if !validSelectionElement(element) {
			return fmt.Errorf("required_elements contains unsupported %q", element)
		}
		if _, exists := seen[element]; exists {
			return fmt.Errorf("required_elements contains duplicate %q", element)
		}
		seen[element] = struct{}{}
	}
	if selection.MaximumReserveCost < 0 || selection.MinimumCards != 0 || selection.MatchPrintedReserveTotal || selection.PrintedReserveTotal != 0 {
		return fmt.Errorf("reserve cost or multi-card selection parameters unsupported")
	}
	return nil
}

// validSelectionElement 判斷選牌條件是否引用目前引擎認識的元素。
// 輸入為元素名稱；輸出為布林值，無副作用。
func validSelectionElement(element string) bool {
	switch element {
	case "FIRE", "WATER", "WIND", "TERA", "NORM":
		return true
	default:
		return false
	}
}

// validateOperationContinuousEffect 驗證持續效果的靜態資料與到期語意；來源、目標與時間戳只能由結算時填入。
// 輸入為 modifier operation；輸出為無效 payload 錯誤或 nil，無副作用。
func validateOperationContinuousEffect(operation effectOperation) error {
	effect := operation.ContinuousEffect
	if effect.ID != 0 || effect.Source != "" || effect.Controller != nil || effect.Target != "" || effect.Timestamp != 0 || len(effect.DependsOn) != 0 || effect.ExpiresAtTurn != 0 {
		return fmt.Errorf("runtime identity or expiry must not be prefilled")
	}
	if effect.Scope != effectScopeObject || (effect.SubLayer != "" && effect.SubLayer != effectSubLayerNone) || effect.PowerLife != powerLifeModify {
		return fmt.Errorf("scope, sublayer or power/life operation unsupported")
	}
	if operation.Duration == modifierDurationEndOfNextTurn {
		if effect.Layer != effectLayerAbility || !isOtherControlledSuitedAllyModifier(effect.Modifier) {
			return fmt.Errorf("next-turn ability modifier unsupported")
		}
		return nil
	}
	if effect.Layer != effectLayerModifier || (!isCardistryModifier(effect.Modifier) && !isActionModifier(effect.Modifier) && !isCardistryObserverModifier(effect.Modifier) && !isNoireStealthModifier(effect.Modifier)) {
		return fmt.Errorf("end-of-turn modifier unsupported")
	}
	return nil
}

// validateOperationPayload 拒絕每種操作不使用的欄位，避免序列化資料藏有第二種解讀。
// 輸入為單一 operation；輸出為第一個多餘欄位的錯誤或 nil，無副作用。
func validateOperationPayload(operation effectOperation) error {
	kind := operation.Kind
	if operation.Target != "" || operation.Source != "" || operation.ExcludedObject != "" || len(operation.Options) != 0 {
		return fmt.Errorf("runtime-only target, source, exclusion or options are unsupported in a definition")
	}
	if operation.TargetReference != "" && kind != effectOperationDamage {
		return fmt.Errorf("target_reference is unused for %q", kind)
	}
	if operation.TargetBinding != "" && kind != effectOperationDamage && kind != effectOperationCounter && kind != effectOperationSacrifice {
		return fmt.Errorf("target_binding is unused for %q", kind)
	}
	if operation.Amount != 0 && kind != effectOperationDraw && kind != effectOperationDrawToMemory && kind != effectOperationCardistryDiscount {
		return fmt.Errorf("amount is unused for %q", kind)
	}
	if operation.Counter != "" && kind != effectOperationCounter {
		return fmt.Errorf("counter is unused for %q", kind)
	}
	if operation.Value != nil && kind != effectOperationDamage && kind != effectOperationCounter {
		return fmt.Errorf("value is unused for %q", kind)
	}
	if operation.CanPass && kind != effectOperationChoose {
		return fmt.Errorf("can_pass is unused for %q", kind)
	}
	if operation.CardZone != "" && kind != effectOperationDiscard {
		return fmt.Errorf("card_zone is unused for %q", kind)
	}
	if !reflect.ValueOf(operation.CardSelection).IsZero() && kind != effectOperationChoose && kind != effectOperationCopyAction {
		return fmt.Errorf("card_selection is unused for %q", kind)
	}
	if operation.Binding != "" && kind != effectOperationChoose && kind != effectOperationDiscard && kind != effectOperationDeploy && kind != effectOperationCopyAction {
		return fmt.Errorf("binding is unused for %q", kind)
	}
	if operation.RequiredType != "" && kind != effectOperationSacrifice {
		return fmt.Errorf("required_type is unused for %q", kind)
	}
	if operation.Selector != "" && kind != effectOperationChoose {
		return fmt.Errorf("selector is unused for %q", kind)
	}
	if operation.Duration != "" && kind != effectOperationContinuousModifier {
		return fmt.Errorf("duration is unused for %q", kind)
	}
	if operation.RequiresRetargetSuccess && kind != effectOperationContinuousModifier {
		return fmt.Errorf("requires_retarget_success is unused for %q", kind)
	}
	if !reflect.ValueOf(operation.ContinuousEffect).IsZero() && kind != effectOperationContinuousModifier {
		return fmt.Errorf("continuous_effect is unused for %q", kind)
	}
	return nil
}

// validCardChoiceZone 判斷 runtime Card choice 是否使用已支援區域。
// 輸入為區域；輸出為布林值，無副作用。
func validCardChoiceZone(zone cardZone) bool {
	return zone == cardZoneHand || zone == cardZoneGraveyard || zone == cardZoneMemory
}

// validObjectChoiceSelector 判斷 runtime Object choice 是否使用已支援的動態查詢。
// 輸入為 selector；輸出為布林值，無副作用。
func validObjectChoiceSelector(selector selectorKind) bool {
	switch selector {
	case selectorControlledSuitedAlly, selectorAllAllies, selectorOtherControlledAllies, selectorUnits:
		return true
	default:
		return false
	}
}
