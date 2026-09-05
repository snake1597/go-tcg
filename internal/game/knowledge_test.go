package game

import (
	"go-tcg/internal/constants"
	"strings"
	"testing"
)

func TestPlayerViewProjectsOnlyTrackedCardsAndVisibleHistory(t *testing.T) {
	game := NewGame(42)
	secretCard := game.addKnowledgeFixtureCard(
		constants.PlayerOne,
		"Secret Flame",
	)
	game.recordVisibleEvent(
		constants.PlayerOne,
		"reveal",
		secretCard,
	)

	firstView, err := game.PlayerView(constants.PlayerOne)
	if err != nil {
		t.Fatalf("first PlayerView() error = %v", err)
	}
	secondView, err := game.PlayerView(constants.PlayerTwo)
	if err != nil {
		t.Fatalf("second PlayerView() error = %v", err)
	}
	if len(firstView.Cards) != 1 || firstView.Cards[0].Name != "Secret Flame" {
		t.Fatalf("first PlayerView().Cards = %#v, want Secret Flame", firstView.Cards)
	}
	if len(secondView.Cards) != 0 {
		t.Fatalf("second PlayerView().Cards = %#v, want no private cards", secondView.Cards)
	}
	if len(firstView.VisibleEvents) != 1 || firstView.VisibleEvents[0].CardName != "Secret Flame" {
		t.Fatalf("first PlayerView().VisibleEvents = %#v, want recorded reveal", firstView.VisibleEvents)
	}
	if len(secondView.VisibleEvents) != 0 {
		t.Fatalf("second PlayerView().VisibleEvents = %#v, want no private event", secondView.VisibleEvents)
	}
	if strings.Contains(string(firstView.Cards[0].Handle), string(secretCard)) {
		t.Fatalf("card handle %q exposed internal card identity %q", firstView.Cards[0].Handle, secretCard)
	}
}

func TestPlayerViewRevokesTrackingHandleButRetainsRevealHistory(t *testing.T) {
	game := NewGame(42)
	secretCard := game.addKnowledgeFixtureCard(
		constants.PlayerOne,
		"Secret Flame",
	)
	game.recordVisibleEvent(
		constants.PlayerOne,
		"reveal",
		secretCard,
	)
	beforeShuffle, err := game.PlayerView(constants.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() before shuffle error = %v", err)
	}
	oldHandle := beforeShuffle.Cards[0].Handle

	game.revokeCardTracking(
		constants.PlayerOne,
		secretCard,
	)
	game.advanceKnowledgeRevision()

	afterShuffle, err := game.PlayerView(constants.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() after shuffle error = %v", err)
	}
	if len(afterShuffle.Cards) != 0 {
		t.Fatalf("PlayerView().Cards = %#v, want no tracked cards after shuffle", afterShuffle.Cards)
	}
	if len(afterShuffle.VisibleEvents) != 1 || afterShuffle.VisibleEvents[0].CardName != "Secret Flame" {
		t.Fatalf("PlayerView().VisibleEvents = %#v, want retained reveal", afterShuffle.VisibleEvents)
	}

	game.grantCardTracking(
		constants.PlayerOne,
		secretCard,
	)
	game.advanceKnowledgeRevision()

	afterReturn, err := game.PlayerView(constants.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() after return error = %v", err)
	}
	if len(afterReturn.Cards) != 1 {
		t.Fatalf("PlayerView().Cards = %#v, want one newly visible card", afterReturn.Cards)
	}
	if afterReturn.Cards[0].Handle == oldHandle {
		t.Fatalf("re-granted card reused revoked handle %q", oldHandle)
	}
}

func TestPendingChoiceAcceptsOnlyCurrentPlayersVisibleHandle(t *testing.T) {
	game := NewGame(42)
	secretCard := game.addKnowledgeFixtureCard(
		constants.PlayerOne,
		"Secret Flame",
	)
	game.setPendingCardChoice(
		constants.PlayerOne,
		secretCard,
	)

	firstView, err := game.PlayerView(constants.PlayerOne)
	if err != nil {
		t.Fatalf("first PlayerView() error = %v", err)
	}
	secondView, err := game.PlayerView(constants.PlayerTwo)
	if err != nil {
		t.Fatalf("second PlayerView() error = %v", err)
	}
	if firstView.PendingChoice == nil || len(firstView.PendingChoice.Options) != 1 {
		t.Fatalf("first PlayerView().PendingChoice = %#v, want one option", firstView.PendingChoice)
	}
	if secondView.PendingChoice != nil {
		t.Fatalf("second PlayerView().PendingChoice = %#v, want nil", secondView.PendingChoice)
	}

	choice := firstView.PendingChoice.Options[0]
	beforeHash := game.StateHash()
	crossPlayerInput := Input{
		Revision: secondView.Revision,
		Choice:   choice,
	}
	err = game.Submit(constants.PlayerTwo, crossPlayerInput)
	if err == nil || !strings.Contains(err.Error(), "invalid view handle") {
		t.Fatalf("cross-player Submit() error = %v, want invalid view handle", err)
	}
	if game.StateHash() != beforeHash {
		t.Fatalf("StateHash() changed after cross-player choice submission")
	}
	input := Input{
		Revision: firstView.Revision,
		Choice:   choice,
	}
	if err := game.Submit(constants.PlayerOne, input); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	afterChoice, err := game.PlayerView(constants.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() after choice error = %v", err)
	}
	if afterChoice.PendingChoice != nil {
		t.Fatalf("PlayerView().PendingChoice = %#v, want nil", afterChoice.PendingChoice)
	}
	if len(afterChoice.Cards) != 1 || afterChoice.Cards[0].Handle != choice {
		t.Fatalf("PlayerView().Cards = %#v, want tracked card with stable handle", afterChoice.Cards)
	}
}

func TestRevokingTrackingRevokesPendingChoiceOptionWithoutChangingStateOnSubmission(t *testing.T) {
	game := NewGame(42)
	secretCard := game.addKnowledgeFixtureCard(
		constants.PlayerOne,
		"Secret Flame",
	)
	game.setPendingCardChoice(
		constants.PlayerOne,
		secretCard,
	)
	view, err := game.PlayerView(constants.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	oldChoice := view.PendingChoice.Options[0]

	game.revokeCardTracking(
		constants.PlayerOne,
		secretCard,
	)
	game.advanceKnowledgeRevision()

	afterRevocation, err := game.PlayerView(constants.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() after revocation error = %v", err)
	}
	if afterRevocation.PendingChoice != nil {
		t.Fatalf("PlayerView().PendingChoice = %#v, want nil", afterRevocation.PendingChoice)
	}
	beforeHash := game.StateHash()
	input := Input{
		Revision: afterRevocation.Revision,
		Choice:   oldChoice,
	}
	err = game.Submit(constants.PlayerOne, input)
	if err == nil || !strings.Contains(err.Error(), "invalid view handle") {
		t.Fatalf("Submit() error = %v, want invalid view handle", err)
	}
	if game.StateHash() != beforeHash {
		t.Fatalf("StateHash() changed after revoked choice submission")
	}
}

func TestPendingChoiceOmitsUntrackedCards(t *testing.T) {
	game := NewGame(42)
	secretCard := game.addKnowledgeFixtureCard(
		constants.PlayerOne,
		"Secret Flame",
	)

	game.setPendingCardChoice(
		constants.PlayerTwo,
		secretCard,
	)

	view, err := game.PlayerView(constants.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if view.PendingChoice != nil {
		t.Fatalf("PlayerView().PendingChoice = %#v, want nil", view.PendingChoice)
	}
}
