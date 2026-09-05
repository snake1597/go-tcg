# Grand Archive v1 工作追蹤

本目錄是 Grand Archive 首版的唯一工作追蹤來源，負責記錄 issue 的執行順序、相依關係、狀態與驗收結果。[`docs/implementation-features.md`](../../../docs/implementation-features.md) 只保存功能目錄與完整驗收範圍，不重複維護工作順序或進度。

## 執行規則

- `Blocked by` 是能否開始 issue 的權威依據；所有列出的前置 issue 都是 `completed` 後才能開始。
- 編號表示預期交付順序；同時有多張 issue 可開始時，優先選擇編號較小者，但編號不能取代 `Blocked by`。
- 狀態只使用 `ready-for-agent`、`in-progress`、`blocked`、`completed`。
- 開始實作時將狀態改為 `in-progress`；只有全部驗收項目均勾選且相關測試通過後，才能改為 `completed`。
- 因外部裁定或未完成前置工作而無法繼續時使用 `blocked`，並在 issue 中記錄具體原因。

## 里程碑對應

| 里程碑 | Issue | 完成結果 |
| --- | --- | --- |
| 0. 規格入口 | [02](./02-fixed-deck-and-support-set-gate.md) | 固定牌組、production registry 與 Support Set 開局 gate 可用 |
| 1. 最小端到端骨架 | [01](./01-versioned-deterministic-game-baseline.md)、[03](./03-knowledge-state-player-view.md) | 版本化確定性基線與玩家資訊邊界完成 |
| 2. Standard 生命週期 | [04](./04-spirit-of-fire-standard-setup.md)～[06](./06-tonoris-lineage-materialization.md) | Standard 開局、回合、Opportunity、Lineage 與 Materialization 可運作 |
| 3. 第一張可操作卡 | [07](./07-blazing-throw-declaration-stack.md)～[10](./10-impact-hammer-trigger-ordering.md) | declaration、Stack、fast action、動態查詢與 trigger ordering 形成完整路徑 |
| 4. 戰鬥縱切 | [11](./11-combat-retaliation-and-death.md) | 攻擊、Retaliation、傷害、死亡與敗北完整串接 |
| 5. 固定牌組擴充 | [12](./12-central-characteristics-with-arthur-and-bulwark.md)～[23](./23-enable-complete-fixed-deck.md) | 固定牌組所有內容通過 Support Set gate |
| 6. 首版收尾 | [24](./24-player-view-only-heuristic-bot.md)～[27](./27-v1-release-quality-gate.md) | bot、production CLI、完整對戰與發布品質 gate 完成 |

里程碑只用來描述產品完成度，不改變 issue 的實際依賴。工作選擇與解鎖仍以各 issue 的 `Blocked by` 和 `Status` 為準。

## 目前可執行

依目前狀態，issue 01 與 issue 02 已完成；下一張可執行工作是 [03：以 Knowledge State 保護 Player View 與 View Handle](./03-knowledge-state-player-view.md)。Issue 04 及後續工作仍受各自的 `Blocked by` 限制。
