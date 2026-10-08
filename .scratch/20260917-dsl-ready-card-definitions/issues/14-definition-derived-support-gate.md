# 14 — 由可執行定義產生 production 支援登錄

**What to build:** 建局時，只有成功編譯且依賴及裁定完整的 Ability Slots 才能標記 Supported。

**Blocked by:** 10 — 複製能力與 runtime 身分；11 — 遷移其餘 Cardistry Ability Slots；12 — 遷移其餘 Action Ability Slots；13 — 遷移其餘 Object 與 Regalia Ability Slots。

**Status:** completed

**Completed:** 固定牌組所有 rules-bearing Card Definitions 已完成編譯；建局由 compiled definitions 衍生並驗證 CardFace、Ability Slot、mechanism、operation、content dependency、ruling reference 與 production status。Arthur、Noire、Peppered Chef、Heated Vengeance、Rouge、Spirit of Fire、Tonoris 的舊 runtime Card ID 分支已由 typed definitions 與共用 runtime 取代。

**Quality gate:** `go test ./...`、`go test -race ./...`、`go vet ./...`、`FuzzRejectedInputPreservesStateHash` 10 秒與 `FuzzReplayHashDeterminism` 10 秒均通過；GitNexus change detection 已檢查建局、Ability Runtime、trigger、evaluator、materialization、replay 與 mirrored Standard Game 受影響流程。

- [x] production metadata 從可執行定義衍生，並驗證 Face、Slot、mechanism、operation、content dependency 及 ruling reference。
- [x] 重複 ID、缺漏或循環依賴、未知原語、Unsupported 或 Needs Ruling 內容均在建局前拒絕。
- [x] 已取代的手寫 registry shape 與平行 Card ID 路徑已移除；固定 Support Set 可完整遊玩。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
