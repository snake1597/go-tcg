# 16: 以 The Duchess's Thornes 支援 Cardistry observer

**What to build:** 讓 Ally 的 Cardistry activation 觸發 The Duchess's Thornes，並支援 Hindered、rest-and-banish cost、暫時 power／true sight 及 next-use Cardistry discount。

**Blocked by:** 10: 以 Impact Hammer 完成觸發收集與排序; 11: 完成攻擊、Retaliation、傷害與死亡的戰鬥縱切; 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器; 13: 建立 Cardistry 核心並支援基本 Cardistry 卡牌.

**Status:** ready-for-agent

- [ ] Observer 只對 Ally 的 Cardistry activation 觸發，不對 Phantasia activation 觸發。
- [ ] 來源 Ally、activation event 與 trigger 透過 typed identity 關聯，來源離場後仍按 LKI 判定。
- [ ] Power、true sight 與 next-use discount 在正確 duration 或使用點失效。
- [ ] Hindered 及 rest-and-banish 費用由合法行動與原子交易強制執行。
