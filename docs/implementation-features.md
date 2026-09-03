# Grand Archive 首版功能實作清單

本文件把 [`development-plan.md`](./development-plan.md) 轉成可追蹤的功能目錄與驗收清單。它描述「需要實現什麼」，不預先固定尚未經第一個 tracer bullet 驗證的 Go 公開介面。

首版產品邊界是：一名真人透過 CLI，使用唯一固定合法牌組，與讀取相同玩家視角模型的啟發式 bot 完成一場可重播的雙人 Standard 鏡像對戰。

## 使用方式

- 狀態只使用 `待實作`、`進行中`、`被阻擋`、`完成`。
- 功能只有在其驗收條件與相關測試都通過後才能標為完成。
- 正式卡牌功能必須先列入 `docs/card-support-matrix.md`。目前矩陣已由固定 Main Deck 與 Material Deck 建立，但在 Outside Game Pool、CardFace／Ability Slot ID 與卡面資料版本補齊前維持 `blocked`。
- 功能實作依照垂直切片推進，不應先把本文件的每個章節各做一個不相連的框架。
- 所有規則測試都要標註規則 commit `602c917f2f8fd4df7198429a72eb596bf7f647c6`、來源檔案與條目。

## 功能總覽

| ID | 功能域 | 首版交付能力 | 前置條件 | 初始狀態 |
| --- | --- | --- | --- | --- |
| DECK | 牌組與支援集合 | 載入、驗證唯一固定牌組及完整內容閉包 | 固定牌組與卡面版本 | 被阻擋 |
| GAME | 單局與回合 | 建立並推進雙人 Standard 單局直到結束 | CORE、DECK | 待實作 |
| CORE | 身分、區域與狀態 | 保存權威且可確定重現的單局狀態 | 無 | 待實作 |
| ACTION | 玩家行動與選擇 | 產生合法行動並原子提交逐步宣告 | CORE、VIEW | 待實作 |
| STACK | Effects Stack | 管理來源卡、StackItem、FILO 與結算 | CORE、ACTION | 待實作 |
| ABILITY | 能力與卡牌行為 | 以 typed Go 註冊並執行卡牌能力 | DECK、CORE、STACK | 被阻擋 |
| TRIGGER | 觸發效果 | 偵測、聚合、排序並建立觸發能力 | EVENT、STACK | 待實作 |
| EFFECT | 效果求值 | 執行 typed operations、持續與替代效果 | ABILITY、EVENT | 被阻擋 |
| RULE | 排程與狀態檢查 | 自動推進到穩定停點並求得 fixed point | CORE、EVENT | 待實作 |
| COMBAT | 戰鬥 | 完成攻擊、回應、傷害與死亡流程 | GAME、STACK、TRIGGER | 待實作 |
| VIEW | 玩家資訊與合法選項 | 防止隱藏資訊與內部身分洩漏 | CORE | 待實作 |
| EVENT | 事件與因果 | 保存批次、同時性、順序及原因鏈 | CORE | 待實作 |
| RNG | 確定性亂數 | 可指定 seed 且可回滾、可重播 | CORE | 待實作 |
| REPLAY | 重播與 state hash | 逐步重現輸入並驗證相同狀態 | GAME、RNG、EVENT | 待實作 |
| BOT | 啟發式 bot | 只憑自身 PlayerView 完成對局 | GAME、VIEW | 待實作 |
| CLI | 命令列介面 | 顯示資訊、收集選擇、輸出 replay | GAME、VIEW、REPLAY | 待實作 |
| GOVERN | 規則治理與支援 gate | 未支援或未裁定內容在開局前被拒絕 | DECK、ABILITY | 待實作 |

## DECK：牌組、卡面資料與支援集合

### DECK-01 固定牌組 manifest

- 表示兩位玩家共用的 Main Deck、Material Deck 與 Outside Game Pool。
- 每個項目使用穩定 Card ID／CardFace ID 與數量，不用名稱作為主要識別。
- 記錄牌組版本及其對應的卡面資料版本。
- 首版不提供自由組牌、sideboard 或執行期間換牌。

驗收條件：

- 相同 manifest 可確定性地建立雙方鏡像牌組。
- 缺卡、數量錯誤、格式不合法或版本不符時無法開始單局。
- 未提供 Outside Game Pool 時，不能執行需要從遊戲外 Generate 的牌組。

### DECK-02 Standard 合法性驗證

- 驗證 Main Deck、Material Deck、Champion 與其他 Standard 格式限制。
- 驗證固定牌組中所有 Card ID 都存在於鎖定資料版本。
- 驗證兩位玩家使用完全相同的牌組版本。

### DECK-03 Support Set 閉包

- 從三個牌組區段遞迴展開所有可能建立或執行的內容。
- 納入 token、生成卡、雙面卡面、Mastery、Status、copy／transform／level-up 目標及其他衍生內容。
- 單純被卡名提及但不會被建立或執行的內容，只加入完整名稱索引。
- 無界生成且沒有封閉 allowlist 的卡牌維持未支援。

驗收條件：

- 任一可達內容、Ability Slot、規則機制或 typed operation 未支援時，開局 fail-fast。
- 正式 registry 不存在「部分支援」狀態，只有完整支援或未支援。
- 測試 fixture 不得被正式 manifest 或 CLI 載入。

### DECK-04 卡面資料載入

- 將目前外部 `entity.Card` 資料轉成引擎內不可變的 `CardDefinition` 與 `CardFace`。
- 保留原文、穩定 ID、類型、subtype、element、費用、數值與資料版本。
- 資料載入與單局 runtime 狀態分離。

## CORE：核心身分、區域與單局狀態

### CORE-01 Runtime 身分

- `CardDefinition`：不可變資料與註冊行為。
- `CardInstance`：一張單局卡牌；跨非場上 zone 保持 CardInstanceID 與 owner。
- `GameObject`：卡牌或 token 在 Field 上的存在；一般物件重新進場取得新 ObjectID。
- `AbilityDefinition`：卡面上的不可變能力定義與 Ability Slot。
- `AbilityInstance`：卡牌、Object 或玩家層級能力的 runtime 身分。
- `ContinuousEffectInstance`：持續效果的來源、時間戳、期間與相依資訊。
- `StackItem`：堆疊上的 activation、Materialization、bestowment 或 ability instance。

驗收條件：

- CardInstanceID、ObjectID、AbilityInstanceID 與 StackItemID 不可互換。
- 規則關聯只保存 typed ID／SourceRef，不讓 handler 長期持有 Go pointer。
- 物件離場時保存規則需要的 Last-Known Information。

### CORE-02 Champion Lineage 與 Transform

- Champion 在同一 Lineage 期間維持相同 ObjectID。
- Level Up 將新 CardInstance 放到 Lineage 頂端，Inner Lineage 不屬於 Field Object。
- Transform 只切換有效 CardFaceID。
- Level Up 與 Transform 依法保留 orientation、counter、戰鬥角色及 continuous effects。

### CORE-03 Zone 模型

- 表示 Main Deck、Material Deck、Hand、Memory、Field、Graveyard、Banishment、Intent、Pantheon 與 Effects Stack。
- Standard 首版不啟用 Pantheon gameplay，但 zone vocabulary 保持與鎖定規則一致。
- 所有 zone move 都透過規則操作產生事件，不允許卡牌 handler 直接改 slice 或 map。
- Token 離開允許存在的區域後，由 state-based checks 依法終止存在。

### CORE-04 GameState

- 保存玩家、卡牌、物件、zone、回合、階段、Opportunity、堆疊、trigger buffer、scheduler／resolution frame、KnowledgeState、PRNG cursor 與 revision。
- Game Module 是唯一可改變 GameState 的元件。
- 每次成功提交後 revision 單調遞增；失敗輸入不改變 revision。
- canonical state 序列化不得依賴 Go map 迭代順序。

## GAME：單局建立、回合與結束

### GAME-01 建立 Standard 單局

- 接受固定牌組版本、規則／資料／引擎／PRNG 版本及 seed。
- 完成 Support Set 與規則 issue gate 後才建立正式單局。
- 建立兩名玩家、牌組實例、初始 zone、Champion Lineage、KnowledgeState 與初始 scheduler frame。

### GAME-02 Standard 開局

- 依鎖定規則完成玩家順序、洗牌及起始設置。
- 由雙方 Level 0 Champion 的 On Enter abilities，按 turn order 取得起始手牌。
- 所有強制開局流程完成前不得授予 Opportunity。
- 實作第一回合修正。

### GAME-03 回合與階段

- Wake Up、Materialize、Recollection、Draw、Main 與 End。
- 區分回合玩家與非回合玩家可開始的行動。
- 階段推進由玩家行動與 scheduler 共同控制，不由 CLI 直接修改階段。
- 回合切換時處理到期效果、戰鬥狀態及每回合使用次數。

### GAME-04 Opportunity 與讓過

- Game Module 明確保存當前 Opportunity 擁有者與連續讓過狀態。
- 玩家可取得的行動由當前時序、速度、階段、堆疊與 permission 決定。
- 所有人依法讓過後，結算最上方 StackItem 或推進規則流程。
- 依 [ADR 0015](./adr/0015-retain-opportunity-until-the-holder-passes.md)，玩家完成需要 Opportunity 的行動後保有 Opportunity，直到讓過才依 turn order 移交。

### GAME-05 單局結果

- Champion 死亡導致玩家落敗。
- 必須抽牌但牌庫耗盡時落敗。
- 任一尚未落敗玩家可隨時投降，包括等待 PendingChoice 時。
- 單局結束後拒絕一般行動，保留最終事件、state hash 與 replay 輸出。

## ACTION：玩家行動、宣告交易與待決選擇

### ACTION-01 合法行動列舉

- 依指定 PlayerView 與 revision 列出可開始的打牌、啟動能力、攻擊、推進階段、讓過及投降等行動。
- 合法選項只能包含該玩家當下可見且可選的 ViewHandle。
- CLI 與 bot 消費相同的合法行動資料，不各自重算規則。

### ACTION-02 DeclarationTransaction

- 以隔離候選狀態逐步收集 cost parameter、mode、target、排序與付款方式。
- 保存起始 revision、候選 zone move、費用快照、候選事件、trigger buffer 與隔離 PRNG cursor。
- 所有選擇完成後重新驗證並一次提交。
- 費用造成新增 mode／target 時，依法重走指定步驟但不重算已固定 cost。

驗收條件：

- 取消、過期、非法目標、無法支付與最終非法都不改變正式 state、事件、trigger 或 PRNG cursor。
- activation 成立後才提交費用；後續 fizzle 或 negate 不退費。
- 開始取樣隨機費用後，不再提供規則以外的任意取消。
- 未提交的 chance outcome 不會出現在 PlayerView 或事件歷史。

### ACTION-03 PendingChoice 與 ResolutionFrame

- 只能在結算時決定的 mode、target、排序、replacement 順序與 Unique 等決策建立 PendingChoice。
- ResolutionFrame 以型別安全資料保存指令位置、局部結果、checkpoint 與後續步驟，不使用 closure。
- 回答者可不是 Opportunity 擁有者；等待期間拒絕不相關行動但允許投降。
- 所有回答都驗證 player、revision、choice ID 與選項合法性。

## STACK：Effects Stack 與結算

### STACK-01 來源卡與 StackItem 分離

- 打出的 CardInstance 作為 Effects Stack zone 中的 Source Card，具有 timestamp 但不參與 FILO 排序。
- 每個 activation、Materialization、bestowment、ability 與 copy 建立獨立 StackItem。
- StackItem 保存 SourceRef、操控者快照、mode、target、cost 與結算需要的 LKI。
- 來源離開後，ability StackItem 仍可依法結算；card-play instance 則依法判斷 fizzle。

### STACK-02 堆疊生命週期

- 完成宣告後放入 StackItem，重設 Opportunity 流程。
- 所有玩家讓過後以 FILO 結算頂端 StackItem。
- 結算可建立子 ResolutionFrame、事件、觸發與 PendingChoice。
- 最後一個關聯 StackItem 離開後，state-based checks 才移動 Source Card 或建立 Field Object。
- 支援 negate、copy 與來源消失後的處理；只實作 Support Set 實際需要的 typed operation。

### STACK-03 Fizzle 與部分目標失效

- 結算前重新檢查目標合法性。
- 依正式裁定區分完整 fizzle、仍對合法目標結算及不受影響的無目標效果。
- 依 RUL-002 的官方裁定，只有所有必要目標都非法或不存在時完整 fizzle；仍有合法必要目標時對其結算。

## ABILITY：能力模型與卡牌行為

### ABILITY-01 能力類型

- Activated ability：玩家依法宣告並支付費用後進入堆疊。
- Triggered ability：由事件偵測並於 checkpoint 整理入堆疊。
- Static ability：在有效期間貢獻 permission 或 ContinuousEffectInstance，不進入堆疊。
- Delayed／reflexive trigger：具有獨立 runtime instance 與明確生命週期。
- On Enter、On Leave、On Death、On Hit、On Kill 等只能透過共用事件與觸發機制實作。

### ABILITY-02 卡牌 registry

- 以 Card ID、CardFace ID 與 Ability Slot 註冊 typed Go handler。
- 每個 handler 宣告其規則版本、所需機制、typed operations 與衍生內容依賴。
- handler 只能呼叫 Game Module 提供的規則操作，不能直接修改 GameState。
- registry 在開局前與 Support Set 雙向檢查，不允許遺漏或孤立的正式內容。

### ABILITY-03 常用效果元件

按 Support Set 的實際需要逐項加入，至少預留以下功能分類：

- 抽牌、棄牌、搜尋、展示、洗牌與移動卡牌。
- 產生／消耗 Memory、支付 reserve 或其他費用。
- 造成、預防與修改傷害；恢復生命。
- 建立、移動、destroy、banish 或變更 Object 狀態。
- 選擇玩家、卡牌、Object、ability 或 StackItem 作為目標。
- 加減 counter、改變 orientation 或設定回合內狀態。
- Generate、建立 token、Mastery、Status、copy、Transform 與 Level Up。
- 增減 ability、permission／prohibition、數值或其他 characteristics。

此列表不是預先全部實作的要求；只有固定 Support Set 可達的元件才進入首版。

## EVENT：遊戲事件、批次與因果

### EVENT-01 EventBatch

- 每個規則指令建立具有 batch ID、cause ID、父流程、順序與 simultaneous flag 的批次。
- 批次可包含多個離散 GameEvent；例如抽三張是三個事件、同一批次。
- 同時傷害先共同計算與提交，再處理衍生的 On Hit／On Kill 判定。
- canonical state 與測試保留 metadata；CLI 可投影成簡化文字。

### EVENT-02 Cause chain

- 玩家行動、StackItem、效果指令、replacement、最終事件與觸發之間可追溯。
- 原始意圖被 replacement 改寫後仍保留每次替代步驟。
- trigger 不得以 CLI 顯示文字或事後掃描最終狀態作為唯一來源。

## TRIGGER：觸發偵測、聚合與排序

### TRIGGER-01 觸發偵測與暫存

- GameEvent 發生時偵測符合條件的 AbilityDefinition／AbilityInstance，先存入 trigger buffer。
- `whenever X` 可按事件觸發；`whenever one or more X` 依 EventBatch 聚合。
- 離場相關觸發使用 LKI 判斷來源及條件。
- 卡牌 handler 不得自行決定 trigger flush 時點。

### TRIGGER-02 Checkpoint flush

- 只有指定 RuleCheckpoint 且 state-based fixed point 完成後，才能把 trigger 放入 Effects Stack。
- 同時 triggers 先依玩家分組，從回合玩家開始按 turn order 處理。
- 同一玩家有多個 triggers 時建立 PendingChoice 取得放置順序；唯一順序時自動推進。
- `choose one` mode 與 target 在入堆疊時固定；`depending on` 在結算時計算。
- 排序、mode 與 target 都寫入 canonical replay。

## RULE：確定性排程與 state-based checks

### RULE-01 Deterministic scheduler

- 每次成功輸入後自動執行所有不需外部決定的步驟。
- 只在 Opportunity、PendingChoice、單局結束或 NeedsRuling 讓出控制權。
- 明確表示目前流程與 checkpoint，不由 CLI 或 bot 逐指令驅動。
- 設定步驟／行動上限並輸出診斷，避免無進展循環。

### RULE-02 State-based fixed point

- 每輪先取得單一 derived-characteristic view，再依官方順序判斷所有檢查。
- 同時結果使用同一 EventBatch 套用。
- 任一結果改變狀態後，從規定起點重跑，直到完整一輪無變化。
- 期間產生的 trigger 先暫存；fixed point 完成前不結算下一項或移交 Opportunity。
- Unique 等需要決定的結果建立 PendingChoice，回答後恢復原檢查流程。
- 具備確定性的循環偵測及保守迭代上限；不收斂時停止並輸出診斷、replay 與 state hash。

## EFFECT：衍生特徵、持續效果與替代效果

### EFFECT-01 中央 derived-characteristic evaluator

- 所有合法性、target、cost、戰鬥與畫面顯示都使用同一 evaluator。
- 結果是指定 canonical state 的唯讀計算值，不回寫 CardDefinition 或 GameObject 基礎資料。
- Support Set 使用 continuous effect 時，需支援完整官方順序：Layer A 基礎值／費用／預設 play permission、B type、C element、D ability、E 數值／費用／play permission，以及 power／life sub-layer。
- 同 layer 內動態解析 dependency，再依 timestamp；dependency loop 依官方 timestamp 規則處理。
- `can't`／`may not` 高於 `can` permission。

### EFFECT-02 Replacement pipeline

- 在行動或事件正式提交前找出目前適用的 replacement／prevention。
- 每套用一項後重新計算候選，直到沒有適用項。
- 多個 replacement 的順序由受影響實體的控制玩家依法選擇。
- 宣告期間的選擇屬 draft-scoped ChoiceRequest；已提交流程則使用 PendingChoice。
- replacement 使用獨立 pipeline，不偽裝成 continuous layer 或卡牌 handler 特例。

## COMBAT：戰鬥

### COMBAT-01 攻擊宣告

- 依回合、階段、orientation、單位狀態、目標及 permission 列出合法攻擊。
- 以 DeclarationTransaction 選擇攻擊者、目標及必要選項後原子提交。
- 攻擊宣告後建立規則要求的 Opportunity 回應流程。

### COMBAT-02 Retaliation 與傷害

- 判斷是否可 Retaliate 及傷害來源。
- 同時戰鬥傷害以單一 EventBatch 計算並提交。
- 使用 derived characteristics 取得 power／life 與適用修正。
- 傷害可經 prevention／replacement pipeline 處理。

### COMBAT-03 戰鬥結果

- 在正確 checkpoint 執行 On Hit、On Kill、死亡、destroy、離場及 Champion 敗北。
- 使用 LKI 支援來源已離場仍需判斷的 trigger。
- 戰鬥結束清除依法到期的角色與暫時狀態。

## VIEW：玩家資訊、ViewHandle 與隱藏資訊

### VIEW-01 KnowledgeState

- 分別記錄每位玩家當前的查看權及追蹤權。
- 自己的手牌、依法公開的牌、已 reveal 的歷史與對手隱藏內容分開處理。
- 洗牌、隨機放回等失去追蹤權的操作撤銷舊關聯；重新可見時取得新 handle。

### VIEW-02 PlayerView

- 顯示回合、階段、Opportunity、Champion、手牌、Field、Effects Stack、最近可見事件、PendingChoice 與合法行動。
- 只包含指定玩家依法可知的資料，不含完整 GameState、牌庫順序或對手手牌。
- 引擎內部 CardInstanceID、ObjectID 等不出現在任何操控端資料。
- CLI 與 bot 使用同一投影，不提供 bot 特權視角。

### VIEW-03 ViewHandle

- handle 綁定玩家、state revision、可見實體與追蹤生命週期。
- 合法選擇提交使用 handle，不接受引擎內部 ID。
- 過期、跨玩家、偽造或已撤銷 handle 一律拒絕且無副作用。
- 已公開事件歷史可保留卡名，但不能藉舊 handle 重新定位洗牌後實體。

## RNG：確定性亂數

### RNG-01 版本化 PRNG

- 單局建立時注入 seed、演算法名稱與版本。
- 固定 shuffle 與所有隨機取樣的消耗順序。
- 不使用系統時間、全域亂數或 Go map 迭代順序決定結果。
- bot 平手亂數使用獨立且可重現的注入來源，不能讀取隱藏遊戲資訊。

### RNG-02 交易與重播

- DeclarationTransaction 使用隔離 cursor，只有成功提交才推進正式 cursor。
- 若某機制無法保證固定取樣順序，canonical replay 額外記錄實際 chance outcome。
- 相同版本、seed 與輸入序列逐步產生相同結果與 state hash。

## REPLAY：canonical replay 與狀態雜湊

### REPLAY-01 Replay 格式

- 記錄引擎、規則 commit、卡面資料、固定牌組及 PRNG 的版本。
- 記錄初始 seed、依序成功提交的玩家行動與所有 PendingChoice 回答。
- 記錄 trigger ordering、mode、target、replacement ordering 及必要 chance outcome。
- 可選擇在每步附 state hash；完整 state 與衍生事件不是 replay 真相來源。

### REPLAY-02 重播執行

- 使用 replay 指定的版本重新建立單局並依序提交輸入。
- 每步驗證 revision、輸入合法性與預期 hash，差異時輸出第一個分歧點。
- 舊 replay 只能搭配相容的原始引擎、資料與 PRNG 版本，不加入靜默 migration。
- canonical replay 可能含隱藏資訊，首版只作私人診斷，不提供公開分享或 spectator 模式。

### REPLAY-03 Canonical state hash

- 為所有影響未來規則結果的權威狀態建立穩定序列化。
- 排除 CLI 文字、非規則性 cache 及不穩定記憶體資訊。
- derived view 若可由 canonical state 重算，不作為獨立真相來源。

## BOT：啟發式機器人

### BOT-01 合法決策

- 只接收 bot 自己的 PlayerView、合法行動及 PendingChoice。
- 每次提交攜帶該 view 的 revision 與 ViewHandle。
- 不得讀取 GameState、對手手牌、牌庫順序或 canonical replay。

### BOT-02 決策策略

- 依序評估：可立即獲勝、避免立即落敗、有利攻擊、有效使用資源、出牌、推進階段。
- 相同 view 與 seed 產生相同決定；平手才使用 seeded random source。
- 為每次決策與整場行動設定上限，偵測合法但無進展的循環。
- 針對 card-specific choice，從引擎提供的合法選項評分，不自行重做規則判定。

## CLI：命令列介面

### CLI-01 單局啟動

- 啟動唯一固定牌組的人類對 bot Standard 單局。
- 支援 `--seed` 與 `--replay-out`。
- 開局 gate 失敗時列出缺少的 Content ID、Ability Slot、typed operation 或 RUL issue。
- replay 輸出時提示檔案可能包含隱藏資訊，不適合公開分享。

### CLI-02 玩家畫面

- 清楚顯示回合、階段、Opportunity 擁有者、Champion、自己的手牌、場上物件、Effects Stack、最近事件與單局結果。
- 使用引擎產生的編號選單顯示合法行動、mode、target、費用及排序。
- 不建立自由文字命令語法或全螢幕 TUI。
- 不顯示任何無法從真人 PlayerView 取得的資料。

### CLI-03 錯誤與中止

- 非法／過期輸入顯示原因並重新取得最新 PlayerView，不自行修補行動。
- NeedsRuling、scheduler 不收斂或 replay 分歧時輸出可追查的 issue、state hash 與診斷檔位置。
- EOF／中斷不可留下半提交的 DeclarationTransaction。

## GOVERN：規則治理、支援 gate 與品質

### GOVERN-01 規則版本與可追溯性

- 引擎版本固定對應規則 commit，不在同一版本混用不同規則快照。
- 每張正式卡牌、Ability Slot、typed operation 與情境測試都可追溯到規則及卡面來源。
- 規則歧義登錄於 [`rules-issues.md`](./rules-issues.md)，不得在 handler 中默默猜測。

### GOVERN-02 NeedsRuling

- 開局前若 Support Set 觸及未裁定 issue，拒絕建立單局。
- 若執行期間遇到未預期的歧義，停止該局並輸出 issue ID、replay 與 state hash。
- 專案自訂裁定只有經負責人確認並以 ADR 記錄後才能進入正式行為。

### GOVERN-03 測試層級

- 每個 vertical slice：先寫有規則來源的失敗情境，再完成最小實作。
- Game Module 測試只走建立單局、取得 PlayerView、提交行動／選擇的正式 seam。
- 加入非法輸入無副作用、rollback、LKI、觸發排序、fixed point、隱藏資訊及 replay hash 測試。
- 執行 `go test ./...`、`go test -race ./...` 與有時限的 fuzz／property tests。
- 發布前至少執行 100 場不同 seed 的 bot 鏡像對戰，全部在行動上限內結束且無 panic、deadlock、非法提交、NeedsRuling 或不收斂。

## 建議的實作檔案分組

以下只是 Game Module 內部的起始分組，目的在分離責任；第一個 tracer bullet 可依實際依賴調整名稱。首版不因此建立額外 package、repository 或抽象介面。

```text
cmd/grandarchive/
  main.go                    # CLI composition root

internal/game/
  game.go                    # 少量公開 use cases
  state.go                   # canonical GameState 與 revision
  identity.go                # runtime IDs、SourceRef、LKI
  zones.go                   # zone 與移動規則
  setup.go                   # Standard 開局
  turn.go                    # 回合與階段
  scheduler.go               # 推進到 Stable Yield Point
  state_checks.go            # state-based fixed point
  actions.go                 # 合法 PlayerAction
  declaration.go             # DeclarationTransaction
  choices.go                 # PendingChoice、ResolutionFrame
  opportunity.go             # Opportunity 與 pass 流程
  stack.go                   # Source Card 與 StackItem
  events.go                  # EventBatch、GameEvent、cause chain
  triggers.go                # 偵測、buffer、排序與 flush
  characteristics.go         # derived-characteristic evaluator
  replacements.go            # replacement/prevention pipeline
  combat.go                  # 攻擊、Retaliation、傷害與死亡
  knowledge.go               # KnowledgeState 與 handle 壽命
  view.go                    # PlayerView 投影
  random.go                  # 版本化 PRNG
  replay.go                  # canonical replay
  hash.go                    # canonical state hash
  support.go                 # Support Set 與開局 gate
  registry.go                # CardDefinition／ability registry
  operations.go              # 共用 typed rule operations

internal/game/cards/
  <card-id>.go               # 每張正式卡牌的 typed 行為

internal/bot/
  bot.go                     # 啟發式策略，只依賴 PlayerView

internal/cli/
  cli.go                     # 編號選單與可見文字投影
```

目前 [`internal/entity/card.go`](../internal/entity/card.go) 應視為外部卡面資料結構；它不足以同時代表 `CardDefinition`、`CardInstance`、`GameObject` 或 `StackItem`，不可直接作為單局狀態模型。

## 里程碑對應

| 里程碑 | 應完成的主要功能 |
| --- | --- |
| 0. 規格入口 | 固定 Main／Material Deck 與初版矩陣已完成；補齊 DECK-01～04、ABILITY-02、GOVERN-01～02 的其餘資料與 gate |
| 1. 最小端到端骨架 | CORE-01／04、ACTION-01／03 的最小 fixture、VIEW-01～03、RNG、REPLAY、CLI 啟動與投降 |
| 2. Standard 生命週期 | GAME-01～05、RULE、EVENT、Level 0 Champion On Enter 與 deckout |
| 3. 第一張可操作卡 | ACTION-02、STACK、ABILITY、TRIGGER，以及該卡所需的最小 EFFECT operation |
| 4. 戰鬥縱切 | COMBAT-01～03、戰鬥事件、On Hit／On Kill、死亡與 Champion 敗北 |
| 5. 固定牌組擴充 | 按 Support Set dependency graph 逐列完成 ABILITY／EFFECT 與跨卡互動 |
| 6. 首版收尾 | BOT、CLI 可讀性、完整 replay 回歸、fuzz/property、100 場批次對戰及文件 gate |

## 目前阻擋項目

正式卡牌與完整首版目前不能排出最終工作量，直到補齊：

1. 明確的 Outside Game Pool，包含 `Rile the Abyss` 的 Card JSON 與數量。
2. 每張卡及所有可達內容的 CardFace／Ability Slot ID。
3. 卡面資料來源的鎖定版本／日期與可驗證 manifest。

規則裁定目前不再構成 blocker：RUL-001 依 [ADR 0015](./adr/0015-retain-opportunity-until-the-holder-passes.md) 採用專案自訂裁定，RUL-002 與 RUL-004 已由官方來源解決。里程碑 1 的 test-only walking skeleton 可先開始，但不能誤列為正式卡牌支援。
