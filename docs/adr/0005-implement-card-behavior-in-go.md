# 以 Go 與版本化 DSL 編寫卡牌行為

卡牌以型別安全的 Go 編寫資料註冊，並用抽牌、傷害、目標選擇與觸發能力等可重用效果元件組合常見行為；特殊卡牌仍只能透過規則引擎提供的操作改變單局狀態。第一個 DSL vertical slice 將已完整驗證的 Wonderland's Reign 序列化為 schema version 1 的 JSON，於 content loading 階段嚴格解碼後交給既有 compiler。它不解析自然語言卡面、不新增 selector、expression、cost 或 operation kind，也不建立另一套 runtime。

DSL schema、compiled content 與 replay format 各自明確版本化。未知欄位、錯誤型別、未知原語、未使用欄位或不支援版本都會附帶 Definition、Face、Ability Slot 與欄位脈絡後拒絕載入；不讀舊 schema、不 migration、不 fallback 到 Go Definition。編譯完成後，Player View、合法操作、resolution 與 replay 只消費同一份 immutable compiled definition。

## 曾考慮的方案

直接為完整卡池建立 DSL 會同時擴張 schema、驗證與除錯面積，因此不採用。先只序列化一張既有 Go Definition，證明同一 compiler、Ability Slot、Player View 與 replay 契約可端到端運作；後續卡牌只能在既有 compiled IR 能完整表達時逐張遷移。Lua、JavaScript 或 Starlark 等腳本語言仍不採用。
