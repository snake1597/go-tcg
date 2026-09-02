# 規則歧義與裁定登錄

本文件只追蹤鎖定規則快照中會影響實作結果的矛盾或疑點。首版規則基準為 `602c917f2f8fd4df7198429a72eb596bf7f647c6`；未解決項目不得由程式碼、測試或卡牌 handler 暗自選邊。

狀態定義：

- `待官方裁定`：公開官方來源尚不足以唯一決定行為；相關 slice 或 Support Set 內容維持未支援。
- `已由官方來源解決`：已找到與鎖定快照相容的官方變更說明或 ruling，並指定應採用的條文。
- `非語意錯字`：可由同一條文中的明確列舉判定，不會改變任何合法遊戲結果。
- `專案自訂裁定`：只有專案負責人明確批准並另建 ADR 後才能使用；目前沒有此類項目。

## RUL-001：玩家行動後的 Opportunity 擁有者

狀態：`待官方裁定`

- [Timing and Permissions](../rules/game-mechanics/game-mechanics-timing-and-permissions.md) 與 card activation／Materialization 規則表示行動執行者繼續取得 Opportunity。
- [Activated Abilities](../rules/game-mechanics/game-mechanics-abilities/abilities-activated-abilities.md) 表示 ability 進入 Stack 後先交給回合玩家。
- [Game Terms](../rules/glossary/game-terms.md) 的 Player Action order 也先指向回合玩家。

影響：Opportunity scheduler、回合外 activated ability 與 bot 合法行動。此項會阻擋里程碑 2 的完整 Opportunity 行為；取得裁定前只能完成不宣稱正式時序正確的 test fixture 骨架。

## RUL-002：必要目標失效時的 fizzle 條件

狀態：`待官方裁定`

- [State-based Checks](../rules/game-mechanics/game-mechanics-miscellaneous-topics/state-based-checks-and-effects.md) 表示所有必要目標都非法或不存在時才 fizzle。
- [Resolution](../rules/game-mechanics/game-mechanics-playing-cards/playing-cards-resolution.md) 表示任一必要目標非法時便不結算。

影響：具有多個必要目標的 activation、Materialization 或 ability。固定 Support Set 若沒有這類效果，可以標記為不適用首版；一旦可達便阻擋該卡牌 slice。

## RUL-003：起始 Level 0 Champion 的放置時間

狀態：`已由官方來源解決`

採用 [Starting the Game](../rules/general-rules/general-rules-starting-the-game.md) 的 Standard step 6：所有玩家在第一回合前揭示並將起始 Level 0 Champion 放入場上，再依 turn order 結算取得起始手牌的 On Enter abilities，全部完成前玩家不能取得 Opportunity。

依據：官方規則庫的 [2025-12-04 changelog](https://github.com/weebsoftheshore/gitbook-rules/blob/602c917f2f8fd4df7198429a72eb596bf7f647c6/README.md) 明確將此列為規則變更；[Champion](../rules/general-rules/general-rules-card-types/card-types-champion.md) 中「各玩家第一回合才放置」視為尚未同步的舊文字，而非專案 house rule。

影響：正式 Standard setup 必須由固定牌組的 Level 0 Champion 行為取得起始手牌，不能以通用 draw 或 fixture 直接發牌取代。

## RUL-004：Token 離場後消失的精確時點

狀態：`待官方裁定`

- [Tokens](../rules/game-mechanics/game-mechanics-tokens.md) 表示 token 移往非 Field zone 時 cease to exist，之後不能再進行區域移動。
- [Token glossary](../rules/glossary/game-terms.md) 表示移動仍完成 state-based checks、triggers 與必要 actions，到 Opportunity 傳遞時才 cease to exist。

影響：On Leave、On Death、zone-change trigger 與後續效果能否引用 token。固定 Support Set 若可建立會離場的 token，此項會阻擋相關 slice。

## RUL-005：Game Zone 數量

狀態：`非語意錯字`

[Game Zones](../rules/game-mechanics/game-mechanics-game-zones/README.md) 宣稱 9 個 zone，但同一句明確列出 Main Deck、Material Deck、Hand、Memory、Field、Graveyard、Banishment、Intent、Pantheon 與 Effects Stack 共 10 個。引擎以明確列出的 10 個名稱作為 zone vocabulary；Standard MVP 不建立 Pantheon gameplay，遇到 Pantheon 內容由 Support Set 驗證拒絕。這不改變任何 Standard 合法行動。

## 處理流程

1. 每個 vertical slice 在進入 `ready` 前，從規則依賴反查本表。
2. 若命中 `待官方裁定`，先檢查官方規則庫、正式 changelog、官方 judge ruling 或勘誤；只接受可穩定引用的來源。
3. 找到裁定後，記錄來源 URL、裁定日期、適用規則 commit 與精確行為，並先建立失敗的規則案例測試。
4. 找不到裁定時，從 Support Set 排除相關內容；若運行時才發現，引擎以 `NeedsRuling` 結束並輸出 issue ID、replay 與 state hash。
5. 若專案負責人選擇 house rule，新增 ADR，標示與官方規則的差異，再將本表狀態改為 `專案自訂裁定`。不得只修改本表便視為已批准。

公開官方 GitHub 規則庫已於 2026-08-31 檢查；除 RUL-003 的 changelog 與 RUL-005 的非語意計數錯字外，未找到其餘三項可穩定引用的正式裁定。
