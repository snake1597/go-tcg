package game

import "go-tcg/internal/model"

const (
	arthurYoungHeirCardID CardID = "GjM8b5fxqj"
	bulwarkSwordCardID    CardID = "8kmoi0a5uh"
)

type characteristics struct {
	Power int
	Life  int
}

// characteristicsFor is the sole source for a unit's derived combat values.
// Later layer modifiers extend this query rather than mutating card instances.
func (g *Game) characteristicsFor(id objectID) characteristics {
	card, exists := g.cardForObject(id)
	if !exists {
		return characteristics{}
	}
	result := characteristics{
		Power: card.Power,
		Life:  card.Life,
	}
	if card.Definition == bulwarkSwordCardID && g.championHasClass(card.Owner, card.Classes) {
		result.Power++
	}
	if containsString(card.Types, "ALLY") && g.hasRestedArthur(card.Owner, id) {
		result.Power++
	}
	return result
}

func (g *Game) championHasClass(player *model.Player, classes []string) bool {
	champion, exists := g.state.Champions[player.UID]
	if !exists {
		return false
	}
	return containsAny(g.state.Cards[champion.Card].Classes, classes)
}

func (g *Game) hasRestedArthur(player *model.Player, exclude objectID) bool {
	for id, object := range g.state.Objects {
		if id == exclude || !samePlayer(object.Owner, player) || !object.Rested || g.state.Cards[object.Card].Definition != arthurYoungHeirCardID {
			continue
		}
		return true
	}
	return false
}

func containsAny(first, second []string) bool {
	for _, value := range first {
		if containsString(second, value) {
			return true
		}
	}
	return false
}

func (g *Game) isImmortal(id objectID) bool {
	object, exists := g.state.Objects[id]
	return exists && object.ImmortalUntilTurn > 0 && object.ImmortalUntilTurn >= g.state.Scheduler.TurnNumber
}

func (g *Game) grantArthurImmortality(id objectID) {
	object, exists := g.state.Objects[id]
	if !exists || g.state.Cards[object.Card].Definition != arthurYoungHeirCardID {
		return
	}
	object.Rested = true
	object.ImmortalUntilTurn = g.state.Scheduler.TurnNumber
	g.state.Objects[id] = object
}

func (g *Game) cardForObject(id objectID) (cardInstance, bool) {
	for _, champion := range g.state.Champions {
		if champion.ID == id {
			card, exists := g.state.Cards[champion.Card]
			return card, exists
		}
	}
	object, exists := g.state.Objects[id]
	if !exists {
		return cardInstance{}, false
	}
	card, exists := g.state.Cards[object.Card]
	return card, exists
}
