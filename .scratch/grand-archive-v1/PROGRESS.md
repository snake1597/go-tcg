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

## 進行中：Issue 12

已新增 `internal/game/characteristics.go`，並將 attack legality、combat damage 與 state-based lethal checks 收斂到 `characteristicsFor`。Arthur rested-other-Allies +1 Power 與 Bulwark matching-class +1 Power 已透過此查詢實作且不回寫 printed stats。`go test ./internal/game` 已通過。

Arthur/Bulwark modifiers 目前：card class 資料、Arthur rested-Allies modifier、Bulwark class bonus 與可到期的 immortality state-based protection 已完成並有測試。

Bulwark wield payment 已完成：合法性與提交時皆要求兩點 reserve，付款會 banish。

尚未完成：Arthur On Enter 的 optional-rest transaction，以及完整 Layer A 至 E 排序與 duration。

Arthur 的 rest + current-turn immortality state transition 已封裝並測試；仍缺將它接到尚未存在的 Ally deployment On Enter choice。

Player View 的 Champion power/life 也已改由 `characteristicsFor` 顯示，並通過 `go test ./internal/game`。

完整驗證：`go test ./...` 與 `git diff --check` 通過。
