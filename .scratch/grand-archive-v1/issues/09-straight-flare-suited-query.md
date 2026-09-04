# 09: 以 Straight Flare 建立 Suited 查詢與動態傷害

**What to build:** 讓 Straight Flare 使用引擎共用的 Suited Card Face 與 printed reserve cost 查詢計算動態傷害，並在結算時依正式裁定處理必要目標失效。

**Blocked by:** 07: 以 Blazing Throw 建立第一條完整卡牌宣告與 Stack 路徑.

**Status:** ready-for-agent

- [ ] 查詢以 Card Face 的 printed 資料計算不同 reserve cost，不由單卡 handler 自行掃描或推測。
- [ ] Player View 只列出合法目標，宣告與結算分別重驗目標合法性。
- [ ] 所有必要目標失效時完整 fizzle；仍有合法必要目標時只對合法目標結算。
- [ ] 正常、零種 cost、重複 cost 與目標失效案例均有規則來源及 replay/hash 驗證。
