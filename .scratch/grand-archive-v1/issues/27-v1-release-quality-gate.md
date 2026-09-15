# 27: 通過首版發布與文件 gate

**What to build:** 對完整真人對 bot 產品執行發布級確定性、安全性、收斂性與文件驗證，確保固定 Support Set 的所有可達互動均可信且可追溯。

**Blocked by:** 26: 完成真人對 bot 的完整鏡像對戰.

**Status:** ready-for-agent

- [ ] 至少 100 個不同 seed 的 bot 鏡像單局在行動上限內結束，沒有 panic、deadlock、非法提交、Needs Ruling 或 scheduler 不收斂。
- [ ] 全部單元、整合、race 與設定時限的 fuzz/property tests 通過，包含 rollback、handle 隔離、LKI、trigger ordering 與 replay hash。
- [ ] 固定牌組、Support Set、registry、Ability Slot、typed operation 與 rules issue 的一致性檢查全部通過。
- [ ] 領域文件、ADR、規則來源、卡牌覆蓋狀態、CLI 說明與文件連結均與實作同步。

## 已知未完成項目

以下項目是 issue 26 完成端到端路徑後仍需處理的發布工作。前四項是已確認的工程缺口；「待規則確認」只能依官方規則與既有 ruling 處理，不得直接猜測實作。

### 執行期與輸入契約

- [ ] 在 production CLI 的 bot 輪替迴圈加入整局提交上限或等價的無進展偵測；超限時必須停止單局並輸出診斷、replay 與 state hash，不得無限 `Decide`／`Submit`。
- [ ] 按 action kind 驗證 `Input` payload，只允許該行動實際使用的 `Reserve`、`FloatingMemory`、target 或 choice 欄位，拒絕被忽略的多餘輸入且保持 state hash 不變。
- [ ] 修正並測試零 Reserve 費用的 bot 決策：即使 `LegalAction.ReserveOptions` 非空，也不得附帶任何 `Input.Reserve` handle；特別覆蓋 Verita alternative cost。
- [ ] 核對新增 `Input.Reserve` 與 canonical state/replay 結構是否改變既有版本契約；若有改變，直接更新 replay/state schema 版本與測試，不保留舊格式相容路徑。

### 端到端與發布測試

- [ ] 將鏡像對戰的 Stack 回應覆蓋收緊為「Stack 非空時提交合法的非 pass 回應」，不能把任意 action 或單純讓過計為已覆蓋。
- [ ] 為 production CLI 的 Reserve 編號選單加入直接測試，涵蓋正確張數、重複選擇、越界、格式錯誤與 EOF，並驗證 rejected input 不改變 state hash。
- [ ] 建立 100 個不同 seed 的批次鏡像測試；每局使用明確行動上限，失敗訊息至少包含 seed、step、診斷與最終 state hash，讓問題可重現。
- [ ] 執行 `go test ./...`、`go test -race ./...` 與設定時限的 fuzz/property suites，保存失敗 seed；不得以單一 seed 的兩次成功取代發布 gate。
- [ ] 收緊 setup 測試對合法 action 的斷言：保留動態卡牌 action 的彈性，但不得只檢查「包含預期 kind」而漏掉意外開放的行動。

### 待規則確認

- [ ] 依官方 Wake Up Phase 規則確認 rested Ally 是否與 Champion 一同 wake up；若需要，統一由 scheduler 喚醒所有合格 object 並補回合測試。
- [ ] 依 Ally 攻擊規則確認 `legalAttackers` 是否應接受所有 obey 且可攻擊的 Ally，而不是只接受 Red Hare；把卡牌特例與一般規則的責任邊界寫入規則文件。
- [ ] 依 play Ally 的正式流程確認 Ally 是否必須先成為 Stack item、允許 fast response 後才進場；若是，移除直接從 Hand 放入 Field 的特殊路徑並共用 declaration/runtime。

### Review 與文件同步

- [ ] 對 commit `3271466`（fixed point `6c27230`）完成 Standards 與 Spec 雙軸 Code Review，處理所有 HIGH／CRITICAL 發現後重新審查至無阻擋項目。先前的自動 review worker 因執行額度限制未產出報告，因此此項仍未通過。
- [ ] 更新 CLI 使用說明、replay/schema 版本文件、rules issue／ADR 與卡牌覆蓋狀態，使其與最終裁定及測試結果一致。

## 完成判定

只有本 issue 頂層四項驗收及以上所有已知未完成項目均勾選，且相關測試與雙軸 Review 通過後，狀態才能改為 `completed`。若規則證據不足，依 ADR 0011 保持 `blocked`，不可用臨時行為通過發布 gate。
