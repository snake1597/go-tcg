# 02 — 首張 Go 可執行定義與編譯驗證

**What to build:** Wonderlands Reign 的抽牌能力由 Card Definition 中不可變的 Ability Definition 執行，卡牌作者能在單一來源看見 Slot、條件與效果。

**Blocked by:** 共通規則操作 02 — 抽牌與抽空規則。

**Status:** completed

- [x] Go builder 產生經驗證的 typed 中介表示；穩定 Ability Slot ID 貫穿宣告、instance、ruling 與診斷。
- [x] 以 Player View 和 View Handle 啟動及結算，正確產生 Effects Stack、Game Event 與 replay；舊 Card ID 路徑已刪除。
- [x] 純資料 compiler 在建局前拒絕重複 ID、無效 Slot、未知 kind、錯誤 payload 或 reference，錯誤指出 Definition、Face、Slot 與欄位；編寫資料不進入可變 Game State。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。

2026-09-27：`ability:0mf1ug6yfi:front:cardistry-draw` 連結 [鎖定卡面與 once-per-instance ruling](../../../card/wonderlands-reign.json) 及 [抽牌與抽空規則](../../../rules/game-mechanics/game-mechanics-drawing-cards.md)。`TestCardistryMemoryPaymentIsDeclaredRecordedAndReplayed` 驗證 View Handle、Slot、Effects Stack、主牌組頂順序、私有可見事件與 replay；`TestCardistryDrawFromEmptyDeckLosesGame` 驗證抽空敗北。Go builder、compiler 錯誤由 `executable_definition_test.go` 驗證。品質門檻：`go test ./...`、`go test -race ./...` 與兩個 10 秒 fuzz 測試。
