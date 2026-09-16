package bot

import (
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/game"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
	"testing"
)

// TestHeuristicDecideUsesOnlyViewPriority 驗證決策僅從 PlayerView 的合法行動與待選項取值。
// 輸入為同時包含各種行動的可見視圖；輸出為優先選取攻擊、待選項與正確 revision 的提交資料，且不會讀取遊戲內部狀態。
func TestHeuristicDecideUsesOnlyViewPriority(t *testing.T) {
	bot := NewHeuristic(
		NewSeededRandom(7),
	)
	view := game.PlayerView{
		Revision: 12,
		LegalActions: []game.LegalAction{
			{
				Handle:        "pass",
				Kind:          constants.ActionPass,
				HeuristicRank: 5,
			},
			{
				Handle:        "activate",
				Kind:          constants.ActionActivate,
				HeuristicRank: 1,
			},
			{
				Handle:        "attack",
				Kind:          constants.ActionAttack,
				HeuristicRank: 0,
			},
		},
	}

	input, err := bot.Decide(view)
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}
	if input.Revision != 12 || input.Action != "attack" || input.Choice != "" {
		t.Fatalf("Decide() = %#v, want attack at revision 12", input)
	}

	view.PendingChoice = &game.PendingChoice{
		Options: []game.ViewHandle{
			"choice-one",
			"choice-two",
		},
		Choices: []game.VisibleChoice{
			{
				Handle:        "choice-one",
				CardName:      "First",
				HeuristicRank: 1,
			},
			{
				Handle:        "choice-two",
				CardName:      "Second",
				HeuristicRank: 0,
			},
		},
	}
	input, err = bot.Decide(view)
	if err != nil {
		t.Fatalf("Decide() pending choice error = %v", err)
	}
	if input.Revision != 12 || input.Action != "" || input.Choice != "choice-two" {
		t.Fatalf("Decide() = %#v, want pending choice at revision 12", input)
	}
}

// TestHeuristicDecideUsesSeededRandomOnlyForTies 驗證相同視圖與 seed 的平手選擇可重現。
// 輸入為兩個同優先級的合法行動與相同 seed；輸出為相同行動，副作用僅為平手時推進注入亂數來源。
func TestHeuristicDecideUsesSeededRandomOnlyForTies(t *testing.T) {
	view := game.PlayerView{
		Revision: 3,
		LegalActions: []game.LegalAction{
			{
				Handle:        "first",
				Kind:          constants.ActionActivate,
				CardName:      "A",
				HeuristicRank: 1,
			},
			{
				Handle:        "second",
				Kind:          constants.ActionActivate,
				CardName:      "B",
				HeuristicRank: 1,
			},
		},
	}
	first, err := NewHeuristic(
		NewSeededRandom(99),
	).Decide(view)
	if err != nil {
		t.Fatalf("first Decide() error = %v", err)
	}
	second, err := NewHeuristic(
		NewSeededRandom(99),
	).Decide(view)
	if err != nil {
		t.Fatalf("second Decide() error = %v", err)
	}
	if first.Revision != second.Revision || first.Action != second.Action || first.Choice != second.Choice {
		t.Fatalf("same seed decisions = %#v and %#v, want equal", first, second)
	}
}

// TestHeuristicDecideOmitsReserveForZeroCostVerita 驗證替代費用已令 ReserveCost 為零時不會提交任何 Reserve handle。
// 輸入為 ReserveOptions 非空的 Verita 合法行動；輸出為只含 action 與 revision 的 Input，副作用僅可能消耗 bot 平手亂數。
func TestHeuristicDecideOmitsReserveForZeroCostVerita(t *testing.T) {
	view := game.PlayerView{
		Revision: 8,
		LegalActions: []game.LegalAction{
			{
				Handle:      "verita",
				Kind:        constants.ActionActivate,
				CardName:    "Verita, Queen of Hearts",
				ReserveCost: 0,
				ReserveOptions: []game.VisibleCard{
					{
						Handle: "reserve",
						Name:   "Two of Hearts",
					},
				},
			},
		},
	}

	input, err := NewHeuristic(
		NewSeededRandom(7),
	).Decide(view)
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}
	if input.Revision != view.Revision || input.Action != "verita" || len(input.Reserve) != 0 {
		t.Fatalf("Decide() = %#v, want zero-reserve Verita input", input)
	}
}

// TestHeuristicRunRefreshesStaleViewAndStopsAtLimit 驗證回合迴圈在過期提交後重新取得 PlayerView，並限制整場提交數。
// 輸入為第一次提交回傳過期 revision 的可控遊戲邊界；輸出為第二次使用新 revision 的提交，副作用是只呼叫兩次 Submit。
func TestHeuristicRunRefreshesStaleViewAndStopsAtLimit(t *testing.T) {
	controller := &scriptedController{
		views: []game.PlayerView{
			{
				Revision: 1,
				LegalActions: []game.LegalAction{
					{
						Handle:        "pass-one",
						Kind:          constants.ActionPass,
						HeuristicRank: 5,
					},
				},
			},
			{
				Revision: 2,
				LegalActions: []game.LegalAction{
					{
						Handle:        "activate-two",
						Kind:          constants.ActionActivate,
						HeuristicRank: 1,
					},
				},
			},
			{
				Revision: 3,
				Finished: true,
			},
		},
		staleFirstSubmission: true,
	}
	bot := NewHeuristic(
		NewSeededRandom(1),
	)

	err := bot.Run(
		controller,
		model.PlayerOne,
		3,
	)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(controller.inputs) != 2 {
		t.Fatalf("Submit() calls = %d, want 2", len(controller.inputs))
	}
	if controller.inputs[1].Revision != 2 {
		t.Fatalf("second submission revision = %d, want 2", controller.inputs[1].Revision)
	}
}

// TestHeuristicRunRejectsNoProgressCycle 驗證連續相同可見決策局面不會讓 bot 無限提交。
// 輸入為兩個語意相同、revision 不同的視圖；輸出為無進展循環錯誤，副作用是提交兩次後停止。
func TestHeuristicRunRejectsNoProgressCycle(t *testing.T) {
	controller := &scriptedController{
		views: []game.PlayerView{
			{
				Revision: 1,
				LegalActions: []game.LegalAction{
					{
						Handle:        "pass-one",
						Kind:          constants.ActionPass,
						HeuristicRank: 5,
					},
				},
			},
			{
				Revision: 2,
				LegalActions: []game.LegalAction{
					{
						Handle:        "pass-two",
						Kind:          constants.ActionPass,
						HeuristicRank: 5,
					},
				},
			},
		},
	}

	err := NewHeuristic(
		NewSeededRandom(1),
	).Run(
		controller,
		model.PlayerOne,
		4,
	)
	if err == nil {
		t.Fatal("Run() error = nil, want no progress cycle")
	}
	if len(controller.inputs) != 2 {
		t.Fatalf("Submit() calls = %d, want 2", len(controller.inputs))
	}
}

type scriptedController struct {
	views                []game.PlayerView
	inputs               []game.Input
	staleFirstSubmission bool
}

// PlayerView 回傳腳本指定的當前玩家視圖，供 bot 回合迴圈以正式控制器邊界讀取。
// 輸入為 bot 玩家但不檢查身分；輸出為下一個視圖，副作用是視圖索引隨成功提交前進。
func (controller *scriptedController) PlayerView(_ *model.Player) (game.PlayerView, error) {
	if len(controller.views) == 0 {
		return game.PlayerView{}, fmt.Errorf("no scripted view")
	}
	return controller.views[0], nil
}

// Submit 記錄 bot 的提交，第一次可模擬過期 revision，成功時前進至下一個腳本視圖。
// 輸入為玩家與決策資料；輸出為過期或成功結果，副作用是追加輸入並在成功時消耗目前視圖。
func (controller *scriptedController) Submit(_ *model.Player, input game.Input) error {
	controller.inputs = append(controller.inputs, input)
	if controller.staleFirstSubmission {
		controller.staleFirstSubmission = false
		controller.views = controller.views[1:]
		return fmt.Errorf("%w", tcgErrors.ErrStaleRevision)
	}
	controller.views = controller.views[1:]
	return nil
}
