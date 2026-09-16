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

// defaultSubmissionLimit 限制正式 CLI 單局可接受的提交數，避免無進展對局無限執行。
const defaultSubmissionLimit = 1000

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
	submissionLimit := flags.Int(
		"submission-limit",
		defaultSubmissionLimit,
		"單局允許的最大成功提交數",
	)
	if parseErr := flags.Parse(arguments); parseErr != nil {
		return fmt.Errorf("parse flags: %w", parseErr)
	}
	if *replayPath == "" {
		return fmt.Errorf("--replay-out is required")
	}
	if *submissionLimit <= 0 {
		return fmt.Errorf("--submission-limit must be positive")
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
	acceptedSubmissions := 0
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
		if acceptedSubmissions >= *submissionLimit {
			diagnostic := fmt.Sprintf("submission limit %d reached", *submissionLimit)
			fmt.Fprintf(
				output,
				"發布診斷：seed=%d step=%d diagnostic=%s replay=%s state_hash=%s\n",
				*seed,
				acceptedSubmissions,
				diagnostic,
				*replayPath,
				match.StateHash(),
			)
			return fmt.Errorf("%s", diagnostic)
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
			afterView, afterViewErr := match.PlayerView(player)
			if afterViewErr != nil {
				return fmt.Errorf("get bot view after submission: %w", afterViewErr)
			}
			acceptedSubmissions++
			writeSubmissionSummary(
				output,
				"bot "+player.UID,
				player,
				view,
				botInput,
				afterView,
			)
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
			continue
		}
		afterView, afterViewErr := match.PlayerView(player)
		if afterViewErr != nil {
			return fmt.Errorf("get player view after submission: %w", afterViewErr)
		}
		acceptedSubmissions++
		writeSubmissionSummary(
			output,
			player.UID,
			player,
			view,
			selected,
			afterView,
		)
	}
}

// writeSubmissionSummary 將成功提交描述為玩家行動、待決選擇或 Opportunity pass，並說明公開的堆疊與 Opportunity 轉移結果。
// 輸入為輸出串流、顯示名稱、提交玩家、提交前後 PlayerView 與已接受 Input；輸出為零值，副作用僅為寫入公開流程文字。
func writeSubmissionSummary(output io.Writer, actorLabel string, actor *model.Player, before game.PlayerView, input game.Input, after game.PlayerView) {
	if before.PendingChoice != nil {
		if input.Choice != "" {
			fmt.Fprintf(output, "%s 已完成一項選擇。\n", actorLabel)
		} else {
			fmt.Fprintf(output, "%s 略過一項選擇。\n", actorLabel)
		}
		writeRetainedOpportunity(output, actor, after)
		return
	}
	action, found := legalActionForInput(before, input)
	if !found {
		fmt.Fprintf(output, "%s 已完成一次提交。\n", actorLabel)
		return
	}
	if action.Kind != constants.ActionPass {
		if action.Kind == constants.ActionActivate && action.CardName != "" {
			fmt.Fprintf(output, "%s 啟動 %s。\n", actorLabel, action.CardName)
		} else {
			label := actionLabel(action)
			fmt.Fprintf(
				output,
				"%s 選擇 %s。\n",
				actorLabel,
				label,
			)
		}
		writeRetainedOpportunity(output, actor, after)
		return
	}
	if !sameVisibleEffectsStack(before.EffectsStack, after.EffectsStack) {
		top := before.EffectsStack[len(before.EffectsStack)-1]
		fmt.Fprintf(output, "%s 選擇 pass。\n", actorLabel)
		fmt.Fprintf(output, "雙方連續 pass，結算堆疊頂端：%s。\n", top.SourceName)
		if len(after.EffectsStack) > 0 {
			fmt.Fprintf(
				output,
				"Effects Stack 尚有 %d 個項目；重新開啟回應窗口，Opportunity 交給 %s。\n",
				len(after.EffectsStack),
				playerName(after.OpportunityHolder),
			)
		} else {
			fmt.Fprintln(output, "Effects Stack 已清空。")
		}
		return
	}
	fmt.Fprintf(
		output,
		"%s 選擇 pass，Opportunity 移交給 %s。\n",
		actorLabel,
		playerName(after.OpportunityHolder),
	)
}

// legalActionForInput 從提交前 PlayerView 找出已接受 action handle 的公開描述。
// 輸入為提交前 PlayerView 與 Input；輸出為相符 LegalAction 及是否存在，無副作用。
func legalActionForInput(view game.PlayerView, input game.Input) (game.LegalAction, bool) {
	for _, action := range view.LegalActions {
		if action.Handle == input.Action {
			return action, true
		}
	}
	var emptyAction game.LegalAction
	return emptyAction, false
}

// sameVisibleEffectsStack 比較兩個公開 Effects Stack 投影的順序、種類、來源與控制者。
// 輸入為提交前後的公開 stack items；輸出為玩家可見內容是否相同，無副作用且不讀取引擎內部身分。
func sameVisibleEffectsStack(first []game.VisibleEffectStackItem, second []game.VisibleEffectStackItem) bool {
	if len(first) != len(second) {
		return false
	}
	for index, firstItem := range first {
		secondItem := second[index]
		if firstItem.Kind != secondItem.Kind || firstItem.SourceName != secondItem.SourceName || playerName(firstItem.Controller) != playerName(secondItem.Controller) {
			return false
		}
	}
	return true
}

// writeRetainedOpportunity 說明成功行動或選擇後 Opportunity 是否仍由原提交玩家持有。
// 輸入為輸出串流、提交玩家與提交後 PlayerView；輸出為零值，副作用僅在 Opportunity 保留時寫入提示。
func writeRetainedOpportunity(output io.Writer, actor *model.Player, after game.PlayerView) {
	if actor == nil || after.OpportunityHolder == nil || actor.UID != after.OpportunityHolder.UID {
		return
	}
	fmt.Fprintf(output, "Opportunity 仍由 %s 持有。\n", actor.UID)
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
	var emptyInput game.Input
	for {
		fmt.Fprint(output, "請輸入編號：")
		if !scanner.Scan() {
			if scannerErr := scanner.Err(); scannerErr != nil {
				return emptyInput, fmt.Errorf("read input: %w", scannerErr)
			}
			return emptyInput, fmt.Errorf("%w", io.EOF)
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
		reserve, reserveErr := ReadReserve(scanner, output, action)
		if reserveErr != nil {
			return emptyInput, fmt.Errorf("read reserve: %w", reserveErr)
		}
		floatingMemory, floatingMemoryErr := readFloatingMemory(scanner, output, action)
		if floatingMemoryErr != nil {
			return emptyInput, fmt.Errorf("read floating memory: %w", floatingMemoryErr)
		}
		return game.Input{
			Revision:       view.Revision,
			Action:         action.Handle,
			Reserve:        reserve,
			FloatingMemory: floatingMemory,
		}, nil
	}
}

// ReadReserve 讓呼叫端以引擎提供的編號選擇恰好數量的手牌支付 reserve cost。
// 輸入為 scanner、輸出與合法 action；輸出為 reserve handles 或讀取錯誤，副作用僅為提示與無效輸入訊息。
func ReadReserve(scanner *bufio.Scanner, output io.Writer, action game.LegalAction) ([]game.ViewHandle, error) {
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
		inputText := scanner.Text()
		selected, valid := parseNumberedHandles(
			inputText,
			action.ReserveOptions,
		)
		if !valid || len(selected) != action.ReserveCost {
			fmt.Fprintln(output, "無效編號，請重新輸入。")
			continue
		}
		return selected, nil
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
		selected, valid := parseNumberedHandles(
			text,
			action.FloatingMemoryOptions,
		)
		if valid {
			return selected, nil
		}
		fmt.Fprintln(output, "無效編號，請重新輸入。")
	}
}

// parseNumberedHandles 將逗號分隔的十進位選項解析成不重複的 PlayerView handles。
// 輸入為使用者文字與可見卡牌選項；輸出為依輸入順序排列的 handles 與是否合法，無副作用。
func parseNumberedHandles(text string, options []game.VisibleCard) ([]game.ViewHandle, bool) {
	parts := strings.Split(
		strings.TrimSpace(text),
		",",
	)
	selectedCapacity := len(parts)
	selected := make([]game.ViewHandle, 0, selectedCapacity)
	seen := make(map[game.ViewHandle]bool, selectedCapacity)
	optionCount := len(options)
	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		index, err := strconv.Atoi(trimmedPart)
		if err != nil || index <= 0 || index > optionCount {
			return nil, false
		}
		handle := options[index-1].Handle
		if seen[handle] {
			return nil, false
		}
		seen[handle] = true
		selected = append(selected, handle)
	}
	return selected, true
}

// renderView 將單一玩家依法可見的狀態、最近事件與引擎合法選項寫成終端畫面。
// 輸入為輸出串流、檢視玩家與 PlayerView；輸出為零值，副作用僅為寫入文字，不計算或補造遊戲規則選項。
func renderView(output io.Writer, player *model.Player, view game.PlayerView) {
	fmt.Fprintf(output, "\n玩家：%s　回合：%d　phase：%s　Opportunity：%s\n", player.UID, view.TurnNumber, view.Phase, playerName(view.OpportunityHolder))
	fmt.Fprintf(output, "回合玩家：%s　等待輸入：%s\n", playerName(view.TurnPlayer), playerName(view.DecisionPlayer))
	fmt.Fprintln(output, "Champion：")
	for _, champion := range view.Champions {
		remainingLife := max(0, champion.Life-champion.Damage)
		fmt.Fprintf(output, "- %s：%s %d/%d（生命上限 %d，傷害 %d），rested=%t，taunt=%t\n", playerName(champion.Owner), champion.CardName, champion.Power, remainingLife, champion.Life, champion.Damage, champion.Rested, champion.Taunt)
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
	if len(view.EffectsStack) > 0 {
		fmt.Fprintln(output, "Effects Stack 尚未清空；只能啟動 Fast 卡或 pass。")
	}
	fmt.Fprintln(output, "最近事件：")
	start := max(0, len(view.VisibleEvents)-recentEventLimit)
	for _, event := range view.VisibleEvents[start:] {
		if event.Kind == "draw" {
			fmt.Fprintf(output, "- 抽牌階段：抽到 %s\n", event.CardName)
			continue
		}
		if event.Kind == "materialize-banish-memory" {
			fmt.Fprintf(output, "- 物質化付款：隨機放逐 %s\n", event.CardName)
			continue
		}
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
		label := actionLabel(action)
		fmt.Fprintf(output, "%d. %s\n", index+1, label)
		if action.Kind == constants.ActionPass && len(view.EffectsStack) > 0 {
			top := view.EffectsStack[len(view.EffectsStack)-1]
			fmt.Fprintf(
				output,
				"   └─ 將 Opportunity 交給下一位玩家；若所有玩家連續 pass，結算堆疊頂端：%s。\n",
				top.SourceName,
			)
		}
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
