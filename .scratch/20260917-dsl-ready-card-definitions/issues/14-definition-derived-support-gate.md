# 14 — 由可執行定義產生 production 支援登錄

**What to build:** 建局時，只有成功編譯且依賴及裁定完整的 Ability Slots 才能標記 Supported。

**Blocked by:** 10 — 複製能力與 runtime 身分；11 — 遷移其餘 Cardistry Ability Slots；12 — 遷移其餘 Action Ability Slots；13 — 遷移其餘 Object 與 Regalia Ability Slots。

**Status:** blocked

**Blocked reason:** DSL-ready Card Definitions 10 與 13 尚未完成；完成所有 Ability Slot 遷移前，不能宣告 production support metadata 已具有單一可執行來源。

**Next implementation step:** 10 與 13 完成後，先列出仍由手寫 registry 提供的 Face、Slot、mechanism、operation、dependency 與 ruling metadata，再逐項改由 compiled definitions 衍生；每移除一個舊來源都保持 Support Set 可建立，不同時設計 DSL schema。

- [ ] production metadata 從可執行定義衍生，並驗證 Face、Slot、mechanism、operation、content dependency 及 ruling reference。
- [ ] 重複 ID、缺漏或循環依賴、未知原語、Unsupported 或 Needs Ruling 內容均在建局前拒絕。
- [ ] 已取代的手寫 registry shape 與平行 Card ID 路徑已移除；固定 Support Set 可完整遊玩。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
