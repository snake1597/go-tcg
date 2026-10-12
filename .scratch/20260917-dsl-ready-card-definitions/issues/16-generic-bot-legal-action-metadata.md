# 16 — Bot 使用通用合法行動 metadata

**What to build:** bot 依自身 Player View 的通用合法行動資訊選擇行動；新增定義卡牌不需在 Game Module 加精確卡名 heuristic。

**Blocked by:** 15 — 版本化 DSL 編譯成相同中介表示。

**Related migration:** 12 — 遷移其餘 Action Ability Slots；13 — 遷移其餘 Object 與 Regalia Ability Slots。通用 legal-action metadata 可先以已具備的 Red Hare 與 Duchess executable definitions 驗證，不等待所有 Object／Regalia Slots 遷移。

**Status:** completed

**Completed:** 15 已完成；Player View 現在為所有有來源的 Legal Action 提供同視圖可見 source，並提供 `can_win_game` 與 definition-derived triggered ability slots；Bot Controller 自行解讀策略並移除精確卡名 heuristic。

**Execution order:** 排在 15 之後作 DSL 主線的最後收尾。只從 Player View 增加通用且玩家可見的決策 metadata，再移除 Game Module 內的 Red Hare、Duchess 精確卡名 heuristic。

- [x] Player View 只提供玩家可見的通用 legal-action metadata；Bot Controller 不讀 Game State 或隱藏資訊。
- [x] Red Hare、Duchess 等決策不依精確卡名；相同 view 與 seed 得到相同決定。
- [x] CLI 及有時限鏡像對戰可完成，無非法提交、stall、desync 或 replay hash 分歧。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
