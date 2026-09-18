# Go Player API for Vue

Status: ready-for-agent

## Problem Statement

現有 Go Game Module 能執行規則與產生玩家視角，但沒有可供獨立 Vue client 使用的 HTTP contract、PlayerSession、遊戲應用生命週期、Declaration Draft、事件分頁或版本化卡牌資產。直接暴露現有方法會讓 transport、隱藏資訊、並行輸入與 UI 工作流程互相滲透，也無法安全支援斷線與單分頁接管。

## Solution

在 Game Module 外新增遊戲應用層與 ConnectRPC 玩家 API。應用層序列化每場 game 的所有輸入、管理 seat-scoped session、Draft、watcher 與 event cursor；Game Module 仍是唯一規則權威。PlayerView 改成完整、卡牌身分穩定且可直接供 Vue 呈現的玩家 snapshot，並由版本化 Card Catalog 補足卡面內容與圖片。

## User Stories

1. As a Vue client, I want to 建立固定人類對 Bot 練習對局, so that 我能取得 game id、token 與 card data version
2. As a 安全設計者, I want to 讓 token 綁定 game 與 seat, so that client 無法自行選擇玩家視角
3. As a 安全設計者, I want to 只保存 token hash, so that server 儲存不暴露 bearer secret
4. As a 玩家, I want to 讓 token 只回傳一次, so that secret 的生命週期清楚
5. As a Vue client, I want to 以 WatchGame 取得首份完整 PlayerView, so that 棋盤能一次建立
6. As a Vue client, I want to 在 revision 改變時取得最新完整 snapshot, so that 不需實作 field delta 合併
7. As a 慢速 client, I want to 允許跳過中間 snapshot, so that 我不會阻塞遊戲
8. As a 玩家, I want to 讓事件歷史獨立於 snapshot 保存, so that 跳 revision 不會遺失紀錄
9. As a Vue client, I want to 用 cursor 取得最近與較舊 EventBatch, so that 可漸進顯示歷史
10. As a 玩家, I want to 只收到自己可見的事件, so that 隱藏資訊不被洩漏
11. As a 玩家, I want to 在 PlayerView 看見雙方公開局面, so that 棋盤能完整呈現
12. As a 玩家, I want to 看見自己的 Hand 與對手 Hand count, so that 私密資訊正確隔離
13. As a 玩家, I want to 依規則查看 Main Deck、Material Deck 與 Memory, so that 受控查看能安全運作
14. As a 玩家, I want to 看見 Graveyard、Banishment、Champion 與 Lineage, so that 所有公開區域可瀏覽
15. As a 玩家, I want to 看見 Effects Stack 的 stack identity 與來源卡, so that 來源離場後仍可理解效果
16. As a 玩家, I want to 看見 Combat、Intent、目標與 wielded Weapon, so that UI 不需猜測戰鬥關係
17. As a Vue client, I want to 以 CardRef 與 ObjectRef 識別內容和物件, so that 同名卡與場上實例不混淆
18. As a Vue client, I want to 取得 Turn、Opportunity 與 Decision player, so that 三種角色可分開呈現
19. As a Vue client, I want to 取得 Legal Actions 與 unavailable reasons, so that UI 只呈現引擎已知選項
20. As a 玩家, I want to 啟動不修改正式狀態的 Declaration Draft, so that 我能逐步組合行動
21. As a 玩家, I want to 逐步選 Parameters、Modes、Targets 與費用, so that 複雜宣告可被導引
22. As a 玩家, I want to 看見所有可選支付來源, so that 我能自行選擇非隨機成本
23. As a 玩家, I want to 讓隨機 Memory 只暴露數量, so that banish 前不洩漏卡牌
24. As a 玩家, I want to 以 Confirm 原子化提交完整 Draft, so that canonical state 不會處於半完成狀態
25. As a 遊戲引擎, I want to 在 Confirm 再次驗證 input, so that Draft 不會取代規則權威
26. As a 玩家, I want to 讓 stale revision 或 ViewHandle 使 Draft 失效, so that 舊選擇不能污染新局面
27. As a Vue client, I want to 讓 command 成功只回 accepted revision, so that 正式狀態仍由 WatchGame 發布
28. As a 遊戲主機, I want to 每場 game 序列化 human、Bot 與排程輸入, so that 狀態轉移不會競爭
29. As a Bot controller, I want to 只用自己的 PlayerView 與 Legal Actions, so that Bot 不繞過資訊邊界
30. As a Game actor, I want to 不被 subscriber 網路速度阻塞, so that 對局能持續前進
31. As a Vue client, I want to 每 15 秒收到 heartbeat, so that 我能辨識停滯連線
32. As a 玩家, I want to 讓新的 WatchGame 接管舊分頁, so that 一個座位只有一個活動控制端
33. As a 舊分頁, I want to 收到 session_replaced, so that 我會停止送 command 與重連
34. As a 新分頁, I want to 不受舊 stream 延遲關閉影響, so that 接管不會被反向撤銷
35. As a Vue client, I want to 區分可重試和不可重試錯誤, so that 重連策略正確
36. As a Vue client, I want to 取得穩定 reason code 與參數, so that 我能使用 i18n 呈現訊息
37. As a 維運者, I want to 使用 correlation id 追查未知錯誤, so that 不需在訊息中洩漏內部細節
38. As a Vue client, I want to 取得指定版本 Card Catalog, so that CardRef 能解析成卡面
39. As a 玩家, I want to 使用 canonical card image, so that 同張卡有一致外觀
40. As a 建置者, I want to 在缺圖或 digest 不符時失敗, so that 不完整資產不會發布
41. As a Vue client, I want to 在 catalog 版本不符時得到不可重試錯誤, so that 不會以錯誤內容操作
42. As a 本機開發者, I want to 從明確 Vue origin 呼叫 loopback API, so that 本機開發可用且 CORS 有邊界
43. As a 安全設計者, I want to 驗證所有 Protobuf、cursor 與 Draft 輸入, so that 惡意資料不進入 Game Module
44. As a 安全設計者, I want to 確保 log 不含 token 或隱藏牌, so that 診斷不造成資料洩漏
45. As a 規則維護者, I want to 在 NeedsRuling 時停止並傳結構化狀態, so that 系統不猜測未裁定規則
46. As a Go 開發者, I want to 讓 transport 只做 mapping, so that 規則與資訊投影保持在深層 Game Module
47. As a Go 開發者, I want to 移除舊名稱式與重複 PlayerView 欄位, so that 新合約沒有雙重真相來源
48. As a 發布者, I want to 以真實 Connect handler 與 Game Module 驗證 contract, so that API 行為可被可靠驗收

## Implementation Decisions

- 保持模組化單體；新增應用層，不把 session、HTTP 或 Draft 放入 Game Module。
- 每場 game 使用序列化 mailbox，human、Bot、choice 與 scheduler 都走相同入口。
- PlayerSession 是 seat-scoped opaque bearer token，server 只保存 hash；request 不接受 player identity。
- connection handle 是短效且屬於最新 WatchGame，正式 command 必須驗證它。
- 單一 PlayerSession 只允許一個活動 WatchGame；新連線原子接管舊連線。
- WatchGame 傳完整 PlayerView snapshot；subscriber 使用單格 latest-value buffer。
- PlayerView 使用 CardRef、ObjectRef、Stack identity 與玩家範圍 zone projection。
- `VisibleEvents` 從 PlayerView 移除；EventBatch 使用 opaque cursor 獨立取得。
- 名稱式 action identity、通用 Cards 欄位與被新投影取代的欄位直接移除，不加相容層。
- Declaration Draft 由應用層保存，綁定 session、game、revision、ViewHandle 與 connection handle。
- Confirm 將 Draft 轉成一個原子 Game input，Game Module 最終重新驗證。
- command 成功只回 accepted revision；Draft RPC 回目前 Draft step。
- Card Catalog 由 repository card data 建置，綁定明確版本與 canonical image digest。
- 圖片使用一般 HTTP static assets；Protobuf 只提供 URL 與 metadata。
- Go repository 擁有 Protobuf source 與 Go generation；TypeScript client 產生與交付不屬於此範圍，也不要求 BSR。
- 本機 HTTP 只允許 loopback 與明確 Vue origin，不使用 wildcard CORS 或 cookie。
- Go 回傳穩定 code、參數與 correlation id，不回傳翻譯後 UI 句子。
- Bot 使用自己的 PlayerView／Legal Actions，經同一 mailbox 操作；不新增策略訓練能力。

## Testing Decisions

- 主要 seam 是產生的 ConnectRPC client 經 `httptest` HTTP server 呼叫真實 Connect handler 與真實 Game Module。
- 在主要 seam 驗證 CreatePracticeGame、token、single-tab takeover、WatchGame、accepted revision、Draft、event cursor、catalog 版本與結構化錯誤。
- 資訊隔離測試以兩個 seat 的真實 PlayerView／EventBatch 比較外部輸出，不測私有 helper。
- 背壓測試建立慢 subscriber，驗證 game revision 仍前進、snapshot 可跳號、事件仍完整。
- 接管測試驗證舊連線被 Aborted、舊 command 被拒絕、延遲關閉不影響新 handle。
- Game Module 沿用 PlayerView／Submit scenario tests 驗證規則，不在 handler mock legality。
- catalog builder、token hash 與 cursor codec 可用窄單元測試，但不能取代 contract tests。
- 端對端 release gate 使用 built Vue 與真實 Go server；測試 game 使用固定 seed 或 deterministic fixture。

## Out of Scope

- Vue component、store、動畫、i18n 文案與瀏覽器 focus。
- TypeScript client 的實際 generation／distribution 與 BSR。
- Replay seek、branch、undo、歷史 snapshot 與長期 replay 保存。
- Bot 策略訓練、難度、風格、人類介入與策略選擇器。
- 帳號、登入、房間、邀請、配對、觀戰。
- Production TLS、正式 ingress、跨網域部署與長期 session storage。
- 舊 PlayerView、舊 action identity 或事件欄位的相容支援。

## Further Notes

- 此合約是破壞性的新玩家 API；實作時直接更新既有呼叫端與測試，避免雙軌模型。
- Live EventBatch 支援 UI 動畫與紀錄，不代表 replay 機制已定案。
- PlayerView 與 EventBatch 必須在 Game Module 的 KnowledgeState 邊界完成資訊裁切，transport 不可補救過度暴露的資料。
- Vue spec 的端對端驗收依賴本 spec 全部核心能力。
