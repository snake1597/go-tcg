package bot

import (
	"errors"
	"fmt"
	"go-tcg/internal/constants"
	"go-tcg/internal/game"
	"go-tcg/internal/model"
	tcgErrors "go-tcg/internal/tcg_errors"
	"sort"
	"strings"
)

const maxCandidatesPerDecision = 128

// RandomSource 提供 bot 僅在同分候選間選擇時所需的可注入亂數。
// 實作者輸出下一個 uint64；bot 不保存或讀取遊戲引擎的亂數狀態。
type RandomSource interface {
	Uint64() uint64
}

// Controller 定義 bot 操作單局所需的最小正式邊界。
// PlayerView 輸出指定玩家的視圖，Submit 接受附帶 revision 的選擇；兩者皆不暴露 GameState。
type Controller interface {
	PlayerView(player *model.Player) (game.PlayerView, error)
	Submit(player *model.Player, input game.Input) error
}

// Heuristic 依固定優先級從 PlayerView 選取單一合法提交。
// random 僅於相同優先級候選間使用，沒有遊戲狀態、replay 或任何隱藏資訊欄位。
type Heuristic struct {
	random RandomSource
}

// NewHeuristic 建立使用指定亂數來源的啟發式 bot。
// 輸入為平手決策的亂數來源；輸出為 bot，nil 來源會在第一次需要平手決策時回傳錯誤而不提交行動。
func NewHeuristic(random RandomSource) *Heuristic {
	return &Heuristic{
		random: random,
	}
}

// Decide 從單一 PlayerView 選出附帶該 view revision 的合法 action 或 PendingChoice。
// 輸入只包含玩家可見資訊；輸出為可直接 Submit 的 Input，副作用僅可能在同分候選時消耗 injected random source。
func (bot *Heuristic) Decide(view game.PlayerView) (game.Input, error) {
	if view.Finished {
		return game.Input{}, fmt.Errorf("cannot decide for finished game")
	}
	if view.PendingChoice != nil {
		return bot.decidePendingChoice(view)
	}
	if len(view.LegalActions) == 0 {
		return game.Input{}, fmt.Errorf("no legal action in player view")
	}
	if len(view.LegalActions) > maxCandidatesPerDecision {
		return game.Input{}, fmt.Errorf("legal action count %d exceeds decision limit %d", len(view.LegalActions), maxCandidatesPerDecision)
	}
	return bot.decideAction(view)
}

// Run 反覆讀取最新 PlayerView 並提交 bot 決策，直到遊戲結束或達到提交上限。
// 輸入為正式控制器、bot 玩家與正整數上限；輸出為結束或錯誤，副作用是最多提交 limit 次且過期 revision 只會重新讀取視圖。
func (bot *Heuristic) Run(controller Controller, player *model.Player, limit int) error {
	if limit <= 0 {
		return fmt.Errorf("decision limit must be positive")
	}
	if controller == nil {
		return fmt.Errorf("controller is required")
	}
	seen := make(map[string]struct{}, limit)
	for attempts := 0; attempts < limit; {
		view, err := controller.PlayerView(player)
		if err != nil {
			return fmt.Errorf("get player view: %w", err)
		}
		if view.Finished {
			return nil
		}
		input, err := bot.Decide(view)
		if err != nil {
			return fmt.Errorf("decide: %w", err)
		}
		attempts++
		if err := controller.Submit(player, input); err != nil {
			if errors.Is(err, tcgErrors.ErrStaleRevision) {
				continue
			}
			return fmt.Errorf("submit bot decision: %w", err)
		}
		signature := viewSignature(view)
		if _, exists := seen[signature]; exists {
			return fmt.Errorf("no progress cycle after %d submissions", attempts)
		}
		seen[signature] = struct{}{}
	}
	return fmt.Errorf("decision limit %d reached", limit)
}

// decidePendingChoice 優先處理引擎要求的 PendingChoice，避免在選擇期間提交一般 action。
// 輸入為含 PendingChoice 的視圖；輸出為 choice Input 或可略過時的 pass action，副作用僅於多個等價選項時消耗亂數。
func (bot *Heuristic) decidePendingChoice(view game.PlayerView) (game.Input, error) {
	choice := view.PendingChoice
	if len(choice.Options) > maxCandidatesPerDecision {
		return game.Input{}, fmt.Errorf("pending choice count %d exceeds decision limit %d", len(choice.Options), maxCandidatesPerDecision)
	}
	if len(choice.Choices) > 0 {
		handle, err := bot.selectChoice(choice.Choices)
		if err != nil {
			return game.Input{}, err
		}
		return game.Input{
			Revision: view.Revision,
			Choice:   handle,
		}, nil
	}
	if len(choice.Options) > 0 {
		option, err := bot.selectHandle(choice.Options)
		if err != nil {
			return game.Input{}, fmt.Errorf("select pending choice: %w", err)
		}
		return game.Input{
			Revision: view.Revision,
			Choice:   option,
		}, nil
	}
	if choice.CanPass {
		for _, action := range view.LegalActions {
			if action.Kind == constants.ActionPass {
				return game.Input{
					Revision: view.Revision,
					Action:   action.Handle,
				}, nil
			}
		}
	}
	return game.Input{}, fmt.Errorf("pending choice has no selectable option")
}

// decideAction 依攻擊、啟動、裝備、materialize、推進 phase、投降的固定優先級選出 action。
// 輸入為沒有 PendingChoice 的視圖；輸出為 action Input，副作用僅於最高優先級平手時消耗亂數。
func (bot *Heuristic) decideAction(view game.PlayerView) (game.Input, error) {
	bestPriority := view.LegalActions[0].HeuristicRank
	candidates := []game.ViewHandle{
		view.LegalActions[0].Handle,
	}
	for _, action := range view.LegalActions[1:] {
		priority := action.HeuristicRank
		switch {
		case priority < bestPriority:
			bestPriority = priority
			candidates = []game.ViewHandle{
				action.Handle,
			}
		case priority == bestPriority:
			candidates = append(candidates, action.Handle)
		}
	}
	handle, err := bot.selectHandle(candidates)
	if err != nil {
		return game.Input{}, fmt.Errorf("select action: %w", err)
	}
	input := game.Input{
		Revision: view.Revision,
		Action:   handle,
	}
	for _, action := range view.LegalActions {
		if action.Handle != handle {
			continue
		}
		if action.ReserveCost > len(action.ReserveOptions) {
			return game.Input{}, fmt.Errorf("reserve options = %d, want at least %d", len(action.ReserveOptions), action.ReserveCost)
		}
		for _, option := range action.ReserveOptions {
			if action.CardName != option.Name && (option.Name == "Red Hare, Unrivaled Stallion" || option.Name == "Duchess, Six of Hearts") {
				continue
			}
			input.Reserve = append(input.Reserve, option.Handle)
			if len(input.Reserve) == action.ReserveCost {
				return input, nil
			}
		}
		for _, option := range action.ReserveOptions {
			if containsHandle(input.Reserve, option.Handle) {
				continue
			}
			input.Reserve = append(input.Reserve, option.Handle)
			if len(input.Reserve) == action.ReserveCost {
				break
			}
		}
		break
	}
	return input, nil
}

// containsHandle 回傳 handles 是否含有 candidate，供 bot 避免將同一張 Reserve 卡重複放入提交。
// 輸入為已選 handles 與候選 handle；輸出為是否存在，無副作用。
func containsHandle(handles []game.ViewHandle, candidate game.ViewHandle) bool {
	for _, handle := range handles {
		if handle == candidate {
			return true
		}
	}
	return false
}

// selectChoice 從引擎提供的可見 choice rank 中保留最高優先級候選，再交由平手機制處理。
// 輸入為同一 PendingChoice 的可見選項；輸出為其中一個合法 handle，副作用只在最高優先級有平手時消耗亂數。
func (bot *Heuristic) selectChoice(choices []game.VisibleChoice) (game.ViewHandle, error) {
	bestPriority := choices[0].HeuristicRank
	candidates := []game.ViewHandle{
		choices[0].Handle,
	}
	for _, choice := range choices[1:] {
		switch {
		case choice.HeuristicRank < bestPriority:
			bestPriority = choice.HeuristicRank
			candidates = []game.ViewHandle{
				choice.Handle,
			}
		case choice.HeuristicRank == bestPriority:
			candidates = append(candidates, choice.Handle)
		}
	}
	return bot.selectHandle(candidates)
}

// selectHandle 從唯一候選直接回傳，或以注入亂數從同分候選選取。
// 輸入為已知合法且同優先級的 handles；輸出為其中一項，副作用是在多項時推進亂數來源一次。
func (bot *Heuristic) selectHandle(handles []game.ViewHandle) (game.ViewHandle, error) {
	if len(handles) == 0 {
		return "", fmt.Errorf("no candidate handle")
	}
	if len(handles) == 1 {
		return handles[0], nil
	}
	if bot.random == nil {
		return "", fmt.Errorf("random source is required for tied decisions")
	}
	return handles[bot.random.Uint64()%uint64(len(handles))], nil
}

// viewSignature 建立不含 revision、handle 或隱藏資訊的可見決策摘要以偵測循環。
// 輸入為 PlayerView；輸出為排序後的行動種類、卡名與待選狀態，無副作用。
func viewSignature(view game.PlayerView) string {
	actions := make([]string, 0, len(view.LegalActions))
	for _, action := range view.LegalActions {
		actions = append(actions, fmt.Sprintf("%s:%s:%d", action.Kind, action.CardName, action.HeuristicRank))
	}
	sort.Strings(actions)
	pending := "none"
	if view.PendingChoice != nil {
		pending = fmt.Sprintf("choice:%d:%t", len(view.PendingChoice.Options), view.PendingChoice.CanPass)
	}
	return pending + ";" + strings.Join(actions, ",")
}

// seededRandom 以固定的 SplitMix64 狀態產生可重現的平手決策亂數。
// state 是唯一可變欄位；每次 Uint64 輸出一個值並將 state 前進一次。
type seededRandom struct {
	state uint64
}

// NewSeededRandom 建立可注入 Heuristic 的確定性亂數來源。
// 輸入為任意 seed；輸出為獨立亂數來源，副作用僅存在於後續 Uint64 呼叫的內部 state 前進。
func NewSeededRandom(seed uint64) RandomSource {
	return &seededRandom{
		state: seed,
	}
}

// Uint64 以 SplitMix64 演算法產生下一個確定性亂數值。
// 輸入隱含目前 state；輸出為 uint64，副作用是 state 增加固定步長以供下一次呼叫。
func (random *seededRandom) Uint64() uint64 {
	random.state += 0x9e3779b97f4a7c15
	value := random.state
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}
