# 06 — 犧牲 Ally 與 Weapon

**What to build:** Peppered Chef 的 Ally 犧牲與 Action 的 Weapon 額外費用使用共通犧牲流程，同時保留各自後續效果。

**Blocked by:** 04 — 場上 Object 放逐；05 — 具名選擇與卡牌篩選。

**Status:** ready-for-agent

- [ ] 犧牲前驗證合法控制者、Object 類型與排除來源條件；Chef 的選擇可略過且只包含其他受控 Ally。
- [ ] 成功時被犧牲的 Object 離場、Card Instance 進入 owner 的 Graveyard，Game Event 與觸發原因標示為犧牲。
- [ ] Chef 的加成只在合法犧牲成功後套用，且持續時間與原有規則一致。
- [ ] Weapon 作為費用時，宣告或後續提交失敗不留下已犧牲武器的半完成狀態。
- [ ] 犧牲與戰鬥破壞在可見事件及相關觸發中仍可區分；Player View 與 replay 正確。
- [ ] 被取代的 Chef 專屬犧牲操作及 Weapon 直接移區路徑已移除。
