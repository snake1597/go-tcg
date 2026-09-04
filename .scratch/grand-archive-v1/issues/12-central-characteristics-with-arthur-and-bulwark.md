# 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器

**What to build:** 讓 Arthur 與 Bulwark Sword 的 static、temporary 與 triggered 行為只透過中央 derived-characteristic evaluator 影響 power、life、ability、cost 及 permission。

**Blocked by:** 10: 以 Impact Hammer 完成觸發收集與排序; 11: 完成攻擊、Retaliation、傷害與死亡的戰鬥縱切.

**Status:** ready-for-agent

- [ ] Evaluator 實作 Support Set 需要的 Layer A 至 E、power/life sub-layer、dependency、dependency loop 與 timestamp 規則。
- [ ] 合法性、目標、費用、戰鬥及 Player View 顯示使用同一份 derived result。
- [ ] Arthur 的 optional rest、immortality 與 rested allies power，以及 Bulwark 的 class bonus 與 wield payment 正確運作。
- [ ] 來源失效或 duration 到期後重新求值，不把衍生數值永久寫入 Object。
