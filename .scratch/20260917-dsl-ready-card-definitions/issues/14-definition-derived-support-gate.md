# 14 — 由可執行定義產生 production 支援登錄

**What to build:** 建局時，只有成功編譯且依賴及裁定完整的 Ability Slots 才能標記 Supported。

**Blocked by:** 11 — 遷移其餘 Cardistry Ability Slots；12 — 遷移其餘 Action Ability Slots；13 — 遷移其餘 Object 與 Regalia Ability Slots。

**Status:** ready-for-agent

- [ ] production metadata 從可執行定義衍生，並驗證 Face、Slot、mechanism、operation、content dependency 及 ruling reference。
- [ ] 重複 ID、缺漏或循環依賴、未知原語、Unsupported 或 Needs Ruling 內容均在建局前拒絕。
- [ ] 已取代的手寫 registry shape 與平行 Card ID 路徑已移除；固定 Support Set 可完整遊玩。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
