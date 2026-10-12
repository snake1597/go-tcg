package productioncli

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"go-tcg/internal/bot"
	"go-tcg/internal/game"
	"go-tcg/internal/model"
)

const mirrorGameActionLimit = 1000
const mirrorGameTimeLimit = 5 * time.Second

// TestMirrorGamesCompleteForTenSeeds 驗證兩個只讀自身 PlayerView 的 bot 能在有界提交數內完成鏡像對戰。
// 輸入為 seed 1 到 10；輸出為每局具有勝者且 replay 可驗證，副作用為依序推進十個隔離測試單局。
func TestMirrorGamesCompleteForTenSeeds(t *testing.T) {
	for seed := uint64(1); seed <= 10; seed++ {
		t.Run(
			fmt.Sprintf("seed-%d", seed),
			func(t *testing.T) {
				match, err := runMirrorGame(
					seed,
					mirrorGameActionLimit,
					mirrorGameTimeLimit,
				)
				if err != nil {
					t.Fatalf("error=%v", err)
				}
				if err := match.Replay().Verify(); err != nil {
					t.Fatalf("Replay().Verify() error=%v", err)
				}
			},
		)
	}
}

// runMirrorGame 執行一局由兩個確定性 bot 操作的正式 Standard 對局。
// 輸入為遊戲 seed、正整數提交上限與 wall-clock 上限；輸出為已完成的 Game 或包含 step、diagnostic 與 state hash 的錯誤，副作用為建立並推進單局。
func runMirrorGame(seed uint64, actionLimit int, timeLimit time.Duration) (*game.Game, error) {
	match, err := game.NewStandardGame(
		game.StandardGameConfig{
			Players: []*model.Player{
				model.PlayerOne,
				model.PlayerTwo,
			},
			RepositoryRoot: filepath.Clean("../.."),
			Seed:           seed,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("start game: %w", err)
	}
	bots := map[string]*bot.Heuristic{
		model.PlayerOne.UID: bot.NewHeuristic(
			bot.NewSeededRandom(seed),
		),
		model.PlayerTwo.UID: bot.NewHeuristic(
			bot.NewSeededRandom(seed + 1),
		),
	}
	startedAt := time.Now()
	deadline := startedAt.Add(timeLimit)
	for step := 0; step < actionLimit; step++ {
		now := time.Now()
		if now.After(deadline) {
			stateHash := match.StateHash()
			return match, fmt.Errorf(
				"time limit %s reached at step %d; state_hash=%s",
				timeLimit,
				step,
				stateHash,
			)
		}
		player, err := currentDecisionPlayer(match)
		if err != nil {
			return match, fmt.Errorf("step %d decision player: %w", step, err)
		}
		view, err := match.PlayerView(player)
		if err != nil {
			return match, fmt.Errorf("step %d player view: %w", step, err)
		}
		if view.Finished {
			if view.Winner == nil || view.Diagnostic != "" {
				stateHash := match.StateHash()
				return match, fmt.Errorf(
					"step %d winner=%v diagnostic=%q state_hash=%s",
					step,
					view.Winner,
					view.Diagnostic,
					stateHash,
				)
			}
			return match, nil
		}
		input, err := bots[player.UID].Decide(view)
		if err != nil {
			return match, fmt.Errorf("step %d decide: %w", step, err)
		}
		if err := match.Submit(player, input); err != nil {
			return match, fmt.Errorf("step %d submit: %w", step, err)
		}
	}
	stateHash := match.StateHash()
	return match, fmt.Errorf(
		"action limit %d reached; state_hash=%s",
		actionLimit,
		stateHash,
	)
}
