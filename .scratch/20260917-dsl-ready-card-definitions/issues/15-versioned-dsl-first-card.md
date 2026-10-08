# 15 — 版本化 DSL 編譯成相同中介表示

**What to build:** 卡牌作者能以版本化 DSL 編寫一張已驗證卡牌；玩家看到的合法操作、Player View 及 replay 與 Go 定義一致。

**Blocked by:** 14 — 由可執行定義產生 production 支援登錄。

**Status:** completed

**Completed:** Wonderland's Reign 已成為唯一 versioned DSL vertical slice；schema version 1 JSON 在 content loading 與 replay definition rebuild 時嚴格編譯成既有 immutable IR，失敗不 fallback。Player View、合法／非法 Cardistry、state hash 與 replay 沿用相同 runtime，Ability Slot ID 與既有 Go Definition 完全一致。

**Quality gate:** `go test ./...`、`go test -race ./...`、`go vet ./...`、`FuzzRejectedInputPreservesStateHash` 10 秒與 `FuzzReplayHashDeterminism` 10 秒均通過；GitNexus change detection 已檢查 content loading、建局、Player View、replay 與 Standard Game 受影響流程。

- [x] DSL 在 content loading 階段編譯一次成相同不可變中介表示；失敗不 fallback，resolution loop 不重新解析。
- [x] schema 及 compiled content version 納入既有契約；不讀舊 schema，Ability Slot ID 在 Go 與 DSL 間穩定。
- [x] compiler 測試拒絕未知 kind、selector、expression、zone、duration、reference、錯誤型別、無效／未使用欄位及不支援版本，並指出 Definition、Face、Slot 和欄位。
- [x] DSL 卡牌的正常、非法及 replay scenario 通過；相關 Go 編寫 ADR 更新。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
