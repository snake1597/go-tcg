# Grand Archive 核心開發計畫

## 目標

首版交付一個可信、可測試且可重播的 Grand Archive 規則引擎：一名真人透過 CLI，使用唯一固定牌組與啟發式 bot 進行 Standard 鏡像對戰，並能依支援的官方規則完成整場單局。

首版規則基準鎖定 `rules` 儲存庫 commit `602c917f2f8fd4df7198429a72eb596bf7f647c6`（2026-08-24）。固定牌組與卡面資料版本由專案負責人稍後指定。

## 首版範圍

### 必須具備

- 雙人 Standard 單局，一名真人對一個啟發式 bot。
- 真人與 bot 使用同一份固定且合法的牌組清單。
- 由 CLI 顯示玩家視角、合法行動、待決選擇、遊戲事件與結果。
- 支援固定牌組可達內容閉包完整需要的規則、卡牌與互動，不宣稱支援完整卡池。
- 能由 Champion 死亡、牌庫耗盡或投降結束。
- 相同版本、seed 與輸入序列可以逐步重現相同 state hash。
- 未支援卡牌與機制在開始單局前 fail-fast；未裁定歧義不得默默猜測。

### 明確排除

- 多人與 Pantheon。
- 自由組牌、sideboard、錦標賽、帳號、排名及網路配對。
- 完整卡池與未被固定牌組使用的關鍵字。
- minimax、MCTS、機器學習或會偷看隱藏資訊的 bot。
- PostgreSQL、GORM、Redis 與執行中單局持久化。
- 卡面自然語言解析、卡牌 DSL 與嵌入式腳本語言。
- 微服務、預先建立的 repository 與沒有第二個 Adapter 的假想 seam。

## 核心原則

1. Game Module 是唯一能改變單局狀態的權威來源。
2. CLI 與 bot 只能取得指定玩家依法可見的 `PlayerView`。
3. 每個單局一次只接受一個輸入；回合外互動由 Opportunity 形成序列，而非平行寫入狀態。
4. 玩家行動以逐步選擇建立，完整驗證後才原子提交。
5. 規則上的同時事件由單次原子轉移表示；私密同時選擇必須全部收集後才共同公開。
6. 所有隨機性由可指定 seed 且具版本識別的 PRNG 提供；同一引擎版本固定 shuffle 與亂數消耗順序，不可依賴 Go map 迭代順序、系統時間或全域亂數。
7. 卡牌行為以型別安全的 Go 與可重用效果元件實作。
8. 規則與卡牌以 vertical slice 逐項加入，不先建立尚無使用者的通用框架。

## 玩家資訊模型

`CardInstanceID`、`ObjectID` 等引擎內部身分不得出現在 CLI 或 bot 可讀取的資料。Game Module 為每位玩家維護 `KnowledgeState`，並依指定玩家與當前 state revision 產生 `PlayerView`；其中的卡牌與物件只能透過不透明 `ViewHandle` 顯示及提交選擇。

只要規則仍允許該玩家追蹤同一實體，ViewHandle 可以保持穩定。洗牌、隨機放回牌庫或其他使追蹤權消失的操作必須撤銷舊 handle；該卡牌日後重新可見時取得新 handle，操控端不能藉此推斷其隱藏位置。Reveal 等已公開內容保留在該玩家可讀取的事件歷史中，讓玩家回顧依法能記錄的資訊，但歷史紀錄不得重新連結到失去追蹤的實體。

CLI 與 bot 使用完全相同的投影規則。所有行動與選擇提交都必須攜帶產生該 ViewHandle 的 state revision；過期、屬於其他玩家或已失去權限的 handle 一律拒絕，且不改變正式狀態。

## 架構形狀

```text
cmd/grandarchive
      │
      ▼
 CLI Adapter ────────┐
                     ▼
                Game Module ◄──── Bot Adapter
                     │
                     ├─ 單局狀態與身分
                     ├─ 行動與待決選擇
                     ├─ Opportunity 與 Effects Stack
                     ├─ 狀態檢查、階段與戰鬥
                     ├─ 卡牌 registry 與效果元件
                     └─ 玩家視角、事件、hash 與 replay
```

首版維持一個執行檔。Game Module 提供少量 Interface，將規則複雜度藏在 Implementation 中；內部先以同一 Go package 的清楚檔案分組開始，只有當實際依賴方向與第二種 Adapter 出現時才增加 package 或 seam。

概念上的外部操作只有三類：建立單局、取得玩家視角、以當前 state revision 提交玩家行動或選擇。實際 Go 型別與方法名稱由第一個 tracer bullet 驗證後固定，不預先擴大公開 Interface。

## 核心身分模型

- `CardDefinition`：不可變卡面資料與已註冊行為。
- `CardInstance`：單局中的一張卡，跨非場上區域維持身分與所有者。
- `GameObject`：卡牌或 token 位於場上期間的穩定物件身分，保存控制者、counter、狀態與暫時效果；一般 Object 離場後結束存在，重新進場會取得新 ID。
- `StackItem`：Effects Stack 有序結構上的 activation、Materialization、bestowment 或 ability instance，保存來源與已宣告選項。
- `AbilityDefinition`：位於 CardDefinition 或 CardFace 上的不可變能力規則，不保存單局狀態。
- `AbilityInstance`：具有獨立 ID、宿主與生命週期的 runtime 能力；卡牌範圍狀態以 CardInstanceID 與 ability slot 識別，Object 範圍狀態以 ObjectID 與 ability slot 識別，玩家層級的 Mastery、Status 等能力則有自己的 instance。
- `ContinuousEffectInstance`：static ability 或其他來源目前生效的持續效果，保存來源、時間戳、期間與相依資訊，供衍生特徵計算使用。

Champion 在同一條 Lineage 存續期間保持 Object ID；Level Up 只把新 CardInstance 放到 Lineage 頂端並改變代表卡，Inner Lineage 卡牌不是 Object。Transform 只切換有效 CardFaceID。兩者都不清除 Object 依法保留的 orientation、counter、戰鬥角色與 continuous effects。目標與關聯使用 ID，不讓卡牌 handler 長期持有 Go pointer；Object 真正離場時才終止 lifetime 並保存規則需要的 last-known information。

啟動或觸發能力進入 Effects Stack 時建立獨立 StackItem，快照來源參照、操控者、模式、目標及結算所需的特徵或 last-known information；來源之後離場不會自動移除該 StackItem。Static ability 不進入 Effects Stack，而是在有效期間貢獻 ContinuousEffectInstance；衍生特徵由規則層重新計算，不把修正永久寫回基礎資料。延遲觸發與 reflexive trigger 也各自建立 runtime instance。卡牌 handler 只能組合 Game Module 提供的規則操作，不得直接修改 GameState。

Effects Stack zone 中的來源卡與有序 StackItem 必須分開保存。打出卡牌時，CardInstance 成為具有 timestamp、但不參與 FILO 排序的來源卡；原始 activation／Materialization／bestowment 及後續 copy 各自成為 StackItem，透過 source association 指回來源卡。來源卡只在最後一個關聯 StackItem 離開後，才由 state-based checks 移往預設區域或成為 Object。若來源卡先離開，仍引用它的 card-play instances 依法 fizzle；ability StackItem 的來源不會被搬進 Effects Stack，改以 SourceRef 與 LKI 追蹤。

## 行動與選擇協定

1. Game Module 依當前玩家視角提供可開始的玩家行動。
2. 操控端選擇一個行動類型或來源。
3. 引擎依序要求目標、模式、費用、排序等必要資料。
4. 宣告期間不修改正式單局狀態。
5. 所有選擇齊全後，引擎依當前 revision 完整驗證並原子提交。
6. 只能在效果結算期間決定的選項會形成待決選擇，並暫停推進直到指定玩家回答。
7. 過期、非法或取消的輸入不留下部分狀態變更。

上述宣告流程必須由 `DeclarationTransaction` 在隔離候選狀態中執行，而不是逐欄修改正式 GameState。交易保存起始 revision、候選區域移動、已選參數／模式／目標、費用快照、候選事件、trigger buffer 與隔離的 PRNG cursor。費用支付造成新增模式或目標時，依規則重走指定步驟，但不得重算已固定的 cost。

只有完整合法的 activation 才把候選狀態、PRNG cursor 與事件一次提交。取消、過期、無法支付或最終非法時整筆丟棄，視為沒有採取行動，也不產生 trigger 或消耗正式亂數。Activation 一旦成立，已付費用即屬正式狀態；之後 fizzle 或 negate 不回滾費用。

操控端不得藉交易窺視未提交的隱藏 chance outcome。玩家只能在規則仍允許撤回宣告的階段主動取消；一旦進入會取樣隨機結果的費用支付階段，介面只提供規則要求的後續選擇或讓交易依法失敗，不再提供任意取消。最終非法而 rollback 的暫定亂數結果不得投影進 PlayerView 或事件歷史。

提交成功後，由 Game Module 內的 deterministic scheduler 自動執行所有不需要外部決定的規則步驟，持續推進到下一個穩定停點：某位玩家取得 Opportunity、指定玩家必須回答 PendingChoice、單局結束，或遇到 NeedsRuling。CLI 與 bot 不逐步驅動效果指令、觸發整理或 state-based checks。

需要暫停的流程以型別安全的 `ResolutionFrame` 保存在 GameState，明確記錄指令位置、局部結果與後續步驟，不使用 Go closure 保存 continuation。恢復時先驗證 state revision，再確定性地繼續。PendingChoice 的回答者不必是 Opportunity 持有者；等待期間拒絕所有不相關的一般 PlayerAction，但任何尚未落敗的玩家都可提交全域 `Concede`。輸入仍逐一序列處理，不引入平行狀態寫入。

## 事件批次與因果

引擎內部不以扁平的 `[]GameEvent` 驅動規則。每個規則指令建立帶有 batch ID、cause ID、父流程、確定順序與 simultaneous flag 的 `EventBatch`；批次內仍保存各個離散事件。例如「抽三張」產生三個 draw event，但三者保留同一批次身分。

事件發生時先偵測並暫存 triggered abilities，直到規則指定的 checkpoint 才整理並放入 Effects Stack。`whenever X` 可逐事件觸發，`whenever one or more X` 依事件批次聚合；同時傷害維持共同批次，先完成共同計算，再處理 On Hit、On Kill 等後續判定。ResolutionFrame 明確標示效果指令與 checkpoint 邊界，不在每個事件或效果句子之後自行執行 state-based checks。

CLI 可把 committed events 投影成易讀列表，但 canonical state、replay 診斷與規則測試保留必要的批次、因果及順序 metadata；外部顯示格式不作為觸發或結算的真相來源。

## 狀態檢查固定點

Scheduler 到達指定 checkpoint 時，依鎖定官方規則的明確順序執行 state-based checks。同一輪以一致的狀態視圖判斷；依法同時發生的結果以同一 EventBatch 套用。任何檢查造成狀態變化後，從規定起點重新開始，直到完整一輪不再改變狀態。

每一輪先透過中央 evaluator 取得該 canonical state 的 derived-characteristic view，再用同一 view 執行該輪檢查；若 state-based result 改變來源、持續效果或其他基礎狀態，下一輪重新求值。Derived view 是計算結果而不是另一份可獨立修改的 GameState。

檢查期間產生的 triggered abilities 先暫存，fixed point 完成後才進入觸發排序流程。若 Unique 或其他 state-based result 要求玩家決定，scheduler 建立 PendingChoice；取得回答後恢復原 ResolutionFrame 並重新檢查。Fixed point 完成前不得移交 Opportunity 或結算下一個 StackItem，卡牌 handler 也不能自行呼叫、跳過或改變檢查時點。

實作需有確定性的循環偵測與保守的迭代上限。若支援內容造成無法收斂的狀態循環，引擎停止該局並輸出診斷、replay 與 state hash，不讓 CLI 無限執行；這類情況視為規則／實作問題，不猜測一個穩定結果。

## 觸發排序

到達 trigger flush checkpoint 且 state-based fixed point 完成後，引擎把同時待放入 Effects Stack 的 triggered abilities 依玩家分組，從回合玩家開始按 turn order 處理。若同一玩家有多個 triggers，由該玩家透過 PendingChoice 決定自己的放置順序；只有一種順序時自動前進。需要「choose one」模式或目標的 trigger，在放入 Stack 時一併取得並固定選擇；依狀態決定的「depending on」模式留到結算時計算。

所有玩家的 triggers 完成放置後才授予規則指定的 Opportunity。Trigger ordering、模式與目標都是正式輸入並進入 canonical replay；CLI 與 bot 不得自行排序或跳過。

## 持續效果與替代效果求值器

固定牌組未確定前不預先實作完整 evaluator。Support Set 完成後，覆蓋矩陣必須列出所有可達的 characteristic modifier、static effect、permission／prohibition、prevention 與 replacement effect，再逐個加入需要的 typed operation。

只要 Support Set 使用 continuous effects，中央 derived-characteristic evaluator 就必須遵守官方完整排序語意：Layer A 設定基礎數值、費用與預設 play permission；Layer B 修改 type；Layer C 修改 element；Layer D 增減 ability；Layer E 增減數值、費用與 play permission，並包含 power／life 的規定 sub-layer。相同 layer 內先動態解析 dependency，再依 timestamp；dependency loop 依官方 timestamp 規則處理。所有合法性、目標、費用、戰鬥與顯示查詢都走同一 evaluator，不把結果永久寫回 Object。靜態 `can't`／`may not` 高於 `can` permission。

Replacement effects 使用獨立 pipeline，不塞入 layer evaluator。事件或行動提交前先找出目前適用的 replacements；多個適用項依受影響卡牌、Object、zone 等的控制玩家決定次序。每套用一項後重新計算仍適用的候選，直到沒有候選。需要玩家排序時沿用所在流程的選擇協定：宣告交易內產生 draft-scoped ChoiceRequest，已提交的結算流程則建立 GameState 中的 PendingChoice。原始意圖、每次替代與最終 committed events 保留 cause chain。

共用 evaluator 只提供 Support Set 已登記且有測試的 modifier／replacement operation。卡牌要求尚未支援的 operation 時，開局驗證直接拒絕，不能在個別 handler 中加入繞過 layer、permission 或 replacement pipeline 的特例。此決策記錄於 [ADR 0014](./adr/0014-centralize-derived-characteristics-and-replacements.md)。

## 卡牌支援流程

固定牌組確定後，從 [`card-support-matrix-template.md`](./card-support-matrix-template.md) 建立正式卡牌覆蓋矩陣與封閉的 `Support Set`。固定牌組 manifest 必須包含 Main Deck、Material Deck 與明確的 `Outside Game Pool`；Support Set 需遞迴包含所有可能實際建立或執行的 token、生成卡、Mastery、Status、雙面卡面及其他衍生內容。Generate 只可從 Outside Game Pool 選擇，列入該 pool 即代表玩家具備規則要求的數位副本；允許無界生成且沒有封閉 allowlist 的卡牌維持未支援。單純指定卡名但不建立或執行該卡時，只要求完整卡名索引。

每個不同 card ID 或衍生內容至少記錄：

- 鎖定的卡面資料版本與原文。
- 所需卡牌類型、區域、費用、時序、效果與關鍵字。
- 依賴的官方規則檔案與條目。
- 正常、非法、目標失效及跨卡互動案例。
- 實作與測試狀態。

正式 registry 只有「完整支援」與「未支援」兩種狀態。開局驗證必須遞迴檢查完整 Support Set；缺少任一可達內容或必要機制便拒絕開始。test-only fixture 放在測試 package，不能進入正式 registry 或正式 CLI。

## 遊戲機器人

首版只有一種正式啟發式 bot。它只能讀取自己的 `PlayerView` 與合法選項，並按「可立即獲勝、避免立即落敗、有利攻擊、有效使用資源、出牌、推進階段」等明確優先級選擇。相同局面與 seed 必須得到相同決策；平手時只能使用注入的 seeded random source。

Bot 思考可以在 Game Module 外進行，但提交時必須攜帶 state revision。策略必須具備決策或行動上限，避免在合法但無進展的行動間無限循環。

## 命令列介面

CLI 使用引擎驅動的編號選單，不建立自由文字語法或全螢幕 TUI。畫面需清楚顯示回合、階段、Opportunity 擁有者、Champion、手牌、場上物件、Effects Stack、最近遊戲事件與可選行動；只顯示真人依法可見的資訊。

首版命令列至少支援 `--seed` 與 `--replay-out`。選擇目標、模式與費用時沿用相同編號選單。`--replay-out` 產生的 canonical replay 是可能包含完整隱藏資訊的私人診斷產物，CLI 必須清楚提示它不適合公開分享。

## 重播

Canonical replay 包含：

- 引擎版本。
- 規則 commit。
- 卡面資料版本。
- 固定牌組版本。
- PRNG 演算法與版本。
- 初始 seed。
- 按順序提交的玩家行動與選擇。

若某種隨機機制無法固定取樣順序，replay 必須一併記錄實際 chance outcome。每一步可附加 canonical state hash 作為回歸檢查，但完整狀態與衍生遊戲事件不是 replay 的真相來源。舊 replay 必須搭配原本的引擎、資料與 PRNG 版本執行。

首版不提供公開或觀戰 replay。未來若新增分享功能，必須依指定玩家或 spectator 視角建立 redacted projection，並通過與 `PlayerView` 同等的隱藏資訊防洩漏測試；不得直接發布 canonical replay。

## 規則治理

只在當前 vertical slice 需要時閱讀相關規則，不預先完整實作規則庫。每個行為測試需標註規則 commit、來源檔案與條目。

遇到歧義時，依 [`rules-issues.md`](./rules-issues.md) 登錄並尋找官方規則或正式 ruling；沒有裁定便保持未支援。若未知歧義於執行期間出現，引擎停止該局並輸出 `NeedsRuling`、issue ID、replay 與狀態 hash。任何專案自訂裁定必須先由專案負責人確認，另以 ADR 記錄並標示偏離官方規則。目前 RUL-001、RUL-002 與 RUL-004 仍待官方裁定；只阻擋實際觸及它們的正式 slice，不阻擋 test-only walking skeleton。

## 垂直切片就緒門檻

一個正式 slice 只有同時滿足以下條件才能進入 red → green：

1. 規則 commit、卡面資料版本、Content ID 與 Ability Slot 已固定。
2. 該內容已列入 Support Set，所有直接可達依賴都在覆蓋矩陣中。
3. 正常案例、至少一個非法或邊界案例及預期結果已有獨立規則來源。
4. 所需 typed operations、layer／replacement 類型與 KnowledgeState 影響已列出。
5. 沒有被此 slice 觸及且仍為 `待官方裁定` 的規則 issue。
6. 可從正式 Game Module Interface 觀察結果，不要求新增只供測試使用的 production seam。

## 開發里程碑

### 0. 規格入口

取得唯一固定牌組、Outside Game Pool 與卡面版本，從範本建立 `docs/card-support-matrix.md`，遞迴完成 Support Set、規則依賴、evaluator operation 與 ruling gate。起始 Level 0 Champion 排在第一個正式內容節點。若可達內容牽涉未支援機制或未裁定 issue，矩陣維持 `blocked`，再由專案負責人更換牌組、縮小 allowlist、等待 ruling 或明確擴大範圍。本里程碑等待牌組時，不阻擋里程碑 1 的 test-only 骨架。

### 1. 最小端到端骨架

以 test-only fixture 走通「啟動 CLI → 建立單局 → 取得 PlayerView → 真人或 bot 提交一個選擇 → 投降結束 → 寫出 replay → 重播驗證 hash」的完整路徑。Fixture 同時驗證 revision、ViewHandle 與非法輸入無副作用。這是第一個 tracer bullet；正式 registry 與正式 CLI 仍不得接受 fixture 或未完整支援的牌組。

### 2. Standard 開局與回合生命週期

加入 deterministic scheduler、state-based fixed point、Standard 第一回合修正、Wake Up、Materialize、Recollection、Draw、Main、End、Opportunity 與讓過。先以 fixture 驗證流程，但完成條件必須使用固定牌組的 Level 0 Champion On Enter abilities 依 turn order 取得起始手牌；全部完成前不能授予 Opportunity。正式對局可推進並以 deckout 結束。RUL-001 未解決前，本里程碑不能宣告 Opportunity 時序完成。

### 3. 第一張可操作卡牌

在起始 Champion 之後，選一張最簡單且能代表固定牌組依賴的卡牌，走通 DeclarationTransaction、費用與 PRNG rollback、區域移動、activation 或 Materialization、來源卡／StackItem 關聯、逐步選擇、Opportunity、結算、trigger flush、狀態檢查與 fizzle。若是多必要目標卡牌，RUL-002 必須先解決。

### 4. 戰鬥縱切

加入攻擊宣告、Opportunity 回應、Retaliation、EventBatch 同時傷害、On Hit／On Kill、單位死亡與 Champion 敗北，使正式支援內容可以透過正常戰鬥結束。

### 5. 固定牌組擴充

依覆蓋矩陣的 dependency graph，一次完成一張正式卡牌或一個最小跨卡互動。遇到 continuous、permission、replacement、token、Mastery 或 Status 時，在該 slice 內加入中央 evaluator 所需的最小 typed operation 與完整排序測試；不先橫向完成所有效果元件。每完成一列便重新執行 Support Set closure validation。

### 6. 首版收尾

啟用完整固定牌組、移除任何誤入 production 的 fixture 路徑、調整啟發式 bot、改善 CLI 可讀性、完成 replay 回歸、fuzz/property tests、批次 bot 對戰及文件。只有通過 release gate 才發布正式 CLI。

### 7. 首版之後

若需要保存牌組、單局摘要與 replay，新增持久化 use case，並在外部 Adapter 使用 PostgreSQL 與 GORM。只有伺服器、跨程序進行中單局、配對或分散式協調出現具體需求時才評估 Redis。

## 每個垂直切片的強制流程

1. 從覆蓋矩陣選擇一個最小可觀察行為。
2. 以固定規則版本與已知結果撰寫一個失敗測試（red）。
3. 實作使它通過所需的最小行為（green）。
4. 只透過已確認的測試 seam 驗證，不 mock 自有的內部 Module。
5. 加入必要的 replay/state-hash 回歸案例。
6. 執行相關測試與整體測試，再開始下一個 slice。
7. 重構留到 review 階段，不混入 red → green 循環。

詳細測試 seam 與禁止模式見 [`testing.md`](./testing.md)。

## 首版發布門檻

- 固定牌組通過 Standard 格式、Support Set 閉包與 registry 一致性驗證；所有可達內容、Ability Slot 及 evaluator operation 都為 `supported`。
- 所有被 Support Set 觸及的規則 issue 已有可追溯裁定與測試；沒有 runtime fallback、近似行為或靜默 no-op。
- 真人能透過 CLI 與啟發式 bot 完成鏡像對戰。
- 正式 Standard setup 由雙方 Level 0 Champion 的 On Enter abilities 取得起始手牌；第一回合修正與 Opportunity gate 皆有情境測試。
- Champion 死亡、牌庫耗盡與投降都有正式情境測試；等待 PendingChoice 時其他尚未落敗玩家仍可投降。
- 固定牌組所需的 Opportunity、來源卡／StackItem／copy、觸發排序、目標失效、fizzle、state-based fixed point、continuous／replacement 與戰鬥規則都有來源可追溯的測試。
- DeclarationTransaction 對取消、過期、非法及費用不足能完整 rollback state、事件與 PRNG；activation 成立後的 fizzle／negate 不退費。
- 玩家視角不洩漏對手手牌或牌庫順序；洗牌後舊 ViewHandle 無法追蹤卡牌，已公開事件歷史仍可回顧。
- 相同引擎、規則、卡面、牌組及 PRNG 版本，加上相同 seed 與輸入序列，必須逐步產生相同 state hash。
- 至少 100 場不同 seed 的 bot 鏡像對戰在行動上限內結束，沒有 panic、deadlock、scheduler 不收斂、非法提交或 NeedsRuling。
- `go test ./...`、`go test -race ./...` 與設定時限的 fuzz/property tests 全數通過。
- `CONTEXT.md`、ADR、規則 issue、卡牌覆蓋矩陣與 CLI 使用說明同步完成，且文件連結檢查通過。

## 尚待提供的輸入

開始正式卡牌實作前，專案負責人必須提供：

1. 唯一固定牌組的完整 Standard 主牌與 Material Deck 清單。
2. 明確列出的 Outside Game Pool。
3. 每張卡及所有可達內容對應的穩定 ID。
4. 卡面資料來源與欲鎖定的版本或日期。

除此之外，首版核心計畫沒有尚待決定的架構選擇。尚未取得的官方裁定屬外部規則輸入，集中追蹤於 [`rules-issues.md`](./rules-issues.md)，不由實作自行決定。
