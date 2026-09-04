# Grand Archive 遊戲領域

本文件定義 Grand Archive 遊戲規則中共用且一致的領域語言。

## 用語

**單局（Game）**：
從遊戲設置開始，到規則判定勝、負或和局為止的一次完整遊玩。
_避免稱為_：系列賽、回合

**系列賽（Match）**：
由一場或多場單局組成的競賽結構，例如三戰兩勝。
_避免稱為_：單局

**標準單局（Standard Game）**：
採用標準雙人設置與規則、而非多人 Pantheon 規則的單局。
_避免稱為_：Pantheon 單局

**玩家（Player）**：
擁有牌組，並在單局中操控 Champion 與其他遊戲物件的參與者。

**Champion**：
在單局中代表玩家的單位；擊敗對手的所有 Champion 是使該玩家落敗的主要方式。

**Lineage**：
代表一個 Champion 的完整卡牌集合，由目前位於頂端、代表 Champion Object 的卡牌及其下方 Inner Lineage 組成。

**Inner Lineage**：
Lineage 中除頂端 Champion 卡牌外的所有卡牌；這些卡牌不是遊戲物件，也不位於 Field。

**Level Up**：
將符合條件的下一等級 Champion 卡牌放到 Lineage 頂端；這會改變代表 Champion 的卡牌，但不建立新的遊戲物件。
_避免稱為_：建立新 Champion、Materialization

**Transform**：
將雙面卡牌 Object 翻到另一個有效卡面；它改變有效 characteristics，但不建立新的遊戲物件。
_避免稱為_：離場再進場

**回合玩家（Turn Player）**：
掌控當前回合，並可在規則允許時於自己的主要階段執行慢速玩家行動的玩家。
_避免稱為_：主動玩家

**非回合玩家（Non-Turn Player）**：
當前不是回合玩家，因而不能執行慢速玩家行動的玩家。
_避免稱為_：對手

**玩家行動（Player Action）**：
玩家依規則許可主動作出的選擇，例如打出卡牌、啟動能力、宣告攻擊、推進回合階段或讓過 Opportunity。
_避免稱為_：命令、移動

**遊戲事件（Game Event）**：
規則可觀察到的一個離散事件；一次玩家行動在結算時可以產生多個遊戲事件。
_避免稱為_：玩家行動

**Opportunity（行動機會）**：
規則授予特定玩家的行動權，使其能在遊戲繼續或效果堆疊頂端結算前執行符合條件的玩家行動。玩家成功完成一個需要 Opportunity 的玩家行動後仍是 Opportunity 持有者；只有在該玩家讓過後，Opportunity 才依 turn order 交給下一位玩家。所有玩家連續讓過完整一輪後，才結算 Effects Stack 頂端或在 Stack 為空時推進規則流程。
_避免稱為_：優先權、回應窗口

**效果堆疊（Effects Stack）**：
所有玩家共用的公開 zone，包含不參與 FILO 排序的 Source Card，以及由 activation、Materialization、bestowment 與 ability 組成的有序 Stack Item；所有玩家依序讓過 Opportunity 後，結算最頂端項目。
_避免稱為_：佇列

**來源卡（Source Card）**：
被打出後位於 Effects Stack zone、供一個或多個 activation／Materialization／bestowment instance 讀取的 CardInstance。來源卡具有 timestamp 但不參與 StackItem 的先進後出排序；最後一個關聯 instance 離開後才依法移往目的區域或代表新 Object。
_避免稱為_：StackItem、堆疊層

**卡牌（Card）**：
位於手牌、牌庫、墓地、Memory、Banishment 或其他卡牌專屬區域中的卡片；卡牌位於場上時由遊戲物件表示。
_避免稱為_：遊戲物件

**卡牌定義（Card Definition）**：
由穩定 card ID 識別的不可變卡面資料與已註冊行為；它不是單局中的實體卡，也不保存 runtime 狀態。

**卡牌定義 ID（Card ID）**：
跨印刷版本識別同一 Card Definition 的穩定內容身分；它不識別特定印刷品、圖像或單局中的卡牌實例。
_避免稱為_：印刷版本 ID、CardInstanceID

**卡面（Card Face）**：
 CardDefinition 中由穩定 face ID 識別的一面；Transform 改變 Object 的有效卡面，但不改變 CardInstanceID 或 ObjectID。

**卡面 ID（CardFace ID）**：
在一個 Card Definition 中識別特定正面或背面的穩定內容身分；它不隨印刷版本、卡名翻譯或單局狀態改變。
_避免稱為_：印刷版本 UUID、CardInstanceID

**卡牌實例（Card Instance）**：
單局中一張實體或數位卡牌的身分，具有 CardInstanceID 與 owner，並跨非場上 zone 保持身分；位於 Field 時由 GameObject 表示。

**遊戲物件（Object）**：
卡牌或 token 位於場上時的實例表示，包括 Champion、Ally、Weapon、Item、Domain 與 Phantasia；Level Up 或 Transform 不會改變該 Object 的身分。
_避免稱為_：卡牌實例、堆疊項目

**單位（Unit）**：
遊戲物件的子集合，只包含 Ally 與 Champion。

**堆疊項目（Stack Item）**：
位於 Effects Stack 有序結構上的 activation、Materialization、bestowment 或 ability instance；它不是 Source Card，也不是場上的遊戲物件。
_避免稱為_：遊戲物件、卡牌

**來源參照（Source Ref）**：
StackItem 或 effect 指向其規則來源的型別化 ID 關聯，可引用 Source Card、CardInstance、GameObject、AbilityInstance 或其 Last-Known Information；不保存長期 Go pointer。

**能力定義（Ability Definition）**：
附著於 CardDefinition 或 CardFace、描述能力規則與執行方式的不可變定義；它不保存單局中的使用次數、目標或其他可變狀態。

**能力實例（Ability Instance）**：
單局中實際存在且具有獨立 ID、宿主、作用範圍與生命週期的能力。卡牌範圍的狀態以 CardInstanceID 與 ability slot 識別；Object 範圍的狀態以 ObjectID 與 ability slot 識別；Mastery、Status 等玩家層級能力則有自己的 runtime identity。

**持續效果實例（Continuous Effect Instance）**：
由 static ability 或其他規則來源產生、在特定期間參與衍生特徵計算的 runtime 實體，保存來源、時間戳、期間及必要的相依資訊；它不永久改寫卡牌或 Object 的原始資料。

**能力欄位（Ability Slot）**：
rules-bearing behavior 在特定 CardFace 中的穩定語意位置，用來區分同一來源上的多個能力及追蹤其作用範圍內的狀態；它不是段落序號或 runtime AbilityInstanceID。
_避免稱為_：能力索引、AbilityInstanceID

**最後已知資訊（Last-Known Information）**：
卡牌或遊戲物件離開規則所檢查的區域前所具有的資訊；它只追溯最近一次區域變更之前的狀態。
_避免稱為_：歷史狀態

**玩家視角（Player View）**：
Game Module 依指定玩家與 state revision 投影出的唯一操控端資料，包含該玩家當下依法可見的資訊、可用 ViewHandle 與合法選項；它不暴露引擎內部 ID 或完整 GameState。

**可見控制代號（View Handle）**：
只在特定玩家視角及可追蹤期間內有效、用於顯示與提交選擇的不透明代號。洗牌、隨機放回等使玩家失去追蹤權的操作會撤銷舊代號；它不是 CardInstanceID 或 ObjectID。
_避免稱為_：CardInstanceID、永久卡牌 ID

**知識狀態（Knowledge State）**：
每位玩家在當下依法具有的查看與追蹤權限。曾公開的資訊可保留於可見事件歷史，但歷史資訊不會讓玩家重新關聯已因洗牌或隨機化而失去追蹤的卡牌。

**待決選擇（Pending Choice）**：
規則流程無法自行繼續時，要求一位或多位指定玩家回答的外部決定；回答者不一定是回合玩家或 Opportunity 持有者。等待期間一般玩家行動停止，但任何尚未落敗的玩家仍可投降。

**結算框架（Resolution Frame）**：
GameState 中型別安全且可重現的暫停流程資料，記錄目前規則指令位置、局部結果及後續步驟；它不是 Go closure。

**穩定停點（Stable Yield Point）**：
Deterministic scheduler 已完成目前所有強制規則推進，只能等待 Opportunity 下的玩家行動、Pending Choice、單局結束或 NeedsRuling 的狀態。

**事件批次（Event Batch）**：
由同一規則指令產生、具有共同原因與明確同時性邊界的一組離散 Game Event；例如抽三張牌包含三個 draw event，但仍屬同一批次。

**規則檢查點（Rule Checkpoint）**：
規則允許 scheduler 整理暫存觸發、執行 state-based checks 或推進下一階段的明確邊界；不是每個 Game Event 或效果句子之後都存在檢查點。

**狀態檢查固定點（State-Based Fixed Point）**：
Scheduler 在規則 checkpoint 依官方順序反覆執行 state-based checks，直到完整一輪不再改變狀態。完成前不移交 Opportunity，也不結算下一個 StackItem；需要玩家決定時可暫停為 PendingChoice，回答後再恢復檢查。

**宣告交易（Declaration Transaction）**：
玩家打牌或啟動能力時，在隔離候選狀態中依規則完成參數、模式、目標、合法性、費用試算與支付的流程。交易失敗或取消時，正式 GameState、PRNG cursor、事件與觸發均不改變；正式 activation 成立後才原子提交。

**單局狀態（Game State）**：
Game Module 內唯一具權威性的單局資料，包含 runtime 身分、zone、scheduler frame、KnowledgeState、PRNG cursor 與單調遞增的 state revision；操控端不能直接取得或修改它。

**衍生特徵（Derived Characteristics）**：
從卡面與 Object 基礎資料套用目前有效的 continuous effects、官方 layer、timestamp 及 dependency 後得到的規則查詢結果；它不會永久覆寫基礎資料。

**替代管線（Replacement Pipeline）**：
在玩家行動或 Game Event 正式提交前，依規則找出適用 replacement effects、取得必要排序選擇、逐一套用並重新計算剩餘候選的流程；它與 continuous layer evaluator 是不同機制。

**支援集合（Support Set）**：
由固定 Main Deck、Material Deck 與 Outside Game Pool 出發，遞迴包含所有可實際建立或執行的卡牌、卡面、token、Mastery、Status 及其他衍生內容的封閉集合。

**需要裁定（Needs Ruling）**：
鎖定規則快照與正式 ruling 仍無法唯一決定行為時的停止結果。它不是勝負結果；相關內容維持未支援，並輸出規則項目、replay 與 state hash 供後續裁定。
