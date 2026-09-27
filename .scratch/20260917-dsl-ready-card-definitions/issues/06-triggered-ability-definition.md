# 06 — Triggered ability 定義

**What to build:** Impact Hammer 由 Ability Definition 觀察 Game Event 並建立 triggered Ability Instance，按既有 checkpoint 與 Effects Stack 流程結算。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證。

**Status:** ready-for-agent

- [ ] Event Batch、同時觸發、玩家排序選擇、Rule Checkpoint 與來源離場 LKI 皆正確。
- [ ] 所有 trigger 完成入 Stack 後才交出 Opportunity；正式介面 scenario 檢查事件和 replay。
- [ ] Impact Hammer 的 Card ID trigger discovery 分支已刪除。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
