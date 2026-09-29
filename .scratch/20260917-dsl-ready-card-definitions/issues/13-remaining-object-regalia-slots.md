# 13 — 遷移其餘 Object 與 Regalia Ability Slots

**What to build:** 固定 Support Set 中剩餘 Object、Weapon 與 Regalia 的 activated、triggered、static、permission 及 replacement 能力皆由 Card Definition 提供。

**Blocked by:** 05 — Activated ability 的費用與使用限制；06 — Triggered ability 定義；07 — Static 與 permission 定義；08 — Replacement ability 定義；09 — 替代費用的可執行定義；10 — 複製能力與 runtime 身分。

**Execution order:** 10 完成後才開始本票。進入本票後一次只遷移一個 Ability Slot，完成 Definition、舊路徑刪除、scenario 與 quality gate 後才開始下一個 Slot。

**Related foundation:** 共通規則操作 04 — Field Object banish；06 — sacrifice；09 — Reserve／Memory payment；10 — alternative／Wield payment。某個 Slot 需要其中尚未具備的 primitive 時，先在對應 Common issue 交付最小 operation slice並記錄 `Delivered slice`，不等待整張 umbrella issue 完成。

**Status:** blocked

**Blocked reason:** 10 — Duchess runtime copy 尚有未結驗收；先關閉目前進行中的 migration slice，再開啟剩餘 Object／Regalia 批次。

- [ ] 每個 Slot 有穩定 ID、規則來源與 scenario，涵蓋正常結算、非法宣告、來源失效及跨機制互動。
- [ ] Object 能力共用宣告生命週期；static 只輸入中央 evaluator，replacement 只輸入 Replacement Pipeline。
- [ ] 每個 Slot 都記錄使用的共通 primitive；缺少的 primitive 已先以 mechanism-level operation 實作並同步更新對應 Common issue，沒有卡牌專用 direct mutation 或暫時 API。
- [ ] 已遷移 Slot 的 Card ID 路徑全部刪除，不留無法驗證的 Supported Slot。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
