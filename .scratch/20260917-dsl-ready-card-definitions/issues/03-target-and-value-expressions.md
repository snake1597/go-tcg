# 03 — 目標與數值表達式能力

**What to build:** Straight Flare 由通用 selector、reference 與 value expression 決定合法目標和傷害數值。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證；共通規則操作 07 — 數值表達式與傷害。

**Status:** completed

- [x] Selector 只產生合法且 visibility-safe 的選項，順序確定；compiler 驗證 reference 型別及此能力所需的數值運算。
- [x] 正常傷害、非法目標、結算時目標失效、Replacement Pipeline 與已付費用保留均由正式介面 scenario 驗證。
- [x] Straight Flare 的卡名專用效果欄位與分支已刪除。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。

2026-09-27：`ability:28bjn8g50v:front:action-damage` 以 Go 編寫資料宣告公開 unit selector、declared-target reference 與 `1 + distinct printed reserve costs(controlled Suited)`；結算時求值並使用既有傷害 replacement pipeline。`straight_flare_definition_test.go` 驗證 compiler、非法目標、Fiery Interference 回應使目標失效後付款保留、Safeguard Amulet 與 replay；既有 `TestStraightFlareCountsDistinctPrintedSuitedReserveCosts` 驗證正常傷害。非戰鬥傷害提交後執行狀態檢查，讓回應擊倒的 Ally 正式離場。品質門檻：`go test ./...`、`go test -race ./...` 與兩個 10 秒 fuzz 測試。
