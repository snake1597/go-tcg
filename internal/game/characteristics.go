package game

import "go-tcg/internal/model"

const (
	arthurYoungHeirCardID CardID = "GjM8b5fxqj"
	bulwarkSwordCardID    CardID = "8kmoi0a5uh"
)

func (g *Game) championHasClass(player *model.Player, classes []string) bool {
	champion, exists := g.state.Champions[player.UID]
	if !exists {
		return false
	}
	return containsAny(g.state.Cards[champion.Card].Classes, classes)
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
	return g.characteristicsFor(id).Immortal
}

func (g *Game) grantArthurImmortality(id objectID) {
	object, exists := g.state.Objects[id]
	if !exists || g.state.Cards[object.Card].Definition != arthurYoungHeirCardID {
		return
	}
	object.Rested = true
	g.state.Objects[id] = object
	g.addContinuousEffect(continuousEffect{
		Source:        id,
		Controller:    object.Owner,
		Target:        id,
		Scope:         effectScopeObject,
		Layer:         effectLayerAbility,
		ExpiresAtTurn: g.state.Scheduler.TurnNumber + uint64(len(g.players)),
		Modifier: continuousModifier{
			GrantImmortality: true,
		},
	})
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
