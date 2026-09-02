# 採用模組化單體與深層 Game Module

首版建置為單一 CLI 執行檔，核心是一個提供小型 Interface 的深層 Game Module；Opportunity、效果堆疊、狀態檢查、選擇流程、戰鬥與卡牌效果都隱藏在其 Implementation 中。CLI 與 bot 作為操控端 Adapter 依賴該 Module；內部可依領域能力分檔或分 package，但不把內部細節擴散到呼叫端。Clean Architecture 在本專案表示依賴方向與職責隔離，不要求每個 entity 都具備 repository、use case、DTO 與 mapper，也不預先拆分微服務或只有單一 Adapter 的假想 seam。

## 曾考慮的方案

固定建立多層資料夾、為每個 entity 配置 Interface，能在檔案結構上呈現常見的 Clean Architecture 外觀，但會產生大量淺層轉呼叫，增加呼叫端與測試需要理解的表面，卻沒有隱藏規則複雜度。
