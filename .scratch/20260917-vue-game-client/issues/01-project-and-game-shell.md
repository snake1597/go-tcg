# 01 — 建立 Vue 專案與最小對局殼層

**What to build:** 建立獨立 Vue SPA，讓玩家可從本機首頁建立 practice game、保存 PlayerSession、載入 Card Catalog，並透過 WatchGame 看見第一個可理解的桌面棋盤。

**Blocked by:** Go Player API 02 — 建立本機 Create／Watch 垂直切片

**Status:** ready-for-agent

- [ ] 專案使用 Vue 3、Vite、TypeScript strict、Composition API、Vue Router、Pinia 與 vue-i18n
- [ ] 首頁可呼叫 CreatePracticeGame 並導向對局頁
- [ ] token 只存於 sessionStorage，不出現在 URL 或 localStorage
- [ ] API adapter 使用產生的 TypeScript ConnectRPC client，component 不直接持有 client
- [ ] 對局頁載入相容 Card Catalog 與首份 WatchGame snapshot
- [ ] 1920×1080 與 1280×720 都能看見基本棋盤區塊
- [ ] Playwright 以 built Vue 和真實 Go server 驗證建立及進入對局
