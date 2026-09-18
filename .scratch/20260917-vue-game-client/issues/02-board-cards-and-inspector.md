# 02 — 完成棋盤殼層、卡牌與 Inspector

**What to build:** 使用最大密度 fixture 完成雙方核心棋盤、卡牌呈現與 Card Inspector，讓滑鼠及鍵盤使用者能在目標桌面尺寸理解局面。

**Blocked by:** 01 — 建立 Vue 基礎與 Fixture Seam

**Status:** ready-for-agent

- [ ] 1920×1080 與 1280×720 都能呈現完整基本棋盤，對手在上、玩家在下
- [ ] Champion／Lineage 固定在各自棋盤左側
- [ ] Field 卡只顯示卡圖與名稱，順序穩定且不可拖曳
- [ ] Hand、雙方公開張數及主要區域入口依 fixture PlayerView 呈現
- [ ] Awake、Rested 與 Banishment 使用已確認的方向及非純色彩狀態提示
- [ ] hover／keyboard focus 暫時更新 Card Inspector，點擊可鎖定及解除
- [ ] 圖片 loading、placeholder 與 retry 不阻擋其他棋盤資訊
- [ ] Vue Test Utils 與 fixture-based Playwright 覆蓋滑鼠、鍵盤及最大密度布局
