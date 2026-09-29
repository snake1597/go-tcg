# 08 — Action 複製 operation 與 runtime copy 生命週期

**What to build:** 建立可由 Card Definition 使用的 Action 複製 operation；它接收已驗證的具名 Card reference，處理原牌移動、免付費建立獨立能力、目標宣告與 runtime copy 清理。Duchess Card Definition 的端到端遷移由 DSL-ready Card Definitions 10 驗收。

**Blocked by:** None — can continue from the existing end-to-end slice using the narrow zone-movement and named-choice capabilities.

**Related foundation:** 01 — 玩家卡牌區域移動、棄牌與 Memory recollection；05 — 具名選擇與卡牌篩選。這兩張 umbrella issue 的其餘入口共通化不阻擋本票剩餘驗收。

**Related migration:** DSL-ready Card Definitions 10 — Duchess 是本 operation 的代表性 vertical slice，也是該卡牌遷移的唯一完成 owner。

**Status:** completed

**Completion evidence:** 與 DSL-ready Card Definitions 10 共用正式介面 scenario：先合法複製有目標的 Action，再於結算前使目標離場，驗證 fizzle、原牌 Banishment、runtime copy cleanup、Game Event、state hash、replay、卡牌守恆與中斷 cleanup。

- [x] Operation 只接受符合已編譯 selector 且提交時仍有效的 Action reference；不合格或已離開來源區域的卡不能被複製。
- [x] 選定後原 Card Instance 進入 Banishment，複製能力不再次支付原 Action 費用，並具有獨立 Ability Instance 與正確來源資訊。
- [x] 玩家能依規則選目標或略過；複製品在完成、略過及中斷後均被清理。
- [x] Player View、Game Event、replay 與 state hash 驗證原牌及複製品生命週期。
- [x] Operation contract、payload validation 與 runtime lifecycle 不含 Duchess Card ID 或卡名判斷。

2026-09-29：`af1d973` 已交付具名 Graveyard Action reference、通用 `copy_action` payload validation、原牌放逐、免付費建立獨立 Ability Instance、可選目標，以及正常完成與略過後的 runtime copy cleanup。Duchess 僅作為 Definition vertical slice，operation contract 與 runtime 不依賴其 Card ID。品質門檻 `go test ./...`、`go test -race ./...` 與兩個 10 秒 fuzz 測試均通過；剩餘工作是補上目標失效與中斷 cleanup，並直接驗證 Game Event、state hash 及卡牌守恆。

2026-09-29：共用正式介面 scenario 已覆蓋複製品被 Fast Action 中斷、已選目標離場後 fizzle、原牌放逐、付款保留、runtime copy cleanup、`ability/banish` Game Event、最終 state hash、replay 與卡牌數守恆。`go test ./...`、`go test -race ./...` 與兩個 10 秒 fuzz 測試全數通過，本票完成。
