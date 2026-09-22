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
	action := actionByCardName(t, view, "Blazing Throw")
	if err := game.Submit(player, Input{
		Revision: view.Revision,
		Action:   action.Handle,
		Reserve:  reserveHandles(action),
	}); err != nil {
		t.Fatalf("Submit() declaration error = %v", err)
	}
	selectPendingChoiceSubject(t, game, player, entityID("champion:"+model.PlayerTwo.UID))
	selectPendingChoice(t, game, player, 0)
	if len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Ability == nil {
		t.Fatalf("EffectsStack = %#v, want one Ability Instance", game.state.EffectsStack)
	}
	ability := game.state.EffectsStack[0].Ability
	if ability.Controller != player || ability.Source != ability.SourceLKI || ability.Target != objectID("champion:"+model.PlayerTwo.UID) {
		t.Fatalf("Ability Instance = %#v, want controller, source LKI, and fixed target", ability)
	}
	if !containsCard(game.state.EffectSources, ability.Source) {
		t.Fatalf("EffectSources = %#v, want declared action source %q", game.state.EffectSources, ability.Source)
	}
	passOpportunityRound(t, game, player)

	if containsCard(game.state.EffectSources, ability.Source) {
		t.Fatalf("EffectSources = %#v, want resolved action source removed", game.state.EffectSources)
	}
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
	action := actionByCardName(t, view, "Fiery Interference")
	if action.CardName != "Fiery Interference" {
		t.Fatalf("action card = %q, want Fiery Interference", action.CardName)
	}
	if err := game.Submit(model.PlayerTwo, Input{
		Revision: view.Revision,
		Action:   action.Handle,
		Reserve:  reserveHandles(action),
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
	if !game.characteristicsFor(objectID("champion:" + model.PlayerOne.UID)).RecoverProhibited {
		t.Fatal("derived characteristics did not prohibit recovery")
	}
	if game.recoverChampion(objectID("champion:"+model.PlayerOne.UID), 1) {
		t.Fatal("recoverChampion() succeeded while Fiery Interference prohibition was active")
	}
	game.state.Scheduler.TurnNumber++
	game.expireTimedChampionEffects()
	if game.characteristicsFor(objectID("champion:" + model.PlayerOne.UID)).RecoverProhibited {
		t.Fatal("derived recovery prohibition remained after turn boundary")
	}
	if !game.recoverChampion(objectID("champion:"+model.PlayerOne.UID), 1) {
		t.Fatal("recoverChampion() failed after Fiery Interference prohibition expired")
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

// TestAllyActivationUsesEffectsStackBeforeEnteringField 驗證 Ally activation 支付後先建立可回應的 Stack item，結算後才成為 Object。
// 輸入為 Red Hare 的 PlayerView action 與引擎提供的 Reserve handles；輸出為 activation window 及最終 Ally object，副作用為提交 activation 並讓雙方 pass 結算。
func TestAllyActivationUsesEffectsStackBeforeEnteringField(t *testing.T) {
	game := newActionGameWithSource(t, redHareCardID)
	player := model.PlayerOne
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := actionByCardName(t, view, "Red Hare, Unrivaled Stallion")
	source := game.state.Knowledge.Players[player.UID].Activations[action.Handle]

	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   action.Handle,
			Reserve:  reserveHandles(action),
		},
	); err != nil {
		t.Fatalf("Submit() activation error = %v", err)
	}
	if _, exists := objectIDForCard(game, source); exists {
		t.Fatal("Ally entered Field before its activation resolved")
	}
	if !containsCard(game.state.EffectSources, source) || len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Kind != effectStackAbility || game.state.EffectsStack[0].Ability == nil {
		t.Fatalf("activation state = sources %#v, stack %#v; want one Ally ability runtime activation", game.state.EffectSources, game.state.EffectsStack)
	}
	if game.state.Scheduler.OpportunityHolder != player {
		t.Fatalf("OpportunityHolder = %q, want activating player", game.state.Scheduler.OpportunityHolder)
	}

	passOpportunityRound(t, game, player)
	if _, exists := objectIDForCard(game, source); !exists {
		t.Fatal("Ally did not enter Field after its activation resolved")
	}
	if containsCard(game.state.EffectSources, source) {
		t.Fatalf("EffectSources = %#v, want resolved source removed", game.state.EffectSources)
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
	action := actionByCardName(t, view, "Straight Flare")
	if err := game.Submit(player, Input{
		Revision: view.Revision,
		Action:   action.Handle,
		Reserve:  reserveHandles(action),
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
	action := actionByCardName(t, view, "Blazing Throw")
	if err := game.Submit(player, Input{
		Revision: view.Revision,
		Action:   action.Handle,
		Reserve:  reserveHandles(action),
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

// objectIDForCard 尋找目前 Field 上代表指定 CardInstance 的 Ally object。
// 輸入為測試單局與卡牌實例；輸出為 object ID 與是否存在，無副作用。
func objectIDForCard(game *Game, card cardInstanceID) (objectID, bool) {
	for id, object := range game.state.Objects {
		if object.Card == card {
			return id, true
		}
	}
	return "", false
}

func newActionGame(t *testing.T) *Game {
	return newActionGameWithSource(t, blazingThrowCardID)
}

// reserveHandles 取出合法 action 前幾個 reserve 選項，供規則案例提交完整付款。
// 輸入為引擎提供的合法 action；輸出為恰好支付其 reserve cost 的 handles，無副作用。
func reserveHandles(action LegalAction) []ViewHandle {
	handles := make([]ViewHandle, 0, action.ReserveCost)
	for index := 0; index < action.ReserveCost; index++ {
		handles = append(handles, action.ReserveOptions[index].Handle)
	}
	return handles
}

func newActionGameWithSource(t *testing.T, definition CardID) *Game {
	return newActionGameForPlayer(t, model.PlayerOne, definition)
}

func newActionGameForPlayer(t *testing.T, player *model.Player, definition CardID) *Game {
	t.Helper()
	game, err := NewStandardGame(StandardGameConfig{
		Players: []*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: "../..",
		Seed:           42,
	})
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
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
