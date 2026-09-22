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
	game.state.Scheduler.TurnNumber = 2
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()

	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	attack := actionByKind(t, view, constants.ActionAttack)
	if attack.CardName != game.cardName(attacker.Card) {
		t.Fatalf("attack action CardName = %q, want %q", attack.CardName, game.cardName(attacker.Card))
	}
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

// TestLegalAttackersIncludesEveryObeyingPositivePowerAlly 驗證一般 Ally 不需卡牌特例即可取得 attack action，零 power Ally 則不可攻擊。
// 輸入為同一玩家控制的兩個 awake Ally objects；輸出為只包含正 power Ally 的合法攻擊者，副作用僅為重建 PlayerView action handles。
func TestLegalAttackersIncludesEveryObeyingPositivePowerAlly(t *testing.T) {
	game := newTestGame(42)
	game.state.Scheduler = schedulerFrame{
		Kind:              schedulerStable,
		TurnPlayer:        model.PlayerOne,
		Phase:             constants.PhaseMain,
		OpportunityHolder: model.PlayerOne,
		TurnNumber:        3,
	}
	game.state.Champions[model.PlayerTwo.UID] = championObject{
		ID:    "champion:player-2",
		Owner: model.PlayerTwo,
	}
	positiveCard := cardInstanceID("ally-card:positive")
	game.state.Cards[positiveCard] = cardInstance{
		ID:    positiveCard,
		Owner: model.PlayerOne,
		Types: []string{
			"ALLY",
		},
		Power: 2,
		Life:  2,
	}
	zeroCard := cardInstanceID("ally-card:zero")
	game.state.Cards[zeroCard] = cardInstance{
		ID:    zeroCard,
		Owner: model.PlayerOne,
		Types: []string{
			"ALLY",
		},
		Power: 0,
		Life:  2,
	}
	positive := objectID("ally:positive")
	zero := objectID("ally:zero")
	game.state.Objects[positive] = fieldObject{
		ID:    positive,
		Card:  positiveCard,
		Owner: model.PlayerOne,
		Types: []string{
			"ALLY",
		},
	}
	game.state.Objects[zero] = fieldObject{
		ID:    zero,
		Card:  zeroCard,
		Owner: model.PlayerOne,
		Types: []string{
			"ALLY",
		},
	}
	game.advanceKnowledgeRevision()

	attackers := game.legalAttackers(model.PlayerOne)
	if !containsObject(attackers, positive) || containsObject(attackers, zero) {
		t.Fatalf("legalAttackers() = %#v, want only positive-power Ally", attackers)
	}
}

// TestLegalAttackersExcludesFirstTurnChampion 驗證先手玩家首回合不能宣告攻擊。
// 輸入為首回合 Main Phase 的正 power、醒著 Champion；輸出為沒有 Attack action，副作用僅為重建 PlayerView action handles。
func TestLegalAttackersExcludesFirstTurnChampion(t *testing.T) {
	game := newTestGame(42)
	firstChampionCard := cardInstanceID("champion-card:player-1")
	secondChampionCard := cardInstanceID("champion-card:player-2")
	game.state.Cards[firstChampionCard] = cardInstance{
		ID:    firstChampionCard,
		Owner: model.PlayerOne,
		Power: 3,
		Life:  15,
	}
	game.state.Cards[secondChampionCard] = cardInstance{
		ID:    secondChampionCard,
		Owner: model.PlayerTwo,
		Power: 0,
		Life:  15,
	}
	game.state.Champions[model.PlayerOne.UID] = championObject{
		ID:    "champion:player-1",
		Card:  firstChampionCard,
		Owner: model.PlayerOne,
	}
	game.state.Champions[model.PlayerTwo.UID] = championObject{
		ID:    "champion:player-2",
		Card:  secondChampionCard,
		Owner: model.PlayerTwo,
	}
	game.state.Scheduler = schedulerFrame{
		Kind:              schedulerStable,
		TurnPlayer:        model.PlayerOne,
		Phase:             constants.PhaseMain,
		OpportunityHolder: model.PlayerOne,
		TurnNumber:        1,
	}
	game.advanceKnowledgeRevision()

	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	for _, action := range view.LegalActions {
		if action.Kind == constants.ActionAttack {
			t.Fatalf("PlayerView().LegalActions = %#v, want no first-turn attack", view.LegalActions)
		}
	}
}

func TestCombatStateBasedCheckEndsGameWhenChampionIsDefeated(t *testing.T) {
	game := newActionGame(t)
	attacker := game.state.Champions[model.PlayerOne.UID]
	target := game.state.Champions[model.PlayerTwo.UID]
	game.state.Cards[attacker.Card] = withCombatStats(game.state.Cards[attacker.Card], 5, 10)
	game.state.Cards[target.Card] = withCombatStats(game.state.Cards[target.Card], 1, 3)
	game.state.Scheduler.TurnNumber = 2
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

// TestWieldIsUnavailableOutsideMainPhase 驗證以武器宣告攻擊遵守 slow action timing。
// 輸入為 End Phase 且持有 Opportunity 的玩家與可用武器；輸出為 PlayerView 不含 Wield，副作用僅為建立測試用場上武器並重建 handles。
func TestWieldIsUnavailableOutsideMainPhase(t *testing.T) {
	game := newActionGame(t)
	weaponCard := findCard(t, game, model.PlayerOne, impactHammerCardID)
	weapon := objectID("weapon:end-phase")
	game.state.Objects[weapon] = fieldObject{
		ID:    weapon,
		Card:  weaponCard,
		Owner: model.PlayerOne,
		Types: []string{
			"WEAPON",
		},
	}
	game.state.Scheduler.Phase = constants.PhaseEnd
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	for _, action := range view.LegalActions {
		if action.Kind == constants.ActionWield {
			t.Fatalf("PlayerView().LegalActions = %#v, want no End Phase Wield", view.LegalActions)
		}
	}
}

func TestCombatRecordsSimultaneousHitAndRetaliationEvents(t *testing.T) {
	game := newActionGame(t)
	attacker := game.state.Champions[model.PlayerOne.UID]
	target := game.state.Champions[model.PlayerTwo.UID]
	game.state.Cards[attacker.Card] = withCombatStats(game.state.Cards[attacker.Card], 3, 10)
	game.state.Cards[target.Card] = withCombatStats(game.state.Cards[target.Card], 2, 10)
	game.state.Scheduler.TurnNumber = 2
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
