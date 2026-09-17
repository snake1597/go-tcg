# 06 — 交付重連、接管與分層錯誤 UX

**What to build:** 讓玩家在 command 等待、暫時斷線、不可重試錯誤及另一分頁接管時得到一致回饋，且過期分頁不再提交或自動重連。

**Blocked by:** 04 — 交付 Declaration 與 Pay Cost 介面；05 — 交付 Combat、Stack、事件與動畫；Go Player API 07 — 完成 Command 與連線生命週期

**Status:** ready-for-agent

- [ ] accepted revision 等待期間顯示對局覆蓋狀態並禁止重複 command
- [ ] 暫時錯誤使用帶 jitter 的 0.5、1、2、4、8 秒退避，最高 10 秒
- [ ] 認證、版本不相容與 NotFound 錯誤不自動重試
- [ ] 重連會清除 Draft、取得最新 snapshot 並從保存的 cursor 補事件
- [ ] session_replaced 後舊分頁停止 command 與自動重連
- [ ] 頁面 loading、阻擋錯誤、對局覆蓋層與局部錯誤彼此分離
- [ ] 局部圖片或 Draft 錯誤不清空仍有效棋盤
- [ ] store test 與 Playwright 覆蓋暫時斷線、永久錯誤及雙分頁接管
