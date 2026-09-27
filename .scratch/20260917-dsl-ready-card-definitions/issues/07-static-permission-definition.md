# 07 — Static 與 permission 定義

**What to build:** Red Hare 的靜態修正及權限由 Ability Definition 提供中央 Derived Characteristics evaluator 使用。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證。

**Status:** ready-for-agent

- [ ] 定義明列 predicate、layer、sublayer、modifier、duration 與 source-presence；permission／prohibition 不藏在 Card ID 檢查中。
- [ ] 正式介面驗證 layer、timestamp、dependency、dependency loop、來源離場與合法行動；printed characteristics 不被永久改寫。
- [ ] Red Hare 的專用 evaluator 與合法性分支已刪除。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
