# 11 — 物件能力共通宣告整合

**What to build:** 讓 Object activated abilities 可共用同一套條件驗證、Cost Payment 與 operation 提交流程。剩餘 Object、Weapon、Regalia 的 Card Definition 遷移與 production support metadata 由 DSL-ready Card Definitions 13、14 驗收。

**Blocked by:** 02 — 抽牌與抽空規則；03 — Source Card 移動及 Ally 部署；04 — 場上 Object 放逐；06 — Sacrifice operation 與規則原因；07 — 數值表達式與傷害；08 — Action 複製 operation 與 runtime copy 生命週期；10 — 替代費用與 Wield 付款。

**Related migration:** DSL-ready Card Definitions 13 — 剩餘 Object 與 Regalia Slots；14 — definition-derived production support gate。

**Status:** blocked

**Blocked reason:** 02、03、04、06、07、08、10 尚未全部完成；本票是共通 Object ability declaration 的整合與清理 gate。

**Next implementation step:** 所有前置 Common issues 完成後，依 Card ID、卡名 operation、單用途旗標與 direct Game State mutation 四類搜尋剩餘路徑；逐條以已完成 operation contract 取代並保留既有正式介面 scenario，最後執行完整 Standard Game、Support Set、replay 與 mirror-game gate。

- [ ] 現有可組合的 Object ability 使用共通啟動條件、費用與 operation 提交語意；玩家可用行動與原有規則一致。
- [ ] Combat、Level Up 與 Materialization 等獨立生命週期保留專用入口，並透過共通規則操作提交可共用的狀態變化。
- [ ] Operation validation 在開局前拒絕未知 operation、錯誤參照與無效參數；production support metadata 的衍生與完整性由 DSL-ready Card Definitions 14 負責。
- [ ] 共通 runtime 中的卡名專用 operation、單用途旗標與重複直接狀態修改均被移除，不保留相容層；個別 Card ID 行為路徑由對應 DSL migration 清理。
- [ ] 受影響卡牌的 Player View、Game Event、Replacement Pipeline、replay、state hash 與完整 Standard Game 驗證通過。
