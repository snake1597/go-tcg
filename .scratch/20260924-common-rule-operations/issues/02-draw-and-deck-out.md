# 02 — 抽牌與抽空規則

**What to build:** 回合抽牌、能力抽牌與抽至 Memory 依正確的 Main Deck 順序和抽空規則執行，並產生一致的可見結果。

**Blocked by:** 01 — 玩家卡牌區域移動、棄牌與 Memory recollection。

**Status:** completed

- [x] 先依鎖定規則確認 Main Deck 的牌組頂、抽牌次序、牌組不足與抽空勝負語意；有歧義時記錄 Needs Ruling，不以現有實作差異代替規則。
- [x] 回合抽牌、能力抽牌與抽至 Memory 各自移動正確張數至正確區域，並依規則處理空牌組。
- [x] Event Batch、Game Event、Knowledge State、Player View 及追蹤權與抽牌結果一致。
- [x] 相同 seed 與玩家輸入可重播且 state hash 確定；完成後移除被取代的重複抽牌路徑。

**Delivered slice:** 回合、開局、能力抽牌與抽至 Memory 均共用同一個由 Main Deck 頂端逐張取牌的抽空提交器；每張牌保留獨立 Game Event，而同一指令只建立一個 Event Batch。抽牌玩家取得追蹤權與私有 Player View 事件；抽空時立即結束對局。已移除能力專用的重複抽牌路徑，並以開局、回合、能力、Memory、抽空與 replay 情境驗證確定性。
