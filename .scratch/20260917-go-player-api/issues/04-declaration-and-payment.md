# 04 — 交付 Declaration 與 Pay Cost

**What to build:** 提供 session-scoped Declaration Draft，讓玩家可逐步完成 Parameters、Modes、Targets 與 Pay Cost，最後以單一原子輸入確認行動。

**Blocked by:** 03 — 交付 Card Catalog、圖片與完整區域投影

**Status:** ready-for-agent

- [ ] 可從有效 Legal Action 或 ability 建立 Draft，且不修改 canonical Game State
- [ ] 每個 step 回傳選項、已選值、限制、完成條件與結構化不可用原因
- [ ] Pay Cost 包含 Hand、Memory 與 Field 上所有可選支付來源
- [ ] 隨機 Memory 支付只回傳數量，不回傳未揭露卡牌身分
- [ ] Done 只完成目前多選步驟，Confirm 才提交完整 Declaration
- [ ] Game Module 在 Confirm 時重新驗證並原子處理輸入
- [ ] Cancel、重連、session takeover、stale revision 或 stale ViewHandle 會清除或拒絕 Draft
- [ ] 真實 ConnectRPC contract test 覆蓋成功、取消、失效與非法支付
