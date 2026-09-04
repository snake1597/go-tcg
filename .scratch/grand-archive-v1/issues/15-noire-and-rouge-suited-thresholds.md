# 15: 以 Noire 與 Rouge 支援 Suited threshold 行為

**What to build:** 讓共用 Suited reserve-total 查詢驅動 Noire 的 stealth 與 On Enter counters，以及 Rouge 的 threshold choice 與 conditional damage。

**Blocked by:** 10: 以 Impact Hammer 完成觸發收集與排序; 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器; 13: 建立 Cardistry 核心並支援基本 Cardistry 卡牌.

**Status:** ready-for-agent

- [ ] Noire 只在存在另一個合格 Suited Ally 時取得 stealth，來源變化後即時重新求值。
- [ ] On Enter reserve-total 門檻使用共用查詢並正確放置 buff counters。
- [ ] Rouge 的 choose-without-target 決策在入 Stack 或結算的正確時點建立並固定。
- [ ] 門檻邊界、唯一來源、來源離場及傷害結果皆有可追溯規則測試。
