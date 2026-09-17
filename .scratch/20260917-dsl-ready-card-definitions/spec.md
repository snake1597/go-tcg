# 可執行卡牌定義與 DSL-ready 能力 Runtime

Status: ready-for-agent

## 問題陳述

Game Module 已具備小型對外 Interface 與確定性的 typed-effect runtime，但卡牌行為尚未由單一、可執行的來源表示。Production registry 記錄 Card Face、Ability Slot、mechanism、operation、dependency 與 ruling metadata；實際的啟動合法性、費用、目標選擇、觸發、靜態效果、替代效果及效果序列，則分散在 Implementation 的多個 Card ID 分支。

這種重複表示只能證明 Supported Ability Slot 的結構有效，無法證明已登錄的行為可以執行，也無法保證 Implementation 與 registry 一致。新增一張卡牌可能需要修改多個 Card ID switch；理解單一卡牌則需要跨越宣告、選擇、觸發收集、特徵求值、替代處理與效果結算等流程。

現有 typed operations 是合適的執行基礎，但目前的資料形狀允許無效欄位組合，也已經包含 Support Set 專用的 operation。若直接把現行 Implementation 改寫成資料 DSL，只會用新語法保留既有分支，無法形成可維護的卡牌編寫模型。

確定性 replay 契約另有一個必須立即處理的完整性缺口：canonical state 宣告了 Cardistry 使用狀態、折扣及數個 runtime 身分計數器，但 state hash 投影沒有填入所有欄位。因此，兩個會產生不同未來行為的狀態可能具有相同 hash。

使用者需要現在就提升 Card Definition 的可讀性與可維護性，同時建立一條直接且可驗證的路徑，讓現有支援卡牌日後能轉成 DSL，而不必再次更換規則 runtime。

## 解決方案

讓每個受支援的 Card Definition 擁有不可變且可執行的 Ability Definitions，並將其編譯成單一、經驗證的中介表示，再交由統一的 Ability 與 Effect Runtime 執行。先以 Go builder 驗證此表示能涵蓋目前 Support Set，再加入版本化 declarative DSL，且 DSL 必須編譯成完全相同的中介表示。每張卡牌完成遷移後立即刪除舊 Card ID 分支，不保留平行相容路徑。

Ability Definition 只描述目前 Support Set 已證明需要的能力：能力類別與作用範圍、啟動時機或觸發條件、宣告選擇、目標 selector、費用、使用限制、效果序列、持續效果輸入、替代行為，以及必要的 ruling 或 content dependency。Selector、reference、value、condition、duration 與 effect 應是可重用的領域原語，而不是以卡牌命名的 operation。

所有 activated 與 triggered ability 都必須進入統一 runtime ADR 所定義的確定性生命週期：合法性與宣告、必要的 Pending Choice、原子費用支付、以 Source Ref 與 Last-Known Information 建立 Ability Instance、放入 Effects Stack、結算、產生 Game Event，最後抵達 Stable Yield Point。Static ability 只產生中央 Derived Characteristics evaluator 的輸入；replacement ability 只產生 Replacement Pipeline 的輸入，不得繞過這些 module。

Runtime 行為使用既有 Game Module Interface，透過 Player View、View Handle submission、replay 與 state hash 測試。唯一新增的測試 seam 是純資料 definition loader/compiler，因為錯誤定義必須在建立 Game 前遭拒絕。內部 interpreter 分支、私有 card handler 與呼叫順序都不成為測試 seam。

Canonical state hash 必須完整涵蓋所有會影響目前或未來可觀察行為的欄位。回歸測試同時證明相同 replay 的確定性，以及任一行為相關 canonical 欄位改變時 hash 必須不同。

## 使用者故事

1. 身為卡牌作者，我希望一個 Card Definition 就包含其可執行的 Ability Definitions，讓我不必在 Game Module 各處搜尋 Card ID 分支也能理解整張卡。
2. 身為卡牌作者，我希望抽牌、傷害、counter、zone movement 與暫時 modifier 等簡單效果由可重用原語組合，讓常見卡牌只需定義資料而不需新增 runtime 程式碼。
3. 身為卡牌作者，我希望目標與卡牌選擇由可重用 selector 表示，讓每種卡牌選擇不必新增專用 operation。
4. 身為卡牌作者，我希望效果數值與條件由經驗證的 value expression 表示，讓動態效果可以依合法的 Derived Characteristics 計算，而不把卡名寫入 interpreter。
5. 身為卡牌作者，我希望 activated ability 一起宣告 timing、cost、target 與 effect sequence，讓合法性與執行內容無法分離漂移。
6. 身為卡牌作者，我希望 triggered ability 宣告其觀察的 Game Event 與條件，讓 trigger discovery 不再增加 Card ID switch。
7. 身為卡牌作者，我希望 static ability 宣告 predicate、layer、modifier 與 duration semantics，讓中央 evaluator 保持 Derived Characteristics 的唯一權威。
8. 身為卡牌作者，我希望 replacement ability 宣告 event filter 與 transformation，讓能力必定參與既有 Replacement Pipeline。
9. 身為卡牌作者，我希望 alternative cost 與 additional cost 使用和一般費用相同的 Declaration Transaction，讓失敗或取消的宣告不改變正式 Game State。
10. 身為卡牌作者，我希望 Pending Choice continuation 能把選取對象綁定至具名 reference，讓後續效果無須保存 Go closure 也能使用選擇結果。
11. 身為卡牌作者，我希望 Ability Slot ID 在 Go 與 DSL 編寫方式間保持穩定，讓 usage tracking、ruling、replay 與診斷都指向相同語意能力。
12. 身為卡牌作者，我希望定義錯誤指出 Card Definition、Card Face、Ability Slot 與失敗欄位，讓編寫錯誤可以直接處理。
13. 身為卡牌作者，我希望未知的 effect kind、selector、expression、zone、duration 與 reference 在載入時被拒絕，讓錯誤內容無法進入 runtime panic。
14. 身為卡牌作者，我希望無效欄位組合與無法解析的 reference 在載入時被拒絕，讓格式錯誤的 tagged data 不會悄悄產生錯誤結算。
15. 身為維護者，我希望 production registry 從可執行定義衍生支援 metadata，讓 Supported 狀態與可執行行為只有一個來源。
16. 身為維護者，我希望保留 mechanism、operation、dependency 與 ruling validation，讓定義變成可執行資料後仍維持 Support Set gate 的強度。
17. 身為維護者，我希望現有領域原語能表達行為時就移除卡牌專用 operation kind，讓 interpreter 複雜度依 mechanism 而非卡牌數量成長。
18. 身為維護者，我希望統一 runtime 保留在 Game Module 內且不建立推測性的 package interface，讓外部 Interface 維持小型且深層。
19. 身為維護者，我希望完成遷移的卡牌立即刪除過時 Card ID 路徑，讓系統沒有長期 compatibility layer 或雙重實作。
20. 身為維護者，我希望以可運作的 vertical slice 遷移，讓每張卡遷移完成後產品仍可遊玩且可測試。
21. 身為維護者，我希望首批 slice 涵蓋簡單序列能力、目標選擇能力、trigger、static ability、replacement，以及複雜費用或 copy 互動，讓全面遷移前先證明表示法涵蓋所有現有能力家族。
22. 身為維護者，我希望在 Go 中的中介表示經過驗證後才定義版本化 DSL schema，讓語法來自已證明的領域需求，而非推測性的語言設計。
23. 身為維護者，我希望 DSL 在啟動或 content loading 階段編譯成不可變定義，讓 resolution loop 不必重複解析編寫資料。
24. 身為維護者，我希望編寫表示不進入可變 Game State，讓 replay 與 runtime identity 保持確定性。
25. 身為維護者，我希望 canonical hash 包含每個會影響行為的 Game State 欄位，讓 replay divergence 在第一個不同步步驟就被偵測。
26. 身為維護者，我希望確定性驗證包含所有 identity counter，讓未來 Object、Ability 與 Continuous Effect identity 不會在相同 hash 背後分歧。
27. 身為玩家，我希望 DSL 編寫的卡牌提供與現有卡牌相同的 Player View 與 View Handle 互動，讓操控端不必知道卡牌的編寫方式。
28. 身為玩家，我希望非法或過期選擇遭拒絕時不改變 Game State、PRNG cursor、Knowledge State、event 或 cost，讓 declarative card 保留權威規則行為。
29. 身為玩家，我希望 effect 依鎖定規則 fizzle、保留已支付費用或繼續結算，而不是採用通用 interpreter 的猜測，讓遷移不改變卡牌行為。
30. 身為 replay 使用者，我希望相同版本、seed 與輸入序列在每一步都產生相同 state hash，讓 Go 與 DSL 定義保留確定性播放。
31. 身為 bot 作者，我希望卡牌評估使用 Player View 與通用 legal-action metadata，而非 Game Module 內的精確卡名，讓新增 DSL 卡牌時無須修改規則引擎 heuristic。
32. 身為 reviewer，我希望每個完成遷移的 Ability Slot 都連結鎖定規則來源與 scenario tests，讓資料轉換仍遵循規則實作的證據標準。
33. 身為 operator，我希望不支援或有歧義的 mechanic 被 Support Set gate 拒絕或回報 Needs Ruling，讓 DSL 不會把未知行為變成猜測行為。
34. 身為 operator，我希望 definition 與 schema version 納入既有 content/version contract，讓 replay 能識別該場 Game 使用的確切可執行卡牌內容。

## 實作決策

- 保留 Game Module 作為 deep module。Caller-facing Interface 仍聚焦於建立 Game、提交 player-scoped input、取得 Player View、產生 replay 與計算 state hash。
- 引入不可變的 executable Card Definitions，其中包含 Card Faces 與 Ability Definitions。Runtime Card Instances、Objects、Stack Items、Ability Instances、Continuous Effect Instances 與 Pending Choices 繼續只存在於 Game State。
- 開發期間採兩階段編寫路徑：先使用 typed Go builders，再加入版本化 declarative DSL。兩者產生相同、經驗證的中介表示並使用相同 runtime；DSL 不直接執行。
- 以中介表示作為語意契約。它只建模目前 Support Set 需要的 mechanism，並透過完成的 vertical slice 擴充。
- 明確建模 activated、triggered、static/continuous 與 replacement 等能力家族。參與合法性與 cost evaluation 的 permission 或 prohibition 也必須有明確定義形式，不得隱藏在 Card ID 檢查中。
- 將 declaration concern 與 resolution effect 分開建模。Timing、mode、selector、target、cost 與 usage restriction 經 Declaration Transaction 驗證並提交後，才建立 Ability Instance 並放入 Effects Stack。
- Continuation state 使用具名 binding 的可序列化 Resolution Frame。任何 definition 或 effect 都不得保存 closure 或指向編寫定義的可變 pointer。
- 以經驗證的 typed variant 取代具有大量 optional field 的 operation bag。每種 effect kind 只接受相關 payload；definition compilation 拒絕缺漏、衝突或未使用欄位。
- 定義符合目前領域詞彙的通用 selector，包括 controller/owner 關係、zone、Card 或 Object kind、type、subtype、element、printed value、Derived Characteristics、排除來源及 visibility-safe choice projection。
- 為 source、Source Ref/LKI、declared target、chosen subject、controller 與目前 event subject 定義 reference。Compiler 必須在 definition 註冊前完成 reference type checking。
- 只定義 Support Set 需要的最小 integer 與 boolean expression，包括 constant、Derived Characteristics query、count、distinct printed-value count、threshold、arithmetic 與 logical predicate。不嵌入通用 scripting language。
- Duration 使用規則概念表示，例如回合結束、指定玩家下一回合開始或結束、來源仍位於必要 zone 期間，以及永久。Duration evaluation 仍由 scheduler/evaluator 負責。
- Zone move、draw、damage、counter、continuous modifier、source cleanup 與 event recording 保留在 Game Module operation 後方。Definition 只描述意圖，不能直接修改 Game State。
- Static definition 提供中央 evaluator 的輸入，並保留 layer、sublayer、timestamp、dependency 與 source-presence semantics；不得永久改寫 printed characteristics。
- Replacement definition 提供 Replacement Pipeline 的輸入，並保留 candidate recomputation、ordering choice、prevention、cause chain 與 continuation behavior。
- 儘可能從 executable definition 衍生 production support registration。Ability Slot 只有在定義成功編譯，且所有 mechanism、operation、dependency 與 ruling reference 都有效時，才能標示為 Supported。
- 通用原語能表達行為時，移除以卡牌命名的 operation kind。若已證明 Support Set 存在無法再分解的 mechanism，新增具有狹窄 typed payload 的穩定 mechanism-level native operation；native code 不得直接取得 Game State mutation 權限。
- 保持確定性順序。Map 或 selector result 在影響 choice、Stack Item、event、ID 或 state hash 前必須正規化排序。
- DSL content 只在 content loading 階段編譯一次。Definition loading failure 會阻止受影響的 production Support Set 啟動，不得 fallback 到舊 Go behavior。
- 透過既有 version contract 管理 DSL schema 與 compiled card content 版本。Schema 變更不提供 backward-compatible reader 或 migration layer；專案更新 pinned content 時直接移除過時格式。
- 使用 vertical slice 遷移。每個 slice 完整取代指定 Ability Slot 的行為並刪除相應舊分支，完成後才開始下一個 slice。
- 先選擇具代表性的能力，而不是一次遷移所有簡單卡牌：簡單序列效果、runtime target selection、triggered behavior、static continuous evaluation、replacement handling，以及複雜 alternative-cost 或 copy behavior。
- 將 bot 的卡名 heuristic 移出 Game Module。若策略需要額外資訊，公開通用且 player-visible 的 legal-action metadata，並由 Bot Controller module 解讀。
- Canonical state projection 必須完整包含所有影響現在或未來行為的欄位，包括 usage tracking、discount、PRNG state、runtime identity counter、scheduler state、choice、effect、event 與 Knowledge State。
- 所有錯誤以 `fmt.Errorf` 包裝，並在每層 compiler 附加 definition context，使錯誤能指出 Card Definition、Card Face、Ability Slot 與語意欄位。
- 遵守 repository formatting rules：YAML 使用展開 mapping；Go struct literal、nested call 與 anonymous function 不使用 inline 格式。

## 測試決策

- 主要 seam 是正式 Game Module Interface。Scenario test 提交 Player View handle，只檢查可觀察的 Player View、replay、error 與 state hash；不呼叫私有 effect executor，也不檢查 Card ID dispatch。
- 唯一新增 seam 是純資料 definition loader/compiler。測試輸入 definition，並斷言得到有效且不可變的中介表示，或具有完整 context 的 validation error。此 seam 的必要性在於 production content 必須於 Game 建立前失敗。
- 既有 action、Cardistry、trigger、characteristics、replacement、knowledge、replay、setup、combat 與 Support Set scenario 是測試先例。將卡牌遷移成 executable definition 時保留其規則斷言。
- 新 runtime 不以舊 runtime 的輸出作為真相來源。Expected behavior 來自 pinned rules snapshot、已解決 ruling、card data 與既有 externally observable scenario outcome。
- 新增 canonical-hash regression tests，分別改變每個 behavior-bearing canonical field 並證明 hash 改變。至少包含 Cardistry usage、Cardistry discount、next-effect identity、next-ability identity 與 next-object identity。
- 保留確定性不變量：相同 engine/rules/card-data/deck/PRNG version、seed 與 accepted input sequence，必須在 replay 每一步產生相同 state hash。
- 為 definition compiler 加入 table tests，涵蓋 duplicate ID、無效 Ability Slot relationship、未知 mechanism 或 operation、未知 reference、錯誤 reference type、不支援 selector、無效 duration、不可能的 effect payload、未使用欄位、禁止的 dependency cycle 與不支援 schema version。
- 每張遷移卡牌必須涵蓋正常結算、非法 activation 或 declaration、取消或失敗的 Declaration Transaction、target invalidation/fizzle、來源離開相關 zone，以及至少一個與其他 Support Set mechanism 的互動。
- Pending Choice 完全透過 Player View 測試：只有 actor 看得到 option、option 使用 View Handle、stale 或 cross-player handle 失敗且不改變狀態、只有宣告允許時才可 pass，且 resolution 必須確定性恢復。
- 透過 Game Module Interface 測試 Declaration Transaction rollback：cost failure、無效最終 target、stale revision 與取消 optional choice 均不得改變 zone、counter、PRNG cursor、event、trigger buffer、Knowledge State 或 state hash。
- 使用 Event Batch 與 Rule Checkpoint semantics 測試 triggered definition，包括 simultaneous trigger、ordering choice、Last-Known Information，以及所有 trigger placement 完成後才能轉移 Opportunity。
- 只透過正式行為可見的 Derived Characteristics，測試 static definition 的適用 layer、timestamp、dependency、dependency loop 與 source-presence predicate。
- 測試 replacement definition 的 applicable-candidate discovery、player ordering、每次 replacement 後的 recomputation、prevention、Pending Choice 後 continuation，以及正確 cause/event history。
- 測試 alternative cost 與 runtime copy 的獨立 Ability Instance identity、正確 Source Ref/LKI、無重複 payment、source cleanup 與 replay stability。
- 加入 property/fuzz coverage，涵蓋 rejected-input state preservation、replay hash determinism、selector ordering、definition compiler 不發生 panic，以及遷移效果中的 card conservation。
- 保留少量 CLI 與 mirror-game tests 作為 end-to-end confidence。它們應偵測 crash、stall、desync 與無法結束，但不重複卡牌規則 scenario。
- 每個完成的 vertical slice 都必須執行完整 repository quality gate。若引入 runtime panic、nondeterministic hash、Support Set failure、mirror-game regression，或仍保留該卡牌過時 Card ID path，遷移即未完成。

## 範圍外

- 將自然語言卡面文字解析成可執行行為。
- 嵌入 Lua、JavaScript、Starlark 或其他通用 scripting runtime。
- 載入不受信任的使用者卡牌定義，或設計第三方內容 security sandbox。
- 在進行中的 Game 內 hot reload definition。
- 支援目前封閉 Support Set 以外的所有 Grand Archive 卡牌。
- 變更鎖定的遊戲規則、解決新的規則歧義，或猜測應維持 Unsupported 或 Needs Ruling 的內容。
- 以第三方引擎取代權威 Game Module、scheduler、Player View model、Effects Stack、中央 characteristics evaluator、Replacement Pipeline 或 replay format。
- 在只有一個 implementation 時新增推測性的 package interface 或 plugin seam。
- 完成取代後仍維護過時 registry shape、Go card handler 或舊 DSL schema version 的相容性。
- 將可變 Game State 持久化至 PostgreSQL 或 Redis。
- 除了通用 Bot Controller strategy 所需 metadata 外，重新設計 CLI presentation。

## 補充說明

- 現有 runtime 是可延伸的基礎，不是拋棄式 Implementation。Draw、damage、counter、zone movement、choice、continuous modifier、Stack placement、Player View 與 replay 應深化為共用 execution engine。
- 目前架構檢視顯示 Ability Instance constructor 是 critical graph hotspot，具有 9 個 direct caller 並影響 16 組 execution flow。因此，Ability Instance 建立流程的修改必須採用小型 vertical slice 並逐一驗證。
- 外部 Game Module Interface 已具備適當深度。主要設計工作是提升內部 locality，並讓 executable Card Definition 成為單一來源，而不是把 module 拆成許多 shallow package。
- 成熟引擎支持此方向：可重用 typed effect 能處理常見卡牌，大型卡池能從 declarative definition 獲益，特殊 mechanic 仍可能需要狹窄的 engine-level primitive。本專案應採用此模式，但不引入這些引擎累積的歷史廣度。
- 撰寫本規格時，Game Module tests 已通過；repository-wide suite 存在既有失敗，原因是 100-seed mirror-game gate 的 seed 1 達到 1,000 action 上限。實作時必須區分此 baseline 與新 regression，且不得弱化或跳過該 gate。
