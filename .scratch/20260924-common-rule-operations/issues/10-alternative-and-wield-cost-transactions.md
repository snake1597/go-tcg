# 10 — 替代費用與 Wield 付款

**What to build:** 玩家選擇替代費用或 Wield 付款時，沿用共通的選擇驗證與原子提交保障。

**Blocked by:** 09 — Reserve Cost 與 Memory Cost 原子提交。

**Status:** ready-for-agent

- [ ] 替代費用的多張選牌在完整合法後一次付款；保留一般付款與替代付款的玩家選擇。
- [ ] Wield 從 Memory 支付的卡牌依規則進入 Banishment，付款張數與區域內容正確。
- [ ] 非法、取消或過期輸入不改變 Game State、PRNG cursor、Game Event 或 Knowledge State。
- [ ] 成功付款不重複，後續效果依既有 Effects Stack 與 Opportunity 流程執行。
- [ ] Player View 與 replay 覆蓋兩種付款路徑；被取代的直接區域修改已移除。
