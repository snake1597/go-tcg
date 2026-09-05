package game

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFixedStandardDeck(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	content, err := loadCardDefinitions(
		filepath.Join(repositoryRoot, "card"),
		filepath.Join(repositoryRoot, "card-data-manifest.json"),
	)
	if err != nil {
		t.Fatalf("loadCardDefinitions() error = %v", err)
	}

	deck := fixedStandardDeck()
	if err := validateFixedStandardDeck(deck, content); err != nil {
		t.Fatalf("validateFixedStandardDeck() error = %v", err)
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
	if got := len(content); got != 32 {
		t.Fatalf("card definition count = %d, want 32", got)
	}

	spirit := content[CardID("LMyKyVC2O9")]
	if spirit.ID() != CardID("LMyKyVC2O9") {
		t.Fatalf("Spirit card ID = %q, want LMyKyVC2O9", spirit.ID())
	}
	if spirit.Face().ID() != CardFaceID("face:LMyKyVC2O9:front") {
		t.Fatalf("Spirit face ID = %q", spirit.Face().ID())
	}
	if spirit.DataVersion() != "card-data-v3" {
		t.Fatalf("Spirit data version = %q, want card-data-v3", spirit.DataVersion())
	}
	if spirit.Name() != "Spirit of Fire" {
		t.Fatalf("Spirit name = %q", spirit.Name())
	}
	if !spirit.Face().HasType("CHAMPION") || spirit.Face().Level() != 0 {
		t.Fatalf("Spirit face does not preserve Champion level: %+v", spirit.Face())
	}
}

func TestFixedStandardDeckRejectsInvalidManifest(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	content, err := loadCardDefinitions(
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
			name: "missing outside game pool",
			mutate: func(deck *DeckManifest) {
				deck.OutsideGamePool = nil
			},
			message: "outside game pool is required",
		},
		{
			name: "wrong version",
			mutate: func(deck *DeckManifest) {
				deck.Version = "standard-fire-v1"
			},
			message: "deck version",
		},
		{
			name: "wrong count",
			mutate: func(deck *DeckManifest) {
				deck.MainDeck[0].Count--
			},
			message: "main deck has 59 cards",
		},
		{
			name: "unknown card",
			mutate: func(deck *DeckManifest) {
				deck.MainDeck[0].CardID = CardID("missing")
			},
			message: "unknown card",
		},
		{
			name: "too many copies",
			mutate: func(deck *DeckManifest) {
				deck.MainDeck[0].Count--
				deck.MainDeck[1].Count++
			},
			message: "maximum is 4",
		},
		{
			name: "substituted known main deck card",
			mutate: func(deck *DeckManifest) {
				deck.MainDeck[0].CardID = CardID("LMyKyVC2O9")
				deck.MainDeck[0].FaceID = CardFaceID("face:LMyKyVC2O9:front")
			},
			message: "does not match the fixed manifest",
		},
		{
			name: "removed material deck card",
			mutate: func(deck *DeckManifest) {
				deck.MaterialDeck = deck.MaterialDeck[:11]
			},
			message: "does not match the fixed manifest",
		},
		{
			name: "added outside game pool card",
			mutate: func(deck *DeckManifest) {
				deck.OutsideGamePool = DeckSection{
					deckEntry(
						"GjM8b5fxqj",
						1,
					),
				}
			},
			message: "does not match the fixed manifest",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			deck := fixedStandardDeck()
			test.mutate(&deck)
			err := validateFixedStandardDeck(deck, content)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("validateFixedStandardDeck() error = %v, want %q", err, test.message)
			}
		})
	}
}

func TestFixedStandardDeckRequiresMirroredPlayers(t *testing.T) {
	first := fixedStandardDeck()
	second := fixedStandardDeck()
	if err := validateMirroredDecks(first, second); err != nil {
		t.Fatalf("validateMirroredDecks() error = %v", err)
	}

	second.MainDeck[0].Count--
	second.MainDeck[1].Count++
	err := validateMirroredDecks(first, second)
	if err == nil || !strings.Contains(err.Error(), "identical") {
		t.Fatalf("validateMirroredDecks() error = %v, want identical-deck error", err)
	}
}
