# 07 — 數值表達式與傷害

**What to build:** 現有門檻與 Suited 費用計數傷害由一般 damage 操作及可序列化數值表達式結算。

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] 現有固定值、Suited Reserve Cost 總和門檻及 distinct printed cost 計數能以已驗證的數值表達式表示。
- [ ] 每個動態數值於規則指定時點求值；Rouge、Noire 與 Straight Flare 的可見結果符合既有規則情境。
- [ ] Damage 仍驗證合法目標並通過 Replacement Pipeline；Pending Choice 後續行與來源 Last-Known Information 正確。
- [ ] 無效表達式與參照在啟動前被拒絕，不在結算中 panic。
- [ ] 卡名專用傷害 operation 及單用途傷害旗標已移除；replay 與 state hash 確定。
