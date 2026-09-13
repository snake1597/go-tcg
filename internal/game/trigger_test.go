package game

import (
	"testing"

	"go-tcg/internal/model"
)

func TestImpactHammerOnWieldUsesLastKnownInformationAfterSourceLeaves(t *testing.T) {
	game := newActionGame(t)
	weaponCard := findCard(t, game, model.PlayerOne, CardID("chsbalegbs"))
	weapon := objectID("weapon:impact-hammer")
	game.state.Objects[weapon] = fieldObject{
		ID:    weapon,
		Card:  weaponCard,
		Owner: model.PlayerOne,
		Types: []string{"WEAPON"},
	}
	unit := objectID("ally:target")
	game.state.Objects[unit] = fieldObject{
		ID:    unit,
		Card:  weaponCard,
		Owner: model.PlayerOne,
		Types: []string{"ALLY"},
	}
	if err := game.wieldWeapon(model.PlayerOne, unit, weapon); err != nil {
		t.Fatalf("wieldWeapon() error = %v", err)
	}
	if len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Ability == nil || game.state.EffectsStack[0].Ability.SourceLKI != weaponCard {
		t.Fatalf("trigger stack entry = %#v, want an Ability Instance with source LKI", game.state.EffectsStack)
	}
	delete(game.state.Objects, weapon)
	passOpportunityRound(t, game, model.PlayerOne)
	if got := game.state.Objects[unit].Damage; got != 3 {
		t.Fatalf("wielded unit damage = %d, want 3", got)
	}
}

func TestSameControllerOrdersSimultaneousTriggersBeforeOpportunity(t *testing.T) {
	game := newActionGame(t)
	first := impactHammerFixture(t, game, "one")
	second := impactHammerFixture(t, game, "two")
	unit := objectID("ally:target")
	game.state.Objects[unit] = fieldObject{
		ID:    unit,
		Card:  first,
		Owner: model.PlayerOne,
		Types: []string{"ALLY"},
	}
	if err := game.wieldWeapons(model.PlayerOne, unit, []cardInstanceID{first, second}); err != nil {
		t.Fatalf("wieldWeapons() error = %v", err)
	}
	if game.state.Knowledge.Choice == nil {
		t.Fatal("same-controller triggers did not create a Pending Choice")
	}
	batch := game.state.Events[len(game.state.Events)-1]
	if !batch.Simultaneous || batch.ParentFlow != "wield" {
		t.Fatalf("wield batch metadata = %#v, want simultaneous wield parent flow", batch)
	}
	if game.state.Scheduler.OpportunityHolder != nil {
		t.Fatalf("OpportunityHolder = %q, want nil until trigger ordering completes", game.state.Scheduler.OpportunityHolder)
	}
	selectPendingChoice(t, game, model.PlayerOne, 0)
	selectPendingChoice(t, game, model.PlayerOne, 0)
	if got := len(game.state.EffectsStack); got != 2 {
		t.Fatalf("EffectsStack entries = %d, want two ordered triggers", got)
	}
	if game.state.Scheduler.OpportunityHolder != model.PlayerOne {
		t.Fatalf("OpportunityHolder = %q, want trigger controller", game.state.Scheduler.OpportunityHolder)
	}
}

func impactHammerFixture(t *testing.T, game *Game, suffix string) cardInstanceID {
	t.Helper()
	card := findCard(t, game, model.PlayerOne, impactHammerCardID)
	clone := game.state.Cards[card]
	clone.ID = cardInstanceID(string(card) + ":" + suffix)
	game.state.Cards[clone.ID] = clone
	return clone.ID
}

func findCard(t *testing.T, game *Game, player *model.Player, definition CardID) cardInstanceID {
	t.Helper()
	for card, instance := range game.state.Cards {
		if samePlayer(instance.Owner, player) && instance.Definition == definition {
			return card
		}
	}
	t.Fatalf("no %q card for %q", definition, player)
	return ""
}
