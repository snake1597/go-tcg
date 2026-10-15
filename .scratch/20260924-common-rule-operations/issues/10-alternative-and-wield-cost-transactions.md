# 10 — 替代費用與 Wield 付款

**What to build:** 玩家選擇替代費用或 Wield 付款時，沿用共通的選擇驗證與原子提交保障。

**Blocked by:** None — can continue from the existing alternative-cost and Wield payment semantics.

**Related foundation:** 09 — Reserve Cost 與 Memory Cost 原子提交。兩票共用 Declaration Transaction 契約，但剩餘入口可各自以 vertical slice 收斂。

**Related migration:** DSL-ready Card Definitions 09 — Verita 負責替代費用的 Card Definition 遷移；本票仍負責替代費用與 Wield 跨入口的共通交易保障。

**Status:** completed

**Delivered slice:** DSL-ready Card Definitions 09 已證明 Verita 的一般與替代付款可在同一交易中取消、失敗或一次提交。剩餘工作是完成其他替代費用與 Wield payment 入口的共通化及直接區域修改清理。

**Delivered slice:** DSL-ready Card Definitions 13 已讓 Bulwark Sword 的 class-matched static Slot 向中央 evaluator 輸出 `ReserveCostDelta: 2`，共用 Wield 合法性與付款流程消費該 derived cost；不足付款為非法，成功時仍由既有 Wield transaction 放逐兩張 Reserve。其餘 Wield 交易清理仍由本票負責。

**Next implementation step:** 先讓 Wield Memory payment 使用既有 Declaration Transaction，保持隨機或明確付款選擇、Banishment 目的地、取消／失敗 rollback 與成功一次提交；完成後再盤點是否仍有其他替代費用入口未共用同一 contract。

**Completed implementation:** 替代費用與 Wield 均以共通 Declaration Transaction 提交。Wield 改用既有 Memory 付款規則，隨機付款卡會放逐並記錄 `banish-wield-reserve` 事件；公開 Player View 到 replay 的流程已覆蓋兩點 Wield 付款，原本直接修改 Memory／Banishment 的路徑已移除。

- [x] 替代費用的多張選牌在完整合法後一次付款；保留一般付款與替代付款的玩家選擇。
- [x] Wield 從 Memory 支付的卡牌依規則進入 Banishment，付款張數與區域內容正確。
- [x] 非法、取消或過期輸入不改變 Game State、PRNG cursor、Game Event 或 Knowledge State。
- [x] 成功付款不重複，後續效果依既有 Effects Stack 與 Opportunity 流程執行。
- [x] Player View 與 replay 覆蓋兩種付款路徑；被取代的直接區域修改已移除。
