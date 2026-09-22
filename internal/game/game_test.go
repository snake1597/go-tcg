package game

import (
	"encoding/json"
	"errors"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
	"reflect"
	"strings"
	"testing"
)

func newTestGame(seed uint64) *Game {
	return NewGame(
		StandardGameConfig{
			Players: []*model.Player{
				model.PlayerOne,
				model.PlayerTwo,
			},
			Seed: seed,
		},
	)
}

func TestNewGamePinsReplayVersionsAndSeed(t *testing.T) {
	const seed uint64 = 42

	game := newTestGame(seed)
	replay := game.Replay()

	wantVersions := Versions{
		Engine:   "grand-archive-v1",
		Rules:    "602c917f2f8fd4df7198429a72eb596bf7f647c6",
		CardData: "card-data-v3",
		Deck:     "standard-fire-v2",
		PRNG:     "splitmix64-v1",
	}
	if replay.FormatVersion != 5 {
		t.Fatalf("Replay().FormatVersion = %d, want 5", replay.FormatVersion)
	}
	if constants.CanonicalStateSchemaVersion != 5 {
		t.Fatalf("CanonicalStateSchemaVersion = %d, want 5", constants.CanonicalStateSchemaVersion)
	}
	if replay.Versions != wantVersions {
		t.Fatalf("Replay().Versions = %#v, want %#v", replay.Versions, wantVersions)
	}
	if replay.InitialSeed != seed {
		t.Fatalf("Replay().InitialSeed = %d, want %d", replay.InitialSeed, seed)
	}
}

func TestPlayerViewScopesOpaqueActionHandles(t *testing.T) {
	game := newTestGame(42)

	firstView, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	secondView, err := game.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	firstActionCount := len(firstView.LegalActions)
	secondActionCount := len(secondView.LegalActions)
	if firstActionCount != 1 || secondActionCount != 1 {
		t.Fatalf("legal action counts = %d and %d, want one each", firstActionCount, secondActionCount)
	}
	firstAction := firstView.LegalActions[0]
	secondAction := secondView.LegalActions[0]
	if firstAction.Kind != constants.ActionConcede || secondAction.Kind != constants.ActionConcede {
		t.Fatalf("legal actions = %#v and %#v, want concede", firstView.LegalActions, secondView.LegalActions)
	}
	if firstAction.Handle == secondAction.Handle {
		t.Fatalf("players received the same action handle %q", firstAction.Handle)
	}
	viewJSON, err := json.Marshal(firstView)
	if err != nil {
		t.Fatalf("marshal player view: %v", err)
	}
	forbiddenValues := []string{
		"CardInstanceID",
		"ObjectID",
		"player-2",
	}
	for _, forbidden := range forbiddenValues {
		viewText := string(viewJSON)
		if strings.Contains(viewText, forbidden) {
			t.Fatalf("PlayerView() exposed %q in %s", forbidden, viewJSON)
		}
	}
	_, err = game.PlayerView(
		&model.Player{
			UID: "intruder",
		},
	)
	if !errors.Is(err, tcgErrors.ErrUnknownPlayer) {
		t.Fatalf("PlayerView() error = %v, want unknown player", err)
	}
}

// TestPlayerViewIncludesHeuristicRank 驗證 PlayerView 對每個合法 action 提供 bot 可用的固定優先級。
// 輸入為新建對局與 Player One；輸出為投降 action 的最低優先級 rank，副作用是無。
func TestPlayerViewIncludesHeuristicRank(t *testing.T) {
	game := newTestGame(1)
	view, err := game.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if len(view.LegalActions) != 1 {
		t.Fatalf("LegalActions = %#v, want one action", view.LegalActions)
	}
	if view.LegalActions[0].Kind != constants.ActionConcede || view.LegalActions[0].HeuristicRank != 6 {
		t.Fatalf("LegalAction = %#v, want concede with rank 6", view.LegalActions[0])
	}
}

func TestSubmitRejectsInvalidActionHandleWithoutChangingGame(t *testing.T) {
	testCases := []struct {
		name       string
		player     *model.Player
		input      func(*Game) Input
		wantReason string
	}{
		{
			name:   "forged handle",
			player: model.PlayerOne,
			input: func(game *Game) Input {
				return Input{
					Revision: 1,
					Action:   ViewHandle("forged"),
				}
			},
			wantReason: "invalid view handle",
		},
		{
			name:   "cross player handle",
			player: model.PlayerTwo,
			input: func(game *Game) Input {
				return Input{
					Revision: 1,
					Action:   actionHandle(t, game, model.PlayerOne),
				}
			},
			wantReason: "invalid view handle",
		},
		{
			name:   "stale revision",
			player: model.PlayerOne,
			input: func(game *Game) Input {
				return Input{
					Revision: 0,
					Action:   actionHandle(t, game, model.PlayerOne),
				}
			},
			wantReason: "stale revision",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				game := newTestGame(42)
				beforeHash := game.StateHash()
				replayBeforeInput := game.Replay()
				beforeReplay, err := json.Marshal(replayBeforeInput)
				if err != nil {
					t.Fatalf("marshal replay before input: %v", err)
				}

				input := testCase.input(game)
				err = game.Submit(testCase.player, input)
				errorMessage := ""
				if err != nil {
					errorMessage = err.Error()
				}
				if err == nil || !strings.Contains(errorMessage, testCase.wantReason) {
					t.Fatalf("Submit() error = %v, want reason %q", err, testCase.wantReason)
				}
				afterHash := game.StateHash()
				if afterHash != beforeHash {
					t.Fatalf("StateHash() changed after rejected input")
				}
				replayAfterInput := game.Replay()
				afterReplay, err := json.Marshal(replayAfterInput)
				if err != nil {
					t.Fatalf("marshal replay after input: %v", err)
				}
				if string(afterReplay) != string(beforeReplay) {
					t.Fatalf("Replay() changed after rejected input: before %s, after %s", beforeReplay, afterReplay)
				}
			},
		)
	}
}

// TestSubmitRejectsUnusedPayloadByActionKind 驗證每類合法 handle 只接受該行動實際消費的 Input 欄位。
// 輸入為帶有 Reserve 或 MemoryPayment 多餘資料的 action／choice；輸出為 ErrInvalidViewHandle，副作用為拒絕後 state hash 與 replay 都不變。
func TestSubmitRejectsUnusedPayloadByActionKind(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(*Game) Input
	}{
		{
			name: "plain action rejects reserve",
			setup: func(game *Game) Input {
				game.state.Knowledge.Players[model.PlayerOne.UID].Actions["plain"] = constants.ActionPass
				return Input{
					Revision: game.state.Revision,
					Action:   "plain",
					Reserve: []ViewHandle{
						"unused",
					},
				}
			},
		},
		{
			name: "materialization rejects reserve",
			setup: func(game *Game) Input {
				game.state.Knowledge.Players[model.PlayerOne.UID].Materializations["materialize"] = "source"
				return Input{
					Revision: game.state.Revision,
					Action:   "materialize",
					Reserve: []ViewHandle{
						"unused",
					},
				}
			},
		},
		{
			name: "card activation rejects memory payment",
			setup: func(game *Game) Input {
				game.state.Knowledge.Players[model.PlayerOne.UID].Activations["activate"] = "source"
				return Input{
					Revision: game.state.Revision,
					Action:   "activate",
					MemoryPayment: []ViewHandle{
						"unused",
					},
				}
			},
		},
		{
			name: "attack rejects reserve",
			setup: func(game *Game) Input {
				game.state.Knowledge.Players[model.PlayerOne.UID].Attacks["attack"] = "attacker"
				return Input{
					Revision: game.state.Revision,
					Action:   "attack",
					Reserve: []ViewHandle{
						"unused",
					},
				}
			},
		},
		{
			name: "wield rejects memory payment",
			setup: func(game *Game) Input {
				game.state.Knowledge.Players[model.PlayerOne.UID].Wields["wield"] = "weapon"
				return Input{
					Revision: game.state.Revision,
					Action:   "wield",
					MemoryPayment: []ViewHandle{
						"unused",
					},
				}
			},
		},
		{
			name: "cardistry rejects reserve",
			setup: func(game *Game) Input {
				game.state.Knowledge.Players[model.PlayerOne.UID].Abilities["cardistry"] = activatedAbility{
					Kind:   activatedAbilityCardistry,
					Source: "source",
				}
				return Input{
					Revision: game.state.Revision,
					Action:   "cardistry",
					Reserve: []ViewHandle{
						"unused",
					},
				}
			},
		},
		{
			name: "object ability rejects reserve",
			setup: func(game *Game) Input {
				game.state.Knowledge.Players[model.PlayerOne.UID].Abilities["ability"] = activatedAbility{
					Kind:   activatedAbilityObject,
					Source: "source",
				}
				return Input{
					Revision: game.state.Revision,
					Action:   "ability",
					Reserve: []ViewHandle{
						"unused",
					},
				}
			},
		},
		{
			name: "choice rejects reserve",
			setup: func(game *Game) Input {
				game.state.Knowledge.Choice = &pendingChoice{
					Actor: model.PlayerOne,
					Options: map[ViewHandle]entityID{
						"choice": "subject",
					},
				}
				return Input{
					Revision: game.state.Revision,
					Choice:   "choice",
					Reserve: []ViewHandle{
						"unused",
					},
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				game := newTestGame(42)
				input := testCase.setup(game)
				beforeHash := game.StateHash()
				beforeReplay := game.Replay()

				err := game.Submit(model.PlayerOne, input)
				if err == nil {
					t.Fatal("Submit() error = nil, want unused input payload")
				}
				errorMessage := err.Error()
				if !errors.Is(err, tcgErrors.ErrInvalidViewHandle) || !strings.Contains(errorMessage, "unused input payload") {
					t.Fatalf("Submit() error = %v, want unused input payload", err)
				}
				afterHash := game.StateHash()
				if afterHash != beforeHash {
					t.Fatal("Submit() changed state hash after rejected payload")
				}
				afterReplay := game.Replay()
				if !reflect.DeepEqual(afterReplay, beforeReplay) {
					t.Fatal("Submit() changed replay after rejected payload")
				}
			},
		)
	}
}

func TestSameSeedAndInputProduceSameStateHash(t *testing.T) {
	first := newTestGame(42)
	second := newTestGame(42)
	input := concedeInput(
		t,
		first,
		model.PlayerOne,
	)

	if err := first.Submit(model.PlayerOne, input); err != nil {
		t.Fatalf("first Submit() error = %v", err)
	}
	if err := second.Submit(model.PlayerOne, input); err != nil {
		t.Fatalf("second Submit() error = %v", err)
	}

	firstHash := first.StateHash()
	secondHash := second.StateHash()
	if firstHash != secondHash {
		t.Fatalf("state hashes differ: %q != %q", firstHash, secondHash)
	}
	firstView, err := first.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("first PlayerView() error = %v", err)
	}
	secondView, err := second.PlayerView(model.PlayerTwo)
	if err != nil {
		t.Fatalf("second PlayerView() error = %v", err)
	}
	if !firstView.Finished || !secondView.Finished || firstView.Winner != model.PlayerTwo || secondView.Winner != model.PlayerTwo {
		t.Fatalf("final views = %#v and %#v, want player two to win", firstView, secondView)
	}
	firstReplay := first.Replay()
	secondReplay := second.Replay()
	firstStepCount := len(firstReplay.Steps)
	secondStepCount := len(secondReplay.Steps)
	if firstStepCount != 1 || secondStepCount != 1 {
		t.Fatalf("replay step counts = %d and %d, want 1", firstStepCount, secondStepCount)
	}
	if firstReplay.Steps[0].StateHash != secondReplay.Steps[0].StateHash {
		t.Fatalf("replay hashes differ: %q != %q", firstReplay.Steps[0].StateHash, secondReplay.Steps[0].StateHash)
	}
}

// TestSubmitRejectsActionAfterGameFinishes 驗證單局結束後不再接受先前合法的一般行動。
// 輸入為 Standard 單局結束前取得的 pass 與投降 action；輸出為 ErrGameFinished，副作用僅有投降步驟且拒絕後 state hash 與 replay 不變。
func TestSubmitRejectsActionAfterGameFinishes(t *testing.T) {
	match, err := NewStandardGame(
		StandardGameConfig{
			Players: []*model.Player{
				model.PlayerOne,
				model.PlayerTwo,
			},
			RepositoryRoot: "../..",
			Seed:           7,
		},
	)
	if err != nil {
		t.Fatalf("NewStandardGame() error = %v", err)
	}
	view, err := match.PlayerView(model.PlayerOne)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	pass := actionByKind(
		t,
		view,
		constants.ActionPass,
	)
	concede := actionByKind(
		t,
		view,
		constants.ActionConcede,
	)
	if err := match.Submit(
		model.PlayerOne,
		Input{
			Revision: view.Revision,
			Action:   concede.Handle,
		},
	); err != nil {
		t.Fatalf("Submit() concede error = %v", err)
	}
	finishedHash := match.StateHash()
	finishedReplay := match.Replay()
	finishedSteps := len(finishedReplay.Steps)
	err = match.Submit(
		model.PlayerOne,
		Input{
			Revision: view.Revision,
			Action:   pass.Handle,
		},
	)
	if !errors.Is(err, tcgErrors.ErrGameFinished) {
		t.Fatalf("Submit() after finish error = %v, want ErrGameFinished", err)
	}
	afterHash := match.StateHash()
	afterReplay := match.Replay()
	if afterHash != finishedHash || len(afterReplay.Steps) != finishedSteps {
		t.Fatal("rejected post-game action changed state hash or replay")
	}
}

func TestStateHashUsesCanonicalVersionedState(t *testing.T) {
	game := newTestGame(42)
	const want = "4af1bf6e6a83084e97ab896860d404599d09e6160c67cd71d4a0fd25a804e3d5"

	if got := game.StateHash(); got != want {
		t.Fatalf("StateHash() = %q, want canonical digest %q", got, want)
	}
	otherGame := newTestGame(43)
	if other := otherGame.StateHash(); other == want {
		t.Fatalf("StateHash() ignored the seed: seed 43 also produced %q", other)
	}
}

// TestKnowledgeStateGroupsViewsByPlayer 驗證新對局為每位玩家建立獨立的可見資訊容器。
// 輸入為固定 seed 的測試對局；輸出為每個玩家的完整 playerKnowledge；副作用僅限測試資料寫入。
func TestKnowledgeStateGroupsViewsByPlayer(t *testing.T) {
	game := newTestGame(42)
	first := game.state.Knowledge.Players[model.PlayerOne.UID]
	second := game.state.Knowledge.Players[model.PlayerTwo.UID]
	if first == nil || second == nil {
		t.Fatalf("Knowledge.Players = %#v, want both players", game.state.Knowledge.Players)
	}
	if first.Actions == nil || first.Materializations == nil || first.Activations == nil || first.Attacks == nil || first.Wields == nil || first.Abilities == nil || first.Cards == nil || first.Events == nil {
		t.Fatalf("first player knowledge = %#v, want initialized collections", first)
	}
	first.Actions["only-first"] = constants.ActionPass
	if _, exists := second.Actions["only-first"]; exists {
		t.Fatalf("second player actions = %#v, want no first player handle", second.Actions)
	}
}

func TestRejectedInputDoesNotChangeGame(t *testing.T) {
	testCases := []struct {
		name       string
		player     *model.Player
		input      func(*Game) Input
		wantReason string
	}{
		{
			name:   "stale revision",
			player: model.PlayerOne,
			input: func(game *Game) Input {
				input := concedeInput(
					t,
					game,
					model.PlayerOne,
				)
				input.Revision = 0
				return input
			},
			wantReason: "stale revision",
		},
		{
			name: "unknown player",
			player: &model.Player{
				UID: "intruder",
			},
			input: func(game *Game) Input {
				return concedeInput(
					t,
					game,
					model.PlayerOne,
				)
			},
			wantReason: "unknown player",
		},
		{
			name:   "invalid action handle",
			player: model.PlayerOne,
			input: func(game *Game) Input {
				return Input{
					Revision: 1,
					Action:   ViewHandle("unsupported"),
				}
			},
			wantReason: "invalid view handle",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				game := newTestGame(42)
				beforeView, err := game.PlayerView(model.PlayerOne)
				if err != nil {
					t.Fatalf("PlayerView() before input error = %v", err)
				}
				beforeHash := game.StateHash()
				replayBeforeInput := game.Replay()
				beforeReplay, err := json.Marshal(replayBeforeInput)
				if err != nil {
					t.Fatalf("marshal replay before input: %v", err)
				}

				input := testCase.input(game)
				err = game.Submit(testCase.player, input)
				if err == nil {
					t.Fatalf("Submit() error = nil, want reason %q", testCase.wantReason)
				}
				errorMessage := err.Error()
				if !strings.Contains(errorMessage, testCase.wantReason) {
					t.Fatalf("Submit() error = %v, want reason %q", err, testCase.wantReason)
				}

				replayAfterInput := game.Replay()
				afterReplay, err := json.Marshal(replayAfterInput)
				if err != nil {
					t.Fatalf("marshal replay after input: %v", err)
				}
				afterView, err := game.PlayerView(model.PlayerOne)
				if err != nil {
					t.Fatalf("PlayerView() after input error = %v", err)
				}
				beforeViewJSON, err := json.Marshal(beforeView)
				if err != nil {
					t.Fatalf("marshal player view before input: %v", err)
				}
				afterViewJSON, err := json.Marshal(afterView)
				if err != nil {
					t.Fatalf("marshal player view after input: %v", err)
				}
				if string(afterViewJSON) != string(beforeViewJSON) {
					t.Fatalf("PlayerView() changed after rejected input")
				}
				if game.StateHash() != beforeHash {
					t.Fatalf("StateHash() changed after rejected input")
				}
				if string(afterReplay) != string(beforeReplay) {
					t.Fatalf("Replay() changed after rejected input: before %s, after %s", beforeReplay, afterReplay)
				}
			},
		)
	}
}

func TestReplayVerifiesFromRecordedVersionsAndSeed(t *testing.T) {
	game := newTestGame(42)
	input := concedeInput(
		t,
		game,
		model.PlayerTwo,
	)
	if err := game.Submit(model.PlayerTwo, input); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	replay := game.Replay()
	if err := replay.Verify(); err != nil {
		t.Fatalf("Replay().Verify() error = %v", err)
	}
}

func TestReplayRejectsEachIncompatibleVersion(t *testing.T) {
	testCases := []struct {
		name       string
		field      string
		wantReason string
	}{
		{
			name:       "format",
			field:      "format",
			wantReason: "replay format version",
		},
		{
			name:       "engine",
			field:      "engine",
			wantReason: "engine version",
		},
		{
			name:       "rules",
			field:      "rules",
			wantReason: "rules version",
		},
		{
			name:       "card data",
			field:      "card_data",
			wantReason: "card data version",
		},
		{
			name:       "deck",
			field:      "deck",
			wantReason: "deck version",
		},
		{
			name:       "PRNG",
			field:      "prng",
			wantReason: "PRNG version",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				game := newTestGame(42)
				replay := game.Replay()
				switch testCase.field {
				case "format":
					replay.FormatVersion++
				case "engine":
					replay.Versions.Engine = "old"
				case "rules":
					replay.Versions.Rules = "old"
				case "card_data":
					replay.Versions.CardData = "old"
				case "deck":
					replay.Versions.Deck = "old"
				case "prng":
					replay.Versions.PRNG = "old"
				}

				err := replay.Verify()
				var diagnostic *ReplayError
				if !errors.As(err, &diagnostic) {
					t.Fatalf("Verify() error = %v, want *ReplayError", err)
				}
				if diagnostic.InputIndex != -1 {
					t.Fatalf("ReplayError.InputIndex = %d, want -1", diagnostic.InputIndex)
				}
				if diagnostic.Failure != constants.ReplayVersionMismatch {
					t.Fatalf("ReplayError.Failure = %q, want %q", diagnostic.Failure, constants.ReplayVersionMismatch)
				}
				errorMessage := diagnostic.Error()
				if !strings.Contains(errorMessage, testCase.wantReason) {
					t.Fatalf("Verify() error = %v, want reason %q", err, testCase.wantReason)
				}
			},
		)
	}
}

func TestReplayReportsFirstStateHashDivergence(t *testing.T) {
	game := newTestGame(42)
	input := concedeInput(
		t,
		game,
		model.PlayerOne,
	)
	if err := game.Submit(model.PlayerOne, input); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	replay := game.Replay()
	replay.Steps[0].StateHash = strings.Repeat("0", 64)

	err := replay.Verify()
	var diagnostic *ReplayError
	if !errors.As(err, &diagnostic) {
		t.Fatalf("Verify() error = %v, want *ReplayError", err)
	}
	if diagnostic.InputIndex != 0 {
		t.Fatalf("ReplayError.InputIndex = %d, want 0", diagnostic.InputIndex)
	}
	if diagnostic.Failure != constants.ReplayStateHashMismatch {
		t.Fatalf("ReplayError.Failure = %q, want %q", diagnostic.Failure, constants.ReplayStateHashMismatch)
	}
	errorMessage := diagnostic.Error()
	if !strings.Contains(errorMessage, "state hash mismatch") {
		t.Fatalf("ReplayError.Error() = %q, want readable hash mismatch reason", errorMessage)
	}
}

func actionHandle(t *testing.T, game *Game, player *model.Player) ViewHandle {
	t.Helper()
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	if len(view.LegalActions) != 1 {
		t.Fatalf("PlayerView().LegalActions = %#v, want one action", view.LegalActions)
	}
	return view.LegalActions[0].Handle
}

func concedeInput(t *testing.T, game *Game, player *model.Player) Input {
	view, err := game.PlayerView(player)
	if err != nil {
		t.Fatalf("PlayerView() error = %v", err)
	}
	return Input{
		Revision: view.Revision,
		Action: actionHandle(
			t,
			game,
			player,
		),
	}
}
