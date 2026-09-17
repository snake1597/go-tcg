# 07 — 完成無障礙、效能與跨瀏覽器發布門檻

**What to build:** 將完整練習對局收斂成可發布的桌面體驗，確保鍵盤、輔助技術、reduced motion、最低解析度、效能預算與主流瀏覽器均達標。

**Blocked by:** 06 — 交付重連、接管與分層錯誤 UX；Go Player API 08 — 完成 Go Contract 與安全發布門檻

**Status:** ready-for-agent

- [ ] 所有核心操作可只用鍵盤完成，dialog focus 與關閉後 focus recovery 正確
- [ ] 選取、Rested、可用性與錯誤不只依賴顏色
- [ ] 重要 revision、Decision Player 與錯誤以適度 ARIA live region 宣告
- [ ] prefers-reduced-motion 使用者可完成相同核心流程
- [ ] 1280×720 與 1920×1080 均無阻擋操作的溢位或遮蔽
- [ ] catalog 與首份 snapshot 後兩秒內呈現可理解棋盤，本機回饋目標 100ms 內
- [ ] 圖片 lazy-load，長事件或大型區域清單按需要虛擬化
- [ ] Chromium、Firefox、WebKit 皆以 built Vue 和真實 Go server 完成整場練習對局
- [ ] release gate 覆蓋 Declaration、動畫、重連、接管與結束狀態
