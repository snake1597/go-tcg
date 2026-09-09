package game

import "go-tcg/internal/model"

func (g *Game) addKnowledgeFixtureCard(player *model.Player, name string) entityID {
	card := entityID("fixture-card")
	g.state.Entities[card] = knowledgeEntity{
		Name: name,
	}
	g.grantCardTracking(
		player,
		card,
	)
	return card
}
