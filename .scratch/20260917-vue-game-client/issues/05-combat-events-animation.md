# 05 — 交付 Combat、Stack、事件與動畫

**What to build:** 呈現中央對局狀態、Effects Stack 與可翻頁事件紀錄，並以獨立 presentation state 播放 EventBatch 動畫而不延遲權威同步。

**Blocked by:** 02 — 交付 Card Inspector 與主要棋盤區域；Go Player API 05 — 交付 Combat、Intent 與 Effects Stack；Go Player API 06 — 交付玩家範圍 EventBatch 歷史

**Status:** ready-for-agent

- [ ] Turn Player、Opportunity Holder、Decision Player 與 Phase 分開呈現
- [ ] Combat／Intent 顯示攻擊者、防禦者、目標與 wielded Weapon
- [ ] Effects Stack 顯示穩定順序、stack identity 與 Source Card
- [ ] 事件依 EventBatch 由舊至新顯示，並可向前載入較舊頁面
- [ ] gameView 立即保存最新 snapshot，presentation 可短暫落後
- [ ] presentation 落後時鎖定正式輸入，但保留 Inspector、瀏覽與事件操作
- [ ] 動畫 queue 有界且可跳至最新狀態，事件歷史仍完整
- [ ] reduced motion 能減少或移除非必要動畫
- [ ] store test 與 Playwright 驗證動畫落後、追趕及 cursor 補取
