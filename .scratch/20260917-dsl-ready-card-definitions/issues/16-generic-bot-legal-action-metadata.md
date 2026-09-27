# 16 — Bot 使用通用合法行動 metadata

**What to build:** bot 依自身 Player View 的通用合法行動資訊選擇行動；新增定義卡牌不需在 Game Module 加精確卡名 heuristic。

**Blocked by:** 12 — 遷移其餘 Action Ability Slots；13 — 遷移其餘 Object 與 Regalia Ability Slots。

**Status:** ready-for-agent

- [ ] Player View 只提供玩家可見的通用 legal-action metadata；Bot Controller 不讀 Game State 或隱藏資訊。
- [ ] Red Hare、Duchess 等決策不依精確卡名；相同 view 與 seed 得到相同決定。
- [ ] CLI 及有時限鏡像對戰可完成，無非法提交、stall、desync 或 replay hash 分歧。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
