package game

import (
	"fmt"

	"go-tcg/internal/model"
)

// playerCardMove 描述同一玩家持有的單張 Card Instance 在兩個玩家卡牌區域間的移動。
type playerCardMove struct {
	Card cardInstanceID
	From cardZone
	To   cardZone
}

// playerCardMoveRequest 描述一批必須全部成功才提交的玩家卡牌移動及其可觀察事件語意。
type playerCardMoveRequest struct {
	Player       *model.Player
	Moves        []playerCardMove
	Cause        string
	EventKind    string
	VisibleToAll bool
}

// commitObjectAbilityBanishment 驗證控制者、場上來源與卡牌擁有者後，原子移除 Object 並將其 Card Instance 放入擁有者的 Banishment。
// 輸入為啟動能力的控制者與場上來源；輸出為離場前 Object 或錯誤；成功時更新 Objects、Banishment 與公開事件，失敗時不改變遊戲狀態。
func (g *Game) commitObjectAbilityBanishment(player *model.Player, source objectID) (fieldObject, error) {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) {
		return fieldObject{}, fmt.Errorf("invalid object ability")
	}
	card, exists := g.state.Cards[object.Card]
	if !exists || !samePlayer(card.Owner, object.Owner) {
		return fieldObject{}, fmt.Errorf("object source card is not owned by object controller")
	}
	zones := g.state.Zones[object.Owner.UID]
	zones.Banishment = append(zones.Banishment, object.Card)
	delete(g.state.Objects, source)
	g.state.Zones[object.Owner.UID] = zones
	g.recordPublicEvent(player, "ability", "banish", object.Card)
	return object, nil
}

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

// commitPlayerCardMoves 重新驗證 owner 與來源區域後，原子提交一批玩家卡牌移動並寫入事件與追蹤權。
// 輸入為玩家、移動清單與可見事件設定；輸出為錯誤或 nil；失敗無副作用，成功時更新區域、Knowledge State 與 Game Event。
func (g *Game) commitPlayerCardMoves(request playerCardMoveRequest) error {
	if !g.hasPlayer(request.Player) {
		return fmt.Errorf("unknown player %q", request.Player)
	}
	if len(request.Moves) == 0 || request.Cause == "" || request.EventKind == "" {
		return fmt.Errorf("player card move requires moves, cause, and event kind")
	}
	updated := g.state.Zones[request.Player.UID]
	for _, move := range request.Moves {
		if !isPlayerCardZone(move.From) || !isPlayerCardZone(move.To) || move.From == move.To {
			return fmt.Errorf("invalid player card move from %q to %q", move.From, move.To)
		}
		card, exists := g.state.Cards[move.Card]
		if !exists || !samePlayer(card.Owner, request.Player) {
			return fmt.Errorf("card %q is not owned by player %q", move.Card, request.Player)
		}
		var err error
		updated, err = moveCardBetweenZones(
			updated,
			move.From,
			move.To,
			move.Card,
		)
		if err != nil {
			return fmt.Errorf("move player card %q: %w", move.Card, err)
		}
	}
	g.state.Zones[request.Player.UID] = updated
	eventCapacity := len(request.Moves)
	batch := eventBatch{
		Player: request.Player,
		Cause:  request.Cause,
		Events: make([]gameEvent, 0, eventCapacity),
	}
	for _, move := range request.Moves {
		cardEntity := entityID(move.Card)
		g.grantCardTracking(request.Player, cardEntity)
		g.state.NextEvent++
		batch.Events = append(
			batch.Events,
			gameEvent{
				Sequence: g.state.NextEvent,
				Kind:     request.EventKind,
				Card:     move.Card,
			},
		)
		if request.VisibleToAll {
			for _, viewer := range g.players {
				g.recordVisibleEvent(viewer, request.EventKind, cardEntity)
			}
			continue
		}
		g.recordVisibleEvent(request.Player, request.EventKind, cardEntity)
	}
	g.state.Events = append(g.state.Events, batch)
	return nil
}

// isPlayerCardZone 判斷區域是否屬於可由玩家 Card Instance 移動流程處理的卡牌區域。
// 輸入為區域種類；輸出為是否合法；無副作用。
func isPlayerCardZone(zone cardZone) bool {
	switch zone {
	case cardZoneHand,
		cardZoneGraveyard,
		cardZoneMemory,
		cardZoneBanishment:
		return true
	default:
		return false
	}
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
