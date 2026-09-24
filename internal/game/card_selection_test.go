package game

import (
	"testing"

	"go-tcg/internal/model"
)

// TestCardSelectionFromGraveyardDoesNotMoveCards 驗證選牌只回傳合格墓地牌，並保留所有區域原狀。
func TestCardSelectionFromGraveyardDoesNotMoveCards(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	zones := game.state.Zones[player.UID]
	card := zones.Hand[0]
	zones.Hand = removeCardAt(zones.Hand, 0)
	zones.Graveyard = append(zones.Graveyard, card)
	game.state.Zones[player.UID] = zones
	spec := cardSelectionSpec{
		Zone:         cardZoneGraveyard,
		MinimumCards: 1,
	}
	selected := game.findCardSelection(player, spec, "", nil)
	if len(selected) != 1 || selected[0] != card {
		t.Fatalf("findCardSelection() = %#v, want graveyard card %q", selected, card)
	}
	if cardIndex(game.state.Zones[player.UID].Graveyard, card) < 0 || cardIndex(game.state.Zones[player.UID].Banishment, card) >= 0 {
		t.Fatal("card selection moved a card")
	}
}
