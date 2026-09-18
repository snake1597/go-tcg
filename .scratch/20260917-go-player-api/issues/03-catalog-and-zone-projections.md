# 03 — 交付 Card Catalog、圖片與完整區域投影

**What to build:** 讓玩家 API 提供可直接繪製完整棋盤的玩家範圍區域資料，並以版本化 Card Catalog 與 canonical image assets 解析每個 CardRef。

**Blocked by:** 02 — 建立本機 Create／Watch 垂直切片

**Status:** ready-for-agent

- [ ] Hand、Field、Main Deck、Material Deck、Memory、Graveyard、Banishment 與 Champion／Lineage 都有明確玩家視角投影
- [ ] 對手 Hand 與未授權牌庫只暴露允許的張數或卡背資訊
- [ ] Field 提供穩定順序、controller 與 Awake／Rested 狀態
- [ ] Memory 在隨機 banish 前不暴露卡牌身分
- [ ] GetCardCatalog 依 card data version 回傳 CardRef 對應內容
- [ ] canonical image 可由本機 HTTP static asset 載入並具有 digest／快取資訊
- [ ] 缺圖、未知 CardRef、digest 不符或 catalog 版本不相容時以預期方式失敗
- [ ] 資訊隔離 contract test 比較不同 seat 的真實輸出
