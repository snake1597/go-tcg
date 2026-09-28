# 08 — Replacement ability 定義

**What to build:** Infernal Vessel 的替代行為由 Ability Definition 進入既有 Replacement Pipeline，玩家可選順序並得到正確結果。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證。

**Status:** completed

- [x] 定義明列 event filter 與 transformation；每次替代後重算候選。
- [x] 正式介面驗證多候選順序、prevention、Pending Choice 後續行與 cause／event history。
- [x] 非法選擇不改變權威狀態；Infernal Vessel 的 Card ID replacement 分支已刪除。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。

2026-09-28：`ability:vgWgu1DUYv:front:recover-reduce` 以 replacement Ability Definition 宣告 recover event filter 與減少 3 的 transformation；Replacement Pipeline 從場上物件的已編譯定義重建候選，不再檢查 Infernal Vessel Card ID。`TestCompileInfernalVesselReplacementDefinition` 驗證 compiler 的 slot、filter、transformation 與錯誤資料拒絕；`TestReplacementChoiceUsesAffectedControllerAndResumesAfterRecalculation` 經 `Submit` 驗證 Pending Choice、非法 handle 的 state hash／replay 不變、候選重算與完整 cause chain。品質門檻：`go test ./...`、`go test -race ./...` 與兩個 10 秒 fuzz 測試。
