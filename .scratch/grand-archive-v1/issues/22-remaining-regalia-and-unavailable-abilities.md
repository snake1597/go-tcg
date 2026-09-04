# 22: 完成固定牌組剩餘 Regalia 與不可啟動能力

**What to build:** 讓 Grand Crusader's Ring、Viridian Protective Trinket 及 Water／Wind Resonance Bauble 完成正式支援，包括牌組限制、自我 banish、抽牌、費用稅與固定鏡像條件下不可啟動的能力。

**Blocked by:** 06: 以 Tonoris 完成 Champion Lineage 與 Materialization; 07: 以 Blazing Throw 建立第一條完整卡牌宣告與 Stack 路徑; 12: 以 Arthur 與 Bulwark Sword 建立中央衍生特徵求值器.

**Status:** ready-for-agent

- [ ] Divine Relic restriction 在 Standard deck validation 階段執行，非法牌組無法開始。
- [ ] Grand Crusader's Ring 的 banish-and-draw 使用原子能力宣告與共用 zone operations。
- [ ] Viridian tax 只在 active-player、opponent 與 water-card 條件同時成立時修改 activation cost。
- [ ] 固定鏡像牌組沒有 water／wind Champion 時，兩種 Bauble ability 不會出現在合法行動中。
