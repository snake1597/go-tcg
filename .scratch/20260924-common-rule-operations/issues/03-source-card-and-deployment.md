# 03 — Source Card 移動及 Ally 部署

**What to build:** Action 結算後的 Source Card 移動與 Ally 從 Effects Stack 或 Memory 部署，透過共通的來源與入場流程完成。

**Blocked by:** 01 — 玩家卡牌區域移動、棄牌與 Memory recollection。

**Status:** blocked

**Blocked reason:** 01 — 玩家卡牌區域移動的共通提交語意尚未完成；本票會在其上增加 Source Card、Stack Item 與 Field Object 生命週期。

- [ ] Source Card 與 Stack Item 保有各自身分；來源離開 Effects Stack 時依規則移到目的區域或成為場上 Object。
- [ ] Ally 部署建立正確 Object，並保留 Hindered、入場觸發與公開事件行為。
- [ ] 來源在結算前失效時依既定規則處理，已付費用不被錯誤退回。
- [ ] Player View、replay、state hash 與卡牌守恆情境驗證成功；被取代的來源移動旗標或分支已移除。
