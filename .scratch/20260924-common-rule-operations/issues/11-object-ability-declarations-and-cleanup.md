# 11 — 物件能力宣告整合與舊分支清理

**What to build:** 現有可共用的 Object activated abilities 由明確的條件、Cost Payment 與效果組合執行，並完成舊分支清理。

**Blocked by:** 02 — 抽牌與抽空規則；03 — Source Card 移動及 Ally 部署；04 — 場上 Object 放逐；06 — 犧牲 Ally 與 Weapon；07 — 數值表達式與傷害；08 — Duchess 選牌與複製 Action；10 — 替代費用與 Wield 付款。

**Status:** ready-for-agent

- [ ] 現有可組合的 Object ability 啟動條件、費用與效果不再散落於 Card ID 分派；玩家可用行動與原有規則一致。
- [ ] Combat、Level Up 與 Materialization 等獨立生命週期保留專用入口，並透過共通規則操作提交可共用的狀態變化。
- [ ] Support Set 在開局前拒絕未知 operation、錯誤參照與無效參數，且錯誤提供可定位的能力資訊。
- [ ] 所有已遷移路徑的卡名專用 operation、單用途旗標與重複直接狀態修改均被移除，不保留相容層。
- [ ] 受影響卡牌的 Player View、Game Event、Replacement Pipeline、replay、state hash 與完整 Standard Game 驗證通過。
