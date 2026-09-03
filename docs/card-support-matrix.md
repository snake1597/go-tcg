# 固定牌組與 Support Set 覆蓋矩陣

本文件依 [`card-support-matrix-template.md`](./card-support-matrix-template.md) 建立，是目前唯一固定 Standard 牌組的正式 registry 輸入。它只記錄已從 `card/*.json` 頂層資料核對過的事實；未釘選的卡面、衍生卡或規則實作一律標示為 `blocked` 或 `unsupported`。

## 官方 card JSON 對照

- Card ID 是 `entity.Card.UUID`，對應各 JSON 頂層 `uuid`；`slug` 與 `name` 僅作重抓與人工核對。
- `cost`、`classes`、`types`、`subtypes`、`elements`、數值、`effect_raw`、`references` 與 `referenced_by` 必須由完整 `entity.Card` 匯入流程保留。
- `editions` 與 `result_editions` 不作 Card ID 或規則來源。
- CardFace ID 不是官方 card JSON 或目前 `entity.Card` 的欄位；尚未建立 registry 配置前，所有值均為 `unassigned`。

## 版本與狀態

| 欄位 | 值 |
| --- | --- |
| 牌組 ID／版本 | `standard-alice-fire-v1` |
| 規則 commit | `602c917f2f8fd4df7198429a72eb596bf7f647c6` |
| 卡面資料來源／版本 | Repository `./card/*.json`；[`card-data-v1`](../card-data-manifest.json)，資料集 SHA-256 `40558216461a31bb9c203f91cf7555c3dfc7b794f9e1a9203e286c2d0eb676bf` |
| 引擎最低版本 | 未建立 |
| 負責人 | 未指定 |
| 狀態 | `blocked` |

阻擋原因：尚未配置 CardFace ID、未取得 `Rile the Abyss` 的 Card JSON 與 Outside Game Pool 數量，且所有卡牌機制仍是 `unsupported`。

## 牌組清單（Deck Manifest）

### 主牌組（Main Deck，60 張）

| Card UUID | 卡名 | 數量 | CardFace ID | 備註 |
| --- | --- | ---: | --- | --- |
| `hbpu4fo8oo` | Blighted Jewel | 4 | `unassigned` | |
| `v0gu8efq08` | Lingering Banshee | 4 | `unassigned` | |
| `s9ICPMYPNx` | Bill, Chimney Sweep | 4 | `unassigned` | |
| `stiyh3pmk3` | Cinder Geyser | 4 | `unassigned` | |
| `BqDw4Mei4C` | Creative Shock | 4 | `unassigned` | |
| `qzv380ujf5` | Duchess, Six of Hearts | 4 | `unassigned` | |
| `PptfA8gG6h` | Emberwrath Witch | 4 | `unassigned` | |
| `1gxrpx8jyp` | Fanatical Devotee | 4 | `unassigned` | |
| `5pw07bh5wf` | Fractal of Sparks | 4 | `unassigned` | |
| `cbNF64gCsS` | Furnace Drone | 4 | `unassigned` | |
| `26ya6zaae8` | Incinerated Templar | 4 | `unassigned` | |
| `Vl03t5rMSA` | Incinerator Felindroid | 1 | `unassigned` | |
| `wtHBZAdTSv` | Nether Dodobird | 4 | `unassigned` | |
| `hdvpug4d5m` | Searing Rebuke | 4 | `unassigned` | Duchess 可 copy 的火元素 action（reserve 2） |
| `4vjkezn49t` | Vengeful Paramour | 4 | `unassigned` | |
| `ecZsQQAYJJ` | Volda, Smolder's Spite | 3 | `unassigned` | |

### Material Deck（12 張）

| Card UUID | 卡名 | 數量 | CardFace ID | 是否為起始 Level 0 Champion | 備註 |
| --- | --- | ---: | --- | --- | --- |
| `LMyKyVC2O9` | Spirit of Fire | 1 | `unassigned` | 是 | `CHAMPION`、Level 0 |
| `GiQxfpKTUC` | Alice, Distorted Queen | 1 | `unassigned` | 否 | `CHAMPION`、Level 1 |
| `9gv4vm4kj3` | Backup Charger | 1 | `unassigned` | 否 | `REGALIA` |
| `2gv7DC0KID` | Grand Crusader's Ring | 1 | `unassigned` | 否 | `REGALIA` |
| `pol1nz0j1n` | Nullifying Mirror | 1 | `unassigned` | 否 | `REGALIA` |
| `yj2rJBREH8` | Safeguard Amulet | 1 | `unassigned` | 否 | `REGALIA` |
| `ScGcOmkoQt` | Smoke Bombs | 1 | `unassigned` | 否 | `REGALIA` |
| `xnrw8qq1uw` | Tariff Ring | 1 | `unassigned` | 否 | `REGALIA` |
| `s3572j3oda` | Viridian Protective Trinket | 1 | `unassigned` | 否 | `REGALIA` |
| `chsbalegbs` | Impact Hammer | 1 | `unassigned` | 否 | `REGALIA` |
| `vgWgu1DUYv` | Infernal Vessel | 1 | `unassigned` | 否 | `REGALIA` |
| `1ubrwubSQN` | Mantle of the Abyss | 1 | `unassigned` | 否 | `REGALIA`；Generate `Rile the Abyss` |

### Outside Game Pool

| Card ID | 卡名 | 可用數量 | 由何效果 Generate／取得 | 備註 |
| --- | --- | ---: | --- | --- |
| 未取得 | Rile the Abyss | 未指定 | `Mantle of the Abyss`（`1ubrwubSQN`）Generate 後放入 memory | `card/` 無資料；未補 Card ID、卡面與可用數量前，此 manifest 不可進入 production registry。 |

## 內容閉包

`card:*` 列均由 deck 可達；所有 CardFace／Ability Slot 均尚未配置。`copy:searing-rebuke` 表示 Duchess 所建立的 copy instance，讀取已列入的 Searing Rebuke 行為。

| Content ID | 種類 | 官方 Card UUID／來源 | 從何內容可達 | 可達方式 | 所有 CardFace／Ability Slot | 實作狀態 | 測試狀態 | 阻擋項目 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `card:LMyKyVC2O9` | card | `LMyKyVC2O9` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:GiQxfpKTUC` | card | `GiQxfpKTUC` | deck root | deck／level up | `unassigned` | unsupported | missing | registry、Phantasmagoria |
| `card:9gv4vm4kj3` | card | `9gv4vm4kj3` | deck root | deck | `unassigned` | unsupported | missing | registry、Powercell |
| `card:2gv7DC0KID` | card | `2gv7DC0KID` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:pol1nz0j1n` | card | `pol1nz0j1n` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:yj2rJBREH8` | card | `yj2rJBREH8` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:ScGcOmkoQt` | card | `ScGcOmkoQt` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:xnrw8qq1uw` | card | `xnrw8qq1uw` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:s3572j3oda` | card | `s3572j3oda` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:chsbalegbs` | card | `chsbalegbs` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:vgWgu1DUYv` | card | `vgWgu1DUYv` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:1ubrwubSQN` | card | `1ubrwubSQN` | deck root | deck | `unassigned` | unsupported | missing | registry、Rile the Abyss |
| `card:hbpu4fo8oo` | card | `hbpu4fo8oo` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:v0gu8efq08` | card | `v0gu8efq08` | deck root | deck／ephemerate | `unassigned` | unsupported | missing | registry |
| `card:s9ICPMYPNx` | card | `s9ICPMYPNx` | deck root | deck／ephemerate | `unassigned` | unsupported | missing | registry |
| `card:stiyh3pmk3` | card | `stiyh3pmk3` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:BqDw4Mei4C` | card | `BqDw4Mei4C` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:qzv380ujf5` | card | `qzv380ujf5` | deck root | deck | `unassigned` | unsupported | missing | registry、copy |
| `card:PptfA8gG6h` | card | `PptfA8gG6h` | deck root | deck／ephemerate | `unassigned` | unsupported | missing | registry |
| `card:1gxrpx8jyp` | card | `1gxrpx8jyp` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:5pw07bh5wf` | card | `5pw07bh5wf` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:cbNF64gCsS` | card | `cbNF64gCsS` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:26ya6zaae8` | card | `26ya6zaae8` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:Vl03t5rMSA` | card | `Vl03t5rMSA` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:wtHBZAdTSv` | card | `wtHBZAdTSv` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `card:hdvpug4d5m` | card | `hdvpug4d5m` | deck root | deck／copy source | `unassigned` | unsupported | missing | registry、copy |
| `card:4vjkezn49t` | card | `4vjkezn49t` | deck root | deck／ephemerate | `unassigned` | unsupported | missing | registry |
| `card:ecZsQQAYJJ` | card | `ecZsQQAYJJ` | deck root | deck | `unassigned` | unsupported | missing | registry |
| `mastery:phantasmagoria` | mastery | Alice reference | `card:GiQxfpKTUC` | create／gain | `unassigned` | unsupported | missing | mastery rules、registry |
| `token:powercell` | token | Backup Charger reference | `card:9gv4vm4kj3` | summon | `unassigned` | unsupported | missing | token characteristics、registry |
| `card:rile-the-abyss` | card | Mantle reference；Card ID 未取得 | `card:1ubrwubSQN` | generate | unassigned | unsupported | missing | Card JSON、Outside Game Pool |
| `copy:searing-rebuke` | copy | `hdvpug4d5m` | `card:qzv380ujf5` | copy | `unassigned` | unsupported | missing | copy／activation rules |

### 卡名索引依賴

| 來源 Card ID | 效果 | 所需名稱集合／資料版本 | 是否建立或執行被命名卡牌 |
| --- | --- | --- | --- |
| 無 | 本固定牌組未發現僅按卡名查詢、且不建立或執行目標卡牌的效果。 | 不適用 | 否 |

## 規則與機制覆蓋

所有列皆為首版未支援；這些不是實作承諾的完成狀態，而是 Support Set 必須逐項消除的清單。

| 機制 ID | 規則來源與條目 | 由哪些內容需要 | 所需 typed operation | 正常案例 | 非法／邊界案例 | 實作狀態 | 測試狀態 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `MEC-001` 開局／抽牌／deckout | `general-rules-starting-the-game.md`；`general-rules-ending-the-game.md` | Spirit、所有 deck card | shuffle、draw、deckout | Spirit On Enter 抽 7 | 牌庫不足 | unsupported | missing |
| `MEC-002` materialize／level up | `playing-cards-card-materialization.md`；`card-types-champion.md` | Material Deck、Alice | materialize、lineage、payment | Spirit 後 materialize Alice | 不合法 lineage／費用 | unsupported | missing |
| `MEC-003` activation 與目標 | `playing-cards-card-activation.md` | actions、regalia、activated abilities | declare、choose target、pay、resolve、fizzle | Cinder Geyser 傷害 unit | 無合法目標 | unsupported | missing |
| `MEC-004` 觸發與延遲觸發 | `abilities-triggered-abilities.md` | On Enter／Death／end phase cards | trigger、delayed trigger、sacrifice | Emberwrath Witch end phase | source 已離場 | unsupported | missing |
| `MEC-005` 連續效果與 characteristic layers | `continuous-effects/README.md` | Mirror、Viridian、Blighted、Fanatical、Impact、Tariff、Volda | layer modifier、cost modifier、permission／prohibition | Mirror 改變 memory element | 相依／timestamp 衝突 | unsupported | missing |
| `MEC-006` prevention／replacement | `replacement-effects.md` | Safeguard、Infernal、Searing、Volda | prevent damage、replace recover、can't prevent | Searing 防止傷害 | 0 或負 recover | unsupported | missing |
| `MEC-007` Ephemerate 與 graveyard | `keywords-and-abilities.md#ephemerate` | Banshee、Bill、Witch、Paramour | alternative activation、graveyard move、ephemeral | Witch 從 graveyard activation | 非法 timing／費用 | unsupported | missing |
| `MEC-008` counters／Mastery／token／Generate | `game-mechanics-counters.md`；`game-mechanics-mastery.md`；`game-terms.md#generate` | Alice、Backup、Mantle | gain mastery、counter、summon token、generate | Backup Charger 生成 Powercell | 缺 Rile card／pool copy | unsupported | missing |
| `MEC-009` copy 與 LKI | `playing-cards-resolution.md`；`game-terms.md#reflexive-trigger` | Duchess、Searing Rebuke | copy activation、source ref、LKI | Cardistry copy Searing | source／target 已失效 | unsupported | missing |
| `MEC-010` combat | `combat-phase-attack-declaration.md`；`keywords-and-abilities.md#taunt` | Impact、Smoke、Tariff、Furnace | attack declare、taunt、stealth、attack tax | Taunt target restriction | 無法支付 attack tax | unsupported | missing |
| `MEC-011` keyword-specific payment | `keywords-and-abilities.md#kindle-n`；`#reservable`；`game-terms.md#label-keywords` | Duchess、Fractal、Furnace | banish payment、rest payment、cost reduction | Kindle／Reservable payment | 重複或非法支付 | unsupported | missing |

## 規則裁定 Gate

| Issue ID | 是否由 Support Set 觸及 | 狀態 | 裁定來源／排除理由 | 對應測試 |
| --- | --- | --- | --- | --- |
| `RUL-001` | 是 | 專案自訂裁定 | [ADR 0015](./adr/0015-retain-opportunity-until-the-holder-passes.md)：玩家完成需要 Opportunity 的行動後保有 Opportunity，直到讓過。 | `MEC-003`、`MEC-004`、`MEC-009` |
| `RUL-002` | 是 | 已由官方來源解決 | 只在所有必要目標都非法或不存在時 fizzle；仍有合法必要目標時對其結算。 | `MEC-003` |
| `RUL-004` | 是 | 已由官方來源解決 | token 先完成 zone change，於下一次 state-based checks 中 cease to exist；已產生的 trigger 可使用 LKI。 | `MEC-008` |

## 實作順序

1. 為 `Spirit of Fire` 配置 CardFace ID，完成 Standard 開局及 On Enter 抽七張。
2. 匯入並配置 Phantasmagoria、Powercell 與 Rile the Abyss；明確指定 Outside Game Pool 數量。
3. 以一張依賴最少的正式卡通過 activation、目標、費用、Effects Stack 與 checkpoint。
4. 依機制表由葉節點向外完成其他 card／token／mastery／copy。
5. 每完成一項，同步補正常與邊界測試；所有 closure 內容完成前保持 `blocked`。

## 完成條件

- Deck Manifest 的 UUID、CardFace ID、資料版本與 Outside Game Pool 均可回溯。
- 重新計算三個牌池的閉包不會發現矩陣外內容。
- 每個內容與機制皆為 `supported` 且有正常及邊界測試。
- 沒有被 Support Set 觸及且未解決的 ruling issue。
