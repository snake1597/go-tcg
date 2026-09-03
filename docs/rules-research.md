# 規則查證指南

本文件定義處理規則歧義時的最小查證流程。它不重述遊戲規則，也不取代 `rules-issues.md` 的逐案裁定。

## 來源優先順序

1. 官方規則庫目前文字。
2. 官方規則庫 `README.md` 的 changelog。
3. 直接修改相關規則的官方 Git commit、diff 與 blame。
4. 可穩定連結的官方 judge ruling 或官方公告。

當兩段同等權威的文字衝突時，只有較新的官方變更明確新增、修訂或澄清該爭點，才能採用其行為；否則維持 `待官方裁定`。不得以一般桌遊慣例或引擎實作補足規則。

若官方來源仍無法唯一決定行為，專案負責人可明確批准專案自訂裁定；此時必須另建 ADR，記錄採用行為、偏離的官方文字、替代方案及重新檢視條件，再將 issue 狀態改為 `專案自訂裁定`。

## 快速查證流程

1. 在 `rules-issues.md` 找到 issue ID、已查來源與下一步查詢。
2. 用 `rg -n -i '<關鍵字>'` 找到現行規則、glossary 與 changelog 的相關段落。
3. 對每個命中的規則檔案依序使用：

   ```sh
   git log --follow --format='%H %ad %s' --date=short -- <檔案>
   git show <commit> -- <檔案>
   git blame -L <起始行>,<結束行> <檔案>
   ```

4. 若本機歷史不足，再搜尋官方 GitHub、官方公告及 judge ruling；將精確 URL 與搜尋詞記錄回 issue。
5. 只有來源唯一決定行為時，才更新狀態為 `已由官方來源解決`，並記錄裁定日期、適用 commit、採用條文與可測試案例。

## GitNexus 的使用邊界

GitNexus 用於導覽與變更安全，而非取代全文與歷史查證：

- 查詢陌生的 Markdown 主題或引用關係時，先讀 `gitnexus://repo/gitbook-rules/context`，再用 query/context 定位檔案與標題。
- 編輯已索引的函式、類別或方法前，先做 upstream impact analysis；提交前執行 detect-changes。
- 規則矛盾的裁定仍以 `rg`、Git 歷史與官方來源為準。本 repo 目前沒有 execution flow，故不應為這個用途建立額外索引或自動化。

## 每案查證記錄

每個 RUL 項目保留以下資訊，避免重複掃描：

- 最後查證日期與規則 commit。
- 已查官方來源與使用過的搜尋詞。
- 下一次可執行的精確查詢或待追蹤官方來源。
