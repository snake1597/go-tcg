# 06 — Triggered ability 定義

**What to build:** Impact Hammer 由 Ability Definition 觀察 Game Event 並建立 triggered Ability Instance，按既有 checkpoint 與 Effects Stack 流程結算。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證。

**Status:** completed

- [x] Event Batch、同時觸發、玩家排序選擇、Rule Checkpoint 與來源離場 LKI 皆正確。
- [x] 所有 trigger 完成入 Stack 後才交出 Opportunity；正式介面 scenario 檢查事件和 replay。
- [x] Impact Hammer 的 Card ID trigger discovery 分支已刪除。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。

2026-09-27：`ability:chsbalegbs:front:on-wield-self-damage` 以 `triggered` Ability Definition 宣告 `wield` event 與 `event-target` reference。Committed Event Batch 會保留 weapon 與 target，觸發探索依已編譯定義建立帶 Source LKI 的 Ability Instance，並重用既有 trigger ordering、Effects Stack 與 Opportunity 流程。`TestImpactHammerTriggeredDefinitionUsesPlayerViewAndReplay` 經 Player View 提交 Wield 與 target choice，驗證 event、stable Slot、3 點傷害與 replay；既有 trigger 測試覆蓋同時排序與來源離場。品質門檻：`go test ./...`、`go test -race ./...` 與兩個 10 秒 fuzz 測試。
