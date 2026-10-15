package game

import "fmt"

// commitDeclarationTransaction 以權威狀態快照保護最終宣告與付款，只有完整成功才保留變更。
// 輸入為會完成驗證、付款與建立 Stack Item 的提交函式；輸出為提交錯誤或 nil；失敗時還原區域、事件、PRNG 與所有其他權威欄位。
func (g *Game) commitDeclarationTransaction(commit func() error) error {
	previousState := cloneGameState(
		g.state,
	)
	if err := commit(); err != nil {
		g.state = previousState
		return fmt.Errorf("commit declaration transaction: %w", err)
	}
	return nil
}
