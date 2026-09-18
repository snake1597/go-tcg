# 以 ConnectRPC HTTP API 連接獨立 Vue 專案

Vue 對局介面位於另一個獨立專案，不加入 `go-tcg` repository，也不由 Go binary 編譯或嵌入其靜態資源。`go-tcg` 負責遊戲引擎、應用服務、Protobuf API 合約，以及以 ConnectRPC 暴露的 HTTP endpoint；Vue 使用由同一份 Protobuf 合約產生的 TypeScript client，直接透過 HTTP 呼叫 Go 服務。

`.proto` 由 `go-tcg` 單一擁有。如何取得 schema、產生 TypeScript client、保存產物與鎖定版本，由 Vue 專案的使用者自行處理，不屬於 `go-tcg` 的建置或發布責任。瀏覽器使用產生出的 service descriptor，搭配 `@connectrpc/connect` 與 `@connectrpc/connect-web` 的 Connect transport；不直接操作低階 gRPC-Web frame。

Game Module 不依賴 ConnectRPC、HTTP 或 Vue。Transport handler 只能呼叫應用層介面，將 Protobuf request 轉成命令，並將玩家專屬 `PlayerView` 轉成 response；不得把完整 `GameState` 或 transport 型別傳入規則引擎。CLI、bot 與 ConnectRPC adapter 共用同一套應用層與玩家視角規則。

只維護 Connect protocol 這一條瀏覽器遠端 API 路徑，不另建功能重疊的 REST、WebSocket 或傳統 gRPC-Web gateway 相容層。瀏覽器所需的 unary 與 server-streaming 能力均由 ConnectRPC HTTP endpoint 提供。

## 曾考慮的方案

把 Vue 放入同一 repository 並由 Go 服務靜態託管，能減少本機啟動步驟，但會把前端發布週期、工具鏈與後端模組混在一起，違反已確認的獨立專案邊界。另建 REST 或 WebSocket gateway 會產生第二份傳輸合約及錯誤語意，因此不採用。BSR 或 npm SDK 發布可以協助跨 repository 交付，但不是 Go API 運作所必需，且使用者已決定自行處理 client 產生，因此不納入本專案要求。
