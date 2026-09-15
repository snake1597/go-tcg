package productioncli

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go-tcg/internal/bot"
	"go-tcg/internal/constants"
	"go-tcg/internal/game"
	"go-tcg/internal/model"
	"io"
	"os"
	"strconv"
	"strings"
)

const recentEventLimit = 8

// Run 建立固定 Standard 單局，以引擎 PlayerView 讓真人與 bot 依序提交行動並輸出私人 canonical replay。
// 輸入為命令列參數、終端輸入輸出與 repository root；輸出為流程或 I/O 錯誤，副作用為建立 replay、提交真人與 bot 輸入，並將 bot 私密視圖保留在終端外。
func Run(arguments []string, input io.Reader, output io.Writer, repositoryRoot string) (err error) {
	flags := flag.NewFlagSet(
		"production-cli",
		flag.ContinueOnError,
	)
	flags.SetOutput(output)
	seed := flags.Uint64(
		"seed",
		1,
		"固定亂數種子",
	)
	replayPath := flags.String(
		"replay-out",
		"",
		"私人 canonical replay 輸出位置",
	)
	if parseErr := flags.Parse(arguments); parseErr != nil {
		return fmt.Errorf("parse flags: %w", parseErr)
	}
	if *replayPath == "" {
		return fmt.Errorf("--replay-out is required")
	}

	match, setupErr := game.NewStandardGame(game.StandardGameConfig{
		Players: [2]*model.Player{
			model.PlayerOne,
			model.PlayerTwo,
		},
		RepositoryRoot: repositoryRoot,
		Seed:           *seed,
	})
	if setupErr != nil {
		var gateError *game.GateError
		if errors.As(setupErr, &gateError) {
			writeGateDiagnostics(output, gateError.Diagnostics)
		}
		return fmt.Errorf("start standard game: %w", setupErr)
	}
	opponent := bot.NewHeuristic(
		bot.NewSeededRandom(*seed),
	)
	writeReplayResult := true
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("scheduler failure: %v", recovered)
			writeReplayResult = false
		}
		if !writeReplayResult {
			fmt.Fprintln(output, "scheduler failure 後未輸出 replay，以避免保留未驗證的半提交狀態。")
			return
		}
		if writeErr := writeReplay(*replayPath, match.Replay()); writeErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf("write replay: %w", writeErr),
			)
		}
	}()

	fmt.Fprintln(output, "隱私警告：canonical replay 可能包含完整隱藏資訊，僅供私人診斷使用，不可公開分享。")
	fmt.Fprintln(output, "真人：player-1　bot：player-2")
	scanner := bufio.NewScanner(input)
	for {
		player, playerErr := currentDecisionPlayer(match)
		if playerErr != nil {
			return fmt.Errorf("get decision player: %w", playerErr)
		}
		view, viewErr := match.PlayerView(player)
		if viewErr != nil {
			return fmt.Errorf("get player view: %w", viewErr)
		}
		if view.Finished {
			renderView(output, model.PlayerOne, view)
			if view.Diagnostic != "" {
				fmt.Fprintf(output, "遊戲診斷：%s\n", view.Diagnostic)
			}
			return nil
		}
		if player == model.PlayerTwo {
			botInput, decideErr := opponent.Decide(view)
			if decideErr != nil {
				return fmt.Errorf("decide bot action: %w", decideErr)
			}
			beforeHash := match.StateHash()
			if submitErr := match.Submit(player, botInput); submitErr != nil {
				if match.StateHash() != beforeHash {
					return fmt.Errorf("rejected bot input changed game state: %w", submitErr)
				}
				return fmt.Errorf("submit bot action: %w", submitErr)
			}
			fmt.Fprintln(output, "bot player-2 已提交行動。")
			continue
		}
		renderView(output, player, view)
		selected, selectErr := readSelection(scanner, output, view)
		if selectErr != nil {
			return fmt.Errorf("read selection: %w", selectErr)
		}
		beforeHash := match.StateHash()
		if submitErr := match.Submit(player, selected); submitErr != nil {
			if match.StateHash() != beforeHash {
				return fmt.Errorf("rejected input changed game state: %w", submitErr)
			}
			fmt.Fprintf(output, "輸入被引擎拒絕：%v\n", submitErr)
		}
	}
}

// currentDecisionPlayer 從公開 PlayerView 的 DecisionPlayer 取得目前應操作終端的玩家。
// 輸入為正式 Game；輸出為待輸入玩家或排程錯誤，無副作用且不讀取其他玩家的私人投影。
func currentDecisionPlayer(match *game.Game) (*model.Player, error) {
	view, err := match.PlayerView(model.PlayerOne)
	if err != nil {
		return nil, fmt.Errorf("get scheduler view: %w", err)
	}
	if view.Finished {
		return model.PlayerOne, nil
	}
	if view.DecisionPlayer == nil {
		return nil, fmt.Errorf("scheduler failure: game is not finished but has no decision player")
	}
	return view.DecisionPlayer, nil
}

// readSelection 讀取一個十進位編號，並將其轉為目前 PlayerView 中的合法 action 或 choice。
// 輸入為 scanner、輸出與單一玩家視圖；輸出為可直接 Submit 的輸入或 EOF／讀取錯誤，副作用僅為提示與無效輸入訊息。
func readSelection(scanner *bufio.Scanner, output io.Writer, view game.PlayerView) (game.Input, error) {
	for {
		fmt.Fprint(output, "請輸入編號：")
		if !scanner.Scan() {
			if scannerErr := scanner.Err(); scannerErr != nil {
				return game.Input{}, fmt.Errorf("read input: %w", scannerErr)
			}
			return game.Input{}, fmt.Errorf("%w", io.EOF)
		}
		selected, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || selected <= 0 {
			fmt.Fprintln(output, "無效編號，請重新輸入。")
			continue
		}
		if view.PendingChoice != nil {
			if selected > len(view.PendingChoice.Options) {
				if selected == len(view.PendingChoice.Options)+1 && view.PendingChoice.CanPass {
					for _, action := range view.LegalActions {
						if action.Kind == constants.ActionPass {
							return game.Input{
								Revision: view.Revision,
								Action:   action.Handle,
							}, nil
						}
					}
				}
				fmt.Fprintln(output, "無效編號，請重新輸入。")
				continue
			}
			return game.Input{
				Revision: view.Revision,
				Choice:   view.PendingChoice.Options[selected-1],
			}, nil
		}
		if selected > len(view.LegalActions) {
			fmt.Fprintln(output, "無效編號，請重新輸入。")
			continue
		}
		action := view.LegalActions[selected-1]
		reserve, reserveErr := readReserve(scanner, output, action)
		if reserveErr != nil {
			return game.Input{}, fmt.Errorf("read reserve: %w", reserveErr)
		}
		floatingMemory, floatingMemoryErr := readFloatingMemory(scanner, output, action)
		if floatingMemoryErr != nil {
			return game.Input{}, fmt.Errorf("read floating memory: %w", floatingMemoryErr)
		}
		return game.Input{
			Revision:       view.Revision,
			Action:         action.Handle,
			Reserve:        reserve,
			FloatingMemory: floatingMemory,
		}, nil
	}
}

// readReserve 讓玩家以引擎提供的編號選擇恰好數量的手牌支付 reserve cost。
// 輸入為 scanner、輸出與合法 action；輸出為 reserve handles 或讀取錯誤，副作用僅為提示與無效輸入訊息。
func readReserve(scanner *bufio.Scanner, output io.Writer, action game.LegalAction) ([]game.ViewHandle, error) {
	if action.ReserveCost == 0 {
		return nil, nil
	}
	fmt.Fprintf(output, "選擇 %d 張 Reserve（以逗號分隔）：\n", action.ReserveCost)
	for index, card := range action.ReserveOptions {
		fmt.Fprintf(output, "%d. %s\n", index+1, card.Name)
	}
	for {
		fmt.Fprint(output, "請輸入編號：")
		if !scanner.Scan() {
			if scannerErr := scanner.Err(); scannerErr != nil {
				return nil, fmt.Errorf("read reserve: %w", scannerErr)
			}
			return nil, fmt.Errorf("%w", io.EOF)
		}
		parts := strings.Split(strings.TrimSpace(scanner.Text()), ",")
		if len(parts) != action.ReserveCost {
			fmt.Fprintln(output, "無效編號，請重新輸入。")
			continue
		}
		selected := make([]game.ViewHandle, 0, action.ReserveCost)
		seen := make(map[game.ViewHandle]struct{}, action.ReserveCost)
		valid := true
		for _, part := range parts {
			index, parseErr := strconv.Atoi(strings.TrimSpace(part))
			if parseErr != nil || index <= 0 || index > len(action.ReserveOptions) {
				valid = false
				break
			}
			handle := action.ReserveOptions[index-1].Handle
			if _, exists := seen[handle]; exists {
				valid = false
				break
			}
			seen[handle] = struct{}{}
			selected = append(selected, handle)
		}
		if valid {
			return selected, nil
		}
		fmt.Fprintln(output, "無效編號，請重新輸入。")
	}
}

// readFloatingMemory 讓玩家以引擎提供的編號選項複選 Cardistry Floating Memory，空白代表不選。
// 輸入為 scanner、輸出與合法 action；輸出為選取的 handle 或讀取錯誤，副作用僅為寫入提示與無效輸入訊息。
func readFloatingMemory(scanner *bufio.Scanner, output io.Writer, action game.LegalAction) ([]game.ViewHandle, error) {
	if len(action.FloatingMemoryOptions) == 0 {
		return nil, nil
	}
	fmt.Fprintln(output, "可選 Floating Memory（可複選，以逗號分隔；直接 Enter 不使用）：")
	for index, card := range action.FloatingMemoryOptions {
		fmt.Fprintf(output, "%d. %s\n", index+1, card.Name)
	}
	for {
		fmt.Fprint(output, "請輸入編號：")
		if !scanner.Scan() {
			if scannerErr := scanner.Err(); scannerErr != nil {
				return nil, fmt.Errorf("read floating memory: %w", scannerErr)
			}
			return nil, fmt.Errorf("%w", io.EOF)
		}
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			return nil, nil
		}
		parts := strings.Split(text, ",")
		selected := make([]game.ViewHandle, 0, len(parts))
		seen := make(map[game.ViewHandle]struct{}, len(parts))
		valid := true
		for _, part := range parts {
			index, parseErr := strconv.Atoi(strings.TrimSpace(part))
			if parseErr != nil || index <= 0 || index > len(action.FloatingMemoryOptions) {
				valid = false
				break
			}
			handle := action.FloatingMemoryOptions[index-1].Handle
			if _, exists := seen[handle]; exists {
				valid = false
				break
			}
			seen[handle] = struct{}{}
			selected = append(selected, handle)
		}
		if valid {
			return selected, nil
		}
		fmt.Fprintln(output, "無效編號，請重新輸入。")
	}
}

// renderView 將單一玩家依法可見的狀態、最近事件與引擎合法選項寫成終端畫面。
// 輸入為輸出串流、檢視玩家與 PlayerView；輸出為零值，副作用僅為寫入文字，不計算或補造遊戲規則選項。
func renderView(output io.Writer, player *model.Player, view game.PlayerView) {
	fmt.Fprintf(output, "\n玩家：%s　回合：%d　phase：%s　Opportunity：%s\n", player.UID, view.TurnNumber, view.Phase, playerName(view.OpportunityHolder))
	fmt.Fprintf(output, "回合玩家：%s　等待輸入：%s\n", playerName(view.TurnPlayer), playerName(view.DecisionPlayer))
	fmt.Fprintln(output, "Champion：")
	for _, champion := range view.Champions {
		fmt.Fprintf(output, "- %s：%s %d/%d，傷害 %d，rested=%t，taunt=%t\n", playerName(champion.Owner), champion.CardName, champion.Power, champion.Life, champion.Damage, champion.Rested, champion.Taunt)
	}
	fmt.Fprintln(output, "自己手牌：")
	for _, card := range view.Hand {
		fmt.Fprintf(output, "- %s\n", card.Name)
	}
	fmt.Fprintln(output, "Field：")
	for _, object := range view.Field {
		fmt.Fprintf(output, "- %s：%s（%s，rested=%t，傷害 %d）\n", playerName(object.Owner), object.CardName, strings.Join(object.Types, "/"), object.Rested, object.Damage)
	}
	fmt.Fprintln(output, "Effects Stack（底 → 頂）：")
	for _, item := range view.EffectsStack {
		fmt.Fprintf(output, "- %s：%s（控制者 %s）\n", item.Kind, item.SourceName, playerName(item.Controller))
	}
	fmt.Fprintln(output, "最近事件：")
	start := max(0, len(view.VisibleEvents)-recentEventLimit)
	for _, event := range view.VisibleEvents[start:] {
		fmt.Fprintf(output, "- %s：%s\n", event.Kind, event.CardName)
	}
	if view.PendingChoice != nil {
		fmt.Fprintln(output, "可選選擇：")
		for index, option := range view.PendingChoice.Options {
			label := "選項"
			for _, choice := range view.PendingChoice.Choices {
				if choice.Handle == option && choice.CardName != "" {
					label = choice.CardName
					break
				}
			}
			fmt.Fprintf(output, "%d. %s\n", index+1, label)
		}
		if view.PendingChoice.CanPass {
			fmt.Fprintf(output, "%d. pass\n", len(view.PendingChoice.Options)+1)
		}
		return
	}
	fmt.Fprintln(output, "可選行動：")
	for index, action := range view.LegalActions {
		fmt.Fprintf(output, "%d. %s\n", index+1, actionLabel(action))
	}
	if view.Finished {
		fmt.Fprintf(output, "結果：%s 獲勝\n", playerName(view.Winner))
	}
}

// actionLabel 產生引擎已驗證合法 action 的人類可讀標籤。
// 輸入為 LegalAction；輸出為動作種類與可見卡牌名稱，無副作用且不接觸 action handle 或遊戲狀態。
func actionLabel(action game.LegalAction) string {
	label := string(action.Kind)
	if action.CardName != "" {
		return label + "：" + action.CardName
	}
	return label
}

// playerName 將可選玩家指標轉為終端安全顯示文字。
// 輸入為玩家或 nil；輸出為 UID 或「無」，無副作用。
func playerName(player *model.Player) string {
	if player == nil {
		return "無"
	}
	return player.UID
}

// writeGateDiagnostics 逐行輸出 Support Set gate 的可行動缺項診斷。
// 輸入為輸出串流與診斷清單；輸出為零值，副作用僅為寫入文字。
func writeGateDiagnostics(output io.Writer, diagnostics []game.GateDiagnostic) {
	fmt.Fprintln(output, "Standard 開局 gate 失敗：")
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(output, "- %s %s：%s\n", diagnostic.Kind, diagnostic.ID, diagnostic.Reason)
	}
}

// writeReplay 將 canonical replay 以只限擁有者讀寫的檔案權限寫入指定位置。
// 輸入為檔案路徑與 replay；輸出為序列化或 I/O 錯誤，副作用為覆寫指定 replay 檔案。
func writeReplay(path string, replay game.Replay) error {
	encoded, err := json.MarshalIndent(
		replay,
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf("marshal canonical replay: %w", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}
