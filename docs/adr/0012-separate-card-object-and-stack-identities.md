# 分離卡牌、場上物件、能力與堆疊項目的身分

引擎分別表示不可變的 `CardDefinition`、單局中的 `CardInstance`、只存在於場上的 `GameObject`、runtime 的 `AbilityInstance` 與 `ContinuousEffectInstance`，以及效果堆疊上的 `StackItem`，每種實體都使用自己的穩定 ID。一般卡牌進場時建立 GameObject，離場即結束該 Object lifetime，日後重新進場會取得新 Object ID；但 Champion 在同一條 Lineage 存續期間保持 Object ID，Level Up 只更換頂端代表卡，Inner Lineage 中各卡保留 CardInstanceID 而不是 Object。Transform 只切換有效 CardFaceID，也不改變 ObjectID。Object 的 orientation、counter、戰鬥角色及 continuous effects 依規則保留；真正離場時才建立 last-known information。目標與關聯以 ID 表示，不讓卡牌 handler 長期持有 Go pointer。

`AbilityDefinition` 位於 CardDefinition 或 CardFace，只保存不可變規則。AbilityInstance 具有自己的 ID、宿主、作用範圍與 lifetime：卡牌範圍的狀態以 CardInstanceID 與 ability slot 識別；Object 範圍的狀態以 ObjectID 與 ability slot 識別，離場再進場即重置；Mastery、Status 等玩家層級能力則擁有獨立 runtime identity。這使「每回合一次」等限制明確歸屬於規則指定的 lifetime，並讓 Champion Level Up 後應保留或重置的狀態依能力作用範圍決定。

啟動或觸發能力進入 Effects Stack 時會建立獨立 StackItem，快照來源參照、操控者、模式、目標及必要的 characteristics 或 last-known information；即使來源之後離場，StackItem 仍可依法結算。Static ability 不進入 Effects Stack，而是貢獻帶有來源、時間戳、期間與相依資訊的 ContinuousEffectInstance；衍生 characteristics 由規則層計算，不永久改寫基礎資料。延遲觸發與 reflexive trigger 也建立獨立 runtime instance。卡牌 handler 只能呼叫引擎提供的規則操作，不得直接修改 GameState。

Effects Stack zone 內還需區分「來源卡」與有序 StackItem。打出卡牌時，CardInstance 成為具有 timestamp、但不參與 FILO 排序的來源卡；原始 activation／Materialization／bestowment 與所有 copy 各有 StackItemID，透過 source association 指向該卡。來源卡要等最後一個關聯 item 離開後，才由 state-based checks 移往預設區域或成為 Object。來源卡先離開會使仍引用它的 card-play instances 依法 fizzle。Ability StackItem 的來源卡或 Object 不搬入 Effects Stack，改以 SourceRef 與必要 LKI 表示。

## 曾考慮的方案

以單一 Card struct 同時代表卡面、實體卡、場上物件、能力狀態與堆疊 activation，初期型別較少，但無法可靠區分區域變更後的新物件、能力狀態的正確 lifetime、堆疊來源、目標失效、持續效果與 last-known information，容易產生跨區域的陳舊參照及永久殘留的衍生數值。把來源卡本身直接放進有序 StackItem slice，也無法表示一張來源卡同時具有原始 instance 與多個 copy，因此不採用。
