# 02 — 建立本機 Create／Watch 垂直切片

**What to build:** 提供可由 ConnectRPC client 建立練習對局並觀看首份完整 PlayerView 的最小端對端路徑，同時建立 seat-scoped PlayerSession 與新的穩定卡牌／物件身分模型。

**Blocked by:** 01 — 建立序列化 Game Host

**Status:** ready-for-agent

- [ ] CreatePracticeGame 回傳 game id、只出現一次的 human token 與 card data version
- [ ] server 只保存 PlayerSession token 的安全 hash，並由 token 推導 game 與 seat
- [ ] WatchGame 連線後先傳送該座位的完整 PlayerView snapshot
- [ ] PlayerView 使用 CardRef、ObjectRef 與明確 ViewHandle，不再以卡名作為身分
- [ ] client 無法藉由 request 欄位切換玩家視角
- [ ] 被新合約取代的名稱式與重複 PlayerView 欄位已移除
- [ ] generated ConnectRPC client 經真實 HTTP handler 與 Game Module 的 contract test 通過
