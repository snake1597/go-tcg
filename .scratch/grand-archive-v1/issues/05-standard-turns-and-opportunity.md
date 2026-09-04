# 05: 讓 Standard 回合與 Opportunity 可持續推進

**What to build:** 讓已完成開局的單局透過合法玩家行動、Opportunity 讓過及 deterministic scheduler 依序推進 Standard phases 與回合，直到下一個穩定停點或單局結束。

**Blocked by:** 04: 以 Spirit of Fire 完成 Standard 開局.

**Status:** ready-for-agent

- [ ] Wake Up、Materialize、Recollection、Draw、Main 與 End 依鎖定規則推進並套用第一回合修正。
- [ ] Opportunity holder 的合法行動由時序、速度、phase、Stack 與 permission 決定。
- [ ] 行動者完成需要 Opportunity 的行動後保有 Opportunity，直到主動讓過。
- [ ] Pending Choice 期間拒絕不相關行動，但任何尚未落敗玩家仍可投降。
