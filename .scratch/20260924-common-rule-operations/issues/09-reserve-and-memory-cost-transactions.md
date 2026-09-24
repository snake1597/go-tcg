# 09 — Reserve Cost 與 Memory Cost 原子提交

**What to build:** 打出 Action、Ally 與啟動 Cardistry 時，Reserve Cost 與 Memory Cost 在完整合法性驗證後一次提交。

**Blocked by:** 01 — 棄牌與一般區域移動。

**Status:** ready-for-agent

- [ ] Action、Ally 與 Cardistry 的付款均保留各自合法來源、明確選牌及隨機支付規則，來源卡與付款卡進入正確區域。
- [ ] 宣告取消、非法目標、非法付款或過期 View Handle 不改變權威 Game State、PRNG cursor、事件與觸發。
- [ ] 成功付款僅提交一次，並建立正確 Source Card、Ability Instance 或 Stack Item。
- [ ] Player View、Game Event、replay 與 state hash 覆蓋各付款入口；被取代的重複付款移區路徑已移除。
