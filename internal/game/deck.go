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
	"time"

	carddata "go-tcg/internal/card_data"
	"go-tcg/internal/constants"
)

type Card struct {
	Classes      []string     `json:"classes"`
	Cost         *Cost        `json:"cost"`
	CreatedAt    time.Time    `json:"created_at"`
	Durability   *int64       `json:"durability"`
	Effect       *string      `json:"effect"`
	EffectRaw    *string      `json:"effect_raw"`
	Elements     []string     `json:"elements"`
	Flavor       *string      `json:"flavor"`
	LastUpdate   time.Time    `json:"last_update"`
	Level        *int64       `json:"level"`
	Life         *int64       `json:"life"`
	Name         string       `json:"name"`
	Power        *int64       `json:"power"`
	ReferencedBy []*Reference `json:"referenced_by"`
	References   []*Reference `json:"references"`
	Rule         []*Rule      `json:"rule"`
	Slug         string       `json:"slug"`
	Speed        *bool        `json:"speed"`
	Subtypes     []string     `json:"subtypes"`
	Types        []string     `json:"types"`
	UUID         string       `json:"uuid"`
}

type Cost struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Rule struct {
	DateAdded   string `json:"date_added"`
	Description string `json:"description"`
	Title       string `json:"title"`
}

type Reference struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Direction string `json:"direction"`
}

type CardID string

type CardFaceID string

type AbilitySlotID string

type CardDefinition struct {
	id          CardID
	dataVersion string
	face        CardFace
	card        Card
}

type CardFace struct {
	id   CardFaceID
	card Card
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

func (definition CardDefinition) faceData() Card {
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

// loadCardDefinitions 先驗證 manifest 版本與卡牌檔案，再建立固定 front 牌面的定義索引。
// 檔案內 UUID 必須與 manifest 相符；任一資料錯誤都拒絕載入，不回傳部分定義。
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

// readCard 要求檔案恰有一個 JSON 值；第二個值或尾端非空白資料都回傳 error。
// 此函式只處理解碼，資料版本、檔案完整性與 UUID 對照由載入層驗證。
func readCard(path string) (Card, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Card{}, fmt.Errorf("read card %s: %w", path, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	var card Card
	if err := decoder.Decode(&card); err != nil {
		return Card{}, fmt.Errorf("decode card %s: %w", path, err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Card{}, fmt.Errorf("card %s contains more than one JSON value", path)
		}
		return Card{}, fmt.Errorf("decode card %s trailing data: %w", path, err)
	}
	return card, nil
}

// validateDeckReferences 確保固定牌組的卡片與牌面可由載入的卡面資料建立，且有唯一的起始 Champion。
// 輸入為固定 DeckManifest 與已驗證 CardDefinition 索引；輸出為引用或起始 Champion 錯誤，副作用為零。
func validateDeckReferences(deck DeckManifest, definitions map[CardID]CardDefinition) error {
	sections := []struct {
		name  string
		cards DeckSection
	}{
		{
			name:  "main deck",
			cards: deck.MainDeck,
		},
		{
			name:  "material deck",
			cards: deck.MaterialDeck,
		},
		{
			name:  "outside game pool",
			cards: deck.OutsideGamePool,
		},
	}
	startingChampions := 0
	for _, section := range sections {
		for _, entry := range section.cards {
			definition, exists := definitions[entry.CardID]
			if !exists {
				return fmt.Errorf("%s contains unknown card %q", section.name, entry.CardID)
			}
			if entry.FaceID != definition.Face().ID() {
				return fmt.Errorf("%s card %q face %q does not match %q", section.name, entry.CardID, entry.FaceID, definition.Face().ID())
			}
			if section.name == "material deck" && definition.Face().HasType("CHAMPION") && definition.Face().Level() == 0 {
				startingChampions += entry.Count
			}
		}
	}
	if startingChampions != 1 {
		return fmt.Errorf("material deck has %d Level 0 Champions, want exactly 1", startingChampions)
	}
	return nil
}
