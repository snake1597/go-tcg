package productioncli

import (
	"fmt"
	"go-tcg/internal/bot"
	"go-tcg/internal/constants"
	"go-tcg/internal/game"
	"go-tcg/internal/model"
	"path/filepath"
	"reflect"
	"testing"
)

// TestMirrorGameCompletesWithRequiredInteractions 驗證兩個僅使用 PlayerView 的決策者可完成固定鏡像單局並覆蓋核心互動。
// 輸入為固定 seed 的兩次自動決策；輸出為相同結果與逐步相同 replay，副作用為推進兩個測試內單局。
func TestMirrorGameCompletesWithRequiredInteractions(t *testing.T) {
	first, firstCoverage, firstWinner, firstErr := runMirrorGame()
	if firstErr != nil {
		t.Fatalf("first game error = %v", firstErr)
	}
	second, secondCoverage, secondWinner, secondErr := runMirrorGame()
	if secondErr != nil {
		t.Fatalf("second game error = %v", secondErr)
	}
	for name, covered := range firstCoverage {
		if !covered {
			t.Fatalf("mirror game did not cover %s: %#v", name, firstCoverage)
		}
	}
	if firstWinner == "" || firstWinner != secondWinner || !reflect.DeepEqual(firstCoverage, secondCoverage) || !reflect.DeepEqual(first.Steps, second.Steps) {
		for index := 0; index < len(first.Steps) && index < len(second.Steps); index++ {
			if reflect.DeepEqual(first.Steps[index], second.Steps[index]) {
				continue
			}
			t.Logf("first different step %d:\nfirst:  %#v\nsecond: %#v", index, first.Steps[index], second.Steps[index])
			break
		}
		t.Fatalf("mirror results differ: winner %q/%q, coverage %#v/%#v, steps %d/%d", firstWinner, secondWinner, firstCoverage, secondCoverage, len(first.Steps), len(second.Steps))
	}
	if err := first.Verify(); err != nil {
		t.Fatalf("first replay.Verify() error = %v", err)
	}
	if err := second.Verify(); err != nil {
		t.Fatalf("second replay.Verify() error = %v", err)
	}
}

// runMirrorGame 執行一局由兩個確定性 bot 代表的 PlayerView 對局並記錄指定互動。
// 輸入為無；輸出為 replay、互動覆蓋、勝者與錯誤，副作用為建立及推進測試內單局。
func runMirrorGame() (game.Replay, map[string]bool, string, error) {
	match, err := game.NewStandardGame(game.StandardGameConfig{
		Players:        [2]*model.Player{model.PlayerOne, model.PlayerTwo},
		RepositoryRoot: filepath.Clean("../.."),
		Seed:           7,
	})
	if err != nil {
		return game.Replay{}, nil, "", fmt.Errorf("start game: %w", err)
	}
	bots := map[string]*bot.Heuristic{
		model.PlayerOne.UID: bot.NewHeuristic(bot.NewSeededRandom(7)),
		model.PlayerTwo.UID: bot.NewHeuristic(bot.NewSeededRandom(8)),
	}
	coverage := map[string]bool{"card play": false, "Cardistry": false, "combat": false, "Pending Choice": false, "Stack response": false}
	for step := 0; step < 1000; step++ {
		player, err := currentDecisionPlayer(match)
		if err != nil {
			return game.Replay{}, nil, "", err
		}
		view, err := match.PlayerView(player)
		if err != nil {
			return game.Replay{}, nil, "", err
		}
		if view.Finished {
			if view.Winner == nil {
				return game.Replay{}, nil, "", fmt.Errorf("finished without winner")
			}
			return match.Replay(), coverage, view.Winner.UID, nil
		}
		if view.PendingChoice != nil {
			coverage["Pending Choice"] = true
		}
		input, err := bots[player.UID].Decide(view)
		if err != nil {
			return game.Replay{}, nil, "", err
		}
		for _, action := range view.LegalActions {
			if action.Handle != input.Action {
				continue
			}
			if action.Kind == constants.ActionAttack {
				coverage["combat"] = true
			}
			if action.Kind == constants.ActionActivate {
				if action.FloatingMemoryOptions != nil {
					coverage["Cardistry"] = true
				} else {
					coverage["card play"] = true
				}
			}
		}
		if len(view.EffectsStack) > 0 && input.Action != "" {
			coverage["Stack response"] = true
		}
		if err := match.Submit(player, input); err != nil {
			return game.Replay{}, nil, "", err
		}
	}
	return game.Replay{}, nil, "", fmt.Errorf("game did not finish")
}
