# 01: 建立版本化且可診斷的確定性 Game Module 基線

**What to build:** 讓現有 walking skeleton 成為可由正式版本、seed 與輸入序列重現的最小 Game Module 路徑，並在版本或狀態分歧時提供一致診斷；移除只服務舊 fixture 的過時路徑。

**Blocked by:** None (can start immediately).

**Status:** completed

- [x] Replay 記錄並驗證引擎、規則、卡面資料、牌組及 PRNG 版本與初始 seed。
- [x] 相同版本、seed 與輸入序列逐步產生相同 canonical state hash。
- [x] 不相容版本或 hash 分歧會指出第一個失敗輸入與可判讀原因，不會靜默接受。
- [x] 非法或過期輸入不改變 state、revision、replay 或 PRNG cursor。
