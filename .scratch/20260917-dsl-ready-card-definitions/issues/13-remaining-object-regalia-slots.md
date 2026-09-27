# 13 — 遷移其餘 Object 與 Regalia Ability Slots

**What to build:** 固定 Support Set 中剩餘 Object、Weapon 與 Regalia 的 activated、triggered、static、permission 及 replacement 能力皆由 Card Definition 提供。

**Blocked by:** 05 — Activated ability 的費用與使用限制；06 — Triggered ability 定義；07 — Static 與 permission 定義；08 — Replacement ability 定義；09 — 替代費用的可執行定義。

**Status:** ready-for-agent

- [ ] 每個 Slot 有穩定 ID、規則來源與 scenario，涵蓋正常結算、非法宣告、來源失效及跨機制互動。
- [ ] Object 能力共用宣告生命週期；static 只輸入中央 evaluator，replacement 只輸入 Replacement Pipeline。
- [ ] 已遷移 Slot 的 Card ID 路徑全部刪除，不留無法驗證的 Supported Slot。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
