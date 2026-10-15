# 11 — 物件能力共通宣告整合

**What to build:** 讓 Object activated abilities 可共用同一套條件驗證、Cost Payment 與 operation 提交流程。剩餘 Object、Weapon、Regalia 的 Card Definition 遷移與 production support metadata 由 DSL-ready Card Definitions 13、14 驗收。

**Blocked by:** 02 — 抽牌與抽空規則；03 — Source Card 移動及 Ally 部署；04 — 場上 Object 放逐；06 — Sacrifice operation 與規則原因；07 — 數值表達式與傷害；08 — Action 複製 operation 與 runtime copy 生命週期；10 — 替代費用與 Wield 付款。

**Related migration:** DSL-ready Card Definitions 13 — 剩餘 Object 與 Regalia Slots；14 — definition-derived production support gate。

**Status:** completed

**Completion evidence:** 02、03、04、06、07、08、10 均已完成。Object 的可用行動與付款共用來源擁有權及費用前綴檢查；編譯時驗證 runtime operation 種類、具名參照、必要參數與多餘欄位。Arthur 的專用休息加 Immortality operation 已拆為共通操作，Red Hare 的單用途衍生旗標與戰鬥手寫效果已改為條件式 Attack trigger 的具名選牌、棄牌與抽牌操作。`go test ./...` 與 `go test -race ./...` 通過，包含 Standard Game、Support Set、replay 與十組 mirror games。

**Delivered slice:** DSL-ready Card Definitions 13 已交付固定 Support Set 的窄整合：Object activated abilities 共用條件查詢、`rest_source`／`banish_source` cost prefix、Ability Instance、Effects Stack 與 resolution operations；Duchess's Thornes 的 `cardistry_discount` 亦在結算時才生效。此 slice 不解除其餘前置 Common issues，剩餘非固定 Support Set 入口仍待本票最終清理。

**Cleanup scope:** Card ID 分派仍僅用於載入各張 Card Definition；Five of Spades 的個別 Card ID 行為由對應 DSL migration 清理。Combat、Level Up 與 Materialization 的獨立生命週期維持專用入口。

- [x] 現有可組合的 Object ability 使用共通啟動條件、費用與 operation 提交語意；玩家可用行動與原有規則一致。
- [x] Combat、Level Up 與 Materialization 等獨立生命週期保留專用入口，並透過共通規則操作提交可共用的狀態變化。
- [x] Operation validation 在開局前拒絕未知 operation、錯誤參照與無效參數；production support metadata 的衍生與完整性由 DSL-ready Card Definitions 14 負責。
- [x] 共通 runtime 中的卡名專用 operation、單用途旗標與重複直接狀態修改均被移除，不保留相容層；個別 Card ID 行為路徑由對應 DSL migration 清理。
- [x] 受影響卡牌的 Player View、Game Event、Replacement Pipeline、replay、state hash 與完整 Standard Game 驗證通過。
