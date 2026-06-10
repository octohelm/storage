---
name: storage-guideline
description: 说明如何在其他项目中引入并使用 `github.com/octohelm/storage` 的主要包与集成入口；当 agent 需要把 storage 作为依赖接入宿主项目时使用。
---

# Storage Guideline

按需接入 `github.com/octohelm/storage`，缺哪层接哪层。

## 做个查询

```go
import "github.com/octohelm/storage/pkg/sqlpipe"

// 定义 source
src := sqlpipe.FromAll[Org]().
    Where(sqlpipe.Eq(sqlpipe.Col("id"), sqlpipe.Value(1))).
    Limit(sqlpipe.Const(10))
```

**选层**：
- 只拼 SQL 片段 → `sqlfrag`
- 需要表/列/条件表达 → `sqlbuilder`
- 需要查询管道（Where/Join/Project/Limit）→ `sqlpipe`
- 需要排序选项枚举（如 API 参数绑定、UI 排序下拉）→ `sort`
- 需要执行 → `sqlpipe` + `session` + `pkg/sqlpipe/ex`
- 需要迁移 → `sqlbuilder` + `migrator`
- 需要过滤 API → `filter` + `sqlpipe/filter`

**原则**：缺哪层接哪层。已有 repository 就只补缺失能力，不整套重建。

API 细节以 `go doc` 为准：`go doc github.com/octohelm/storage/pkg/sqlfrag` / `sqlbuilder` / `sqlpipe` / `sort` / `session`。

## 更多

- 包选型和完整接入顺序 → [references/package-map.md](references/package-map.md)
- 集成工作流和最小验证 → [references/integration-workflow.md](references/integration-workflow.md)
