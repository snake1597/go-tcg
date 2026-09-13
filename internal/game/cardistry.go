package game

import (
	"fmt"
	"sort"

	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
)

const (
	fiveOfSpadesCardID     CardID = "i9hf5lhl5f"
	fourOfSpadesCardID     CardID = "8bolq2y5qp"
	fourOfHeartsCardID     CardID = "xgax8bbjqj"
	threeOfHeartsCardID    CardID = "1db8hz4prm"
	threeOfSpadesCardID    CardID = "o09csnorqv"
	twoOfHeartsCardID      CardID = "rufki4o41y"
	twoOfSpadesCardID      CardID = "e8ygl32jef"
	wonderlandsReignCardID CardID = "0mf1ug6yfi"
	noireCardID            CardID = "wbjc9t8ycp"
	rougeCardID            CardID = "h68dr63eo5"
)

func (g *Game) legalCardistries(player *model.Player) []objectID {
	objects := make([]objectID, 0, len(g.state.Objects))
	for id := range g.state.Objects {
		if g.canActivateCardistry(player, id) {
			objects = append(objects, id)
		}
	}
	sort.Slice(objects, func(first, second int) bool { return objects[first] < objects[second] })
	return objects
}

func (g *Game) canActivateCardistry(player *model.Player, source objectID) bool {
	object, exists := g.state.Objects[source]
	if !exists || !samePlayer(object.Owner, player) || g.state.CardistryUsed[source] {
		return false
	}
	baseCost, fast := g.cardistryBaseCost(object.Card)
	if baseCost < 0 || !samePlayer(g.state.Scheduler.OpportunityHolder, player) {
		return false
	}
	if !fast && (!samePlayer(g.state.Scheduler.TurnPlayer, player) || g.state.Scheduler.Phase != PhaseMain || len(g.state.EffectsStack) != 0) {
		return false
	}
	return len(g.state.Zones[player.UID].Memory) >= g.cardistryCost(player, baseCost)
}

func (g *Game) cardistryBaseCost(card cardInstanceID) (int, bool) {
	switch g.state.Cards[card].Definition {
	case fiveOfSpadesCardID:
		return 5, false
	case fourOfSpadesCardID:
		return 4, false
	case fourOfHeartsCardID:
		return 4, false
	case threeOfHeartsCardID:
		return 3, false
	case threeOfSpadesCardID:
		return 3, false
	case twoOfHeartsCardID:
		return 2, false
	case twoOfSpadesCardID:
		return 2, true
	case wonderlandsReignCardID:
		return 10, false
	default:
		return -1, false
	}
}

func (g *Game) cardistryCost(player *model.Player, baseCost int) int {
	costs := make(map[int]struct{})
	for _, object := range g.state.Objects {
		if samePlayer(object.Owner, player) && g.cardHasSubtype(object.Card, "SUITED") {
			costs[g.printedReserveCost(object.Card)] = struct{}{}
		}
	}
	cost := baseCost - len(costs)
	if cost < 0 {
		return 0
	}
	return cost
}

func (g *Game) activateCardistry(player *model.Player, source objectID) error {
	if !g.canActivateCardistry(player, source) {
		return fmt.Errorf("%w %q", tcgErrors.ErrInvalidViewHandle, source)
	}
	object := g.state.Objects[source]
	baseCost, _ := g.cardistryBaseCost(object.Card)
	g.payCardistryCost(player, g.cardistryCost(player, baseCost))
	g.state.CardistryUsed[source] = true
	g.pushAbility(g.cardistryAbility(player, source, object.Card))
	g.recordPublicEvent(player, "cardistry", "ability-activated", object.Card)
	g.grantOpportunity(player)
	return nil
}

func (g *Game) payCardistryCost(player *model.Player, cost int) {
	zones := g.state.Zones[player.UID]
	for payment := 0; payment < cost; payment++ {
		index := int(g.nextRandom() % uint64(len(zones.Memory)))
		card := zones.Memory[index]
		zones.Memory = removeCardAt(zones.Memory, index)
		zones.Banishment = append(zones.Banishment, card)
	}
	g.state.Zones[player.UID] = zones
}

func (g *Game) cardistryAbility(player *model.Player, source objectID, card cardInstanceID) abilityInstance {
	operations := []effectOperation{}
	switch g.state.Cards[card].Definition {
	case wonderlandsReignCardID:
		operations = append(operations, effectOperation{Kind: effectOperationDraw, Amount: 1})
	case fiveOfSpadesCardID:
		operations = append(operations, g.temporaryModifierOperation(5, 0))
	case fourOfSpadesCardID:
		operations = append(operations, effectOperation{Kind: effectOperationDrawToMemory, Amount: 1})
	case fourOfHeartsCardID:
		operations = append(operations, effectOperation{Kind: effectOperationDrawToMemory, Amount: 1})
		operations = append(operations, effectOperation{Kind: effectOperationChooseMemoryAlly})
		operations = append(operations, effectOperation{Kind: effectOperationDeploy})
	case threeOfHeartsCardID:
		operations = append(operations, effectOperation{Kind: effectOperationDraw, Amount: 1})
		operations = append(operations, effectOperation{Kind: effectOperationChooseHandCard})
		operations = append(operations, effectOperation{Kind: effectOperationDiscard})
	case threeOfSpadesCardID:
		operations = append(operations, effectOperation{Kind: effectOperationChoose, Options: g.controlledSuitedAllies(player)})
		operations = append(operations, g.temporaryModifierOperation(0, 2))
	case twoOfHeartsCardID:
		operations = append(operations, g.temporaryModifierOperation(2, 0))
	case twoOfSpadesCardID:
		operations = append(operations, effectOperation{Kind: effectOperationCounter, Counter: "BUFF", Amount: 1})
	}
	return g.newAbilityInstance(player, card, source, operations)
}

func (g *Game) controlledSuitedAllies(player *model.Player) []objectID {
	objects := []objectID{}
	for id, object := range g.state.Objects {
		if samePlayer(object.Owner, player) && containsString(object.Types, "ALLY") && g.cardHasSubtype(object.Card, "SUITED") {
			objects = append(objects, id)
		}
	}
	sort.Slice(objects, func(first, second int) bool { return objects[first] < objects[second] })
	return objects
}

func (g *Game) qualifiedMemoryAllies(player *model.Player) []cardInstanceID {
	zones := g.state.Zones[player.UID]
	cards := []cardInstanceID{}
	for _, card := range zones.Memory {
		candidate := g.state.Cards[card]
		if containsString(candidate.Types, "ALLY") && containsString(candidate.Subtypes, "SUITED") && candidate.ReserveCost <= 3 && (containsString(candidate.Elements, "FIRE") || containsString(candidate.Elements, "NORM")) {
			cards = append(cards, card)
		}
	}
	sort.Slice(cards, func(first, second int) bool { return cards[first] < cards[second] })
	return cards
}

func (g *Game) deployAlly(player *model.Player, card cardInstanceID) {
	zones := g.state.Zones[player.UID]
	index := cardIndex(zones.Memory, card)
	if index < 0 {
		return
	}
	zones.Memory = removeCardAt(zones.Memory, index)
	g.state.Zones[player.UID] = zones
	g.state.NextObject++
	id := objectID(fmt.Sprintf("ally:%d", g.state.NextObject))
	candidate := g.state.Cards[card]
	g.state.Objects[id] = fieldObject{ID: id, Card: card, Owner: player, Types: candidate.Types}
	g.recordPublicEvent(player, "ability", "deploy", card)
	g.enqueueSuitedEnterAbility(player, id, card)
}

func (g *Game) suitedReserveTotal(player *model.Player) int {
	total := 0
	for _, object := range g.state.Objects {
		if samePlayer(object.Owner, player) && g.cardHasSubtype(object.Card, "SUITED") {
			total += g.printedReserveCost(object.Card)
		}
	}
	return total
}

func suitedThresholdAmount(total int) int {
	switch {
	case total >= 21:
		return 4
	case total >= 10:
		return 2
	case total >= 6:
		return 1
	default:
		return 0
	}
}

func (g *Game) enqueueSuitedEnterAbility(player *model.Player, source objectID, card cardInstanceID) {
	switch g.state.Cards[card].Definition {
	case noireCardID:
		amount := suitedThresholdAmount(g.suitedReserveTotal(player))
		if amount > 0 {
			g.pushAbility(g.newAbilityInstance(player, card, source, []effectOperation{{Kind: effectOperationCounter, Counter: "BUFF", Amount: amount}}))
		}
	case rougeCardID:
		if suitedThresholdAmount(g.suitedReserveTotal(player)) > 0 {
			g.pushAbility(g.newAbilityInstance(player, card, "", []effectOperation{{Kind: effectOperationChoose, Options: g.legalTargets()}, {Kind: effectOperationSuitedThresholdDamage}}))
		}
	}
}

func (g *Game) temporaryModifierOperation(power, life int) effectOperation {
	return effectOperation{
		Kind: effectOperationContinuousModifier,
		ContinuousEffect: continuousEffect{
			Scope:         effectScopeObject,
			Layer:         effectLayerModifier,
			PowerLife:     powerLifeModify,
			ExpiresAtTurn: g.state.Scheduler.TurnNumber + 1,
			Modifier: continuousModifier{
				PowerDelta: power,
				LifeDelta:  life,
			},
		},
	}
}
