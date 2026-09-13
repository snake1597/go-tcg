# 15: 以 Noire 與 Rouge 支援 Suited threshold 行為

**What to build:** 讓共用 Suited reserve-total 查詢驅動 Noire 的 stealth 與 On Enter counters，以及 Rouge 的 threshold choice 與 conditional damage。

**Blocked by:** 10: 以 Impact Hammer 完成觸發收集與排序; 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器; 12.5: 建立統一 Ability 與 Effect Runtime; 13: 建立 Cardistry 核心並支援基本 Cardistry 卡牌.

**Status:** in-progress

- [x] Noire 只在存在另一個合格 Suited Ally 時取得 stealth，來源變化後即時重新求值。
- [x] On Enter reserve-total 門檻使用共用查詢並正確放置 buff counters。
- [x] Rouge 的「depending on」門檻在結算時計算；`Choose a unit` 的必要 target 則在 trigger 入 Stack 時以合法 View Handle 固定。
- [ ] 門檻邊界、唯一來源、來源離場及傷害結果皆有可追溯規則測試。
