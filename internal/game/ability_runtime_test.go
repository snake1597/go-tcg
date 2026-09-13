package game

import (
	"testing"

	"go-tcg/internal/model"
)

func TestAbilityChoiceUsesVisibleHandleAndResumesTypedOperations(t *testing.T) {
	game := newActionGame(t)
	source := findCard(t, game, model.PlayerOne, blazingThrowCardID)
	target := objectID("champion:" + model.PlayerTwo.UID)
	instance := game.newAbilityInstance(
		model.PlayerOne,
		source,
		"",
		[]effectOperation{
			{
				Kind:    effectOperationChoose,
				Options: []objectID{target},
			},
			{
				Kind:   effectOperationDamage,
				Amount: 3,
			},
		},
	)
	game.pushAbility(instance)
	game.resolveTopEffectStack()
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if view.PendingChoice == nil || len(view.PendingChoice.Options) != 1 {
		t.Fatalf("PendingChoice = %#v, want one visible option", view.PendingChoice)
	}
	if otherView, err := game.PlayerView(model.PlayerTwo); err != nil || otherView.PendingChoice != nil {
		t.Fatalf("other player choice = %#v, %v; want nil, no error", otherView.PendingChoice, err)
	}
	if err := game.Submit(model.PlayerOne, Input{
		Revision: game.state.Revision,
		Choice:   view.PendingChoice.Options[0],
	}); err != nil {
		t.Fatalf("Submit() choice error = %v", err)
	}
	if len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Ability == nil {
		t.Fatalf("EffectsStack = %#v, want resumed Ability Instance", game.state.EffectsStack)
	}
	game.resolveTopEffectStack()
	if got := game.state.Champions[model.PlayerTwo.UID].Damage; got != 3 {
		t.Fatalf("target damage = %d, want 3", got)
	}
}

func TestTypedOperationsDrawAndCounterProduceStateChanges(t *testing.T) {
	game := newActionGame(t)
	source := findCard(t, game, model.PlayerOne, blazingThrowCardID)
	target := objectID("counter-target")
	game.state.Objects[target] = fieldObject{
		ID:    target,
		Card:  source,
		Owner: model.PlayerOne,
	}
	beforeDraw := len(game.state.Zones[model.PlayerOne.UID].Hand)
	game.pushAbility(game.newAbilityInstance(
		model.PlayerOne,
		source,
		target,
		[]effectOperation{
			{
				Kind:   effectOperationDraw,
				Amount: 1,
			},
			{
				Kind:    effectOperationCounter,
				Counter: "POWER",
				Amount:  2,
			},
		},
	))
	game.resolveTopEffectStack()
	if got := len(game.state.Zones[model.PlayerOne.UID].Hand); got != beforeDraw+1 {
		t.Fatalf("hand size = %d, want %d", got, beforeDraw+1)
	}
	if got := game.state.Objects[target].Counters["POWER"]; got != 2 {
		t.Fatalf("POWER counters = %d, want 2", got)
	}
	if got := game.characteristicsFor(target).Power; got != game.state.Cards[source].Power+2 {
		t.Fatalf("derived power = %d, want printed power plus counters", got)
	}
}
