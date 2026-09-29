package game

import (
	"testing"

	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

func TestCardistryMemoryPaymentIsDeclaredRecordedAndReplayed(t *testing.T) {
	game := newCardistryGame(t, wonderlandsReignCardID)
	player := model.PlayerOne
	floatingMemory := findCard(t, game, player, fiveOfSpadesCardID)
	moveCardToGraveyard(t, game, player, floatingMemory)
	game.grantCardTracking(player, entityID(floatingMemory))
	fillMemory(t, game, player, 8, floatingMemory)
	zones := game.state.Zones[player.UID]
	knownTop := -1
	for index, card := range zones.MainDeck {
		if game.state.Cards[card].Definition == twoOfHeartsCardID {
			knownTop = index
			break
		}
	}
	if knownTop < 0 {
		t.Fatal("fixture needs a Two of Hearts in Main Deck")
	}
	lastIndex := len(zones.MainDeck) - 1
	zones.MainDeck[knownTop], zones.MainDeck[lastIndex] = zones.MainDeck[lastIndex], zones.MainDeck[knownTop]
	game.state.Zones[player.UID] = zones
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()

	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	cardistryHandle := cardistryAction(t, game, player)
	for _, action := range view.LegalActions {
		if action.Handle == cardistryHandle && action.AbilitySlot != "ability:0mf1ug6yfi:front:cardistry-draw" {
			t.Fatalf("Cardistry slot = %q, want stable Wonderland's Reign slot", action.AbilitySlot)
		}
	}
	if got := cardistryAction(t, game, player); got == "" {
		t.Fatal("cardistry action = empty, want a legal action with Floating Memory")
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   cardistryAction(t, game, player),
			MemoryPayment: []ViewHandle{
				game.getPlayerKnowledge(player).Cards[entityID(floatingMemory)],
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
	stackView, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() after activation error = %v", err)
	}
	if len(stackView.EffectsStack) == 0 || stackView.EffectsStack[len(stackView.EffectsStack)-1].AbilitySlot != "ability:0mf1ug6yfi:front:cardistry-draw" {
		t.Fatalf("Effects Stack view = %#v, want Wonderland's Reign slot", stackView.EffectsStack)
	}
	opponentBefore, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("opponent PlayerView() error = %v", err)
	}
	handCount := len(stackView.Hand)
	passOpportunityRound(t, game, player)
	if !hasGameEvent(game, "draw") {
		t.Fatalf("events = %#v, want draw event", game.state.Events)
	}
	resolvedView, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("resolved PlayerView() error = %v", err)
	}
	if len(resolvedView.Hand) != handCount+1 || resolvedView.Hand[len(resolvedView.Hand)-1].Name != "Two of Hearts" {
		t.Fatalf("hand = %#v, want top card Two of Hearts drawn", resolvedView.Hand)
	}
	opponentAfter, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("opponent PlayerView() after draw error = %v", err)
	}
	for _, event := range opponentAfter.VisibleEvents[len(opponentBefore.VisibleEvents):] {
		if event.Kind == "draw" {
			t.Fatalf("opponent saw private draw: %#v", event)
		}
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

// TestThreeOfHeartsCompiledChoiceBindsSelectedCardAndReplays 驗證 Three of Hearts 由編譯定義建立 actor 專屬選牌，並以同一張 Card Instance 完成棄牌與重播。
// 輸入為固定標準對局與 Three of Hearts；輸出為已棄牌與可驗證 replay，副作用為推進對局的 Cardistry 結算。
func TestThreeOfHeartsCompiledChoiceBindsSelectedCardAndReplays(t *testing.T) {
	game := newCardistryGame(t, threeOfHeartsCardID)
	player := model.PlayerOne
	source := cardistrySource(t, game, player)
	if _, compiled := game.compiledCardistry(game.state.Objects[source].Card); !compiled {
		t.Fatal("Three of Hearts Cardistry is not compiled")
	}
	fillMemory(t, game, player, 3, "")
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()

	if err := game.Submit(
		player,
		Input{
			Revision: game.state.Revision,
			Action:   cardistryAction(t, game, player),
		},
	); err != nil {
		t.Fatalf("Submit() Cardistry error = %v", err)
	}
	passOpportunityRound(t, game, player)

	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if view.PendingChoice == nil || len(view.PendingChoice.Options) == 0 {
		t.Fatalf("PendingChoice = %#v, want card choice", view.PendingChoice)
	}
	if opponentView, err := game.PlayerView(model.PlayerTwo); err != nil || opponentView.PendingChoice != nil {
		t.Fatalf("opponent PendingChoice = %#v, error = %v; want nil, nil", opponentView.PendingChoice, err)
	}
	choice := view.PendingChoice.Options[0]
	selected := game.state.Knowledge.Choice.Options[choice]
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Choice:   choice,
		},
	); err != nil {
		t.Fatalf("Submit() choice error = %v", err)
	}
	if len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Ability == nil || game.state.EffectsStack[0].Ability.Bindings["discard-card"] != selected {
		t.Fatalf("resumed ability = %#v, want selected card bound by name", game.state.EffectsStack)
	}
	passOpportunityRound(t, game, player)
	if cardIndex(game.state.Zones[player.UID].Graveyard, cardInstanceID(selected)) < 0 {
		t.Fatalf("selected card %q was not discarded", selected)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

// TestThreeOfHeartsRejectedChoicePreservesState 驗證選牌 handle 過期、跨玩家、偽造或失效時，Submit 不改變任何權威狀態。
// 輸入為已暫停於 Three of Hearts 選牌的對局與各種無效輸入；輸出為拒絕錯誤，副作用為零。
func TestThreeOfHeartsRejectedChoicePreservesState(t *testing.T) {
	newPendingChoice := func(t *testing.T) (*Game, ViewHandle) {
		t.Helper()
		game := newCardistryGame(t, threeOfHeartsCardID)
		fillMemory(t, game, model.PlayerOne, 3, "")
		game.advanceKnowledgeRevision()
		if err := game.Submit(
			model.PlayerOne,
			Input{
				Revision: game.state.Revision,
				Action:   cardistryAction(t, game, model.PlayerOne),
			},
		); err != nil {
			t.Fatalf("Submit() Cardistry error = %v", err)
		}
		passOpportunityRound(t, game, model.PlayerOne)
		view, err := game.PlayerView(model.PlayerOne)
		if err != nil {
			t.Fatalf("PlayerView() error = %v", err)
		}
		if view.PendingChoice == nil || len(view.PendingChoice.Options) == 0 {
			t.Fatalf("PendingChoice = %#v, want card choice", view.PendingChoice)
		}
		return game, view.PendingChoice.Options[0]
	}

	testCases := []struct {
		name  string
		input func(*Game, ViewHandle) Input
	}{
		{
			name: "forged handle",
			input: func(game *Game, _ ViewHandle) Input {
				return Input{
					Revision: game.state.Revision,
					Choice:   "forged",
				}
			},
		},
		{
			name: "cross player handle",
			input: func(game *Game, choice ViewHandle) Input {
				return Input{
					Revision: game.state.Revision,
					Choice:   choice,
				}
			},
		},
		{
			name: "stale revision",
			input: func(game *Game, choice ViewHandle) Input {
				return Input{
					Revision: game.state.Revision - 1,
					Choice:   choice,
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				game, choice := newPendingChoice(t)
				player := model.PlayerOne
				if testCase.name == "cross player handle" {
					player = model.PlayerTwo
				}
				before := game.StateHash()
				if err := game.Submit(player, testCase.input(game, choice)); err == nil {
					t.Fatal("Submit() error = nil, want rejection")
				}
				if got := game.StateHash(); got != before {
					t.Fatalf("StateHash() = %q, want unchanged %q", got, before)
				}
			},
		)
	}

	t.Run(
		"selected card left zone",
		func(t *testing.T) {
			game, choice := newPendingChoice(t)
			selected := cardInstanceID(game.state.Knowledge.Choice.Options[choice])
			moveCardToGraveyard(t, game, model.PlayerOne, selected)
			before := game.StateHash()
			if err := game.Submit(
				model.PlayerOne,
				Input{
					Revision: game.state.Revision,
					Choice:   choice,
				},
			); err == nil {
				t.Fatal("Submit() error = nil, want rejection")
			}
			if got := game.StateHash(); got != before {
				t.Fatalf("StateHash() = %q, want unchanged %q", got, before)
			}
		},
	)
}

// TestThreeOfHeartsRequiresCompiledDefinition 驗證 Three of Hearts 缺少已編譯定義時不可啟動，不會退回舊的 Card ID 分支。
// 輸入為移除 Three of Hearts 定義的標準對局；輸出為不可啟動結果，副作用為零。
func TestThreeOfHeartsRequiresCompiledDefinition(t *testing.T) {
	game := newCardistryGame(t, threeOfHeartsCardID)
	player := model.PlayerOne
	source := cardistrySource(t, game, player)
	fillMemory(t, game, player, 3, "")
	delete(game.definitions, threeOfHeartsCardID)
	if game.canActivateCardistry(player, source) {
		t.Fatal("canActivateCardistry() = true without compiled Three of Hearts definition")
	}
}

func TestCardistryInvalidMemoryPaymentDoesNotChangeState(t *testing.T) {
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
			game.getPlayerKnowledge(player).Cards[entityID(notFloatingMemory)],
		},
	)
	if err == nil {
		t.Fatal("activateCardistry() error = nil, want invalid Floating Memory")
	}
	if got := game.StateHash(); got != before {
		t.Fatalf("StateHash() = %q, want unchanged %q", got, before)
	}
}

// TestCardistryStaleDeclarationDoesNotChangeState 驗證 Player View 過期後的 Cardistry 宣告會在付款與建 instance 前遭拒絕。
// 輸入為已過期 revision 的 Cardistry handle；輸出為提交錯誤與相同 state hash，副作用為零。
func TestCardistryStaleDeclarationDoesNotChangeState(t *testing.T) {
	game := newCardistryGame(t, wonderlandsReignCardID)
	player := model.PlayerOne
	fillMemory(t, game, player, 10, "")
	game.advanceKnowledgeRevision()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	action := cardistryAction(
		t,
		game,
		player,
	)
	before := game.StateHash()
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision - 1,
			Action:   action,
		},
	); err == nil {
		t.Fatal("Submit() error = nil, want stale revision rejection")
	}
	if got := game.StateHash(); got != before {
		t.Fatalf("StateHash() = %q, want unchanged %q", got, before)
	}
}

// TestCardistryCreatesIndependentInstanceForReenteredObject 驗證同一卡牌以新 Object 進場後有獨立 usage 與 Ability Instance。
// 輸入為已成功啟動並重新進場的 Wonderland's Reign；輸出為不同 instance ID 與相同 Card Instance LKI，副作用為兩次成功的 Cardistry 宣告。
func TestCardistryCreatesIndependentInstanceForReenteredObject(t *testing.T) {
	game := newCardistryGame(t, wonderlandsReignCardID)
	player := model.PlayerOne
	source := cardistrySource(t, game, player)
	card := game.state.Objects[source].Card
	fillMemory(t, game, player, 20, "")
	game.advanceKnowledgeRevision()

	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	firstAction := cardistryAction(
		t,
		game,
		player,
	)
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   firstAction,
		},
	); err != nil {
		t.Fatalf("first Submit() error = %v", err)
	}
	if len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Ability == nil {
		t.Fatalf("EffectsStack = %#v, want first Ability Instance", game.state.EffectsStack)
	}
	first := *game.state.EffectsStack[0].Ability
	if first.SourceLKI != card {
		t.Fatalf("first SourceLKI = %q, want %q", first.SourceLKI, card)
	}
	passOpportunityRound(t, game, player)

	delete(game.state.Objects, source)
	returned := objectID("cardistry:returned")
	game.state.Objects[returned] = fieldObject{
		ID:    returned,
		Card:  card,
		Owner: player,
		Types: game.state.Cards[card].Types,
	}
	game.advanceKnowledgeRevision()
	if !game.state.CardistryUsed[source] || game.state.CardistryUsed[returned] {
		t.Fatalf("CardistryUsed = %#v, want only departed Object marked used", game.state.CardistryUsed)
	}
	game.captureReplayInitialState()
	view, err = game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() after reentry error = %v", err)
	}
	var returnedAction ViewHandle
	for _, action := range view.LegalActions {
		if action.Kind == constants.ActionActivate && action.AbilitySlot == "ability:0mf1ug6yfi:front:cardistry-draw" {
			returnedAction = action.Handle
			break
		}
	}
	if returnedAction == "" {
		t.Fatal("PlayerView() has no Cardistry action for reentered object")
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   returnedAction,
		},
	); err != nil {
		t.Fatalf("second Submit() error = %v", err)
	}
	if len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Ability == nil {
		t.Fatalf("EffectsStack = %#v, want second Ability Instance", game.state.EffectsStack)
	}
	second := game.state.EffectsStack[0].Ability
	if second.ID == first.ID || second.SourceLKI != card {
		t.Fatalf("second Ability Instance = %#v, want new ID and SourceLKI %q", second, card)
	}
	passOpportunityRound(t, game, player)
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestCardistryRejectedActivationsDoNotChangeState(t *testing.T) {
	t.Run(
		"Floating Memory is only accepted with Cardistry",
		func(t *testing.T) {
			game := newCardistryGame(t, twoOfSpadesCardID)
			player := model.PlayerOne
			var pass ViewHandle
			for handle, kind := range game.getPlayerKnowledge(player).Actions {
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
					MemoryPayment: []ViewHandle{
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

// TestCardistryDrawFromEmptyDeckLosesGame 驗證能力從空牌組抽牌時，玩家依抽空規則敗北。
// 輸入為空主牌組的 Wonderland's Reign 能力；輸出為對手勝利且沒有虛構抽牌事件，副作用為結束對局。
func TestCardistryDrawFromEmptyDeckLosesGame(t *testing.T) {
	game := newCardistryGame(t, wonderlandsReignCardID)
	player := model.PlayerOne
	fillMemory(t, game, player, 10, "")
	zones := game.state.Zones[player.UID]
	zones.MainDeck = nil
	game.state.Zones[player.UID] = zones
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	var action ViewHandle
	for _, candidate := range view.LegalActions {
		if candidate.AbilitySlot == "ability:0mf1ug6yfi:front:cardistry-draw" {
			action = candidate.Handle
			break
		}
	}
	if action == "" {
		t.Fatal("PlayerView has no Wonderland's Reign action")
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   action,
		},
	); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	passOpportunityRound(t, game, player)
	result, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() after resolution error = %v", err)
	}
	if !result.Finished || result.Winner == nil || result.Winner.UID != model.PlayerTwo.UID {
		t.Fatalf("finished = %v, winner = %v; want player two win", result.Finished, result.Winner)
	}
	for _, event := range result.VisibleEvents[len(view.VisibleEvents):] {
		if event.Kind == "draw" {
			t.Fatalf("empty deck produced draw event %#v", event)
		}
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
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

// TestFourOfSpadesDrawsToMemoryAndReplays 驗證 Four of Spades 由定義抽牌至 Memory，並維持卡牌守恆與 replay。
// 輸入為具足額 Memory 付款的 Four of Spades；輸出為一張新 Memory 卡與可驗證 replay，副作用為放逐付款卡並記錄抽至 Memory 事件。
func TestFourOfSpadesDrawsToMemoryAndReplays(t *testing.T) {
	game := newCardistryGame(t, fourOfSpadesCardID)
	player := model.PlayerOne
	fillMemory(t, game, player, 4, "")
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()
	before := game.state.Zones[player.UID]
	cost := game.cardistryCost(
		player,
		4,
		reductionDistinctSuitedCosts,
	)
	if err := game.Submit(
		player,
		Input{
			Revision: game.state.Revision,
			Action:   cardistryAction(t, game, player),
		},
	); err != nil {
		t.Fatalf("Submit() Cardistry error = %v", err)
	}
	passOpportunityRound(t, game, player)
	after := game.state.Zones[player.UID]
	if len(after.Memory) != len(before.Memory)-cost+1 || len(after.Banishment) != len(before.Banishment)+cost {
		t.Fatalf("zones after Four of Spades = %#v, want %d payments and one Memory draw", after, cost)
	}
	if !hasGameEvent(game, "draw-to-memory") {
		t.Fatalf("events = %#v, want draw-to-memory", game.state.Events)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

// TestDuchessCopySelectionIncludesOnlyQualifiedFireActions 驗證 Duchess 的定義選牌只提供合格墓地 Action。
// 輸入為墓地中的合格與不合格卡牌；輸出為只含 Fire、Action 且 Reserve Cost 不超過二的候選，副作用為零。
func TestDuchessCopySelectionIncludesOnlyQualifiedFireActions(t *testing.T) {
	game := newCardistryGame(t, duchessCardID)
	player := model.PlayerOne
	qualified := []CardID{
		blazingThrowCardID,
		fieryInterferenceCardID,
		straightFlareCardID,
	}
	qualifiedCards := make([]cardInstanceID, 0, len(qualified))
	for _, definition := range qualified {
		card := findCard(t, game, player, definition)
		moveCardToGraveyard(t, game, player, card)
		qualifiedCards = append(qualifiedCards, card)
	}
	notQualified := findCard(t, game, player, duchessCardID)
	moveCardToGraveyard(t, game, player, notQualified)

	duchess := findCard(
		t,
		game,
		player,
		duchessCardID,
	)
	ability, exists := game.compiledCardistry(duchess)
	if !exists {
		t.Fatal("compiledCardistry() = false, want Duchess definition")
	}
	got := game.cardSelectionCandidates(
		player,
		ability.operations()[0].CardSelection,
		"",
	)
	for _, card := range qualifiedCards {
		if !containsCard(got, card) {
			t.Fatalf("eligible copies = %#v, missing %q", got, card)
		}
	}
	if containsCard(got, notQualified) {
		t.Fatalf("eligible copies = %#v, unexpectedly included %q", got, notQualified)
	}
}

func TestDuchessCopyCanBeDeclinedAndRuntimeCopyIsDestroyed(t *testing.T) {
	game := newCardistryGame(t, duchessCardID)
	player := model.PlayerOne
	source := findCard(t, game, player, blazingThrowCardID)
	moveCardToGraveyard(t, game, player, source)
	game.grantCardTracking(player, entityID(source))
	fillMemory(t, game, player, 6, source)
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()

	if err := game.Submit(
		player,
		Input{
			Revision: game.state.Revision,
			Action:   cardistryAction(t, game, player),
		},
	); err != nil {
		t.Fatalf("Submit() Cardistry error = %v", err)
	}
	passOpportunityRound(t, game, player)
	selectPendingChoiceSubject(t, game, player, entityID(source))
	passOpportunityRound(t, game, player)
	if len(game.state.EffectsStack) != 1 || game.state.EffectsStack[0].Ability == nil {
		t.Fatalf("EffectsStack = %#v, want copied Action ability", game.state.EffectsStack)
	}
	copied := game.state.EffectsStack[0].Ability
	if !copied.RuntimeCopy || copied.Source == source || copied.SourceLKI != copied.Source {
		t.Fatalf("copied ability = %#v, want independent runtime copy source and LKI", copied)
	}
	passOpportunityRound(t, game, player)

	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if view.PendingChoice == nil || !view.PendingChoice.CanPass {
		t.Fatalf("PendingChoice = %#v, resolution frame = %#v, stack = %#v, want optional copy activation", view.PendingChoice, game.state.ResolutionFrame, game.state.EffectsStack)
	}
	var pass LegalAction
	for _, action := range view.LegalActions {
		if action.Kind == constants.ActionPass {
			pass = action
			break
		}
	}
	if pass.Handle == "" {
		t.Fatalf("LegalActions = %#v, want pass for declining copy", view.LegalActions)
	}
	if err := game.Submit(
		player,
		Input{
			Revision: view.Revision,
			Action:   pass.Handle,
		},
	); err != nil {
		t.Fatalf("Submit() decline error = %v", err)
	}
	if len(game.state.EffectsStack) != 0 || game.state.ResolutionFrame != nil {
		t.Fatalf("copy state = stack %#v, frame %#v, want no pending copy", game.state.EffectsStack, game.state.ResolutionFrame)
	}
	if cardIndex(game.state.Zones[player.UID].Banishment, source) < 0 {
		t.Fatalf("source zones = %#v, want source banished", game.state.Zones[player.UID])
	}
	for card := range game.state.Cards {
		if len(card) >= 5 && card[:5] == "copy:" {
			t.Fatalf("runtime copy %q remains after decline", card)
		}
	}
	if game.state.Champions[model.PlayerTwo.UID].Damage != 0 {
		t.Fatalf("opponent damage = %d, want no damage after decline", game.state.Champions[model.PlayerTwo.UID].Damage)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestDuchessCopyActivatesForFreeAndResolvesItsCopiedFace(t *testing.T) {
	game := newCardistryGame(t, duchessCardID)
	player := model.PlayerOne
	source := findCard(t, game, player, blazingThrowCardID)
	moveCardToGraveyard(t, game, player, source)
	game.grantCardTracking(player, entityID(source))
	fillMemory(t, game, player, 6, source)
	game.advanceKnowledgeRevision()
	game.captureReplayInitialState()

	if err := game.Submit(
		player,
		Input{
			Revision: game.state.Revision,
			Action:   cardistryAction(t, game, player),
		},
	); err != nil {
		t.Fatalf("Submit() Cardistry error = %v", err)
	}
	passOpportunityRound(t, game, player)
	selectPendingChoiceSubject(t, game, player, entityID(source))
	passOpportunityRound(t, game, player)
	passOpportunityRound(t, game, player)
	selectPendingChoiceSubject(t, game, player, entityID("champion:"+model.PlayerTwo.UID))
	passOpportunityRound(t, game, player)

	if got := game.state.Champions[model.PlayerTwo.UID].Damage; got != 4 {
		t.Fatalf("opponent damage = %d, want 4 from copied Blazing Throw", got)
	}
	for card := range game.state.Cards {
		if len(card) >= 5 && card[:5] == "copy:" {
			t.Fatalf("runtime copy %q remains after resolution", card)
		}
	}
	if cardIndex(game.state.Zones[player.UID].Banishment, source) < 0 {
		t.Fatalf("source zones = %#v, want source banished", game.state.Zones[player.UID])
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

// TestDuchessCopiedActionFizzlesAfterTargetLeaves 驗證 Duchess 複製的 Action 在已選目標離場後依正式介面 fizzle。
// 輸入為墓地 Blazing Throw、可被 Fast Action 移除的 Ally 與 Cardistry 付款；輸出驗證原牌放逐、費用保留、事件、清理、hash、replay 與卡牌守恆，副作用為完成一段測試對局。
func TestDuchessCopiedActionFizzlesAfterTargetLeaves(t *testing.T) {
	game := newCardistryGame(t, duchessCardID)
	player := model.PlayerOne
	opponent := model.PlayerTwo
	source := findCard(t, game, player, blazingThrowCardID)
	moveCardToGraveyard(t, game, player, source)
	game.grantCardTracking(player, entityID(source))

	zones := game.state.Zones[player.UID]
	targetCard := cardInstanceID("")
	for index, card := range zones.MainDeck {
		if game.state.Cards[card].Definition != redHareCardID {
			continue
		}
		targetCard = card
		zones.MainDeck = removeCardAt(zones.MainDeck, index)
		break
	}
	if targetCard == "" {
		t.Fatal("fixture has no Red Hare")
	}
	game.state.Zones[player.UID] = zones
	target := objectID("ally:duchess-copy-target")
	game.state.Objects[target] = fieldObject{
		ID:    target,
		Card:  targetCard,
		Owner: player,
		Types: []string{
			"ALLY",
		},
	}
	targetObject := game.state.Objects[target]
	targetObject.Damage = game.characteristicsFor(target).Life - 2
	game.state.Objects[target] = targetObject

	zones = game.state.Zones[opponent.UID]
	responseCard := cardInstanceID("")
	for index, card := range zones.MainDeck {
		if game.state.Cards[card].Definition != fieryInterferenceCardID {
			continue
		}
		responseCard = card
		zones.MainDeck = removeCardAt(zones.MainDeck, index)
		zones.Hand = append(zones.Hand, card)
		break
	}
	if responseCard == "" {
		t.Fatal("fixture has no Fiery Interference")
	}
	for len(zones.Memory) < game.state.Cards[responseCard].ReserveCost {
		payment := zones.Hand[0]
		if payment == responseCard {
			payment = zones.Hand[1]
		}
		zones.Hand = removeCardAt(
			zones.Hand,
			cardIndex(zones.Hand, payment),
		)
		zones.Memory = append(zones.Memory, payment)
	}
	game.state.Zones[opponent.UID] = zones
	fillMemory(t, game, player, 6, source)
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
	selectVisibleTargetByName(t, game, player, "Blazing Throw")
	passOpportunityRound(t, game, player)
	passOpportunityRound(t, game, player)
	selectVisibleTargetByName(t, game, player, "Red Hare, Unrivaled Stallion")

	submitActionKind(t, game, player, constants.ActionPass)
	responseView, err := game.PlayerView(opponent)
	if err != nil {
		t.Fatalf("PlayerView() for response error = %v", err)
	}
	response := actionByCardName(t, responseView, "Fiery Interference")
	if err := game.Submit(
		opponent,
		Input{
			Revision: responseView.Revision,
			Action:   response.Handle,
			Reserve:  reserveHandles(response),
		},
	); err != nil {
		t.Fatalf("Submit() Fiery Interference error = %v", err)
	}
	selectVisibleTargetByName(t, game, opponent, "Red Hare, Unrivaled Stallion")
	for {
		stackView, viewErr := game.PlayerView(player)
		if viewErr != nil {
			t.Fatalf("PlayerView() during resolution error = %v", viewErr)
		}
		if len(stackView.EffectsStack) == 0 {
			break
		}
		if stackView.OpportunityHolder == nil {
			t.Fatal("EffectsStack is nonempty without an opportunity holder")
		}
		submitActionKind(t, game, stackView.OpportunityHolder, constants.ActionPass)
	}

	finalView, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() after fizzle error = %v", err)
	}
	if finalView.PendingChoice != nil || len(finalView.EffectsStack) != 0 {
		t.Fatalf("copy state = choice %#v, stack %#v; want cleaned runtime copy", finalView.PendingChoice, finalView.EffectsStack)
	}
	for _, object := range finalView.Field {
		if object.CardName == "Red Hare, Unrivaled Stallion" {
			t.Fatalf("target remains after Fiery Interference: %#v", object)
		}
	}
	if got := championByOwner(t, finalView, opponent).Damage; got != 0 {
		t.Fatalf("opponent damage = %d, want copied Action to fizzle", got)
	}
	banishedMemory := 0
	for _, event := range finalView.VisibleEvents {
		if event.Kind == "banish-memory" {
			banishedMemory++
		}
	}
	if banishedMemory != 5 || !hasVisibleEvent(finalView.VisibleEvents, "banish", "Blazing Throw") || !hasVisibleEvent(finalView.VisibleEvents, "ability-activated", "Duchess, Six of Hearts") {
		t.Fatalf("events = %#v, want five retained payments, copied source banishment, and activation", finalView.VisibleEvents)
	}
	replay := game.Replay()
	if len(replay.Steps) == 0 || replay.Steps[len(replay.Steps)-1].StateHash != game.StateHash() {
		t.Fatalf("final replay hash = %#v, want %q", replay.Steps, game.StateHash())
	}
	const wantStateHash = "7d74349c3f56138605f61b0d051cabaaaf92bb5587bee6ad4b67986e951a8ca9"
	if got := game.StateHash(); got != wantStateHash {
		t.Fatalf("StateHash() = %q, want conserved Duchess fizzle state %q", got, wantStateHash)
	}
	if err := replay.Verify(); err != nil {
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
	baseLife := game.characteristicsFor(target).Life
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
	if len(game.state.ContinuousEffects) != 1 || game.state.ContinuousEffects[0].ExpiresAtTurn != game.state.Scheduler.TurnNumber+1 {
		t.Fatalf("continuous effects = %#v, want modifier through the current turn", game.state.ContinuousEffects)
	}
	if err := game.Replay().Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
	game.state.Scheduler.TurnNumber = game.state.ContinuousEffects[0].ExpiresAtTurn
	if got := game.characteristicsFor(target).Life; got != baseLife {
		t.Fatalf("target life = %d, want expired base life %d", got, baseLife)
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
	game.state.Scheduler.Phase = constants.PhaseMain
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
	for handle, ability := range game.getPlayerKnowledge(player).Abilities {
		if ability.Kind == activatedAbilityCardistry && ability.Source == objectID("cardistry:source") {
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
