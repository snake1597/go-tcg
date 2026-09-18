# 05 — 交付 Combat、Intent 與 Effects Stack

**What to build:** 擴充完整 PlayerView，使 Vue 能直接呈現攻防關係、Intent、wielded Weapon 與 Effects Stack，而不需從事件或卡名推測。

**Blocked by:** 03 — 交付 Card Catalog、圖片與完整區域投影

**Status:** ready-for-agent

- [ ] Combat projection 明確指出攻擊者、防禦者、目標與目前狀態
- [ ] Intent projection 提供玩家可見且穩定的關聯身分
- [ ] wielded Weapon 以 ObjectRef／CardRef 連結正確場上物件
- [ ] 每個 Effects Stack item 有獨立 stack identity、順序與 Source Card reference
- [ ] Source Card 離場後 stack item 仍保有可呈現資訊
- [ ] 不同 seat 的 projection 不洩漏隱藏選擇或不可見來源
- [ ] Game Module scenario test 與 ConnectRPC contract test 覆蓋建立、更新及清空狀態
