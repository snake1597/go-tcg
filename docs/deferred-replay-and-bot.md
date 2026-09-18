# Replay 與 Bot 策略訓練待續議題

## 狀態

兩項主題均延後，尚未形成實作規格或架構決策，因此不放在 `docs/specs`。Vue 與 Go Player API 首版不得預留猜測性的 Replay 或訓練介面。

已確認：取消人類即時介入、批准、覆寫或接管 Bot 行動的設計。改善方向是訓練或調整 Bot 策略，使其更接近正常玩家並提供更好的練習對局。

## Replay 待討論

1. Replay 僅供觀看、研究回溯，或允許從歷史點建立新對局。
2. 可定位至 action、turn、phase 或 EventBatch 的哪個層級。
3. checkpoint、狀態 snapshot 與重新模擬 input 的分工。
4. 玩家視角隱藏資訊與結束後完整揭露的權限。
5. PRNG、Game Module、rules snapshot、API schema 與 card data version 的釘選。
6. 進行中與已結束對局的可回看範圍。
7. 儲存、保留期限、匯出與完整性驗證。
8. timeline、播放速度、事件跳轉與 Card Inspector。
9. 舊引擎版本不可使用時的處理。
10. 從 Replay branch 是否建立新的 canonical game，以及起始證明。

觀看、seek、branch 與 undo 是不同能力；不得預設「回溯」等於撤銷正式對局。

## Bot 現況與邊界

目前 Bot 是固定優先級 heuristic，從自身 PlayerView 的 Legal Actions 或 PendingChoice 選擇候選，同順位時使用注入的亂數來源。部分戰術排序目前仍位於 Game Module，尚無訓練資料、模型版本、self-play pipeline、評測或發布機制。

Bot 必須持續遵守：

- 只讀取自己的 PlayerView。
- 只從引擎提供的合法 action／choice 中選擇。
- 透過與 human 相同的 Game Host mailbox 提交。
- 訓練隨機性不得破壞正式對局的 deterministic replay。

## Bot 待討論

1. 勝率、操作自然度、難度分級與練習價值的優先順序及量測方式。
2. 訓練與評估資料來源、授權與隱私邊界。
3. heuristic search、模仿學習、self-play reinforcement learning 或混合策略。
4. 不完全資訊與跨決策 observation 的編碼。
5. 模型輸出如何對應可變長 Legal Actions、payment 與 PendingChoice。
6. 如何避免對固定牌組與少量對手過度擬合。
7. 固定 seeds、基準 heuristic、歷史模型與真人風格的評測 gate。
8. 模型版本、發布、回退、A／B 評估與 replay 重現契約。
9. 戰術排序如何從 Game Module 移至 Bot／policy 邊界。
10. Replay 是否保存 observation、候選 action、機率、value 與結果資料。
11. 訓練基礎設施與線上推論的技術棧、延遲及資源預算。

## 恢復討論方式

- Replay 與 Bot 分開建立規格，不在同一輪混合決策。
- Replay 先確認觀看、seek、branch、undo 的產品邊界。
- Bot 先定義「更符合正常玩家」的可量測成功條件。
- 不重新詢問已完成的 Vue／PlayerView／ConnectRPC 決策，也不得恢復人類即時介入方案。
