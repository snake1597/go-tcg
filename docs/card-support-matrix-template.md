# 固定牌組與 Support Set 覆蓋矩陣範本

取得固定牌組後，複製本文件為 `docs/card-support-matrix.md`，填完所有欄位並刪除範例提示。矩陣是正式 registry、開發排序與 release gate 的共同輸入；不能只維護一份沒有規則依賴的牌表。

## 版本與狀態

| 欄位 | 值 |
| --- | --- |
| 牌組 ID／版本 | `<待填>` |
| 規則 commit | `602c917f2f8fd4df7198429a72eb596bf7f647c6` |
| 卡面資料來源／版本 | `<待填>` |
| 引擎最低版本 | `<待填>` |
| 負責人 | `<待填>` |
| 狀態 | `draft`／`blocked`／`supported` |

只有所有閉包內容與機制都達到 `supported`，整份矩陣才能標記為 `supported`。

## 牌組清單（Deck Manifest）

### 主牌組（Main Deck）

| Card ID | 卡名 | 數量 | CardFace ID | 備註 |
| --- | --- | ---: | --- | --- |
| `<待填>` | `<待填>` | 0 | `<待填>` |  |

### Material Deck

| Card ID | 卡名 | 數量 | CardFace ID | 是否為起始 Level 0 Champion | 備註 |
| --- | --- | ---: | --- | --- | --- |
| `<待填>` | `<待填>` | 0 | `<待填>` | 是／否 |  |

### Outside Game Pool

| Card ID | 卡名 | 可用數量 | 由何效果 Generate／取得 | 備註 |
| --- | --- | ---: | --- | --- |
| `<待填或明確寫無>` | `<待填>` | 0 | `<待填>` | 列入即代表玩家具備規則要求的數位副本 |

## 內容閉包

每一列代表 Support Set 中一個可實際建立或執行的內容。雙面卡的所有 faces、token、生成卡、Mastery、Status、copy 可讀取的來源行為及其他衍生內容都要列入。只要求指定卡名、但不建立或執行該卡時，改記錄於「卡名索引依賴」。

| Content ID | 種類 | 從何內容可達 | 可達方式 | 所有 CardFace／Ability Slot | 實作狀態 | 測試狀態 | 阻擋項目 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `<待填>` | card／token／mastery／status／其他 | `<Card ID>` | deck／generate／create／copy／transform／level up | `<待填>` | unsupported／supported | missing／passing | `<RUL-ID 或無>` |

### 卡名索引依賴

| 來源 Card ID | 效果 | 所需名稱集合／資料版本 | 是否建立或執行被命名卡牌 |
| --- | --- | --- | --- |
| `<待填>` | `<待填>` | `<待填>` | 否；若為是，移入內容閉包 |

## 規則與機制覆蓋

| 機制 ID | 規則來源與條目 | 由哪些內容需要 | 所需 typed operation | 正常案例 | 非法／邊界案例 | 實作狀態 | 測試狀態 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `<待填>` | `<rules 相對路徑與條目>` | `<Content ID>` | `<例如 draw、deal damage、Layer E modifier>` | `<scenario ID>` | `<scenario ID>` | unsupported／supported | missing／passing |

以下類別必須逐一檢查，即使最後填「不適用」：

- 區域、公開／私密資訊、shuffle 與 ViewHandle 撤銷。
- Opportunity、Fast／Slow timing、Effects Stack、trigger 與 reflexive／delayed trigger。
- 宣告交易、cost parameter、mode、target、payment、fizzle 與 negate。
- Card／Object／Lineage／Transform／copy／LKI 身分。
- Continuous Layer A–E、timestamp、dependency、permission／prohibition。
- Replacement、prevention、玩家排序與 cause chain。
- State-based checks、Unique、死亡、destroy、deckout 與 game ending。
- 戰鬥、attack declaration、retaliation、同時傷害、On Hit 與 On Kill。
- Token、Generate、Mastery、Status、Intent 與固定牌組實際使用的 keywords。

## 規則裁定 Gate

| Issue ID | 是否由 Support Set 觸及 | 狀態 | 裁定來源／排除理由 | 對應測試 |
| --- | --- | --- | --- | --- |
| `RUL-001` | 是／否 | 待官方裁定／已解決／不適用 | `<待填>` | `<待填>` |

Issue 定義與處理流程見 [`rules-issues.md`](./rules-issues.md)。任何被觸及且仍待裁定的項目都會使矩陣保持 `blocked`。

## 實作順序

1. 起始 Level 0 Champion 及其取得起始手牌的 On Enter ability。
2. 被上述 Champion 直接依賴的內容與機制。
3. 能走通第一個正式 activation／Materialization 的最小卡牌。
4. 戰鬥與勝負需要的最小內容。
5. 其餘內容依 dependency graph 由葉節點向外完成；一次只推進一張卡或一個可觀察互動。

## 完成條件

- Deck Manifest 通過 Standard 格式與數量驗證。
- 從三個牌池出發重新計算閉包，不會發現矩陣外的可達內容。
- 每個內容、Ability Slot 與機制都為 `supported`，且至少有來源可追溯的正常與邊界測試。
- 沒有被觸及且仍待裁定的 issue。
- 正式 registry 可由矩陣生成或驗證一致結果；不能存在只在程式碼內宣告支援的卡牌。
