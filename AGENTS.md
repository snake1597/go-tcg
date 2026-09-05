# AGENTS.MD

## 概述

go-tcg 是一個實現tcg能全自動遊玩的專案。

## 技術棧

- Language: Golang
- Database: Postgresql，Redis
- API: gRPC

## 主要目錄結構

```sh
go-tcg/
├── cmd/                         # 應用程式進入點
│   └── server/                  # API Server 主程式
│   └── tool_executor/           # 本地執行小工具
├── config/                      # 配置相關
│   ├── config.go                # 配置管理
├── database/                    # 資料庫相關內容
│   ├── migration/               # DDL 資料庫遷移
│   └── samples/                 # Mock 資料
├── docs/                        # 文件與規格
│   ├── api/                     # API 規範文件連結
│   │   └── README.md            # 文件說明
│   ├── architecture/            # 架構設計文檔
│   │   └── README.md            # 文件說明
│   └── README.md                # 文件說明
├── internal/                    # 專案內部邏輯與工具
│   ├── enums/                   # 列舉與常數定義
│   ├── model/                   # 資料模型
│   ├── entity/                   # 資料模型
│   ├── repo/                    # Repository 資料訪問層
│   │   ├──     ⦙
│   │   └── ticket.go            # 門票資料邏輯
│   └── service/                 # 核心業務邏輯
│       ├──     ⦙
│       └── ticket_client.go     # 會員端門票業務邏輯
├── pkg/                         # 可重用的工具模組
├── test/                        # 測試相關模組
│   ├── helper/                  # 測試輔助工具
│   ├── integration/             # 整合測試
│   └── setup/                   # 測試環境設定
├── transport/                   # 傳輸層 (gRPC 或 HTTP)
├── .gitignore                   # Git 忽略規則
├── create_migrate.sh            # 資料庫遷移腳本生成工具
├── go.mod                       # Go Modules 定義檔
├── go.sum                       # Go Modules 依賴檔
├── app.go                       # 主程式執行點
└── README.md                    # 專案說明
```

## 程式碼風格

### 命名慣例
- 檔案名稱：snake_case（例：labor_order.go）
- 類別名稱：PascalCase（例：LaborOrder）
- 函數名稱：公開的使用 PascalCase，私有的使用 lowerCamelCase
- 常數：公開的使用 PascalCase，私有的使用 lowerCamelCase
- 專有名詞（如 AppSDK）的命名方式以 Code Review 時的判斷為主，遇到特別奇怪的情況再提出討論

### 其他慣例
- 每次實作、修改或新增檔案前，Agent 必須重新讀取完整 `AGENTS.md`；未完成讀取不得修改檔案。
- YAML 測試資料不得使用 inline map；每筆資料的欄位需逐行展開，以維持可讀性與方便檢視差異。
- struct literal、包含匿名函式或巢狀呼叫的函式呼叫不得 inline；即使只有一個欄位或參數，欄位、參數與匿名函式 body 均須逐行展開。

## 安全規範

### 絕對禁止
- 在程式碼中寫死 API key、密碼、token
- 直接串接 SQL 查詢（SQL Injection 風險）
- 未驗證的使用者輸入直接使用

## 當遇到不確定的情況
如果不確定該怎麼做，請：
1. 先暫停，不要猜測
2. 詢問使用者
3. 參考專案中已有的類似實作

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **go-tcg** (1383 symbols, 2814 relationships, 32 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> If any GitNexus tool warns the index is stale, run `npx gitnexus analyze` in terminal first.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `gitnexus_impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `gitnexus_detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `gitnexus_query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `gitnexus_context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `gitnexus_impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `gitnexus_rename` which understands the call graph.
- NEVER commit changes without running `gitnexus_detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/go-tcg/context` | Codebase overview, check index freshness |
| `gitnexus://repo/go-tcg/clusters` | All functional areas |
| `gitnexus://repo/go-tcg/processes` | All execution flows |
| `gitnexus://repo/go-tcg/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
