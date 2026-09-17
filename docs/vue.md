# Vue 對局介面主規格

## 狀態與責任

本文件是 Grand Archive 玩家對局介面的產品與互動主規格。UI 設計已收斂、尚未實作。Vue 位於獨立 repository；本 repository 的 Go 服務負責權威遊戲狀態、玩家視角、ConnectRPC HTTP API、卡牌 catalog 與圖片資產。

本文件說明玩家看見什麼、如何操作與 Vue 的責任邊界。PlayerView、區域可見性、Declaration Draft、事件及傳輸的精確合約由文末專項規格負責。Replay 與 Bot 策略訓練延後；首版不預留未定介面。人類即時介入 Bot 行動的方案已取消。

## 產品目標與範圍

首版提供可在本機完成一場「人類對 Bot」練習對局的桌面網頁介面。Go Game Module 是唯一規則權威；Vue 不重算合法性、隱藏資訊、觸發順序或勝負，也不樂觀修改正式遊戲狀態。

- `/`：建立固定設定的人類對 Bot 練習對局。
- `/games/:gameId`：持有效 PlayerSession 進入對局。
- 不包含帳號、房間、邀請、配對、牌組編輯、Bot 策略選擇、Replay 或行動版。

## 版面區塊

介面以 1920×1080 為基準、最低 1280×720，採 16:9 桌面布局。對手在上、玩家在下，雙方卡牌皆正向顯示。`background.png` 只作布局與密度參考。

1. 左上 Card Inspector：完整卡牌資訊。
2. 對手棋盤：Champion／Lineage、Field、手牌張數與公開區域。
3. 中央狀態帶：Turn Player、Opportunity Holder、Decision Player、Phase、Combat 與 Intent。
4. Effects Stack：待處理項目、來源卡及順序。
5. 玩家棋盤：Champion／Lineage、Field、Hand、Memory 與其他區域。
6. 事件紀錄：由舊至新顯示玩家可見 EventBatch，可向前載入。
7. 動作區：只呈現引擎提供的 Pass、Next Phase、End Turn 或其他合法選擇。

## 卡牌與區域互動

### Card Inspector

- hover 與 keyboard focus 暫時顯示卡牌；點擊鎖定 Card Inspector。
- 點擊卡牌不直接出牌或提交指令。
- 玩家要從哪個區域出牌，必須先開啟該實際區域，再選擇卡牌。
- 不可操作的卡仍可檢視；點擊後顯示 Go 提供的結構化不可用原因。

### Field 與方向

- Field 卡只露出卡圖與名稱，維持穩定排序，不支援拖曳或手動重排。
- Awake 正向；Rested 向左旋轉；Banishment 內卡牌向右旋轉。
- 選擇 Field 物件使用點擊與非純色彩提示。

### 區域與資訊公開

- Hand：擁有者看卡牌；對手只見張數。
- Main Deck：一般只見張數；受控查看由 PlayerView 明確授權。
- Material Deck：依規則與玩家權限顯示卡牌或張數。
- Memory：隨機支付只顯示剩餘張數；banish 後才由事件揭露。
- Graveyard、Banishment：公開且可瀏覽。
- Champion／Lineage：固定置於各玩家棋盤左側。

純瀏覽視窗隨新 revision 即時同步。Declaration Draft 的基礎 revision 若失效，Draft 立即作廢；PendingChoice 則在 ViewHandle 有效時持續存在。Vue 不自行推斷 owner、controller、face、可見範圍或隱藏卡身分。

### 區域視窗與 Card Inspector 生命週期

- 純瀏覽視窗直接讀取最新 PlayerView；新 revision 更新卡牌、群組、張數與可用性，但不自動關閉。
- 更新後保持捲動錨點；區域變空時顯示 empty state，仍由玩家以 `Close` 關閉。
- 被鎖定卡牌若仍可追蹤，Inspector 跟隨至新區域並更新 CardRef；handle 撤銷時清除內容並顯示「已無法追蹤」。
- snapshot 更新不得任意清除仍有效的 PendingChoice；choice handle 消失、被取代或完成後才關閉。
- 純瀏覽、Draft 與 PendingChoice 可以並存，但提交能力只來自目前 Draft step 或 PendingChoice。

## Declaration 與 Pay Cost

Declaration Draft 流程固定為：Parameters → Modes → Targets → Pay Cost → Confirm。Draft 由 Go 應用層保存並綁定 PlayerSession、game、基礎 revision 與 ViewHandle；Confirm 前不得修改正式 Game State。

- `Done`：完成目前多選步驟。
- `Confirm`：確認整個 Declaration，提交原子化正式輸入。
- `Cancel`：放棄 Draft。
- `Close`：關閉純瀏覽視窗。

Pay Cost 顯示所有可由玩家決定的支付來源，包括 Hand、Memory，以及 Field 上 `Fractal of Sparks` 類似能力的卡牌。玩家選擇所有非隨機來源；引擎驗證組合並安排效果與觸發順序。隨機 Memory 不預先揭露卡牌，只顯示數量。

## 對局狀態與事件

Turn Player、Opportunity Holder、Decision Player 必須分開顯示。Effects Stack 的每個項目有獨立身分、來源卡與順序；來源卡離場後仍可辨識。Combat／Intent 顯示攻擊者、防禦者、目標、wielded Weapon 與引擎提供的狀態。

事件依 EventBatch 分組、由舊至新顯示，使用 cursor 載入較舊歷史。每批只含該 PlayerSession 可見資訊。動畫可略過，但事件不得遺失；事件紀錄不包含聊天。

## 權威狀態、動畫與輸入鎖

`WatchGame` 首次及每次 revision 都傳送完整 PlayerView snapshot。Vue 立即保存最新權威狀態；呈現狀態可為動畫短暫落後。落後期間鎖定正式輸入，但 Card Inspector、純瀏覽與事件紀錄仍可使用。動畫佇列有界，必要時跳至最新狀態，並尊重 `prefers-reduced-motion`。

Command 成功只回傳 `accepted_revision`。Vue 必須等 `WatchGame` 到達該 revision 或更高後才解除等待，不得預測正式局面。

## 連線與 Session

- Vue 使用 Protobuf 產生的 TypeScript client，透過 Connect native protocol 連接 Go。
- PlayerSession 是 seat-scoped opaque bearer token；request 不可自行聲明 player。
- token 只回傳一次，存於 `sessionStorage`，不得放入 URL 或 `localStorage`。
- 同一 PlayerSession 只允許一個活動分頁。新 `WatchGame` 接管後，舊分頁收到 `session_replaced` 並停止 command 與重連。
- 串流每 15 秒 heartbeat。暫時錯誤採帶 jitter 的 0.5、1、2、4、8 秒退避，最高 10 秒；認證、版本與 NotFound 錯誤不重試。
- 重連時顯示覆蓋層、清除 Draft、取得最新 snapshot，再從既有 cursor 補事件。
- subscriber 只保留最新待送 snapshot；慢 client 可跳 revision，但不得阻塞 Game actor。

本機只允許明確設定的 Vue localhost origin 連接 loopback Go API，不使用 wildcard origin 或 cookie。Production TLS、正式網域及跨站政策延後。

## Card Catalog 與圖片

Go 提供版本化 Card Catalog 與普通 HTTP static image assets。PlayerView 用穩定 CardRef 連結 catalog，不重複傳完整卡面，也不只用卡名識別。

資料建置會選擇 canonical edition、產出本機圖片並驗證 digest；缺圖或 digest 不符使建置失敗。瀏覽器暫時載圖失敗顯示 placeholder 與重試，不阻擋棋盤。catalog 版本不符時阻擋正式 command。

## Vue 架構

技術基線為 Vue 3、Vite、TypeScript strict、Composition API、`script setup`、Vue Router、Pinia、vue-i18n、`@connectrpc/connect` 與 `@connectrpc/connect-web`。不使用 Nuxt、SSR、Connect Query 或大型 UI framework。

Store 責任：

- `session`：PlayerSession、game id、connection handle、接管狀態。
- `catalog`：card data version、CardRef 查找、圖片狀態。
- `gameView`：最新權威 PlayerView、accepted revision、串流生命週期。
- `presentation`：呈現 snapshot、EventBatch queue、輸入鎖。
- `declaration`：Draft step、選擇、失效、局部錯誤。
- `ui`：Card Inspector、區域視窗、focus recovery、偏好。

只有 API adapter 可呼叫 RPC；component 不持有 RPC client 或規則邏輯。文案使用 i18n key；首版介面為繁體中文，官方術語與卡名可保留英文。介面語言與 Card Catalog 語言分開處理。

其他架構限制：

- transport、base URL、bearer interceptor、Connect error 轉換與 stream 生命週期集中在 API adapter。
- store 保存可序列化的應用狀態；DOM reference、focus trap 與短暫 hover 留在 component／composable。
- component 只 emit 使用者意圖，不自行產生 action、choice、draft 或 payment handle。
- `gameView` 只能由 WatchGame 更新；`presentation` 不得回寫權威狀態。
- `declaration` 不得把 Draft step 混入 PlayerView，也不得在 reload／reconnect 後恢復舊選擇。
- 不建立第二套 server-state cache 或以 component local copy 保存正式棋盤。

## Loading 與錯誤

狀態分為頁面 loading、阻擋錯誤、對局覆蓋層、局部錯誤四級。局部 Draft 或圖片錯誤不得清空仍有效的棋盤。

- 頁面 loading：等待 session、相容 catalog 與首份 snapshot；三者完成前不顯示可操作空棋盤。
- 阻擋錯誤：session 無效、game 不存在、API／catalog 版本不相容或建立失敗，提供明確離開或重試入口。
- 對局覆蓋層：重連、等待 accepted revision、動畫落後或 session 被接管；保留最後畫面但禁止正式輸入。
- 局部錯誤：圖片、單一區域或 Draft step 失敗；只處理受影響範圍。
- 合法的空 Hand、空 Stack、空區域或無 Legal Action 是 normal empty state，不當成錯誤。
- Go reason code 由 i18n key 顯示；未知 code 顯示安全通用訊息與 correlation id，不展示 stack trace。

## 無障礙

鍵盤必須能到達所有卡牌、區域、對話框與動作。Dialog 開啟時移入 focus，關閉後回到觸發元素。選取、Rested、可用性與錯誤不得只靠顏色。重要 revision、Decision Player 變化與錯誤以適度 ARIA live region 宣告。

- `Enter`／`Space` 執行聚焦元素目前合法的主要操作，不得繞過 availability 或輸入鎖。
- `Escape` 只關閉可安全關閉的純瀏覽 dialog，不自動 Confirm、取消 Draft 或提交 PendingChoice。
- Card Inspector 鎖定／解除具有可感知控制與 `aria-pressed`。
- modal focus 被限制在最上層；觸發元素消失時，關閉後回到對應區域控制。
- snapshot 更新不得把 focus 重設至頁面起點；focused card 消失時移至最接近的有效控制。
- actionable、selected、unavailable、hidden、Turn、Opportunity、Decision 與方向狀態都有文字、圖示或形狀，不只使用顏色或旋轉。
- 一般 snapshot 與每個動畫 frame 不逐項朗讀；只宣告重要角色變化、連線狀態、錯誤與結果。
- 優先使用原生 button、dialog、list 與 option 語意，不以只有 click handler 的元素重造控制項。

## 瀏覽器、效能與驗收

- 支援最新 Chrome、Edge、Firefox、Safari；Playwright 覆蓋 Chromium、Firefox、WebKit。
- catalog 與首份 snapshot 完成後 2 秒內呈現可理解棋盤。
- 本機操作回饋目標 100ms 內；一般 store／layout 工作 50ms 內；動畫目標 60fps。
- 卡圖 lazy-load；長事件與大型區域清單按需要虛擬化。
- 低於 1280×720 顯示不支援提示，不把完整棋盤壓成不可操作版面。
- 1280×720 至 1920×1080 間維持棋盤比例；Field 到達 compact tile 最小寬度後使用水平捲動。
- 效能不足可減少陰影、粒子與過場，但不能移除狀態文字、focus indicator、不可用原因或輸入鎖。
- 效能基準使用最大正常 Hand、Field、公開區域、Stack、event log 與 modal，不只測空棋盤。

首版驗收須以 built Vue 對真實本機 Go server 完成：建立對局、正確玩家視角、區域瀏覽、Card Inspector、完整 Declaration／Pay Cost、Combat／Stack／事件顯示、動畫輸入鎖、斷線重連、單分頁接管、鍵盤操作與 reduced motion。

### 測試層級

- Vitest：Pinia store、API adapter、accepted revision、重連、接管、Draft invalidation、presentation queue 與 i18n error mapping。
- Vue Test Utils：Inspector、zone dialog、Declaration step、focus trap／restore、ARIA state、empty／error state 與 reduced motion。
- Go contract：generated client 經 `httptest`、真實 Connect handler、應用層與 Game Module；不 mock legality 或隱藏資訊。
- Playwright：built Vue 對真實 Go server，覆蓋 CreatePracticeGame 至 game finished／NeedsRuling，並驗證 response 與 DOM 都沒有對手私密資訊。
- Release gate：typecheck、lint、production build、三個 browser engine、Go unit／rules／contract／race 與既有 replay verification 全部通過；不得用無限 retry 掩蓋 flaky test。

## 相關規格

| 主題 | 規格 |
| --- | --- |
| PlayerView、各區域、Catalog、Stack 與 Combat | [PlayerView 與卡牌呈現合約](./specs/player-view-contract.md) |
| API、Game Host、session、stream 與事件 | [ConnectRPC 玩家 API](./specs/connectrpc-player-api.md) |
| Declaration／Pay Cost | [Declaration Draft](./specs/declaration-draft-and-cost-payment.md) |
| 延後範圍 | [Replay 與 Bot 待續議題](./deferred-replay-and-bot.md) |
