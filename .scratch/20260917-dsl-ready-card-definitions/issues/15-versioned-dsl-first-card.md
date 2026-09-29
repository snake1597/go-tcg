# 15 — 版本化 DSL 編譯成相同中介表示

**What to build:** 卡牌作者能以版本化 DSL 編寫一張已驗證卡牌；玩家看到的合法操作、Player View 及 replay 與 Go 定義一致。

**Blocked by:** 14 — 由可執行定義產生 production 支援登錄。

**Status:** blocked

**Blocked reason:** 14 — definition-derived production support gate 尚未完成；在中介表示與 production registration 成為單一來源前不固定 DSL schema。

**Next implementation step:** 14 完成後，選一張已由 Go Definition 完整驗證的卡牌作唯一 DSL vertical slice，序列化既有中介表示並沿用相同 compiler validation 與 runtime；不得藉此新增 selector、expression、cost 或 operation kind。

- [ ] DSL 在 content loading 階段編譯一次成相同不可變中介表示；失敗不 fallback，resolution loop 不重新解析。
- [ ] schema 及 compiled content version 納入既有契約；不讀舊 schema，Ability Slot ID 在 Go 與 DSL 間穩定。
- [ ] compiler 測試拒絕未知 kind、selector、expression、zone、duration、reference、錯誤型別、無效／未使用欄位及不支援版本，並指出 Definition、Face、Slot 和欄位。
- [ ] DSL 卡牌的正常、非法及 replay scenario 通過；相關 Go 編寫 ADR 更新。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
