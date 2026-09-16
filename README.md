# go-tcg

## 概述

go-tcg 是 Grand Archive 的確定性規則引擎。首版產品提供一名真人透過 production CLI，使用固定 Standard 牌組與只讀取自身 `PlayerView` 的啟發式 bot 完成鏡像單局，並輸出可驗證的私人 canonical replay。

## 執行 production CLI

在 repository root 執行：

```sh
go run ./cmd/production_cli \
  --seed 7 \
  --submission-limit 1000 \
  --replay-out ./game.replay.json
```

- `--replay-out` 必填；檔案會以 `0600` 權限覆寫，且可能包含完整隱藏資訊，不可公開分享。
- `--seed` 預設為 `1`，相同版本、seed 與輸入序列會得到相同 replay state hash。
- `--submission-limit` 預設為 `1000`，必須是正整數。超限時 CLI 會停止單局並輸出 seed、step、診斷、replay 路徑與最終 state hash。
- 所有 action、target、choice、Reserve 與 Floating Memory 都使用引擎 `PlayerView` 提供的編號；格式錯誤、重複或越界選擇會被拒絕並重新提示。

## 驗證

發布級測試命令與 fuzz 時限見 [`docs/testing.md`](./docs/testing.md)。固定牌組、Support Set、CardFace 與 Ability Slot 覆蓋狀態見 [`docs/card.md`](./docs/card.md)。規則基準與裁定見 [`docs/rules-issues.md`](./docs/rules-issues.md)。
