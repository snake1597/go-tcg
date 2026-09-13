package game

import (
	"testing"

	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

func TestCardistryFloatingMemoryIsDeclaredRecordedAndReplayed(t *testing.T) {
	game := newCardistryGame(t, wonderlandsReignCardID)
	player := model.PlayerOne
	floatingMemory := findCard(t, game, player, fiveOfSpadesCardID)
	moveCardToGraveyard(t, game, player, floatingMemory)
	game.grantCardTracking(player, entityID(floatingMemory))
	fillMemory(t, game, player, 8, floatingMemory)
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()

	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if got := cardistryAction(t, game, player); got == "" {
		t.Fatal("cardistry action = empty, want a legal action with Floating Memory")
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   cardistryAction(t, game, player),
			FloatingMemory: []ViewHandle{
				game.state.Knowledge.Cards[player.UID][entityID(floatingMemory)],
			},
		},
	); err != nil {
		t.Fatalf("Submit() Cardistry error = %v", err)
	}
	if cardIndex(game.state.Zones[player.UID].Banishment, floatingMemory) < 0 {
		t.Fatal("Floating Memory card was not banished")
	}
	if !hasGameEvent(game, "banish-floating-memory") || !hasGameEvent(game, "banish-memory") || !hasGameEvent(game, "ability-activated") {
		t.Fatalf("events = %#v, want Floating Memory payment, memory payment, and activation", game.state.Events)
	}
	passOpportunityRound(t, game, player)
	if !hasGameEvent(game, "draw") {
		t.Fatalf("events = %#v, want draw event", game.state.Events)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestCardistryInvalidFloatingMemoryDoesNotChangeState(t *testing.T) {
	game := newCardistryGame(t, wonderlandsReignCardID)
	player := model.PlayerOne
	notFloatingMemory := findCard(t, game, player, twoOfHeartsCardID)
	moveCardToGraveyard(t, game, player, notFloatingMemory)
	game.grantCardTracking(player, entityID(notFloatingMemory))
	fillMemory(t, game, player, 9, notFloatingMemory)
	game.advanceKnowledgeRevision()
	before := game.StateHash()

	err := game.activateCardistry(
		player,
		cardistrySource(t, game, player),
		[]ViewHandle{
			game.state.Knowledge.Cards[player.UID][entityID(notFloatingMemory)],
		},
	)
	if err == nil {
		t.Fatal("activateCardistry() error = nil, want invalid Floating Memory")
	}
	if got := game.StateHash(); got != before {
		t.Fatalf("StateHash() = %q, want unchanged %q", got, before)
	}
}

func TestCardistryRejectedActivationsDoNotChangeState(t *testing.T) {
	t.Run(
		"Floating Memory is only accepted with Cardistry",
		func(t *testing.T) {
			game := newCardistryGame(t, twoOfSpadesCardID)
			player := model.PlayerOne
			var pass ViewHandle
			for handle, kind := range game.state.Knowledge.Actions[player.UID] {
				if kind == constants.ActionPass {
					pass = handle
					break
				}
			}
			if pass == "" {
				t.Fatal("no pass action")
			}
			before := game.StateHash()
			if err := game.Submit(
				player,
				Input{
					Revision: game.state.Revision,
					Action:   pass,
					FloatingMemory: []ViewHandle{
						"not-a-floating-memory-handle",
					},
				},
			); err == nil {
				t.Fatal("Submit() error = nil, want Floating Memory rejection for pass")
			}
			if got := game.StateHash(); got != before {
				t.Fatalf("StateHash() = %q, want unchanged %q", got, before)
			}
		},
	)
	t.Run(
		"insufficient payment",
		func(t *testing.T) {
			game := newCardistryGame(t, wonderlandsReignCardID)
			player := model.PlayerOne
			zones := game.state.Zones[player.UID]
			zones.Memory = nil
			game.state.Zones[player.UID] = zones
			before := game.StateHash()
			if err := game.activateCardistry(
				player,
				cardistrySource(t, game, player),
				nil,
			); err == nil {
				t.Fatal("activateCardistry() error = nil, want insufficient payment rejection")
			}
			if got := game.StateHash(); got != before {
				t.Fatalf("StateHash() = %q, want unchanged %q", got, before)
			}
		},
	)
	t.Run(
		"once per instance",
		func(t *testing.T) {
			game := newCardistryGame(t, twoOfSpadesCardID)
			player := model.PlayerOne
			fillMemory(t, game, player, 2, "")
			game.advanceKnowledgeRevision()
			view, err := game.PlayerView(player)
			if err != nil {
				t.Fatalf("PlayerView() error = %v", err)
			}
			action := cardistryAction(t, game, player)
			if err := game.Submit(
				player,
				Input{
					Revision: view.Revision,
					Action:   action,
				},
			); err != nil {
				t.Fatalf("first Submit() error = %v", err)
			}
			before := game.StateHash()
			if err := game.Submit(
				player,
				Input{
					Revision: game.state.Revision,
					Action:   action,
				},
			); err == nil {
				t.Fatal("second Submit() error = nil, want once-per-instance rejection")
			}
			if got := game.StateHash(); got != before {
				t.Fatalf("StateHash() = %q, want unchanged %q", got, before)
			}
		},
	)
}

func TestCardistryDrawFromEmptyDeckHasNoDrawEvent(t *testing.T) {
	game := newCardistryGame(t, wonderlandsReignCardID)
	player := model.PlayerOne
	zones := game.state.Zones[player.UID]
	zones.MainDeck = nil
	game.state.Zones[player.UID] = zones
	source := cardistrySource(t, game, player)
	beforeEvents := len(game.state.Events)
	game.pushAbility(game.cardistryAbility(player, source, game.state.Objects[source].Card))
	game.resolveTopEffectStack()
	if got := len(game.state.Events); got != beforeEvents {
		t.Fatalf("event count = %d, want unchanged %d after draw from an empty deck", got, beforeEvents)
	}
}

func TestCardistrySourceLeavingFizzesTemporaryModifier(t *testing.T) {
	game := newCardistryGame(t, fiveOfSpadesCardID)
	player := model.PlayerOne
	source := cardistrySource(t, game, player)
	game.pushAbility(game.cardistryAbility(player, source, game.state.Objects[source].Card))
	delete(game.state.Objects, source)
	game.resolveTopEffectStack()
	if len(game.state.ContinuousEffects) != 0 {
		t.Fatalf("continuous effects = %#v, want no effect after source left", game.state.ContinuousEffects)
	}
	if hasGameEvent(game, "continuous-effect") {
		t.Fatalf("events = %#v, want no modifier event after source left", game.state.Events)
	}
}

func TestCardistryOperationsRecordCountersAndModifiers(t *testing.T) {
	game := newCardistryGame(t, twoOfSpadesCardID)
	player := model.PlayerOne
	fillMemory(t, game, player, 2, "")
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   cardistryAction(t, game, player),
		},
	); err != nil {
		t.Fatalf("Submit() Cardistry error = %v", err)
	}
	passOpportunityRound(t, game, player)
	if !hasGameEvent(game, "counter") {
		t.Fatalf("events = %#v, want counter event", game.state.Events)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestFourOfHeartsDeploymentAndTriggerAreEventedAndReplayed(t *testing.T) {
	game := newCardistryGame(t, fourOfHeartsCardID)
	player := model.PlayerOne
	noire := findCard(t, game, player, noireCardID)
	moveCardToMemory(t, game, player, noire)
	game.grantCardTracking(player, entityID(noire))
	addSuitedAlly(t, game, player, "zero", 0)
	addSuitedAlly(t, game, player, "one", 1)
	addSuitedAlly(t, game, player, "two", 2)
	addSuitedAlly(t, game, player, "three", 3)
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()

	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   cardistryAction(t, game, player),
		},
	); err != nil {
		t.Fatalf("Submit() Cardistry error = %v", err)
	}
	passOpportunityRound(t, game, player)
	if game.state.Knowledge.Choice == nil {
		t.Fatalf("choice = nil, memory = %#v, stack = %#v", game.state.Zones[player.UID].Memory, game.state.EffectsStack)
	}
	selectPendingChoiceSubject(t, game, player, entityID(noire))
	passOpportunityRound(t, game, player)
	passOpportunityRound(t, game, player)

	if !hasGameEvent(game, "draw-to-memory") || !hasGameEvent(game, "deploy") || !hasGameEvent(game, "counter") {
		t.Fatalf("events = %#v, want draw-to-memory, deploy, and resulting counter", game.state.Events)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestThreeOfSpadesTemporaryModifierIsEventedAndExpires(t *testing.T) {
	game := newCardistryGame(t, threeOfSpadesCardID)
	player := model.PlayerOne
	target := addSuitedAlly(t, game, player, "target", 0)
	if !game.cardHasSubtype(game.state.Objects[target].Card, "SUITED") {
		t.Fatalf("target card = %#v, want SUITED subtype", game.state.Cards[game.state.Objects[target].Card])
	}
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   cardistryAction(t, game, player),
		},
	); err != nil {
		t.Fatalf("Submit() Cardistry error = %v", err)
	}
	passOpportunityRound(t, game, player)
	if game.state.Knowledge.Choice == nil {
		t.Fatalf("choice = nil, stack = %#v", game.state.EffectsStack)
	}
	selectPendingChoiceSubject(t, game, player, entityID(target))
	passOpportunityRound(t, game, player)
	if !hasGameEvent(game, "continuous-effect") {
		t.Fatalf("events = %#v, want continuous-effect", game.state.Events)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestSuitedThresholdsHandleBoundariesAndSourceChanges(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		total int
		want  int
	}{
		{
			name:  "below six",
			total: 5,
			want:  0,
		},
		{
			name:  "six",
			total: 6,
			want:  1,
		},
		{
			name:  "ten",
			total: 10,
			want:  2,
		},
		{
			name:  "twenty one",
			total: 21,
			want:  4,
		},
	} {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				if got := suitedThresholdAmount(testCase.total); got != testCase.want {
					t.Fatalf("suitedThresholdAmount(%d) = %d, want %d", testCase.total, got, testCase.want)
				}
			},
		)
	}

	game := newActionGame(t)
	player := model.PlayerOne
	noireCard := findCard(t, game, player, noireCardID)
	noire := objectID("ally:noire")
	game.state.Objects[noire] = fieldObject{
		ID:    noire,
		Card:  noireCard,
		Owner: player,
		Types: []string{
			"ALLY",
		},
	}
	companion := addSuitedAlly(t, game, player, "companion", 3)
	if !game.characteristicsFor(noire).Stealth {
		t.Fatal("Noire has no stealth with another Suited ally")
	}
	delete(game.state.Objects, companion)
	if game.characteristicsFor(noire).Stealth {
		t.Fatal("Noire retained stealth after the only other Suited ally left")
	}
}

func TestSuitedEnterTriggersReevaluateThresholdsAndHandleSourceLeaving(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	rouge := findCard(t, game, player, rougeCardID)
	moveCardToMemory(t, game, player, rouge)
	addSuitedAlly(t, game, player, "rouge-companion", 5)
	game.deployAlly(player, rouge)
	game.resolveTopEffectStack()
	if game.state.Knowledge.Choice == nil {
		t.Fatal("Rouge On Enter did not create a target choice")
	}
	selectPendingChoiceSubject(t, game, player, entityID("champion:"+model.PlayerTwo.UID))
	delete(game.state.Objects, objectID("ally:rouge-companion"))
	game.resolveTopEffectStack()
	if got := game.state.Champions[model.PlayerTwo.UID].Damage; got != 0 {
		t.Fatalf("Rouge damage = %d, want 0 after threshold source left before resolution", got)
	}

	noire := findCard(t, game, player, noireCardID)
	moveCardToMemory(t, game, player, noire)
	addSuitedAlly(t, game, player, "noire-companion", 4)
	game.deployAlly(player, noire)
	if len(game.state.EffectsStack) == 0 {
		t.Fatal("Noire On Enter did not create a counter trigger")
	}
	delete(game.state.Objects, objectForCard(t, game, noire))
	game.resolveTopEffectStack()
	if hasGameEvent(game, "counter") {
		t.Fatalf("events = %#v, want no counter after Noire left", game.state.Events)
	}
}

func newCardistryGame(t *testing.T, definition CardID) *Game {
	t.Helper()
	game := newActionGame(t)
	player := model.PlayerOne
	source := findCard(t, game, player, definition)
	id := objectID("cardistry:source")
	game.state.Objects[id] = fieldObject{
		ID:    id,
		Card:  source,
		Owner: player,
		Types: game.state.Cards[source].Types,
	}
	game.state.Scheduler.Kind = schedulerStable
	game.state.Scheduler.Phase = PhaseMain
	game.state.Scheduler.TurnPlayer = player
	game.state.Scheduler.OpportunityHolder = player
	game.advanceKnowledgeRevision()
	return game
}

func cardistrySource(t *testing.T, game *Game, player *model.Player) objectID {
	t.Helper()
	for source, object := range game.state.Objects {
		if samePlayer(object.Owner, player) && gCardistrySource(game, object.Card) {
			return source
		}
	}
	t.Fatal("no Cardistry source")
	return ""
}

func gCardistrySource(game *Game, card cardInstanceID) bool {
	baseCost, _ := game.cardistryBaseCost(card)
	return baseCost >= 0
}

func cardistryAction(t *testing.T, game *Game, player *model.Player) ViewHandle {
	t.Helper()
	for handle, source := range game.state.Knowledge.Cardistries[player.UID] {
		if source == objectID("cardistry:source") {
			return handle
		}
	}
	t.Fatal("no Cardistry source action")
	return ""
}

func fillMemory(t *testing.T, game *Game, player *model.Player, want int, exclude cardInstanceID) {
	t.Helper()
	zones := game.state.Zones[player.UID]
	for len(zones.Memory) < want {
		card, zone := cardForMemoryPayment(&zones, exclude)
		if card == "" {
			t.Fatal("not enough cards to fill memory")
		}
		*zone = removeCardAt(*zone, cardIndex(*zone, card))
		zones.Memory = append(zones.Memory, card)
	}
	game.state.Zones[player.UID] = zones
}

func cardForMemoryPayment(zones *playerZones, exclude cardInstanceID) (cardInstanceID, *[]cardInstanceID) {
	for _, zone := range []*[]cardInstanceID{
		&zones.Hand,
		&zones.MainDeck,
		&zones.MaterialDeck,
	} {
		for _, card := range *zone {
			if card != exclude {
				return card, zone
			}
		}
	}
	return "", nil
}

func moveCardToGraveyard(t *testing.T, game *Game, player *model.Player, card cardInstanceID) {
	t.Helper()
	zones := game.state.Zones[player.UID]
	for _, zone := range []*[]cardInstanceID{
		&zones.Hand,
		&zones.MainDeck,
		&zones.MaterialDeck,
	} {
		index := cardIndex(*zone, card)
		if index < 0 {
			continue
		}
		*zone = removeCardAt(*zone, index)
		zones.Graveyard = append(zones.Graveyard, card)
		game.state.Zones[player.UID] = zones
		return
	}
	t.Fatalf("card %q is not in a movable zone", card)
}

func moveCardToMemory(t *testing.T, game *Game, player *model.Player, card cardInstanceID) {
	t.Helper()
	zones := game.state.Zones[player.UID]
	for _, zone := range []*[]cardInstanceID{
		&zones.Hand,
		&zones.MainDeck,
		&zones.MaterialDeck,
	} {
		index := cardIndex(*zone, card)
		if index < 0 {
			continue
		}
		*zone = removeCardAt(*zone, index)
		zones.Memory = append(zones.Memory, card)
		game.state.Zones[player.UID] = zones
		return
	}
	t.Fatalf("card %q is not in a movable zone", card)
}

func hasGameEvent(game *Game, want string) bool {
	for _, batch := range game.state.Events {
		for _, event := range batch.Events {
			if event.Kind == want {
				return true
			}
		}
	}
	return false
}

func addSuitedAlly(t *testing.T, game *Game, player *model.Player, suffix string, reserveCost int) objectID {
	t.Helper()
	card := findCard(t, game, player, twoOfHeartsCardID)
	instance := game.state.Cards[card]
	instance.ID = cardInstanceID(string(card) + ":" + suffix)
	instance.ReserveCost = reserveCost
	game.state.Cards[instance.ID] = instance
	id := objectID("ally:" + suffix)
	game.state.Objects[id] = fieldObject{
		ID:    id,
		Card:  instance.ID,
		Owner: player,
		Types: []string{
			"ALLY",
		},
	}
	return id
}

func objectForCard(t *testing.T, game *Game, card cardInstanceID) objectID {
	t.Helper()
	for id, object := range game.state.Objects {
		if object.Card == card {
			return id
		}
	}
	t.Fatalf("no object for card %q", card)
	return ""
}
