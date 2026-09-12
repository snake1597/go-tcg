package game

import (
	"testing"

	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

func TestActionCardsUsePlayerViewDeclarationAndResolveToGraveyard(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := actionByKind(t, view, constants.ActionActivate)
	if action.CardName != "Blazing Throw" {
		t.Fatalf("action card = %q, want Blazing Throw", action.CardName)
	}
	if err := game.Submit(player, Input{
		Revision: view.Revision,
		Action:   action.Handle,
	}); err != nil {
		t.Fatalf("Submit() declaration error = %v", err)
	}
	selectPendingChoiceSubject(t, game, player, entityID("champion:"+model.PlayerTwo.UID))
	selectPendingChoice(t, game, player, 0)
	passOpportunityRound(t, game, player)

	if got := len(game.state.Zones[player.UID].Graveyard); got != 2 {
		t.Fatalf("graveyard cards = %d, want sacrificed weapon and action", got)
	}
	if game.state.Champions[model.PlayerTwo.UID].Damage != 4 {
		t.Fatalf("target damage = %d, want 4", game.state.Champions[model.PlayerTwo.UID].Damage)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestFieryInterferenceCanBeActivatedByNonTurnPlayerAtFastTiming(t *testing.T) {
	game := newActionGameForPlayer(t, model.PlayerTwo, fieryInterferenceCardID)
	submitActionKind(t, game, model.PlayerOne, constants.ActionPass)
	view, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := actionByKind(t, view, constants.ActionActivate)
	if action.CardName != "Fiery Interference" {
		t.Fatalf("action card = %q, want Fiery Interference", action.CardName)
	}
	if err := game.Submit(model.PlayerTwo, Input{
		Revision: view.Revision,
		Action:   action.Handle,
	}); err != nil {
		t.Fatalf("Submit() declaration error = %v", err)
	}
	if game.state.Scheduler.OpportunityHolder != model.PlayerTwo {
		t.Fatalf("OpportunityHolder = %q, want activating non-turn player", game.state.Scheduler.OpportunityHolder)
	}
	selectPendingChoiceSubject(t, game, model.PlayerTwo, entityID("champion:"+model.PlayerOne.UID))
	passOpportunityRound(t, game, model.PlayerTwo)

	champion := game.state.Champions[model.PlayerOne.UID]
	if champion.Damage != 2 {
		t.Fatalf("target damage = %d, want 2", champion.Damage)
	}
	if champion.RecoverProhibitedUntilTurn != 1 {
		t.Fatalf("recover prohibition turn = %d, want current turn 1", champion.RecoverProhibitedUntilTurn)
	}
	if game.recoverChampion(objectID("champion:"+model.PlayerOne.UID), 1) {
		t.Fatal("recoverChampion() succeeded while Fiery Interference prohibition was active")
	}
	game.state.Scheduler.TurnNumber++
	game.expireTimedChampionEffects()
	if got := game.state.Champions[model.PlayerOne.UID].RecoverProhibitedUntilTurn; got != 0 {
		t.Fatalf("recover prohibition after turn boundary = %d, want expired", got)
	}
	if !game.recoverChampion(objectID("champion:"+model.PlayerOne.UID), 1) {
		t.Fatal("recoverChampion() failed after Fiery Interference prohibition expired")
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestStraightFlareCountsDistinctPrintedSuitedReserveCosts(t *testing.T) {
	game := newActionGameWithSource(t, straightFlareCardID)
	player := model.PlayerOne
	for index, wantCost := range []int{0, 1, 1} {
		card := suitedCardWithPrintedReserveCost(t, game, player, wantCost, index)
		object := objectID("suited:fixture:" + string(rune('a'+index)))
		game.state.Objects[object] = fieldObject{
			ID:    object,
			Card:  card,
			Owner: player,
			Types: []string{"SUITED"},
		}
	}
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if err := game.Submit(player, Input{
		Revision: view.Revision,
		Action:   actionByKind(t, view, constants.ActionActivate).Handle,
	}); err != nil {
		t.Fatalf("Submit() declaration error = %v", err)
	}
	selectPendingChoiceSubject(t, game, player, entityID("champion:"+model.PlayerTwo.UID))
	passOpportunityRound(t, game, player)
	if got := game.state.Champions[model.PlayerTwo.UID].Damage; got != 3 {
		t.Fatalf("target damage = %d, want 3 from two distinct printed costs plus one", got)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestActionFizzleMovesSourceToGraveyardWithoutUndoingCosts(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if err := game.Submit(player, Input{
		Revision: view.Revision,
		Action:   actionByKind(t, view, constants.ActionActivate).Handle,
	}); err != nil {
		t.Fatalf("Submit() declaration error = %v", err)
	}
	selectPendingChoiceSubject(t, game, player, entityID("champion:"+model.PlayerTwo.UID))
	selectPendingChoice(t, game, player, 0)
	delete(game.state.Champions, model.PlayerTwo.UID)
	passOpportunityRound(t, game, player)
	if got := len(game.state.Zones[player.UID].Graveyard); got != 2 {
		t.Fatalf("graveyard cards after fizzle = %d, want sacrificed weapon and action", got)
	}
	if got := len(game.state.EffectsStack); got != 0 {
		t.Fatalf("EffectsStack after fizzle = %#v, want empty", game.state.EffectsStack)
	}
}

func newActionGame(t *testing.T) *Game {
	return newActionGameWithSource(t, blazingThrowCardID)
}

func newActionGameWithSource(t *testing.T, definition CardID) *Game {
	return newActionGameForPlayer(t, model.PlayerOne, definition)
}

func newActionGameForPlayer(t *testing.T, player *model.Player, definition CardID) *Game {
	t.Helper()
	game, err := NewStandardSetup(StandardGameConfig{
		Players: [2]*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: "../..",
		Seed:           42,
	})
	if err != nil {
		t.Fatalf("NewStandardSetup() error = %v", err)
	}
	zones := game.state.Zones[player.UID]
	var source cardInstanceID
	for index, card := range zones.MainDeck {
		if game.state.Cards[card].Definition != definition {
			continue
		}
		source = card
		zones.MainDeck = removeCardAt(zones.MainDeck, index)
		zones.Hand = append(zones.Hand, card)
		break
	}
	if source == "" {
		t.Fatalf("fixed deck has no %q", definition)
	}
	for len(zones.Memory) < game.state.Cards[source].ReserveCost {
		payment := zones.Hand[0]
		if payment == source {
			payment = zones.Hand[1]
		}
		zones.Hand = removeCardAt(zones.Hand, cardIndex(zones.Hand, payment))
		zones.Memory = append(zones.Memory, payment)
	}
	game.state.Zones[player.UID] = zones
	if definition == blazingThrowCardID {
		weaponCard := zones.MaterialDeck[0]
		weapon := objectID("weapon:fixture")
		game.state.Objects[weapon] = fieldObject{
			ID:    weapon,
			Card:  weaponCard,
			Owner: player,
			Types: []string{"WEAPON"},
		}
	}
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	return game
}

func selectPendingChoice(t *testing.T, game *Game, player *model.Player, index int) {
	t.Helper()
	choice := game.state.Knowledge.Choice
	if choice == nil {
		t.Fatal("expected pending choice")
	}
	options := game.pendingChoice(player).Options
	if index >= len(options) {
		t.Fatalf("pending options = %#v, index = %d", options, index)
	}
	if err := game.Submit(player, Input{
		Revision: game.state.Revision,
		Choice:   options[index],
	}); err != nil {
		t.Fatalf("Submit() choice error = %v", err)
	}
}

func selectPendingChoiceSubject(t *testing.T, game *Game, player *model.Player, subject entityID) {
	t.Helper()
	choice := game.state.Knowledge.Choice
	if choice == nil {
		t.Fatal("expected pending choice")
	}
	for handle, candidate := range choice.Options {
		if candidate != subject {
			continue
		}
		if err := game.Submit(player, Input{
			Revision: game.state.Revision,
			Choice:   handle,
		}); err != nil {
			t.Fatalf("Submit() choice error = %v", err)
		}
		return
	}
	t.Fatalf("pending options = %#v, want %q", choice.Options, subject)
}

func suitedCardWithPrintedReserveCost(t *testing.T, game *Game, player *model.Player, wantCost, occurrence int) cardInstanceID {
	t.Helper()
	cards := append([]cardInstanceID(nil), game.state.Zones[player.UID].MainDeck...)
	cards = append(cards, game.state.Zones[player.UID].MaterialDeck...)
	seen := 0
	for _, card := range cards {
		if game.state.Cards[card].ReserveCost != wantCost || !game.cardHasSubtype(card, "SUITED") {
			continue
		}
		if seen == occurrence {
			return card
		}
		seen++
	}
	t.Fatalf("no card with printed reserve cost %d at occurrence %d", wantCost, occurrence)
	return ""
}
