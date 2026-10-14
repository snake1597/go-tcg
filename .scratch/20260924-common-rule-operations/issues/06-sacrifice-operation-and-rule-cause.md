# 06 — Sacrifice operation 與規則原因

**What to build:** 建立可供能力效果與額外費用共用的 sacrifice operation，統一驗證受控 Object、原子離場與 Card Instance 移動，同時保留 sacrifice、destroy 等不同規則原因。Peppered Chef 與 Weapon 額外費用只作為代表性驗證案例。

**Blocked by:** 04 — 場上 Object 放逐。

**Related foundation:** 05 — 具名選擇與卡牌篩選。已存在的窄 Object choice 能力足以開始 sacrifice vertical slice，不等待所有選牌入口完成共通化。

**Related migration:** DSL-ready Card Definitions 13 — Peppered Chef 與其他 Object／Weapon Slots 的 Card Definition 遷移；本票只負責 sacrifice operation 與交易語意。

**Status:** completed

**Delivered slice:** 04 完成後，Peppered Chef 與 Weapon 額外費用皆改用 `commitObjectSacrifice`。operation 在提交前驗證控制權、Object 類型、排除來源及 Card Instance 所有權，成功時原子移除 Object、移入 owner Graveyard，並以 `ability:sacrifice` 或 `cost:sacrifice` 記錄公開事件。Peppered Chef 的 +2 Power 維持為後續 generic modifier；Weapon 付款在最終提交前重驗證，失效時 state hash 不變。

**Next implementation step:** 04 完成後，以 Peppered Chef optional sacrifice 作第一個 operation slice，只抽出合法 Object 驗證、sacrifice 原因事件及原子離場；加成效果仍由 Definition 的後續 operation 表示。接著讓 Weapon additional cost 重用同一 sacrifice operation，驗證宣告失敗完全 rollback。

- [x] 犧牲前驗證合法控制者、Object 類型與排除來源條件；代表性 optional choice 只包含依法可犧牲的其他受控 Object。
- [x] 成功時被犧牲的 Object 離場、Card Instance 進入 owner 的 Graveyard，Game Event 與觸發原因標示為犧牲。
- [x] 後續效果只在合法犧牲成功後套用，且不由 sacrifice operation 直接寫入卡牌專屬效果。
- [x] Weapon 作為費用時，宣告或後續提交失敗不留下已犧牲武器的半完成狀態。
- [x] 犧牲與戰鬥破壞在可見事件及相關觸發中仍可區分；Player View 與 replay 正確。
- [x] 共通 runtime 不再包含 Peppered Chef 專屬犧牲 operation 或 Weapon 直接移區路徑；個別 Card Definition 遷移由 DSL-ready Card Definitions 13 驗收。
