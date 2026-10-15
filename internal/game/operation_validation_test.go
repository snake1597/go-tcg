package game

import (
	"strings"
	"testing"
)

// TestValidateAbilityOperationsRejectsInvalidRuntimeData 驗證載入前攔截未知操作、錯誤參照與無效參數。
// 輸入為有缺陷的 operation sequences；輸出為含操作索引與欄位的錯誤，無遊戲狀態副作用。
func TestValidateAbilityOperationsRejectsInvalidRuntimeData(t *testing.T) {
	tests := []struct {
		name       string
		operations []effectOperation
		field      string
	}{
		{
			name: "unknown operation",
			operations: []effectOperation{
				{
					Kind: "unknown",
				},
			},
			field: "operations[0].kind",
		},
		{
			name: "unbound Object reference",
			operations: []effectOperation{
				{
					Kind:          effectOperationDamage,
					TargetBinding: "missing",
				},
			},
			field: "operations[0].target_binding",
		},
		{
			name: "invalid draw amount",
			operations: []effectOperation{
				{
					Kind: effectOperationDraw,
				},
			},
			field: "operations[0].amount",
		},
		{
			name: "unbound Card reference",
			operations: []effectOperation{
				{
					Kind:    effectOperationCopyAction,
					Binding: "missing",
				},
			},
			field: "operations[0].binding",
		},
		{
			name: "unused payload",
			operations: []effectOperation{
				{
					Kind:     effectOperationDraw,
					Amount:   1,
					Selector: selectorAllAllies,
				},
			},
			field: "selector is unused",
		},
		{
			name: "invalid Card filter",
			operations: []effectOperation{
				{
					Kind: effectOperationChoose,
					CardSelection: cardSelectionSpec{
						Zone:         cardZoneHand,
						RequiredType: "UNKNOWN",
					},
					Binding: "card",
				},
			},
			field: "card_selection",
		},
		{
			name: "empty continuous effect",
			operations: []effectOperation{
				{
					Kind:     effectOperationContinuousModifier,
					Duration: modifierDurationEndOfTurn,
				},
			},
			field: "continuous_effect",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateAbilityOperations(test.operations)
			if err == nil || !strings.Contains(err.Error(), test.field) {
				t.Fatalf("validateAbilityOperations() error = %v, want %q", err, test.field)
			}
		})
	}
}

// TestCompileAbilityDefinitionsRejectsInvalidOperationSelection 驗證 Card Definition 編譯會套用 operation payload gate。
// 輸入為錯誤選牌篩選的 Red Hare 定義；輸出為載入前錯誤，無對局狀態副作用。
func TestCompileAbilityDefinitionsRejectsInvalidOperationSelection(t *testing.T) {
	definition := CardDefinition{
		id: redHareCardID,
		face: CardFace{
			id: CardFaceID("face:5du8f077ua:front"),
		},
	}
	authored := redHareAbilities()
	authored[2].effects[0].chooseCard.selection.RequiredType = "UNKNOWN"
	_, err := compileAbilityDefinitions(definition, authored)
	if err == nil || !strings.Contains(err.Error(), "card_selection") {
		t.Fatalf("compileAbilityDefinitions() error = %v, want pre-game selection rejection", err)
	}
}
