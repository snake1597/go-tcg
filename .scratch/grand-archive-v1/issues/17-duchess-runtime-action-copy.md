# 17: 以 Duchess 支援 runtime action copy

**What to build:** 讓 Duchess 的 Cardistry ability 從墓地選擇合格 fire action、banish 來源、建立獨立 runtime copy，並讓玩家決定是否免費 activation 該 copy。

**Blocked by:** 07: 以 Blazing Throw 建立第一條完整卡牌宣告與 Stack 路徑; 08: 以 Fiery Interference 支援 fast action 與恢復禁止; 09: 以 Straight Flare 建立 Suited 查詢與動態傷害; 13: 建立 Cardistry 核心並支援基本 Cardistry 卡牌.

**Status:** ready-for-agent

- [ ] 合法候選精確涵蓋牌組內所有符合 element、type 與 reserve cost 的 action。
- [ ] Copy 沿用來源 Card Face 的 resolution behavior，但具有獨立 Object／Stack Item 身分與一次性追蹤。
- [ ] 玩家可選擇不啟動；選擇啟動時仍完成 mode、target 與合法性驗證但不支付被免除的 activation cost。
- [ ] 來源離開墓地、copy 目標失效及結算完成後的生命週期均有 replay/hash 回歸案例。
