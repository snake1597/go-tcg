# 02 — 抽牌與抽空規則

**What to build:** 回合抽牌、能力抽牌與抽至 Memory 依正確的 Main Deck 順序和抽空規則執行，並產生一致的可見結果。

**Blocked by:** 01 — 玩家卡牌區域移動、棄牌與 Memory recollection。

**Status:** blocked

**Blocked reason:** 01 — 玩家卡牌區域移動的共通提交語意尚未完成。

- [ ] 先依鎖定規則確認 Main Deck 的牌組頂、抽牌次序、牌組不足與抽空勝負語意；有歧義時記錄 Needs Ruling，不以現有實作差異代替規則。
- [ ] 回合抽牌、能力抽牌與抽至 Memory 各自移動正確張數至正確區域，並依規則處理空牌組。
- [ ] Event Batch、Game Event、Knowledge State、Player View 及追蹤權與抽牌結果一致。
- [ ] 相同 seed 與玩家輸入可重播且 state hash 確定；完成後移除被取代的重複抽牌路徑。
