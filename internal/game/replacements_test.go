package game

import (
	"testing"

	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

// TestSafeguardAmuletPreventsOnlyNonCombatDamageForItsDuration 驗證 Amulet 的放逐、Champion 限制、傷害類型與到期行為。
// 輸入為固定對局與 6 點能力傷害；輸出驗證防止後為 2 點，副作用涵蓋放逐、延遲 replacement 與 event cause chain。
func TestSafeguardAmuletPreventsOnlyNonCombatDamageForItsDuration(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	champion := game.state.Champions[player.UID]
	amulet := replacementObject(t, game, player, safeguardAmuletCardID, "amulet")
	if err := game.activateSafeguardAmulet(player, amulet); err != nil {
		t.Fatalf("activateSafeguardAmulet() error = %v", err)
	}
	if _, exists := game.state.Objects[amulet]; exists {
		t.Fatal("Safeguard Amulet remained on the Field")
	}
	source := findCard(t, game, player, blazingThrowCardID)
	if !game.damageNonCombat(champion.ID, 6, source, nil) {
		t.Fatal("damageNonCombat() did not complete")
	}
	if got := game.state.Champions[player.UID].Damage; got != 2 {
		t.Fatalf("non-combat damage = %d, want 2 after prevention", got)
	}
	batch := game.state.Events[len(game.state.Events)-1]
	if len(batch.CauseChain) != 3 || batch.CauseChain[1].Kind != string(replacementDamagePrevent) {
		t.Fatalf("CauseChain = %#v, want intent, prevention, and committed damage", batch.CauseChain)
	}
	game.damageUnit(champion.ID, 3)
	if got := game.state.Champions[player.UID].Damage; got != 5 {
		t.Fatalf("combat damage = %d, want 5 because Safeguard excludes combat", got)
	}
	game.state.Scheduler.TurnNumber++
	if !game.damageNonCombat(champion.ID, 4, source, nil) {
		t.Fatal("expired damageNonCombat() did not complete")
	}
	if got := game.state.Champions[player.UID].Damage; got != 9 {
		t.Fatalf("expired prevention damage = %d, want 9", got)
	}
}

// TestSafeguardAmuletAbilityAndReplacementCauseChainReplay 驗證公開 Ability path 產生的 delayed prevention 與 cause chain 可由 replay 重建。
// 輸入為 Player One 的 Amulet activation 與 Player Two 的 Fiery Interference；輸出驗證傷害全被防止且 Replay.Verify 成功。
func TestSafeguardAmuletAbilityAndReplacementCauseChainReplay(t *testing.T) {
	game := newActionGameForPlayer(t, model.PlayerTwo, fieryInterferenceCardID)
	_ = replacementObject(t, game, model.PlayerOne, safeguardAmuletCardID, "replay-amulet")
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	amuletAction := actionByCardName(t, view, "Safeguard Amulet")
	if err := game.Submit(model.PlayerOne, Input{
		Revision: view.Revision,
		Action:   amuletAction.Handle,
	}); err != nil {
		t.Fatalf("Submit() Safeguard Amulet error = %v", err)
	}
	submitActionKind(t, game, model.PlayerOne, constants.ActionPass)
	view, err = game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	fieryAction := actionByCardName(t, view, "Fiery Interference")
	if err := game.Submit(model.PlayerTwo, Input{
		Revision: view.Revision,
		Action:   fieryAction.Handle,
	}); err != nil {
		t.Fatalf("Submit() Fiery Interference error = %v", err)
	}
	selectPendingChoiceSubject(t, game, model.PlayerTwo, entityID("champion:"+model.PlayerOne.UID))
	passOpportunityRound(t, game, model.PlayerTwo)
	if got := game.state.Champions[model.PlayerOne.UID].Damage; got != 0 {
		t.Fatalf("Safeguard prevented damage = %d, want 0", got)
	}
	batch := game.state.Events[len(game.state.Events)-2]
	if len(batch.CauseChain) != 3 || batch.CauseChain[1].Kind != string(replacementDamagePrevent) {
		t.Fatalf("CauseChain = %#v, want replayable prevention chain", batch.CauseChain)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

// TestInfernalVesselRecomputesRecoverAndDoesNotChangeRecoverCosts 驗證每次 replacement 後的數量與 recover 付款的原始數量。
// 輸入為 5、3 與 5 點 recover；輸出分別驗證減為 2、零值不 recover、付款路徑仍移除 5 點傷害。
func TestInfernalVesselRecomputesRecoverAndDoesNotChangeRecoverCosts(t *testing.T) {
	game := newActionGame(t)
	player := model.PlayerOne
	champion := game.state.Champions[player.UID]
	_ = replacementObject(t, game, player, infernalVesselCardID, "vessel")
	champion.Damage = 12
	game.state.Champions[player.UID] = champion
	if !game.recoverChampion(champion.ID, 5) {
		t.Fatal("recoverChampion() returned false for reduced positive recovery")
	}
	if got := game.state.Champions[player.UID].Damage; got != 10 {
		t.Fatalf("damage after recover 5 = %d, want 10 because Infernal reduces it to 2", got)
	}
	if game.recoverChampion(champion.ID, 3) {
		t.Fatal("recoverChampion() accepted recovery reduced to zero")
	}
	if got := game.state.Champions[player.UID].Damage; got != 10 {
		t.Fatalf("damage after recover 3 = %d, want unchanged 10", got)
	}
	if !game.recoverChampionUnreplaced(champion.ID, 5) {
		t.Fatal("recoverChampionUnreplaced() rejected recover cost payment")
	}
	if got := game.state.Champions[player.UID].Damage; got != 5 {
		t.Fatalf("damage after recover cost = %d, want 5 without Infernal replacement", got)
	}
}

// TestReplacementChoiceUsesAffectedControllerAndResumesAfterRecalculation 驗證兩個 Vessel 時由 recover 玩家選順序並逐次重算。
// 輸入為雙方各一個 Vessel 與 10 點 recover；輸出為 Player One 的 PendingChoice，副作用是最後只 recover 4 點。
func TestReplacementChoiceUsesAffectedControllerAndResumesAfterRecalculation(t *testing.T) {
	game := newActionGame(t)
	champion := game.state.Champions[model.PlayerOne.UID]
	champion.Damage = 10
	game.state.Champions[model.PlayerOne.UID] = champion
	_ = replacementObject(t, game, model.PlayerOne, infernalVesselCardID, "first-vessel")
	_ = replacementObject(t, game, model.PlayerTwo, infernalVesselCardID, "second-vessel")
	source := findCard(
		t,
		game,
		model.PlayerOne,
		blazingThrowCardID,
	)
	continuation := game.newAbilityInstance(
		model.PlayerOne,
		source,
		"",
		[]effectOperation{},
	)
	intent := replacementIntent{
		Kind:     replacementIntentRecover,
		Target:   champion.ID,
		Affected: model.PlayerOne,
		Amount:   10,
		Cause:    "test",
		Chain: []replacementCause{
			{
				Kind:   "recover-intent",
				Amount: 10,
			},
		},
	}
	if game.applyReplacementIntent(intent, &continuation) {
		t.Fatal("applyReplacementIntent() completed instead of requesting replacement order")
	}
	choice := game.pendingChoice(model.PlayerOne)
	if choice == nil || len(choice.Options) != 2 {
		t.Fatalf("Player One PendingChoice = %#v, want two Vessel choices", choice)
	}
	if otherChoice := game.pendingChoice(model.PlayerTwo); otherChoice != nil {
		t.Fatalf("Player Two PendingChoice = %#v, want nil", otherChoice)
	}
	if err := game.Submit(model.PlayerOne, Input{
		Revision: game.state.Revision,
		Choice:   choice.Options[0],
	}); err != nil {
		t.Fatalf("Submit() replacement choice error = %v", err)
	}
	if got := game.state.Champions[model.PlayerOne.UID].Damage; got != 6 {
		t.Fatalf("damage after two Vessel replacements = %d, want 6", got)
	}
}

// replacementObject 將指定 Regalia 放進控制者場上，供 replacement 測試使用。
// 輸入為測試、對局、擁有者、卡牌定義與穩定 suffix；輸出為新物件 ID，副作用是加入 state.Objects。
func replacementObject(t *testing.T, game *Game, player *model.Player, definition CardID, suffix string) objectID {
	t.Helper()
	card := findCard(t, game, player, definition)
	id := objectID("replacement:" + suffix)
	game.state.Objects[id] = fieldObject{
		ID:    id,
		Card:  card,
		Owner: player,
		Types: []string{
			"REGALIA",
			"ITEM",
		},
	}
	return id
}

// actionByCardName 從玩家視圖取得指定卡名的啟動 action，避免測試依賴 handle 順序。
// 輸入為測試與 PlayerView；輸出為唯一同名啟動 action，找不到時使測試失敗且不改變遊戲狀態。
func actionByCardName(t *testing.T, view PlayerView, name string) LegalAction {
	t.Helper()
	for _, action := range view.LegalActions {
		if action.Kind == constants.ActionActivate && action.CardName == name {
			return action
		}
	}
	t.Fatalf("action %q not found in %#v", name, view.LegalActions)
	return LegalAction{}
}
