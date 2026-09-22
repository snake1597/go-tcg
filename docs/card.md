# 卡牌資料與固定 Standard 牌組

首版只會建立一種固定的鏡像 Standard 對局。卡牌行為由 Game Module 的 Go 程式碼執行；此文件只說明載入資料與建局的最小關係，不維護另一份 runtime 支援清單。

## 資料來源

- `card/*.json` 是卡面資料來源。
- `card-data-manifest.json` 固定資料版本、檔案清單與 SHA-256 digest。
- [deck.go](../internal/game/deck.go) 的 `fixedStandardDeck` 是唯一牌組定義。

每次 [NewStandardGame](../internal/game/support.go) 建局時，會讀取 manifest、驗證資料集 digest、載入每張卡面，並確認固定牌組引用的 Card ID 與 CardFace 存在。

## 固定牌組的責任劃分

runtime 只保護「無法安全建立對局」的條件：

- 兩名玩家存在且 UID 不同。
- 卡面檔案與 manifest 相符。
- 牌組中的 Card ID 與 CardFace 可解析。
- Material Deck 有且只有一名 Level 0 Champion。

固定牌組的格式規則屬於開發時的測試責任，而不是每局重跑的 gate：60 張主牌組、12 張 Material Deck、卡片張數上限、Divine Relic 限制與固定牌組版本都由 `internal/game` 測試覆蓋。

## 目前建局流程

```text
NewStandardGame
├── 驗證玩家
├── 載入並驗證 card-data manifest
├── 取得 fixedStandardDeck
├── 驗證卡片引用與起始 Champion
└── newStandardSetup
    ├── 建立卡片實例與 zones
    ├── 洗主牌組
    ├── 處理起始 Champion 的 On Enter
    └── 啟動 scheduler
```

## 未來接受不同牌組時

現在不預先建立 validator interface。當第一個外部牌組輸入出現時，才在輸入邊界加入該格式的驗證函式，例如 `ValidateStandardDeck(deck, definitions)`；若後續格式規則不同，再各自實作獨立驗證。遊戲初始化只接收已可建立的牌組，避免讓格式規則、檔案載入與 scheduler 交纏。
