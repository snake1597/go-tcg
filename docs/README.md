# Grand Archive 開發文件索引

本頁只負責文件導航，不保存牌組版本、支援狀態、實作進度或規則裁定。需要資料時，先依下表讀取唯一負責該資訊的文件。

## 優先查找表

| 想知道什麼 | 優先讀取 | 必要時接著讀 |
| --- | --- | --- |
| 固定牌組、卡面版本、CardFace、Ability Slot 或 Support Set | [`card.md`](./card.md) | 對應 `card/*.json`、[`../card-data-manifest.json`](../card-data-manifest.json) |
| 某張卡依賴哪些機制或其他卡牌 | [`card.md`](./card.md#support-set-dependency-graph) | [`rules-issues.md`](./rules-issues.md)、對應 `rules/` 文件 |
| 現在應該實作什麼 | [Grand Archive v1 工作追蹤](../.scratch/grand-archive-v1/issues/README.md#目前可執行) | [`card.md`](./card.md#support-set-dependency-graph) |
| 功能是否完成或被什麼阻擋 | [Grand Archive v1 工作追蹤](../.scratch/grand-archive-v1/issues/README.md) | [`implementation-features.md`](./implementation-features.md)、[`rules-issues.md`](./rules-issues.md) |
| 首版範圍、核心模型或完成門檻 | [`development-plan.md`](./development-plan.md) | 對應 ADR |
| 規則歧義與採用的裁定 | [`rules-issues.md`](./rules-issues.md) | 官方 `rules/`、對應 ADR |
| 測試策略與禁止模式 | [`testing.md`](./testing.md) | [`development-plan.md`](./development-plan.md#每個垂直切片的強制流程) |
| 為什麼採用某項架構決策 | [`adr/`](./adr/) | [`development-plan.md`](./development-plan.md) |
| 研究背景與外部模式 | [`research/`](./research/) | 研究文件引用的一手來源 |

## 文件責任與關聯

| 文件 | 唯一負責內容 | 可以引用 | 不應重複 |
| --- | --- | --- | --- |
| [`card.md`](./card.md) | 卡面資料流程、固定牌組、Content ID、Ability Slot、Support Set 與 dependency graph | `card/*.json`、data manifest、rules、ruling ID | 開發里程碑、架構原則、完整 ruling 內容 |
| [`development-plan.md`](./development-plan.md) | 產品範圍、核心模型、架構原則、slice 與 release gate | `card.md`、ADR、rules issues | 具體牌表、目前資料版本、工作狀態 |
| [`implementation-features.md`](./implementation-features.md) | 功能目錄與完整驗收範圍 | development plan、card dependency、ruling ID | 執行順序、工作狀態、牌表、卡面能力全文、架構決策全文 |
| [Grand Archive v1 工作追蹤](../.scratch/grand-archive-v1/issues/README.md) | 里程碑、issue 執行順序、相依、狀態與逐項驗收結果 | implementation features、card dependency、rules issues | 完整功能規格、牌表、規則裁定內容 |
| [`rules-issues.md`](./rules-issues.md) | 規則歧義、裁定狀態、證據與處理流程 | 官方 rules、ADR | 牌組清單、實作進度 |
| [`testing.md`](./testing.md) | 測試層級、情境格式、品質 gate 與禁止模式 | development plan、rules | 功能 backlog、卡牌 registry |
| [`adr/`](./adr/) | 已批准且難以逆轉的單一決策及理由 | plan、rules、research | 當前版本、工作狀態 |
| [`research/`](./research/) | 研究過程、風險與來源證據 | primary sources、ADR | 當前執行狀態 |

關聯方向保持單向：工作追蹤讀取 `implementation-features.md`、`card.md` 與 `rules-issues.md`；`implementation-features.md` 讀取 `development-plan.md` 與 `card.md`；`card.md` 讀取卡面 manifest、rules 與 `rules-issues.md`；README 只導向這些文件。下游文件不得反向複製上游事實。

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
| [0016](./adr/0016-use-hierarchical-content-ids.md) | CardFace 與 Ability Slot 的顯式階層式 ID |
