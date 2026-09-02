# 卡牌遊戲引擎模式研究

本文件以一手來源檢查本專案目前已確認的核心決策，並刻意區分「來源明確展示的事實」與「對本專案的推論」。它不是建議採用下列引擎，也不把其他專案的規模或歷史包袱照搬進來。

- 研究日期：2026-08-31
- 研究對象：boardgame.io、Google DeepMind OpenSpiel、XMage、Forge
- 來源限制：只使用官方文件、官方 GitHub repository 與其中的原始碼
- 本專案對照基準：[權威確定性規則引擎](../adr/0001-authoritative-deterministic-rule-engine.md)、[逐步選擇與原子行動](../adr/0004-use-guided-choices-and-atomic-actions.md)、[Go 卡牌行為](../adr/0005-implement-card-behavior-in-go.md)、[玩家視角](../adr/0006-expose-player-scoped-views-to-controllers.md)、[版本化輸入 replay](../adr/0008-record-replays-as-versioned-inputs.md)、[單局序列化](../adr/0010-serialize-state-transitions-within-each-game.md)與[測試策略](../testing.md)

## 結論摘要

目前架構方向與成熟引擎反覆出現的模式一致：由單一權威核心依序接受合法輸入、把完整狀態投影成玩家視角、集中管理隨機性，並以情境測試與完整 playthrough 驗證規則。

研究同時揭露三個需要在實作規格中明文化的條件：

1. `Apply` 必須自行重新驗證 actor、state revision 與行動合法性，不能把先前取得的 `LegalActions` 當成授權。
2. 只記錄 seed 仍不足以跨版本重播；PRNG 演算法及其取樣順序必須是引擎版本契約的一部分。若日後無法保證，replay 還要記錄實際 chance outcome。
3. Canonical replay、事件紀錄與公開觀戰紀錄是不同視角；公開輸出必須移除隱藏選擇與卡牌資訊，不能直接發布完整輸入序列。

## 1. boardgame.io：權威 reducer、玩家投影與受控隨機性

### 來源事實

- boardgame.io 把遊戲狀態分為遊戲資料 `G` 與框架中繼資料 `ctx`。玩家 move 是改變 `G` 的函式，文件要求 move 不依賴外部狀態且不產生修改 `G` 以外的副作用。[官方 Concepts](https://github.com/boardgameio/boardgame.io/blob/main/docs/documentation/concepts.md)
- 多人模式中，client 只送出 move／event，由 game master 執行規則並廣播新狀態；client 的 optimistic update 最終仍會被 master 結果覆蓋，因此權威來源只有一個。[官方 Multiplayer 文件](https://github.com/boardgameio/boardgame.io/blob/main/docs/documentation/multiplayer.md)
- `playerView({ G, ctx, playerID })` 會依玩家產生刪除秘密後的狀態；使用秘密資料的 move 可標成 `client: false`，避免在資訊不足的 client 上執行。[官方 Secret State 文件](https://github.com/boardgameio/boardgame.io/blob/main/docs/documentation/secret-state.md)
- turn 內的 stage 可以暫時只允許指定的 active players 行動；官方例子包含目前玩家出牌後，要求其他玩家各自丟一張牌。這些回合外操作仍是個別 move，而不是多個執行緒共同寫入狀態。[官方 Stages 文件](https://github.com/boardgameio/boardgame.io/blob/main/docs/documentation/stages.md)
- 內建 Random API 的目標包含精確 replay；文件指出 PRNG 狀態必須留在 server，且允許設定初始 seed。測試文件則使用固定 seed 得到可預期序列，或用 `MockRandom` 指定特定結果。[官方 Random 文件](https://github.com/boardgameio/boardgame.io/blob/main/docs/documentation/random.md)、[官方 Testing 文件](https://github.com/boardgameio/boardgame.io/blob/main/docs/documentation/testing.md)
- 官方測試建議同時涵蓋獨立 move 單元測試、從特定局面開始的 scenario test、本機多人 client 測試與 UI integration test。[官方 Testing 文件](https://github.com/boardgameio/boardgame.io/blob/main/docs/documentation/testing.md)
- 框架的 log entry 記錄 action、`_stateID`、回合與階段；move 也可把參數標成 `redact`，表示操作紀錄本身可能包含不應公開的資訊。[官方型別定義](https://github.com/boardgameio/boardgame.io/blob/main/src/types.ts)、[官方 Game API](https://github.com/boardgameio/boardgame.io/blob/main/docs/documentation/api/Game.md)

### 對本專案的推論

- 這直接支持「Game Module 是唯一狀態寫入者」。CLI 與 bot 應送 request，不應取得可寫的 state reference；未來即使加入 server，權威規則也不應搬進 transport adapter。
- 對手回合可回應不構成平行狀態修改的理由。Opportunity、效果堆疊與 `PendingChoice` 可以讓 acting player 隨規則切換，每次仍只提交一項輸入。
- `ViewFor(playerID)` 應是 Game Module 的正式輸出能力，而不是 CLI 自行隱藏欄位。bot 測試也只能收到同一份玩家視角。
- `stateRevision` 應附在所有 action／choice submission 上；引擎在提交點重新檢查 revision 與合法性。這是日後加入非同步 bot 或網路 client 時避免 stale action 的最小契約。
- 事件與 replay 要有資訊分級。內部 canonical replay 可以保存重建所需資料；給玩家或觀戰者的版本必須經過與 `ViewFor` 相同概念的 redaction。

## 2. OpenSpiel：Action/State 介面、資訊狀態與 history replay

### 來源事實

- OpenSpiel 把一局表示成狀態樹：節點是 `State`，轉移是玩家 action；chance 也被建模為特殊 player。循序遊戲每個狀態有一個 current player，同時行動遊戲則以 joint action 表示一次轉移。[官方 Concepts](https://github.com/google-deepmind/open_spiel/blob/master/docs/concepts.md)
- 核心介面提供 `CurrentPlayer`、`LegalActions(player)`、`ApplyAction` 與會先做合法性檢查的 `ApplyActionWithLegalityCheck`。結構化 action 的 `ApplyActionStruct` 明確保證先驗證；非法時 state 不變。[官方 `spiel.h`](https://github.com/google-deepmind/open_spiel/blob/master/open_spiel/spiel.h)
- OpenSpiel 明確區分一般整數 action 與 structured action；官方原始碼說明，當 action 空間具有複雜或組合爆炸特性時，攤平成整數不可行，可改用 `ActionStruct`、驗證函式與 sampler。[官方 `spiel.h`](https://github.com/google-deepmind/open_spiel/blob/master/open_spiel/spiel.h)
- 不完全資訊遊戲可依 player 回傳 `InformationStateString/Tensor` 與 `ObservationString/Tensor`。官方定義要求 observation 同時涵蓋該玩家的公開與私有觀察，但資訊量不得超過其 information state。[官方 `spiel.h`](https://github.com/google-deepmind/open_spiel/blob/master/open_spiel/spiel.h)
- `State` 保存 `(player, action)` history。預設 `Serialize` 會從初始狀態寫出 action 序列；官方特別警告，這對內部自行取樣的 stochastic game 不足，因為沒有通用方式設定相同 seed 以重現 chance event。[官方 `spiel.h`](https://github.com/google-deepmind/open_spiel/blob/master/open_spiel/spiel.h)、[官方 API Reference](https://github.com/google-deepmind/open_spiel/blob/master/docs/api_reference.md)
- 對 explicit stochastic game，`ChanceOutcomes` 列出所有 outcome 及機率，之後 `ApplyAction` 是確定性的；sampled stochastic game 則在 `ApplyAction` 內取樣並自行維護 RNG。[官方 `spiel.h`](https://github.com/google-deepmind/open_spiel/blob/master/open_spiel/spiel.h)
- 共用 `RandomSimTest` 會隨機跑完整對局，可在每個 state 執行檢查，預設也測 serialization；`RandomSimTestWithUndo` 會逐步倒回開局並比對歷史上的 state。Kuhn Poker 的官方測試直接重用這些共用測試，另枚舉小型完整 state space。[官方測試介面](https://github.com/google-deepmind/open_spiel/blob/master/open_spiel/tests/basic_tests.h)、[Kuhn Poker 測試](https://github.com/google-deepmind/open_spiel/blob/master/open_spiel/games/kuhn_poker/kuhn_poker_test.cc)

### 對本專案的推論

- `LegalActions + Apply` 是合適的核心形狀，但複雜宣告不應預先展開成所有「卡牌 × 模式 × 目標 × 費用」組合。已決定的逐步 `ChoiceRequest` 相當於保留 structured action 的優點，又更適合 CLI。
- 每個 `PendingChoice` 至少要明確包含 `ActorID`、choice kind、合法 options、是否可取消及其所屬 state revision。CLI 與 bot 不應從 prompt 文字反推語意。
- Opportunity 互動最自然的表示是「下一個決策 state 換另一位 current actor」，不是 goroutine。規則上的同時事件可在一個 `Apply` 內原子套用；私密同時選擇則先收集、不公開，收齊後才一次套用。
- 目前「版本 + seed + 輸入序列」的 replay 決策成立，但有前提：引擎版本必須鎖定 PRNG 演算法、shuffle 實作與每次效果消耗亂數的順序。建議 replay header 明列 `rng_algorithm`；若未來無法維持這項契約，應把每個實際 chance outcome 一併記錄。
- 完整 playthrough fuzz/property test 應像 `RandomSimTest` 一樣在每一步檢查不變量，而不是只看最後贏家。這支持目前對卡牌守恆、行動權限、相同輸入 state hash 的驗證方向。

## 3. XMage：型別化卡牌組合與特殊效果逃生口

### 來源事實

- XMage 的官方 repository 說明其具完整規則執行、超過 31,000 張獨特卡牌，並支援真人、AI、本機 test mode 與多種對局模式。[官方 repository](https://github.com/magefree/mage)
- 簡單卡牌 `LightningBolt` 是一個 Java `CardImpl` 類別，透過共用 `TargetAnyTarget` 與 `DamageTargetEffect(3)` 組成行為；卡面自然語言本身不是執行來源。[官方 `LightningBolt.java`](https://github.com/magefree/mage/blob/master/Mage.Sets/src/mage/cards/l/LightningBolt.java)
- 複雜卡牌 `SelvalaExplorerReturned` 同時使用共用 `DrawCardAllEffect` 與該卡專屬的 effect class，顯示共用效果元件之外仍保留專屬程式碼逃生口。[官方 `SelvalaExplorerReturned.java`](https://github.com/magefree/mage/blob/master/Mage.Sets/src/mage/cards/s/SelvalaExplorerReturned.java)
- `Player` 介面的註解要求新增選擇 dialog 時，同時支援 human、computer、stub、proxy 與 test player；同一介面也包含 `choose`、`sendPlayerAction`、reveal/look 與 state restore 等大量能力。[官方 `Player.java`](https://github.com/magefree/mage/blob/master/Mage/src/main/java/mage/players/Player.java)
- repository 具有按規則機制分類的大量卡牌測試目錄，官方 README 另說明可用預先條件建立特殊 test mode 場景。[官方測試目錄](https://github.com/magefree/mage/tree/master/Mage.Tests/src/test/java/org/mage/test/cards)、[官方 repository](https://github.com/magefree/mage)

### 對本專案的推論

- XMage 提供了與目前 Go 決策最接近的成熟先例：卡牌以型別安全程式碼註冊，常見行為組合共用 effect，罕見卡牌使用專屬 handler。首副固定牌組沒有先建 DSL 的必要。
- 但不應複製 XMage 寬廣的 `Player` 介面。卡牌 handler 若能直接要求 human dialog、讀取完整 player state 或觸碰 adapter，會讓規則、互動與測試耦合。專案目前把所有互動收斂為資料化 `ChoiceRequest`，邊界更符合首版需求。
- 卡牌專屬 handler 仍必須透過受控的 engine operations，例如 `DealDamage`、`DrawCards`、`MoveCard` 與 `CreateChoice`；它不能直接修改 aggregate 欄位。這能保留 XMage 的表達能力而縮小權限面。
- XMage 本次查到的一手入口適合校驗卡牌表示與 scenario test，但沒有足夠證據建立它的穩定 canonical replay 契約，因此 replay 設計不以 XMage 為依據。

## 4. Forge：DSL 的規模效益、啟發式 bot 與真實縱切測試

### 來源事實

- Forge 官方將自己定義為 Magic: The Gathering rules engine，並稱已支援超過 99% 的卡片；它也支援最多八名玩家及 human/AI controller。[官方 Wiki](https://github.com/Card-Forge/forge/wiki/)
- Forge 的卡牌檔是由引擎解析的腳本 API：`A`、`T`、`R`、`S` 分別表示能力、觸發、替代與靜態效果，能力由以 pipe 分隔的 name-value parameters 組成；trigger 會指定監聽事件及要執行的能力。[官方 Card Scripting API](https://github.com/Card-Forge/forge/wiki/Card-scripting-API)、[官方 AbilityFactory](https://github.com/Card-Forge/forge/wiki/AbilityFactory)、[官方 Triggers](https://github.com/Card-Forge/forge/wiki/Triggers)
- Forge 文件也列出仍由 hardcoded plaintext 辨認的機制，表示 DSL 並沒有消除特殊案例或底層程式碼。[官方 Card Scripting API](https://github.com/Card-Forge/forge/wiki/Card-scripting-API)
- Forge AI 使用基本規則與 heuristics；判斷邏輯主要分散於 effect API 與其他遊戲決策，偶爾有單卡 hardcode，但官方文件說通常不健康。它提供無 GUI 的 CLI simulation mode，可批次跑 AI 對局並分析 log。[官方 AI 文件](https://github.com/Card-Forge/forge/wiki/AI)
- Forge 的 network integration tests 啟動真實 TCP server 與 headless AI clients，檢查 crash、desync 與效能；其中最小 vertical slice 使用兩副十張基本地牌讓單局很快抽空結束。[官方 Network Testing 文件](https://github.com/Card-Forge/forge/wiki/Network-Testing)

### 對本專案的推論

- Forge 證明 DSL 在超大型卡池能產生維護效益，但其 parser、參數規格、觸發語法、AI hints 與 hardcoded escape hatches 本身就是一個大型子系統。這不是只有一副固定牌組的首版應先支付的成本。
- Go-first 並不阻礙未來 DSL。現在若把 `Draw`、`Damage`、`SelectTarget`、`Trigger` 等效果元件設計成穩定、可組合且受 engine operation 約束，未來 DSL interpreter 可以呼叫同一批元件。
- 啟發式 bot 是合理首版：它應只消費玩家視角與合法選項，策略可替換，但不需先做 search tree 或 machine learning。應避免把大量單卡 AI 特例放入 card handler；必要的 bot hints 應由策略層解讀。
- Forge 的真實 network vertical slice 支持本專案保留少量 stdin/stdout 完整對局 smoke tests；規則組合的主要覆蓋仍應留在 Game Module scenario tests，而不是全數堆在端到端測試。
- Forge 本次一手資料主要用來校驗卡牌表示、bot 與測試層次，不把它當作本專案隱藏資訊或 canonical replay 的設計依據。

## 跨案例決策校驗

| 本專案決策 | 校驗結果 | 最重要的實作約束 |
| --- | --- | --- |
| 唯一權威的確定性 Game Module | 強支持 | 狀態不可由 CLI、bot、卡牌或 storage adapter 旁路修改；`Apply` 是唯一提交點 |
| 逐步選擇，完成後原子提交 | 強支持 | 宣告草稿不改正式 state；結算中 choice 則成為明確 pending state |
| 單局狀態轉移序列化 | 強支持 | 回合外反應是 actor 交替；同時規則語意不等於程式並行 |
| 玩家限定視角 | 強支持 | `ViewFor(player)` 是核心契約；bot、CLI、未來 client 共用，並以防洩漏測試保護 |
| 卡牌行為首版使用 Go | 對目前範圍強支持 | 共用 typed effects + 少量專屬 handler；禁止 handler 越權改 state 或呼叫 UI |
| replay = 版本 + seed + 輸入 | 有條件支持 | PRNG 演算法與消耗順序必須被版本鎖定；公開 replay 另做 redaction |
| 規則案例先行與三個測試 seam | 強支持 | scenario tests 為主，補固定亂數、完整 playthrough property tests 與少量 CLI smoke tests |

## 建議轉成驗收條件

以下項目不要求新增架構層，只把既有決策收緊成可測的契約：

1. 同一引擎版本、規則版本、卡面版本、PRNG 識別、seed 與輸入序列，必須逐步產生相同的完整 state hash。
2. 每個提交都帶 `ActorID` 與 `ExpectedRevision`；錯誤 actor、過期 revision 或非法 action 必須失敗且 state hash 不變。
3. 對任一玩家與 spectator 產生的 view，不得包含對手手牌 identity、牌庫順序或尚未公開的私密選擇；bot 測試 fixture 也不得繞過此 view。
4. 宣告取消、參數非法或目標在提交前失效時，正式 state 不得殘留部分費用或部分效果。
5. 私密同時選擇在所有人完成前不得互相可見；全部完成後才由一次狀態轉移公開並套用。
6. 每個支援卡牌至少有正常、非法、目標失效及與牌組中另一機制互動的 scenario tests；隨機完整對局在每一步檢查卡牌守恆與 actor 合法性。
7. Canonical replay 可供內部重建，但若輸出給其中一位玩家或 spectator，必須產生相應的 redacted projection，不能只靠 CLI 不顯示秘密欄位。

## 研究邊界

- 本研究沒有評估這些專案的效能、授權相容性或是否適合成為 Go dependency，因為目前目標是校驗設計模式，不是選用框架。
- XMage 與 Forge 是大規模 Magic 引擎；它們證明某些模式能擴張，不代表其 package 數量、介面寬度或 DSL 複雜度適合首版。
- GitHub `main`／`master` 與 Wiki 內容會持續變動。正式採納任何外部行為前，應在 ADR 或測試註記查閱日期或來源 commit；本專案自己的規則行為仍以已鎖定的 Grand Archive 規則 commit 為準。
