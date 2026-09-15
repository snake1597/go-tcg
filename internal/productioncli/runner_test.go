package productioncli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"go-tcg/internal/constants"
	"go-tcg/internal/game"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunRendersNumberedMenuAndWritesReplayOnEOF 驗證 CLI 以引擎視圖輸出編號選單，並在 EOF 時寫出可驗證的私人 replay。
// 輸入為固定 seed、replay 路徑與一筆非法編號後的 EOF；輸出為診斷與 replay，副作用為建立 replay 檔案但不提交任何行動。
func TestRunRendersNumberedMenuAndWritesReplayOnEOF(t *testing.T) {
	replayPath := filepath.Join(
		t.TempDir(),
		"game.replay.json",
	)
	var output bytes.Buffer
	err := Run(
		[]string{
			"--seed",
			"7",
			"--replay-out",
			replayPath,
		},
		strings.NewReader("0\n"),
		&output,
		filepath.Clean("../.."),
	)
	if err == nil || !strings.Contains(err.Error(), "EOF") {
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

// TestRunCompletesHumanVsBotGameAndWritesVerifiableReplay 驗證 CLI 由真人與 bot 輪流透過各自 PlayerView 提交，並可在真人投降後完成單局。
// 輸入為真人先讓過再投降的編號輸入、固定 seed 與 replay 路徑；輸出為完成後的畫面與可驗證 replay，副作用為 bot 提交回應並建立 replay 檔案。
func TestRunCompletesHumanVsBotGameAndWritesVerifiableReplay(t *testing.T) {
	replayPath := filepath.Join(
		t.TempDir(),
		"human-vs-bot.replay.json",
	)
	var output bytes.Buffer
	err := Run(
		[]string{
			"--seed",
			"7",
			"--replay-out",
			replayPath,
		},
		strings.NewReader("8\n2\n"),
		&output,
		filepath.Clean("../.."),
	)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	text := output.String()
	for _, want := range []string{
		"真人：player-1　bot：player-2",
		"bot player-2 已提交行動。",
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
	if len(replay.Steps) < 2 {
		t.Fatalf("replay steps = %d, want human and bot submissions", len(replay.Steps))
	}
	if replay.Steps[1].Player == nil || replay.Steps[1].Player.UID != "player-2" {
		t.Fatalf("bot replay player = %#v, want player-2", replay.Steps[1].Player)
	}
	lastStep := replay.Steps[len(replay.Steps)-1]
	if lastStep.Player == nil || lastStep.Player.UID != "player-1" {
		t.Fatalf("final replay player = %#v, want conceding player-1", lastStep.Player)
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
	passInput, passErr := readSelection(
		bufio.NewScanner(strings.NewReader("2\n")),
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
	paymentInput, paymentErr := readSelection(
		bufio.NewScanner(strings.NewReader("1\n1,2\n")),
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
