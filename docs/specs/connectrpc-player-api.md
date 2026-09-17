# ConnectRPC 玩家 API 邊界規格

## 狀態

已確認並可進入實作。TypeScript client 的產生與交付由使用者處理；Production 部署與 Replay 權限仍延後。

## 背景

Vue 對局介面預計是獨立專案。`go-tcg` 不包含 Vue source、Node build 或前端部署；它提供可讓瀏覽器直接透過 HTTP 呼叫的 ConnectRPC API。

本規格落實 [ADR 0018](../adr/0018-expose-connectrpc-http-api-to-separate-vue-client.md)，並沿用 [ADR 0006](../adr/0006-expose-player-scoped-views-to-controllers.md) 的玩家視角與隱藏資訊邊界。

PlayerView 的內容、區域、Stack 與 Combat 合約統一由 [PlayerView 與卡牌呈現合約](./player-view-contract.md)負責；Declaration 的隔離交易由[Declaration Draft 與費用支付規格](./declaration-draft-and-cost-payment.md)負責。本文件只負責 Game Host、PlayerSession、RPC、stream、事件 cursor 與本機 HTTP 邊界。

## 責任邊界

### `go-tcg` repository

- 擁有權威 Protobuf schema 與 API 版本。
- 產生 Go message、service interface 與 ConnectRPC handler。
- 以 ConnectRPC HTTP endpoint 提供 unary 與 server-streaming RPC。
- 驗證身分、玩家座位、game、revision、action／choice／draft handle。
- 只回傳呼叫者依法可見的 `PlayerView` 與結構化錯誤。
- 不包含 Vue component、CSS、Node package 或前端靜態資源。

### Game Module

- 唯一負責規則合法性、狀態轉移、觸發順序、隱藏資訊、PlayerView、deterministic PRNG、revision、ViewHandle、EventBatch 與勝負。
- 不負責 HTTP、token、CORS、subscriber queue、Declaration Draft 保存或圖片傳輸。

### 遊戲應用層

- 建立與定位 practice game。
- 每場 game 使用一條序列化 mailbox，依序處理 human、Bot 與內部 scheduler 輸入。
- 建立、雜湊保存、驗證及撤銷 PlayerSession。
- 建立 connection handle，實施單一活動分頁接管。
- 保存 session-scoped Declaration Draft，並在 revision、handle、重連或接管時依合約清除。
- 發布最新 PlayerView 且不讓慢 subscriber 阻塞 Game actor。
- 保存玩家可見 EventBatch 的 cursor 歷史。
- 將 typed domain 結果轉換為穩定 reason code 與 transport error。

Bot 只使用自己的 PlayerView／Legal Actions，並經相同 mailbox 提交；策略訓練不屬於本規格。

### 獨立 Vue repository

- 使用 `go-tcg` 權威 schema 產生的 TypeScript client；如何取得 schema、執行 code generation、保存產物及鎖定版本由使用者處理。
- 直接透過 HTTP 連線至 Go ConnectRPC endpoint。
- 負責畫面狀態、Card Inspector、彈窗、動畫與 i18n。
- 不複製規則判定、隱藏資訊過濾或合法行動生成。
- 不依賴 Go internal package、`GameState` JSON 或手寫 HTTP response shape。

## Vue Client

Vue 使用下列邊界：

- Protobuf message 與 service descriptor 來自 `go-tcg` schema 的產生結果。
- runtime 使用 `@bufbuild/protobuf`、`@connectrpc/connect` 與 `@connectrpc/connect-web`。
- 以 `createConnectTransport` 建立整個應用共用的 transport，使用 Connect protocol；不選用低階 `createGrpcWebTransport`。
- base URL、認證 interceptor 與共通錯誤轉換集中在 API adapter，不散落在 component。
- component 不直接建立 client；由 game API service／store 封裝 unary command 與 server stream，再投影成 Vue 狀態。
- 不為 unary RPC 額外加入 Connect Query；遊戲狀態由單一串流與明確 command result 管理，避免兩套 cache 成為不同真相來源。

`card_data_version` 是對局釘選的卡牌內容版本，不等同 Protobuf API／SDK 版本；Vue 必須分別檢查兩者。

## Schema 提供

- `.proto` 只存在於 `go-tcg` 的權威 schema 目錄。
- `go-tcg` 不負責發布 BSR module、TypeScript npm SDK 或 Vue 產生碼。
- 使用者自行決定 Vue 如何取得 schema、產生 client 及保存產物。
- Go 端產生設定與 plugin 版本固定在 repository；不得依開發者機器上的未鎖版 plugin 隱式產生。
- 不相容 schema 變更直接更新 API contract 與消費者，不在 server 保留舊欄位、舊 endpoint 或轉接 fallback。

## Transport Adapter

ConnectRPC handler 是 Game Module 外的 adapter：

1. 驗證 transport 層身分與 request 格式。
2. 將 Protobuf request 轉成應用層 command／query。
3. 經由應用層序列化同一局遊戲的輸入。
4. 取得指定玩家的 `PlayerView`。
5. 轉為 Protobuf response 或穩定 Connect error。

Game Module 不 import ConnectRPC 或 Protobuf transport package。API message 不直接序列化內部 `GameState`，也不暴露 runtime ID。

## 單一路徑原則

- 遠端 Vue 操作只走 ConnectRPC HTTP API。
- 不為相同能力維護 REST fallback、獨立 WebSocket protocol 或第二套 JSON DTO。
- CLI 與 bot 可以使用程序內 adapter，但必須共用同一應用層命令、合法性檢查與 `PlayerView` 語意。
- schema 發生不相容變更時直接同步更新生產者與消費者；不保留舊 endpoint 或舊欄位相容層。

## `WatchGame` 狀態串流

Vue 以 server-streaming `WatchGame` 取得指定玩家的即時狀態：

- 建立連線後第一筆訊息必須是目前最新、完整且僅屬於該玩家視角的 `PlayerView`。
- 每個正式 state revision 產生新的完整 `PlayerView` snapshot，不傳送 field-level delta／patch。
- 每筆訊息攜帶單調遞增的 revision；同一 revision 的內容在相同玩家視角下必須具確定性。
- Vue 以最新完整 snapshot 取代舊狀態；不得依賴先前 snapshot 才能解讀目前狀態。
- 卡牌 catalog、圖片與完整規則文字不放入 snapshot，只保留 `CardRef` 與對局狀態，避免重複傳送靜態資料。
- Vue 可以比較相鄰 snapshot 輔助畫面轉場，但遊戲語意與合法性只能來自最新 snapshot 與 EventBatch，不由 diff 推導。
- 重新連線不補送 PlayerView delta；server 直接先送最新完整 snapshot。
- revision 單調遞增但不保證對單一 subscriber 連續；完整 snapshot 必須能跨 revision 跳號直接取代舊 view。

完整 snapshot 優先於 delta，因為一次行動可能同時改變區域、handle、可見性、合法行動及 Opportunity。這也確保洗牌或進入隱藏區域後，Vue 不會因漏套刪除 patch 而保留已失效資訊。

## Command 與 Snapshot 一致性

正式遊戲狀態只以 `WatchGame` 的完整 PlayerView snapshot 為 Vue 的權威來源。Command response 不回傳另一份 PlayerView：

- `SubmitAction`、`SubmitChoice` 與 `ConfirmDeclaration` 成功時回傳 `accepted_revision`。
- `accepted_revision` 表示該 command 已原子提交後所產生的正式 state revision，不表示 Vue 已收到畫面更新。
- Vue 送出後進入 processing 狀態，停用同一意圖的重複提交，但不樂觀修改 Field、Memory、Life、Stack 或 Legal Actions。
- Vue 收到 `WatchGame.revision >= accepted_revision` 的完整 snapshot 後，才套用正式畫面並結束 processing 狀態。
- Command 被拒絕時，RPC 立即回傳結構化 Connect error／error detail；失敗不得增加正式 revision。
- `BeginDeclaration` 與 `UpdateDeclaration` 只更新隔離的 Declaration Draft，不修改正式 Game State，因此直接回傳最新 Draft step，不回傳 `accepted_revision`。
- `CancelDeclaration` 成功丟棄 draft 且不修改正式 Game State，回傳明確的取消結果；Vue 清除 Wizard，不等待新 snapshot。

若 snapshot 先於 unary response 抵達，Vue 可以保存最新 revision；收到 `accepted_revision` 後，只要目前 revision 已達到或超過該值，便立即結束 processing，不等待重複 snapshot。

## Snapshot 背壓

慢速或暫停中的瀏覽器不得阻塞 Game actor、其他 subscriber 或正式狀態推進：

- 每個 `WatchGame` subscriber 最多保留一份尚未開始傳送的 pending snapshot。
- 新 snapshot 到達時，原子取代尚未傳送的舊 pending snapshot，只保留目前最新 revision。
- 已開始寫入 HTTP stream 的訊息不在中途替換；寫入完成後再取得當時最新 pending snapshot。
- 發布 PlayerView 給 subscriber 的路徑必須 non-blocking，不得讓網路背壓進入單局序列化 command loop。
- Vue 必須接受例如 `42 → 45 → 48` 的 revision 跳號，直接以 48 完整取代目前 view。
- 若 processing 中的 `accepted_revision` 是 43，而下一份 snapshot 是 45，視為該 command 已反映完成。
- 被略過 revision 的可見 EventBatch 不得遺失；Vue 從最後成功保存的 event cursor 呼叫 `ListVisibleEventBatches` 補取。
- 中間動畫可以略過，但 event log、最終狀態與可見資訊不得缺失。

## Heartbeat 與重連

`WatchGame` 必須讓 proxy、server 與 Vue 能在無遊戲更新時辨識連線仍存活：

- 若 15 秒內沒有送出 snapshot，server 傳送一筆輕量 heartbeat。
- Heartbeat 只包含目前 revision 與 server time，不增加 revision，也不構成 PlayerView 或 EventBatch。
- Vue 對網路中斷、`Unavailable` 與 `DeadlineExceeded` 自動重連。
- 重連採帶 jitter 的指數退避，基準約為 `0.5s → 1s → 2s → 4s → 8s`，上限 10 秒；持續至恢復或玩家主動離開。
- 收到重新連線後的第一份有效 snapshot，便重設退避狀態。
- `Unauthenticated`、`PermissionDenied`、`NotFound`、API contract 不相容及 `card_data_version` 不相容不得無限自動重試，必須顯示對應錯誤。
- 重連期間保留並淡化最後一份畫面，顯示阻擋操作的重新連線遮罩，不允許提交 command。
- 連線中斷立即清除 Declaration Draft；重連後不得自動建立、更新、確認或重送 draft。
- 重連成功時 server 直接傳送最新完整 snapshot；Vue 使用最後持久保存的 event cursor 補取遺漏的可見 EventBatch。
- 部署環境的 proxy idle timeout 必須大於 heartbeat 間隔，且不得緩衝 server-streaming response。

## 玩家座位認證

遠端玩家以單局、單座位的 opaque bearer token 建立 `PlayerSession`：

- 建立或加入遊戲時，server 分別為兩個玩家座位核發高熵 token。
- Vue 以 `Authorization: Bearer <token>` 呼叫 ConnectRPC；token 不得放入 URL、game ID 或任何可被一般 access log 記錄的 query parameter。
- server 由 token 推導唯一的 game 與 seat。Game Module 只接收已驗證的 actor，不處理 HTTP credential。
- request 不接受 client 自行宣告 `player_one`／`player_two` 以決定 PlayerView 或操作權。
- URL 或 request 中的 game ID 只作資源定位，不能授予查看、事件、command 或 draft 權限。
- `WatchGame`、`ListVisibleEventBatches`、所有正式 command 與 Declaration Draft RPC 必須使用同一個已驗證座位身分。
- 重連可以沿用仍有效的相同 token；token 失效時不得自動改用其他 seat 或匿名視角。
- 對手 token、跨 game handle、竄改 game ID 或不屬於該座位的 action／choice／draft handle 一律拒絕，不得洩漏另一座位的 PlayerView。
- 對未獲授權的 game，外部錯誤不得用差異化訊息透露該 game 是否存在。
- 首版不把帳號系統放入 Game Module；未來帳號／大廳服務只負責核發或交換 `PlayerSession` token。

`PlayerSession` 是 transport／應用層對單局座位的授權，不是領域中的 Player，也不是 Card、Object 或 ViewHandle，因此不加入遊戲領域 glossary。

### Token 交付與瀏覽器保存

- `CreateGame`／`JoinGame` 成功時，原始 token 只在 response 回傳一次；server 只保存可驗證的 token hash。
- Vue 以 `game_id` 為 key 將 token 存入 `sessionStorage`，讓同一瀏覽器 session 重新整理後可以恢復座位。
- token 不保存於 `localStorage`、URL、route parameter、query string 或可被一般 log 記錄的欄位。
- API adapter 從 `sessionStorage` 取得 token，透過共用 Connect interceptor 加入 `Authorization` header；component 不直接讀寫 token。
- 關閉分頁或瀏覽器 session 後，不保證能用原 token 恢復。未來若需要跨裝置或跨 browser session 恢復，由帳號／大廳服務驗證後交換新 token。
- token 洩漏時可撤銷單一 PlayerSession，不影響同局另一座位。
- 遊戲結束後 PlayerSession 是否繼續授權 Replay，留待 Replay 權限規格決定；目前不得預設永久有效或立即撤銷。

## 本機 CORS 範圍

首版只保證獨立 Vue dev server 與本機 Go ConnectRPC server 能直接互通，不在本規格提前決定 production domain 或正式 proxy topology：

- Go server 預設只監聽 loopback address，不為本機開發綁定所有網路介面。
- 本機設定明確允許 Vue dev server origin；預設涵蓋 `http://localhost:5173`，需要使用 `127.0.0.1` 或不同 port 時由開發設定明確加入。
- 不使用 `Access-Control-Allow-Origin: *`，也不以任意 localhost port pattern 放寬 origin。
- 認證仍使用 `Authorization: Bearer`；本機模式不繞過 PlayerSession 驗證。
- 不啟用 cookie credentials。
- CORS middleware 必須允許 Connect protocol 所需 methods／headers／exposed headers，並額外允許 `Authorization`。
- CLI、bot 與沒有 `Origin` header 的 server-side client 不受瀏覽器 CORS 判斷，但仍受 API 認證與授權限制。
- production origin allowlist、TLS termination、reverse proxy 與正式 CORS policy 延後至部署規格決定，不從本機設定推導。

## 首版本機入口

首版 Vue 只提供能完成本機人類對 bot 練習對局的最小產品入口：

- `/` 是本機啟動頁，只提供「開始練習對局」及必要的建立中／失敗狀態。
- `CreatePracticeGame` 使用目前支援的固定牌組建立一局人類對 bot 單局，不要求使用者選牌組或 bot 策略。
- 建立成功回傳 `game_id`、人類座位的 PlayerSession token 與該局釘選的 `card_data_version`。
- Vue 依既定 token 規則保存 session，然後導向 `/games/:gameId`。
- `/games/:gameId` 是完整對局介面；URL 中的 game ID 不授予座位或觀看權。
- 直接進入 game route 但 sessionStorage 沒有該 game 的 token 時，顯示無法存取，不嘗試猜測座位或建立匿名 PlayerView。
- 首版不包含帳號登入、房間、邀請、真人配對、牌組編輯、牌組選擇或 bot 策略選擇。
- 未來大廳／帳號服務可以在 Game Module 外建立單局與核發 PlayerSession，不改變玩家視角及正式 command 邊界。

## Card Catalog 與圖片

- Vue 使用 `GetCardCatalog(card_data_version)` 取得指定版本、精簡且不可變的 presentation catalog。
- Go server 只提供 repository 已釘選並通過建置驗證的版本，不執行 runtime 外部下載或 latest-version fallback。
- `GetCardCatalog` 回傳 Card ID、CardFace ID、顯示文字、靜態 characteristics 與 canonical image 相對 URL；不回傳完整原始 edition／市場資料。
- canonical card image 是版本化本機資產，由同一 Go server 的一般 HTTP static asset path 提供，不包入 Protobuf message。
- Vue 以 Go API base URL 解析相對 asset URL，在 browser session 依 `card_data_version` 快取 catalog。
- PlayerView 的版本缺少、catalog digest 不符或版本不一致時阻擋所有 command 並顯示版本錯誤，不改用其他 catalog。
- 瀏覽器單張圖片暫時載入失敗只顯示 placeholder 與 retry；不得因此替換 CardRef 或阻斷 Game Module。

## 單一活動分頁

- 同一個 PlayerSession 同一時間最多只能有一個活動的 Vue 遊戲連線。
- 不支援以同一 seat token 在多個分頁同時觀看或操作同一局。
- server 必須在連線層執行限制，不能只依賴 Vue 的 local flag 或瀏覽器事件，避免複製 sessionStorage 後繞過。
- 新的 `WatchGame` 驗證 PlayerSession 後成為最新活動連線，原子取代相同 token 的舊連線；不等待舊 TCP／HTTP stream timeout。
- server 為新活動連線核發短期 opaque connection handle；後續遊戲 RPC 必須同時驗證 seat token 與目前 connection handle。
- connection handle 屬於 transport session，不是 PlayerView handle，不得傳入 Game Module 或 canonical Replay。
- 被取代的舊 stream 以結構化 `Aborted`／`session_replaced` 結束；這是終止狀態，不套用一般網路錯誤的自動重連。
- 舊分頁顯示「此對局已在另一個分頁開啟」，停止 heartbeat、事件補取與 command，且不能只憑仍有效的 seat token 恢復操作。
- 新連線接管時清除該座位尚未提交的 Declaration Draft，不自動恢復或重送。
- 正常頁面重新整理會以新 `WatchGame` 取代可能尚未被 server 察覺關閉的舊 stream，因此不產生暫時鎖死。
- 若舊連線稍後送達關閉通知，不得撤銷或中斷已接管的新 connection handle。

## 可見事件歷史

完整事件歷史不內嵌於每一份 `PlayerView`，避免 snapshot 大小隨對局時間無界增長。

- 既有無界的 `PlayerView.VisibleEvents` 移除，不保留相容欄位。
- `ListVisibleEventBatches` 以 game、玩家身分及不透明 cursor 分頁取得依法可見的 EventBatch。
- 頁面順序固定由舊到新；每個 EventBatch 保留事件邊界與因果關聯，不拆成無關的單筆文字。
- cursor 只代表該玩家可見事件序列的位置，不能推導對手隱藏事件數量或內部 event ID。
- 即時更新可以附帶新公開的 EventBatch 供 event log 與動畫使用，但事件不是重建目前 PlayerView 的必要輸入。
- 若即時事件遺漏或動畫被跳過，Vue 以最新 PlayerView 保持正確，並可從最後 cursor 補取事件歷史。

動畫與正式輸入鎖定由 [`vue.md`](../vue.md#權威狀態動畫與輸入鎖)定義；`WatchGame` 不等待 client 動畫完成才推進或發布下一個 snapshot。

## 結構化錯誤

- 預期拒絕使用穩定 code、非敏感參數與 correlation id，不回傳已翻譯句子。
- 至少區分非法 action／choice／target／payment、stale revision／ViewHandle、Draft invalidated、session takeover、無效 token、game 不存在、card data version 不符與 NeedsRuling。
- Vue 依 code 對應 i18n key；未知 code 視為 contract mismatch，使用安全通用訊息。
- 未授權錯誤不得用差異化訊息揭露 game 是否存在。

## 尚待確認

- production CORS、TLS 與 reverse proxy（不阻擋本機首版）。

## 測試接縫

主要 contract seam 是「產生的 ConnectRPC client → `httptest` HTTP server → 真實 Connect handler → 真實遊戲應用層與 Game Module」。此 seam 驗證 session、connection handle、PlayerView、revision、event cursor、Draft、command、背壓與資訊隔離，不 mock 規則合法性或隱藏資訊。

必要範圍：

- CreatePracticeGame 核發 game、seat token 與 card data version。
- token hash、跨 seat／game 拒絕、接管與舊 handle 失效。
- WatchGame 首份 snapshot、heartbeat、revision 跳號與慢 subscriber。
- EventBatch cursor 分頁、補取與玩家可見性。
- command 只回 accepted revision；Draft 更新回 Draft step。
- catalog version、static asset、missing version 與 CORS。
- deterministic fixture 可供 built Vue 執行跨瀏覽器端對端測試。

## 驗收條件

1. Vue repository 不需要讀取 `go-tcg` internal Go 型別即可產生型別安全 client。
2. 瀏覽器可直接透過 HTTP 呼叫 ConnectRPC endpoint，無額外 gateway。
3. API 只回傳指定玩家的 PlayerView，不洩漏完整 GameState。
4. CLI、bot 與遠端 Vue 對相同行動使用相同規則判定。
5. repository 中沒有重複功能的 REST、WebSocket 或相容 API 路徑。
6. `go-tcg` 不依賴 BSR 或 TypeScript npm SDK 才能建置及提供 API。
7. Vue component 不直接配置 transport；base URL、認證與錯誤處理集中在 API adapter。
8. API contract 版本與 `card_data_version` 可分別升級及驗證。
9. `WatchGame` 首筆及後續 revision 都能獨立解讀成完整 PlayerView，不需先套用舊訊息。
10. 重新連線直接取得最新 snapshot，不需補送狀態 delta。
11. PlayerView 大小不因整場事件歷史持續增長；事件能以玩家專屬 cursor 分頁補取。
12. 洗牌或失去追蹤權後，最新 snapshot 不再包含已撤銷 handle，即使前端漏掉中間 revision 也不洩漏目前資訊。
13. 正式 command response 不包含 PlayerView；Vue 只在 WatchGame snapshot 更新正式畫面。
14. Vue 在收到不小於 `accepted_revision` 的 snapshot 前保持 processing，且不會重複提交相同意圖。
15. Draft update 直接回傳 Draft step；取消 draft 不等待正式 state revision。
16. 慢速 subscriber 不阻塞 Game actor，且尚未傳送的舊 snapshot 會被較新 revision 取代。
17. Vue 收到跳號 revision 後仍能直接得到正確完整狀態，並以 cursor 補齊可見 EventBatch。
18. 無狀態更新時每 15 秒送出 heartbeat，且 heartbeat 不改變 revision。
19. 可重試連線錯誤會以有上限的指數退避重連；認證、授權、找不到遊戲及版本不相容不會無限重試。
20. 重連期間不能提交 command，Declaration Draft 被清除且不自動重送。
21. client 無法以 request 欄位切換 PlayerView 座位；所有視角與 command actor 均由 PlayerSession 推導。
22. game ID、對手 handle 或跨 game token 不授予額外資訊，未授權錯誤不揭露 game 是否存在。
23. token 不出現在 URL 或 localStorage；重新整理可由 sessionStorage 恢復同一座位。
24. server 不保存可直接使用的原始 bearer token，且能撤銷單一 PlayerSession。
25. 本機 Vue dev server 可使用 bearer token 跨 origin 呼叫 loopback Go API，不需要關閉瀏覽器安全機制或允許任意 origin。
26. 同一 PlayerSession 無法同時維持兩個可操作的 Vue 遊戲連線，舊連線被停用後不能繼續提交 command。
27. 新 WatchGame 會立即取代舊連線；重新整理無須等待舊 stream timeout。
28. 被取代的分頁收到 `session_replaced` 後停止自動重連，舊 connection handle 無法呼叫任何遊戲 RPC。
29. 延遲到達的舊連線關閉通知不會影響目前活動連線。
30. 使用者可從本機啟動頁建立固定人類對 bot 單局並進入遊戲，不需要帳號、大廳或牌組選擇。
31. 缺少對應 PlayerSession token 時，game route 不會只憑 game ID 顯示任何 PlayerView。
32. 本機 Vue 可從 Go server 取得與 PlayerView 完全相同版本的 catalog 及 canonical images，不依賴外部網路。
33. catalog 版本不符會阻擋 command；單張圖片載入失敗只影響呈現並可重試。
