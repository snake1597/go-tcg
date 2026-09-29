# 10 — 複製能力與 runtime 身分

**What to build:** Duchess 的選牌與 Action 複製由通用定義及操作完成；玩家能合法選擇、略過或結算複製品。

**Blocked by:** 04 — 具名選擇與可序列化續行；05 — Activated ability 的費用與使用限制。

**Related foundation:** 共通規則操作 08 — Action 複製 operation 與 runtime copy 生命週期；本票是 Duchess Card Definition 遷移的唯一完成 owner。

**Status:** completed

**Completion evidence:** 正式 Game Module Interface scenario 使已複製 Action 的合法目標在結算前離場，並驗證 fizzle、原牌仍位於 Banishment、已付 Cardistry 費用不退回、runtime copy 清理、Game Event、state hash、replay 與卡牌守恆。同一組 production path 與測試亦完成共通規則操作 08 的剩餘驗收。

- [x] 複製品具獨立 Ability Instance identity、正確 Source Ref／LKI，且不再支付原 Action 費用。
- [x] 正式介面驗證原牌移動、目標失效、pass、續行與 source cleanup；卡牌守恆及 replay hash 確定。
- [x] Duchess 專用 operation kind 與 Card ID 分支已刪除。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。

2026-09-29：`af1d973` 將 Duchess 的墓地 Action 選擇與免費複製改由 executable Card Definition、具名 reference 及通用 `copy_action` operation 執行。現有正式介面測試已覆蓋合法 selector、獨立 runtime copy identity、Source Ref／LKI、pass、免費目標宣告、原牌放逐、正常完成與略過後的複製品清理及 replay。品質門檻 `go test ./...`、`go test -race ./...` 與兩個 10 秒 fuzz 測試均通過；完成本票前仍須補上複製 Action 目標於結算前失效，以及中斷路徑清理／事件與卡牌守恆的直接 scenario。

2026-09-29：新增 Duchess 複製 Blazing Throw 後由 Fiery Interference 先移除目標的正式介面 scenario；複製傷害正確 fizzle，原牌與 Cardistry 付款留在 Banishment，runtime copy 在中斷結算後清理，且 `ability/banish` Game Event、最終 state hash、replay 驗證與卡牌數守恆均有直接斷言。`go test ./...`、`go test -race ./...` 與兩個 10 秒 fuzz 測試全數通過，本票完成。
