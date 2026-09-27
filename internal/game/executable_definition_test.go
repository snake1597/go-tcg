package game

import (
	"strings"
	"testing"
)

// TestCompileWonderlandsReignDefinition 驗證 Go 編寫資料在建局前編譯成具穩定 Slot 的 typed 定義。
// 輸入為 Wonderland's Reign 的卡牌與牌面；輸出為費用、Slot 與抽牌效果，副作用為零。
func TestCompileWonderlandsReignDefinition(t *testing.T) {
	definition := CardDefinition{
		id: wonderlandsReignCardID,
		face: CardFace{
			id: CardFaceID("face:0mf1ug6yfi:front"),
		},
	}
	authored := wonderlandsReignAbilities()
	abilities, err := compileAbilityDefinitions(definition, authored)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 1 || abilities[0].slot != "ability:0mf1ug6yfi:front:cardistry-draw" || abilities[0].baseCost != 10 || abilities[0].usage != usageOncePerObject || abilities[0].reduction != reductionDistinctSuitedCosts {
		t.Fatalf("abilities = %#v, want Cardistry draw slot with cost 10", abilities)
	}
}

// TestCompileThreeOfHeartsDefinitionRejectsUnboundDiscard 驗證 Three of Hearts 的棄牌只能引用同一能力中先前具名的選牌 binding。
// 輸入為含未知 binding 的 Three of Hearts 編寫資料；輸出為附帶欄位脈絡的編譯錯誤，副作用為零。
func TestCompileThreeOfHeartsDefinitionRejectsUnboundDiscard(t *testing.T) {
	definition := CardDefinition{
		id: threeOfHeartsCardID,
		face: CardFace{
			id: CardFaceID("face:1db8hz4prm:front"),
		},
	}
	authored := threeOfHeartsAbilities()
	authored[0].effects[2].discard.binding = "unknown-card"
	_, err := compileAbilityDefinitions(definition, authored)
	if err == nil {
		t.Fatal("compileAbilityDefinitions() error = nil")
	}
	for _, want := range []string{"1db8hz4prm", "discard.binding", "unknown-card"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}
}

// TestCompileAbilityDefinitionsRejectsInvalidData 驗證 compiler 拒絕重複 ID、錯誤關聯、未知 kind 與 payload。
// 輸入為數組無效 Go 定義；輸出為含 Definition、Face、Slot 與欄位的錯誤，副作用為零。
func TestCompileAbilityDefinitionsRejectsInvalidData(t *testing.T) {
	definition := CardDefinition{
		id: wonderlandsReignCardID,
		face: CardFace{
			id: CardFaceID("face:0mf1ug6yfi:front"),
		},
	}
	valid := wonderlandsReignAbilities()[0]
	tests := []struct {
		name    string
		entries []authoredAbilityDefinition
		field   string
	}{
		{
			name: "duplicate slot",
			entries: []authoredAbilityDefinition{
				valid,
				valid,
			},
			field: "slot",
		},
		{
			name: "wrong face",
			entries: []authoredAbilityDefinition{
				{
					slot:      "ability:0mf1ug6yfi:back:cardistry-draw",
					kind:      abilityKindCardistry,
					timing:    valid.timing,
					usage:     valid.usage,
					reduction: valid.reduction,
					baseCost:  10,
					effects:   valid.effects,
				},
			},
			field: "slot",
		},
		{
			name: "unregistered semantic slot",
			entries: []authoredAbilityDefinition{
				{
					slot:      "ability:0mf1ug6yfi:front:other-cardistry",
					kind:      valid.kind,
					timing:    valid.timing,
					usage:     valid.usage,
					reduction: valid.reduction,
					baseCost:  valid.baseCost,
					effects:   valid.effects,
				},
			},
			field: "slot",
		},
		{
			name: "unknown ability kind",
			entries: []authoredAbilityDefinition{
				{
					slot:      valid.slot,
					kind:      "unknown",
					timing:    valid.timing,
					usage:     valid.usage,
					reduction: valid.reduction,
					baseCost:  10,
					effects:   valid.effects,
				},
			},
			field: "kind",
		},
		{
			name: "unknown effect kind",
			entries: []authoredAbilityDefinition{
				{
					slot:      valid.slot,
					kind:      abilityKindCardistry,
					timing:    valid.timing,
					usage:     valid.usage,
					reduction: valid.reduction,
					baseCost:  10,
					effects: []authoredEffectDefinition{
						{
							kind: "unknown",
						},
					},
				},
			},
			field: "effects[0].kind",
		},
		{
			name: "invalid payload",
			entries: []authoredAbilityDefinition{
				{
					slot:      valid.slot,
					kind:      abilityKindCardistry,
					timing:    valid.timing,
					usage:     valid.usage,
					reduction: valid.reduction,
					baseCost:  10,
					effects: []authoredEffectDefinition{
						{
							kind: effectKindDraw,
							draw: &drawEffectDefinition{
								amount:    0,
								recipient: referenceController,
							},
						},
					},
				},
			},
			field: "effects[0].draw.amount",
		},
		{
			name: "unknown reference",
			entries: []authoredAbilityDefinition{
				{
					slot:      valid.slot,
					kind:      abilityKindCardistry,
					timing:    valid.timing,
					usage:     valid.usage,
					reduction: valid.reduction,
					baseCost:  10,
					effects: []authoredEffectDefinition{
						{
							kind: effectKindDraw,
							draw: &drawEffectDefinition{
								amount:    1,
								recipient: "unknown",
							},
						},
					},
				},
			},
			field: "effects[0].draw.recipient",
		},
	}
	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				_, err := compileAbilityDefinitions(definition, test.entries)
				if err == nil {
					t.Fatal("compileAbilityDefinitions() error = nil")
				}
				for _, want := range []string{"0mf1ug6yfi", "face:0mf1ug6yfi:front", "ability:0mf1ug6yfi", test.field} {
					message := err.Error()
					if !strings.Contains(message, want) {
						t.Fatalf("error %q missing %q", err, want)
					}
				}
			},
		)
	}
}
