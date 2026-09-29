# 11 — 遷移其餘 Cardistry Ability Slots

**What to build:** 固定 Support Set 中剩餘 Cardistry 能力由各自 Card Definition 描述，玩家仍能抽牌、抽至 Memory、加 counter、取得暫時修正與作選擇。

**Blocked by:** 03 — 目標與數值表達式能力；04 — 具名選擇與可序列化續行；05 — Activated ability 的費用與使用限制；10 — 複製能力與 runtime 身分。

**Status:** completed

- [x] 每個 Slot 有穩定 ID、規則或 ruling 來源；scenario 涵蓋正常結算、非法宣告、來源失效及跨機制互動。
- [x] 重用通用 selector、value、duration 與 typed effect；完成一個 Slot 就刪其舊 Card ID 或卡名 operation 路徑。
- [x] Player View、Knowledge State、Game Event、卡牌守恆與 replay 維持正確。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
