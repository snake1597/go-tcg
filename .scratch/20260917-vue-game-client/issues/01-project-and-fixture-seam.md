# 01 — 建立 Vue 基礎與 Fixture Seam

**What to build:** 建立獨立 Vue SPA 與 Game Client Interface，讓 UI 可透過 deterministic Fixture Adapter 載入完整 PlayerView、Draft、EventBatch 與錯誤情境，不等待 Go API 即可開發及驗證。

**Blocked by:** None — can start immediately

**Status:** ready-for-agent

- [ ] 專案使用 Vue 3、Vite、TypeScript strict、Composition API、Vue Router、Pinia 與 vue-i18n
- [ ] UI 與 stores 只依賴 Game Client Interface，不直接依賴 ConnectRPC 或 Protobuf message
- [ ] Fixture Adapter 能載入版本化 Card Catalog、完整 PlayerView、Draft step、EventBatch 與結構化錯誤情境
- [ ] Fixture 資料遵守 PlayerView、Declaration 與 ConnectRPC 規格的語意及身分邊界
- [ ] Fixture Adapter 不計算合法性、cost、target、trigger、隱藏資訊或 Bot 行動
- [ ] 開發模式可明確選擇 fixture scenario，production build 不會誤用 Fixture Adapter
- [ ] Vitest 驗證 Interface、fixture 選擇、情境載入與錯誤輸出
