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

// TestCompileRemainingCardistryDefinitions 驗證其餘 Support Set Cardistry 全部以穩定 Slot 與編譯定義提供。
// 輸入為六張固定卡牌及其 Go 編寫資料；輸出為各一個可執行 Cardistry 定義，副作用為零。
func TestCompileRemainingCardistryDefinitions(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		id         CardID
		slot       AbilitySlotID
		cost       int
		timing     abilityTiming
		operations []effectOperationKind
	}{
		{
			name:   "Five of Spades",
			id:     fiveOfSpadesCardID,
			slot:   "ability:i9hf5lhl5f:front:cardistry-power",
			cost:   5,
			timing: timingMainPhase,
			operations: []effectOperationKind{
				effectOperationContinuousModifier,
			},
		},
		{
			name:   "Four of Spades",
			id:     fourOfSpadesCardID,
			slot:   "ability:8bolq2y5qp:front:cardistry-draw-to-memory",
			cost:   4,
			timing: timingMainPhase,
			operations: []effectOperationKind{
				effectOperationDrawToMemory,
			},
		},
		{
			name:   "Four of Hearts",
			id:     fourOfHeartsCardID,
			slot:   "ability:xgax8bbjqj:front:cardistry-deploy",
			cost:   4,
			timing: timingMainPhase,
			operations: []effectOperationKind{
				effectOperationDrawToMemory,
				effectOperationChoose,
				effectOperationDeploy,
			},
		},
		{
			name:   "Three of Spades",
			id:     threeOfSpadesCardID,
			slot:   "ability:o09csnorqv:front:cardistry-life",
			cost:   3,
			timing: timingMainPhase,
			operations: []effectOperationKind{
				effectOperationChoose,
				effectOperationContinuousModifier,
			},
		},
		{
			name:   "Two of Hearts",
			id:     twoOfHeartsCardID,
			slot:   "ability:rufki4o41y:front:cardistry-power",
			cost:   2,
			timing: timingMainPhase,
			operations: []effectOperationKind{
				effectOperationContinuousModifier,
			},
		},
		{
			name:   "Two of Spades",
			id:     twoOfSpadesCardID,
			slot:   "ability:e8ygl32jef:front:cardistry-buff-counter",
			cost:   2,
			timing: timingFast,
			operations: []effectOperationKind{
				effectOperationCounter,
			},
		},
	} {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				definition := CardDefinition{
					id: testCase.id,
					face: CardFace{
						id: CardFaceID("face:" + string(testCase.id) + ":front"),
					},
				}
				abilities, err := compileAbilityDefinitions(
					definition,
					authoredAbilitiesForCard(testCase.id),
				)
				if err != nil {
					t.Fatalf("compileAbilityDefinitions() error = %v", err)
				}
				if len(abilities) != 1 || abilities[0].kind != abilityKindCardistry || abilities[0].slot != testCase.slot || abilities[0].baseCost != testCase.cost || abilities[0].timing != testCase.timing {
					t.Fatalf("compiled abilities = %#v, want stable executable Cardistry slot %q", abilities, testCase.slot)
				}
				operations := abilities[0].operations()
				if len(operations) != len(testCase.operations) {
					t.Fatalf("operations = %#v, want %d operations", operations, len(testCase.operations))
				}
				for index, operation := range operations {
					if operation.Kind != testCase.operations[index] {
						t.Fatalf("operations[%d] = %q, want %q", index, operation.Kind, testCase.operations[index])
					}
				}
			},
		)
	}
}

// TestCompileCardistryDefinitionRejectsMixedModifier 驗證 Cardistry modifier 不可混入未登錄的 continuous effect 欄位。
// 輸入為夾帶 Immortality 的 Two of Hearts 編寫資料；輸出為編譯錯誤，副作用為零。
func TestCompileCardistryDefinitionRejectsMixedModifier(t *testing.T) {
	authored := twoOfHeartsAbilities()
	authored[0].effects[0].modifier.modifier.GrantImmortality = true
	_, err := compileAbilityDefinitions(
		CardDefinition{
			id: twoOfHeartsCardID,
			face: CardFace{
				id: "face:rufki4o41y:front",
			},
		},
		authored,
	)
	if err == nil || !strings.Contains(err.Error(), "modifier") {
		t.Fatalf("compileAbilityDefinitions() error = %v, want modifier validation failure", err)
	}
	authored = twoOfHeartsAbilities()
	authored[0].cost = &actionCostDefinition{
		kind: actionCostSacrificeControlledWeapon,
	}
	_, err = compileAbilityDefinitions(
		CardDefinition{
			id: twoOfHeartsCardID,
			face: CardFace{
				id: "face:rufki4o41y:front",
			},
		},
		authored,
	)
	if err == nil || !strings.Contains(err.Error(), "Cardistry") {
		t.Fatalf("compileAbilityDefinitions() error = %v, want unused Action cost validation failure", err)
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
	if len(operations) != 2 || operations[0].Kind != effectOperationChoose || operations[0].Binding != "copy-source" || operations[1].Kind != effectOperationCopyAction || operations[1].Binding != "copy-source" {
		t.Fatalf("operations = %#v, want graveyard choice followed by bound copy action", operations)
	}
}

// TestCompileObjectChoiceBinding 驗證正式能力定義可將 Object 選擇綁定為名稱，並由後續 counter operation 引用。
// 輸入為含 Object 選擇與具名 counter 目標的 Cardistry 定義；輸出為保留 binding 的 typed operations；副作用為零。
func TestCompileObjectChoiceBinding(t *testing.T) {
	definition := CardDefinition{
		id: threeOfHeartsCardID,
		face: CardFace{
			id: CardFaceID("face:1db8hz4prm:front"),
		},
	}
	authored := threeOfHeartsAbilities()
	authored[0].effects = []authoredEffectDefinition{
		{
			kind: effectKindChooseObject,
			chooseObject: &chooseObjectEffectDefinition{
				selector: selectorControlledSuitedAlly,
				binding:  "chosen-object",
			},
		},
		{
			kind: effectKindCounter,
			counter: &counterEffectDefinition{
				counter:       "BUFF",
				amount:        1,
				targetBinding: "chosen-object",
			},
		},
	}
	abilities, err := compileAbilityDefinitions(definition, authored)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	operations := abilities[0].operations()
	if len(operations) != 2 || operations[0].Binding != "chosen-object" || operations[1].TargetBinding != "chosen-object" {
		t.Fatalf("operations = %#v, want Object binding and named counter target", operations)
	}
}

// TestCompileImpactHammerTriggeredDefinition 驗證 Impact Hammer 的 class bonus 與 On Wield 能力都由定義提供。
// 輸入為 Impact Hammer 的卡牌與牌面；輸出為 stable static 與 triggered slots，副作用為零。
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
	if len(abilities) != 2 || abilities[0].kind != abilityKindStatic || abilities[0].slot != "ability:chsbalegbs:front:class-bonus" || abilities[1].kind != abilityKindTriggered || abilities[1].trigger == nil || abilities[1].trigger.event != eventKindWield {
		t.Fatalf("compiled abilities = %#v, want class bonus and On Wield definitions", abilities)
	}
	operations := abilities[1].operations()
	if len(operations) != 1 || operations[0].Kind != effectOperationDamage || operations[0].TargetReference != referenceEventTarget || operations[0].Value == nil || operations[0].Value.Kind != valueConstant || operations[0].Value.Constant != 3 {
		t.Fatalf("triggered operations = %#v, want 3 damage to wield event target", operations)
	}
}

// TestCompileBulwarkSwordStaticDefinition 驗證 Bulwark Sword 的 class bonus 與 wield cost 由中央 static evaluator 消費。
// 輸入為 Bulwark Sword 卡牌定義；輸出為 stable static slot 與 +1 Power/+2 Reserve modifier，副作用為零。
func TestCompileBulwarkSwordStaticDefinition(t *testing.T) {
	definition := CardDefinition{
		id: bulwarkSwordCardID,
		face: CardFace{
			id: CardFaceID("face:8kmoi0a5uh:front"),
		},
	}
	abilities, err := compileAbilityDefinitions(
		definition,
		authoredAbilitiesForCard(definition.id),
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 1 || abilities[0].kind != abilityKindStatic || abilities[0].slot != "ability:8kmoi0a5uh:front:class-bonus-wield-cost" || abilities[0].static == nil || abilities[0].static.modifier.PowerDelta != 1 || abilities[0].static.modifier.ReserveCostDelta != 2 {
		t.Fatalf("compiled abilities = %#v, want Bulwark static modifier", abilities)
	}
}

// TestCompileViridianPermissionDefinition 驗證 Viridian 的 activation tax 只存在於穩定 permission Slot。
// 輸入為 Viridian 卡牌定義；輸出為 Water、對手、控制者回合與來源存續條件，副作用為零。
func TestCompileViridianPermissionDefinition(t *testing.T) {
	definition := CardDefinition{
		id: viridianProtectiveTrinketCardID,
		face: CardFace{
			id: CardFaceID("face:s3572j3oda:front"),
		},
	}
	abilities, err := compileAbilityDefinitions(
		definition,
		authoredAbilitiesForCard(definition.id),
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 1 || abilities[0].slot != "ability:s3572j3oda:front:opponent-water-activation-tax" || abilities[0].kind != abilityKindPermission || abilities[0].permission == nil || abilities[0].permission.candidateElement != "WATER" || abilities[0].permission.reserveCostDelta != 2 {
		t.Fatalf("compiled abilities = %#v, want Viridian permission definition", abilities)
	}
}

// TestCompileDuchessThornesDefinitions 驗證 Hindered、Cardistry trigger 與 activated discount 都有穩定 Slot。
// 輸入為 Duchess's Thornes 卡牌定義；輸出為三個已編譯能力及共用 operations，副作用為零。
func TestCompileDuchessThornesDefinitions(t *testing.T) {
	definition := CardDefinition{
		id: duchessThornesCardID,
		face: CardFace{
			id: CardFaceID("face:bEXmm4rKOs:front"),
		},
	}
	abilities, err := compileAbilityDefinitions(
		definition,
		authoredAbilitiesForCard(definition.id),
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 3 || abilities[0].slot != "ability:bEXmm4rKOs:front:hindered" || abilities[0].keyword != keywordHindered || abilities[1].slot != "ability:bEXmm4rKOs:front:ally-cardistry-bonus" || abilities[1].trigger == nil || abilities[1].trigger.event != eventKindCardistryActivated || abilities[2].slot != "ability:bEXmm4rKOs:front:rest-banish-cardistry-discount" {
		t.Fatalf("compiled abilities = %#v, want Duchess's Thornes keyword, trigger, and Object slots", abilities)
	}
	operations := abilities[2].operations()
	if len(operations) != 3 || operations[0].Kind != effectOperationRestSource || operations[1].Kind != effectOperationBanishSource || operations[2].Kind != effectOperationCardistryDiscount || operations[2].Amount != 6 {
		t.Fatalf("activated operations = %#v, want rest, banish, and Cardistry discount", operations)
	}
}

// TestCompileGrandCrusadersRingObjectDefinition 驗證 Grand Crusader's Ring 以穩定 Object slot 宣告放逐來源後抽牌。
// 輸入為 Grand Crusader's Ring 的卡牌定義與 Go 編寫資料；輸出為一個可執行的 Object ability，副作用為零。
func TestCompileGrandCrusadersRingObjectDefinition(t *testing.T) {
	definition := CardDefinition{
		id: grandCrusadersRingCardID,
		face: CardFace{
			id: CardFaceID("face:2gv7DC0KID:front"),
		},
	}
	abilities, err := compileAbilityDefinitions(
		definition,
		authoredAbilitiesForCard(definition.id),
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 1 || abilities[0].kind != abilityKindObject || abilities[0].slot != "ability:2gv7DC0KID:front:banish-draw" {
		t.Fatalf("compiled abilities = %#v, want one stable Object slot", abilities)
	}
	operations := abilities[0].operations()
	if len(operations) != 2 || operations[0].Kind != effectOperationBanishSource || operations[1].Kind != effectOperationDraw || operations[1].Amount != 1 {
		t.Fatalf("operations = %#v, want banish source followed by draw one", operations)
	}
}

// TestCompileWaterResonanceBaubleObjectDefinition 驗證 Water Resonance Bauble 將元素條件、放逐來源與抽牌編譯為 Object slot。
// 輸入為 Water Resonance Bauble 的卡牌定義與 Go 編寫資料；輸出為帶 Water 對手 Champion 條件的能力，副作用為零。
func TestCompileWaterResonanceBaubleObjectDefinition(t *testing.T) {
	definition := CardDefinition{
		id: waterResonanceBaubleCardID,
		face: CardFace{
			id: CardFaceID("face:dSSRtNnPtw:front"),
		},
	}
	abilities, err := compileAbilityDefinitions(
		definition,
		authoredAbilitiesForCard(definition.id),
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 1 || abilities[0].kind != abilityKindObject || abilities[0].objectCondition == nil || abilities[0].objectCondition.opponentChampionElement != "WATER" {
		t.Fatalf("compiled abilities = %#v, want Water-conditioned Object slot", abilities)
	}
}

// TestCompileWindResonanceBaubleObjectDefinition 驗證 Wind Resonance Bauble 將 Wind 條件編譯為獨立的 Object slot。
// 輸入為 Wind Resonance Bauble 的卡牌定義與 Go 編寫資料；輸出為帶 Wind 對手 Champion 條件的能力，副作用為零。
func TestCompileWindResonanceBaubleObjectDefinition(t *testing.T) {
	definition := CardDefinition{
		id: windResonanceBaubleCardID,
		face: CardFace{
			id: CardFaceID("face:bHGUNMFLg9:front"),
		},
	}
	abilities, err := compileAbilityDefinitions(
		definition,
		authoredAbilitiesForCard(definition.id),
	)
	if err != nil {
		t.Fatalf("compileAbilityDefinitions() error = %v", err)
	}
	if len(abilities) != 1 || abilities[0].objectCondition == nil || abilities[0].objectCondition.opponentChampionElement != "WIND" {
		t.Fatalf("compiled abilities = %#v, want Wind-conditioned Object slot", abilities)
	}
}

// TestCompileSmokeBombsAndSafeguardAmuletObjectDefinitions 驗證具有選擇與 replacement 的 Object slot 都由 Definition 編譯為可執行操作。
// 輸入為 Smoke Bombs、Safeguard Amulet 的卡牌定義與 Go 編寫資料；輸出為穩定 slot 和對應 operation 序列，副作用為零。
func TestCompileSmokeBombsAndSafeguardAmuletObjectDefinitions(t *testing.T) {
	testCases := []struct {
		name       string
		definition CardDefinition
		wantSlot   AbilitySlotID
		wantKinds  []effectOperationKind
	}{
		{
			name: "Smoke Bombs",
			definition: CardDefinition{
				id: smokeBombsCardID,
				face: CardFace{
					id: CardFaceID("face:ScGcOmkoQt:front"),
				},
			},
			wantSlot: "ability:ScGcOmkoQt:front:banish-stealth-draw",
			wantKinds: []effectOperationKind{
				effectOperationBanishSource,
				effectOperationChoose,
				effectOperationContinuousModifier,
				effectOperationDraw,
			},
		},
		{
			name: "Safeguard Amulet",
			definition: CardDefinition{
				id: safeguardAmuletCardID,
				face: CardFace{
					id: CardFaceID("face:yj2rJBREH8:front"),
				},
			},
			wantSlot: "ability:yj2rJBREH8:front:banish-prevent-non-combat-damage",
			wantKinds: []effectOperationKind{
				effectOperationBanishSource,
				effectOperationPreventNonCombatDamage,
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				abilities, err := compileAbilityDefinitions(
					testCase.definition,
					authoredAbilitiesForCard(testCase.definition.id),
				)
				if err != nil {
					t.Fatalf("compileAbilityDefinitions() error = %v", err)
				}
				if len(abilities) != 1 || abilities[0].slot != testCase.wantSlot {
					t.Fatalf("compiled abilities = %#v, want slot %q", abilities, testCase.wantSlot)
				}
				operations := abilities[0].operations()
				if len(operations) != len(testCase.wantKinds) {
					t.Fatalf("operations = %#v, want %d operations", operations, len(testCase.wantKinds))
				}
				for index, wantKind := range testCase.wantKinds {
					if operations[index].Kind != wantKind {
						t.Fatalf("operations[%d].Kind = %q, want %q", index, operations[index].Kind, wantKind)
					}
				}
			},
		)
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
