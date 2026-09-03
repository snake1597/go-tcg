# 卡面資料版本

## 權威來源

首版唯一的卡面資料來源是 repository 內的 `./card/*.json`。Game Module、測試及建置流程不得在執行期間向外部服務取得較新的卡面，也不得在版本不符時自動 fallback。

每張卡以 JSON 頂層 `uuid` 作為 Card ID；`editions` 與 `result_editions` 內的 UUID 只代表印刷版本，不作為 Card ID 或 CardFace ID。

## 固定版本

目前版本為 `card-data-v1`，由 repository 根目錄的 [`card-data-manifest.json`](../card-data-manifest.json) 固定。Manifest 記錄：

- schema 與資料版本。
- 固定來源 `./card/*.json`。
- 每個檔案的相對路徑、Card ID、slug、名稱及 `last_update`。
- 每個檔案的 SHA-256。
- 整份資料集的 SHA-256。

資料集 digest 依檔名排序後，對每張卡依序加入 UTF-8 檔名、NUL byte、原始檔案 SHA-256 bytes 與換行，再計算 SHA-256。`last_update` 只供來源追溯；內容是否一致以 digest 為準。

## 驗證與更新

驗證目前工作目錄沒有卡面漂移：

```bash
go run ./cmd/tool/card_manifest
```

有意更新 `./card` 後，明確提高 `-version` 並重新產生 manifest：

```bash
go run ./cmd/tool/card_manifest -write -version card-data-v2
```

更新必須與 manifest、Support Set、受影響測試及 replay 相容性變更一起 review。Manifest 驗證會拒絕無效 JSON、多份 JSON 串接、缺少必要 metadata、重複 Card ID／slug、檔名與 slug 不一致，以及任何未反映在 manifest 的內容變更。
