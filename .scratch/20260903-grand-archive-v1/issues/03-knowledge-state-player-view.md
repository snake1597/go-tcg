# 03: 以 Knowledge State 保護 Player View 與 View Handle

**What to build:** 讓所有操控端只能取得指定玩家依法可見的 Player View，並只能使用限定玩家、追蹤期間與 revision 的不透明 View Handle 提交行動或選擇。

**Blocked by:** 01: 建立版本化且可診斷的確定性 Game Module 基線.

**Status:** completed

- [x] Player View 不暴露 Card Instance、Object 或其他引擎內部 ID，也不洩漏對手手牌或牌庫順序。
- [x] 合法行動與待決選擇只包含該玩家當下可見且可選的 handle。
- [x] 過期、跨玩家、偽造或已撤銷 handle 均被拒絕且不改變正式狀態。
- [x] 洗牌或失去追蹤權會撤銷舊 handle；已公開事件仍可回顧但不能重新連結該實體。
