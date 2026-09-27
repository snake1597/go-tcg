# 12 — 遷移其餘 Action Ability Slots

**What to build:** Blazing Throw、Fiery Interference、Trump Set 等剩餘 Action 由定義完成宣告、目標、效果及來源移動。

**Blocked by:** 03 — 目標與數值表達式能力；04 — 具名選擇與可序列化續行；05 — Activated ability 的費用與使用限制。

**Status:** ready-for-agent

- [ ] 每個 Slot 保留規則來源，並驗證正常結算、非法宣告、付款、fizzle 與來源移動。
- [ ] 傷害、暫時修正及目標重選重用 typed 原語；不可分 mechanism 才使用狹窄 native operation，且不直接修改 Game State。
- [ ] 每個已遷移 Action 的 Card ID 分支立即刪除；Player View、事件及 replay 正確。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
