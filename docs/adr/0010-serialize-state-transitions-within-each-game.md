# 每個單局內序列化狀態轉移

Game Module 一次只處理一個玩家行動或待決選擇，不自行啟動 goroutine，也不允許多個操控端同時修改同一單局。對手在回合外的互動仍依 Opportunity 與效果堆疊形成明確的交錯序列；規則定義的同時事件由一次原子狀態轉移表示，私密的同時選擇則可依序收集但必須等所有玩家完成後才共同公開及套用。

每次有效輸入提交後，Game Module 的 deterministic scheduler 自動執行所有不需要外部決定的強制流程，直到下一個穩定停點：某位玩家取得 Opportunity、指定玩家必須回答 PendingChoice、單局結束，或 NeedsRuling。暫停中的流程以型別安全、可 hash 且可重現的 `ResolutionFrame` 保存在 GameState；不使用 Go closure 保存 continuation。恢復時先驗證 state revision，再從 frame 確定性地繼續。

PendingChoice 的回答者不一定是回合玩家或 Opportunity 持有者。等待期間一般規則行動不能插入，但 Concede 是不受 Opportunity 或 timing restriction 限制的全域特殊輸入，因此任何尚未落敗的玩家仍可提交；它與其他輸入一樣由引擎逐一處理，不代表並行修改狀態。未來伺服器應在外層以每單局 mailbox 或 event loop 排序輸入；bot 可在外層非同步思考，但提交時必須攜帶 state revision，過期決策會被拒絕。

在規則 checkpoint，scheduler 依官方順序反覆執行 state-based checks，直到完整一輪不再改變狀態。每輪使用一致狀態視圖，依法同時發生的結果形成同一 EventBatch；任何變化都使檢查從規定起點重啟。期間產生的 triggered abilities 先暫存，fixed point 完成後才整理。需要 Unique 等玩家決定時建立 PendingChoice，回答後從保存的 ResolutionFrame 恢復並重新檢查。Fixed point 完成前不移交 Opportunity，也不結算下一個 StackItem。

每輪開始先由中央 evaluator 取得該 canonical state 的 derived-characteristic view；同一輪共用這份結果。若 state-based result 改變效果來源或基礎狀態，下一輪重新求值。Derived view 不能被獨立修改或當成第二份 GameState。

卡牌 handler 不能自行呼叫或跳過 state-based checks。引擎以確定性的循環偵測及迭代上限防止無法收斂的支援內容卡死程序；觸發上限時停止單局並輸出診斷、replay 與 state hash，不任意選擇結果。

Fixed point 後要 flush 的 triggered abilities 依 turn order 從回合玩家開始分組放入 Effects Stack。同一玩家控制多個 triggers 時，由該玩家透過 PendingChoice 決定自己的順序；只有一種合法順序時自動前進。「Choose one」模式與必要目標在 trigger 放入 Stack 時選定，「depending on」模式在結算時計算。所有 triggers 都完成放置後，才授予規則指定的 Opportunity。排序、模式及目標選擇都屬正式 replay 輸入。

## 曾考慮的方案

讓不同玩家或效果以 goroutine 平行修改單局狀態看似能直接表達回合外互動與同時事件，但會引入競態、非確定順序與難以重播的結果；遊戲規則本身已提供 Opportunity、回合順序與同時結算語意，不需要共享狀態並行寫入。另一方案是由 CLI 逐步呼叫每個內部規則動作；這會把結算狀態機洩漏至 Adapter，造成 CLI、bot 與未來伺服器各自實作不同的推進邏輯，因此不採用。
