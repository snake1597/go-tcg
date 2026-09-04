# 13: 建立 Cardistry 核心並支援基本 Cardistry 卡牌

**What to build:** 讓玩家透過一致的 Cardistry 行動啟動 Wonderland's Reign、Five/Four of Spades 與 Two of Hearts/Spades，涵蓋折扣、一次性追蹤、抽牌、draw-to-memory、buff counter、Floating Memory 及暫時 power。

**Blocked by:** 07: 以 Blazing Throw 建立第一條完整卡牌宣告與 Stack 路徑; 09: 以 Straight Flare 建立 Suited 查詢與動態傷害; 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器.

**Status:** ready-for-agent

- [ ] Cardistry cost 使用共用 Suited distinct-cost query 與中央 cost evaluator。
- [ ] Once-per-instance 依 Ability Instance 的正確 lifetime 記錄，離場再進場的新 Object 不沿用舊狀態。
- [ ] 抽牌、draw-to-memory、buff counter、Floating Memory 與 temporary power 都透過 typed operations 產生事件。
- [ ] 重複啟動、費用不足、空牌庫及來源失效均有無副作用或依法失敗的情境測試。
