# 完整規則閱讀後的候選問題

本報告完整閱讀 `rules` 目錄下 108 份 Markdown 規則文件（約 51,500 字），基準為 commit `602c917f2f8fd4df7198429a72eb596bf7f647c6`。下方原始內容只列出可能影響首版計畫或既有 ADR 的問題，不把規則文字中的疑點自行解釋成 house rules；研究後形成的決策另記錄於「後續落地狀態」。

## 後續落地狀態

本節記錄研究完成後的決策去向；下方原始問題分析保留不改寫，方便追溯當時風險。

| 問題 | 狀態 | 落地位置 |
| --- | --- | --- |
| 1. 支援內容閉包 | 已決定 | [ADR 0002](../adr/0002-reject-unsupported-card-mechanics-before-play.md)、[卡牌與 Support Set](../card.md) |
| 2. Champion／ability identity | 已決定 | [ADR 0012](../adr/0012-separate-card-object-and-stack-identities.md)、[`CONTEXT.md`](../../CONTEXT.md) |
| 3. Knowledge／visibility | 已決定 | [ADR 0006](../adr/0006-expose-player-scoped-views-to-controllers.md) |
| 4. Scheduler／穩定停點 | 已決定 | [ADR 0010](../adr/0010-serialize-state-transitions-within-each-game.md) |
| 5. 事件批次與因果 | 已決定 | [ADR 0013](../adr/0013-preserve-event-batches-and-causality.md) |
| 6. Continuous／replacement evaluator | 已決定 | [ADR 0014](../adr/0014-centralize-derived-characteristics-and-replacements.md) |
| 7. Effects Stack source association | 已決定 | [ADR 0012](../adr/0012-separate-card-object-and-stack-identities.md) |
| 8. 宣告交易 | 已決定 | [ADR 0004](../adr/0004-use-guided-choices-and-atomic-actions.md) |
| 9. PendingChoice 中的 Concede | 已決定 | [ADR 0010](../adr/0010-serialize-state-transitions-within-each-game.md) |
| 10. Level 0 Champion 開局依賴 | 已納入流程 | [開發計畫里程碑 2](../development-plan.md#2-standard-開局與回合生命週期) |
| 11–15. 規則文字疑點 | 已處理 | [`rules-issues.md`](../rules-issues.md)；RUL-001 依 [ADR 0015](../adr/0015-retain-opportunity-until-the-holder-passes.md) 採用專案自訂裁定，RUL-002～004 已由官方來源解決，RUL-005 為非語意錯字 |

## P0：會影響核心模型

### 1. 「一副固定牌組」不是完整的支援邊界

牌組內的效果可能召喚 token、從遊戲外生成卡牌、取得 Mastery／Status、複製其他物件，或要求玩家指定任意合法卡名。正式支援清單因此可能必須涵蓋從牌組出發可到達的完整內容閉包，而不只是牌組中的 card ID。

特別是 Generate 仍要求玩家擁有對應實體卡，數位模擬需要決定「遊戲外卡池」如何提供及驗證；選牌後的覆蓋矩陣也必須遞迴列出 token、生成卡與其他非牌組內容。

來源：[Generate](../../rules/glossary/game-terms.md)、[Tokens](../../rules/game-mechanics/game-mechanics-tokens.md)、[Mastery](../../rules/game-mechanics/game-mechanics-mastery.md)、[Statuses](../../rules/game-mechanics/game-mechanics-statuses.md)、[Naming](../../rules/game-mechanics/game-mechanics-miscellaneous-topics/naming.md)

### 2. 目前四種身分仍不足以完整表示 Champion 與能力生命週期

既有 ADR 分開 `CardDefinition`、`CardInstance`、`GameObject` 與 `StackItem`，方向正確，但完整規則還要求：

- Champion level up 時，代表 Object 的頂層卡牌改變，但不是新 Object。
- Transform 改變有效卡面，但不是新 Object；新啟用的 static ability 取得新 timestamp。
- Champion Object、lineage 中每張卡與各自能力具有不同 timestamp。
- 有些能力的追蹤狀態屬於特定卡牌，有些如 Cascade 則屬於特定 Object。
- Intent 中的卡牌可以像 Object 一樣提供 static、activated 與 triggered abilities，但本身仍不是 Object。

因此可能需要明確的 `AbilityInstance`／effect identity，以及「Object 的目前代表卡／卡面」關係，不能假設一個 Object 永遠對應同一張頂層卡與同一組能力狀態。

來源：[Champion](../../rules/general-rules/general-rules-card-types/card-types-champion.md)、[Double-faced Cards](../../rules/general-rules/general-rules-card-characteristics/double-faced-cards.md)、[Ability Tracking](../../rules/game-mechanics/game-mechanics-abilities/abilities-ability-tracking.md)、[Abilities](../../rules/game-mechanics/game-mechanics-abilities/README.md)、[Continuous Effect Timestamps](../../rules/game-mechanics/game-mechanics-types-of-effects/types-of-effects-continuous-effects/README.md)

### 3. `PlayerView` 不能只靠「公開區域／自己的私有區域」投影

資訊權限是每位玩家、每張卡、每個時間點的狀態：公開卡轉為公開區域中的 face-down 卡後，曾看過它的玩家可能保留查看權；face-down 卡控制權改變時，原控制者與 owner 又可能失去查看權。Reveal 是暫時公開，但規則允許玩家記錄曾公開的資訊；shuffle 或隨機放回牌庫後則不能讓穩定內部 ID 洩漏新位置。

因此 Game Module 可能需要 knowledge／visibility 模型；`PlayerView` 也不能直接暴露跨 shuffle 穩定的 `CardInstanceID`，否則即使隱藏卡名仍能追蹤牌庫位置。

來源：[Public vs Private Information](../../rules/game-mechanics/game-mechanics-game-zones/game-zones-public-vs-private-information.md)、[Reveal](../../rules/glossary/game-terms.md)、[Ordering and Tracking](../../rules/game-mechanics/game-mechanics-miscellaneous-topics/ordering-and-tracking.md)

### 4. 引擎需要「推進到穩定點」的排程器，不只是單次 `Apply`

一個輸入之後可能連續發生：效果逐句結算、暫存觸發、完整效果結束、state-based checks、玩家排序同時觸發、Unique 犧牲選擇、Opportunity 轉移，再決定是否結算 Stack 頂端。State-based checks 本身也可能要求玩家選擇，並可能反覆產生新的狀態變更。

因此 `Apply` 的內部需要可暫停／恢復的 deterministic state machine 或 continuation，持續推進到下一個「需要外部玩家決定」或「單局結束」的穩定點。待決選擇的 owner 也不一定是回合玩家或 Opportunity holder。

來源：[State-based Checks](../../rules/game-mechanics/game-mechanics-miscellaneous-topics/state-based-checks-and-effects.md)、[Triggered Abilities](../../rules/game-mechanics/game-mechanics-abilities/abilities-triggered-abilities.md)、[Resolution](../../rules/game-mechanics/game-mechanics-playing-cards/playing-cards-resolution.md)

### 5. 扁平的 `[]GameEvent` 不足以驅動規則

「抽 N 張」由 N 個離散 draw events 組成，但它們仍屬於同一玩家行動，且相關觸發要等全部抽完才放入 Stack；效果中的多句指令之間不進行 state-based checks。另一方面，`whenever X` 與 `whenever one or more X` 對同一批事件的觸發數量不同，同時傷害也要先共同計算，再依來源與 replacement order 判定 On Hit／On Kill。

外部顯示仍可輸出 `[]GameEvent`，但內部可能必須保存 resolution frame、事件批次、因果來源與「何時 flush triggers」等結構，否則無法正確聚合或延後觸發。

來源：[Drawing Cards](../../rules/game-mechanics/game-mechanics-drawing-cards.md)、[Game Effects](../../rules/game-mechanics/game-mechanics-types-of-effects/types-of-effects-game-effects.md)、[Resolution](../../rules/game-mechanics/game-mechanics-playing-cards/playing-cards-resolution.md)、[Damage Step](../../rules/game-mechanics/game-mechanics-turn-order/turn-order-combat-phase/combat-phase-damage-step.md)

### 6. 衍生 characteristics 需要 layer／dependency evaluator

Card type、supertype、element、ability、power、life、durability、level、cost 與 play permission 都可能被 continuous effects 修改。規則要求 A–E layer、layer 內 timestamp、跨效果 dependency，以及 dependency loop 的順序；`can't` 又高於 `can`。直接把最終數值寫回 Object 容易留下過期狀態，也無法在來源失效時可靠還原。

Replacement effects 則另外要求受影響一方選擇順序、逐個重算仍可套用的 replacement，且被替代的 action 在部分規則觀察中仍視為已執行。若固定牌組碰到 static、prevention、replacement 或 type-changing，這會成為首版核心，而非單卡 handler 可以局部解決的小功能。

來源：[Continuous Effects](../../rules/game-mechanics/game-mechanics-types-of-effects/types-of-effects-continuous-effects/README.md)、[Replacement Effects](../../rules/game-mechanics/game-mechanics-types-of-effects/types-of-effects-replacement-effects.md)、[Permission](../../rules/game-mechanics/game-mechanics-miscellaneous-topics/permission.md)

## P1：會影響 Interface 與里程碑

### 7. Effects Stack 不是單純的 `[]StackItem`

被打出的 source card 位於 Effects Stack，但本身不依 Stack 順序排列；有順序的是 activation、Materialization、bestowment 與 ability instance。同一 source card 可以同時關聯原始 instance 與多個 copy instance，只有最後一個 instance 離開後，source card 才移往目的區域或成為 Object。Ability instance 又能在來源離場後獨立結算並使用 last-known information。

目前 `CardInstance` 與 `StackItem` 的分離可以保留，但必須明確加入 source association、instance count、copy semantics、controller snapshot 與 LKI，而不能把 Effects Stack 實作成只有一個 slice 的卡牌堆疊。

來源：[Playing Cards](../../rules/game-mechanics/game-mechanics-playing-cards/README.md)、[Card Activation](../../rules/game-mechanics/game-mechanics-playing-cards/playing-cards-card-activation.md)、[Effects Stack](../../rules/game-mechanics/game-mechanics-game-zones/game-zones-effects-stack.md)、[Resolving Abilities](../../rules/game-mechanics/game-mechanics-abilities/abilities-resolving-triggered-and-activated-abilities.md)

### 8. 「逐步選擇後原子提交」還需要正式的宣告交易模型

Activation 依序經過宣告、元素、cost parameters、modes、targets、合法性、cost calculation、payment 與最終 activation。支付 costs 後還可能新增 modes／targets 並重做部分步驟，但不重算 cost；memory payment 又包含先支付 Floating Memory、再隨機選取並同時 banish。若整個 play 最終非法，狀態要回復且不得產生 triggers；若合法後才 fizzle 或 negate，已付 costs 則不退還。

這要求 action draft 能在隔離狀態中試算 characteristics、cost、隨機結果與 replacement，並在 commit 時一次產生正確事件；不能只把 UI 已選參數組成 struct 後直接逐欄修改正式 state。

來源：[Card Activation](../../rules/game-mechanics/game-mechanics-playing-cards/playing-cards-card-activation.md)、[Costs and Memory](../../rules/game-mechanics/game-mechanics-playing-cards/playing-cards-costs-and-memory.md)、[Negate/Fizzle](../../rules/glossary/game-terms.md)

### 9. Concession 是序列化輸入模型的明確例外

Concession 是 special game action，可以在任何時間發生，不受 Opportunity 或其他 timing restriction 約束。因此即使引擎正等待某位玩家的 Pending Choice，其他玩家仍必須能投降。Interface 需要區分一般規則行動、回答待決選擇，以及任何玩家隨時可提交的全域 special game action。

來源：[Concession](../../rules/game-mechanics/game-mechanics-miscellaneous-topics/concession.md)、[Special Game Actions](../../rules/game-mechanics/game-mechanics-miscellaneous-topics/special-game-actions-and-turn-based-actions.md)

### 10. 正式開局依賴 Level 0 Champion 卡牌行為

起始手牌不是一條固定的通用 draw 規則，而是由各玩家起始 Level 0 Champion 的 On Enter ability 取得。在所有起始 On Enter abilities 依回合順序結算完成前，玩家也不能取得 Opportunity。

因此 Fixture 可以先走通開局骨架，但正式牌組的 Level 0 Champion 必須是第一批完成的正式卡牌；否則無法聲稱已完成真正的 Standard setup。卡牌覆蓋矩陣的依賴排序應把起始 Champion 放在最前面。

來源：[Starting the Game](../../rules/general-rules/general-rules-starting-the-game.md)、[Champion](../../rules/general-rules/general-rules-card-types/card-types-champion.md)

## P0：規則文字本身需要 ruling

以下不是單純的程式設計問題，而是目前快照內可直接觀察到的文字衝突或疑點。依既有 ADR，實作前應尋找官方 ruling；找不到時維持未支援，不應由程式碼暗自選邊。

### 11. Opportunity 的下一位擁有者不一致

- Timing 規則與 card activation／Materialization 規則說：執行者在行動後繼續取得 Opportunity，即使他不是回合玩家。
- Activated Ability 規則卻說：ability 支付 costs 並進入 Stack 後，Opportunity 先給回合玩家。
- Game Terms 的 Player Action order 也寫成先給回合玩家。

來源：[Timing and Permissions](../../rules/game-mechanics/game-mechanics-timing-and-permissions.md)、[Activated Abilities](../../rules/game-mechanics/game-mechanics-abilities/abilities-activated-abilities.md)、[Card Activation](../../rules/game-mechanics/game-mechanics-playing-cards/playing-cards-card-activation.md)、[Player Action](../../rules/glossary/game-terms.md)

### 12. 必要目標失效時的 fizzle 條件不一致

- State-based checks 說「所有必要目標」都非法或不存在時才 fizzle。
- Resolution 規則卻說任一必要目標非法時，整個 card／effect 都不結算。

這會直接改變多目標卡牌的結果，必須先取得 ruling。

來源：[State-based Checks](../../rules/game-mechanics/game-mechanics-miscellaneous-topics/state-based-checks-and-effects.md)、[Resolution](../../rules/game-mechanics/game-mechanics-playing-cards/playing-cards-resolution.md)

### 13. Level 0 Champion 的放置時間描述不一致

- Starting the Game 規則在第一回合開始前揭示並放置所有起始 Level 0 Champion。
- Champion 類型規則則寫成每位玩家的第一回合才放置。

來源：[Starting the Game](../../rules/general-rules/general-rules-starting-the-game.md)、[Champion](../../rules/general-rules/general-rules-card-types/card-types-champion.md)

### 14. Token 消失的精確時點不一致

- Token 機制頁說 token 被移到非 Field 區域時便 cease to exist，且不能再進行其他區域移動。
- Glossary 則說該次移動仍會完成 state-based checks、triggers 與必要 actions，到 Opportunity 傳給下一位玩家時才 cease to exist。

這會影響 On Leave、On Death、zone-change trigger 與後續效果是否還能引用 token。

來源：[Tokens](../../rules/game-mechanics/game-mechanics-tokens.md)、[Token Glossary](../../rules/glossary/game-terms.md)

### 15. Zone 數量文字有明顯計數疑點

Game Zones 總覽宣稱共有 9 個 zone，但同一句實際列出 Main Deck、Material Deck、Hand、Memory、Field、Graveyard、Banishment、Intent、Pantheon 與 Effects Stack 共 10 個。這多半是文件計數錯字，但 enum 與 Standard／Pantheon 可用 zone 的定義仍應明確。

來源：[Game Zones](../../rules/game-mechanics/game-mechanics-game-zones/README.md)

## 結論

現有高階方向仍成立：權威確定性 Game Module、序列化輸入、玩家視角、Go typed effects、fail-fast 支援清單與 vertical-slice TDD 都沒有被完整規則推翻。

原先需要重新確認的四組核心問題均已完成決策並落入 ADR、領域詞彙、開發計畫與測試策略，不需要繼續 grill。固定牌組、卡面資料、Content ID 與 Support Set dependency graph 由 [`card.md`](../card.md) 管理；剩餘內容工作是完成可執行 registry 與對應測試。實際被牌組觸及的規則 issue 已依 [`rules-issues.md`](../rules-issues.md) 處理；未來新增的未裁定處仍不得以未記錄的產品偏好替代官方語意。
