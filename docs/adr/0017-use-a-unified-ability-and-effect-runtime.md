# 使用統一的 Ability 與 Effect Runtime

所有 activated 與 triggered ability 共用同一條宣告、選擇、付款、入 Effects Stack 與結算流程。每次能力使用建立帶有 controller、source LKI、已選 mode／target 與 usage identity 的 Ability Instance；結算由 typed operation sequence 產生 zone move、counter、damage、draw、modifier 與 choice 事件。這取代依卡牌類型或關鍵字建立 Action 專用分支，確保 replay、state hash、trigger ordering 與 Player View 均沿用相同的 deterministic lifecycle。

Continuous effect 保留為中央 evaluator 的輸入：`get`、`gain`、`become` 建立目標快照的 instanced effect，`has`、`are`、`as long as` 建立每次求值的 static effect。Pending Choice 保存 continuation 且僅揭露 actor 可見的 View Handle；once-per-instance 使用 Ability Instance identity，因此 Object 離場再進場會建立新的追蹤生命週期。

## 曾考慮的方案

將 Cardistry 接在既有 Action 宣告旁邊較快，但會重複付款、選擇、LKI、Stack 與事件邏輯，並使後續 triggered ability、runtime copy、replacement 與 alternative cost 無法共用。預先為完整卡池建立 DSL 則超出固定 Support Set 的需求；因此採用 Go 中的通用 runtime 與 Support Set 驅動的 typed operations。
