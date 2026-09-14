package game

import (
	"testing"

	"go-tcg/internal/constants"
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
		Controller: model.PlayerTwo,
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

func TestSmokeBombsRevalidatesAttackTargetsForStealthAndTrueSight(t *testing.T) {
	game := newActionGame(t)
	defender := model.PlayerOne
	attacker := model.PlayerTwo
	targetCard := findCard(t, game, defender, duchessCardID)
	target := objectID("ally:smoke-target")
	game.state.Objects[target] = fieldObject{
		ID:    target,
		Card:  targetCard,
		Owner: defender,
		Types: []string{
			"ALLY",
		},
	}
	smokeCard := findCard(t, game, defender, smokeBombsCardID)
	smoke := objectID("regalia:smoke")
	game.state.Objects[smoke] = fieldObject{
		ID:    smoke,
		Card:  smokeCard,
		Owner: defender,
		Types: []string{
			"ITEM",
		},
	}
	if err := game.activateSmokeBombs(defender, smoke, target); err != nil {
		t.Fatalf("activateSmokeBombs() error = %v", err)
	}
	attackingChampion := game.state.Champions[attacker.UID]
	defendingChampion := game.state.Champions[defender.UID]
	defendingChampion.TauntUntilTurn = 0
	game.state.Champions[defender.UID] = defendingChampion
	game.resolveCombat(effectStackItem{
		Kind:       effectStackCombat,
		Controller: attacker,
		Target:     target,
		Attacker:   attackingChampion.ID,
	})
	if got := game.state.Objects[target].Damage; got != 0 {
		t.Fatalf("stealthed target damage = %d, want 0", got)
	}
	if _, exists := game.state.Objects[target]; !exists {
		t.Fatal("stealthed target left the field")
	}
	game.addContinuousEffect(continuousEffect{
		Target: attackingChampion.ID,
		Scope:  effectScopeObject,
		Layer:  effectLayerAbility,
		Modifier: continuousModifier{
			GrantTrueSight: true,
			PowerDelta:     1,
		},
	})
	game.resolveCombat(effectStackItem{
		Kind:       effectStackCombat,
		Controller: attacker,
		Target:     target,
		Attacker:   attackingChampion.ID,
	})
	if got := game.state.Objects[target].Damage; got == 0 {
		t.Fatalf("true sight attack did not damage the stealthed target; attacker characteristics = %#v, attack targets = %#v", game.characteristicsFor(attackingChampion.ID), game.attackTargets(attacker, attackingChampion.ID))
	}
	damageBeforeTaunt := game.state.Objects[target].Damage
	defendingChampion = game.state.Champions[defender.UID]
	defendingChampion.TauntUntilTurn = game.state.Scheduler.TurnNumber + 1
	game.state.Champions[defender.UID] = defendingChampion
	game.resolveCombat(effectStackItem{
		Kind:       effectStackCombat,
		Controller: attacker,
		Target:     target,
		Attacker:   attackingChampion.ID,
	})
	if got := game.state.Objects[target].Damage; got != damageBeforeTaunt {
		t.Fatalf("taunt-restricted target damage = %d, want %d", got, damageBeforeTaunt)
	}
}

func TestTrumpSetRequiresDifferentLegalSuitedAllyAndFizzlesDeterministically(t *testing.T) {
	game := newActionGameWithSource(t, trumpSetCardID)
	defender := model.PlayerOne
	attacker := model.PlayerTwo
	firstCard := findCard(t, game, defender, twoOfHeartsCardID)
	first := objectID("ally:first-suited")
	game.state.Objects[first] = fieldObject{
		ID:    first,
		Card:  firstCard,
		Owner: defender,
		Types: []string{
			"ALLY",
		},
	}
	secondCard := findCard(t, game, defender, threeOfSpadesCardID)
	second := objectID("ally:second-suited")
	game.state.Objects[second] = fieldObject{
		ID:    second,
		Card:  secondCard,
		Owner: defender,
		Types: []string{
			"ALLY",
		},
	}
	attackingChampion := game.state.Champions[attacker.UID]
	game.state.EffectsStack = append(game.state.EffectsStack, effectStackItem{
		Kind:       effectStackCombat,
		Controller: attacker,
		Target:     first,
		Attacker:   attackingChampion.ID,
	})
	if got := game.trumpSetTargets(defender); len(got) != 1 || got[0] != second {
		t.Fatalf("trumpSetTargets() = %#v, want only %q", got, second)
	}
	if !game.canActivateAction(defender, findCard(t, game, defender, trumpSetCardID)) {
		t.Fatal("Trump Set was not legal with a distinct suited ally")
	}
	if err := game.retargetAttackWithTrumpSet(defender, second); err != nil {
		t.Fatalf("retargetAttackWithTrumpSet() error = %v", err)
	}
	if got := game.state.EffectsStack[0].Target; got != second {
		t.Fatalf("retargeted attack target = %q, want %q", got, second)
	}
	if got := game.characteristicsFor(second).Power; got != game.state.Cards[secondCard].Power+3 {
		t.Fatalf("Trump Set power = %d, want +3", got)
	}
	game.state.Scheduler.TurnNumber++
	if got := game.characteristicsFor(second).Power; got != game.state.Cards[secondCard].Power {
		t.Fatalf("expired Trump Set power = %d, want %d", got, game.state.Cards[secondCard].Power)
	}
	delete(game.state.Champions, attacker.UID)
	trumpCard := findCard(t, game, defender, trumpSetCardID)
	ability := game.newAbilityInstance(
		defender,
		trumpCard,
		second,
		[]effectOperation{
			{
				Kind: effectOperationRetargetAttack,
			},
			{
				Kind:                  effectOperationMove,
				MoveSourceToGraveyard: true,
			},
		},
	)
	game.resolveAbility(ability)
	if cardIndex(game.state.Zones[defender.UID].Graveyard, trumpCard) < 0 {
		t.Fatal("Trump Set did not move to the graveyard after its attack source became invalid")
	}
}

func TestRedHareAndVeritaContinuousEffects(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	redHareCard := findCard(t, game, player, redHareCardID)
	redHare := objectID("ally:red-hare")
	game.state.Objects[redHare] = fieldObject{ID: redHare, Card: redHareCard, Owner: player, Types: []string{"ALLY"}}
	if game.canAttackWith(player, redHare) {
		t.Fatal("Red Hare obeyed without a level-three champion or qualifying ally")
	}
	duchessCard := findCard(t, game, player, duchessCardID)
	duchess := objectID("ally:duchess")
	game.state.Objects[duchess] = fieldObject{ID: duchess, Card: duchessCard, Owner: player, Types: []string{"ALLY", "UNIQUE"}}
	if !game.canAttackWith(player, redHare) {
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

func TestHeatedVengeanceTracksChampionDamageAndResolvesOptionalOnAttack(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	heatedCard := findCard(t, game, player, heatedVengeanceCardID)
	heated := objectID("attack:heated-vengeance")
	game.state.Objects[heated] = fieldObject{
		ID:    heated,
		Card:  heatedCard,
		Owner: player,
		Types: []string{
			"ALLY",
		},
	}
	champion := game.state.Champions[player.UID]
	championCard := game.state.Cards[champion.Card]
	championCard.Classes = []string{
		"WARRIOR",
	}
	game.state.Cards[champion.Card] = championCard
	if got := game.characteristicsFor(heated).Power; got != 2 {
		t.Fatalf("Heated Vengeance power = %d, want 2 before champion damage", got)
	}
	game.damageUnit(champion.ID, 1)
	if got := game.characteristicsFor(heated).Power; got != 5 {
		t.Fatalf("Heated Vengeance power = %d, want 5 after champion damage", got)
	}
	game.state.Scheduler.TurnNumber++
	if got := game.characteristicsFor(heated).Power; got != 2 {
		t.Fatalf("Heated Vengeance power = %d, want 2 after turn change", got)
	}
	triggers := game.onAttackTriggers(player, heated)
	if len(triggers) != 1 || triggers[0].Ability == nil {
		t.Fatalf("On Attack triggers = %#v, want one Ability Instance", triggers)
	}
	game.flushTriggers(triggers)
	game.resolveTopEffectStack()
	if game.state.AbilityChoice == nil || !game.state.AbilityChoice.CanPass {
		t.Fatalf("Heated Vengeance choice = %#v, want optional self-damage choice", game.state.AbilityChoice)
	}
	delete(game.state.Objects, heated)
	selectPendingChoice(t, game, player, 0)
	game.resolveTopEffectStack()
	if got := game.state.Champions[player.UID].Damage; got != 4 {
		t.Fatalf("champion damage = %d, want prior 1 plus 3 from LKI trigger", got)
	}
	game.state.Objects[heated] = fieldObject{
		ID:    heated,
		Card:  heatedCard,
		Owner: player,
		Types: []string{
			"ALLY",
		},
	}
	game.flushTriggers(game.onAttackTriggers(player, heated))
	game.resolveTopEffectStack()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   actionByKind(t, view, constants.ActionPass).Handle,
		},
	); err != nil {
		t.Fatalf("Submit() optional self-damage pass error = %v", err)
	}
	if got := game.state.Champions[player.UID].Damage; got != 4 {
		t.Fatalf("champion damage = %d, want unchanged after optional pass", got)
	}
}

func TestRedHarePermissionAndGrantedAttackAbilityUseDerivedCharacteristics(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	redHareCard := findCard(t, game, player, redHareCardID)
	redHare := objectID("ally:red-hare")
	game.state.Objects[redHare] = fieldObject{
		ID:    redHare,
		Card:  redHareCard,
		Owner: player,
		Types: []string{
			"ALLY",
		},
	}
	if game.canAttackWith(player, redHare) || game.characteristicsFor(redHare).GrantedOnAttack {
		t.Fatal("Red Hare gained permission or its granted ability without a qualifying Human ally")
	}
	duchessCard := findCard(t, game, player, duchessCardID)
	duchess := objectID("ally:duchess")
	game.state.Objects[duchess] = fieldObject{
		ID:    duchess,
		Card:  duchessCard,
		Owner: player,
		Types: []string{
			"ALLY",
			"UNIQUE",
		},
	}
	if !game.canAttackWith(player, redHare) || game.characteristicsFor(redHare).Pride != 0 || !game.characteristicsFor(redHare).GrantedOnAttack {
		t.Fatalf("Red Hare characteristics = %#v, want removed Pride and granted On Attack", game.characteristicsFor(redHare))
	}
	triggers := game.onAttackTriggers(player, redHare)
	if len(triggers) != 1 || triggers[0].Ability == nil || !triggers[0].Ability.Operations[0].CanPass {
		t.Fatalf("granted On Attack triggers = %#v, want optional Ability Instance", triggers)
	}
	game.state.Zones[player.UID] = playerZones{}
	if got := game.onAttackTriggers(player, redHare); got != nil {
		t.Fatalf("On Attack triggers without discard options = %#v, want none", got)
	}
	delete(game.state.Objects, duchess)
	if game.canAttackWith(player, redHare) || game.characteristicsFor(redHare).GrantedOnAttack {
		t.Fatal("Red Hare retained removed Pride or granted ability after source left")
	}
	game.state.Objects[duchess] = fieldObject{
		ID:    duchess,
		Card:  duchessCard,
		Owner: player,
		Types: []string{
			"ALLY",
			"UNIQUE",
		},
	}
	delete(game.state.Objects, redHare)
	game.state.Objects[redHare] = fieldObject{
		ID:    redHare,
		Card:  redHareCard,
		Owner: player,
		Types: []string{
			"ALLY",
		},
	}
	if !game.canAttackWith(player, redHare) || !game.characteristicsFor(redHare).GrantedOnAttack {
		t.Fatal("re-entered Red Hare did not receive fresh derived permission and ability")
	}
}
