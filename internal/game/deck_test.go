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
