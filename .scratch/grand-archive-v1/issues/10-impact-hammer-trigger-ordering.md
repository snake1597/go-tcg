# 10: 以 Impact Hammer 完成觸發收集與排序

**What to build:** 讓 On Wield 等 triggered abilities 從 committed Event Batch 被偵測及暫存，在規則 checkpoint 完成 state-based checks 後按 turn order 與玩家指定順序進入 Effects Stack。

**Blocked by:** 06: 以 Tonoris 完成 Champion Lineage 與 Materialization; 07: 以 Blazing Throw 建立第一條完整卡牌宣告與 Stack 路徑.

**Status:** ready-for-agent

- [ ] 每事件觸發與每批次聚合觸發保留正確 batch、cause、parent flow 與 simultaneous metadata。
- [ ] 單一合法順序自動推進；同玩家多個 triggers 時建立 Pending Choice 並等待其排序。
- [ ] mode 與必要 target 在入 Stack 時固定，depending-on 類結果留到結算時計算。
- [ ] Impact Hammer 的 On Wield 自傷即使來源之後離場仍依 Source Ref 與 LKI 正確結算。
