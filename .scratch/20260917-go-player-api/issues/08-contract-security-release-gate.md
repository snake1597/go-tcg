# 08 — 完成 Go Contract 與安全發布門檻

**What to build:** 將玩家 API 的完整合約、安全界線與 deterministic 測試環境收斂為發布門檻，供獨立 Vue client 穩定執行端對端對局。

**Blocked by:** 07 — 完成 Command 與連線生命週期

**Status:** ready-for-agent

- [ ] 所有玩家 RPC 都驗證 Protobuf 欄位、輸入大小、cursor、Draft selection 與 bearer token
- [ ] 本機 HTTP 只接受 loopback 與明確設定的 Vue origin，不允許 wildcard 或 cookie credential
- [ ] log、diagnostic、error 與 event 不包含 token 或該 seat 不可見資訊
- [ ] 所有預期拒絕回傳穩定 reason code、結構化參數與 correlation id
- [ ] NeedsRuling 會停止對局並安全投影，不猜測規則結果
- [ ] generated client 到真實 handler／Game Module 的完整 contract suite 通過
- [ ] 提供固定 seed 或 deterministic fixture 供 Vue 跨瀏覽器端對端測試
- [ ] 文件化的 API 行為與已發布 Protobuf contract 一致
