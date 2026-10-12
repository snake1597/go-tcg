# 01 — 玩家卡牌區域移動、棄牌與 Memory recollection

**What to build:** 建立玩家卡牌區域間的共通移動提交語意，先以棄牌及 Memory recollection 證明 Card Instance、owner、順序、Knowledge State 與事件結果正確。Field Object、Source Card 與 Stack Item 的生命週期不屬於本票。

**Blocked by:** None — can start immediately.

**Status:** completed

**Next implementation step:** 先盤點 Hand、Memory、Graveyard 與 Banishment 的直接 slice mutation，定義只處理玩家卡牌區域的 typed move request 與原子提交；以一個 discard scenario 和一個 Memory recollection scenario 證明順序、owner、Knowledge State、Game Event、Player View、replay 及非法輸入無副作用，再逐一替換相同語意的剩餘入口。

- [x] 從 Hand 及規則允許的其他來源區域棄牌時，選擇與結算指向同一張 Card Instance，目的區域與 owner 正確。
- [x] Memory recollection 依規則順序將卡牌移回 Hand，並保留正確的追蹤權、Player View 與 Knowledge State。
- [x] 來源已失效時依既定規則處理，且不會憑空增加或遺失卡牌。
- [x] Player View、Game Event、Knowledge State 與 replay 反映正確的棄牌結果。
- [x] 玩家提交非法或過期 View Handle 時，權威 Game State 不改變。
- [x] 完成路徑不再直接使用被取代的玩家卡牌區域移動實作；仍可進行 Standard Game。

**Delivered slice:** `discardCard` 與 `recollectMemory` 已改由同一批次提交器驗證 owner、來源區域與 Card Instance，並原子寫入 zone、tracking、Game Event 與可見事件。直接修改區域的 Reserve／Memory payment、抽牌與其他生命週期，保留給各自具不同隨機、付款或抽空語意的後續 issue；本票沒有將其納入 discard／recollection 移動語意。
