# 卡牌資料、固定牌組與 Support Set

本文件是卡牌資料流程、固定 Standard 牌組及 Support Set 的唯一文件入口。CardFace、Ability Slot 與可達依賴都在此固定；其他開發文件只引用本文件，不複製牌表、資料版本或支援狀態。

## 文件責任

本文件負責：

- 卡面資料的權威來源、版本與驗證流程。
- 固定 Main Deck、Material Deck 與 Outside Game Pool。
- CardFace／Ability Slot ID 與完整能力 inventory。
- Support Set closure、dependency graph 與內容支援狀態。

規則語意由 `rules/` 與 [`rules-issues.md`](./rules-issues.md) 擁有；產品架構與完成門檻由 [`development-plan.md`](./development-plan.md) 擁有；目前工作狀態由 [`implementation-features.md`](./implementation-features.md) 擁有。

## 卡面資料來源與版本

唯一卡面資料來源是 repository 內的 `./card/*.json`。Game Module、測試及建置流程不得在執行期間向外部服務取得較新的卡面，也不得在版本不符時自動 fallback。

每張卡以 JSON 頂層 `uuid` 作為 Card ID；`editions` 與 `result_editions` 內的 UUID 只代表印刷版本，不作為 Card ID 或 CardFace ID。

Repository 根目錄的 [`card-data-manifest.json`](../card-data-manifest.json) 是卡面資料版本、檔案清單及 digest 的機器可檢查來源，記錄：

- schema 與資料版本。
- 固定來源 `./card/*.json`。
- 每個檔案的相對路徑、Card ID、slug、名稱及 `last_update`。
- 每個檔案及整份資料集的 SHA-256。

資料集 digest 依檔名排序後，對每張卡依序加入 UTF-8 檔名、NUL byte、原始檔案 SHA-256 bytes 與換行，再計算 SHA-256。`last_update` 只供來源追溯；內容是否一致以 digest 為準。

### 驗證與更新

驗證目前工作目錄沒有卡面漂移：

```bash
go run ./cmd/tool/card_manifest
```

有意更新 `./card` 後，明確提高 data version 並重新產生 manifest：

```bash
go run ./cmd/tool/card_manifest -write -version card-data-vNEXT
```

更新必須與 manifest、本文件、受影響測試及 replay 相容性變更一起 review。Manifest 驗證會拒絕無效 JSON、多份 JSON 串接、缺少必要 metadata、重複 Card ID／slug、檔名與 slug 不一致，以及任何未反映在 manifest 的內容變更。

## 官方 card JSON 對照

- Card ID 是 `entity.Card.UUID`，對應各 JSON 頂層 `uuid`；`slug` 與 `name` 僅作人工核對。
- `cost`、`classes`、`types`、`subtypes`、`elements`、數值、`effect_raw`、`references` 與 `referenced_by` 必須由完整 `entity.Card` 匯入流程保留。
- `editions` 與 `result_editions` 不作 Card ID 或規則來源。
- CardFace ID 使用 `face:<card-id>:<face-key>`。目前 32 種內容都是單面卡，因此 `face-key` 均為 `front`；未來雙面卡使用 `front`／`back`。它是 registry ID，不借用印刷版本 UUID。
- Ability Slot ID 使用 `ability:<card-id>:<face-key>:<slot-key>`。`slot-key` 描述 rules-bearing behavior 的規則語意，不使用段落序號、效果文字或 hash 自動產生。
- `card-id` 保留權威資料的大小寫；`face-key` 與 `slot-key` 必須符合小寫 kebab-case。完整格式、變更生命週期與拒絕規則由 [ADR 0016](./adr/0016-use-hierarchical-content-ids.md) 固定。
- `effect_raw` 必須連同 `rule` 中的勘誤與裁定一起解析；兩者衝突時，以日期較新的有效勘誤文字建立行為。

## 版本與狀態

| 欄位 | 值 |
| --- | --- |
| 牌組 ID／版本 | `standard-fire-v2` |
| 規則基準 | 由 [`development-plan.md`](./development-plan.md) 與 [ADR 0003](./adr/0003-pin-rules-snapshot-per-engine-version.md) 固定 |
| 卡面資料來源／版本 | Repository `./card/*.json`；實際 data version 與 SHA-256 只讀取 [`card-data-manifest.json`](../card-data-manifest.json) |
| 引擎最低版本 | 未建立 |
| 負責人 | 未指定 |
| 狀態 | `blocked` |

阻擋原因：本牌組涉及的卡牌機制尚未完成實作與測試。內容 ID 與閉包已固定，可以依下方 dependency graph 切分正式 slices。

## 牌組清單（Deck Manifest）

### 主牌組（Main Deck，60 張）

| Card UUID | 卡名 | 數量 | CardFace ID | 備註 |
| --- | --- | ---: | --- | --- |
| `i9hf5lhl5f` | Five of Spades | 3 | `face:i9hf5lhl5f:front` | |
| `8bolq2y5qp` | Four of Spades | 4 | `face:8bolq2y5qp:front` | |
| `wbjc9t8ycp` | Noire, Ace of Spades | 3 | `face:wbjc9t8ycp:front` | |
| `o09csnorqv` | Three of Spades | 3 | `face:o09csnorqv:front` | |
| `w7g91ru45w` | Trump Set | 2 | `face:w7g91ru45w:front` | |
| `e8ygl32jef` | Two of Spades | 4 | `face:e8ygl32jef:front` | |
| `0mf1ug6yfi` | Wonderland's Reign | 1 | `face:0mf1ug6yfi:front` | |
| `GjM8b5fxqj` | Arthur, Young Heir | 4 | `face:GjM8b5fxqj:front` | |
| `iohZMWh5v5` | Blazing Throw | 3 | `face:iohZMWh5v5:front` | |
| `qzv380ujf5` | Duchess, Six of Hearts | 3 | `face:qzv380ujf5:front` | |
| `gt2zqtgs42` | Fiery Interference | 3 | `face:gt2zqtgs42:front` | |
| `xgax8bbjqj` | Four of Hearts | 4 | `face:xgax8bbjqj:front` | |
| `td460e8ig0` | Heated Vengeance | 1 | `face:td460e8ig0:front` | |
| `lcy0lw1veb` | Peppered Chef | 2 | `face:lcy0lw1veb:front` | |
| `5du8f077ua` | Red Hare, Unrivaled Stallion | 3 | `face:5du8f077ua:front` | |
| `h68dr63eo5` | Rouge, Ace of Hearts | 3 | `face:h68dr63eo5:front` | |
| `28bjn8g50v` | Straight Flare | 4 | `face:28bjn8g50v:front` | |
| `1db8hz4prm` | Three of Hearts | 4 | `face:1db8hz4prm:front` | |
| `rufki4o41y` | Two of Hearts | 4 | `face:rufki4o41y:front` | |
| `4qc47amgpp` | Verita, Queen of Hearts | 2 | `face:4qc47amgpp:front` | |

### Material Deck（12 張）

| Card UUID | 卡名 | 數量 | CardFace ID | 是否為起始 Level 0 Champion | 備註 |
| --- | --- | ---: | --- | --- | --- |
| `LMyKyVC2O9` | Spirit of Fire | 1 | `face:LMyKyVC2O9:front` | 是 | `CHAMPION`、Level 0 |
| `zb14m4c8lj` | Tonoris, Lone Mercenary | 1 | `face:zb14m4c8lj:front` | 否 | `CHAMPION`、Level 1 |
| `8kmoi0a5uh` | Bulwark Sword | 1 | `face:8kmoi0a5uh:front` | 否 | `REGALIA` |
| `2gv7DC0KID` | Grand Crusader's Ring | 1 | `face:2gv7DC0KID:front` | 否 | `REGALIA` |
| `yj2rJBREH8` | Safeguard Amulet | 1 | `face:yj2rJBREH8:front` | 否 | `REGALIA` |
| `ScGcOmkoQt` | Smoke Bombs | 1 | `face:ScGcOmkoQt:front` | 否 | `REGALIA` |
| `s3572j3oda` | Viridian Protective Trinket | 1 | `face:s3572j3oda:front` | 否 | `REGALIA` |
| `dSSRtNnPtw` | Water Resonance Bauble | 1 | `face:dSSRtNnPtw:front` | 否 | `REGALIA` |
| `bHGUNMFLg9` | Wind Resonance Bauble | 1 | `face:bHGUNMFLg9:front` | 否 | `REGALIA` |
| `chsbalegbs` | Impact Hammer | 1 | `face:chsbalegbs:front` | 否 | `REGALIA` |
| `vgWgu1DUYv` | Infernal Vessel | 1 | `face:vgWgu1DUYv:front` | 否 | `REGALIA` |
| `bEXmm4rKOs` | The Duchess's Thornes | 1 | `face:bEXmm4rKOs:front` | 否 | `REGALIA` |

### Outside Game Pool

| Card ID | 卡名 | 可用數量 | 由何效果 Generate／取得 | 備註 |
| --- | --- | ---: | --- | --- |
| 無 | 空集合 | 0 | 不適用 | 32 種卡牌沒有 Generate 或其他建立牌組外卡牌的效果。 |

## 內容閉包

本次 manifest 的根內容為 32 種卡牌（主牌組 20 種、Material Deck 12 種）。完整解析 `effect_raw` 與 `rule` 後，沒有 token、Mastery、Status、Generate、雙面卡或牌組外卡牌依賴。Duchess 會建立既有卡牌的 runtime copy；copy 沿用來源 CardFace 的規則行為，不取得新的 Card ID，但 registry 必須支援獨立 Object／StackItem 身分。

| Content ID | 種類 | 官方 Card UUID／來源 | 從何內容可達 | 可達方式 | 所有 CardFace／Ability Slot | 實作狀態 | 測試狀態 | 阻擋項目 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `card:<Card UUID>`（32 筆，見 Deck Manifest） | card | 對應 Deck Manifest 的 Card UUID | deck root | deck | 見下方完整 inventory | unsupported | missing | registry、mechanics |
| `runtime:copied-action` | runtime object／stack item | 被複製 CardFace | Duchess Cardistry | copy、optional activate | 沿用來源 face 的 resolution slot | unsupported | missing | copy、object identity、free activation |

## CardFace 與 Ability Slot inventory

Ability Slot 只切分卡面上可獨立追蹤的 printed behavior。一般卡牌 activation／materialization、進入區域及規則預設行為屬共用機制，不為每張卡重複建立 slot。

| 卡牌 | CardFace ID | Ability Slot ID | 分類 | 主要需求 |
| --- | --- | --- | --- | --- |
| Arthur, Young Heir | `face:GjM8b5fxqj:front` | `ability:GjM8b5fxqj:front:on-enter-rest-immortality`<br>`ability:GjM8b5fxqj:front:rested-allies-power` | triggered<br>static continuous | optional rest、duration、immortality、ally power layer |
| Blazing Throw | `face:iohZMWh5v5:front` | `ability:iohZMWh5v5:front:sacrifice-weapon-cost`<br>`ability:iohZMWh5v5:front:deal-four` | additional cost<br>resolution | controlled weapon、atomic sacrifice、target、damage、fizzle |
| Bulwark Sword | `face:8kmoi0a5uh:front` | `ability:8kmoi0a5uh:front:class-bonus-power`<br>`ability:8kmoi0a5uh:front:wield-payment` | static continuous<br>additional cost | class match、weapon power layer、wield payment；採用 2026-08-16 errata |
| Duchess, Six of Hearts | `face:qzv380ujf5:front` | `ability:qzv380ujf5:front:kindle-six`<br>`ability:qzv380ujf5:front:cardistry-copy-action` | keyword<br>activated | kindle、Cardistry discount、once per instance、graveyard banish、copy、optional free activation |
| Fiery Interference | `face:gt2zqtgs42:front` | `ability:gt2zqtgs42:front:deal-two-recover-lock` | resolution | fast activation、target、damage、conditional champion controller、turn-duration prohibition |
| Five of Spades | `face:i9hf5lhl5f:front` | `ability:i9hf5lhl5f:front:cardistry-power-five`<br>`ability:i9hf5lhl5f:front:floating-memory` | activated<br>keyword | Cardistry discount、once per instance、power layer、Floating Memory |
| Four of Hearts | `face:xgax8bbjqj:front` | `ability:xgax8bbjqj:front:cardistry-memory-deploy` | activated | Cardistry discount、once per instance、draw to memory、qualified selection、put onto field |
| Four of Spades | `face:8bolq2y5qp:front` | `ability:8bolq2y5qp:front:cardistry-memory-draw` | activated | Cardistry discount、once per instance、draw to memory |
| Grand Crusader's Ring | `face:2gv7DC0KID:front` | `ability:2gv7DC0KID:front:divine-relic`<br>`ability:2gv7DC0KID:front:banish-draw` | deck restriction<br>activated | material deck validation、self banish、draw |
| Heated Vengeance | `face:td460e8ig0:front` | `ability:td460e8ig0:front:damaged-champion-power`<br>`ability:td460e8ig0:front:on-attack-self-damage` | static continuous<br>triggered | damage-this-turn history、attack power layer、optional self damage、class bonus |
| Impact Hammer | `face:chsbalegbs:front` | `ability:chsbalegbs:front:class-bonus-power`<br>`ability:chsbalegbs:front:on-wield-self-damage` | static continuous<br>triggered | class match、weapon power layer、wield event、damage；採用 2026-08-16 errata |
| Infernal Vessel | `face:vgWgu1DUYv:front` | `ability:vgWgu1DUYv:front:replace-recover` | replacement | recover replacement、recompute amount、zero-or-less result、recover cost distinction |
| Noire, Ace of Spades | `face:wbjc9t8ycp:front` | `ability:wbjc9t8ycp:front:suited-stealth`<br>`ability:wbjc9t8ycp:front:on-enter-suited-counters` | static continuous<br>triggered | another Suited ally predicate、grant stealth、Suited reserve-total thresholds、buff counters |
| Peppered Chef | `face:lcy0lw1veb:front` | `ability:lcy0lw1veb:front:on-enter-sacrifice-power` | triggered | optional other-ally sacrifice、atomic choice、turn-duration power layer |
| Red Hare, Unrivaled Stallion | `face:5du8f077ua:front` | `ability:5du8f077ua:front:pride-three`<br>`ability:5du8f077ua:front:unique-human-removes-pride`<br>`ability:5du8f077ua:front:granted-on-attack-filter` | keyword<br>static continuous<br>granted trigger | obedience restriction、remove keyword、grant ability、optional discard then draw |
| Rouge, Ace of Hearts | `face:h68dr63eo5:front` | `ability:h68dr63eo5:front:on-enter-suited-threshold-damage` | triggered | Suited reserve-total thresholds、choose without target、conditional damage |
| Safeguard Amulet | `face:yj2rJBREH8:front` | `ability:yj2rJBREH8:front:banish-prevent-noncombat` | activated replacement | self banish、delayed prevention、non-combat damage、champion scope、replacement expiry |
| Smoke Bombs | `face:ScGcOmkoQt:front` | `ability:ScGcOmkoQt:front:banish-stealth-draw` | activated | self banish、target ally、grant stealth until end of turn、draw、attack target revalidation |
| Spirit of Fire | `face:LMyKyVC2O9:front` | `ability:LMyKyVC2O9:front:on-enter-draw-seven` | triggered | starting Champion、On Enter ordering、draw seven、deckout |
| Straight Flare | `face:28bjn8g50v:front` | `ability:28bjn8g50v:front:suited-count-damage` | resolution | fast activation、target、distinct Suited costs、dynamic damage、fizzle |
| The Duchess's Thornes | `face:bEXmm4rKOs:front` | `ability:bEXmm4rKOs:front:hindered`<br>`ability:bEXmm4rKOs:front:cardistry-trigger-power-true-sight`<br>`ability:bEXmm4rKOs:front:cardistry-discount` | keyword<br>triggered<br>activated continuous | Hindered、Cardistry activation event、source ally、power／true sight duration、rest and banish cost、next-use discount |
| Three of Hearts | `face:1db8hz4prm:front` | `ability:1db8hz4prm:front:cardistry-draw-discard` | activated | Cardistry discount、once per instance、draw、mandatory discard |
| Three of Spades | `face:o09csnorqv:front` | `ability:o09csnorqv:front:cardistry-life-two` | activated | Cardistry discount、once per instance、controlled Suited ally target、life layer、duration |
| Tonoris, Lone Mercenary | `face:zb14m4c8lj:front` | `ability:zb14m4c8lj:front:on-enter-taunt` | triggered | lineage materialization、grant taunt、until-next-turn duration、attack declaration restriction |
| Trump Set | `face:w7g91ru45w:front` | `ability:w7g91ru45w:front:class-bonus-discount`<br>`ability:w7g91ru45w:front:retarget-attack-buff` | static continuous<br>resolution | fast reaction timing、cost modifier、active attack、different Suited ally、retarget、power／life duration |
| Two of Hearts | `face:rufki4o41y:front` | `ability:rufki4o41y:front:cardistry-power-two` | activated | Cardistry discount、once per instance、power layer、duration |
| Two of Spades | `face:e8ygl32jef:front` | `ability:e8ygl32jef:front:fast-activation`<br>`ability:e8ygl32jef:front:cardistry-buff-counter` | permission<br>activated | fast activation、Cardistry discount、once per instance、buff counter |
| Verita, Queen of Hearts | `face:4qc47amgpp:front` | `ability:4qc47amgpp:front:suited-alternative-cost`<br>`ability:4qc47amgpp:front:suited-immortality`<br>`ability:4qc47amgpp:front:on-death-suited-power` | alternative cost<br>static continuous<br>triggered | choose 3+ graveyard cards totaling exactly 10、atomic banish、grant immortality、On Death、until-end-of-next-turn duration |
| Viridian Protective Trinket | `face:s3572j3oda:front` | `ability:s3572j3oda:front:opponent-water-tax` | static continuous | active-player condition、opponent water card predicate、activation cost layer；採用 2023-08-15 errata |
| Water Resonance Bauble | `face:dSSRtNnPtw:front` | `ability:dSSRtNnPtw:front:conditional-banish-draw` | activated | opponent water Champion condition、self banish、draw；固定鏡像牌組中無合法啟動時點 |
| Wind Resonance Bauble | `face:bHGUNMFLg9:front` | `ability:bHGUNMFLg9:front:conditional-banish-draw` | activated | opponent wind Champion condition、self banish、draw；固定鏡像牌組中無合法啟動時點 |
| Wonderland's Reign | `face:0mf1ug6yfi:front` | `ability:0mf1ug6yfi:front:cardistry-draw` | activated | Phantasia、Cardistry discount、once per instance、draw |

共配置 49 個 Ability Slot。每個 slot 都必須個別具有 registry 狀態及測試狀態；同一卡牌的所有 slot 均完成後，該 CardFace 才能標記 `supported`。

## Support Set dependency graph

```text
fixed Standard deck
├── lifecycle
│   ├── Spirit of Fire ── On Enter / draw seven / deckout
│   └── Tonoris ───────── materialize / lineage / taunt duration
├── action transaction
│   ├── Blazing Throw ─── weapon sacrifice ── Bulwark Sword, Impact Hammer
│   ├── Fiery Interference ─ damage / recover prohibition
│   ├── Straight Flare ── distinct Suited costs / damage
│   └── Trump Set ─────── attack retarget / temporary characteristics
├── Suited and Cardistry kernel
│   ├── Cardistry actors ─ Duchess, Five of Spades, Four of Hearts,
│   │                      Four of Spades, Three of Hearts, Three of Spades,
│   │                      Two of Hearts, Two of Spades, Wonderland's Reign
│   ├── observer ───────── The Duchess's Thornes
│   └── Suited queries ─── Noire, Rouge, Straight Flare, Trump Set, Verita
├── runtime copy
│   └── Duchess ────────── Blazing Throw, Fiery Interference, Straight Flare
├── combat
│   ├── weapon/wield ───── Bulwark Sword, Impact Hammer, Blazing Throw
│   ├── attack state ───── Heated Vengeance, Red Hare, Trump Set
│   └── targeting ──────── Tonoris, Noire, Smoke Bombs, The Duchess's Thornes
├── continuous and granted abilities
│   ├── power/life ─────── Arthur, Heated Vengeance, Suited cards, Trump Set
│   ├── immortality ────── Arthur, Verita
│   ├── cost ───────────── Cardistry kernel, Trump Set, Viridian Trinket
│   └── granted trigger ── Red Hare
└── replacement
    ├── damage prevention  Safeguard Amulet
    └── recover amount ─── Infernal Vessel
```

### 精確跨卡邊

| 來源 | 可達或觀察的目標 | Dependency 原因 |
| --- | --- | --- |
| Duchess Cardistry | Blazing Throw、Fiery Interference、Straight Flare | 牌組中所有 fire element、ACTION、reserve cost ≤ 2 的卡；copy 沿用被選 CardFace 的 resolution slots。 |
| Four of Hearts | Noire、Rouge、Three of Hearts、Three of Spades、Two of Hearts、Two of Spades | 牌組中所有 fire／norm、Suited、ALLY、reserve cost ≤ 3 的合法 memory 選項。 |
| The Duchess's Thornes | Duchess、Five of Spades、Four of Hearts、Four of Spades、Three of Hearts、Three of Spades、Two of Hearts、Two of Spades | 只有 ally 的 Cardistry activation 會觸發；Wonderland's Reign 是 Phantasia，不在觸發集合。 |
| Blazing Throw | Bulwark Sword、Impact Hammer | 固定牌組中可作為追加犧牲費用的 WEAPON CardFace。 |
| Red Hare | Arthur、Duchess、Rouge、Verita | 固定牌組中符合 fire／tera、UNIQUE、HUMAN 的條件來源。 |
| Noire | 其餘 Suited allies | stealth 條件要求另一個 Suited ally。 |
| Verita alternative cost | 所有 Suited ally CardFace | 從 graveyard 選至少三張且 printed reserve cost 合計必須恰好為 10。 |
| Suited reserve-total／distinct-cost queries | 15 種 Suited CardFace | Noire、Rouge、Straight Flare 與全部 Cardistry cost 計算共用同一個 derived query，不得由 card handler 各自計算。 |
| Water／Wind Resonance Bauble | Spirit、Tonoris | 固定鏡像牌組 Champion 只有 fire 與 norm element，因此條件永遠為 false；合法行動列舉必須排除這兩個 ability。 |

15 種 Suited CardFace 是 Duchess、Five of Spades、Four of Hearts、Four of Spades、Noire、Rouge、Straight Flare、The Duchess's Thornes、Three of Hearts、Three of Spades、Trump Set、Two of Hearts、Two of Spades、Verita 與 Wonderland's Reign。

### 卡名索引依賴

| 來源 Card ID | 效果 | 所需名稱集合／資料版本 | 是否建立或執行被命名卡牌 |
| --- | --- | --- | --- |
| 無 | 卡面中的卡名皆為自身指涉；其他選擇均以 characteristic predicate 決定。 | 目前的 [`card-data-manifest.json`](../card-data-manifest.json) | 否 |

## 規則與機制覆蓋

以下是完整解析後的新牌組機制範圍；CardFace 與 Ability Slot 已配置，但程式與測試完成前均不視為已支援。

| 機制 ID | 規則來源與條目 | 由哪些內容需要 | 所需 typed operation | 正常案例 | 非法／邊界案例 | 實作狀態 | 測試狀態 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `MEC-001` 開局／抽牌／deckout | `general-rules-starting-the-game.md`；`general-rules-ending-the-game.md` | Spirit、所有 deck card | shuffle、draw、deckout | Spirit 開局效果 | 牌庫不足 | unsupported | missing |
| `MEC-002` materialize／level up | `playing-cards-card-materialization.md`；`card-types-champion.md` | Spirit、Tonoris、Material Deck | materialize、lineage、payment | Spirit 後合法 materialize | 不合法 lineage／費用 | unsupported | missing |
| `MEC-003` declaration transaction／activation／target | `playing-cards-card-activation.md`；`player-action-legality.md` | actions、Regalia、activated abilities、Blazing Throw、Verita | declare、choose、target、pay、rollback、resolve、fizzle | 合法 action 與追加／替代費用 | 取消、無合法目標、費用失效 | unsupported | missing |
| `MEC-004` trigger／duration／granted ability | `abilities-triggered-abilities.md`；`abilities-ability-tracking.md` | Spirit、Tonoris、Arthur、Heated、Impact、Noire、Peppered Chef、Red Hare、Rouge、Thornes、Verita | buffer trigger、order、grant、expire、LKI | On Enter／On Attack／On Death／On Wield | source 離場、跨回合 expiry | unsupported | missing |
| `MEC-005` continuous effects／characteristic layers | `continuous-effects/README.md` | Arthur、Bulwark、Heated、Noire、Red Hare、Verita、Viridian、Cardistry cards | characteristic query、layer modifier、cost modifier、permission／prohibition | 同 layer 與 dependency 正確套用 | source 失效、timestamp／dependency 衝突 | unsupported | missing |
| `MEC-006` prevention／replacement | `replacement-effects.md`；`game-mechanics-damage-prevention.md` | Safeguard、Infernal、Fiery Interference | prevent damage、replace recover、prohibit recover、expire | 合法防止、替代與禁止 | 0 或負 recover、非戰鬥／戰鬥分類 | unsupported | missing |
| `MEC-007` combat／weapon／attack retarget | `combat-phase-attack-declaration.md`；`keywords-and-abilities.md` | Arthur、Bulwark、Heated、Impact、Red Hare、Smoke Bombs、Tonoris、Trump Set | attack declare、wield、taunt、stealth、true sight、retarget、damage | 合法攻擊、wield 與 reaction | 追加費用不足、retarget 原目標、失去目標 | unsupported | missing |
| `MEC-008` Cardistry／Suited／once tracking | `playing-cards-resolution.md`；`game-terms.md#label-keywords` | Duchess、Wonderland's Reign、Spades／Hearts cards、Thornes | distinct-cost query、discount、rest、banish、once-per-instance、activation event | 合法 Cardistry activation | 同 instance 重複啟動、source 離場後新 instance | unsupported | missing |
| `MEC-009` copy／source identity | `playing-cards-resolution.md`；[ADR 0012](./adr/0012-separate-card-object-and-stack-identities.md) | Duchess 與三張合格 fire action | copy object、source face、free optional activation、stack identity | copy 後選擇啟動並正確結算 | 不啟動、來源離開 graveyard、target 失效 | unsupported | missing |
| `MEC-010` counters／draw-discard／zone movement | `game-mechanics-counters.md`；`game-mechanics-drawing-cards.md`；game zones | Noire、Two of Spades、Four of Hearts、Three of Hearts、Red Hare、Grand Crusader's Ring、Baubles | buff counter、draw、draw-to-memory、discard、put、banish、sacrifice | 正常移動與卡牌守恆 | 空牌庫、沒有合法選項、mandatory discard | unsupported | missing |
| `MEC-011` keyword與格式限制 | `keywords-and-abilities.md`；Standard format rules | Divine Relic、Floating Memory、Fast Activation、Hindered、Immortality、Kindle、Pride、Stealth、Taunt、True Sight | keyword permission／restriction、deck validation | 關鍵字改變合法行動 | 關鍵字來源失效、互斥 permission | unsupported | missing |

## 規則裁定依賴

裁定狀態、證據與實際結論只由 [`rules-issues.md`](./rules-issues.md) 管理。本 Support Set 會觸及：

| Issue ID | 受影響機制 |
| --- | --- |
| `RUL-001` | `MEC-003`、`MEC-004`、`MEC-008` 的 Opportunity 時序 |
| `RUL-002` | `MEC-003`、`MEC-009` 的目標失效與 fizzle |
| `RUL-003` | `MEC-001`、`MEC-002` 的起始 Level 0 Champion 放置與 On Enter 順序 |

## 完成條件

- Deck Manifest 的 UUID、CardFace ID、Ability Slot ID 與資料版本均可回溯。
- 重新計算三個牌池的閉包不會發現矩陣外內容。
- 每個內容與機制皆為 `supported` 且有正常及邊界測試。
- 沒有被 Support Set 觸及且未解決的 ruling issue。
