# 07 — 完成 Command 與連線生命週期

**What to build:** 讓正式 command、WatchGame、heartbeat、背壓、重連與單分頁接管形成一致的玩家 session 生命週期，使慢速或過期 client 不會阻塞或污染遊戲。

**Blocked by:** 04 — 交付 Declaration 與 Pay Cost；05 — 交付 Combat、Intent 與 Effects Stack；06 — 交付玩家範圍 EventBatch 歷史

**Status:** ready-for-agent

- [ ] SubmitAction、SubmitChoice 與 ConfirmDeclaration 成功只回 accepted_revision
- [ ] command 必須攜帶並驗證目前有效的 connection handle
- [ ] 每個 subscriber 只保留最新待送 snapshot，慢 client 不阻塞 Game Host
- [ ] WatchGame 每 15 秒發送 heartbeat
- [ ] 新 WatchGame 原子接管同一 PlayerSession，舊 stream 收到 Aborted／session_replaced
- [ ] 舊 handle 的 command 被拒絕，舊 stream 延遲關閉不影響新 handle
- [ ] 暫時 transport 錯誤與不可重試錯誤具有穩定分類
- [ ] 背壓、接管、accepted revision 與延遲關閉 contract test 通過
