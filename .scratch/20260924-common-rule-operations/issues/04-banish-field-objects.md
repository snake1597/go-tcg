# 04 — 場上 Object 放逐

**What to build:** 玩家啟動需要放逐場上來源的能力時，Object 移除、卡牌進入 Banishment 與效果建立使用同一套規則流程。

**Blocked by:** None — can start from the existing narrow card-zone transition and Field Object lifecycle.

**Related foundation:** 01 — 玩家卡牌區域移動、棄牌與 Memory recollection。玩家區域入口的完整共通化不阻擋本票建立 Object 離場、LKI 與 Banishment operation slice。

**Status:** ready-for-agent

**Next implementation step:** 選一個現有「放逐自身作為費用」的 Object ability，先抽出不含 Card ID 的 typed banish operation，經 Declaration Transaction 同時提交 Object 移除與 owner Banishment 移動，再以原 scenario 驗證 LKI、事件、失敗無副作用與 replay。若由 DSL 13 觸發此工作，完成本 operation slice 後立即回到該 Ability Slot。

- [ ] Regalia 與其他現有放逐來源的能力，均驗證控制者、來源狀態及付款條件後才原子提交。
- [ ] 成功時移除正確的 Object，將所屬 Card Instance 放入 owner 的 Banishment，並保留 Source Ref、Last-Known Information 與 Game Event。
- [ ] 失敗或過期輸入不改變區域、Object、事件或 PRNG 狀態。
- [ ] 後續能力仍沿 Effects Stack 與 Replacement Pipeline 的既有生命週期執行；Player View 與 replay 正確。
- [ ] 完成路徑不再直接重複刪除 Object 與追加 Banishment 的邏輯。
