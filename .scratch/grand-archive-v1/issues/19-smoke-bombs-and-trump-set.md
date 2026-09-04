# 19: 以 Smoke Bombs 與 Trump Set 完成戰鬥重定向

**What to build:** 讓玩家在戰鬥相關 Opportunity 中使用 Smoke Bombs 或 Trump Set，授予暫時 stealth、重定向 active attack，並套用對應的 power、life、true sight 與 cost modifier。

**Blocked by:** 11: 完成攻擊、Retaliation、傷害與死亡的戰鬥縱切; 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器.

**Status:** ready-for-agent

- [ ] Smoke Bombs 只能選擇合法 Ally，banish 自身後授予本回合 stealth 並抽牌。
- [ ] 攻擊目標在 stealth、taunt 與 true sight 改變後會透過共用規則重新驗證。
- [ ] Trump Set 只能把 active attack 改指向不同且合格的 Suited Ally，並套用正確 duration 的 power/life。
- [ ] 失去原目標、沒有替代目標、來源失效及 phase 結束案例均有確定性測試。
