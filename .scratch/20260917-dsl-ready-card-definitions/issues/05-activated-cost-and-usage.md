# 05 — Activated ability 的費用與使用限制

**What to build:** 代表性 Cardistry 能力以同一定義宣告 timing、target、費用、使用限制及效果，成功提交後建立獨立 Ability Instance。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證；共通規則操作 09 — Reserve Cost 與 Memory Cost 原子提交；共通規則操作 10 — 替代費用與 Wield 付款。

**Status:** completed

- [x] Declaration Transaction 在建 instance 和入 Stack 前完成合法性及付款；成功付款只發生一次。
- [x] 費用不足、取消、非法最終目標及 stale revision 不改變 zone、counter、PRNG、event、trigger buffer、Knowledge State 或 hash。
- [x] Source Ref／LKI 與 Object 離場再進場後的 usage 生命週期正確；對應 Card ID 宣告分支已刪除。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。

2026-09-27：驗證已遷移 Cardistry 由編譯定義提供 timing、Memory cost、once-per-object usage 與 effect sequence；付款成功後才建立帶有 Source LKI 的獨立 Ability Instance 並推入 Effects Stack。付款不足、無效付款與 stale revision 都在提交前失敗，沒有狀態副作用。使用紀錄以 Object ID 分隔，物件離場再進場不會繼承前一個 Object 的使用狀態；Wonderland's Reign 與 Three of Hearts 的舊 Card ID 宣告分支均已移除。品質門檻：`go test ./...`。
