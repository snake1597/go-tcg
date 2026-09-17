# 08: 以 Fiery Interference 支援 fast action 與恢復禁止

**What to build:** 讓回合玩家或非回合玩家在合法 Opportunity 中打出 Fiery Interference，造成傷害並在條件成立時禁止對應 Champion controller 於本回合恢復生命。

**Blocked by:** 07: 以 Blazing Throw 建立第一條完整卡牌宣告與 Stack 路徑.

**Status:** completed

- [x] 合法行動列舉正確反映 fast timing、目標與付款條件。
- [x] 非回合玩家成功打出後仍保有 Opportunity，直到其讓過。
- [x] 傷害及 recover prohibition 以同一因果鏈提交，並在正確回合邊界失效。
- [x] 非法、過期或失效目標不留下費用以外的不合法結果，replay 可重現完整互動。
