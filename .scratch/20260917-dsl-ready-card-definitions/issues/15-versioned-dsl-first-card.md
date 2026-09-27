# 15 — 版本化 DSL 編譯成相同中介表示

**What to build:** 卡牌作者能以版本化 DSL 編寫一張已驗證卡牌；玩家看到的合法操作、Player View 及 replay 與 Go 定義一致。

**Blocked by:** 14 — 由可執行定義產生 production 支援登錄。

**Status:** ready-for-agent

- [ ] DSL 在 content loading 階段編譯一次成相同不可變中介表示；失敗不 fallback，resolution loop 不重新解析。
- [ ] schema 及 compiled content version 納入既有契約；不讀舊 schema，Ability Slot ID 在 Go 與 DSL 間穩定。
- [ ] compiler 測試拒絕未知 kind、selector、expression、zone、duration、reference、錯誤型別、無效／未使用欄位及不支援版本，並指出 Definition、Face、Slot 和欄位。
- [ ] DSL 卡牌的正常、非法及 replay scenario 通過；相關 Go 編寫 ADR 更新。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
