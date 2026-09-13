package game

import (
	"testing"

	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

func TestChampionAttackUsesPlayerViewAndDealsSimultaneousCombatDamage(t *testing.T) {
	game := newActionGame(t)
	attacker := game.state.Champions[model.PlayerOne.UID]
	target := game.state.Champions[model.PlayerTwo.UID]
	game.state.Cards[attacker.Card] = withCombatStats(game.state.Cards[attacker.Card], 3, 10)
	game.state.Cards[target.Card] = withCombatStats(game.state.Cards[target.Card], 2, 10)
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()

	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	attack := actionByKind(t, view, constants.ActionAttack)
	if err := game.Submit(model.PlayerOne, Input{
		Revision: view.Revision,
		Action:   attack.Handle,
	}); err != nil {
		t.Fatalf("Submit() attack declaration error = %v", err)
	}
	selectPendingChoiceSubject(t, game, model.PlayerOne, entityID(target.ID))
	passOpportunityRound(t, game, model.PlayerOne)

	if got := game.state.Champions[model.PlayerOne.UID].Damage; got != 2 {
		t.Fatalf("attacker damage = %d, want retaliation 2", got)
	}
	if got := game.state.Champions[model.PlayerTwo.UID].Damage; got != 3 {
		t.Fatalf("target damage = %d, want attack 3", got)
	}
	if !game.state.Champions[model.PlayerOne.UID].Rested {
		t.Fatal("attacker was not rested as attack cost")
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestCombatStateBasedCheckEndsGameWhenChampionIsDefeated(t *testing.T) {
	game := newActionGame(t)
	attacker := game.state.Champions[model.PlayerOne.UID]
	target := game.state.Champions[model.PlayerTwo.UID]
	game.state.Cards[attacker.Card] = withCombatStats(game.state.Cards[attacker.Card], 5, 10)
	game.state.Cards[target.Card] = withCombatStats(game.state.Cards[target.Card], 1, 3)
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if err := game.Submit(model.PlayerOne, Input{
		Revision: view.Revision,
		Action:   actionByKind(t, view, constants.ActionAttack).Handle,
	}); err != nil {
		t.Fatalf("Submit() attack declaration error = %v", err)
	}
	selectPendingChoiceSubject(t, game, model.PlayerOne, entityID(target.ID))
	passOpportunityRound(t, game, model.PlayerOne)
	if !game.state.Finished || game.state.Winner != model.PlayerOne {
		t.Fatalf("combat result = finished:%t winner:%q, want player one victory", game.state.Finished, game.state.Winner)
	}
	batch := game.state.Events[len(game.state.Events)-1]
	if batch.Cause != "combat:on-kill" || batch.ParentFlow != "combat:damage" || batch.Events[0].Kind != "on-kill" {
		t.Fatalf("on-kill batch = %#v, want combat-linked on-kill event", batch)
	}
}

func TestCombatStateBasedCheckDestroysLethalAllyToOwnerGraveyard(t *testing.T) {
	game := newActionGame(t)
	card := findCard(t, game, model.PlayerTwo, CardID("rufki4o41y"))
	game.state.Cards[card] = withCombatStats(game.state.Cards[card], 1, 1)
	ally := objectID("ally:lethal")
	game.state.Objects[ally] = fieldObject{
		ID:     ally,
		Card:   card,
		Owner:  model.PlayerTwo,
		Types:  []string{"ALLY"},
		Damage: 1,
	}
	game.resolveCombatStateBased()
	if _, exists := game.state.Objects[ally]; exists {
		t.Fatal("lethal ally remained on the Field")
	}
	if cardIndex(game.state.Zones[model.PlayerTwo.UID].Graveyard, card) < 0 {
		t.Fatalf("owner graveyard = %#v, want destroyed ally %q", game.state.Zones[model.PlayerTwo.UID].Graveyard, card)
	}
}

func TestWieldUsesPlayerViewLegalActionAndTargetChoice(t *testing.T) {
	game := newActionGame(t)
	weaponCard := findCard(t, game, model.PlayerOne, impactHammerCardID)
	weapon := objectID("weapon:impact-hammer")
	game.state.Objects[weapon] = fieldObject{
		ID:    weapon,
		Card:  weaponCard,
		Owner: model.PlayerOne,
		Types: []string{"WEAPON"},
	}
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if err := game.Submit(model.PlayerOne, Input{
		Revision: view.Revision,
		Action:   actionByKind(t, view, constants.ActionWield).Handle,
	}); err != nil {
		t.Fatalf("Submit() wield declaration error = %v", err)
	}
	target := game.state.Champions[model.PlayerTwo.UID]
	selectPendingChoiceSubject(t, game, model.PlayerOne, entityID(target.ID))
	if got := game.state.Events[len(game.state.Events)-1].Cause; got != "wield" {
		t.Fatalf("wield event cause = %q, want wield", got)
	}
}

func TestCombatRecordsSimultaneousHitAndRetaliationEvents(t *testing.T) {
	game := newActionGame(t)
	attacker := game.state.Champions[model.PlayerOne.UID]
	target := game.state.Champions[model.PlayerTwo.UID]
	game.state.Cards[attacker.Card] = withCombatStats(game.state.Cards[attacker.Card], 3, 10)
	game.state.Cards[target.Card] = withCombatStats(game.state.Cards[target.Card], 2, 10)
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if err := game.Submit(model.PlayerOne, Input{
		Revision: view.Revision,
		Action:   actionByKind(t, view, constants.ActionAttack).Handle,
	}); err != nil {
		t.Fatalf("Submit() attack declaration error = %v", err)
	}
	selectPendingChoiceSubject(t, game, model.PlayerOne, entityID(target.ID))
	passOpportunityRound(t, game, model.PlayerOne)
	batch := game.state.Events[len(game.state.Events)-1]
	if !batch.Simultaneous || batch.Cause != "combat:damage" || batch.ParentFlow != "attack" {
		t.Fatalf("combat batch metadata = %#v, want simultaneous attack damage", batch)
	}
	if len(batch.Events) != 2 || batch.Events[0].Kind != "on-hit:attack" || batch.Events[1].Kind != "on-hit:retaliation" {
		t.Fatalf("combat events = %#v, want ordered hit and retaliation", batch.Events)
	}
}

func withCombatStats(card cardInstance, power, life int) cardInstance {
	card.Power = power
	card.Life = life
	return card
}
