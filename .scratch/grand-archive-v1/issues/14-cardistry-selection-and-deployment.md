# 14: 支援 Cardistry 的選擇、棄牌與部署互動

**What to build:** 讓 Three of Hearts、Three of Spades 與 Four of Hearts 透過相同 Cardistry 協定完成 mandatory discard、temporary life 與從 Memory 選擇合格 Ally 放到 Field。

**Blocked by:** 12.5: 建立統一 Ability 與 Effect Runtime; 13: 建立 Cardistry 核心並支援基本 Cardistry 卡牌.

**Status:** in-progress

- [x] 待選內容只以當前玩家可用的 View Handle 呈現，不暴露其他隱藏卡牌。
- [x] Mandatory discard 與 deployment 使用 typed operation sequence 和可重播 Pending Choice continuation；沒有可合法完成的情況下依法處理，不能留下半完成效果。
- [x] Four of Hearts 只提供符合 element、type 與 reserve cost 的 Memory 卡，入場後建立新的 Object lifetime。
- [ ] 每個選擇、zone move、modifier 與 resulting trigger 都可由 events、replay 與 state hash 重現。
