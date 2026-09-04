# 18: 以 Heated Vengeance 與 Red Hare 擴充攻擊狀態

**What to build:** 讓本回合 Champion 受傷歷史、On Attack 自傷、Pride、Human 條件及 granted attack trigger 共同影響合法攻擊與戰鬥結果。

**Blocked by:** 10: 以 Impact Hammer 完成觸發收集與排序; 11: 完成攻擊、Retaliation、傷害與死亡的戰鬥縱切; 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器; 14: 支援 Cardistry 的選擇、棄牌與部署互動.

**Status:** ready-for-agent

- [ ] Heated Vengeance 只在本回合 Champion 已受傷時取得正確 power，On Attack optional self-damage 使用觸發流程。
- [ ] Red Hare 的 Pride restriction 及 Unique Human 條件由中央 permission/characteristic query 決定。
- [ ] Granted ability 建立獨立 Ability Instance，optional discard-then-draw 不由卡牌直接修改 zones。
- [ ] 回合切換、來源離場、重新進場及沒有可棄牌選項時的 lifetime 與合法性均正確。
