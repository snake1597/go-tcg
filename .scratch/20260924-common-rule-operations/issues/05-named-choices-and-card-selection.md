# 05 — 具名選擇與卡牌篩選

**What to build:** 玩家在能力結算中選擇 Card 或 Object 後，後續操作可用具名參照取得結果，且選項由共通條件產生。

**Blocked by:** None — can start independently from zone mutation work.

**Related migration:** DSL-ready Card Definitions 04 — Three of Hearts 使用本票能力完成卡牌 Definition slice，但不是本票的完成 owner。

**Status:** in-progress

**Delivered slice:** DSL-ready Card Definitions 04 已以 Three of Hearts 證明 Hand 選擇、具名 binding、可序列化 Resolution Frame、私有 View Handle、提交時重驗證及 replay 確定性。剩餘工作是完成 Memory、Graveyard 與 Object 選擇入口的共通化及移除其專用分支。

**Next implementation step:** 先以一個現有 Memory 選牌能力擴充 selector 所需的最小欄位並重用既有 Resolution Frame；完成提交時重驗證與隱藏資訊 scenario 後，再處理 Graveyard，最後處理 Object choice。每增加一種 selector 條件都必須由當前 Support Set 案例證明需要。

- [ ] 現有 Hand、Memory 與 Graveyard 選牌案例可用共通的區域、owner/controller、類型、子類型、元素、費用及排除來源條件表達；只加入已用到的條件。
- [ ] 選擇結果保存在可序列化的 Resolution Frame，後續操作可引用不同的具名結果，不需覆寫單一 Target。
- [ ] Pending Choice 只向指定玩家暴露依法可見的 View Handle；可略過、無選項及續行行為明確。
- [ ] 提交時重新驗證選項，過期或跨玩家 Handle 被拒絕且權威 Game State 不改變。
- [ ] 選擇前後的 replay 與 state hash 確定；共通選擇 runtime 不再需要卡牌專用 selector 或 continuation 分支。
