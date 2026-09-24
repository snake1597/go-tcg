# 05 — 具名選擇與卡牌篩選

**What to build:** 玩家在能力結算中選擇 Card 或 Object 後，後續操作可用具名參照取得結果，且選項由共通條件產生。

**Blocked by:** 01 — 棄牌與一般區域移動。

**Status:** ready-for-agent

- [ ] 現有 Hand、Memory 與 Graveyard 選牌案例可用共通的區域、owner/controller、類型、子類型、元素、費用及排除來源條件表達；只加入已用到的條件。
- [ ] 選擇結果保存在可序列化的 Resolution Frame，後續操作可引用不同的具名結果，不需覆寫單一 Target。
- [ ] Pending Choice 只向指定玩家暴露依法可見的 View Handle；可略過、無選項及續行行為明確。
- [ ] 提交時重新驗證選項，過期或跨玩家 Handle 被拒絕且權威 Game State 不改變。
- [ ] 選擇前後的 replay 與 state hash 確定；被取代的選牌專用 operation 分支已移除。
