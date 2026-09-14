package game

import (
	"encoding/json"
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/model"
)

type ReplayError struct {
	InputIndex int // -1 when verification fails before the first input.
	Failure    constants.ReplayFailure
	Reason     string
	Cause      error
}

func (e *ReplayError) Error() string {
	description := describeReplayFailure(e.Failure)
	if e.Failure == constants.ReplayVersionMismatch {
		return fmt.Sprintf("replay header %s: %s", description, e.Reason)
	}
	return fmt.Sprintf("replay input %d %s: %s", e.InputIndex, description, e.Reason)
}

func (e *ReplayError) Unwrap() error {
	return e.Cause
}

func describeReplayFailure(f constants.ReplayFailure) string {
	switch f {
	case constants.ReplayVersionMismatch:
		return "version mismatch"
	case constants.ReplayInputRejected:
		return "rejected"
	case constants.ReplayStateHashMismatch:
		return "state hash mismatch"
	default:
		return string(f)
	}
}

type Versions struct {
	Engine   string `json:"engine"`
	Rules    string `json:"rules"`
	CardData string `json:"card_data"`
	Deck     string `json:"deck"`
	PRNG     string `json:"prng"`
}

type Replay struct {
	FormatVersion  int             `json:"format_version"`
	Versions       Versions        `json:"versions"`
	InitialSeed    uint64          `json:"initial_seed"`
	InitialState   *gameState      `json:"initial_state"`
	InitialPlayers []*model.Player `json:"initial_players"`
	Steps          []ReplayStep    `json:"steps"`
}

type ReplayStep struct {
	Player    *model.Player `json:"player"`
	Input     Input         `json:"input"`
	StateHash string        `json:"state_hash"`
}

// Verify 先檢查格式與各資料版本，再由初始快照逐步 Submit 並比對狀態雜湊。
// 遇到第一個輸入拒絕或雜湊不符即停止；標頭不相容或缺少初始狀態以 InputIndex=-1 回報。
func (r Replay) Verify() error {
	if r.FormatVersion != constants.ReplayFormatVersion {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible replay format version %d, want %d",
				r.FormatVersion,
				constants.ReplayFormatVersion,
			),
		)
	}
	if err := verifyVersions(r.Versions); err != nil {
		return err
	}

	if r.InitialState == nil || len(r.InitialPlayers) == 0 {
		return newReplayVersionMismatch("missing initial state")
	}
	game := NewGame(r.InitialSeed)
	game.state = cloneGameState(*r.InitialState)
	game.players = append(
		[]*model.Player(nil),
		r.InitialPlayers...,
	)
	for index, step := range r.Steps {
		if err := game.Submit(step.Player, step.Input); err != nil {
			return &ReplayError{
				InputIndex: index,
				Failure:    constants.ReplayInputRejected,
				Reason:     err.Error(),
				Cause:      err,
			}
		}
		if got := game.StateHash(); got != step.StateHash {
			return &ReplayError{
				InputIndex: index,
				Failure:    constants.ReplayStateHashMismatch,
				Reason:     fmt.Sprintf("got %s, want %s", got, step.StateHash),
			}
		}
	}
	return nil
}

// captureReplayInitialState 將目前狀態設為 replay 起點，複製玩家切片並清除先前步驟。
// 標準開局在配置完成後呼叫，讓重播從開局事件處理後的狀態開始。
func (g *Game) captureReplayInitialState() {
	state := cloneGameState(g.state)
	g.replay.InitialState = &state
	g.replay.InitialPlayers = append(
		[]*model.Player(nil),
		g.players...,
	)
	g.replay.Steps = nil
}

// cloneGameState 透過 JSON 往返複製 gameState 中可序列化的資料，避免初始快照共用可變狀態。
// 編碼或解碼失敗視為內部狀態錯誤並 panic，不回傳零值或可恢復的 error。
func cloneGameState(state gameState) gameState {
	encoded, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	var clone gameState
	if err := json.Unmarshal(
		encoded,
		&clone,
	); err != nil {
		panic(err)
	}
	return clone
}

func verifyVersions(got Versions) error {
	want := currentVersions()
	if got.Engine != want.Engine {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible engine version %q, want %q",
				got.Engine,
				want.Engine,
			),
		)
	}
	if got.Rules != want.Rules {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible rules version %q, want %q",
				got.Rules,
				want.Rules,
			),
		)
	}
	if got.CardData != want.CardData {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible card data version %q, want %q",
				got.CardData,
				want.CardData,
			),
		)
	}
	if got.Deck != want.Deck {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible deck version %q, want %q",
				got.Deck,
				want.Deck,
			),
		)
	}
	if got.PRNG != want.PRNG {
		return newReplayVersionMismatch(
			fmt.Sprintf(
				"incompatible PRNG version %q, want %q",
				got.PRNG,
				want.PRNG,
			),
		)
	}
	return nil
}

func newReplayVersionMismatch(reason string) *ReplayError {
	return &ReplayError{
		InputIndex: -1,
		Failure:    constants.ReplayVersionMismatch,
		Reason:     reason,
	}
}
