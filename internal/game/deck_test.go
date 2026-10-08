package game

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go-tcg/internal/constants"
)

// TestFixedStandardDeckReferencesLoadedDefinitions 驗證固定牌組可由 repository 的不可變卡面資料完整建立。
// 輸入為固定牌組與載入的卡面資料；輸出為固定張數、版本與引用一致性的 assertion，副作用為讀取卡面檔案。
func TestFixedStandardDeckReferencesLoadedDefinitions(t *testing.T) {
	repositoryRoot := filepath.Join(
		"..",
		"..",
	)
	definitions, err := loadCardDefinitions(
		filepath.Join(repositoryRoot, "card"),
		filepath.Join(repositoryRoot, "card-data-manifest.json"),
	)
	if err != nil {
		t.Fatalf("loadCardDefinitions() error = %v", err)
	}

	deck := fixedStandardDeck()
	if err := validateDeckReferences(deck, definitions); err != nil {
		t.Fatalf("validateDeckReferences() error = %v", err)
	}
	if deck.Version != constants.FixedDeckVersion || deck.CardDataVersion != constants.FixedCardDataVersion {
		t.Fatalf("fixed deck version = %q, card data version = %q", deck.Version, deck.CardDataVersion)
	}
	if got := deck.MainDeck.Count(); got != 60 {
		t.Fatalf("main deck count = %d, want 60", got)
	}
	if got := deck.MaterialDeck.Count(); got != 12 {
		t.Fatalf("material deck count = %d, want 12", got)
	}
	if got := deck.OutsideGamePool.Count(); got != 0 {
		t.Fatalf("outside game pool count = %d, want 0", got)
	}
	if got := len(definitions); got != 32 {
		t.Fatalf("card definition count = %d, want 32", got)
	}
	for _, section := range []struct {
		name    string
		cards   DeckSection
		maximum int
	}{
		{
			name:    "main deck",
			cards:   deck.MainDeck,
			maximum: 4,
		},
		{
			name:    "material deck",
			cards:   deck.MaterialDeck,
			maximum: 1,
		},
		{
			name:  "outside game pool",
			cards: deck.OutsideGamePool,
		},
	} {
		seen := make(map[CardID]struct{}, len(section.cards))
		for _, entry := range section.cards {
			if entry.Count <= 0 {
				t.Fatalf("%s card %q has count %d, want positive", section.name, entry.CardID, entry.Count)
			}
			if section.maximum > 0 && entry.Count > section.maximum {
				t.Fatalf("%s card %q has count %d, want at most %d", section.name, entry.CardID, entry.Count, section.maximum)
			}
			if _, exists := seen[entry.CardID]; exists {
				t.Fatalf("%s repeats card %q", section.name, entry.CardID)
			}
			seen[entry.CardID] = struct{}{}
		}
	}
	if !slices.Equal(deck.OutsideGamePool, DeckSection{}) {
		t.Fatalf("outside game pool = %#v, want empty", deck.OutsideGamePool)
	}

	divineRelics := 0
	for _, entry := range deck.MaterialDeck {
		definition := definitions[entry.CardID]
		if definition.card.EffectRaw != nil && strings.Contains(*definition.card.EffectRaw, "Divine Relic") {
			divineRelics += entry.Count
		}
	}
	if divineRelics > 1 {
		t.Fatalf("material deck has %d Divine Relic cards, want at most 1", divineRelics)
	}
}

// TestDeriveSupportSetBuildsMetadataFromCompiledDefinitions 驗證 production Support Set 直接取自已編譯定義，並保留卡面、Slot、operation、相依內容與卡面裁定來源。
// 輸入為 repository 的固定卡面資料及牌組；輸出為可驗證的衍生支援集合，副作用為讀取卡面檔案。
func TestDeriveSupportSetBuildsMetadataFromCompiledDefinitions(t *testing.T) {
	repositoryRoot := filepath.Join(
		"..",
		"..",
	)
	definitions, err := loadCardDefinitions(
		filepath.Join(
			repositoryRoot,
			"card",
		),
		filepath.Join(
			repositoryRoot,
			"card-data-manifest.json",
		),
	)
	if err != nil {
		t.Fatalf("loadCardDefinitions() error = %v", err)
	}

	support, err := deriveSupportSet(
		fixedStandardDeck(),
		definitions,
	)
	if err != nil {
		t.Fatalf("deriveSupportSet() error = %v", err)
	}
	duchess := support.abilities[AbilitySlotID("ability:qzv380ujf5:front:cardistry-copy-action")]
	if duchess.Face != CardFaceID("face:qzv380ujf5:front") || duchess.Mechanism != abilityKindCardistry {
		t.Fatalf("Duchess support = %#v, want compiled Cardistry metadata", duchess)
	}
	if !containsOperation(duchess.Operations, effectOperationCopyAction) {
		t.Fatalf("Duchess operations = %#v, want copy_action", duchess.Operations)
	}
	if !containsSupportDependency(duchess.Dependencies, blazingThrowCardID) || !containsSupportDependency(duchess.Dependencies, fieryInterferenceCardID) || !containsSupportDependency(duchess.Dependencies, straightFlareCardID) {
		t.Fatalf("Duchess dependencies = %#v, want every eligible copied Action", duchess.Dependencies)
	}
	if len(duchess.Rulings) == 0 {
		t.Fatal("Duchess rulings = none, want card-data ruling reference")
	}
}

// TestFixedSupportSetCompilesEveryRulesBearingCard 驗證固定牌組所有規則文字及最後遷移的 Slots 都有 production 定義。
// 輸入為 repository 固定卡面與牌組；輸出為完整性斷言，副作用僅讀取測試資料。
func TestFixedSupportSetCompilesEveryRulesBearingCard(t *testing.T) {
	repositoryRoot := filepath.Join(
		"..",
		"..",
	)
	definitions, err := loadCardDefinitions(
		filepath.Join(repositoryRoot, "card"),
		filepath.Join(repositoryRoot, "card-data-manifest.json"),
	)
	if err != nil {
		t.Fatalf("loadCardDefinitions() error = %v", err)
	}
	support, err := deriveSupportSet(
		fixedStandardDeck(),
		definitions,
	)
	if err != nil {
		t.Fatalf("deriveSupportSet() error = %v", err)
	}
	for _, cardID := range reachableDeckCards(fixedStandardDeck()) {
		definition := definitions[cardID]
		if cardHasRulesText(definition.card) && len(definition.abilities) == 0 {
			t.Fatalf("rules-bearing Card Definition %q has no compiled abilities", cardID)
		}
	}
	for _, slot := range []AbilitySlotID{
		"ability:GjM8b5fxqj:front:on-enter-rest-immortality",
		"ability:wbjc9t8ycp:front:on-enter-suited-counters",
		"ability:lcy0lw1veb:front:on-enter-sacrifice-power",
		"ability:td460e8ig0:front:on-attack-self-damage",
		"ability:h68dr63eo5:front:on-enter-suited-threshold-damage",
		"ability:LMyKyVC2O9:front:on-enter-draw-seven",
		"ability:zb14m4c8lj:front:on-enter-taunt",
	} {
		if _, exists := support.abilities[slot]; !exists {
			t.Fatalf("Support Set missing migrated slot %q", slot)
		}
	}
}

// TestSupportSetRejectsNonProductionStatusAndDependencyCycles 驗證 production gate 拒絕未裁定內容與循環相依。
// 輸入為刻意破壞的 compiled metadata；輸出為明確拒絕錯誤，副作用為零。
func TestSupportSetRejectsNonProductionStatusAndDependencyCycles(t *testing.T) {
	repositoryRoot := filepath.Join(
		"..",
		"..",
	)
	definitions, err := loadCardDefinitions(
		filepath.Join(repositoryRoot, "card"),
		filepath.Join(repositoryRoot, "card-data-manifest.json"),
	)
	if err != nil {
		t.Fatalf("loadCardDefinitions() error = %v", err)
	}
	definition := definitions[wonderlandsReignCardID]
	definition.abilities[0].supportStatus = supportStatusNeedsRuling
	definitions[wonderlandsReignCardID] = definition
	if _, err := deriveSupportSet(fixedStandardDeck(), definitions); err == nil || !strings.Contains(err.Error(), "non-production support status") {
		t.Fatalf("deriveSupportSet() error = %v, want non-production status rejection", err)
	}
	definition.abilities[0].supportStatus = supportStatusSupported
	definition.abilities[0].effects = append(
		definition.abilities[0].effects,
		compiledEffect{
			kind: authoredEffectKind("unknown"),
		},
	)
	definitions[wonderlandsReignCardID] = definition
	if _, err := deriveSupportSet(fixedStandardDeck(), definitions); err == nil || !strings.Contains(err.Error(), "unknown or unmapped primitive") {
		t.Fatalf("deriveSupportSet() error = %v, want unknown primitive rejection", err)
	}

	firstCard := CardID("first")
	secondCard := CardID("second")
	firstFace := CardFaceID("face:first:front")
	secondFace := CardFaceID("face:second:front")
	cyclic := compiledSupportSet{
		faces: map[CardFaceID]CardID{
			firstFace:  firstCard,
			secondFace: secondCard,
		},
		abilities: map[AbilitySlotID]compiledSupportAbility{
			"ability:first:front:dependency": {
				Face:         firstFace,
				Dependencies: []CardID{secondCard},
			},
			"ability:second:front:dependency": {
				Face:         secondFace,
				Dependencies: []CardID{firstCard},
			},
		},
	}
	if err := validateSupportDependencyCycles(cyclic); err == nil || !strings.Contains(err.Error(), "cyclic content dependency") {
		t.Fatalf("validateSupportDependencyCycles() error = %v, want cycle rejection", err)
	}
}

// TestValidateDeckReferencesRejectsUnbuildableCards 驗證建局前會拒絕無法轉成卡牌實例的固定牌組引用。
// 輸入為被修改的固定牌組與正確 CardDefinition 索引；輸出為未知卡、錯誤牌面或錯誤起始 Champion 的錯誤，副作用為零。
func TestValidateDeckReferencesRejectsUnbuildableCards(t *testing.T) {
	repositoryRoot := filepath.Join(
		"..",
		"..",
	)
	definitions, err := loadCardDefinitions(
		filepath.Join(repositoryRoot, "card"),
		filepath.Join(repositoryRoot, "card-data-manifest.json"),
	)
	if err != nil {
		t.Fatalf("loadCardDefinitions() error = %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*DeckManifest)
		message string
	}{
		{
			name: "unknown card",
			mutate: func(deck *DeckManifest) {
				deck.MainDeck[0].CardID = CardID("missing")
			},
			message: "unknown card",
		},
		{
			name: "wrong face",
			mutate: func(deck *DeckManifest) {
				deck.MainDeck[0].FaceID = CardFaceID("face:i9hf5lhl5f:back")
			},
			message: "does not match",
		},
		{
			name: "no starting champion",
			mutate: func(deck *DeckManifest) {
				deck.MaterialDeck = deck.MaterialDeck[1:]
			},
			message: "0 Level 0 Champions",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			deck := fixedStandardDeck()
			test.mutate(&deck)
			err := validateDeckReferences(deck, definitions)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("validateDeckReferences() error = %v, want %q", err, test.message)
			}
		})
	}
}
