# 共通規則操作與未來 DSL 的調整計畫

## 目標

卡片、回合流程與付款流程以相同的規則操作改變遊戲狀態。卡片只宣告條件、選擇與效果的組合；Go runtime 執行可序列化的 typed operations。未來 DSL 應產生同一種操作資料，不另建一套規則引擎。本計畫不要求現在實作 DSL parser。

## 現況與候選共通能力

| 現有位置與行為 | 目標共通能力 | 必須保留的語意 |
| --- | --- | --- |
| `ability_runtime.go` 的 `drawCards`、`drawToMemory`；`turn.go` 的 `drawCardsWithDeckOut` | 抽牌／牌組頂移動，目的區域、數量與抽空規則為明確參數 | 三條路徑目前取牌方向、抽空勝負、追蹤資訊和事件格式不同；先依規則確認方向，再共用實作 |
| `discardCard`、`putInGraveyard`、`recollectMemory`、`deployAlly` | `move_card` 與具名的 `discard`、`recollect`、`deploy` 規則操作 | 來源與目的區域、卡牌所有者、進場觸發、公開資訊及失效時行為 |
| `banishObjectForAbility`、`activateDuchessThornes`、`activateSafeguardAmulet`、`activateSmokeBombs`、`copyDuchessAction` | `banish` 使用同一套物件／卡牌移動提交流程 | 場上物件移除、來源 LKI、事件、觸發與區域卡牌移動須一致 |
| `sacrificeForPepperedChef`、`commitActionDeclarationWithWeapon`、戰鬥後物件移墓地 | `sacrifice` 與 `destroy` 分開表達，共用底層場上物件離場流程 | 犧牲與被破壞的原因不同；死亡觸發、replacement 和事件不得因共用移區而混淆 |
| `choose`、`choose_zone_card`、`choose_memory_ally`、`choose_duchess_copy`、宣告選擇、替代費用選擇 | 具篩選條件與具名結果的 `choose_card`／`choose_object` | 可選數量、可略過、隱藏資訊、提交時重新驗證、無選項時的續行規則 |
| `suited_threshold_damage`、`damage` 的 `DistinctSuitedCostsDamage` 旗標 | 一般 `damage` 加上可序列化的數值表達式 | 數值求值時點、合法目標、replacement choice 與 continuation |
| `copy_duchess_action` | `copy_action` | 原牌放逐、免付費、複製來源識別、再選目標、略過與結算後清理 |
| `move` 的 `MoveSourceToGraveyard` 旗標、`put_ally_on_field`、Materialization 來源移動 | 明確的來源移動與入場操作 | Effects Stack 來源區、場上物件與玩家區域不是同一種容器；來源遺失時須有定義 |
| Action／Ally 宣告、Cardistry 付款、替代費用、Wield 付款 | 共通的付款驗證與原子提交流程 | 付款與效果分開；提交前重驗證，失敗不留下已移動的牌或物件 |
| 物件能力的卡 ID 分派與 `retarget_attack` | 以能力宣告資料組合通用條件、費用與操作；保留必要的規則專用操作 | 重導攻擊涉及戰鬥堆疊，不應硬塞進一般 `move_card` |

`zone_movement.go` 已有玩家區域間的純資料移動；`card_selection.go` 已有選牌條件與組合驗證。優先延伸這些既有能力。共通操作負責規則語意；單純的 slice 搬移只作為其內部實作。

## 操作資料的最小形狀

- **參照**：能力來源卡、來源物件、宣告目標與選擇結果使用明確參照。選擇結果可具名，例如 `sacrifice_target`；後續操作不再依賴覆寫單一 `Target`。
- **選擇條件**：區域、控制者／擁有者、類型、子類型、元素、費用範圍、排除來源等。從現有卡片真正需要的欄位開始擴充，不建立任意條件語言。
- **數值**：常數、已需要的加總／計數、門檻對照與簡單算術。每個欄位明確指定宣告時或結算時求值。
- **狀態改變**：`move_card`、`banish`、`sacrifice`、`discard`、`draw`、`damage`、`continuous_modifier` 等操作描述規則原因；中央流程負責驗證、replacement、事件、可見性與觸發。
- **續行**：選擇與 replacement 中斷時，保存可序列化的剩餘操作與具名結果；所有離開路徑都清理 runtime copy。

## 分層實作順序

1. **規則語意盤點**：先確認抽牌頂端與抽空結果、各類離場原因、公開事件、追蹤資訊及觸發時點。記錄目前行為與預期規則的差異，避免把偶然的實作差異固化為參數。
2. **共通移動提交**：擴充玩家區域移動，並提供場上物件及 Effects Stack 來源的離場流程。先讓棄牌、放逐、recollect、deploy 和現有付款路徑逐一使用；每完成一條路徑就移除對應舊實作。
3. **共通抽牌**：在規則語意確認後，統一牌組頂取牌、目的區域、抽空結果、追蹤與事件；分別驗證回合抽牌、能力抽牌、抽至 Memory。
4. **共通選擇與參照**：把現有選牌條件與 ability continuation 接起來，改用具名結果，涵蓋 Duchess、Chef、一般棄牌與替代費用。
5. **卡片操作組合**：以 `sacrifice` 加 modifier、數值表達式加 `damage`、`copy_action` 替換四個卡名／機制名 operation，並整理 `DistinctSuitedCostsDamage`、`MoveSourceToGraveyard` 這類單用途旗標。
6. **能力宣告整合**：將可共用的啟動條件、費用與效果從卡 ID 分派中移出；戰鬥、Materialization 等有獨立生命週期的規則保留其專用入口，但透過共通操作改變狀態。
7. **刪除舊路徑**：每層通過驗證後刪除被取代的函式與分派，不保留相容層。更新 operation 支援檢查與文件。

## 驗收條件

- 同一種區域移動與場上離場原因經同一個規則流程提交，不由卡片 handler 直接改 `Objects` 或區域 slice。
- 抽牌、棄牌、放逐、犧牲、傷害與選擇都可由可序列化的 typed operation 或明確的規則流程表達；卡片資料不需要 Go 閉包。
- 宣告和結算時的合法性、費用原子性、replacement、觸發、Player View、replay 與 state hash 行為正確。
- 每張受影響卡片有規則情境測試；跨入口的共通操作以整體流程測試驗證。
- 未支援的 operation 或參數在開局驗證時被拒絕。

參照：[ADR 0005](adr/0005-implement-card-behavior-in-go.md)、[ADR 0017](adr/0017-use-a-unified-ability-and-effect-runtime.md)、[既有引擎模式研究](research/card-game-engine-patterns.md)。
