# 07 — 介接 ConnectRPC

**What to build:** 實作 ConnectRPC Adapter 並在不修改既有 UI Module 的前提下替換 Fixture Adapter，接通真實 CreatePracticeGame、Catalog、WatchGame、事件、Draft、command、PlayerSession 與 connection handle。

**Blocked by:** 06 — 完成 Session、連線與錯誤狀態 UI；Go Player API 08 — 完成 Go Contract 與安全發布門檻

**Status:** ready-for-agent

- [ ] ConnectRPC Adapter 滿足既有 Game Client Interface，component 與 feature stores 不直接 import generated client
- [ ] 首頁可建立 practice game、將 token 存入 sessionStorage 並進入對局頁
- [ ] Catalog、static image URL 與 WatchGame snapshot 正確映射成 UI-facing models
- [ ] Event cursor、Draft RPC、SubmitAction、SubmitChoice 與 accepted revision 正確接通
- [ ] bearer token 與 connection handle 只由 Adapter／session store 管理
- [ ] heartbeat、重連、不可重試錯誤及 session_replaced 驅動既有 UI 狀態機
- [ ] Protobuf mapping test 覆蓋 optional、enum、oneof、未知 reason code 與版本不相容
- [ ] Fixture Adapter 保留給開發及 UI 測試，但不進入 production runtime
- [ ] built Vue 對真實 Go server 完成一條 CreatePracticeGame 至可操作棋盤的 smoke test
