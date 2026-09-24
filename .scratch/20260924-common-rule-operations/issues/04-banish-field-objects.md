# 04 — 場上 Object 放逐

**What to build:** 玩家啟動需要放逐場上來源的能力時，Object 移除、卡牌進入 Banishment 與效果建立使用同一套規則流程。

**Blocked by:** 01 — 棄牌與一般區域移動。

**Status:** ready-for-agent

- [ ] Regalia 與其他現有放逐來源的能力，均驗證控制者、來源狀態及付款條件後才原子提交。
- [ ] 成功時移除正確的 Object，將所屬 Card Instance 放入 owner 的 Banishment，並保留 Source Ref、Last-Known Information 與 Game Event。
- [ ] 失敗或過期輸入不改變區域、Object、事件或 PRNG 狀態。
- [ ] 後續能力仍沿 Effects Stack 與 Replacement Pipeline 的既有生命週期執行；Player View 與 replay 正確。
- [ ] 完成路徑不再直接重複刪除 Object 與追加 Banishment 的邏輯。
