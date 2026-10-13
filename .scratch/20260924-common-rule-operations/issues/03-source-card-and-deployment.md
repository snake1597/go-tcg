# 03 — Source Card 移動及 Ally 部署

**What to build:** Action 結算後的 Source Card 移動與 Ally 從 Effects Stack 或 Memory 部署，透過共通的來源與入場流程完成。

**Blocked by:** 01 — 玩家卡牌區域移動、棄牌與 Memory recollection。

**Status:** completed

- [x] Source Card 與 Stack Item 保有各自身分；來源離開 Effects Stack 時依規則移到目的區域或成為場上 Object。
- [x] Ally 部署建立正確 Object，並保留 Hindered、入場觸發與公開事件行為。
- [x] 來源在結算前失效時依既定規則處理，已付費用不被錯誤退回。
- [x] Player View、replay、state hash 與卡牌守恆情境驗證成功；被取代的來源移動旗標或分支已移除。

**Delivered slice:** `takeSourceCard` 集中 Effects Stack 與 Memory 的有效 Source Card 取得語意；Action 結算改由具名 `move_source_to_graveyard` operation 進墓地，Ally 則共用相同來源提交與既有入場流程建立 Object、Hindered、公開事件與 enter trigger。來源已不在指定來源區時不再進場或退回已付費用；正式 Player View、replay、state hash 與來源卡守恆情境皆已驗證。
