# 05: 讓 Standard 回合與 Opportunity 可持續推進

**What to build:** 建立 Standard 回合與 Opportunity 的基礎生命週期：讓已完成開局的單局透過目前已支援的玩家行動、Opportunity 讓過及 deterministic scheduler 推進 Standard phases 與回合，直到下一個穩定停點或單局結束。

**Blocked by:** 04: 以 Spirit of Fire 完成 Standard 開局.

**Status:** complete

- [x] Wake Up、Materialize、Recollection、Draw、Main 與 End 依鎖定規則推進並套用第一回合修正。
- [x] 目前已支援的合法行動（pass、Materialize、skip materialize）由時序、phase、Effects Stack 與既有 Materialize 合法性決定。
- [x] 行動者完成需要 Opportunity 的 Materialize 後保有 Opportunity，直到主動讓過。
- [x] Pending Choice 期間拒絕不相關行動，但任何尚未落敗玩家仍可投降。

**Scope boundary:** 通用 Fast／Slow speed、特殊 permission、任意 card／ability action 的合法性模型由 07–08 處理；通用 Effects Stack lifecycle 與 trigger checkpoint／ordering 由 07–10 處理；多種 Field Object 的 Wake Up 與戰鬥相關 phase 行為由後續戰鬥票處理。此票只交付目前 Support Set 已支援行動所需的回合基礎。
