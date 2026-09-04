# 04: 以 Spirit of Fire 完成 Standard 開局

**What to build:** 透過正式 Game Module seam 建立雙人鏡像 Standard 單局，以 seed 確定洗牌，建立起始 Champion，並由 Spirit of Fire 的 On Enter 能力按 turn order 抽取起始手牌直到第一個穩定停點。

**Blocked by:** 01: 建立版本化且可診斷的確定性 Game Module 基線; 02: 載入固定牌組並實施 Support Set 開局 gate; 03: 以 Knowledge State 保護 Player View 與 View Handle.

**Status:** ready-for-agent

- [ ] 建立玩家、Card Instance、Champion Object、起始 zones、Knowledge State、scheduler frame 與 revision。
- [ ] 洗牌只使用版本化 PRNG，且 Spirit of Fire 的 On Enter 依規則順序各抽七張牌。
- [ ] 抽牌以具有批次、原因及確定順序的 committed events 表示並正確投影給各玩家。
- [ ] 開局抽牌遇到空牌庫時依法結束單局；相同 seed 的 setup 與 hash 可重播。
