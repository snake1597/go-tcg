package game

import (
	"testing"

	"go-tcg/internal/model"
)

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

// newWinningAttackViewFixture 建立公開戰力足以擊敗對方 Champion 的玩家視圖。
// 輸入為測試控制器；輸出為已刷新合法行動的 PlayerView，副作用只限隔離的測試單局。
func newWinningAttackViewFixture(t *testing.T) PlayerView {
	t.Helper()
	game := newActionGame(t)
	attacker := game.state.Champions[model.PlayerOne.UID]
	target := game.state.Champions[model.PlayerTwo.UID]
	game.state.Cards[attacker.Card] = withCombatStats(
		game.state.Cards[attacker.Card],
		5,
		10,
	)
	game.state.Cards[target.Card] = withCombatStats(
		game.state.Cards[target.Card],
		1,
		3,
	)
	game.state.Scheduler.TurnNumber = 2
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	return view
}

// newWieldMetadataViewFixture 建立含已追蹤 Impact Hammer 的合法 Wield 玩家視圖。
// 輸入為測試控制器；輸出為已刷新合法行動的 PlayerView，副作用只限隔離的測試單局。
func newWieldMetadataViewFixture(t *testing.T) PlayerView {
	t.Helper()
	game := newActionGame(t)
	weaponCard := findCard(t, game, model.PlayerOne, impactHammerCardID)
	weapon := objectID("weapon:metadata")
	game.state.Objects[weapon] = fieldObject{
		ID:    weapon,
		Card:  weaponCard,
		Owner: model.PlayerOne,
		Types: []string{
			"WEAPON",
		},
	}
	game.grantCardTracking(
		model.PlayerOne,
		entityID(weaponCard),
	)
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	return view
}
