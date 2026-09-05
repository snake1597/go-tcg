package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	carddata "go-tcg/internal/card_data"
	"go-tcg/internal/constants"
	"go-tcg/internal/entity"
)

type CardID string

type CardFaceID string

type AbilitySlotID string

type CardDefinition struct {
	id          CardID
	dataVersion string
	face        CardFace
	card        entity.Card
}

type CardFace struct {
	id   CardFaceID
	card entity.Card
}

func (definition CardDefinition) ID() CardID {
	return definition.id
}

func (definition CardDefinition) DataVersion() string {
	return definition.dataVersion
}

func (definition CardDefinition) Name() string {
	return definition.card.Name
}

func (definition CardDefinition) Face() CardFace {
	return definition.face
}

func (face CardFace) ID() CardFaceID {
	return face.id
}

func (definition CardDefinition) faceData() entity.Card {
	return definition.card
}

func (face CardFace) HasType(want string) bool {
	return slices.Contains(face.card.Types, want)
}

func (face CardFace) Level() int64 {
	level := face.card.Level
	if level == nil {
		return -1
	}
	return *level
}

type DeckEntry struct {
	CardID CardID
	FaceID CardFaceID
	Count  int
}

type DeckSection []DeckEntry

func (section DeckSection) Count() int {
	total := 0
	for _, entry := range section {
		total += entry.Count
	}
	return total
}

type DeckManifest struct {
	Version         string
	CardDataVersion string
	MainDeck        DeckSection
	MaterialDeck    DeckSection
	OutsideGamePool DeckSection
}

func fixedStandardDeck() DeckManifest {
	return DeckManifest{
		Version:         constants.FixedDeckVersion,
		CardDataVersion: constants.FixedCardDataVersion,
		MainDeck: DeckSection{
			deckEntry("i9hf5lhl5f", 3),
			deckEntry("8bolq2y5qp", 4),
			deckEntry("wbjc9t8ycp", 3),
			deckEntry("o09csnorqv", 3),
			deckEntry("w7g91ru45w", 2),
			deckEntry("e8ygl32jef", 4),
			deckEntry("0mf1ug6yfi", 1),
			deckEntry("GjM8b5fxqj", 4),
			deckEntry("iohZMWh5v5", 3),
			deckEntry("qzv380ujf5", 3),
			deckEntry("gt2zqtgs42", 3),
			deckEntry("xgax8bbjqj", 4),
			deckEntry("td460e8ig0", 1),
			deckEntry("lcy0lw1veb", 2),
			deckEntry("5du8f077ua", 3),
			deckEntry("h68dr63eo5", 3),
			deckEntry("28bjn8g50v", 4),
			deckEntry("1db8hz4prm", 4),
			deckEntry("rufki4o41y", 4),
			deckEntry("4qc47amgpp", 2),
		},
		MaterialDeck: DeckSection{
			deckEntry("LMyKyVC2O9", 1),
			deckEntry("zb14m4c8lj", 1),
			deckEntry("8kmoi0a5uh", 1),
			deckEntry("2gv7DC0KID", 1),
			deckEntry("yj2rJBREH8", 1),
			deckEntry("ScGcOmkoQt", 1),
			deckEntry("s3572j3oda", 1),
			deckEntry("dSSRtNnPtw", 1),
			deckEntry("bHGUNMFLg9", 1),
			deckEntry("chsbalegbs", 1),
			deckEntry("vgWgu1DUYv", 1),
			deckEntry("bEXmm4rKOs", 1),
		},
		OutsideGamePool: DeckSection{},
	}
}

func deckEntry(cardID string, count int) DeckEntry {
	id := CardID(cardID)
	return DeckEntry{
		CardID: id,
		FaceID: CardFaceID("face:" + cardID + ":front"),
		Count:  count,
	}
}

func loadCardDefinitions(cardDirectory, manifestPath string) (map[CardID]CardDefinition, error) {
	manifest, err := carddata.ReadManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	if manifest.DataVersion != constants.FixedCardDataVersion {
		return nil, fmt.Errorf("card data version %q does not match fixed version %q", manifest.DataVersion, constants.FixedCardDataVersion)
	}
	if err := carddata.VerifyManifest(cardDirectory, manifest); err != nil {
		return nil, err
	}

	definitions := make(map[CardID]CardDefinition, len(manifest.Cards))
	for _, entry := range manifest.Cards {
		path := filepath.Join(cardDirectory, entry.Path)
		card, err := readCard(path)
		if err != nil {
			return nil, err
		}
		if card.UUID != entry.CardID {
			return nil, fmt.Errorf("%s card ID %q does not match manifest %q", entry.Path, card.UUID, entry.CardID)
		}
		id := CardID(card.UUID)
		faceID := CardFaceID("face:" + card.UUID + ":front")
		definition := CardDefinition{
			id:          id,
			dataVersion: manifest.DataVersion,
			face: CardFace{
				id:   faceID,
				card: card,
			},
			card: card,
		}
		definitions[id] = definition
	}
	return definitions, nil
}

func readCard(path string) (entity.Card, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return entity.Card{}, fmt.Errorf("read card %s: %w", path, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	var card entity.Card
	if err := decoder.Decode(&card); err != nil {
		return entity.Card{}, fmt.Errorf("decode card %s: %w", path, err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return entity.Card{}, fmt.Errorf("card %s contains more than one JSON value", path)
		}
		return entity.Card{}, fmt.Errorf("decode card %s trailing data: %w", path, err)
	}
	return card, nil
}

func validateFixedStandardDeck(deck DeckManifest, definitions map[CardID]CardDefinition) error {
	if deck.Version != constants.FixedDeckVersion {
		return fmt.Errorf("deck version %q does not match %q", deck.Version, constants.FixedDeckVersion)
	}
	if deck.CardDataVersion != constants.FixedCardDataVersion {
		return fmt.Errorf("deck card data version %q does not match %q", deck.CardDataVersion, constants.FixedCardDataVersion)
	}
	if deck.MainDeck.Count() != 60 {
		return fmt.Errorf("main deck has %d cards, want 60", deck.MainDeck.Count())
	}
	if deck.MaterialDeck.Count() > 12 {
		return fmt.Errorf("material deck has %d cards, maximum is 12", deck.MaterialDeck.Count())
	}
	if deck.OutsideGamePool == nil {
		return errors.New("outside game pool is required")
	}
	if err := validateDeckSection("main deck", deck.MainDeck, 4, definitions); err != nil {
		return err
	}
	if err := validateDeckSection("material deck", deck.MaterialDeck, 1, definitions); err != nil {
		return err
	}
	if err := validateDeckSection("outside game pool", deck.OutsideGamePool, 0, definitions); err != nil {
		return err
	}
	canonicalDeck := fixedStandardDeck()
	if !slices.Equal(deck.MainDeck, canonicalDeck.MainDeck) ||
		!slices.Equal(deck.MaterialDeck, canonicalDeck.MaterialDeck) ||
		!slices.Equal(deck.OutsideGamePool, canonicalDeck.OutsideGamePool) {
		return errors.New("deck does not match the fixed manifest")
	}

	startingChampions := 0
	for _, entry := range deck.MaterialDeck {
		definition := definitions[entry.CardID]
		if definition.Face().HasType("CHAMPION") && definition.Face().Level() == 0 {
			startingChampions += entry.Count
		}
	}
	if startingChampions == 0 {
		return errors.New("material deck has no Level 0 Champion")
	}
	return nil
}

func validateMirroredDecks(first, second DeckManifest) error {
	if first.Version != second.Version ||
		first.CardDataVersion != second.CardDataVersion ||
		!slices.Equal(first.MainDeck, second.MainDeck) ||
		!slices.Equal(first.MaterialDeck, second.MaterialDeck) ||
		!slices.Equal(first.OutsideGamePool, second.OutsideGamePool) {
		return errors.New("both players must use identical fixed deck manifests")
	}
	return nil
}

func validateDeckSection(name string, section DeckSection, maximumCopies int, definitions map[CardID]CardDefinition) error {
	seen := make(map[CardID]struct{}, len(section))
	for _, entry := range section {
		if entry.Count <= 0 {
			return fmt.Errorf("%s card %q has invalid count %d", name, entry.CardID, entry.Count)
		}
		if maximumCopies > 0 && entry.Count > maximumCopies {
			return fmt.Errorf("%s card %q has %d copies, maximum is %d", name, entry.CardID, entry.Count, maximumCopies)
		}
		if _, exists := seen[entry.CardID]; exists {
			return fmt.Errorf("%s repeats card %q", name, entry.CardID)
		}
		seen[entry.CardID] = struct{}{}
		definition, exists := definitions[entry.CardID]
		if !exists {
			return fmt.Errorf("%s contains unknown card %q", name, entry.CardID)
		}
		if definition.DataVersion() != constants.FixedCardDataVersion {
			return fmt.Errorf("%s card %q has data version %q", name, entry.CardID, definition.DataVersion())
		}
		if entry.FaceID != definition.Face().ID() {
			return fmt.Errorf("%s card %q face %q does not match %q", name, entry.CardID, entry.FaceID, definition.Face().ID())
		}
	}
	return nil
}
