# 20: 以 Verita 與 Peppered Chef 支援複合費用和 On Death

**What to build:** 讓 Verita 與 Peppered Chef 透過逐步選擇完成精確總和 alternative cost、其他 Ally sacrifice、immortality、On Death 與跨回合 temporary power。

**Blocked by:** 10: 以 Impact Hammer 完成觸發收集與排序; 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器; 12.5: 建立統一 Ability 與 Effect Runtime; 14: 支援 Cardistry 的選擇、棄牌與部署互動.

**Status:** completed

- [x] Verita 只接受至少三張且 printed reserve cost 總和恰為十的 graveyard Suited cards，並原子 banish。
- [x] Verita 的 immortality 與 On Death Suited power 使用中央 evaluator、trigger 與正確跨回合 duration。
- [x] Peppered Chef 的 optional other-Ally sacrifice 只提供合法 handle，且犧牲與 temporary power 原子提交。
- [x] 取消、找不到精確組合、來源死亡及 duration 到期均不留下部分修改。
