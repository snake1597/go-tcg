# PlayerView 與卡牌呈現合約

## 狀態與責任

本文件是所有操控端共同使用的玩家視角合約，整合卡牌內容參照、區域投影、Effects Stack、Combat／Intent 與 handle 生命週期。它由 Game Module 的 KnowledgeState 邊界產生，供 Vue、Bot、CLI 與未來其他 adapter 使用。

本文件不定義 Vue 版面、ConnectRPC transport、Declaration Draft 或規則本身。UI 行為見 [`vue.md`](../vue.md)，多步驟宣告見[Declaration Draft 與費用支付](./declaration-draft-and-cost-payment.md)，HTTP 邊界見 [ConnectRPC 玩家 API](./connectrpc-player-api.md)。

## 核心原則

- PlayerView 是指定玩家在指定 revision 可知資訊的完整權威 snapshot。
- Game Module 在投影時完成資訊裁切；transport 與 Vue 不得事後過濾完整 GameState。
- 內容身分、單局物件身分、可提交能力與 transport session 使用不同種類的 opaque identity。
- 所有 action、choice 與追蹤能力都受玩家、game、revision 與 handle 生命週期約束。
- 事件歷史不能用來重建目前區域、Stack、Combat 或隱藏資訊。
- 被新合約取代的名稱式、通用 `Cards` 與 `VisibleEvents` 欄位直接移除，不保留 fallback。

## 身分模型

### CardRef

每張依法可辨識的卡面以 `Card ID` 與目前 `CardFace ID` 組成 CardRef。CardRef 是版本化公開內容參照，不是單局卡牌實例，也不授權任何行動。

- 名稱、類型、規則文字、印刷 characteristics 與圖片從相同 `card_data_version` 的 Card Catalog 取得。
- Transform 保留 Object 身分，只更新 CardFace。
- Level Up 保留 Champion Object 身分，更新頂端 CardRef，舊頂牌進入 Inner Lineage。
- 卡牌離開 Field 後，使用目的區域規則要求的有效或預設 CardFace，不沿用 Vue 的舊猜測。
- 同名卡或同一 Card 的多個實例可以共用 CardRef，但必須具有不同 ViewHandle。

### ViewHandle

ViewHandle 是單一玩家視角中的短期 opaque identity。不同用途的 handle 不可互換：

- card handle：依法可追蹤的實體卡。
- object handle：Field 或 Champion 上的目前 Object。
- action handle：可提交的 Legal Action。
- choice／option handle：PendingChoice 與其選項。
- stack item handle：Effects Stack 中的 instance。
- group handle：Banishment tracked／temporary group。

Action 以 `LegalAction.Source` 關聯同一 PlayerView 中可見的 card 或 object handle，再以自己的 action handle 提交。PendingChoice 也必須明確關聯可見選項，不得以卡名、文字、陣列位置或 zone index 配對。

### 生命週期

- 物件或卡牌仍依法可追蹤時，可以跨公開區域沿用同一視角 handle。
- Transform 與 Level Up 保留 object handle；Object 離場即撤銷。
- 卡牌進入隱藏區域、洗牌、隨機插回或發生其他失去追蹤權的操作時，撤銷受影響 handle。
- 公開卡重新隱藏時，非 owner 的舊 handle 立即失效。
- 事件可以保留過去依法公開的 CardRef，但不能將其重新關聯到目前隱藏位置。
- 過期、跨玩家、跨 game、錯誤種類或已撤銷 handle 必須被拒絕，且不得改變 state、event 或 PRNG cursor。

## PlayerView 頂層內容

完整 snapshot 至少包含：

- revision、card data version、finished、winner、NeedsRuling／diagnostic。
- Turn Player、Turn Number、Phase、Opportunity Holder、Decision Player。
- 雙方 Champion／Lineage、Field 與公開區域。
- 視角玩家 Hand、雙方公開 zone counts。
- Main Deck、Material Deck、Memory、Graveyard 與 Banishment 投影。
- Effect Sources、Effects Stack、Combat 與 Intent。
- Legal Actions、PendingChoice 與所有結構化 availability。

PlayerView 不包含完整事件歷史；EventBatch 由玩家範圍 cursor API 提供。

## 卡牌資料版本與 Catalog

- 每份 PlayerView 必須帶 `card_data_version`。
- 可辨識卡牌只輸出 CardRef，不重複保存名稱與內容 ID 兩套真相。
- Vue 在呈現卡牌與提交 command 前，必須載入完全相同版本的不可變 Catalog。
- 版本缺失或不一致時阻擋 command，不可退回其他版本或以卡名猜測。
- 每個 CardFace 對應唯一 canonical image；建置時必須驗證 parent、重複 ID、圖片存在性與 digest。
- Catalog 只保存靜態內容；damage、counters、Rested、衍生 power／life 等 runtime 狀態由 PlayerView 提供。
- 隱藏牌背沒有 CardRef、card handle 或可跨 revision 追蹤的穩定佔位資料。

Catalog 的 HTTP 交付與 static asset 規則由 ConnectRPC 玩家 API 規格負責。

## 共通區域規則

- 每個區域都明確提供 owner、種類、權威順序及該視角可見內容。
- 空區域仍提供可判斷的空投影或公開張數，不製造虛構卡牌。
- 卡牌移動後只出現在新區域，不同區域不得同時殘留同一可見實例。
- Vue 不從初始牌表、事件差異、卡牌文字或 phase 推算目前內容或合法性。
- 瀏覽本身不建立事件、不增加 revision，也不代表卡牌可操作。

## Hand 與公開張數

- 視角玩家取得自己 Hand 的有序可見卡牌與 handle。
- 其他玩家只取得公開 Hand count，不取得 CardRef、順序或佔位 handle。
- 雙方公開 counts 以玩家為單位，不以 `self`／`opponent` 固定欄位寫死座位。
- Main Deck 與 Material Deck count 隨下一個 revision 更新；Memory count 由 Memory 投影的 total count 唯一負責。

## Main Deck

- 雙方一般只看見剩餘張數與統一牌背；owner 也不能任意查看內容。
- 非空顯示牌背與張數，空牌組只顯示 `0`。
- 規則授權查看時，只投影指定範圍；需要決策時使用 PendingChoice，純查看則使用唯讀 reveal projection。
- reveal 中每張可見卡使用該玩家專屬 handle 與 CardRef。
- 洗牌或失去追蹤後不得保留舊 index、位置或可回連 CardRef 的資料。
- Main Deck 隨機化只由 Game Module PRNG 執行。

## Material Deck

- owner 可依 deck manifest 固定順序查看所有剩餘卡牌。
- 對手只看見牌背與剩餘張數，不取得內容、順序或 handle。
- 所有卡牌都可瀏覽；只有具有對應 Legal Action 的卡可進入 Materialization。
- `LegalAction.Source` 唯一指向 Material Deck 中的 card handle，action handle 與 card handle 不可互換。
- 同名實例各自具有 card handle 及其 action handle。
- 卡牌離開 Material Deck 後，下一個 revision 移除該項目並更新張數。

## Memory

- 每位玩家固定具有一個 Memory projection，包含 owner、total count 與目前視角可辨識的 cards。
- owner 看見自己全部 Memory；非 owner 只看見總張數及規則明確公開的卡。
- 可見項目標示是否對所有玩家公開；未公開項目不得用空 handle 或穩定位置佔位。
- owner 順序依權威 Memory 順序；非 owner 的公開卡保留相對順序，但不暴露中間隱藏卡位置。
- 一般 memory cost 由引擎隨機 banish，Memory projection 本身不授權玩家選牌。
- Recollection 由引擎處理，不提供 `Recollect All` 行動。
- Hand 與 Memory 間若 owner 持續具追蹤權，可沿用 owner handle；非 owner 在卡牌重新隱藏時失去 handle。

## Graveyard 與 Banishment

- 每份 PlayerView 固定包含雙方各一個 Graveyard 與 Banishment，即使為空。
- 同一玩家先 Graveyard、後 Banishment；玩家順序依座位。
- Graveyard 的 regular group 依進入時間由新到舊。
- Banishment 的一般牌位於 regular group；特定效果追蹤或暫時 banish 分別位於 tracked／temporary group。
- tracked／temporary group 提供 group handle、穩定 label code，以及依法公開時的來源 CardRef。
- group 與卡牌順序由引擎提供，Vue 不合併、不以卡名重排。
- 雙方看見相同公開內容，但各視角 handle 可以不同。

## Champion、Lineage 與 Field

- Champion 與每個 Field Object 都具有 object handle 與目前 CardRef。
- Field 提供穩定權威順序、owner、controller、orientation、damage、counters 及必要 derived characteristics。
- Awake、Rested 與其他規則狀態是引擎輸出，不由 Vue 根據動畫或事件推算。
- Field card handle 與 object handle 代表不同概念，不可互換。
- Lineage 顯示目前公開內容與順序；Level Up 不呈現為 Champion 離場再進場。

## Effects Stack

Effect Source Card 與 Stack Item 是兩種資料：

- Effect Sources 依首次進入 Effects Stack zone 的時間排序，並帶可見 CardRef、card handle 與 controller。
- Effects Stack 陣列直接使用結算順序，第一項永遠是下一個結算項目。
- Stack Item 有自己的 handle、kind、controller，並以 source handle 引用 Effect Source。
- 多個 Stack Items 可共用一張 Source Card；沒有 Card Source 的 ability 可以省略 source。
- Source Card 不是 Stack Item；不得複製名稱或以名稱關聯。
- Item 結算後從下一個 view 消失；Source 直到最後一個關聯 instance 離開後才依法移往目的區域。
- Stack 可見不代表可操作；只有 Legal Action 或 PendingChoice 能產生提交。

## Combat 與 Intent

- 無 active combat 時不輸出 Combat projection。
- active combat 明確提供正式 combat stage、controller、attacker、defender、Intent、wielded Weapons、目前 attack power 與 retaliation power。
- attacker／defender 引用同一 PlayerView 中存在的 Champion 或 Field object handle。
- Intent 依權威順序投影公開 card／copy；copy 具有可見身分但不偽造 CardInstanceID。
- wielded Weapon 只引用 Field object，不複製第二個 Object。
- power 是目前 revision 的 derived 值，不保證等於最終 damage。
- 參與者離場造成 fizzle 或 combat 結束時，下一份 view 更新或移除 Combat，不保留幽靈 handle。

## Availability 與提交邊界

- 可見但不可用的 action、card、target 或 payment source 提供穩定 reason code 與非敏感參數。
- Legal Action 與 PendingChoice 是提交能力的唯一來源；卡牌出現在任何區域不代表可操作。
- Game Module 在提交時重新驗證 actor、revision、handle、來源與目前合法性。
- 拒絕非法提交時，canonical state、event、trigger、Replay 與 PRNG cursor 都不改變。
- JSON／Protobuf 投影不得包含 CardInstanceID、ObjectID、StackItemID、AbilityInstanceID 或其他內部 runtime ID。

## 驗收與測試

至少驗證：

- 每個區域的空、非空、owner／非 owner 與公開／隱藏差異。
- 同名卡、多實例、Transform、Level Up 與跨公開區域移動。
- 洗牌、重新隱藏及失去追蹤後撤銷 handle。
- Materialization action、PendingChoice 與可見卡牌的唯一關聯。
- Banishment regular／tracked／temporary groups 與順序。
- 單一／多項 Stack、共用 Source、無 Card Source 與逐項結算。
- Champion／Ally 攻擊、不同 target、Intent、wielded Weapons、derived power 與 fizzle。
- 各 seat 公開內容一致且 handle 隔離，隱藏資訊不出現在序列化輸出。
- PlayerView／Catalog 版本一致、canonical image 唯一性與錯誤版本阻擋。
- JSON／Protobuf round-trip 與既有 replay verification 不洩漏內部 ID。
