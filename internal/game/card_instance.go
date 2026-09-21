package game

import (
	"fmt"
	"strconv"

	"go-tcg/internal/model"
)

// cardInstanceID 識別單局遊戲中的一張實體卡牌。
type cardInstanceID string

type cardInstance struct {
	ID          cardInstanceID `json:"id"`
	Owner       *model.Player  `json:"owner"`
	Definition  CardID         `json:"definition"`
	Face        CardFaceID     `json:"face"`
	Level       int64          `json:"level"`
	Types       []string       `json:"types"`
	Subtypes    []string       `json:"subtypes"`
	Elements    []string       `json:"elements"`
	Classes     []string       `json:"classes"`
	MemoryCost  int            `json:"memory_cost"`
	ReserveCost int            `json:"reserve_cost"`
	Fast        bool           `json:"fast"`
	Power       int            `json:"power"`
	Life        int            `json:"life"`
}

func (g *Game) newCardInstance(player *model.Player, entry DeckEntry, definitions map[CardID]CardDefinition) cardInstanceID {
	definition := definitions[entry.CardID]
	face := definition.Face()
	cardData := definition.faceData()
	cardCount := len(g.state.Cards) + 1
	identifierText := fmt.Sprintf(
		"card:%s:%d",
		player,
		cardCount,
	)
	identifier := cardInstanceID(identifierText)
	g.state.Cards[identifier] = cardInstance{
		ID:          identifier,
		Owner:       player,
		Definition:  entry.CardID,
		Face:        entry.FaceID,
		Level:       face.Level(),
		Types:       append([]string(nil), cardData.Types...),
		Subtypes:    append([]string(nil), cardData.Subtypes...),
		Elements:    append([]string(nil), cardData.Elements...),
		Classes:     append([]string(nil), cardData.Classes...),
		MemoryCost:  memoryCost(cardData),
		ReserveCost: reserveCost(cardData),
		Fast:        cardData.Speed != nil && *cardData.Speed,
		Power:       cardStat(cardData.Power),
		Life:        cardStat(cardData.Life),
	}
	g.state.Entities[entityID(identifier)] = knowledgeEntity{
		Name: definitions[entry.CardID].Name(),
	}
	return identifier
}

func cardStat(stat *int64) int {
	if stat == nil {
		return 0
	}
	return int(*stat)
}

func memoryCost(card Card) int {
	if card.Cost == nil || card.Cost.Type != "memory" {
		return 0
	}
	cost, err := strconv.Atoi(card.Cost.Value)
	if err != nil {
		return 0
	}
	return cost
}

func reserveCost(card Card) int {
	if card.Cost == nil || card.Cost.Type != "reserve" {
		return 0
	}
	cost, err := strconv.Atoi(card.Cost.Value)
	if err != nil {
		return 0
	}
	return cost
}
