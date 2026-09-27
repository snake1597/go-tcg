# 07 — Static 與 permission 定義

**What to build:** Red Hare 的靜態修正及權限由 Ability Definition 提供中央 Derived Characteristics evaluator 使用。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證。

**Status:** completed

- [x] 定義明列 predicate、layer、sublayer、modifier、duration 與 source-presence；permission／prohibition 不藏在 Card ID 檢查中。
- [x] 正式介面驗證 layer、timestamp、dependency、dependency loop、來源離場與合法行動；printed characteristics 不被永久改寫。
- [x] Red Hare 的專用 evaluator 與合法性分支已刪除。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。

2026-09-27：`ability:5du8f077ua:front:pride` 與 `ability:5du8f077ua:front:qualified-human` 以 static Ability Definition 宣告 Pride、條件式移除 Pride 與 On Attack；compiler 驗證完整 static payload，中央 evaluator 依已編譯定義生成效果。`TestRedHarePermissionAndGrantedAttackAbilityUseDerivedCharacteristics` 經 Player View 與 Submit 驗證條件式合法攻擊、來源離場與重新進場；既有 characteristic tests 覆蓋 layer、timestamp、dependency loop 與不改寫 printed characteristics。品質門檻：`go test ./...`、`go test -race ./...` 與兩個 10 秒 fuzz 測試。
