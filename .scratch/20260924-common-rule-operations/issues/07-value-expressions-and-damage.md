# 07 — 數值表達式與傷害

**What to build:** 現有門檻與 Suited 費用計數傷害由一般 damage 操作及可序列化數值表達式結算。

**Blocked by:** None — can start immediately.

**Related migration:** DSL-ready Card Definitions 03 — Straight Flare 負責其 Card Definition 遷移；本票負責所有既有傷害入口可重用的 expression 與 damage operation。

**Status:** in-progress

**Delivered slice:** DSL-ready Card Definitions 03 已以 Straight Flare 證明 declared-target reference、distinct printed cost expression、結算時求值、Replacement Pipeline 及 replay。剩餘工作是完成 Rouge、Noire 與其他現有 damage 入口的共通化及清理。

**Next implementation step:** 先遷移 Rouge 的 Suited Reserve Cost 總和門檻，僅加入該規則需要的 count、sum 與 threshold expression；完成正常值、門檻邊界、目標失效與 replacement scenario 後，再以相同 expression primitives 遷移 Noire，最後刪除兩者專用傷害旗標。

- [ ] 現有固定值、Suited Reserve Cost 總和門檻及 distinct printed cost 計數能以已驗證的數值表達式表示。
- [ ] 每個動態數值於規則指定時點求值；Rouge、Noire 與 Straight Flare 的可見結果符合既有規則情境。
- [ ] Damage 仍驗證合法目標並通過 Replacement Pipeline；Pending Choice 後續行與來源 Last-Known Information 正確。
- [ ] 無效表達式與參照在啟動前被拒絕，不在結算中 panic。
- [ ] 共通 damage runtime 不再需要卡名專用 operation 或單用途傷害旗標；各卡牌 Definition 路徑的最終清理由對應 DSL migration 驗收。
- [ ] replay 與 state hash 確定。
