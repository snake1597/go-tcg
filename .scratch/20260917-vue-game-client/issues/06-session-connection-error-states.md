# 06 — 完成 Session、連線與錯誤狀態 UI

**What to build:** 以 scripted fixtures 完成 command processing、暫時斷線、不可重試錯誤、session replacement、NeedsRuling 與 Game Finished 的前端狀態機和使用者回饋。

**Blocked by:** 04 — 完成 Declaration 與 Pay Cost UI；05 — 完成 Combat、Stack、事件與動畫

**Status:** ready-for-agent

- [ ] accepted revision 等待期間顯示對局覆蓋層並禁止重複 command
- [ ] scripted reconnect 使用帶 jitter 的 0.5、1、2、4、8 秒退避狀態，最高 10 秒
- [ ] 認證、版本不相容與 NotFound 情境進入不可重試畫面
- [ ] 重連情境會清除 Draft、套用最新 snapshot 並從保存 cursor 補事件
- [ ] session_replaced 會停止 command 與自動重連並顯示接管訊息
- [ ] 頁面 loading、阻擋錯誤、對局覆蓋層與局部錯誤彼此分離
- [ ] 局部圖片或 Draft 錯誤不清空仍有效棋盤
- [ ] NeedsRuling 與 Game Finished 有明確、不可繼續操作的結束狀態
- [ ] store test 與 fixture-based Playwright 覆蓋所有 scripted lifecycle 情境
