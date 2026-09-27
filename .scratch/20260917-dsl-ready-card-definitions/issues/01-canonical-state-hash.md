# 01 — 補齊 canonical state hash

**What to build:** 相同版本、seed 與輸入的逐步重播能以 hash 辨識所有影響目前或未來規則結果的權威狀態。

**Blocked by:** None — can start immediately.

**Status:** completed

- [x] Cardistry 使用狀態、折扣、next effect／ability／object identity，以及其餘行為相關 canonical 欄位都參與 hash；逐一改變時 hash 必須不同。
- [x] 相同輸入序列的逐步 replay hash 相同；被拒絕的輸入不改變 hash、PRNG cursor、Knowledge State、event 或費用。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
