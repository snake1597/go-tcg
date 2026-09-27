# 03 — 目標與數值表達式能力

**What to build:** Straight Flare 由通用 selector、reference 與 value expression 決定合法目標和傷害數值。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證；共通規則操作 07 — 數值表達式與傷害。

**Status:** ready-for-agent

- [ ] Selector 只產生合法且 visibility-safe 的選項，順序確定；compiler 驗證 reference 型別及此能力所需的數值運算。
- [ ] 正常傷害、非法目標、結算時目標失效、Replacement Pipeline 與已付費用保留均由正式介面 scenario 驗證。
- [ ] Straight Flare 的卡名專用效果欄位與分支已刪除。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
