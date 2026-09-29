# 16 — Bot 使用通用合法行動 metadata

**What to build:** bot 依自身 Player View 的通用合法行動資訊選擇行動；新增定義卡牌不需在 Game Module 加精確卡名 heuristic。

**Blocked by:** 15 — 版本化 DSL 編譯成相同中介表示。

**Related migration:** 12 — 遷移其餘 Action Ability Slots；13 — 遷移其餘 Object 與 Regalia Ability Slots。通用 legal-action metadata 可先以已具備的 Red Hare 與 Duchess executable definitions 驗證，不等待所有 Object／Regalia Slots 遷移。

**Status:** blocked

**Blocked reason:** 15 — 先完成 Ability Slot 遷移、definition-derived Support Gate 與第一張版本化 DSL 卡牌，再固定 bot 消費的通用 legal-action metadata。

**Execution order:** 排在 15 之後作 DSL 主線的最後收尾。只從 Player View 增加通用且玩家可見的決策 metadata，再移除 Game Module 內的 Red Hare、Duchess 精確卡名 heuristic。

- [ ] Player View 只提供玩家可見的通用 legal-action metadata；Bot Controller 不讀 Game State 或隱藏資訊。
- [ ] Red Hare、Duchess 等決策不依精確卡名；相同 view 與 seed 得到相同決定。
- [ ] CLI 及有時限鏡像對戰可完成，無非法提交、stall、desync 或 replay hash 分歧。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
