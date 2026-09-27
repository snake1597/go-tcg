# 04 — 具名選擇與可序列化續行

**What to build:** Three of Hearts 的選牌與棄牌透過具名 binding 串接；玩家回答後從可序列化 Resolution Frame 確定地續行。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證；共通規則操作 05 — 具名選擇與卡牌篩選。

**Status:** ready-for-agent

- [ ] 只有 actor 看得見 Pending Choice；選取後的 reference 指向同一合法 Card Instance，只有允許時才能 pass。
- [ ] 過期、跨玩家或偽造 handle 遭拒且不改變 state hash、PRNG、Knowledge State、event 或費用。
- [ ] Continuation 不保存 closure 或可變 definition pointer；replay 結果一致，舊選牌分支已刪除。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
