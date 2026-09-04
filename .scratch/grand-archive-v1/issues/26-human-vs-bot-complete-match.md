# 26: 完成真人對 bot 的完整鏡像對戰

**What to build:** 將 production CLI、Game Module 與啟發式 bot 組成單一執行檔，讓真人使用固定鏡像牌組從 Standard setup 玩到可信的單局結果。

**Blocked by:** 24: 讓啟發式 bot 僅憑 Player View 完成決策; 25: 提供 production CLI 編號選單.

**Status:** ready-for-agent

- [ ] 真人與 bot 從各自 Player View 輪流提交行動，adapter 不直接推進 scheduler 或修改 Game State。
- [ ] 單局能由 Champion 死亡、牌庫耗盡或任一玩家投降結束，結束後拒絕一般行動。
- [ ] 一場涵蓋打牌、Cardistry、戰鬥、Pending Choice 與 Stack 回應的鏡像對戰可完整完成。
- [ ] 相同版本、seed 與記錄輸入重播時，每一步皆得到相同 state hash 與最終結果。
