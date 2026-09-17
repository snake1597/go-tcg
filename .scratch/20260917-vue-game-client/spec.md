# Vue Game Client

Status: ready-for-agent

## Problem Statement

玩家目前只能透過 Go 端的非圖形入口執行對局，無法在符合卡牌遊戲心智模型的棋盤上理解局面、查看自己的合法操作、完成複雜宣告，或在斷線與動畫期間安全地繼續遊戲。既有視覺參考尚未形成可實作、可驗收的獨立 Vue 產品規格。

## Solution

建立一個獨立 Vue 3 桌面 SPA，透過產生的 TypeScript ConnectRPC client 連線本機 Go 服務。它提供極簡的人類對 Bot 練習入口、完整玩家視角棋盤、實際區域導向的卡牌互動、Declaration／Pay Cost 流程、事件與動畫、明確的連線生命週期，以及鍵盤與 reduced-motion 支援。所有正式狀態與合法性以 Go 傳來的 PlayerView 和結構化選項為準。

## User Stories

1. As a 練習玩家, I want to 從本機首頁建立固定人類對 Bot 對局, so that 我不需先設定帳號或房間
2. As a 練習玩家, I want to 建立後直接進入對局頁, so that 開局流程保持簡單
3. As a 玩家, I want to 看見上下鏡像但都正向的雙方棋盤, so that 我能快速理解局面
4. As a 玩家, I want to 在左上 Card Inspector 查看完整卡牌, so that 棋盤本身可以保持精簡
5. As a 滑鼠使用者, I want to hover 卡牌時暫時查看資訊, so that 我能快速掃描場面
6. As a 鍵盤使用者, I want to focus 卡牌時得到同等資訊, so that 我不依賴滑鼠
7. As a 玩家, I want to 點擊鎖定 Card Inspector, so that 游標移開後仍可閱讀
8. As a 玩家, I want to 點擊實際區域後再選要出的牌, so that 操作與卡牌所在位置一致
9. As a 玩家, I want to 在區域視窗查看不可操作卡牌, so that 我仍能研究完整公開資訊
10. As a 玩家, I want to 看見不可用的結構化原因, so that 我知道如何改變選擇
11. As a 玩家, I want to 在 Field 只看卡圖與名稱, so that 大量卡牌仍保持可讀
12. As a 玩家, I want to 以方向區分 Awake、Rested 與 Banishment, so that 狀態一眼可辨
13. As a 玩家, I want to 看見自己完整 Hand 與對手手牌張數, so that 私密資訊不被洩漏
14. As a 玩家, I want to 瀏覽 Graveyard 與 Banishment, so that 我能查閱公開卡牌
15. As a 玩家, I want to 只在規則授權時查看 Main Deck 或 Material Deck 卡牌, so that UI 遵守資訊邊界
16. As a 玩家, I want to 在 Memory 隨機支付前只看張數, so that 未揭露卡牌保持隱藏
17. As a 玩家, I want to 在 banish 後由事件得知隨機 Memory 卡牌, so that 揭露時機符合規則
18. As a 玩家, I want to 依序完成 Parameters、Modes、Targets、Pay Cost, so that 複雜行動可逐步理解
19. As a 玩家, I want to 用 Done 完成當前多選步驟, so that 步驟與整體提交不會混淆
20. As a 玩家, I want to 用 Confirm 一次提交完整 Declaration, so that 正式狀態只接受原子化行動
21. As a 玩家, I want to 取消尚未確認的 Draft, so that 我可無副作用地重新思考
22. As a 玩家, I want to 在 Pay Cost 看見 Hand、Memory 與 Field 支付來源, so that 我能自行選擇非隨機成本
23. As a 玩家, I want to 看見 Turn Player、Opportunity Holder 與 Decision Player, so that 我知道誰的回合、誰能動、誰須決策
24. As a 玩家, I want to 看見 Effects Stack 的順序與來源卡, so that 我理解效果即將如何處理
25. As a 玩家, I want to 看見 Combat、Intent、目標與 wielded Weapon, so that 戰鬥關係清楚
26. As a 玩家, I want to 依 EventBatch 閱讀由舊至新的紀錄, so that 我能理解局面如何形成
27. As a 玩家, I want to 向前載入較舊事件, so that 長對局不需一次載入全部歷史
28. As a 玩家, I want to 略過過多動畫並追上最新局面, so that 慢呈現不妨礙對局
29. As a 玩家, I want to 在動畫落後時仍能查看唯讀資訊, so that 等待期間仍可分析
30. As a 玩家, I want to 在動畫落後時不能提交正式輸入, so that 操作不會針對過期畫面
31. As a 玩家, I want to 在 command 被接受後等待權威 revision, so that 畫面不會預測錯誤結果
32. As a 玩家, I want to 在暫時斷線時看到重連覆蓋層, so that 我知道遊戲未被放棄
33. As a 玩家, I want to 重連後回到最新 snapshot 並補齊事件, so that 中斷不造成歷史缺口
34. As a 玩家, I want to 在新分頁接管後讓舊分頁停止操作, so that 同一座位不會送出衝突指令
35. As a 玩家, I want to 在卡圖載入失敗時仍看見 placeholder 與重試, so that 棋盤仍可使用
36. As a 玩家, I want to 在 catalog 版本不符時被明確阻擋, so that 我不會用錯誤卡面操作
37. As a 繁體中文使用者, I want to 看見本地化 UI 文案, so that 狀態與錯誤容易理解
38. As a 使用輔助技術的玩家, I want to 以鍵盤操作所有核心流程, so that 我能獨立完成對局
39. As a 使用輔助技術的玩家, I want to 得到適度的狀態與錯誤宣告, so that 重要變化不依賴視覺
40. As a 對動態敏感的玩家, I want to 使用 reduced motion, so that 動畫不造成不適
41. As a 桌面玩家, I want to 在 1280×720 以上維持完整操作, so that 常見桌面尺寸都能遊玩
42. As a 開發者, I want to 讓 component 只消費 store 與 view model, so that transport 與規則責任不滲入畫面
43. As a 開發者, I want to 只由 API adapter 呼叫 RPC, so that 連線與錯誤政策集中管理
44. As a 開發者, I want to 分離權威 gameView 與動畫 presentation, so that 動畫不延遲資料同步
45. As a 維護者, I want to 以 i18n key 處理 reason code, so that Go 不需傳送已翻譯句子
46. As a 發布者, I want to 在三種瀏覽器引擎上跑真實端對端流程, so that 首版有可重現的發布門檻

## Implementation Decisions

- Vue client 是獨立 repository 的 Vue 3、Vite、TypeScript strict SPA。
- 使用 Composition API、script setup、Vue Router、Pinia 與 vue-i18n；不使用 Nuxt、SSR、Connect Query 或大型 UI framework。
- 使用產生的 TypeScript client、Connect 與 Connect Web，以 native Connect protocol 直連本機 Go HTTP server。
- 首版路由只有本機建立頁與對局頁；建立固定人類對 Bot 對局。
- 介面為 16:9 桌面優先，1920×1080 為基準、1280×720 為最低，不支援行動版。
- 棋盤分成 Card Inspector、雙方棋盤、中央狀態帶、Effects Stack、事件紀錄與動作區。
- Field 採卡圖加名稱的 compact tile；不提供拖曳與手動重排。
- hover／focus 暫時更新 Inspector，點擊鎖定；卡牌點擊不直接提交正式 action。
- 出牌與發動行動從卡牌實際所在區域開始，不建立全域 Playable Cards 視窗。
- Declaration 依 Parameters、Modes、Targets、Pay Cost、Confirm 導引；Done 與 Confirm 語意分離。
- 所有非隨機支付來源由玩家選擇；隨機 Memory 只顯示數量，揭露由後續事件呈現。
- 純瀏覽區域隨 revision 同步；Draft 基礎 revision 失效即清除。
- 最新 PlayerView 是權威狀態，presentation store 可為動畫落後；落後期間鎖定正式輸入。
- command 不樂觀更新；以 accepted revision 等待 WatchGame snapshot。
- PlayerSession token 只存 sessionStorage；同一 session 只允許一個活動分頁。
- 新 WatchGame 接管舊連線；舊分頁停止 command 與自動重連。
- 事件以 cursor 獨立載入，不從 PlayerView 重建歷史。
- card catalog 與 image assets 由 Go 提供，Vue 以 CardRef 查找。
- Store 分為 session、catalog、gameView、presentation、declaration、ui；只有 API adapter 接觸 RPC。
- loading／error 分為頁面、阻擋、對局覆蓋、局部四級。
- 所有文案預留 i18n；介面語言與 card catalog 語言彼此獨立。

## Testing Decisions

- 主要驗收 seam 是 built Vue 對真實本機 Go server 的 Playwright 測試，從 CreatePracticeGame 經 catalog、WatchGame、Declaration、Pay Cost、動畫、重連到對局結束。
- 主要端對端測試使用 Chromium、Firefox 與 WebKit，驗證外部可見行為，不查 component 內部狀態。
- Vitest 測試 Pinia store、API adapter 狀態機、accepted revision、重連、session takeover、Draft invalidation 與 presentation queue。
- Vue Test Utils 測試 Card Inspector、區域 dialog、Declaration step、focus recovery、ARIA 狀態與 reduced motion。
- 不 mock 規則合法性、隱藏資訊或 Game Module；規則選項一律來自真實 Go server。
- 視覺及效能驗收涵蓋最低解析度、兩秒首屏、100ms 本機回饋、長事件清單與圖片失敗。

## Out of Scope

- Replay、seek、branch、undo 與 replay UI。
- Bot 策略訓練、難度或風格選擇、人類介入 Bot 行動。
- 帳號、登入、房間、邀請、配對、牌組編輯。
- 行動版、觸控專用版、SSR 與 SEO。
- Production CORS、TLS、正式網域及多裝置 session。
- 在 Vue 重作規則引擎或預測正式狀態。
- 聊天與觀戰模式。

## Further Notes

- UI 以現有棋盤圖片作參考，不要求逐像素複製。
- 官方卡名與術語首版可保留英文；UI 說明與錯誤使用繁體中文。
- TypeScript client 的產生與交付由使用者流程負責，不以 BSR 為必要條件。
- 本 spec 依賴 Go Player API spec 提供完整合約後才能完成端對端驗收。
