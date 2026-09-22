package game

import (
	"go-tcg/internal/model"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlayerViewProjectsOnlyTrackedCardsAndVisibleHistory(t *testing.T) {
	game := newTestGame(42)
	secretCard := game.addKnowledgeFixtureCard(
		model.PlayerOne,
		"Secret Flame",
	)
	game.recordVisibleEvent(
		model.PlayerOne,
		"reveal",
		secretCard,
	)

	firstView, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("first PlayerView() error = %v", err)
	}
	secondView, err := game.PlayerView(model.PlayerTwo)
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

// TestPlayerViewProjectsOwnHandPublicFieldAndEffectStack 驗證玩家視圖會分別投影自己的手牌、公開場上物件與公開效果堆疊。
// 輸入為含手牌、場上物件及效果項目的標準單局；輸出為不含內部識別的可見資料，副作用為零。
func TestPlayerViewProjectsOwnHandPublicFieldAndEffectStack(t *testing.T) {
	game, err := NewStandardGame(StandardGameConfig{
		Players: []*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: filepath.Clean("../.."),
		Seed:           42,
	})
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
	}
	firstView, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("first PlayerView() error = %v", err)
	}
	if len(firstView.Hand) != 7 {
		t.Fatalf("first PlayerView().Hand count = %d, want 7", len(firstView.Hand))
	}
	if len(firstView.Field) != 0 {
		t.Fatalf("first PlayerView().Field = %#v, want no field objects", firstView.Field)
	}
	if len(firstView.EffectsStack) != 0 {
		t.Fatalf("first PlayerView().EffectsStack = %#v, want empty stack", firstView.EffectsStack)
	}

	card := game.state.Zones[model.PlayerOne.UID].Hand[0]
	object := objectID("fixture-field")
	game.state.Objects[object] = fieldObject{
		ID:    object,
		Card:  card,
		Owner: model.PlayerOne,
		Types: []string{
			"ALLY",
		},
	}
	game.state.EffectsStack = append(
		game.state.EffectsStack,
		effectStackItem{
			Kind:       effectStackAbility,
			Controller: model.PlayerOne,
			Source:     card,
		},
	)

	secondView, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("second PlayerView() error = %v", err)
	}
	if len(secondView.Hand) != 7 {
		t.Fatalf("second PlayerView().Hand count = %d, want 7", len(secondView.Hand))
	}
	if len(secondView.Field) != 1 || secondView.Field[0].CardName == "" {
		t.Fatalf("second PlayerView().Field = %#v, want one named public object", secondView.Field)
	}
	if len(secondView.EffectsStack) != 1 || secondView.EffectsStack[0].SourceName == "" {
		t.Fatalf("second PlayerView().EffectsStack = %#v, want one named public stack item", secondView.EffectsStack)
	}
	if strings.Contains(string(secondView.Field[0].CardName), string(object)) {
		t.Fatalf("field card name %q exposed internal object identity %q", secondView.Field[0].CardName, object)
	}
}

// TestPlayerViewGroupsFieldByPlayerAndPlayOrder 驗證場上卡牌先依玩家座位分組，再依進場順序呈現。
// 輸入為兩位玩家交錯進場的四張 Ally；輸出為各玩家維持自身出牌順序的公開場上投影，副作用為零。
func TestPlayerViewGroupsFieldByPlayerAndPlayOrder(t *testing.T) {
	game := newTestGame(42)
	addCard := func(id cardInstanceID, owner *model.Player, name string) cardInstanceID {
		game.state.Cards[id] = cardInstance{
			ID:    id,
			Owner: owner,
		}
		game.state.Entities[entityID(id)] = knowledgeEntity{
			Name: name,
		}
		return id
	}
	firstHeart := addCard(
		"card:first-heart",
		model.PlayerOne,
		"First Heart",
	)
	secondHeart := addCard(
		"card:second-heart",
		model.PlayerTwo,
		"Second Heart",
	)
	firstSpade := addCard(
		"card:first-spade",
		model.PlayerOne,
		"First Spade",
	)
	secondNoire := addCard(
		"card:second-noire",
		model.PlayerTwo,
		"Second Noire",
	)
	for _, fixture := range []struct {
		id    objectID
		card  cardInstanceID
		owner *model.Player
	}{
		{
			id:    "ally:1",
			card:  firstHeart,
			owner: model.PlayerOne,
		},
		{
			id:    "ally:2",
			card:  secondHeart,
			owner: model.PlayerTwo,
		},
		{
			id:    "ally:3",
			card:  firstSpade,
			owner: model.PlayerOne,
		},
		{
			id:    "ally:4",
			card:  secondNoire,
			owner: model.PlayerTwo,
		},
	} {
		game.state.Objects[fixture.id] = fieldObject{
			ID:    fixture.id,
			Card:  fixture.card,
			Owner: fixture.owner,
			Types: []string{
				"ALLY",
			},
		}
	}

	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	want := []struct {
		owner string
		name  string
	}{
		{
			owner: model.PlayerOne.UID,
			name:  "First Heart",
		},
		{
			owner: model.PlayerOne.UID,
			name:  "First Spade",
		},
		{
			owner: model.PlayerTwo.UID,
			name:  "Second Heart",
		},
		{
			owner: model.PlayerTwo.UID,
			name:  "Second Noire",
		},
	}
	if len(view.Field) != len(want) {
		t.Fatalf("PlayerView().Field = %#v, want %d objects", view.Field, len(want))
	}
	for index, expected := range want {
		if view.Field[index].Owner.UID != expected.owner || view.Field[index].CardName != expected.name {
			t.Fatalf("PlayerView().Field[%d] = %#v, want owner %q card %q", index, view.Field[index], expected.owner, expected.name)
		}
	}
}

func TestPlayerViewRevokesTrackingHandleButRetainsRevealHistory(t *testing.T) {
	game := newTestGame(42)
	secretCard := game.addKnowledgeFixtureCard(
		model.PlayerOne,
		"Secret Flame",
	)
	game.recordVisibleEvent(
		model.PlayerOne,
		"reveal",
		secretCard,
	)
	beforeShuffle, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() before shuffle error = %v", err)
	}
	oldHandle := beforeShuffle.Cards[0].Handle

	game.revokeCardTracking(
		model.PlayerOne,
		secretCard,
	)
	game.advanceKnowledgeRevision()

	afterShuffle, err := game.PlayerView(model.PlayerOne)
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
		model.PlayerOne,
		secretCard,
	)
	game.advanceKnowledgeRevision()

	afterReturn, err := game.PlayerView(model.PlayerOne)
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
	game := newTestGame(42)
	secretCard := game.addKnowledgeFixtureCard(
		model.PlayerOne,
		"Secret Flame",
	)
	game.setPendingCardChoice(
		model.PlayerOne,
		secretCard,
	)

	firstView, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("first PlayerView() error = %v", err)
	}
	secondView, err := game.PlayerView(model.PlayerTwo)
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
	err = game.Submit(model.PlayerTwo, crossPlayerInput)
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
	if err := game.Submit(model.PlayerOne, input); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	afterChoice, err := game.PlayerView(model.PlayerOne)
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

// TestPendingChoiceNamesChampionTarget 驗證攻擊目標以公開 Champion 卡名投影，而非無意義的通用選項。
// 輸入為玩家可見的 Champion object 選項；輸出為含卡名的 PendingChoice，副作用僅為建立測試狀態。
func TestPendingChoiceNamesChampionTarget(t *testing.T) {
	game := newTestGame(42)
	championCard := cardInstanceID("champion-card:player-2")
	championID := objectID("champion:player-2")
	game.state.Cards[championCard] = cardInstance{
		ID:    championCard,
		Owner: model.PlayerTwo,
	}
	game.state.Entities[entityID(championCard)] = knowledgeEntity{
		Name: "Spirit of Fire",
	}
	game.state.Champions[model.PlayerTwo.UID] = championObject{
		ID:    championID,
		Card:  championCard,
		Owner: model.PlayerTwo,
	}
	game.state.Knowledge.Choice = &pendingChoice{
		Actor: model.PlayerOne,
		Options: map[ViewHandle]entityID{
			"target": entityID(championID),
		},
	}

	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if view.PendingChoice == nil || len(view.PendingChoice.Choices) != 1 || view.PendingChoice.Choices[0].CardName != "Spirit of Fire" {
		t.Fatalf("PlayerView().PendingChoice = %#v, want named Champion target", view.PendingChoice)
	}
}

func TestRevokingTrackingRevokesPendingChoiceOptionWithoutChangingStateOnSubmission(t *testing.T) {
	game := newTestGame(42)
	secretCard := game.addKnowledgeFixtureCard(
		model.PlayerOne,
		"Secret Flame",
	)
	game.setPendingCardChoice(
		model.PlayerOne,
		secretCard,
	)
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	oldChoice := view.PendingChoice.Options[0]

	game.revokeCardTracking(
		model.PlayerOne,
		secretCard,
	)
	game.advanceKnowledgeRevision()

	afterRevocation, err := game.PlayerView(model.PlayerOne)
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
	err = game.Submit(model.PlayerOne, input)
	if err == nil || !strings.Contains(err.Error(), "invalid view handle") {
		t.Fatalf("Submit() error = %v, want invalid view handle", err)
	}
	if game.StateHash() != beforeHash {
		t.Fatalf("StateHash() changed after revoked choice submission")
	}
}

func TestPendingChoiceOmitsUntrackedCards(t *testing.T) {
	game := newTestGame(42)
	secretCard := game.addKnowledgeFixtureCard(
		model.PlayerOne,
		"Secret Flame",
	)

	game.setPendingCardChoice(
		model.PlayerTwo,
		secretCard,
	)

	view, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if view.PendingChoice != nil {
		t.Fatalf("PlayerView().PendingChoice = %#v, want nil", view.PendingChoice)
	}
}
