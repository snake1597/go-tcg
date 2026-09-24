# 01 — 棄牌與一般區域移動

**What to build:** 玩家選擇一張可棄的卡牌後，該卡從指定來源區域進入 Graveyard；同一套區域移動語意可供後續規則流程使用。

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] 從 Hand 及規則允許的其他來源區域棄牌時，選擇與結算指向同一張 Card Instance，目的區域與 owner 正確。
- [ ] 來源已失效時依既定規則處理，且不會憑空增加或遺失卡牌。
- [ ] Player View、Game Event、Knowledge State 與 replay 反映正確的棄牌結果。
- [ ] 玩家提交非法或過期 View Handle 時，權威 Game State 不改變。
- [ ] 完成路徑不再直接使用被取代的棄牌移區實作；仍可進行 Standard Game。
