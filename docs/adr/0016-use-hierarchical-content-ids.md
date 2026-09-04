# CardFace 與 Ability Slot 使用顯式階層式 ID

CardFace ID 固定為 `face:<card-id>:<face-key>`，Ability Slot ID 固定為 `ability:<card-id>:<face-key>:<slot-key>`；`card-id` 保留權威卡面資料的大小寫，`face-key` 與 `slot-key` 使用小寫 kebab-case，且由 registry 顯式配置。單面卡使用 `front`，雙面卡依實體卡面位置使用 `front`／`back`。Ability Slot 以規則語意命名，不使用卡名、效果文字、段落順序、hash 或印刷版本 UUID 自動產生，從而讓勘誤與排版變更不會意外改變身分，也讓多卡面能力不會碰撞。

同一規則能力的文字澄清或勘誤保留既有 slot；新增能力配置新 slot。能力被移除、拆分、合併或替換時，舊 slot 直接移除並配置新的 semantic key，同時提高卡面資料版本；不建立 alias、fallback 或 migration，舊 replay 只能由其釘選的舊內容版本讀取。Registry 必須拒絕未知 parent Card／CardFace、重複完整 ID、非小寫 kebab-case key，以及沒有恰好一個 Ability Slot 對應的 rules-bearing behavior。

曾考慮直接使用印刷版本 UUID，但同一規則卡面可有多個印刷版本；也考慮使用段落序號或效果文字 hash，但新增段落、排版或勘誤會使後續能力全部漂移。兩者都不能提供 registry、replay 與 ability lifetime 所需的穩定語意身分，因此不採用。
