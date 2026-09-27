package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"go-tcg/internal/model"
)

// TestStateHashIncludesAllStateFields 驗證新增權威狀態欄位時 canonicalState 也有同名欄位。
// 輸入為 gameState 與 canonicalState 的型別；輸出為欄位完整性檢查結果，無副作用。
func TestStateHashIncludesAllStateFields(t *testing.T) {
	stateType := reflect.TypeFor[gameState]()
	canonicalType := reflect.TypeFor[canonicalState]()
	for index := range stateType.NumField() {
		field := stateType.Field(index)
		if _, exists := canonicalType.FieldByName(field.Name); !exists {
			t.Errorf("canonicalState omits gameState.%s", field.Name)
		}
	}
}

// TestStateHashChangesForCardistryAndIdentityState 驗證各個 Cardistry 狀態與未來 identity 計數器均參與 hash。
// 輸入為固定 seed 的獨立對局及各欄位的單一變動；輸出為每次變動後不同的 hash，副作用僅為測試對局狀態變動。
func TestStateHashChangesForCardistryAndIdentityState(t *testing.T) {
	testCases := []struct {
		name   string
		change func(*Game)
	}{
		{
			name: "Cardistry used",
			change: func(game *Game) {
				game.state.CardistryUsed[objectID("source")] = true
			},
		},
		{
			name: "Cardistry discount",
			change: func(game *Game) {
				game.state.CardistryDiscounts[model.PlayerOne.UID] = 1
			},
		},
		{
			name: "next effect",
			change: func(game *Game) {
				game.state.NextEffect++
			},
		},
		{
			name: "next ability",
			change: func(game *Game) {
				game.state.NextAbility++
			},
		},
		{
			name: "next object",
			change: func(game *Game) {
				game.state.NextObject++
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				game := newTestGame(42)
				before := game.StateHash()
				testCase.change(game)
				if got := game.StateHash(); got == before {
					t.Fatalf("StateHash() ignored %s", testCase.name)
				}
			},
		)
	}
}

// TestRejectedInputPreservesAuthoritativeState 驗證拒絕的輸入不消耗亂數、知識、事件或費用。
// 輸入為固定 seed 與偽造的 action handle；輸出為拒絕錯誤及完全不變的狀態、hash、replay，無持久副作用。
func TestRejectedInputPreservesAuthoritativeState(t *testing.T) {
	game := newTestGame(42)
	beforeHash := game.StateHash()
	beforeState, err := json.Marshal(game.state)
	if err != nil {
		t.Fatalf("marshal state before input: %v", err)
	}
	beforeReplay, err := json.Marshal(game.Replay())
	if err != nil {
		t.Fatalf("marshal replay before input: %v", err)
	}
	err = game.Submit(
		model.PlayerOne,
		Input{
			Revision: game.state.Revision,
			Action:   ViewHandle("forged"),
		},
	)
	if err == nil {
		t.Fatal("Submit() accepted forged handle")
	}
	afterState, err := json.Marshal(game.state)
	if err != nil {
		t.Fatalf("marshal state after input: %v", err)
	}
	afterReplay, err := json.Marshal(game.Replay())
	if err != nil {
		t.Fatalf("marshal replay after input: %v", err)
	}
	if game.StateHash() != beforeHash || string(afterState) != string(beforeState) || string(afterReplay) != string(beforeReplay) {
		t.Fatal("rejected input changed state, hash, or replay")
	}
}
