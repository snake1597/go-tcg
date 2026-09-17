# 02 — 交付 Card Inspector 與主要棋盤區域

**What to build:** 在權威 PlayerView 上呈現雙方核心棋盤、Field compact cards 與 Card Inspector，讓滑鼠及鍵盤使用者能快速查看卡牌與理解方向狀態。

**Blocked by:** 01 — 建立 Vue 專案與最小對局殼層；Go Player API 03 — 交付 Card Catalog、圖片與完整區域投影

**Status:** ready-for-agent

- [ ] 對手在上、玩家在下，雙方卡牌均保持正向閱讀
- [ ] Champion／Lineage 固定在各自棋盤左側
- [ ] Field 卡只顯示卡圖與名稱，順序穩定且不可拖曳
- [ ] Awake、Rested 與 Banishment 使用已確認的方向呈現
- [ ] hover／keyboard focus 暫時更新 Card Inspector，點擊可鎖定
- [ ] Hand 與對手 Hand count 遵守 PlayerView 資訊邊界
- [ ] 圖片 loading、placeholder 與重試不阻擋其他棋盤資訊
- [ ] component test 與真實對局 E2E 覆蓋滑鼠及鍵盤檢視
