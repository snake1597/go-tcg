# 10 — 複製能力與 runtime 身分

**What to build:** Duchess 的選牌與 Action 複製由通用定義及操作完成；玩家能合法選擇、略過或結算複製品。

**Blocked by:** 04 — 具名選擇與可序列化續行；05 — Activated ability 的費用與使用限制；共通規則操作 08 — Duchess 選牌與複製 Action。

**Status:** ready-for-agent

- [ ] 複製品具獨立 Ability Instance identity、正確 Source Ref／LKI，且不再支付原 Action 費用。
- [ ] 正式介面驗證原牌移動、目標失效、pass、續行與 source cleanup；卡牌守恆及 replay hash 確定。
- [ ] Duchess 專用 operation kind 與 Card ID 分支已刪除。
- [ ] 通過完整 repository quality gate；Standard Game、Support Set、replay 與鏡像對戰無回歸。
