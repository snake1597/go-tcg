package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"go-tcg/internal/model"
)

// FuzzRejectedInputPreservesStateHash 驗證任意 seed 與偽造 handle 都不能突破玩家隔離或改變單局。
// 輸入為 fuzz 產生的 seed 與 handle 文字；輸出為拒絕結果及不變的 state hash／replay，副作用僅為建立兩個隔離的記憶體單局。
func FuzzRejectedInputPreservesStateHash(f *testing.F) {
	const firstSeed uint64 = 1
	const secondSeed uint64 = 42
	f.Add(firstSeed, "forged")
	f.Add(secondSeed, "")
	f.Fuzz(
		func(t *testing.T, seed uint64, forged string) {
			match := newTestGame(seed)
			beforeHash := match.StateHash()
			beforeReplay := match.Replay()
			if forged == "" {
				forged = "forged"
			}
			err := match.Submit(
				model.PlayerOne,
				Input{
					Revision: match.state.Revision,
					Action:   ViewHandle(forged),
				},
			)
			if err == nil {
				t.Fatal("Submit() error = nil, want forged handle rejection")
			}
			afterHash := match.StateHash()
			afterReplay := match.Replay()
			if afterHash != beforeHash || !reflect.DeepEqual(afterReplay, beforeReplay) {
				t.Fatal("rejected input changed state hash or replay")
			}
		},
	)
}

// FuzzReplayHashDeterminism 驗證相同版本、seed 與 Input 對任意 seed 都產生相同 replay state hash。
// 輸入為 fuzz 產生的 seed；輸出為兩個相同 canonical replay 與可通過 Verify 的 hash，副作用僅為各提交一次投降至隔離單局。
func FuzzReplayHashDeterminism(f *testing.F) {
	const firstSeed uint64 = 1
	const secondSeed uint64 = 42
	f.Add(firstSeed)
	f.Add(secondSeed)
	f.Fuzz(
		func(t *testing.T, seed uint64) {
			first := newTestGame(seed)
			second := newTestGame(seed)
			view, err := first.PlayerView(model.PlayerOne)
			if err != nil {
				t.Fatalf("first PlayerView() error = %v", err)
			}
			input := Input{
				Revision: view.Revision,
				Action:   view.LegalActions[0].Handle,
			}
			if err := first.Submit(model.PlayerOne, input); err != nil {
				t.Fatalf("first Submit() error = %v", err)
			}
			if err := second.Submit(model.PlayerOne, input); err != nil {
				t.Fatalf("second Submit() error = %v", err)
			}
			firstResult := first.Replay()
			firstReplay, marshalErr := json.Marshal(firstResult)
			if marshalErr != nil {
				t.Fatalf("marshal first replay: %v", marshalErr)
			}
			secondResult := second.Replay()
			secondReplay, marshalErr := json.Marshal(secondResult)
			if marshalErr != nil {
				t.Fatalf("marshal second replay: %v", marshalErr)
			}
			firstHash := first.StateHash()
			secondHash := second.StateHash()
			if firstHash != secondHash || string(firstReplay) != string(secondReplay) {
				t.Fatalf("same seed %d produced different state or replay", seed)
			}
			if err := firstResult.Verify(); err != nil {
				t.Fatalf("Replay().Verify() error = %v", err)
			}
		},
	)
}
