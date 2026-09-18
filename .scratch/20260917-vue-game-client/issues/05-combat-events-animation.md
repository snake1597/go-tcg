# 05 — 完成 Combat、Stack、事件與動畫

**What to build:** 使用 scripted fixtures 呈現中央對局狀態、Effects Stack 與事件紀錄，並以獨立 presentation state 播放 EventBatch 動畫而不延遲權威 UI state。

**Blocked by:** 02 — 完成棋盤殼層、卡牌與 Inspector

**Status:** ready-for-agent

- [ ] Turn Player、Opportunity Holder、Decision Player 與 Phase 分開呈現
- [ ] Combat／Intent 顯示攻擊者、防禦者、目標與 wielded Weapon
- [ ] Effects Stack 顯示穩定順序、stack identity 與 Source Card
- [ ] 事件依 EventBatch 由舊至新顯示，並可向前載入 scripted older pages
- [ ] gameView 立即保存最新 fixture snapshot，presentation 可短暫落後
- [ ] presentation 落後時鎖定正式輸入，但保留 Inspector、瀏覽與事件操作
- [ ] animation queue 有界且可跳至最新狀態，事件紀錄仍完整
- [ ] reduced motion 能減少或移除非必要動畫
- [ ] store test 與 fixture-based Playwright 驗證動畫落後、追趕、跳號與 cursor 補取 UI
