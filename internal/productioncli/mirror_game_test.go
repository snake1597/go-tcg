package productioncli

import (
	"fmt"
	"go-tcg/internal/bot"
	"go-tcg/internal/constants"
	"go-tcg/internal/game"
	"go-tcg/internal/model"
	"path/filepath"
	"testing"
)

// mirrorGameActionLimit 為每個 seed 的鏡像發布 gate 提供明確收斂上限。
const mirrorGameActionLimit = 1000

// mirrorGameResult 保存單一鏡像 gate 的可重現結果與失敗現場。
// Replay 與 Coverage 供驗收，Seed／Step／Diagnostic／StateHash 供失敗定位；此型別本身無副作用。
type mirrorGameResult struct {
	Replay     game.Replay
	Coverage   map[string]bool
	Winner     string
	Seed       uint64
	Step       int
	Diagnostic string
	StateHash  string
}

// TestMirrorGamesCompleteFor100Seeds 是首版發布 gate，要求 100 個不同 seed 都在固定行動上限內確定結束。
// 輸入為 seed 1 到 100；輸出為每局有勝者、無 Needs Ruling 且 replay hash 可重播，副作用為依序推進 100 個隔離測試單局。
func TestMirrorGamesCompleteFor100Seeds(t *testing.T) {
	for seed := uint64(1); seed <= 100; seed++ {
		result, err := runMirrorGame(seed, mirrorGameActionLimit)
		if err != nil {
			failMirrorGame(t, result, err)
		}
		if result.Winner == "" || result.Diagnostic != "" {
			failMirrorGame(
				t,
				result,
				fmt.Errorf("finished with winner %q and diagnostic %q", result.Winner, result.Diagnostic),
			)
		}
		if err := result.Replay.Verify(); err != nil {
			failMirrorGame(
				t,
				result,
				fmt.Errorf("verify replay: %w", err),
			)
		}
	}
}

// runMirrorGame 執行一局由兩個確定性 bot 代表的 PlayerView 對局，並保留 action limit 或 panic 的診斷現場。
// 輸入為遊戲 seed 與正整數行動上限；輸出為 replay、互動覆蓋、勝者、步數、診斷、hash 與錯誤，副作用為建立及推進一個測試內單局。
func runMirrorGame(seed uint64, actionLimit int) (result mirrorGameResult, err error) {
	result = mirrorGameResult{
		Coverage: map[string]bool{
			"card play":      false,
			"Cardistry":      false,
			"combat":         false,
			"Pending Choice": false,
			"Stack response": false,
		},
		Seed: seed,
	}
	var match *game.Game
	defer func() {
		if match != nil {
			result.StateHash = match.StateHash()
			result.Replay = match.Replay()
		}
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panic: %v", recovered)
		}
	}()
	repositoryRoot := filepath.Clean("../..")
	match, err = game.NewStandardGame(
		game.StandardGameConfig{
			Players: [2]*model.Player{
				model.PlayerOne,
				model.PlayerTwo,
			},
			RepositoryRoot: repositoryRoot,
			Seed:           seed,
		},
	)
	if err != nil {
		return result, fmt.Errorf("start game: %w", err)
	}
	bots := map[string]*bot.Heuristic{
		model.PlayerOne.UID: bot.NewHeuristic(
			bot.NewSeededRandom(seed),
		),
		model.PlayerTwo.UID: bot.NewHeuristic(
			bot.NewSeededRandom(seed + 1),
		),
	}
	for step := 0; step < actionLimit; step++ {
		result.Step = step
		player, playerErr := currentDecisionPlayer(match)
		if playerErr != nil {
			return result, fmt.Errorf("current decision player: %w", playerErr)
		}
		view, viewErr := match.PlayerView(player)
		if viewErr != nil {
			return result, fmt.Errorf("read player view: %w", viewErr)
		}
		result.Diagnostic = view.Diagnostic
		if view.Finished {
			if view.Winner == nil {
				return result, fmt.Errorf("finished without winner")
			}
			result.Winner = view.Winner.UID
			return result, nil
		}
		if view.PendingChoice != nil {
			result.Coverage["Pending Choice"] = true
		}
		input, decideErr := bots[player.UID].Decide(view)
		if decideErr != nil {
			return result, fmt.Errorf("decide: %w", decideErr)
		}
		for _, action := range view.LegalActions {
			if action.Handle != input.Action {
				continue
			}
			if action.Kind == constants.ActionAttack {
				result.Coverage["combat"] = true
			}
			if action.Kind == constants.ActionActivate {
				if action.FloatingMemoryOptions != nil {
					result.Coverage["Cardistry"] = true
				} else {
					result.Coverage["card play"] = true
				}
				if len(view.EffectsStack) > 0 {
					result.Coverage["Stack response"] = true
				}
			}
		}
		beforeHash := match.StateHash()
		if submitErr := match.Submit(player, input); submitErr != nil {
			if match.StateHash() != beforeHash {
				return result, fmt.Errorf("rejected input changed state: %w", submitErr)
			}
			return result, fmt.Errorf("submit: %w", submitErr)
		}
	}
	result.Step = actionLimit
	return result, fmt.Errorf("action limit %d reached", actionLimit)
}

// failMirrorGame 以發布 gate 要求的可重現欄位終止目前測試。
// 輸入為 testing 邊界、鏡像結果與根因；輸出不返回，副作用為記錄 seed、step、diagnostic、state hash 並標記測試失敗。
func failMirrorGame(t *testing.T, result mirrorGameResult, err error) {
	t.Helper()
	t.Fatalf(
		"seed=%d step=%d diagnostic=%q state_hash=%s error=%v",
		result.Seed,
		result.Step,
		result.Diagnostic,
		result.StateHash,
		err,
	)
}
