# 集中處理衍生特徵與替代效果

卡牌與 Object 的基礎資料保持不可變或只保存原始 runtime 狀態；所有衍生 characteristics 由 Game Module 的中央 evaluator 查詢時計算。只要固定 Support Set 使用 continuous effects，evaluator 就實作官方完整排序語意：Layer A 至 E、Layer E 的 power／life sub-layer、同 layer 的 dependency、dependency loop 與 timestamp。所有合法性、目標、費用、戰鬥及 PlayerView 顯示共用同一套結果；靜態 `can't`／`may not` 高於 `can` permission。

Replacement effects 使用獨立 pipeline，不混入 continuous layer evaluator。Action 或 Event 提交前找出目前適用的 replacements；多個候選依受影響卡牌、Object、zone 等的控制玩家決定順序。每套用一項後重新計算其餘候選，直到沒有適用項。需要外部排序時沿用所在流程的選擇協定：DeclarationTransaction 內回傳 draft-scoped ChoiceRequest，已提交的結算流程則建立 PendingChoice。原始意圖、每次替代與最終 committed events 保存 cause chain。

目前不為完整卡池預先實作所有 modifier 與 replacement operation。固定牌組確定後，卡牌覆蓋矩陣列出 Support Set 可達的 characteristic modifiers、static effects、permissions、preventions 與 replacements；只加入這些操作，但共用上述官方排序核心。Registry 若偵測卡牌要求尚未支援的操作，開局直接拒絕，單卡 handler 不得繞過中央 evaluator 或 replacement pipeline。

## 曾考慮的方案

把最終 power、life、type 或 permission 直接寫回 Object，單張卡看似容易實作，但來源失效時難以可靠還原，也無法處理 layer、dependency 與 timestamp。另一極端是固定牌組尚未選定前完成整個卡池所需的所有操作，會延後第一個可玩縱切並產生大量未被測試需求驅動的抽象；因此採用官方通用排序核心與 Support Set 驅動的操作集合。
