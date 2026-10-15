# 09 — Reserve Cost 與 Memory Cost 原子提交

**What to build:** 打出 Action、Ally 與啟動 Cardistry 時，Reserve Cost 與 Memory Cost 在完整合法性驗證後一次提交。

**Scope:** 本票收斂目前 `Game.Submit` 引導選擇流程中的付款與取消原子性；獨立 Begin／Update／Confirm／Cancel Draft RPC 與完整隔離 Draft 模型不在本票範圍。

**Blocked by:** None — can continue from the existing narrow Reserve and Memory payment semantics.

**Related foundation:** 01 — 玩家卡牌區域移動、棄牌與 Memory recollection。其餘玩家區域入口的共通化不阻擋本票逐一收斂既有付款路徑。

**Related migration:** DSL-ready Card Definitions 05 — 代表性 Cardistry Definition 使用本票的 Memory payment 能力，但不取代本票對 Action、Ally 與其他入口的共通化責任。

**Status:** completed

**Delivered slice:** DSL-ready Card Definitions 05 已證明 Cardistry Memory payment 在完整合法性驗證後只提交一次，拒絕或過期輸入不改變權威狀態。剩餘工作是完成 Action、Ally 及其他 Reserve／Memory payment 入口的共通化與直接移區清理。

**Delivered slice:** DSL-ready Card Definitions 13 已讓 Viridian Protective Trinket 的 permission Slot 只向共用 activation Reserve cost query 輸出費用增量；來源離場、非 Water、控制者自身或非控制者回合都不加費，沒有卡牌專用付款 mutation。其餘入口仍依本票 Next implementation step 收斂。

**Completed implementation:** Action 宣告以可還原的 declaration snapshot 支援零正式副作用取消，來源、Reserve 與 Stack Item 只在最終合法後提交；Ally、Cardistry 與 Materialization 共用 declaration transaction 回滾邊界，非法付款與 stale handle 不改變權威狀態。

- [x] Action、Ally 與 Cardistry 的付款均保留各自合法來源、明確選牌及隨機支付規則，來源卡與付款卡進入正確區域。
- [x] 宣告取消、非法目標、非法付款或過期 View Handle 不改變權威 Game State、PRNG cursor、事件與觸發。
- [x] 成功付款僅提交一次，並建立正確 Source Card、Ability Instance 或 Stack Item。
- [x] Player View、Game Event、replay 與 state hash 覆蓋各付款入口；被取代的重複付款移區路徑已移除。
