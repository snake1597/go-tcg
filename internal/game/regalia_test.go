package game

import (
	"testing"

	"go-tcg/internal/model"
)

func TestGrandCrusadersRingBanishesThenDrawsThroughAbilityStack(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	card := findCard(t, game, player, grandCrusadersRingCardID)
	source := objectID("regalia:grand-crusaders-ring")
	game.state.Objects[source] = fieldObject{
		ID:    source,
		Card:  card,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	beforeHand := len(game.state.Zones[player.UID].Hand)
	if err := game.beginObjectAbility(player, source); err != nil {
		t.Fatalf("beginObjectAbility() error = %v", err)
	}
	if _, exists := game.state.Objects[source]; exists {
		t.Fatal("Grand Crusader's Ring remained on the field after paying banish cost")
	}
	if cardIndex(game.state.Zones[player.UID].Banishment, card) < 0 {
		t.Fatal("Grand Crusader's Ring was not banished")
	}
	if got := len(game.state.EffectsStack); got != 1 {
		t.Fatalf("EffectsStack = %d, want 1", got)
	}
	game.resolveTopEffectStack()
	if got := len(game.state.Zones[player.UID].Hand); got != beforeHand+1 {
		t.Fatalf("hand size = %d, want %d after draw", got, beforeHand+1)
	}
}

func TestViridianProtectiveTrinketTaxesOnlyOpponentWaterActivationOnControllerTurn(t *testing.T) {
	game := newActionGame(t)
	controller := model.PlayerOne
	opponent := model.PlayerTwo
	viridianCard := findCard(t, game, controller, viridianProtectiveTrinketCardID)
	viridian := objectID("regalia:viridian")
	game.state.Objects[viridian] = fieldObject{
		ID:    viridian,
		Card:  viridianCard,
		Owner: controller,
		Types: []string{
			"ITEM",
		},
	}
	action := findCard(t, game, opponent, blazingThrowCardID)
	candidate := game.state.Cards[action]
	candidate.Elements = []string{
		"WATER",
	}
	game.state.Cards[action] = candidate
	baseCost := candidate.ReserveCost
	if got := game.actionReserveCost(opponent, action); got != baseCost+2 {
		t.Fatalf("opponent water activation cost = %d, want %d", got, baseCost+2)
	}
	if got := game.actionReserveCost(controller, action); got != baseCost {
		t.Fatalf("controller water activation cost = %d, want %d", got, baseCost)
	}
	game.state.Scheduler.TurnPlayer = opponent
	if got := game.actionReserveCost(opponent, action); got != baseCost {
		t.Fatalf("off-turn water activation cost = %d, want %d", got, baseCost)
	}
}

func TestResonanceBaublesAreUnavailableWithoutMatchingOpponentChampion(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	for _, definition := range []CardID{
		waterResonanceBaubleCardID,
		windResonanceBaubleCardID,
	} {
		card := findCard(t, game, player, definition)
		source := objectID("regalia:" + string(definition))
		game.state.Objects[source] = fieldObject{
			ID:    source,
			Card:  card,
			Owner: player,
			Types: []string{
				"ITEM",
			},
		}
		if game.canActivateObjectAbility(player, source) {
			t.Fatalf("%q ability is legal without a matching opponent Champion", definition)
		}
	}
	game.advanceKnowledgeRevision()
	if got := len(game.state.Knowledge.ObjectAbilities[player.UID]); got != 0 {
		t.Fatalf("object ability handles = %#v, want none", game.state.Knowledge.ObjectAbilities[player.UID])
	}
}
