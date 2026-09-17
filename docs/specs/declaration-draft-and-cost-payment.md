# Declaration Draft 與費用支付規格

## 狀態

已確認需求，尚未實作。

## 背景

玩家從卡牌所在的 Hand、Graveyard、Banishment 或 Material Deck 區域視窗選擇可用卡牌後，需要依規則逐步選擇參數、模式、目標與支付來源。Reserve Cost 可以由 Hand 中移入 Memory 的卡牌及 Field 上的 Reservable Object 等來源共同支付；Memory Cost 必須先處理 Floating Memory 等非隨機來源，再由引擎從 Memory 公平隨機 banish 剩餘數量。

現有 `LegalAction` 只為部分行動提供 Reserve／Floating Memory 選項，部分宣告則先把內部 declaration 與 `PendingChoice` 寫入正式狀態。這不足以支援可取消、多步驟、Confirm 前無副作用的 Vue 流程。

本規格實作 [ADR 0004](../adr/0004-use-guided-choices-and-atomic-actions.md) 已決定的 `DeclarationTransaction`：宣告位於隔離候選狀態，正式 `GameState` 只在完整合法後原子提交。

## 目標

- Vue 以引擎回傳的順序呈現 Parameters、Modes、Targets、Pay Cost 與 Confirm。
- Confirm 前所有選擇只修改隔離 Declaration Draft，不修改正式 Game State。
- 支援 Hand reserve、Field Reservable、Floating Memory 及其他規則註冊的支付來源。
- 一般 Memory 的隨機支付只顯示數量，不讓玩家指定卡牌。
- 引擎提供卡牌與支付來源的結構化不可用原因。
- 完成宣告後只產生一次正式 Input、一次原子狀態轉移及一筆 Replay step。

## 非目標

- Vue 不計算 cost、合法目標、支付總額或規則步驟順序。
- 不把所有支付機制硬編碼成 Hand、Field、Graveyard 三種固定分支。
- 不讓玩家選擇一般 Memory 中將被隨機 banish 的牌。
- 不將 Declaration Draft 寫入 canonical Replay。
- 不保留目前各 action kind 自訂 payload 的前端 fallback。

## Declaration Draft 邊界

Declaration Draft 由伺服器維護，與正式 Game State 隔離。它至少保存：

- 不透明 draft handle
- 建立時的 state revision
- actor 與來源 action handle
- 隔離候選狀態與候選區域移動
- 已選參數、模式、目標與支付來源
- 已固定的 cost snapshot
- 候選事件與 trigger buffer
- 隔離的 PRNG cursor
- 目前步驟及是否仍可取消

Draft handle 只用於更新、取消或提交草稿，不是 Card、Object、action 或 choice handle。

### 開始草稿

Vue 使用當前 revision 與合法 action handle 開始 declaration。引擎重新驗證 action 後建立 draft，並回傳第一個尚未完成的步驟。若該 action 不需宣告流程，引擎可以直接回傳 Confirm 步驟，但仍不得在使用者確認前修改正式狀態。

若同一張卡有多個玩家可主動選擇的合法 action，Vue 先顯示引擎提供的 Action Chooser；只有一個合法 action 時直接開始。強制或同時觸發的能力不出現在此 Chooser，其順序由引擎依規則推進；只有規則明確要求玩家決定順序時，才建立正式 `PendingChoice`。

### 更新草稿

Vue 每次只回答目前步驟。引擎驗證選擇、更新隔離候選狀態並回傳下一步；Vue 不可以自行跳步或提前提交後續欄位。

### 取消草稿

規則仍允許撤回時，Cancel 丟棄整個 draft。取消不建立事件、Replay step、trigger、KnowledgeState 變更或正式 PRNG 消耗。

### 提交草稿

Confirm 會重新驗證 actor、base revision、所有選擇、cost snapshot 與最終合法性。成功時一次提交候選狀態、正式 PRNG cursor、事件與 trigger；失敗時整份丟棄且正式狀態不變。

正式 state revision 只要離開 draft 的 base revision，該 draft 就失效。Vue 必須清除 Wizard 並以最新 PlayerView 重新開始，不自動重送。

純區域瀏覽視窗不受此規則關閉，而是依 [`vue.md` 的區域生命週期](../vue.md#區域視窗與-card-inspector-生命週期)持續同步最新 PlayerView。Draft 與瀏覽視窗不得共用同一份凍結資料或失效生命週期。

## 步驟模型

```go
type DeclarationStepKind string

const (
	DeclarationStepParameters DeclarationStepKind = "parameters"
	DeclarationStepModes      DeclarationStepKind = "modes"
	DeclarationStepTargets    DeclarationStepKind = "targets"
	DeclarationStepPayCost    DeclarationStepKind = "pay_cost"
	DeclarationStepConfirm    DeclarationStepKind = "confirm"
)
```

引擎只回傳當前步驟及該步驟合法選項。正常順序是 `Parameters → Modes → Targets → Pay Cost → Confirm`，未需要的步驟省略。若付款依法新增模式或目標，引擎依規則回到新增步驟，但不得重算已固定的 cost。

`PendingChoice` 只用於正式規則流程已提交後、效果結算或其他無法自行前進的外部決定；尚未提交的 Declaration Draft 不冒充正式 `PendingChoice`。

## Pay Cost View

Pay Cost 是 Declaration Wizard 的一個引擎驅動步驟。

```go
type CostKind string

const (
	CostReserve CostKind = "reserve"
	CostMemory  CostKind = "memory"
)

type PaymentConsequence string

const (
	PaymentMoveToMemory PaymentConsequence = "move_to_memory"
	PaymentRestObject   PaymentConsequence = "rest_object"
	PaymentBanishCard   PaymentConsequence = "banish_card"
)

type VisiblePaymentSource struct {
	Handle       ViewHandle          `json:"handle"`
	Kind         string              `json:"kind"`
	Card         *CardRef            `json:"card,omitempty"`
	Value        int                 `json:"value"`
	Consequence  PaymentConsequence  `json:"consequence"`
	Availability ActionAvailability  `json:"availability"`
}

type PayCostStep struct {
	Kind                 CostKind              `json:"kind"`
	Required             int                   `json:"required"`
	Selected             int                   `json:"selected"`
	Sources              []VisiblePaymentSource `json:"sources"`
	RandomMemoryRequired int                   `json:"random_memory_required,omitempty"`
}
```

欄位名稱可在實作時依既有 Go type 收斂，但語意不得改變：每個來源由不透明 handle 識別，明確表示支付值、使用後果與可用性；Vue 不從 zone 或卡牌文字推算。

### Reserve Cost

- Hand 卡牌來源的 consequence 是面朝下移入 Memory。
- Reservable Object 的 consequence 是 Rest 該 Object，通常支付 1 點。
- 引擎可以提供其他已註冊的合法來源與其支付值。
- 所有非隨機支付來源都由玩家自行選擇；即使只有唯一合法組合，Vue 也不得預先選取。
- 來源組合必須符合引擎要求；Vue 只依 `Selected` 與 `Required` 顯示進度。
- 不能使用來源卡本身、已 Rested Object、失去 Reservable 的 Object 或其他不合法來源。

### Memory Cost

- Floating Memory 與其他非隨機來源列為可選 `VisiblePaymentSource`。
- 非隨機來源依法先於一般 Memory 支付。
- `RandomMemoryRequired` 顯示 Confirm 時仍需由引擎從 Memory 公平隨機 banish 的數量。
- 一般 Memory 卡牌不列為可選來源，也不提供即將被選中的預覽。
- Confirm 前不得消耗正式 PRNG，亦不得把隔離抽樣結果投影給玩家。
- Confirm 成功並正式 banish 後，支付結果才透過公開 Banishment 與玩家可見 EventBatch 揭露。
- 同一次隨機 Memory 支付的所有 banish 屬於同一 Event Batch 並同時公開；Vue 可以錯開動畫，但不得把它們呈現成可插入行動的逐張事件。

## 結構化可用性

所有可顯示但不可選的卡牌、行動或支付來源都由引擎提供結構化原因。

```go
type AvailabilityReasonCode string

type ActionAvailability struct {
	Available bool                   `json:"available"`
	Reason    AvailabilityReasonCode `json:"reason,omitempty"`
	Args      map[string]int         `json:"args,omitempty"`
}
```

Reason code 是穩定、可翻譯的協定值；`Args` 只攜帶格式化所需的非敏感數值。首版至少涵蓋：

- wrong-phase
- no-opportunity
- insufficient-cost
- no-legal-target
- requirement-not-met
- once-per-turn-used
- source-is-action-card
- source-is-rested
- source-not-reservable
- source-no-longer-present
- stale-revision

Vue 依 reason code 顯示本地化文字，不自行重新判定規則。未知 reason code 是前後端版本不一致的合約錯誤，不退回由前端猜測。

## Vue 互動

- 每個區域視窗只顯示實際位於該 Hand、Graveyard、Banishment 或 Material Deck 的卡牌，不提供跨區域的 `Playable Cards` 彈窗。
- 純瀏覽區域視窗只使用 `Close`；`Done` 只用於完成引擎要求的當前多選步驟，`Confirm` 只用於原子提交整個 Declaration，`Cancel` 只用於丟棄尚未提交的 Draft。
- 點擊不可用卡只鎖定 Card Inspector 並顯示結構化原因，不建立 draft。
- 點擊只有一個合法 action 的卡，以其 action handle 開始 draft。
- 點擊具有多個主動合法 action 的卡，先由 Action Chooser 選定 action handle；自動觸發能力不列入。
- 不需前置步驟時直接顯示 Pay Cost；需要參數、模式或目標時依引擎順序顯示。
- Pay Cost 中所有來源可開啟 Card Inspector，但只有 `Available == true` 的來源可選。
- Vue 不預選支付來源，也不提供 `Auto Pay`；玩家必須明確選擇每個非隨機來源。
- 選擇符合引擎條件後才進入 Confirm。
- Confirm 畫面摘要來源、模式、目標、費用來源及可公開的隨機付款數量。
- Cancel 返回發起宣告的原區域視窗；若引擎標示不可取消，UI 不顯示或停用 Cancel。
- 隨機 Memory 支付提交成功後，Vue 一次取得整批公開結果並更新 Banishment；動畫可跳過，但播放期間不接受下一個玩家輸入。

## 權威性與安全限制

- Game Module 是步驟、選項、cost、來源、結果與 reason code 的唯一權威。
- Draft 只能讀取 actor 的 Player View 權限，不得因預覽洩漏對手隱藏資訊。
- Draft 更新不增加正式 state revision。
- 同一座位同時只允許一個 active declaration draft；開始新 draft 前必須取消舊 draft。
- Draft 具有明確逾時與連線中斷清理機制，但逾時不得改變正式狀態。
- 任何拒絕、取消、過期或最終非法流程都不得改變 Game State、正式 PRNG cursor、事件、trigger 或 Replay。

## 驗收條件

1. 不需前置選擇的合法卡牌會直接進入 Pay Cost。
2. 需要參數、模式或目標的牌依引擎順序完成後才進入 Pay Cost。
3. Reserve Cost 可以同時使用 Hand 卡牌與 Field Reservable Object。
4. 使用 Reservable Object 會在正式提交時 Rest 該 Object，取消草稿則不改變它。
5. Memory Cost 先接受合法非隨機來源，再於 Confirm 時由引擎隨機支付剩餘數量。
6. Vue 無法選擇一般 Memory 卡牌，也無法在 Confirm 前預知隨機結果；成功 banish 後才從公開投影得知結果。
7. 不可用卡牌與支付來源帶有結構化 reason code。
8. state revision 改變會使 draft 失效，且不自動重送。
9. Cancel、拒絕、逾時與非法提交不留下任何正式副作用。
10. 成功 Confirm 只產生一次原子狀態轉移與一筆 canonical Replay step。
11. 同一次隨機 Memory 支付以單一 Event Batch 同時揭露，動畫不建立額外 Opportunity。

## 必要測試

- 無前置步驟與完整多步驟 declaration。
- 付款新增模式／目標但不重算固定 cost。
- 純 Hand reserve、純 Reservable 及混合支付。
- Rested、失去 Reservable、離場與來源卡自身的不可用原因。
- Floating Memory 加隨機 Memory 的支付順序。
- Confirm 前不推進正式 PRNG，取消後相同正式狀態保持相同 hash。
- stale revision、跨玩家 draft handle、重複來源與過度／不足支付。
- 成功、取消、逾時及最終非法流程的 Replay step 數量。
- reason code JSON round-trip 與前端未知版本錯誤。
