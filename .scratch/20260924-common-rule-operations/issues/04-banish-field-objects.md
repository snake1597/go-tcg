# 04 — 場上 Object 放逐

**What to build:** 玩家啟動需要放逐場上來源的能力時，Object 移除、卡牌進入 Banishment 與效果建立使用同一套規則流程。

**Blocked by:** None — can start from the existing narrow card-zone transition and Field Object lifecycle.

**Related foundation:** 01 — 玩家卡牌區域移動、棄牌與 Memory recollection。玩家區域入口的完整共通化不阻擋本票建立 Object 離場、LKI 與 Banishment operation slice。

**Status:** completed

**Delivered slice:** DSL-ready Card Definitions 13 已以 Grand Crusader's Ring、Water／Wind Resonance Bauble、Smoke Bombs、Safeguard Amulet 與 The Duchess's Thornes 證明共用 typed `banish_source` cost：先驗證控制者、來源狀態與必要選擇，再一次移除 Field Object、移入 owner Banishment、保留 Ability Source LKI 與公開事件，後續效果統一進 Effects Stack；stale source 拒絕且 state hash 不變。剩餘工作是盤點非本 Support Set 的其他放逐入口並移除重複 zone mutation。

**Completed implementation:** `commitObjectAbilityBanishment` 以不含 Card ID 的 typed operation，在完整驗證控制者、來源與卡牌擁有者後，一次提交 Object 移除、owner Banishment 移動與公開事件。所有 Object 的 `banish_source` 費用均經此 operation；Grand Crusader's Ring scenario 驗證 Source LKI、公開事件、Effects Stack 與 replay，失敗與過期輸入以 StateHash 驗證無副作用。

- [x] Regalia 與其他現有放逐來源的能力，均驗證控制者、來源狀態及付款條件後才原子提交。
- [x] 成功時移除正確的 Object，將所屬 Card Instance 放入 owner 的 Banishment，並保留 Source Ref、Last-Known Information 與 Game Event。
- [x] 失敗或過期輸入不改變區域、Object、事件或 PRNG 狀態。
- [x] 後續能力仍沿 Effects Stack 與 Replacement Pipeline 的既有生命週期執行；Player View 與 replay 正確。
- [x] 完成路徑不再直接重複刪除 Object 與追加 Banishment 的邏輯。
