# 21: 以 Safeguard Amulet 與 Infernal Vessel 建立 Replacement Pipeline

**What to build:** 讓非戰鬥傷害 prevention 與 recover replacement 在 action 或 event 提交前由獨立 Replacement Pipeline 判定、排序、套用並重新計算候選。

**Blocked by:** 08: 以 Fiery Interference 支援 fast action 與恢復禁止; 11: 完成攻擊、Retaliation、傷害與死亡的戰鬥縱切; 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器.

**Status:** ready-for-agent

- [ ] Safeguard Amulet 的自我 banish 建立限定 Champion、非戰鬥傷害及 duration 的 delayed prevention。
- [ ] Infernal Vessel 每次套用後重新計算 recover amount，並正確處理零或負結果及 recover cost 的差異。
- [ ] 多個適用 replacement 由受影響內容的控制玩家排序；宣告中與已提交流程使用各自正確的 choice scope。
- [ ] 原始意圖、每次替代與最終 committed events 保留完整 cause chain 並可重播。
