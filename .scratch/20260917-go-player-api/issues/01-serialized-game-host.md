# 01 — 建立序列化 Game Host

**What to build:** 建立 Game Module 外的遊戲應用層，讓每場 practice game 的 human、Bot 與內部排程輸入都經過同一條序列化 mailbox，並可用固定 seed 重現局面。

**Blocked by:** None — can start immediately

**Status:** ready-for-agent

- [ ] 可建立、定位及結束一場固定設定的 practice game
- [ ] 同一 game 的所有狀態轉移依序執行，不會並行改寫 canonical state
- [ ] Bot 只使用自身 PlayerView 與 Legal Actions，並透過同一 mailbox 提交輸入
- [ ] PendingChoice、NeedsRuling 或遊戲結束時，Bot 不會繼續提交無效輸入
- [ ] 固定 seed 的整合測試可重現相同 revision 與結果
