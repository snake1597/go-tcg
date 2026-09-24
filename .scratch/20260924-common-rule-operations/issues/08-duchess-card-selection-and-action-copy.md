# 08 — Duchess 選牌與複製 Action

**What to build:** Duchess 選擇合格 Graveyard Action 後，透過共通複製操作放逐原牌並建立可選目標的獨立能力。

**Blocked by:** 04 — 場上 Object 放逐；05 — 具名選擇與卡牌篩選。

**Status:** ready-for-agent

- [ ] 合格 Action 由共通選牌條件決定；不合格或已離開 Graveyard 的卡不能被複製。
- [ ] 選定後原 Card Instance 進入 Banishment，複製能力不再次支付原 Action 費用，並具有獨立 Ability Instance 與正確來源資訊。
- [ ] 玩家能依規則選目標或略過；複製品在完成、略過及中斷後均被清理。
- [ ] Player View、Game Event、replay 與 state hash 驗證原牌及複製品生命週期。
- [ ] Duchess 專用選牌及複製 operation 已移除。
