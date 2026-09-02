# 測試策略

本專案採用規則案例先行的 test-first 流程。測試描述呼叫端可觀察的行為，並以鎖定的官方規則快照、卡面資料與已知結果作為獨立真相來源。

## 已確認的測試 seam

### Game Module Interface

大多數測試由正式 Game Module Interface 驗證完整規則行為、非法操作、玩家視角、確定性與狀態 hash。測試不得直接存取或 mock 內部卡牌 handler、效果元件及私有函式。

玩家視角測試至少涵蓋：不同玩家取得不同 KnowledgeState 投影；內部 ID 不出現在 PlayerView；可追蹤期間 ViewHandle 保持可用；洗牌或隨機化後舊 handle 被撤銷且無法關聯新位置；已 reveal 資訊仍可從可見事件歷史回顧；過期、跨玩家或無權限 handle 的提交不改變狀態。

### Bot Controller Interface

給定玩家視角與合法選項，驗證 bot 回傳合法且符合既定優先級的選擇。測試不得依賴完整隱藏狀態，也不得 mock Game Module 內部 Implementation。

### CLI Process Interface

透過 stdin/stdout 執行少量端到端 smoke tests，驗證真人輸入、bot 回合、完整結束與 replay 輸出可以正確串接。CLI 測試不得重複整套規則案例。

未來加入 PostgreSQL 時，先與專案負責人確認新的持久化 use case seam，再以真實測試資料庫優先驗證，不 mock GORM。

## 每個 vertical slice 的流程

1. 標註規則來源、規則 commit 與卡面資料版本。
2. 只寫一個會失敗的行為測試（red）。
3. 實作使該測試通過所需的最小行為（green）。
4. 使用正式 Interface 驗證結果，不越過 seam 檢查內部狀態。
5. 需要時加入 replay 與狀態 hash 回歸案例。
6. 完成該 slice 後才開始下一個測試；不先橫向寫完所有測試。
7. 重構留待 review 階段，不混入 red → green 循環。

## 額外驗證

- 將官方規則例子轉為 table-driven scenario tests。
- 每張支援卡牌涵蓋正常、非法、目標失效與相關互動案例。
- 以 fuzz/property tests 驗證卡牌守恆、行動權限與相同 seed 加相同輸入必須得到相同 hash 等不變量。
- 驗證每次提交都自動推進至穩定停點；ResolutionFrame 可確定性暫停及恢復；PendingChoice 只接受指定玩家回答，但等待期間任何尚未落敗的玩家都能投降。
- 以「抽多張」、「whenever」與「whenever one or more」、同時傷害等規則案例驗證 EventBatch 的批次、因果、順序及 trigger flush checkpoint；CLI 事件投影不得成為規則判定輸入。
- 驗證 state-based checks 依官方順序反覆收斂、同時結果共用 EventBatch、需要玩家決定時可暫停及恢復，且 fixed point 完成前不移交 Opportunity；另以刻意循環的 test fixture 驗證診斷與停止行為。
- 驗證 DeclarationTransaction 的成功提交與完整 rollback：取消、過期、費用不足及最終非法都不能改變正式 zone、cost、trigger buffer、KnowledgeState 或 PRNG cursor；合法 activation 後再 fizzle／negate 則不得退費。另驗證未提交或 rollback 的 chance outcome 不會出現在 PlayerView／事件歷史，取樣開始後也不提供規則外的任意取消。
- 驗證 Effects Stack 的來源卡與 StackItem 分離：一張來源卡可關聯原始 instance 與多個 copy；只在最後一個 item 離開後移動來源卡；來源卡先離開時相關 card-play instances fizzle；ability item 在來源離場後仍可依 LKI 結算。
- 驗證同時 triggers 依 turn order 分組、同一玩家可決定自己的放置順序、模式與目標在正確時點固定，且全部放置完之前不授予 Opportunity。
- 對 Support Set 用到的 continuous effects 建立跨 Layer A–E、timestamp、dependency 與 dependency loop 案例；驗證同一 state-based round 共用 derived view、基礎狀態改變後下一輪重新求值。對 replacements 驗證 draft-scoped 與 resolution-scoped 玩家排序、每步重新計算候選、prevention 與 cause chain。所有合法性、費用、戰鬥及 PlayerView assertion 必須觀察同一 derived result。
- 正式 Standard setup 測試必須使用固定牌組的 Level 0 Champion On Enter ability 取得起始手牌；test-only fixture 的直接發牌不能成為正式 registry 路徑。
- 每個涉及規則歧義的案例都必須引用 [`rules-issues.md`](./rules-issues.md) 中已解決的項目；未解決或不適用 Support Set 的項目不得以猜測 expected value 建立測試。
- 只在真正的系統 seam 替換時間、隨機來源、檔案系統或外部儲存；不 mock 自有的內部 Module。

## 禁止模式

- 測試私有函式、內部呼叫次數或呼叫順序。
- 用與 Implementation 相同的演算法重新計算 expected value。
- 透過資料庫或其他旁路驗證正式 Interface 的結果。
- 為尚未開始的多個 slice 預先建立大量測試骨架。
