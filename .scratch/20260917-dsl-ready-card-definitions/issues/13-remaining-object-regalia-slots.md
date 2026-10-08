# 13 — 遷移其餘 Object 與 Regalia Ability Slots

**What to build:** 固定 Support Set 中剩餘 Object、Weapon 與 Regalia 的 activated、triggered、static、permission 及 replacement 能力皆由 Card Definition 提供。

**Blocked by:** 05 — Activated ability 的費用與使用限制；06 — Triggered ability 定義；07 — Static 與 permission 定義；08 — Replacement ability 定義；09 — 替代費用的可執行定義；10 — 複製能力與 runtime 身分。

**Execution order:** 10 完成後才開始本票。進入本票後一次只遷移一個 Ability Slot，完成 Definition、舊路徑刪除、scenario 與 quality gate 後才開始下一個 Slot。

**Related foundation:** 共通規則操作 04 — Field Object banish；06 — sacrifice；09 — Reserve／Memory payment；10 — alternative／Wield payment。某個 Slot 需要其中尚未具備的 primitive 時，先在對應 Common issue 交付最小 operation slice並記錄 `Delivered slice`，不等待整張 umbrella issue 完成。

**Status:** completed

**Completed slices:**

- Grand Crusader's Ring 的 `ability:2gv7DC0KID:front:banish-draw`。Definition 編譯為放逐來源與抽牌的 typed operations，正式 Player View scenario 驗證宣告、放逐、Effects Stack、結算與 stable slot metadata；已刪除其 Object ability 的 Card ID 分流。
- Water Resonance Bauble 的 `ability:dSSRtNnPtw:front:banish-draw-water-champion`。Definition 的對手 Water Champion 條件決定合法宣告，正式 Player View scenario 驗證條件、放逐、Effects Stack 與抽牌；已刪除其 Card ID 分流。
- Wind Resonance Bauble 的 `ability:bHGUNMFLg9:front:banish-draw-wind-champion`。Definition 使用相同共通條件 primitive；編譯與無效條件 scenario 已驗證，已刪除其 Card ID 分流。
- Smoke Bombs 的 `ability:ScGcOmkoQt:front:banish-stealth-draw`。Definition 提供放逐來源、任一 Ally 選擇、至回合結束的 stealth 與抽牌；正式 Player View scenario 驗證來源費用、Effects Stack、選擇與攻擊互動，已刪除 Card ID 宣告及直接效果 mutation。
- Safeguard Amulet 的 `ability:yj2rJBREH8:front:banish-prevent-non-combat-damage`。Definition 提供放逐來源與 delayed non-combat damage prevention operation；正式 Player View/replay scenario 驗證 Effects Stack 結算及 Replacement Pipeline cause chain，已刪除 Card ID 宣告及 replacement 建立路徑。
- Impact Hammer 的 `ability:chsbalegbs:front:class-bonus` 與 `ability:chsbalegbs:front:on-wield-self-damage`。前者只向中央 static evaluator 輸出 class-matched `PowerDelta`；後者以 Wield event、event target、damage operation 與 Source LKI 建立 trigger。compiler、中央 evaluator、正式 Player View、來源離場與 replay scenarios 覆蓋兩個 Slot。
- Bulwark Sword 的 `ability:8kmoi0a5uh:front:class-bonus-wield-cost`。同一 static Slot 向中央 evaluator 輸出 class-matched `PowerDelta: 1` 與 `ReserveCostDelta: 2`，共用 Wield query／payment 消費 derived cost；不足付款與成功放逐 Reserve scenarios 通過，舊 Card ID evaluator 分支已刪除。
- Viridian Protective Trinket 的 `ability:s3572j3oda:front:opponent-water-activation-tax`。permission 定義輸入共用 activation Reserve cost query，驗證 Water、對手、控制者回合與來源存續；非適用玩家／回合及來源離場 scenarios 不加費，舊 Card ID query 已刪除。
- Infernal Vessel 的 `ability:vgWgu1DUYv:front:recover-reduce`。replacement 定義只由 Replacement Pipeline 收集，正常重算、多 replacement 選序、recover cost 排除與來源離場 scenarios 通過。
- The Duchess's Thornes 的 `ability:bEXmm4rKOs:front:hindered`、`ability:bEXmm4rKOs:front:ally-cardistry-bonus` 與 `ability:bEXmm4rKOs:front:rest-banish-cardistry-discount`。Hindered 由 keyword Slot 決定進場狀態；trigger 由 Cardistry event 建立 `+1 Power`／true sight operation 並保留 LKI；activated Slot 使用共用 `rest_source`、`banish_source`、Ability Instance、Effects Stack 與 `cardistry_discount` operation。正常 Player View、rest 非法宣告、來源離場、跨 Cardistry 與 replay scenarios 通過，所有 Thornes Card ID 行為路徑已刪除。

**Common primitive record:**

- `banish_source`、Object source validation、Source LKI、Effects Stack：Grand Crusader's Ring、兩張 Resonance Bauble、Smoke Bombs、Safeguard Amulet、The Duchess's Thornes；同步更新 Common 04 與 11 的 `Delivered slice`。
- `rest_source` 與 resolution-time `cardistry_discount`：The Duchess's Thornes；以 mechanism-level operation 實作並同步更新 Common 11，沒有卡牌專用 mutation API。
- event target、trigger ordering、damage operation：Impact Hammer；沿用 Common 07 已交付 primitive。
- central static evaluator 與 Wield derived Reserve cost：Impact Hammer、Bulwark Sword；同步更新 Common 10 的 `Delivered slice`。
- activation Reserve cost permission query：Viridian Protective Trinket；同步更新 Common 09 的 `Delivered slice`。
- Replacement Pipeline transformation：Infernal Vessel；沒有 replacement direct mutation。

**Quality gate:** `go test ./...`、`go test -race ./...`、`FuzzRejectedInputPreservesStateHash` 10 秒與 `FuzzReplayHashDeterminism` 10 秒均通過；涵蓋 Standard Game、完整 Support Set materialization、Player View、replay/state hash、trigger ordering、Replacement Pipeline 與鏡像玩家資料。

- [x] 每個 Slot 有穩定 ID、規則來源與 scenario，涵蓋正常結算、非法宣告、來源失效及跨機制互動。
- [x] Object 能力共用宣告生命週期；static 只輸入中央 evaluator，replacement 只輸入 Replacement Pipeline。
- [x] 每個 Slot 都記錄使用的共通 primitive；缺少的 primitive 已先以 mechanism-level operation 實作並同步更新對應 Common issue，沒有卡牌專用 direct mutation 或暫時 API。
- [x] 已遷移 Slot 的 Card ID 路徑全部刪除，不留無法驗證的 Supported Slot。
- [x] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
