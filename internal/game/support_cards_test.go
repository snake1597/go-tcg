package game

import (
	"testing"

	"go-tcg/internal/model"
)

func TestDuchessThornesObservesAllyCardistryAndDiscountIsConsumed(t *testing.T) {
	game := newCardistryGame(t, twoOfHeartsCardID)
	player := model.PlayerOne
	thornesCard := findCard(t, game, player, duchessThornesCardID)
	thornes := objectID("regalia:thornes")
	game.state.Objects[thornes] = fieldObject{
		ID:    thornes,
		Card:  thornesCard,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	source := cardistrySource(t, game, player)
	if err := game.activateCardistry(player, source, nil); err != nil {
		t.Fatalf("activateCardistry() error = %v", err)
	}
	if len(game.state.EffectsStack) != 2 {
		t.Fatalf("effects stack = %#v, want Cardistry and Thornes trigger", game.state.EffectsStack)
	}
	game.resolveTopEffectStack()
	if got := game.characteristicsFor(source); got.Power != game.state.Cards[game.state.Objects[source].Card].Power+1 || !got.TrueSight {
		t.Fatalf("observed Cardistry characteristics = %#v, want +1 power and true sight", got)
	}
	if err := game.activateDuchessThornes(player, thornes); err != nil {
		t.Fatalf("activateDuchessThornes() error = %v", err)
	}
	if got := game.cardistryCost(player, 6); got != 0 {
		t.Fatalf("discounted Cardistry cost = %d, want 0", got)
	}
}

func TestDuchessThornesDoesNotObservePhantasiaCardistry(t *testing.T) {
	game := newCardistryGame(t, wonderlandsReignCardID)
	player := model.PlayerOne
	fillMemory(t, game, player, 10, findCard(t, game, player, wonderlandsReignCardID))
	thornesCard := findCard(t, game, player, duchessThornesCardID)
	thornes := objectID("regalia:thornes")
	game.state.Objects[thornes] = fieldObject{
		ID:    thornes,
		Card:  thornesCard,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}

	if err := game.activateCardistry(player, cardistrySource(t, game, player), nil); err != nil {
		t.Fatalf("activateCardistry() error = %v", err)
	}
	if len(game.state.EffectsStack) != 1 {
		t.Fatalf("effects stack = %#v, want only Phantasia ability", game.state.EffectsStack)
	}
}

func TestDuchessThornesTriggerUsesLastKnownObserverAfterItLeaves(t *testing.T) {
	game := newCardistryGame(t, twoOfHeartsCardID)
	player := model.PlayerOne
	thornesCard := findCard(t, game, player, duchessThornesCardID)
	thornes := objectID("regalia:thornes")
	game.state.Objects[thornes] = fieldObject{
		ID:    thornes,
		Card:  thornesCard,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	source := cardistrySource(t, game, player)
	if err := game.activateCardistry(player, source, nil); err != nil {
		t.Fatalf("activateCardistry() error = %v", err)
	}
	delete(game.state.Objects, thornes)
	game.resolveTopEffectStack()

	got := game.characteristicsFor(source)
	card := game.state.Cards[game.state.Objects[source].Card]
	if got.Power != card.Power+1 || !got.TrueSight {
		t.Fatalf("observed Cardistry characteristics = %#v, want LKI trigger effects", got)
	}
}

func TestDuchessThornesDiscountExpiresAtTurnBoundary(t *testing.T) {
	game := newCardistryGame(t, twoOfHeartsCardID)
	player := model.PlayerOne
	game.state.CardistryDiscounts[player.UID] = 6
	game.state.Scheduler.Phase = PhaseEnd
	game.state.Scheduler.OpportunityHolder = player
	game.state.Scheduler.ConsecutivePasses = len(game.players) - 1

	if err := game.passOpportunity(player); err != nil {
		t.Fatalf("passOpportunity() error = %v", err)
	}
	if _, exists := game.state.CardistryDiscounts[player.UID]; exists {
		t.Fatal("Cardistry discount survived the turn boundary")
	}
}

func TestDuchessCopyAndVeritaAlternativeCostUseZonesAtomically(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	action := findCard(t, game, player, fieryInterferenceCardID)
	zones := game.state.Zones[player.UID]
	zones.Graveyard = append(zones.Graveyard, action)
	game.state.Zones[player.UID] = zones
	copy, err := game.copyDuchessAction(player, action, objectID("champion:"+model.PlayerTwo.UID))
	if err != nil {
		t.Fatalf("copyDuchessAction() error = %v", err)
	}
	if copy.Source == action || cardIndex(game.state.Zones[player.UID].Banishment, action) < 0 {
		t.Fatalf("copy = %#v, banishment = %#v; want independent copied source", copy, game.state.Zones[player.UID].Banishment)
	}
	cards := []cardInstanceID{}
	for _, card := range game.state.Zones[player.UID].MainDeck {
		if gCard := game.state.Cards[card]; containsString(gCard.Types, "ALLY") && gCard.ReserveCost > 0 {
			cards = append(cards, card)
			zones := game.state.Zones[player.UID]
			zones.MainDeck = removeCardAt(zones.MainDeck, cardIndex(zones.MainDeck, card))
			zones.Graveyard = append(zones.Graveyard, card)
			game.state.Zones[player.UID] = zones
		}
		if len(cards) == 3 {
			break
		}
	}
	if len(cards) != 3 {
		t.Fatal("fixture did not contain three ally cards")
	}
	for index := range cards {
		card := game.state.Cards[cards[index]]
		card.Subtypes = append(card.Subtypes, "SUITED")
		game.state.Cards[cards[index]] = card
	}
	for index, cost := range []int{3, 3, 4} {
		card := game.state.Cards[cards[index]]
		card.ReserveCost = cost
		game.state.Cards[cards[index]] = card
	}
	if err := game.payVeritaAlternativeCost(player, cards); err != nil {
		t.Fatalf("payVeritaAlternativeCost() error = %v", err)
	}
}

func TestSmokeBombsTrumpSetAndPepperedChefApplyTemporaryEffects(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	targetCard := findCard(t, game, player, twoOfHeartsCardID)
	target := objectID("ally:target")
	game.state.Objects[target] = fieldObject{
		ID:    target,
		Card:  targetCard,
		Owner: player,
		Types: []string{
			"ALLY",
		},
	}
	smokeCard := findCard(t, game, player, smokeBombsCardID)
	smoke := objectID("regalia:smoke")
	game.state.Objects[smoke] = fieldObject{
		ID:    smoke,
		Card:  smokeCard,
		Owner: player,
		Types: []string{
			"ITEM",
		},
	}
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	var smokeAction ViewHandle
	for _, action := range view.LegalActions {
		if action.CardName == "Smoke Bombs" {
			smokeAction = action.Handle
			break
		}
	}
	if smokeAction == "" {
		t.Fatal("Smoke Bombs action was not exposed")
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   smokeAction,
		},
	); err != nil {
		t.Fatalf("Submit() Smoke Bombs error = %v", err)
	}
	selectPendingChoiceSubject(t, game, player, entityID(target))
	if !game.characteristicsFor(target).Stealth {
		t.Fatal("Smoke Bombs did not grant stealth")
	}
	chefCard := findCard(t, game, player, pepperedChefCardID)
	chefDefinition := game.state.Cards[chefCard]
	chefDefinition.Subtypes = append(chefDefinition.Subtypes, "SUITED")
	game.state.Cards[chefCard] = chefDefinition
	chef := objectID("ally:chef")
	game.state.Objects[chef] = fieldObject{ID: chef, Card: chefCard, Owner: player, Types: []string{"ALLY"}}
	if err := game.sacrificeForPepperedChef(player, chef, target); err != nil {
		t.Fatalf("sacrificeForPepperedChef() error = %v", err)
	}
	if got := game.characteristicsFor(chef).Power; got != game.state.Cards[chefCard].Power+2 {
		t.Fatalf("Peppered Chef power = %d, want +2", got)
	}
	game.state.EffectsStack = append(game.state.EffectsStack, effectStackItem{
		Kind:       effectStackCombat,
		Controller: player,
		Target:     objectID("champion:" + model.PlayerTwo.UID),
	})
	if err := game.retargetAttackWithTrumpSet(player, chef); err != nil {
		t.Fatalf("retargetAttackWithTrumpSet() error = %v", err)
	}
	if got := game.state.EffectsStack[len(game.state.EffectsStack)-1].Target; got != chef {
		t.Fatalf("retargeted attack target = %q, want %q", got, chef)
	}
	if got := game.characteristicsFor(chef).Life; got != game.state.Cards[chefCard].Life+3 {
		t.Fatalf("Trump Set life = %d, want +3", got)
	}
}

func TestRedHareAndVeritaContinuousEffects(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	redHareCard := findCard(t, game, player, redHareCardID)
	redHare := objectID("ally:red-hare")
	game.state.Objects[redHare] = fieldObject{ID: redHare, Card: redHareCard, Owner: player, Types: []string{"ALLY"}}
	if game.redHareObeys(player, redHare) {
		t.Fatal("Red Hare obeyed without a level-three champion or qualifying ally")
	}
	duchessCard := findCard(t, game, player, duchessCardID)
	duchess := objectID("ally:duchess")
	game.state.Objects[duchess] = fieldObject{ID: duchess, Card: duchessCard, Owner: player, Types: []string{"ALLY", "UNIQUE"}}
	if !game.redHareObeys(player, redHare) {
		t.Fatal("Red Hare did not obey with a fire unique Human ally")
	}
	veritaCard := findCard(t, game, player, veritaCardID)
	verita := objectID("ally:verita")
	game.state.Objects[verita] = fieldObject{ID: verita, Card: veritaCard, Owner: player, Types: []string{"ALLY"}}
	if !game.characteristicsFor(duchess).Immortal {
		t.Fatal("Verita did not grant another Suited ally immortality")
	}
	delete(game.state.Objects, verita)
	game.enqueueVeritaDeath(player, veritaCard)
	if got := game.characteristicsFor(duchess).Power; got != game.state.Cards[duchessCard].Power+1 {
		t.Fatalf("Verita On Death power = %d, want +1", got)
	}
}
