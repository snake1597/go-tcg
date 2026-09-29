# 01 — 玩家卡牌區域移動、棄牌與 Memory recollection

**What to build:** 建立玩家卡牌區域間的共通移動提交語意，先以棄牌及 Memory recollection 證明 Card Instance、owner、順序、Knowledge State 與事件結果正確。Field Object、Source Card 與 Stack Item 的生命週期不屬於本票。

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

**Next implementation step:** 先盤點 Hand、Memory、Graveyard 與 Banishment 的直接 slice mutation，定義只處理玩家卡牌區域的 typed move request 與原子提交；以一個 discard scenario 和一個 Memory recollection scenario 證明順序、owner、Knowledge State、Game Event、Player View、replay 及非法輸入無副作用，再逐一替換相同語意的剩餘入口。

- [ ] 從 Hand 及規則允許的其他來源區域棄牌時，選擇與結算指向同一張 Card Instance，目的區域與 owner 正確。
- [ ] Memory recollection 依規則順序將卡牌移回 Hand，並保留正確的追蹤權、Player View 與 Knowledge State。
- [ ] 來源已失效時依既定規則處理，且不會憑空增加或遺失卡牌。
- [ ] Player View、Game Event、Knowledge State 與 replay 反映正確的棄牌結果。
- [ ] 玩家提交非法或過期 View Handle 時，權威 Game State 不改變。
- [ ] 完成路徑不再直接使用被取代的玩家卡牌區域移動實作；仍可進行 Standard Game。
