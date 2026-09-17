# 03 — 交付區域瀏覽與不可用原因

**What to build:** 讓玩家從每個實際棋盤區域開啟同步瀏覽視窗、檢視可見卡牌，並在選到不可操作項目時理解 Go 回傳的原因。

**Blocked by:** 02 — 交付 Card Inspector 與主要棋盤區域

**Status:** ready-for-agent

- [ ] Main Deck、Material Deck、Memory、Graveyard 與 Banishment 都從實際區域入口開啟
- [ ] 區域視窗只顯示 PlayerView 授權的卡牌或張數
- [ ] 純瀏覽視窗在新 revision 到達時保持同步
- [ ] 點擊卡牌只更新 Inspector 或開始引擎允許的流程，不直接提交 action
- [ ] 不可操作卡牌仍可檢視，並顯示 reason code 對應的 i18n 訊息
- [ ] Close 關閉瀏覽視窗並把 focus 還給原觸發元素
- [ ] Vue Test Utils 覆蓋同步、空狀態、不可用原因與 focus recovery
- [ ] Playwright 驗證公開與私密區域不發生資訊洩漏
