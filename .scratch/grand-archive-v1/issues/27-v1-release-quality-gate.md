# 27: 通過首版發布與文件 gate

**What to build:** 對完整真人對 bot 產品執行發布級確定性、安全性、收斂性與文件驗證，確保固定 Support Set 的所有可達互動均可信且可追溯。

**Blocked by:** 26: 完成真人對 bot 的完整鏡像對戰.

**Status:** ready-for-agent

- [ ] 至少 100 個不同 seed 的 bot 鏡像單局在行動上限內結束，沒有 panic、deadlock、非法提交、Needs Ruling 或 scheduler 不收斂。
- [ ] 全部單元、整合、race 與設定時限的 fuzz/property tests 通過，包含 rollback、handle 隔離、LKI、trigger ordering 與 replay hash。
- [ ] 固定牌組、Support Set、registry、Ability Slot、typed operation 與 rules issue 的一致性檢查全部通過。
- [ ] 領域文件、ADR、規則來源、卡牌覆蓋狀態、CLI 說明與文件連結均與實作同步。
