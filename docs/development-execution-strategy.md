# Grand Archive 開發執行策略

本文件補充 [`development-plan.md`](./development-plan.md) 與 [`implementation-features.md`](./implementation-features.md)，說明功能的執行順序、並行邊界、規格與 ticket 的建立時機，以及正式開發前可先完成的工作。

## 核心執行策略

先完成一條能真正執行的垂直切片，驗證 Game Module 的核心 seam，再從已驗證的邊界展開並行工作。工作不應單純按檔案或技術層分割，避免各自完成的 state、event、stack 與 trigger 無法整合。

```text
規格入口 ──────────────┐
固定牌組、卡面版本      │
Support Set、規則裁定   │
                       ▼
測試骨架 → Game Module tracer bullet
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
 PlayerView        Replay/RNG      CLI/Bot Adapter
        └──────────────┼──────────────┘
                       ▼
       Standard 開局、Scheduler、回合
                       ▼
           第一張正式可操作卡牌
                       ▼
                    戰鬥
                       ▼
       依 dependency graph 擴充固定牌組
                       ▼
          整合、壓力測試、發布
```

## 執行階段

### 階段 0：兩條起始工作軌

以下兩條工作軌可立即並行。

#### 規格與內容工作軌

1. 完成唯一固定 Standard Main Deck 與 Material Deck（見[固定 Standard 牌組](#固定-standard-牌組)）。
2. 取得明確的 Outside Game Pool。
3. 固定卡面資料來源、版本及所有穩定 Content ID。
4. 從 [`card-support-matrix-template.md`](./card-support-matrix-template.md) 建立正式 `card-support-matrix.md`。
5. 遞迴展開 Support Set 並建立 dependency graph。
6. 列出需要的規則、Ability Slot、typed operation、continuous／replacement 類型及 ruling gate。
7. 對被固定牌組觸及的規則歧義取得裁定或保持 blocked。

## 固定 Standard 牌組

此 manifest 是目前唯一可進入 production registry 的 Standard 牌組。卡牌識別一律使用 `entity.Card.UUID`（各 `card/*.json` 頂層的 `uuid`）；不使用 `editions` 或 `result_editions` 內的 UUID，因為後兩者是印刷版本而非卡牌定義。

`entity.Card.Types` 決定牌組區域：`CHAMPION` 與 `REGALIA` 放入 Material Deck，其餘列出的卡牌放入 Main Deck。未列入 manifest 的 `card` 檔案不是這副牌的一部分。

### Main Deck（60 張）

前十一張與 `Nether Dodobird`、`Searing Rebuke`、`Vengeful Paramour` 各四張，`Incinerator Felindroid` 一張，`Volda, Smolder's Spite` 三張；總數為 `14 × 4 + 1 + 3 = 60`，每個名稱均不超過 Standard 的四張上限。

| Card UUID | 名稱 | 張數 |
| --- | --- | ---: |
| `hbpu4fo8oo` | Blighted Jewel | 4 |
| `v0gu8efq08` | Lingering Banshee | 4 |
| `s9ICPMYPNx` | Bill, Chimney Sweep | 4 |
| `stiyh3pmk3` | Cinder Geyser | 4 |
| `BqDw4Mei4C` | Creative Shock | 4 |
| `qzv380ujf5` | Duchess, Six of Hearts | 4 |
| `PptfA8gG6h` | Emberwrath Witch | 4 |
| `1gxrpx8jyp` | Fanatical Devotee | 4 |
| `5pw07bh5wf` | Fractal of Sparks | 4 |
| `cbNF64gCsS` | Furnace Drone | 4 |
| `26ya6zaae8` | Incinerated Templar | 4 |
| `Vl03t5rMSA` | Incinerator Felindroid | 1 |
| `wtHBZAdTSv` | Nether Dodobird | 4 |
| `hdvpug4d5m` | Searing Rebuke | 4 |
| `4vjkezn49t` | Vengeful Paramour | 4 |
| `ecZsQQAYJJ` | Volda, Smolder's Spite | 3 |

### Material Deck（12 張）

所有卡牌各一張；`Spirit of Fire` 的 `Level` 為 `0` 且 `Types` 含 `CHAMPION`，符合 Standard 起始 Champion 要求。

| Card UUID | 名稱 | `Types` | `Level` | 張數 |
| --- | --- | --- | ---: | ---: |
| `LMyKyVC2O9` | Spirit of Fire | `CHAMPION` | 0 | 1 |
| `GiQxfpKTUC` | Alice, Distorted Queen | `CHAMPION` | 1 | 1 |
| `9gv4vm4kj3` | Backup Charger | `REGALIA`, `ITEM` | — | 1 |
| `2gv7DC0KID` | Grand Crusader's Ring | `REGALIA`, `ITEM` | — | 1 |
| `pol1nz0j1n` | Nullifying Mirror | `REGALIA`, `ITEM` | — | 1 |
| `yj2rJBREH8` | Safeguard Amulet | `REGALIA`, `ITEM` | — | 1 |
| `ScGcOmkoQt` | Smoke Bombs | `REGALIA`, `ITEM` | — | 1 |
| `xnrw8qq1uw` | Tariff Ring | `REGALIA`, `ITEM` | — | 1 |
| `s3572j3oda` | Viridian Protective Trinket | `REGALIA`, `ITEM` | — | 1 |
| `chsbalegbs` | Impact Hammer | `REGALIA`, `WEAPON` | — | 1 |
| `vgWgu1DUYv` | Infernal Vessel | `REGALIA`, `ITEM` | — | 1 |
| `1ubrwubSQN` | Mantle of the Abyss | `REGALIA`, `ITEM` | — | 1 |

Material Deck 已達 12 張上限，且卡名均唯一。`Everflame Staff` 與 `Fanned Synchron` 雖然也是 `REGALIA`，但不屬於這份固定 manifest，因此不會放入 Main Deck 或 Material Deck。

#### Test-only walking skeleton 工作軌

在正式 registry 仍因內容 metadata 不完整而 blocked 時，以不可能進入 production registry 的 fixture 走通：

1. 啟動 CLI。
2. 建立測試單局。
3. 取得指定玩家的 PlayerView。
4. 以 revision 與 ViewHandle 提交一個選擇。
5. 允許真人或 bot 投降。
6. 結束單局並寫出 canonical replay。
7. 重新播放輸入並驗證逐步 state hash。
8. 驗證過期、跨玩家或非法輸入不產生任何狀態副作用。

walking skeleton 只驗證端到端 seam，不代表任何正式卡牌已被支援。

### 階段 1：固定最小公開 seam

第一個 tracer bullet 完成後，再固定少量 Game Module use cases：

- 建立單局。
- 取得指定玩家 PlayerView。
- 以當前 revision 提交 PlayerAction 或 PendingChoice 回答。

此階段同時確認：

- GameState 的權威寫入邊界。
- revision 的遞增與拒絕規則。
- ViewHandle 的玩家範圍與生命週期。
- canonical replay 的輸入模型。
- canonical state hash 的序列化邊界。
- CLI 與 bot 只依賴 PlayerView 的限制。

公開 seam 尚未經 tracer bullet 驗證前，不大量建立 Adapter、介面或卡牌 handler。

### 階段 2：可安全並行的基礎能力

最小 seam 穩定後，可以按可觀察行為並行開發以下工作：

| 工作軌 | 內容 | 整合界面 |
| --- | --- | --- |
| 玩家資訊 | KnowledgeState、PlayerView、ViewHandle、隱藏資訊測試 | Game Module view use case |
| 確定性 | 版本化 PRNG、canonical serialization、state hash、replay | 已提交輸入與 canonical state |
| 內容驗證 | Deck manifest、資料版本、Support Set closure validator | 建立單局 gate |
| CLI | PlayerView 顯示、編號選單、錯誤與 replay 提示 | Game Module 公開 use cases |
| Bot | 最小合法決策、revision 提交、決策上限 | 與 CLI 相同的 PlayerView |
| 品質 | 情境測試格式、規則來源標註、race／fuzz／CI 入口 | 公開測試 seam |
| 診斷 | NeedsRuling、scheduler loop、replay divergence 格式 | 統一診斷結果 |

這些工作必須各自具備明確輸入、輸出及驗收測試，不能直接修改其他工作軌的內部狀態。

### 階段 3：Standard 生命週期

依下列順序建立可推進的正式單局：

1. Standard 單局建立與格式 gate。
2. Level 0 Champion 與 Lineage 初始狀態。
3. 由 Champion On Enter abilities 取得起始手牌。
4. deterministic scheduler 與 Stable Yield Point。
5. RuleCheckpoint 與 state-based fixed point。
6. Wake Up、Materialize、Recollection、Draw、Main、End。
7. 第一回合修正。
8. Opportunity、讓過與階段推進。
9. 牌庫耗盡與投降結束。

正式 Opportunity 時序依 [ADR 0015](./adr/0015-retain-opportunity-until-the-holder-passes.md) 實作：玩家完成需要 Opportunity 的行動後保有 Opportunity，直到讓過才依 turn order 移交。

### 階段 4：第一張正式可操作卡牌

從固定牌組選擇一張依賴較少，但足以驗證完整規則路徑的卡牌。該 slice 應涵蓋：

1. DeclarationTransaction。
2. mode、target、cost 與 payment。
3. 取消、非法、過期及 PRNG rollback。
4. Source Card 與 StackItem 關聯。
5. Opportunity 與讓過。
6. FILO 結算。
7. EventBatch 與 cause chain。
8. trigger buffer 與 checkpoint flush。
9. state-based fixed point。
10. 目標失效、fizzle 或 negate 的適用行為。

第一張卡不應簡單到完全避開核心行動模型，也不應同時引入 continuous、replacement、token 與複雜戰鬥等所有高風險機制。

### 階段 5：戰鬥垂直切片

第一張卡的堆疊流程穩定後，再依序加入：

1. 合法攻擊列舉。
2. 攻擊宣告交易。
3. Opportunity 回應。
4. Retaliation。
5. EventBatch 同時傷害。
6. damage prevention／replacement。
7. On Hit／On Kill。
8. 單位死亡、離場與 LKI。
9. Champion 敗北。

### 階段 6：固定牌組擴充

依 Support Set dependency graph 一次完成一張卡牌或一個最小跨卡互動：

1. 先完成被多張卡共用的最低層 typed operation。
2. 再完成依賴它的卡牌 handler。
3. 每列都加入正常、非法、目標失效與必要跨卡測試。
4. 每完成一列，重新執行 Support Set closure validation。
5. 所有依賴、Ability Slot 與測試完成後，才能把該內容標記為 supported。

不同卡牌 handler 可並行，但必須符合以下條件：

- 共用事件與 typed operation 語意已經固定。
- 卡牌之間不存在尚未完成的 dependency edge。
- 不會各自在 handler 中繞過 evaluator、replacement pipeline 或 scheduler。
- 修改範圍與驗收情境彼此獨立。

### 階段 7：首版收尾

- 完成固定牌組的所有 Support Set 內容。
- 移除誤入 production 的 fixture 路徑。
- 調整 bot 啟發式與無進展保護。
- 改善 CLI 可讀性。
- 完成 replay regression corpus。
- 執行 race、fuzz／property tests。
- 至少執行 100 場不同 seed 的 bot 鏡像對戰。
- 驗證沒有 panic、deadlock、scheduler 不收斂、非法提交或 NeedsRuling。

## 不適合並行的工作

以下工作具有共同語意核心，應由一條垂直切片先固定行為，再讓其他工作依賴它：

- 多人同時修改 GameState、scheduler、RuleCheckpoint 與 state-based checks。
- EventBatch 語意尚未固定時，大量撰寫 triggered abilities。
- Source Card、StackItem、CardInstance 與 GameObject 身分尚未驗證時，同時展開戰鬥與複雜卡牌。
- Derived evaluator 的 layer、dependency 與 timestamp 尚未固定時，各自在 card handler 寫 modifier 特例。
- Replacement pipeline 尚未確立時，在個別 handler 中直接改寫事件結果。
- 第一張正式卡牌尚未走通時，就把固定牌組全部切成實作 tickets。

工作應按「可驗收行為」分割，而不是把 `state.go`、`events.go`、`stack.go` 分給不同開發者各自完成。

## Spec 建立時機

### 現在可以建立的 spec

下列規格不依賴尚未補齊的 production registry 內容，可以立即完成：

1. Game Module 最小 use-case contract。
2. Game／PlayerView／PlayerAction／PendingChoice 的狀態協定。
3. test-only walking skeleton 的完整 Given／When／Then。
4. canonical replay 格式與版本欄位。
5. canonical state hash 的包含與排除規則。
6. 固定牌組 manifest schema。
7. Support Set closure validator 規格。
8. NeedsRuling、scheduler 不收斂及 replay divergence 的診斷格式。
9. 規則與卡面來源的測試追溯格式。

### 需要輸入後才能建立的 spec

- 正式 Standard setup 的完整卡牌行為。
- 每張正式卡牌的能力與效果。
- Support Set 實際需要的 continuous／replacement operations。
- 跨卡互動與特殊關鍵字。
- 固定牌組的完整 bot 評分策略。

這些規格必須等待 Outside Game Pool、卡面版本、CardFace／Ability Slot ID 與相關規則輸入到位。

## Ticket 建立時機

### 現在可以建立的 tickets

只建立能獨立驗收、且不依賴正式牌組的 walking-skeleton tickets：

- 建立 test-only fixture。
- 建立單局並取得初始 PlayerView。
- 驗證 revision 與非法輸入無副作用。
- 使用 ViewHandle 提交選擇。
- 在 PendingChoice 期間投降。
- 寫出最小 canonical replay。
- 重播輸入並驗證 state hash。
- 建立 CLI 啟動及編號選單最小路徑。
- 建立 production registry 無法載入 fixture 的 gate。

### 目前不應建立的 tickets

- 實作所有卡牌效果。
- 完成整個 Effects Stack。
- 完成完整 continuous evaluator。
- 完成整套戰鬥系統。
- 尚未完成 Support Set 分析的單張正式卡牌。

上述項目過大或仍缺少規則輸入。應先轉成小型 slice spec，再建立具有明確驗收條件的 ticket。

### Spec 到 ticket 的 gate

```text
功能目錄
  → 最小可觀察行為
  → Slice Spec
  → 規則與卡面來源確認
  → Given／When／Then 驗收案例
  → Dependency／Ruling Gate
  → Ready Ticket
  → Red → Green
  → Replay／State Hash Regression
```

一張 ticket 只有符合 [`development-plan.md`](./development-plan.md#垂直切片就緒門檻) 的條件後才能進入 ready。

## 建議的其他前置作業

### PREP-01 固定版本 metadata

把引擎版本、規則 commit、卡面資料版本、牌組版本及 PRNG 版本變成機器可檢查的資料，而不只存在於文字文件。

完成條件：建立單局與讀取 replay 時都能驗證全部版本；不相符時明確拒絕。

### PREP-02 第一個 executable scenario

將 walking skeleton 寫成完整範例，包含：

- 初始狀態。
- 指定玩家可見的 PlayerView。
- 可提交的 action／choice。
- 預期 EventBatch。
- 預期 revision。
- 預期結果與 state hash。

### PREP-03 Canonical serialization 規格

先決定：

- map 與集合的穩定排序。
- runtime ID 的表示方式。
- 哪些權威狀態必須進入 hash。
- 哪些 cache、顯示文字及 derived view 應排除。
- schema 或 hash algorithm 的版本方式。

### PREP-04 卡面資料品質檢查

- Card ID 與 CardFace ID 唯一性。
- reference／referenced-by 完整性。
- 必要欄位與已知 card type 驗證。
- Ability Slot 的穩定性。
- 相同資料版本內不得發生內容漂移。
- 匯入錯誤需指出資料來源及 Content ID。

### PREP-05 規則追溯模板

每個情境測試固定記錄：

- 規則 commit。
- 規則檔案與條目。
- 卡面資料版本。
- Card ID／CardFace ID。
- Ability Slot。
- 相關 RUL issue。
- 預期事件與最終狀態。

### PREP-06 診斷資料模型

定義以下失敗的共同診斷欄位：

- 非法或過期輸入。
- Support Set 缺漏。
- NeedsRuling。
- scheduler／fixed-point 不收斂。
- replay divergence。
- state hash mismatch。

診斷至少包含引擎／規則／資料版本、Game ID、revision、cause chain、相關 Content ID／Ability Slot、replay 位置與 state hash。

### PREP-07 Fixture 隔離

- Fixture 只存在於測試 package 或測試建置路徑。
- Production registry 不接受 fixture Content ID。
- 正式 CLI 使用 fixture 時必須測試失敗。
- release gate 掃描並拒絕 fixture registry entry。

### PREP-08 Definition of Done

單張卡或機制完成至少代表：

- 正常案例通過。
- 至少一個非法或邊界案例通過。
- 目標失效或來源離場案例已處理。
- 必要跨卡互動通過。
- replay 與 state hash 可重現。
- PlayerView 不洩漏資訊。
- 規則與卡面來源可追溯。
- Support Set 與 registry 都標記 supported。

### PREP-09 自動化品質入口

- 建立 `go test ./...` 的一致執行方式。
- 建立 `go test -race ./...` 的 CI gate。
- 為 scheduler、declaration rollback、ViewHandle 及 replay 建立有時限的 fuzz／property tests。
- 為批次 bot 對戰建立 seed、行動上限、失敗 replay 保存與摘要輸出。

## 暫不需要的前置作業

首版不應預先建立：

- 資料庫與 repository abstraction。
- 卡牌 DSL 或嵌入式腳本。
- 完整卡池支援。
- 沒有 Support Set 使用者的完整 evaluator operation 集合。
- 微服務或分散式單局協調。
- 只為未來 Adapter 準備的 interface。
- 大量空 package 或只有型別沒有端到端行為的 scaffolding。

## 待確認的 Grill 決策

以下問題尚未得到專案負責人的答案。答案會直接決定規格與 tickets 的下一層分解。

### DEC-01 固定牌組責任與期限

已完成：本文件的[固定 Standard 牌組](#固定-standard-牌組)是唯一的 Main Deck 與 Material Deck manifest；卡牌定義 ID 由 `entity.Card.UUID` 提供。

仍待完成：指定 Outside Game Pool、卡面版本與其餘 Content ID；這些資料到位前，Support Set closure 與正式卡牌 tickets 仍維持 blocked。

### DEC-02 Walking skeleton 啟動決策

是否在 production registry 尚未 ready 時，先使用 test-only fixture 完成建立單局、PlayerView、投降、replay 與 hash 驗證？

建議：立即開始，優先驗證整體 seam 與確定性模型。

### DEC-03 第一張正式卡牌選擇原則

從已固定的牌組選擇第一張正式卡牌時，應選效果最簡單的卡，還是最能揭露引擎風險的卡？

建議：選擇依賴少，但能完整經過 DeclarationTransaction、StackItem、Opportunity、結算與 checkpoint 的卡。

### DEC-04 未裁定規則政策（已完成）

RUL-001 未取得可唯一決定行為的官方 ruling，專案負責人已選擇以 [ADR 0015](./adr/0015-retain-opportunity-until-the-holder-passes.md) 記錄專案自訂裁定。RUL-002 與 RUL-004 已由官方來源解決。

決策：RUL-001 不再阻擋 Opportunity slice；實作與測試必須明確遵循 ADR 0015。未來若出現新的未裁定 issue，仍依規則治理流程保持 blocked，直到取得官方裁定或專案負責人批准新的 ADR。

## 下一個決策點

完成 DEC-02～03，並補齊 DEC-01 尚待項目後，才進一步決定；DEC-04 已完成：

- manifest 的正式 schema 與保存位置。
- walking skeleton 的第一個 executable scenario。
- 第一張正式卡牌與 dependency path。
- spec 的文件範本。
- ticket 的大小、依賴與驗收格式。
- 哪些工作可以分派給不同開發者而不共享未穩定的核心語意。
