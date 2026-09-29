# 04 — 具名選擇與可序列化續行

**What to build:** Three of Hearts 的選牌與棄牌透過具名 binding 串接；玩家回答後從可序列化 Resolution Frame 確定地續行。

**Blocked by:** 02 — 首張 Go 可執行定義與編譯驗證。

**Related foundation:** 共通規則操作 05 — 具名選擇與卡牌篩選；該票負責所有現有選牌入口的共通化，本票只證明 Three of Hearts slice。

**Status:** completed

- [x] 只有 actor 看得見 Pending Choice；選取後的 reference 指向同一合法 Card Instance，只有允許時才能 pass。
- [x] 過期、跨玩家或偽造 handle 遭拒且不改變 state hash、PRNG、Knowledge State、event 或費用。
- [x] Continuation 不保存 closure 或可變 definition pointer；replay 結果一致，舊選牌分支已刪除。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。

2026-09-27：Three of Hearts 改由 `ability:1db8hz4prm:front:cardistry-draw-discard` 編譯為抽牌、`discard-card` 具名選牌與棄牌操作。等待輸入時，Game State 保存可序列化的 Resolution Frame，並在提交時重新確認選擇的 Card Instance 仍在合法來源區；過期、跨玩家、偽造及失效選項均無副作用。品質門檻：`go test ./...`、`go test -race ./...` 與兩個 10 秒 fuzz 測試。
