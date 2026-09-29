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

// TestCompileDuchessDefinition 驗證 Duchess 的選牌與免費 Action 複製完全由可執行定義描述。
// 輸入為 Duchess 的卡牌定義與 Go 編寫資料；輸出為具選牌條件、具名結果與複製效果的 Cardistry 操作，副作用為零。
func TestCompileDuchessDefinition(t *testing.T) {
	definition := CardDefinition{
		id: duchessCardID,
		face: CardFace{
			id: CardFaceID("face:qzv380ujf5:front"),
		},
	}
	abilities, err := compileAbilityDefinitions(
		definition,
		duchessAbilities(),
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	operations := abilities[0].operations()
	if len(operations) != 2 || operations[0].Kind != effectOperationChooseZoneCard || operations[0].Binding != "copy-source" || operations[1].Kind != effectOperationCopyAction || operations[1].Binding != "copy-source" {
		t.Fatalf("operations = %#v, want graveyard choice followed by bound copy action", operations)
	}
}

// TestCompileImpactHammerTriggeredDefinition 驗證 Impact Hammer 的 On Wield 能力由定義宣告事件條件與事件目標傷害。
// 輸入為 Impact Hammer 的卡牌與牌面；輸出為帶穩定 Slot、wield 事件條件及事件目標參照的 triggered 定義，副作用為零。
func TestCompileImpactHammerTriggeredDefinition(t *testing.T) {
	definition := CardDefinition{
		id: impactHammerCardID,
		face: CardFace{
			id: CardFaceID("face:chsbalegbs:front"),
		},
	}
	authored := impactHammerAbilities()
	abilities, err := compileAbilityDefinitions(
		definition,
		authored,
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 1 || abilities[0].kind != abilityKindTriggered || abilities[0].trigger == nil || abilities[0].trigger.event != eventKindWield {
		t.Fatalf("compiled triggered ability = %#v, want one On Wield definition", abilities)
	}
	operations := abilities[0].operations()
	if len(operations) != 1 || operations[0].Kind != effectOperationDamage || operations[0].TargetReference != referenceEventTarget || operations[0].Value == nil || operations[0].Value.Kind != valueConstant || operations[0].Value.Constant != 3 {
		t.Fatalf("triggered operations = %#v, want 3 damage to wield event target", operations)
	}
}

// TestCompileVeritaDefinitions 驗證 Verita 由同一份可執行定義提供替代付款、靜態效果及死亡後持續效果。
// 輸入為 Verita 的卡牌定義與 Go 編寫資料；輸出為包含三種能力行為的已編譯定義，副作用為零。
func TestCompileVeritaDefinitions(t *testing.T) {
	definition := CardDefinition{
		id: veritaCardID,
		face: CardFace{
			id: CardFaceID("face:4qc47amgpp:front"),
		},
	}
	authored := veritaAbilities()
	abilities, err := compileAbilityDefinitions(
		definition,
		authored,
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 3 {
		t.Fatalf("compiled Verita abilities = %#v, want three definitions", abilities)
	}
	if abilities[0].alternative == nil || abilities[0].alternative.Selection.Zone != cardZoneGraveyard || abilities[0].alternative.Selection.MinimumCards != 3 || abilities[0].alternative.Selection.PrintedReserveTotal != 10 || abilities[0].alternative.Payment != alternativeCostBanishGraveyard {
		t.Fatalf("compiled Verita alternative cost = %#v, want three Suited Allies totaling ten from graveyard", abilities[0].alternative)
	}
	if abilities[1].static == nil || abilities[1].static.predicate != staticPredicateOtherControlledSuitedAlly || !abilities[1].static.modifier.GrantImmortality {
		t.Fatalf("compiled Verita static ability = %#v, want other controlled Suited Ally immortality", abilities[1].static)
	}
	if abilities[2].deathModifier == nil || abilities[2].deathModifier.selector != selectorControlledSuited || abilities[2].deathModifier.duration != modifierDurationEndOfNextTurn || abilities[2].deathModifier.modifier.PowerDelta != 1 {
		t.Fatalf("compiled Verita death ability = %#v, want +1 power through owner's next turn", abilities[2].deathModifier)
	}
}

// TestCompileRedHareStaticDefinitions 驗證 Red Hare 將 Pride 的攻擊限制與條件解除定義為靜態能力。
// 輸入為 Red Hare 的卡牌與牌面；輸出為具 predicate、layer、sublayer、modifier、duration 及來源存在條件的編譯定義，副作用為零。
func TestCompileRedHareStaticDefinitions(t *testing.T) {
	definition := CardDefinition{
		id: redHareCardID,
		face: CardFace{
			id: CardFaceID("face:5du8f077ua:front"),
		},
	}
	abilities, err := compileAbilityDefinitions(
		definition,
		redHareAbilities(),
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 2 {
		t.Fatalf("compiled abilities = %#v, want two static abilities", abilities)
	}
	for _, ability := range abilities {
		if ability.kind != abilityKindStatic || ability.static == nil {
			t.Fatalf("compiled ability = %#v, want static definition", ability)
		}
		if ability.static.predicate == "" || ability.static.layer != effectLayerAbility || ability.static.sublayer != effectSubLayerNone || ability.static.duration != staticDurationWhilePredicate || ability.static.sourcePresence != sourcePresenceRequired {
			t.Fatalf("static definition = %#v, want explicit predicate, layer, sublayer, duration, and source presence", ability.static)
		}
	}
	if abilities[0].static.modifier.SetPride == nil || *abilities[0].static.modifier.SetPride != 3 {
		t.Fatalf("Pride modifier = %#v, want Pride 3", abilities[0].static.modifier)
	}
	if !abilities[1].static.modifier.RemovePride || !abilities[1].static.modifier.GrantRedHareOnAttack {
		t.Fatalf("qualified Human modifier = %#v, want Pride removal and granted On Attack", abilities[1].static.modifier)
	}
	invalid := redHareAbilities()
	invalid[0].static.layer = effectLayerModifier
	_, err = compileAbilityDefinitions(
		definition,
		invalid,
	)
	if err == nil || !strings.Contains(err.Error(), "static") {
		t.Fatalf("invalid static layer error = %v, want static validation failure", err)
	}
	invalid = redHareAbilities()
	invalid[0].static.modifier.PowerDelta = 1
	_, err = compileAbilityDefinitions(
		definition,
		invalid,
	)
	if err == nil || !strings.Contains(err.Error(), "Pride modifier") {
		t.Fatalf("unexpected static modifier error = %v, want static validation failure", err)
	}
}

// TestCompileInfernalVesselReplacementDefinition 驗證 Infernal Vessel 以 event filter 與 transformation 宣告 recover replacement。
// 輸入為 Infernal Vessel 的卡牌與牌面；輸出為具穩定 Slot、recover event filter 和減少 3 的編譯定義，副作用為零。
func TestCompileInfernalVesselReplacementDefinition(t *testing.T) {
	definition := CardDefinition{
		id: infernalVesselCardID,
		face: CardFace{
			id: CardFaceID("face:vgWgu1DUYv:front"),
		},
	}
	abilities, err := compileAbilityDefinitions(
		definition,
		infernalVesselAbilities(),
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 1 || abilities[0].kind != abilityKindReplacement || abilities[0].replacement == nil {
		t.Fatalf("compiled abilities = %#v, want one replacement definition", abilities)
	}
	if abilities[0].slot != "ability:vgWgu1DUYv:front:recover-reduce" || abilities[0].replacement.event != replacementEventRecover || abilities[0].replacement.transformation != replacementTransformReduce || abilities[0].replacement.amount != 3 {
		t.Fatalf("replacement definition = %#v, want recover reduced by 3", abilities[0].replacement)
	}
	for _, testCase := range []struct {
		name   string
		change func(*authoredAbilityDefinition)
		field  string
	}{
		{
			name: "unknown event",
			change: func(ability *authoredAbilityDefinition) {
				ability.replacement.event = "damage"
			},
			field: "event",
		},
		{
			name: "unknown transformation",
			change: func(ability *authoredAbilityDefinition) {
				ability.replacement.transformation = "prevent"
			},
			field: "transformation",
		},
		{
			name: "wrong reduction",
			change: func(ability *authoredAbilityDefinition) {
				ability.replacement.amount = 2
			},
			field: "amount",
		},
	} {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				authored := infernalVesselAbilities()
				invalid := authored[0]
				testCase.change(&invalid)
				_, err := compileAbilityDefinitions(
					definition,
					[]authoredAbilityDefinition{invalid},
				)
				if err == nil || !strings.Contains(err.Error(), testCase.field) || !strings.Contains(err.Error(), string(definition.id)) {
					t.Fatalf("compile error = %v, want %s with definition context", err, testCase.field)
				}
			},
		)
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
