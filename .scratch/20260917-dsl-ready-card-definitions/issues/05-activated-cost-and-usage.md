# 05 — Activated ability 的費用與使用限制

**What to build:** 代表性 Cardistry 能力以同一定義宣告 timing、target、費用、使用限制及效果，成功提交後建立獨立 Ability Instance。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證；共通規則操作 09 — Reserve Cost 與 Memory Cost 原子提交；共通規則操作 10 — 替代費用與 Wield 付款。

**Status:** ready-for-agent

- [ ] Declaration Transaction 在建 instance 和入 Stack 前完成合法性及付款；成功付款只發生一次。
- [ ] 費用不足、取消、非法最終目標及 stale revision 不改變 zone、counter、PRNG、event、trigger buffer、Knowledge State 或 hash。
- [ ] Source Ref／LKI 與 Object 離場再進場後的 usage 生命週期正確；對應 Card ID 宣告分支已刪除。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
