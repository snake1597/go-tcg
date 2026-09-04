# 07: 以 Blazing Throw 建立第一條完整卡牌宣告與 Stack 路徑

**What to build:** 讓玩家從 Player View 宣告 Blazing Throw、逐步選擇目標及 Weapon 犧牲費用，原子建立 Source Card 與 Stack Item，並在所有玩家讓過後結算傷害或依法 fizzle。

**Blocked by:** 06: 以 Tonoris 完成 Champion Lineage 與 Materialization.

**Status:** ready-for-agent

- [ ] Declaration Transaction 隔離保存選擇、候選 zone move、cost snapshot、events、triggers 與 PRNG cursor。
- [ ] 取消、過期、非法目標、沒有可犧牲 Weapon 或最終非法皆完整 rollback。
- [ ] 成功宣告才提交犧牲費用並建立分離的 Source Card、Stack Item、Source Ref 與必要 LKI。
- [ ] 結算前重驗目標；fizzle 或 negate 不退回已提交費用，且來源卡生命週期正確完成。
