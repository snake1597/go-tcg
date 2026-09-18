# 04 — 完成 Declaration 與 Pay Cost UI

**What to build:** 使用 scripted Draft fixtures 完成 Declaration wizard，讓玩家逐步選擇 Parameters、Modes、Targets 與支付來源，最後確認或無副作用取消。

**Blocked by:** 03 — 完成區域瀏覽與不可用原因

**Status:** ready-for-agent

- [ ] wizard 依 Parameters、Modes、Targets、Pay Cost、Confirm 顯示目前 step
- [ ] Done 只完成目前多選 step，Confirm 才 emit 完整 Declaration 意圖
- [ ] Pay Cost 顯示 Hand、Memory 與 Field 的所有 fixture 支付來源
- [ ] 隨機 Memory 只顯示剩餘或待支付張數，不預先揭露卡牌
- [ ] Cancel 關閉 Draft 且不改寫 gameView
- [ ] stale revision、ViewHandle、重連或接管情境會清除 Draft 並顯示原因
- [ ] accepted revision 情境會等待更新 snapshot，不樂觀修改棋盤
- [ ] UI 只消費 step、availability 與 handle，不自行計算合法性、cost 或 target
- [ ] component、store 與 fixture-based Playwright 覆蓋成功、取消、非法與失效流程
