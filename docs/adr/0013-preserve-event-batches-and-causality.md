# 保存規則事件的批次與因果關係

引擎內部以 `EventBatch` 表示由同一規則指令產生的一組離散 `GameEvent`，並保留 batch ID、cause ID、父流程、確定順序與 simultaneous flag。一次「抽三張」仍產生三個 draw event，但三者共享批次身分；同時傷害也必須維持共同批次，不能過早拆成互不相關的狀態修改。

Triggered abilities 在事件發生時先偵測並暫存，只在規則指定的 checkpoint 整理並放入 Effects Stack。`whenever X` 可以逐事件觸發，`whenever one or more X` 則依批次聚合；ResolutionFrame 明確表示效果指令與 checkpoint 邊界，因此不會在每個事件或每句效果文字之後擅自執行 state-based checks。

CLI 仍可將 committed events 投影成扁平且易讀的顯示列表，但該列表不是規則真相來源。Canonical state、replay 診斷與規則測試保留執行所需的批次、因果及順序 metadata，以支援觸發聚合、同時事件、On Hit／On Kill 判定及確定性重播。

## 曾考慮的方案

只保存扁平的 `[]GameEvent` 型別較少，也容易直接輸出，但會遺失事件是否同源、同時或屬於同一聚合範圍等規則資訊。日後只能從事件順序猜測批次，不足以可靠判定 `whenever one or more`、同時傷害與 trigger flush 時點，因此不採用。
