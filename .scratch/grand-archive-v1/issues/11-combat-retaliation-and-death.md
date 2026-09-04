# 11: 完成攻擊、Retaliation、傷害與死亡的戰鬥縱切

**What to build:** 讓玩家從合法行動宣告攻擊，經 Opportunity 回應、Retaliation 與同時傷害抵達戰鬥結束，並由 state-based fixed point 處理死亡及 Champion 落敗。

**Blocked by:** 06: 以 Tonoris 完成 Champion Lineage 與 Materialization; 07: 以 Blazing Throw 建立第一條完整卡牌宣告與 Stack 路徑; 10: 以 Impact Hammer 完成觸發收集與排序.

**Status:** ready-for-agent

- [ ] 攻擊者、目標、wield 與時序皆由 Player View 的合法選項建立並在提交時重驗。
- [ ] 戰鬥與 Retaliation 傷害保留同時性、On Hit／On Kill 因果及確定順序。
- [ ] State-based checks 每輪使用一致 derived view，變更後重跑直到收斂，再 flush triggers。
- [ ] Champion 死亡會結束單局；循環或超過保守上限則停止並輸出 replay、state hash 與診斷。
