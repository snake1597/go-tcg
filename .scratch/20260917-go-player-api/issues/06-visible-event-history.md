# 06 — 交付玩家範圍 EventBatch 歷史

**What to build:** 提供與 snapshot stream 分離的玩家可見事件歷史，讓 client 能以 opaque cursor 載入最近及較舊 EventBatch，並可靠驅動動畫與事件紀錄。

**Blocked by:** 03 — 交付 Card Catalog、圖片與完整區域投影

**Status:** ready-for-agent

- [ ] ListVisibleEventBatches 能取得最近一頁並以 cursor 讀取較舊批次
- [ ] 每批保留 batch identity、revision 範圍、因果順序與穩定 entity reference
- [ ] 每個 seat 只取得自己可見的事件內容
- [ ] 隨機 Memory banish、卡牌揭露與區域移動在正確 EventBatch 出現
- [ ] snapshot 跳過 revision 時，事件歷史仍完整且無重複
- [ ] 無效、過期或跨 session cursor 回傳結構化錯誤
- [ ] PlayerView 不再內嵌完整 VisibleEvents
- [ ] contract test 覆蓋分頁、資訊隔離與 revision 缺口
