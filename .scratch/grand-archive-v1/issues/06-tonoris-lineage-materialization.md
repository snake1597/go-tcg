# 06: 以 Tonoris 完成 Champion Lineage 與 Materialization

**What to build:** 讓玩家在合法 Materialize Phase 透過引擎選項支付費用並將 Tonoris 放到 Champion Lineage 頂端，同時保留 Champion Object 身分並執行其 On Enter taunt 行為。

**Blocked by:** 05: 讓 Standard 回合與 Opportunity 可持續推進.

**Status:** ready-for-agent

- [ ] 合法 Materialization 經 Effects Stack 與 Opportunity 流程完成，非法 lineage、時序或付款不改變狀態。
- [ ] Level Up 保持 Champion Object ID，原頂端卡成為 Inner Lineage 且不再是 Field Object。
- [ ] orientation、counter、戰鬥角色與依法保留的 runtime 狀態不因 Level Up 被重置。
- [ ] Tonoris 的 On Enter 正確授予有期限的 taunt，結果可由 Player View、events、hash 與 replay 驗證。
