# Grand Archive 開發文件索引

目前狀態：核心架構與開發流程已定案，唯一固定 Standard Main Deck 與 Material Deck 已建立，卡面資料已鎖定為 `card-data-v1`，初版 Support Set 也已記錄於 [`card-support-matrix.md`](./card-support-matrix.md)。進入正式卡牌實作前，仍需補齊 Outside Game Pool 與 CardFace／Ability Slot ID。已登錄的規則 issue 目前皆已有正式處理結果。

## 建議閱讀順序

1. [`development-plan.md`](./development-plan.md)：首版範圍、核心模型、vertical slices 與 release gate。
2. [`implementation-features.md`](./implementation-features.md)：依牌組、單局、卡牌能力、觸發、效果、戰鬥與介面拆解的功能實作清單。
3. [`development-execution-strategy.md`](./development-execution-strategy.md)：實作順序、並行工作軌、spec／ticket gate、前置作業與待確認決策。
4. [`../CONTEXT.md`](../CONTEXT.md)：全專案共用的領域詞彙。
5. [`testing.md`](./testing.md)：rule-example-first TDD、測試 seam 與禁止模式。
6. [`card-support-matrix.md`](./card-support-matrix.md)：目前固定 Standard 牌組的正式 Support Set 與阻擋狀態；[`card-support-matrix-template.md`](./card-support-matrix-template.md) 保留為日後建立新矩陣的格式。
7. [`card-data-versioning.md`](./card-data-versioning.md)：`./card` 權威快照、manifest 格式與更新流程。
8. [`rules-issues.md`](./rules-issues.md)：會阻擋相關 slice 的官方規則歧義。
9. [`adr/`](./adr/)：已確認且難以逆轉的架構決策。

研究背景：

- [`research/full-rules-review.md`](./research/full-rules-review.md)：完整閱讀 108 份鎖定規則文件後的風險清單與落地狀態。
- [`research/card-game-engine-patterns.md`](./research/card-game-engine-patterns.md)：成熟遊戲引擎的一手來源模式研究。

## ADR 索引

| ADR | 決策 |
| --- | --- |
| [0001](./adr/0001-authoritative-deterministic-rule-engine.md) | 權威且確定性的 Game Module |
| [0002](./adr/0002-reject-unsupported-card-mechanics-before-play.md) | 開局前驗證封閉 Support Set |
| [0003](./adr/0003-pin-rules-snapshot-per-engine-version.md) | 每個引擎版本鎖定規則快照 |
| [0004](./adr/0004-use-guided-choices-and-atomic-actions.md) | Guided choices 與 DeclarationTransaction |
| [0005](./adr/0005-implement-card-behavior-in-go.md) | 首版卡牌行為使用 typed Go |
| [0006](./adr/0006-expose-player-scoped-views-to-controllers.md) | KnowledgeState、PlayerView 與 ViewHandle |
| [0007](./adr/0007-keep-storage-outside-the-rule-engine.md) | 儲存基礎設施位於 Game Module 外 |
| [0008](./adr/0008-record-replays-as-versioned-inputs.md) | 版本化輸入序列 replay |
| [0009](./adr/0009-use-a-modular-monolith-with-a-deep-game-module.md) | 模組化單體與深層 Game Module |
| [0010](./adr/0010-serialize-state-transitions-within-each-game.md) | 序列化輸入、scheduler、fixed point 與 trigger ordering |
| [0011](./adr/0011-stop-on-unsupported-rules-ambiguity.md) | 未裁定歧義停止而不猜測 |
| [0012](./adr/0012-separate-card-object-and-stack-identities.md) | Card、Object、Ability、來源卡與 StackItem 身分 |
| [0013](./adr/0013-preserve-event-batches-and-causality.md) | EventBatch 與因果關係 |
| [0014](./adr/0014-centralize-derived-characteristics-and-replacements.md) | 中央 derived evaluator 與 replacement pipeline |
| [0015](./adr/0015-retain-opportunity-until-the-holder-passes.md) | Opportunity 由行動者保留至讓過 |

## 下一個可執行動作

固定 Main Deck、Material Deck、`card-data-v1` 與初版內容閉包已建立。下一步整合並驗收里程碑 1 的 test-only walking skeleton，同時補齊 Outside Game Pool 與 CardFace／Ability Slot ID；矩陣達到 `ready` 前不開始正式卡牌 handler。
