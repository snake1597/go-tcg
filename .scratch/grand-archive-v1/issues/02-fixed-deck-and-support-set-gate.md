# 02: 載入固定牌組並實施 Support Set 開局 gate

**What to build:** 讓 Game Module 載入唯一固定 Standard 牌組與不可變卡面定義，遞迴檢查完整 Support Set、Card Face、Ability Slot、機制及 typed operation，並在任何項目未支援時於開局前拒絕。

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

- [ ] Main Deck、Material Deck 與 Outside Game Pool 的數量、內容 ID、資料版本及鏡像一致性皆會驗證。
- [ ] Registry 拒絕非法階層式 ID、未知 parent、重複 slot、遺漏行為或孤立正式內容。
- [ ] Support Set closure 包含所有可達內容，正式內容只有 supported 或 unsupported 兩種狀態。
- [ ] Gate 失敗會完整列出缺少的內容、Ability Slot、機制、operation 或未裁定 issue。
