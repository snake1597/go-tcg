# 08 — Replacement ability 定義

**What to build:** Infernal Vessel 的替代行為由 Ability Definition 進入既有 Replacement Pipeline，玩家可選順序並得到正確結果。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證。

**Status:** ready-for-agent

- [ ] 定義明列 event filter 與 transformation；每次替代後重算候選。
- [ ] 正式介面驗證多候選順序、prevention、Pending Choice 後續行與 cause／event history。
- [ ] 非法選擇不改變權威狀態；Infernal Vessel 的 Card ID replacement 分支已刪除。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
