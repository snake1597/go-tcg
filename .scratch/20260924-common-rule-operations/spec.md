# 共通規則操作與 DSL-ready 效果基礎

Status: ready-for-agent

## Problem Statement

規則維護者新增或修正能力時，同一種遊戲行為可能散落在卡牌專用分支、回合流程、費用支付與效果結算中。抽牌、棄牌、banish、犧牲與選擇已有多條直接修改單局狀態的路徑；部分 typed operation 直接以卡片或 Support Set 機制命名。這讓相同的規則語意難以只修正一次，也讓未來 DSL 難以安全地重用現有 Game Module。

目前能力抽牌與回合抽牌的牌組取牌方向、抽空結果、知識追蹤與事件紀錄不同。未確認規則語意前，直接合併程式可能改變玩家能觀察到的行為。場上 Object、Source Card 與玩家卡牌區域也各有不同生命週期，不能視為同一個 slice 移動。

## Solution

讓卡牌能力、回合流程與 Cost Payment 透過共通的 typed rule operations 提交狀態變化。操作描述規則行為與原因；Game Module 在一處處理重新驗證、原子提交、Replacement Pipeline、Game Event、Knowledge State、觸發、Player View 與 replay。卡片專屬內容改成選擇條件、具名參照、數值表達式及操作序列的組合。

先以現有 Go runtime 與 Support Set 完成可遊玩的逐步遷移。未來 DSL 編譯成同一種已驗證的操作資料；本規格不實作 DSL parser，也不替代更廣的可執行卡牌定義規格。

## User Stories

1. As a 規則維護者, I want to 讓相同的規則行為共用一個操作, so that 修正一次就能涵蓋卡牌與非卡牌流程。
2. As a 卡牌作者, I want to 用抽牌、傷害、選擇與區域移動組合能力, so that 新卡牌不必增加卡名專用的 runtime 分支。
3. As a 玩家, I want to 讓回合抽牌遵守正確的牌組頂與抽空規則, so that 對局結果符合鎖定規則。
4. As a 玩家, I want to 讓能力抽牌與抽至 Memory 保留各自正確的目的區域及事件, so that 可見資訊與卡牌數量正確。
5. As a 玩家, I want to 在牌組不足時得到明確且一致的結果, so that 單局不會因抽牌入口不同而意外分歧。
6. As a 玩家, I want to 讓棄牌從指定合法區域移入 Graveyard, so that 選牌與結算使用同一張卡。
7. As a 玩家, I want to 讓 Memory 的 recollection 維持卡牌順序及追蹤權, so that Player View 能正確顯示回到 Hand 的牌。
8. As a 玩家, I want to 讓 Ally 部署後正確建立 Object 與入場觸發, so that 共通移區不遺失卡牌能力。
9. As a 玩家, I want to 讓 Source Card 結算後移到正確目的地, so that Effects Stack 來源與 Stack Item 的生命週期清楚。
10. As a 玩家, I want to 讓場上 Object 被 banish 時移除 Object 並將卡牌放入 Banishment, so that 區域與公開事件一致。
11. As a 玩家, I want to 讓犧牲與被破壞保留不同的規則原因, so that 觸發及 replacement 依正確事件處理。
12. As a 玩家, I want to 讓犧牲 Weapon 的額外費用在宣告失敗時不留下半完成狀態, so that Cost Payment 原子提交。
13. As a 玩家, I want to 讓 Peppered Chef 只能選擇其他受控 Ally 並可略過, so that 能力選項符合法律條件。
14. As a 玩家, I want to 讓選擇只顯示在指定玩家的 Player View, so that 隱藏資訊不會洩漏。
15. As a 玩家, I want to 讓過期或跨玩家 View Handle 被拒絕且不改變單局狀態, so that 選擇不能污染對局。
16. As a 卡牌作者, I want to 讓選擇結果以具名參照供後續操作使用, so that 多步能力不必覆寫單一 Target。
17. As a 卡牌作者, I want to 用區域、控制者、類型、子類型、元素及費用條件選牌, so that Duchess 和一般選牌使用相同機制。
18. As a 卡牌作者, I want to 用門檻與計數等可序列化數值表達式設定傷害, so that Rouge 和 Straight Flare 不需專屬傷害分支或旗標。
19. As a 玩家, I want to 讓動態傷害於規則指定時點計算並通過 Replacement Pipeline, so that 目標失效及替代效果正確結算。
20. As a 玩家, I want to 讓 Duchess 複製的 Action 保持獨立身分且免付原費用, so that 原牌放逐、目標選擇及複製品清理正確。
21. As a 玩家, I want to 在略過 Duchess 的複製能力後清理暫時複製品, so that Game State 不殘留幽靈卡。
22. As a 玩家, I want to 讓 Reserve Cost、Memory Cost 與替代費用各自依規則原子提交, so that 取消或失敗不會消耗資源。
23. As a 玩家, I want to 讓 Wield 等非卡牌能力支付也使用相同的付款保障, so that 入口不同不造成費用重複或遺失。
24. As a 規則維護者, I want to 讓物件能力的啟動條件、費用與效果有明確宣告, so that 可共用的能力不再散落於 Card ID 分派。
25. As a 規則維護者, I want to 讓未支援的 operation 或參數在 Support Set 驗證時遭拒絕, so that 對局不會在結算時 panic。
26. As a replay 使用者, I want to 讓相同版本、seed 與輸入產生相同 state hash, so that 重整後仍可重播與定位分歧。
27. As a 規則維護者, I want to 每完成一條遷移路徑就刪除舊分支, so that 系統不留下兩套行為來源。
28. As a 未來 DSL 作者, I want to 編譯成和 Go 宣告相同的 typed operations, so that DSL 不需另建規則引擎。

## Implementation Decisions

- 保留 Game Module 為權威 deep module，讓 Player View、玩家輸入、replay 與 state hash 維持對外 Interface。
- 使用現有區域移動與選牌驗證能力作為內部基礎；只在已證明存在不同規則語意時加入新 operation 或 seam。
- 每個 operation 的資料必須可序列化，且只能包含該 kind 的合法參數；不保存 Go closure 或可變定義 pointer。
- Source Card、Object、Card Instance、Stack Item 與 Last-Known Information 使用型別明確的參照。選擇結果可綁定具名參照，Resolution Frame 保存續行位置與結果。
- 選擇條件只加入 Support Set 已需要的欄位。選擇時建立合法選項，提交時再次驗證可見 View Handle 及當前規則條件。
- 抽牌共用牌組頂取牌與卡牌移動語意；回合抽牌、能力抽牌與抽至 Memory 只在已確認的規則差異上給出明確結果。取牌端、抽空勝負、追蹤與 Event Batch 語意須先核對鎖定規則。
- 以共通移動提交流程處理玩家區域與場上 Object 離場；`discard`、`banish`、`sacrifice`、`destroy` 等操作保留不同的規則原因及其事件語意。
- Cost Payment 與效果結算分離。Declaration Transaction 在正式提交前驗證完整付款、目標與來源，失敗或取消不得改變權威 Game State。
- 所有 damage 走既有 Replacement Pipeline。數值表達式只涵蓋現有卡片需要的常數、計數、門檻與簡單算術，並明確指定求值時點。
- `copy_action` 管理原牌移動、免付費建立 Ability Instance、來源身分、目標選擇、Pass 與 runtime copy 清理。
- 物件能力中可組合的條件、費用與效果改用共通表示；Combat、Level Up 與 Materialization 等獨立規則生命週期保留專用入口。
- 各路徑遷移完成後立即刪除被取代的卡名專用 operation、旗標與函式，不保留相容層。
- 此工作深化 typed operations，並作為既有「可執行卡牌定義與 DSL-ready runtime」規格的前置能力；不重複規定完整卡牌定義編譯與 DSL schema。

## Testing Decisions

- 主要測試 seam 為正式 Game Module Interface：使用 Player View 取得 View Handle，提交玩家行動與選擇，再檢查可見結果、遊戲事件、replay 與 state hash。
- 純資料的操作定義驗證可設窄測試 seam，以保證未知 kind、無效參數、錯誤參照及不支援選擇條件在開局前被拒絕。
- 好的測試斷言玩家可觀察的規則結果、拒絕輸入後的狀態不變性與 replay 確定性；不把私有 helper、switch 分支或呼叫順序作為規格。
- 既有 Ability Choice、Cardistry、Action、Combat、Regalia、Replacement、Materialization、Zone Movement 與 Standard Setup 情境測試是先例；受影響卡牌保留其規則情境覆蓋。
- 抽牌情境須覆蓋牌組頂順序、牌組不足、回合抽牌勝負、能力抽牌、抽至 Memory、Knowledge State 與事件批次。
- 移動情境須覆蓋 Hand、Memory、Graveyard、Banishment、Effects Stack Source Card 及 Field Object，並檢查 owner、LKI、入場／離場觸發與事件原因。
- 付款情境須覆蓋 Reserve Cost、Memory Cost、替代費用與 Wield，驗證無效或取消的 Declaration Transaction 不消耗卡牌、PRNG 或事件。
- 選擇、傷害及複製情境須覆蓋目標失效、Pass、Replacement Pending Choice 後續行、隱藏資訊與暫時複製品清理。
- 每個完成的遷移 slice 均須保持 Standard Game 可遊玩，並通過既有 replay 與 Support Set 驗證。

## Out of Scope

- 自然語言卡面解析、完整 DSL parser、第三方腳本 runtime 與不受信任內容沙箱。
- 支援封閉 Support Set 以外的全部卡池，或為尚未出現的機制建立推測性表達式與設定。
- 改變已鎖定的遊戲規則；有歧義時記錄 Needs Ruling，而非由重構自行決定。
- 以新模組取代 Game Module、Scheduler、Replacement Pipeline、中央 Derived Characteristics evaluator 或 replay 格式。
- 為被取代的卡片專用分支、舊 operation 格式或新 DSL 格式維持相容層。

## Further Notes

- 此規格聚焦於規則操作的共通語意；較廣的 executable Card Definition 與 DSL 編寫路徑已有獨立規格。兩者實作時必須共用同一種 typed operation contract。
- 取牌方向與抽空差異目前是待核對的現況，不是希望長期保留的規則選項。
- 場上 Object 的犧牲、被破壞與 banish 共用內部離場流程時，仍需保留各自的 Game Event 與觸發原因。
