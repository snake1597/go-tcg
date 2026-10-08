package game

import (
	"reflect"
	"strings"
	"testing"
)

// TestCompileWonderlandsReignDSLMatchesGoDefinition 驗證版本化 DSL 與既有 Go 編寫資料產生完全相同的 immutable 中介表示。
// 輸入為內嵌 Wonderland's Reign DSL 與既有 Card Definition；輸出為逐欄相同的 compiled abilities，副作用為零。
func TestCompileWonderlandsReignDSLMatchesGoDefinition(t *testing.T) {
	definition := CardDefinition{
		id: wonderlandsReignCardID,
		face: CardFace{
			id: "face:0mf1ug6yfi:front",
		},
	}
	want, err := compileAbilityDefinitions(
		definition,
		wonderlandsReignAbilities(),
	)
	if err != nil {
		t.Fatalf("compile Go definition: %v", err)
	}
	got, err := compileDSLCardDefinition(
		definition,
		wonderlandsReignDSL,
	)
	if err != nil {
		t.Fatalf("compile DSL definition: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("compiled DSL = %#v, want Go definition %#v", got, want)
	}
}

// TestCompileWonderlandsReignDSLRejectsInvalidSchema 驗證 DSL 邊界拒絕未知原語、錯誤型別、未使用欄位及舊版本。
// 輸入為逐一破壞的 Wonderland's Reign DSL；輸出為含 Definition、Face、Slot 與欄位的錯誤，副作用為零。
func TestCompileWonderlandsReignDSLRejectsInvalidSchema(t *testing.T) {
	definition := CardDefinition{
		id: wonderlandsReignCardID,
		face: CardFace{
			id: "face:0mf1ug6yfi:front",
		},
	}
	valid := string(wonderlandsReignDSL)
	tests := []struct {
		name        string
		old         string
		replacement string
		field       string
		ability     bool
	}{
		{
			name:        "unsupported schema version",
			old:         `"schema_version": 1`,
			replacement: `"schema_version": 0`,
			field:       "schema_version",
		},
		{
			name:        "unsupported compiled content version",
			old:         `"compiled_content_version": "compiled-content-v1"`,
			replacement: `"compiled_content_version": "compiled-content-v0"`,
			field:       "compiled_content_version",
		},
		{
			name:        "unknown ability kind",
			old:         `"kind": "cardistry"`,
			replacement: `"kind": "unknown"`,
			field:       "kind",
			ability:     true,
		},
		{
			name:        "unknown operation kind",
			old:         `"kind": "draw"`,
			replacement: `"kind": "teleport"`,
			field:       "effects[0].kind",
			ability:     true,
		},
		{
			name:        "unknown selector",
			old:         `"base_cost": 10,`,
			replacement: `"base_cost": 10, "selector": "hidden-hand",`,
			field:       "selector",
			ability:     true,
		},
		{
			name:        "unknown expression",
			old:         `"amount": 1,`,
			replacement: `"amount": 1, "expression": {"kind": "multiply"},`,
			field:       "effects[0].expression",
			ability:     true,
		},
		{
			name:        "unknown zone",
			old:         `"amount": 1,`,
			replacement: `"amount": 1, "zone": "secret",`,
			field:       "effects[0].zone",
			ability:     true,
		},
		{
			name:        "unknown duration",
			old:         `"amount": 1,`,
			replacement: `"amount": 1, "duration": "forever",`,
			field:       "effects[0].duration",
			ability:     true,
		},
		{
			name:        "unknown reference",
			old:         `"recipient": "controller"`,
			replacement: `"recipient": "spectator"`,
			field:       "effects[0].draw.recipient",
			ability:     true,
		},
		{
			name:        "unknown cost",
			old:         `"base_cost": 10,`,
			replacement: `"base_cost": 10, "cost": "discard-random",`,
			field:       "cost",
			ability:     true,
		},
		{
			name:        "wrong type",
			old:         `"base_cost": 10`,
			replacement: `"base_cost": "ten"`,
			field:       "base_cost",
			ability:     true,
		},
		{
			name:        "invalid value",
			old:         `"amount": 1`,
			replacement: `"amount": -1`,
			field:       "effects[0].amount",
			ability:     true,
		},
		{
			name:        "unused valid field",
			old:         `"base_cost": 10,`,
			replacement: `"base_cost": 10, "selector": "units",`,
			field:       "selector",
			ability:     true,
		},
		{
			name:        "unknown field",
			old:         `"base_cost": 10,`,
			replacement: `"base_cost": 10, "mystery": true,`,
			field:       "mystery",
			ability:     true,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				contents := strings.Replace(valid, test.old, test.replacement, 1)
				if contents == valid {
					t.Fatalf("test mutation did not match %q", test.old)
				}
				_, err := compileDSLCardDefinition(
					definition,
					[]byte(contents),
				)
				if err == nil {
					t.Fatalf("compile error = nil, want field %q", test.field)
				}
				errorMessage := err.Error()
				if !strings.Contains(errorMessage, test.field) {
					t.Fatalf("compile error = %v, want field %q", err, test.field)
				}
				if test.ability {
					for _, want := range []string{
						string(definition.id),
						string(definition.face.id),
						"ability:0mf1ug6yfi:front:cardistry-draw",
					} {
						if !strings.Contains(errorMessage, want) {
							t.Fatalf("compile error %q missing context %q", err, want)
						}
					}
				}
			},
		)
	}
}
