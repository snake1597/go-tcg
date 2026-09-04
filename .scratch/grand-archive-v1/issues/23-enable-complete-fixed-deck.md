# 23: 解除完整固定牌組的 Support Set gate

**What to build:** 將所有已驗證卡面行為、Ability Slot、機制與 typed operation 登錄為完整支援，使唯一固定牌組首次能通過 Support Set gate 並建立正式 Standard 單局。

**Blocked by:** 02: 載入固定牌組並實施 Support Set 開局 gate; 04: 以 Spirit of Fire 完成 Standard 開局; 06: 以 Tonoris 完成 Champion Lineage 與 Materialization; 07: 以 Blazing Throw 建立第一條完整卡牌宣告與 Stack 路徑; 08: 以 Fiery Interference 支援 fast action 與恢復禁止; 09: 以 Straight Flare 建立 Suited 查詢與動態傷害; 10: 以 Impact Hammer 完成觸發收集與排序; 11: 完成攻擊、Retaliation、傷害與死亡的戰鬥縱切; 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器; 13: 建立 Cardistry 核心並支援基本 Cardistry 卡牌; 14: 支援 Cardistry 的選擇、棄牌與部署互動; 15: 以 Noire 與 Rouge 支援 Suited threshold 行為; 16: 以 The Duchess's Thornes 支援 Cardistry observer; 17: 以 Duchess 支援 runtime action copy; 18: 以 Heated Vengeance 與 Red Hare 擴充攻擊狀態; 19: 以 Smoke Bombs 與 Trump Set 完成戰鬥重定向; 20: 以 Verita 與 Peppered Chef 支援複合費用和 On Death; 21: 以 Safeguard Amulet 與 Infernal Vessel 建立 Replacement Pipeline; 22: 完成固定牌組剩餘 Regalia 與不可啟動能力.

**Status:** ready-for-agent

- [ ] 32 種 Card Face 與 49 個 Ability Slot 均恰好對應一個正式、已測試的 registry entry。
- [ ] Support Set closure、registry 與牌組 manifest 雙向一致，沒有遺漏、孤立或部分支援內容。
- [ ] 所有 required mechanism、typed operation 與受影響 rules issue 均已支援或取得記錄完整的裁定。
- [ ] 完整固定牌組可建立單局；移除任一必要支援項時會在開局前以具體診斷拒絕。
