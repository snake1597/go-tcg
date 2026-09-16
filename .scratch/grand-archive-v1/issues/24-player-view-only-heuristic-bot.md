# 24: 讓啟發式 bot 僅憑 Player View 完成決策

**What to build:** 建立唯一正式啟發式 bot，使其只使用自己的 Player View、合法行動與 Pending Choice，依固定優先級在 revision 有效期間提交決策。

**Blocked by:** 23: 解除完整固定牌組的 Support Set gate.

**Status:** completed

- [x] Bot 無法存取 Game State、對手手牌、牌庫順序、內部 ID 或 canonical replay。
- [x] Player View 的合法 action 與 Pending Choice 提供不透明 handle、可見描述及戰術 rank；決策依序考慮立即獲勝、避免立即落敗、有利攻擊、有效使用資源、出牌及推進 phase。
- [x] 相同 Player View 與 bot seed 產生相同選擇，只有評分相同時才消耗注入的 seeded randomness。
- [x] 每次決策與整場行動具有上限，能偵測無進展循環及安全處理過期提交。
