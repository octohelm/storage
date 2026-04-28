# 包选型地图

`storage` 更像一组可组合的基础包，而不是必须整套接入的框架。给其他项目接入时，先选最小能力层。

## 公开包按职责划分

- `github.com/octohelm/storage/pkg/sqlfrag`
  适合只需要 SQL 片段、参数展开、命名参数替换或片段拼接的场景。
- `github.com/octohelm/storage/pkg/sqlbuilder`
  适合需要表、列、索引、条件、语句和 catalog 定义，但还不需要查询管道的场景。
- `github.com/octohelm/storage/pkg/sqlpipe`
  适合需要以组合方式构造查询、写入、分页、排序、关联、聚合和冲突处理的场景。
- `github.com/octohelm/storage/pkg/session`
  适合需要把模型、catalog 和数据库执行上下文绑定在一起的场景。
- `github.com/octohelm/storage/pkg/migrator`
  适合需要根据目标 catalog 做数据库结构迁移的场景。
- `github.com/octohelm/storage/pkg/sqltype`
  适合复用时间、JSON、可空值、软删除等常见字段类型。
- `github.com/octohelm/storage/pkg/filter`
  适合需要可序列化过滤规则，并与 `sqlpipe/filter` 组合使用的场景。

## 每层先看哪些入口

- `sqlfrag`
  先看 `Fragment`、`Const(...)`、`Pair(...)`、`Join(...)`、`Collect(...)`。
- `sqlbuilder`
  先看 `T(...)`、`Col(...)`、`Cols(...)`、`TableFromModel(...)`、`Where(...)`、`OrderBy(...)`、`OnConflict(...)`。
- `sqlpipe`
  先看 `FromAll[T]()`、`Value(...)`、`Values(...)`、`Where(...)`、`JoinOn...(...)`、`Project(...)`、`Limit(...)`、`OnConflictDo...(...)`。
- `session`
  先看 `New(...)`、`RegisterCatalog(...)`、`InjectContext(...)`、`For(...)`、`MustFor(...)`。
- `pkg/sqlpipe/ex`
  先看执行器如何把 `sqlpipe.Source` 接到 `session`。
- `migrator`
  先看如何基于 `sqlbuilder.Catalog` 比较目标结构与当前数据库结构。
- `filter`
  先看 `Filter[T]`、`Eq(...)`、`In(...)`、`Where(...)`、`Compose(...)`。

## 常见需求从哪里进

- “只想减少手写 SQL 字符串”
  先从 `sqlfrag` 或 `sqlbuilder` 开始，不要直接引入 `session`。
- “已有模型，想自动得到表和列定义”
  从 `sqlbuilder.TableFromModel` 开始，再决定是否需要 `session` 或 `migrator`。
- “想把查询、排序、分页、关联拼成一条数据源”
  从 `sqlpipe.FromAll[T]()` 和各类 `Pipe(...)` 操作符开始。
- “想把查询真正执行出去”
  在已有 `sqlpipe` 或 `sqlbuilder` 基础上，再接 `session` 和 `pkg/sqlpipe/ex`。
- “想在部署或启动时做结构迁移”
  从 `migrator` 和 `sqlbuilder/catalog` 开始。
- “想让 API 过滤条件可序列化、可反序列化”
  从 `filter` 和 `sqlpipe/filter` 开始。

## 选型判断

- 已经有稳定 repository 层，只缺 SQL 结构表达：
  先用 `sqlbuilder`，不要默认引入 `sqlpipe`。
- 已经能表达表和列，但查询组合逻辑分散：
  引入 `sqlpipe` 统一过滤、排序、分页和写入来源。
- 已经能拼出 SQL，但执行上下文、catalog 和模型定位不稳定：
  再引入 `session`。
- 已经有外部迁移工具，且不想切换：
  可以只用 `sqlbuilder` / `sqlpipe`，不必强接 `migrator`。
- 只是想让过滤参数具备文本编解码能力：
  先用 `filter`，只有在需要落到 SQL 条件时再接 `sqlpipe/filter`。

## 最小组合建议

- 片段级：`sqlfrag`
- 结构化 SQL：`sqlfrag` + `sqlbuilder`
- 查询构造：`sqlfrag` + `sqlbuilder` + `sqlpipe`
- 查询执行：`sqlpipe` + `session` + `pkg/sqlpipe/ex`
- 迁移：`sqlbuilder` + `session` + `migrator`
- 过滤 API：`filter` + `sqlpipe/filter`

## 不要怎么接

- 不要一开始同时引入 `sqlfrag`、`sqlbuilder`、`sqlpipe`、`session`、`migrator` 全套，只因为“以后可能会用到”。
- 不要让宿主项目直接依赖本仓库的 `internal/` 包。
- 不要先复制测试里的组合方式再理解分层；先判断宿主项目缺哪一层能力。
