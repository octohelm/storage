# 包选型地图

`storage` 是一组可组合的基础包，不强制整套接入。按需选层。

## 公开包

| 包 | 适合场景 |
|----|----------|
| `sqlfrag` | SQL 片段、参数展开、命名参数替换、片段拼接 |
| `sqlbuilder` | 表、列、索引、条件、语句、catalog 定义 |
| `sqlpipe` | 查询组合、写入、分页、排序、关联、聚合、冲突处理 |
| `session` | 模型、catalog 与数据库执行上下文绑定 |
| `migrator` | 基于 catalog 的数据库结构迁移 |
| `sqltype` | 时间、JSON、可空值、软删除等常见字段类型 |
| `filter` | 可序列化过滤规则，配合 `sqlpipe/filter` 使用 |
| `sort` | 基于枚举类型的排序选项，正序/逆序可枚举，适用于 API 参数绑定与排序 UI 生成 |

每层入口通过 `go doc` 查阅：
- `go doc github.com/octohelm/storage/pkg/sqlfrag`
- `go doc github.com/octohelm/storage/pkg/sqlbuilder`
- `go doc github.com/octohelm/storage/pkg/sqlpipe`
- `go doc github.com/octohelm/storage/pkg/session`
- `go doc github.com/octohelm/storage/pkg/migrator`
- `go doc github.com/octohelm/storage/pkg/sqltype`
- `go doc github.com/octohelm/storage/pkg/filter`
- `go doc github.com/octohelm/storage/pkg/sort`

## 选型指南

- 只想减少手写 SQL：`sqlfrag` 或 `sqlbuilder`。
- 已有模型，想自动得表和列：`sqlbuilder.TableFromModel`。
- 想统一查询/排序/分页/关联：`sqlpipe`。
- 想真执行出去：`sqlpipe` + `session` + `pkg/sqlpipe/ex`。
- 想做结构迁移：`sqlbuilder` + `session` + `migrator`。
- 想让 API 过滤可序列化：`filter` + `sqlpipe/filter`。

**原则**：缺哪层接哪层。已有稳定 repository 层就只补缺失能力，不要整套重建。不要一开始就全量引入。
