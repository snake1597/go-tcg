package productioncli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-tcg/internal/constants"
	"go-tcg/internal/game"
	"go-tcg/internal/model"
)

// TestRunRendersNumberedMenuAndWritesReplayOnEOF 驗證 CLI 以引擎視圖輸出編號選單，並在 EOF 時寫出可驗證的私人 replay。
// 輸入為固定 seed、replay 路徑與一筆非法編號後的 EOF；輸出為診斷與 replay，副作用為建立 replay 檔案但不提交任何行動。
func TestRunRendersNumberedMenuAndWritesReplayOnEOF(t *testing.T) {
	replayPath := filepath.Join(
		t.TempDir(),
		"game.replay.json",
	)
	var output bytes.Buffer
	repositoryRoot := filepath.Clean("../..")
	input := strings.NewReader("0\n")
	err := Run(
		[]string{
			"--seed",
			"7",
			"--replay-out",
			replayPath,
		},
		input,
		&output,
		repositoryRoot,
	)
	if err == nil {
		t.Fatal("Run() error = nil, want EOF")
	}
	errorMessage := err.Error()
	if !strings.Contains(errorMessage, "EOF") {
		t.Fatalf("Run() error = %v, want EOF", err)
	}
	text := output.String()
	for _, want := range []string{
		"隱私警告",
		"回合：1",
		"自己手牌",
		"Effects Stack",
		"可選行動",
		"請輸入編號",
		"無效編號",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("CLI output missing %q:\n%s", want, text)
		}
	}
	encoded, readErr := os.ReadFile(replayPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", replayPath, readErr)
	}
	var replay game.Replay
	if unmarshalErr := json.Unmarshal(encoded, &replay); unmarshalErr != nil {
		t.Fatalf("unmarshal replay error = %v", unmarshalErr)
	}
	if verifyErr := replay.Verify(); verifyErr != nil {
		t.Fatalf("replay.Verify() error = %v", verifyErr)
	}
	if len(replay.Steps) != 0 {
		t.Fatalf("replay steps = %d, want 0 after invalid menu input", len(replay.Steps))
	}
}

// TestRunExplainsBotSubmissionsAndEffectStackOpportunityFlow 驗證 CLI 區分 bot 行動、選擇與 pass，並說明堆疊逐項結算後重新開啟 Opportunity。
// 輸入為 seed 7 的 Four of Hearts 與 Fiery Interference 回應流程；輸出為具體提交及流程提示，副作用為建立測試目錄內的 replay。
func TestRunExplainsBotSubmissionsAndEffectStackOpportunityFlow(t *testing.T) {
	replayPath := filepath.Join(
		t.TempDir(),
		"opportunity-flow.replay.json",
	)
	var output bytes.Buffer
	repositoryRoot := filepath.Clean("../..")
	input := strings.NewReader("3\n1,2,3,4\n2\n2\n")
	err := Run(
		[]string{
			"--seed",
			"7",
			"--replay-out",
			replayPath,
		},
		input,
		&output,
		repositoryRoot,
	)
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	if err == nil || !strings.Contains(errorMessage, "EOF") {
		t.Fatalf("Run() error = %v, want EOF", err)
	}
	text := output.String()
	for _, want := range []string{
		"bot player-2 啟動 Fiery Interference。",
		"Opportunity 仍由 player-2 持有。",
		"bot player-2 已完成一項選擇。",
		"bot player-2 選擇 pass，Opportunity 移交給 player-1。",
		"雙方連續 pass，結算堆疊頂端：Fiery Interference。",
		"Effects Stack 尚有 1 個項目；重新開啟回應窗口，Opportunity 交給 player-1。",
		"若所有玩家連續 pass，結算堆疊頂端：Four of Hearts。",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("CLI output missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "bot player-2 已提交行動。") {
		t.Fatalf("CLI output retained ambiguous bot submission text:\n%s", text)
	}
}

// TestRunCompletesHumanVsBotGameAndWritesVerifiableReplay 驗證 CLI 由真人與 bot 輪流透過各自 PlayerView 提交，並可在真人投降後完成單局。
// 輸入為真人先讓過再投降的編號輸入、固定 seed 與 replay 路徑；輸出為完成後的畫面與可驗證 replay，副作用為 bot 提交回應並建立 replay 檔案。
func TestRunCompletesHumanVsBotGameAndWritesVerifiableReplay(t *testing.T) {
	replayPath := filepath.Join(
		t.TempDir(),
		"human-vs-bot.replay.json",
	)
	var output bytes.Buffer
	repositoryRoot := filepath.Clean("../..")
	input := strings.NewReader("8\n2\n")
	err := Run(
		[]string{
			"--seed",
			"7",
			"--replay-out",
			replayPath,
		},
		input,
		&output,
		repositoryRoot,
	)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	text := output.String()
	for _, want := range []string{
		"真人：player-1　bot：player-2",
		"bot player-2 啟動 Fiery Interference。",
		"結果：player-2 獲勝",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("CLI output missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "- Two of Hearts\n") {
		t.Fatalf("CLI output leaked bot private hand:\n%s", text)
	}

	encoded, readErr := os.ReadFile(replayPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", replayPath, readErr)
	}
	var replay game.Replay
	if unmarshalErr := json.Unmarshal(encoded, &replay); unmarshalErr != nil {
		t.Fatalf("unmarshal replay error = %v", unmarshalErr)
	}
	if verifyErr := replay.Verify(); verifyErr != nil {
		t.Fatalf("replay.Verify() error = %v", verifyErr)
	}
	replayStepCount := len(replay.Steps)
	if replayStepCount < 2 {
		t.Fatalf("replay steps = %d, want human and bot submissions", replayStepCount)
	}
	if replay.Steps[1].Player == nil || replay.Steps[1].Player.UID != "player-2" {
		t.Fatalf("bot replay player = %#v, want player-2", replay.Steps[1].Player)
	}
	lastStep := replay.Steps[len(replay.Steps)-1]
	if lastStep.Player == nil || lastStep.Player.UID != "player-1" {
		t.Fatalf("final replay player = %#v, want conceding player-1", lastStep.Player)
	}
}

// TestRunStopsAtSubmissionLimitWithReplayDiagnostics 驗證 production CLI 不會在 bot 輪替迴圈無限提交。
// 輸入為只允許一次提交的固定 seed 單局；輸出為含 seed、step、replay 與 state hash 的錯誤診斷，副作用為寫出可驗證 replay。
func TestRunStopsAtSubmissionLimitWithReplayDiagnostics(t *testing.T) {
	replayPath := filepath.Join(
		t.TempDir(),
		"limited.replay.json",
	)
	var output bytes.Buffer
	repositoryRoot := filepath.Clean("../..")
	input := strings.NewReader("8\n")
	err := Run(
		[]string{
			"--seed",
			"7",
			"--submission-limit",
			"1",
			"--replay-out",
			replayPath,
		},
		input,
		&output,
		repositoryRoot,
	)
	if err == nil {
		t.Fatal("Run() error = nil, want submission limit")
	}
	errorMessage := err.Error()
	if !strings.Contains(errorMessage, "submission limit 1 reached") {
		t.Fatalf("Run() error = %v, want submission limit", err)
	}
	outputText := output.String()
	for _, want := range []string{
		"seed=7",
		"step=1",
		"diagnostic=submission limit 1 reached",
		"replay=" + replayPath,
		"state_hash=",
	} {
		if !strings.Contains(outputText, want) {
			t.Fatalf("CLI output missing %q:\n%s", want, outputText)
		}
	}
	encoded, readErr := os.ReadFile(replayPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", replayPath, readErr)
	}
	var replay game.Replay
	if unmarshalErr := json.Unmarshal(encoded, &replay); unmarshalErr != nil {
		t.Fatalf("unmarshal replay error = %v", unmarshalErr)
	}
	if verifyErr := replay.Verify(); verifyErr != nil {
		t.Fatalf("replay.Verify() error = %v", verifyErr)
	}
	replayStepCount := len(replay.Steps)
	if replayStepCount != 1 {
		t.Fatalf("replay steps = %d, want 1", replayStepCount)
	}
}

// TestReadSelectionUsesEngineProvidedPassAndFloatingMemoryOptions 驗證可略過選擇與 Floating Memory 都只使用 PlayerView 的編號選項。
// 輸入為含 pass 與兩張付款卡的可見視圖；輸出為對應 handle，副作用僅為消耗腳本輸入。
func TestReadSelectionUsesEngineProvidedPassAndFloatingMemoryOptions(t *testing.T) {
	passView := game.PlayerView{
		Revision: 4,
		LegalActions: []game.LegalAction{
			{
				Handle: "pass-handle",
				Kind:   constants.ActionPass,
			},
		},
		PendingChoice: &game.PendingChoice{
			Options: []game.ViewHandle{
				"choice-handle",
			},
			CanPass: true,
		},
	}
	passReader := strings.NewReader("2\n")
	passScanner := bufio.NewScanner(passReader)
	passInput, passErr := readSelection(
		passScanner,
		io.Discard,
		passView,
	)
	if passErr != nil {
		t.Fatalf("readSelection() pass error = %v", passErr)
	}
	if passInput.Action != "pass-handle" || passInput.Choice != "" {
		t.Fatalf("pass input = %#v, want engine pass handle", passInput)
	}

	paymentView := game.PlayerView{
		Revision: 5,
		LegalActions: []game.LegalAction{
			{
				Handle: "cardistry-handle",
				Kind:   constants.ActionActivate,
				FloatingMemoryOptions: []game.VisibleCard{
					{
						Handle: "floating-one",
						Name:   "Five of Spades",
					},
					{
						Handle: "floating-two",
						Name:   "Five of Spades",
					},
				},
			},
		},
	}
	paymentReader := strings.NewReader("1\n1,2\n")
	paymentScanner := bufio.NewScanner(paymentReader)
	paymentInput, paymentErr := readSelection(
		paymentScanner,
		io.Discard,
		paymentView,
	)
	if paymentErr != nil {
		t.Fatalf("readSelection() payment error = %v", paymentErr)
	}
	if paymentInput.Action != "cardistry-handle" || len(paymentInput.FloatingMemory) != 2 || paymentInput.FloatingMemory[0] != "floating-one" || paymentInput.FloatingMemory[1] != "floating-two" {
		t.Fatalf("payment input = %#v, want engine floating memory handles", paymentInput)
	}
}

// TestRenderViewExplainsDrawChoiceDamageAndStackTiming 驗證 CLI 清楚呈現自動抽牌、攻擊目標、剩餘生命與 Effects Stack 限制。
// 輸入為含抽牌事件、受傷 Champion、待選目標與 Effects Stack 的玩家視圖；輸出為含各項中文提示的終端文字。
func TestRenderViewExplainsDrawChoiceDamageAndStackTiming(t *testing.T) {
	var output bytes.Buffer
	renderView(
		&output,
		model.PlayerOne,
		game.PlayerView{
			Phase: game.PhaseMain,
			Champions: []game.VisibleChampion{
				{
					Owner:    model.PlayerOne,
					CardName: "Spirit of Fire",
					Life:     15,
					Damage:   2,
				},
			},
			EffectsStack: []game.VisibleEffectStackItem{
				{
					Kind:       "ability",
					Controller: model.PlayerOne,
					SourceName: "Four of Hearts",
				},
			},
			VisibleEvents: []game.VisibleEvent{
				{
					Kind:     "draw",
					CardName: "Fiery Interference",
				},
				{
					Kind:     "materialize-banish-memory",
					CardName: "Four of Hearts",
				},
			},
			PendingChoice: &game.PendingChoice{
				Options: []game.ViewHandle{
					"target",
				},
				Choices: []game.VisibleChoice{
					{
						Handle:   "target",
						CardName: "Spirit of Fire",
					},
				},
			},
		},
	)
	text := output.String()
	for _, want := range []string{
		"0/13（生命上限 15，傷害 2）",
		"抽牌階段：抽到 Fiery Interference",
		"物質化付款：隨機放逐 Four of Hearts",
		"Effects Stack 尚未清空；只能啟動 Fast 卡或 pass。",
		"1. Spirit of Fire",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("renderView() output missing %q:\n%s", want, text)
		}
	}
}

// TestReadReserveValidatesNumberedSelections 驗證公開 Reserve 選單只接受正確張數、唯一且範圍內的十進位編號。
// 輸入為正確、重複、越界、格式錯誤、張數錯誤與 EOF 腳本；輸出為合法 handles 或 EOF，副作用僅為消耗 scanner 並寫入提示。
func TestReadReserveValidatesNumberedSelections(t *testing.T) {
	action := game.LegalAction{
		ReserveCost: 2,
		ReserveOptions: []game.VisibleCard{
			{
				Handle: "first",
				Name:   "First",
			},
			{
				Handle: "second",
				Name:   "Second",
			},
		},
	}
	testCases := []struct {
		name             string
		input            string
		wantInvalidCount int
		wantEOF          bool
	}{
		{
			name:  "correct count",
			input: "1,2\n",
		},
		{
			name:             "duplicate selection",
			input:            "1,1\n1,2\n",
			wantInvalidCount: 1,
		},
		{
			name:             "out of range",
			input:            "1,3\n1,2\n",
			wantInvalidCount: 1,
		},
		{
			name:             "malformed",
			input:            "one,2\n1,2\n",
			wantInvalidCount: 1,
		},
		{
			name:             "wrong count",
			input:            "1\n1,2\n",
			wantInvalidCount: 1,
		},
		{
			name:    "EOF",
			wantEOF: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var output bytes.Buffer
				reader := strings.NewReader(testCase.input)
				scanner := bufio.NewScanner(reader)
				handles, err := ReadReserve(
					scanner,
					&output,
					action,
				)
				if testCase.wantEOF {
					if !errors.Is(err, io.EOF) {
						t.Fatalf("ReadReserve() error = %v, want EOF", err)
					}
					return
				}
				if err != nil {
					t.Fatalf("ReadReserve() error = %v", err)
				}
				if len(handles) != 2 || handles[0] != "first" || handles[1] != "second" {
					t.Fatalf("ReadReserve() = %#v, want first and second", handles)
				}
				if got := strings.Count(output.String(), "無效編號"); got != testCase.wantInvalidCount {
					t.Fatalf("invalid messages = %d, want %d:\n%s", got, testCase.wantInvalidCount, output.String())
				}
			},
		)
	}
}

// TestRejectedReserveInputDoesNotChangeStateHash 驗證 CLI 所對接的 Game.Submit 拒絕偽造或重複 Reserve handles 時維持原狀態。
// 輸入為 Standard PlayerView 中一個需 Reserve 的 activation 與重複 handle；輸出為拒絕錯誤及相同 state hash，副作用僅為建立測試單局。
func TestRejectedReserveInputDoesNotChangeStateHash(t *testing.T) {
	repositoryRoot := filepath.Clean("../..")
	match, err := game.NewStandardGame(
		game.StandardGameConfig{
			Players: [2]*model.Player{
				model.PlayerOne,
				model.PlayerTwo,
			},
			RepositoryRoot: repositoryRoot,
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
	var activation game.LegalAction
	for _, action := range view.LegalActions {
		if action.Kind == constants.ActionActivate && action.ReserveCost > 0 && len(action.ReserveOptions) > 0 {
			activation = action
			break
		}
	}
	if activation.Handle == "" {
		t.Fatalf("LegalActions = %#v, want paid activation", view.LegalActions)
	}
	beforeHash := match.StateHash()
	err = match.Submit(
		model.PlayerOne,
		game.Input{
			Revision: view.Revision,
			Action:   activation.Handle,
			Reserve: []game.ViewHandle{
				activation.ReserveOptions[0].Handle,
				activation.ReserveOptions[0].Handle,
			},
		},
	)
	if err == nil {
		t.Fatal("Submit() error = nil, want invalid Reserve rejection")
	}
	if match.StateHash() != beforeHash {
		t.Fatal("StateHash() changed after rejected Reserve input")
	}
}
