package game

import "fmt"

type cardZone string

const (
	cardZoneMainDeck        cardZone = "main_deck"
	cardZoneHand            cardZone = "hand"
	cardZoneMaterialDeck    cardZone = "material_deck"
	cardZoneGraveyard       cardZone = "graveyard"
	cardZoneMemory          cardZone = "memory"
	cardZoneBanishment      cardZone = "banishment"
	cardZoneOutsideGamePool cardZone = "outside_game_pool"
)

// cardsInZone 讀取玩家指定區域的卡牌順序；未知區域回傳 nil，無副作用。
func cardsInZone(zones playerZones, zone cardZone) []cardInstanceID {
	switch zone {
	case cardZoneMainDeck:
		return zones.MainDeck
	case cardZoneHand:
		return zones.Hand
	case cardZoneMaterialDeck:
		return zones.MaterialDeck
	case cardZoneGraveyard:
		return zones.Graveyard
	case cardZoneMemory:
		return zones.Memory
	case cardZoneBanishment:
		return zones.Banishment
	case cardZoneOutsideGamePool:
		return zones.OutsideGamePool
	default:
		return nil
	}
}

// setCardsInZone 將卡牌順序寫回指定玩家區域；未知區域回傳錯誤，僅修改傳入的區域副本。
func setCardsInZone(zones playerZones, zone cardZone, cards []cardInstanceID) (playerZones, error) {
	switch zone {
	case cardZoneMainDeck:
		zones.MainDeck = cards
	case cardZoneHand:
		zones.Hand = cards
	case cardZoneMaterialDeck:
		zones.MaterialDeck = cards
	case cardZoneGraveyard:
		zones.Graveyard = cards
	case cardZoneMemory:
		zones.Memory = cards
	case cardZoneBanishment:
		zones.Banishment = cards
	case cardZoneOutsideGamePool:
		zones.OutsideGamePool = cards
	default:
		return zones, fmt.Errorf("unknown card zone %q", zone)
	}
	return zones, nil
}

// moveCardBetweenZones 驗證卡牌仍在來源區域，並在玩家區域副本中移至目的區域；不提交遊戲狀態。
func moveCardBetweenZones(zones playerZones, from, to cardZone, card cardInstanceID) (playerZones, error) {
	source := append([]cardInstanceID(nil), cardsInZone(zones, from)...)
	index := cardIndex(source, card)
	if index < 0 {
		return zones, fmt.Errorf("card %q is not in %q", card, from)
	}
	updated, err := setCardsInZone(zones, from, removeCardAt(source, index))
	if err != nil {
		return zones, fmt.Errorf("remove card from zone: %w", err)
	}
	destination := append([]cardInstanceID(nil), cardsInZone(updated, to)...)
	updated, err = setCardsInZone(updated, to, append(destination, card))
	if err != nil {
		return zones, fmt.Errorf("add card to zone: %w", err)
	}
	return updated, nil
}

// banishCardsFromGraveyard 驗證所有指定牌位於墓地且不重複，再於區域副本中一併放逐；失敗不修改原區域。
func banishCardsFromGraveyard(zones playerZones, cards []cardInstanceID) (playerZones, error) {
	seen := make(map[cardInstanceID]bool, len(cards))
	for _, card := range cards {
		if seen[card] || cardIndex(zones.Graveyard, card) < 0 {
			return zones, fmt.Errorf("card %q is not available in graveyard", card)
		}
		seen[card] = true
	}
	updated := zones
	for _, card := range cards {
		var err error
		updated, err = moveCardBetweenZones(updated, cardZoneGraveyard, cardZoneBanishment, card)
		if err != nil {
			return zones, fmt.Errorf("banish graveyard card: %w", err)
		}
	}
	return updated, nil
}
