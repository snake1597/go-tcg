---
name: implement-ticket
description: use the grill-me implement skill and automatically review completed targets
---

## 核心工作流程 (必需嚴格執行)

### 階段一：使用skill
使用 grill-me `/implement` skill

### 階段二：檢查issues
確認issues裡的項目是否都完成
完成的要標註
並將status改為 `completed`

### 階段三：對每個新增項目增加註解
註解內容須包含每個func的用途與行為描述，並說明其輸入、輸出及副作用