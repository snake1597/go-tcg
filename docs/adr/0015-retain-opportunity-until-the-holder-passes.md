# Opportunity 由持有者保留至讓過

RUL-001 的鎖定官方規則文字互相衝突：Timing and Permissions 與 card activation／materialization 表示行動者取得並保有 Opportunity；Activated Abilities 與 Player Action glossary 則寫成先交給回合玩家。專案負責人已明確確認本專案採用前者。

每次成功完成一個需要 Opportunity 的玩家行動後，在 state-based checks 與必要強制流程完成後，**該行動的玩家**取得 Opportunity；只要他沒有讓過，就可以繼續採取合法的 fast 玩家行動，包括啟動卡牌與 activated ability。持有者讓過後，Opportunity 才依 turn order 交給下一位玩家。所有玩家連續讓過完整一輪後，才結算 Effects Stack 頂端或在 Stack 為空時推進規則流程。

此決策同樣適用於非回合玩家在回合外啟動 activated ability：該玩家不會因 ability 進入 Effects Stack 而自動把 Opportunity 交回回合玩家。獨立產生的 triggered ability、phase／step 起始與效果結算後的 Opportunity 指派，仍遵循各自明確的官方規則；本 ADR 不把玩家行動的保有規則擴張到那些事件。

這是針對 RUL-001 的專案自訂裁定，偏離 `Activated Abilities`「Opportunity is then first given to the turn player」及 Player Action glossary 的衝突表述；若官方發布可唯一決定此行為的更正或 ruling，應重新檢視本 ADR 與 RUL-001。引擎須以情境測試覆蓋：非回合玩家啟動 fast card、materialize card 與 activated ability 後，在其讓過前皆仍是 Opportunity holder。

## 曾考慮的方案

每次玩家行動後一律改由回合玩家先取得 Opportunity，雖符合兩段衝突文字，卻會否定 Timing and Permissions 的「行動者預設取得 Opportunity」以及 card activation／materialization 的明確條文，並使玩家無法在不讓過的情況下連續採取 fast 行動，因此不採用。
