package game

import "go-tcg/internal/constants"

func (g *Game) addKnowledgeFixtureCard(player constants.PlayerID, name string) entityID {
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
