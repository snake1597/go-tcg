# 03 — 完成區域瀏覽與不可用原因

**What to build:** 讓玩家從每個實際棋盤區域開啟同步瀏覽視窗、檢視可見卡牌，並依 fixture 提供的 availability 理解不可操作原因。

**Blocked by:** 02 — 完成棋盤殼層、卡牌與 Inspector

**Status:** ready-for-agent

- [ ] Main Deck、Material Deck、Memory、Graveyard 與 Banishment 都從實際區域入口開啟
- [ ] 區域視窗只顯示 fixture PlayerView 提供的卡牌或張數
- [ ] 純瀏覽視窗在新 fixture revision 到達時保持同步並維持捲動錨點
- [ ] 點擊卡牌只更新 Inspector 或 emit 開始行動意圖，不直接製造 action
- [ ] 不可操作卡牌仍可檢視，並顯示 reason code 對應的 i18n 訊息
- [ ] Close 關閉瀏覽視窗並把 focus 還給原觸發元素或區域控制
- [ ] Vue Test Utils 覆蓋同步、空狀態、不可用原因與 focus recovery
- [ ] fixture-based Playwright 驗證 UI 不會渲染 fixture 中未提供的隱藏內容
- [ ] 此票不宣稱已驗證 server-side 資訊隔離
