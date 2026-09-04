# 25: 提供 production CLI 編號選單

**What to build:** 讓真人以編號選單啟動並操作固定牌組 Standard 單局，查看依法可見的狀態、事件與合法選擇，並輸出可驗證的私人 canonical replay。

**Blocked by:** 23: 解除完整固定牌組的 Support Set gate.

**Status:** ready-for-agent

- [ ] 啟動接受 seed 與 replay 輸出位置，開局 gate 失敗時顯示可行動的缺項診斷。
- [ ] 畫面顯示回合、phase、Opportunity、Champion、自己手牌、Field、Effects Stack、最近事件與結果。
- [ ] 行動、mode、target、費用、trigger ordering 與其他選擇均使用引擎提供的編號選項，不重算規則。
- [ ] 非法／過期輸入、EOF、Needs Ruling 或 scheduler failure 不留下半提交交易；replay 輸出附隱私警告。
