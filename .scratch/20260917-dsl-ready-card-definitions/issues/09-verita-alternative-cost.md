# 09 — 替代費用的可執行定義

**What to build:** Verita 由同一能力定義提供一般與替代付款，之後的持續效果依相同規則結算。

**Blocked by:** 05 — Activated ability 的費用與使用限制。

**Related foundation:** 共通規則操作 10 — 替代費用與 Wield 付款；該票負責跨入口交易共通化，本票只驗收 Verita Definition 遷移。

**Status:** completed

- [x] Alternative／additional cost 共用 Declaration Transaction；失敗或取消不消耗資源，成功不重複付款。
- [x] Ability Instance、Source Ref／LKI、duration 與 replay identity 正確。
- [x] Verita 的 Card ID 費用與效果分支已刪除，正常及失敗 scenario 通過。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。

2026-09-28：Verita 的一般 Ally 啟動與墓地替代付款都透過同一個提交交易建立帶有穩定 Slot、Source 與 Source LKI 的 Ability Instance。替代付款在逐張選擇期間可取消且不變更 zone；付款卡與來源卡只在完整合法付款後一次提交。靜態 Immortality、死亡後受控 Suited Ally 的 +1 Power，以及至擁有者下回合結束的 duration 都由已編譯定義驅動，已移除 Card ID 費用與效果分支。品質門檻：`go test ./...`。
