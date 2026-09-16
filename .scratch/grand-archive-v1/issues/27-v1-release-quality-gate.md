# 27: 通過首版發布與文件 gate

**What to build:** 對完整真人對 bot 產品執行發布級確定性、安全性、收斂性與文件驗證，確保固定 Support Set 的所有可達互動均可信且可追溯。

**Blocked by:** 26: 完成真人對 bot 的完整鏡像對戰.

**Status:** completed

- [x] 100-seed bot 鏡像發布 gate 已由產品負責人核准延後；本次 ticket 驗收不執行且不可視為通過。
- [x] 其餘單元、整合、race 與兩組設定時限 fuzz/property tests 通過，涵蓋 rollback、handle 隔離、LKI、trigger ordering 與 replay hash。
- [x] 固定牌組、Support Set、registry、Ability Slot、typed operation 與 rules issue 的一致性檢查通過。
- [x] 領域文件、ADR、規則來源、卡牌覆蓋狀態、CLI 說明與文件連結已與實作同步。

## 已知未完成項目

以下項目是 issue 26 完成端到端路徑後仍需處理的發布工作。前四項是已確認的工程缺口；「待規則確認」只能依官方規則與既有 ruling 處理，不得直接猜測實作。

### 執行期與輸入契約

- [x] production CLI 的 bot 輪替迴圈具整局提交上限；超限會停止並輸出診斷、replay 與 state hash。
- [x] `Input` payload 依 action kind 驗證，多餘的 `Reserve`、`FloatingMemory`、target 或 choice 會被拒絕且 state hash 不變。
- [x] 零 Reserve 費用的 bot 不附帶 `Input.Reserve`；已覆蓋 Verita alternative cost。
- [x] `Input.Reserve` 變更已將 replay 與 canonical state schema 升至 v4，未保留舊格式相容路徑。

### 端到端與發布測試

- [x] 鏡像對戰的 Stack 回應只在 Stack 非空且提交合法非 pass action 時才計入覆蓋。
- [x] production CLI 的公開 `ReadReserve` 直接測試覆蓋正確張數、重複、越界、格式錯誤與 EOF；rejected Reserve input 維持 state hash。
- [x] 100-seed 批次鏡像測試已建立，具每局上限與 seed、step、診斷、state hash 失敗資訊；依產品負責人核准延後執行。
- [x] 除延後的 100-seed 鏡像 package test 外，單元、race 與兩組設定時限 fuzz/property suites 已執行；不得將此記錄解讀為完整 `go test ./...` 已通過。
- [x] setup 測試已收緊對合法 action 的斷言，保留動態卡牌 action 並拒絕意外開放的 action。

### 待規則確認

- [x] Wake Up Phase 已確認並由 scheduler 同時喚醒 Champion 與所有合格 rested Ally，且有回合測試。
- [x] `legalAttackers` 已接受所有 obey 且可攻擊的 Ally；卡牌特例與一般規則責任邊界已記錄於規則文件。
- [x] play Ally 已統一經 Effects Stack，允許 fast response 後才進場；直接從 Hand 放入 Field 的特殊路徑已移除。

### Review 與文件同步

- [x] 已完成 Standards 與 Spec 雙軸 Code Review，處理發現後重審至無 HIGH／CRITICAL 阻擋項目。
- [x] CLI 使用說明、replay/schema v4 文件、rules issue／ADR 與卡牌覆蓋狀態已同步。

## 完成判定

本次驗收依產品負責人核准，將 100-seed 鏡像執行 gate 延後；該 gate 保留為後續發布前必須執行的工作，不可視為已通過。其餘已知未完成項目、相關測試與雙軸 Review 均已完成，因此本 issue 狀態為 `completed`。
