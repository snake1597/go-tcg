# 自動實作進度

最後更新：2026-09-13

## 已完成

- Issue 07：Action 宣告、stack、Graveyard、Blazing Throw。
- Issue 08：Fiery Interference 的 fast timing、傷害與本回合 recover prohibition。
- Issue 09：Straight Flare 的 Suited Card Face printed reserve cost 查詢與動態傷害。
- Issue 10：wield EventBatch metadata、Impact Hammer On Wield LKI、單 trigger 自動入 stack、同 controller 多 trigger 的 Pending Choice 排序。

所有上述變更仍由使用者自行 staging/commit；工作樹中可見的變更不得覆蓋或重設。

## 已完成：Issue 11

已確認 Issue 11 的第一個縱切應從 `PlayerView -> attack declaration -> target choice -> Effects Stack -> Opportunity` 開始。現有 `internal/game` 已有：

- `championObject.Rested`、`CombatRole`、`Damage`；
- `fieldObject` 與 target handle 投影；
- Effects Stack、Opportunity、trigger ordering 與 replay/hash 基礎。

已完成：

- `PlayerView` 提供 attack 與 wield 的合法 handle，宣告後以 pending choice 選 target，並在提交時重驗時序、控制權與目標。
- combat stack item 在雙方 Opportunity 都讓過後，同時套用 attack damage 與 Retaliation，寫入具順序的 `on-hit:attack`／`on-hit:retaliation` EventBatch。
- state-based checks 以固定 object／player 順序逐輪處理，銷毀致命 Ally 至 owner Graveyard；Champion 死亡結束遊戲並在傷害事件後記錄 `combat:on-kill`。
- 32 輪保守上限會結束對局並透過 `PlayerView.Diagnostic` 保留診斷；每次提交的 replay step 照常附帶 canonical state hash。

驗證：`gofmt` 已執行，`go test ./...` 全數通過。

下一步：Issue 12（中央 derived-characteristics evaluator）。

暫停檢查點：Issue 12 尚未開始；可從其第一項「中央 evaluator 的 Layer A 至 E 與 power/life sub-layer」著手。

## 已完成：Issue 12 與 12.5

已完成中央 evaluator 與 unified runtime：

- `continuousEffect` 支援 Layer A-E、power/life sub-layer、dependency、dependency loop timestamp fallback、duration、instanced target snapshot 與 static predicate；Arthur、Bulwark、immortality、recover prohibition、combat、state-based 與 Player View 皆透過它重算。
- Action、Impact Hammer trigger 均建立有 controller、source LKI、fixed target 與 identity 的 `Ability Instance`，統一進 Effects Stack。
- typed operation 支援 choose continuation、move、draw、counter、damage、continuous modifier；choice 僅向 actor 揭露 View Handle，並能把剩餘 operations 重新入 stack。
- Action、wield 與 Materialization 費用均讀 central evaluator；完整 replay/hash 測試已更新並通過。

完整驗證：`gofmt`、`go test ./...` 與 `git diff --check` 通過。

下一步：Issue 13（Basic Cardistry cards）。
