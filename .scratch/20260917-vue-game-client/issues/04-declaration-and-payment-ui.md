# 04 — 交付 Declaration 與 Pay Cost 介面

**What to build:** 提供完整 Declaration wizard，讓玩家從區域中的合法卡牌開始，逐步選擇 Parameters、Modes、Targets 與支付來源，最後確認或無副作用取消。

**Blocked by:** 03 — 交付區域瀏覽與不可用原因；Go Player API 04 — 交付 Declaration 與 Pay Cost

**Status:** ready-for-agent

- [ ] wizard 依 Parameters、Modes、Targets、Pay Cost、Confirm 顯示目前 step
- [ ] Done 只完成目前多選 step，Confirm 才提交完整 Declaration
- [ ] Pay Cost 顯示 Hand、Memory 與 Field 的所有可選來源
- [ ] 隨機 Memory 只顯示剩餘或待支付張數，不預先揭露卡牌
- [ ] Cancel 關閉 Draft 且不修改正式遊戲狀態
- [ ] stale revision、ViewHandle、重連或接管會清除 Draft 並顯示原因
- [ ] command 成功後等待 WatchGame 到達 accepted revision，不樂觀修改棋盤
- [ ] component、store 與真實 Go E2E 測試覆蓋成功、取消、非法與失效流程
