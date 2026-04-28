# 仓库概览

`storage` 提供一套面向结构化数据访问的 Go 基础库。它不是 ORM，也不是单一 SQL builder，而是围绕“模型定义 -> SQL 构造 -> 数据源组合 -> 执行与迁移”组织的一组库。

## 适合解决的问题

- 定义可复用的模型、列、索引和数据库类型。
- 以组合方式构造 SQL，而不是手写大段字符串。
- 为 `sqlite`、`postgres` 等数据库复用统一执行与迁移逻辑。
- 用可测试的方式组织过滤、投影、分页、关联、聚合与写入逻辑。

## 核心包地图

- `pkg/sqlfrag`：最底层 SQL 片段抽象，负责片段拼接、参数收集和上下文注入。
- `pkg/sqlbuilder`：表、列、索引、条件、语句和方言无关的 SQL 结构化构造器。
- `pkg/sqlpipe`：面向查询与写入流程的管道式组合层。
- `pkg/session`：会话、catalog 注册、上下文注入和适配器选择。
- `pkg/migrator`：按目标 catalog 对数据库结构做差异迁移。
- `pkg/sqltype`：模型字段常用类型、时间字段、JSON 字段和软删除辅助接口。
- `pkg/filter`：可序列化的过滤规则模型，配合 `pkg/sqlpipe/filter` 生成 SQL 条件。
- `pkg/er`：从 catalog 提取 ER 结构，输出更适合展示和序列化的模型。

## internal 目录

- `internal/sql/adapter`：数据库连接、方言与事务执行的内部实现。
- `internal/sql/scanner`：结果扫描、迭代器与空值处理。
- `internal/sql/loggingdriver`：数据库驱动日志包装。
- `internal/testutil`：仓库内部测试辅助。
- `internal/xiter`：少量序列工具函数。

## 推荐阅读顺序

1. 先看 `pkg/sqlfrag`，理解片段如何产出 SQL 和参数。
2. 再看 `pkg/sqlbuilder`，理解表、列、条件和语句的表达方式。
3. 然后看 `pkg/sqlpipe`，理解如何把 SQL 结构组织成可组合的数据源。
4. 最后看 `pkg/session`、`internal/sql/adapter` 和 `pkg/migrator`，理解执行与迁移面。

## 什么时候看什么

- 想查“一个字段怎么变成 SQL 列”：看 `pkg/sqlbuilder` 和 `pkg/sqltype`。
- 想查“一个查询怎么被组合出来”：看 `pkg/sqlpipe`。
- 想查“为什么最终执行成这条 SQL”：先看 `pkg/sqlpipe/internal`，再看 `pkg/sqlfrag`。
- 想查“为什么数据库结构会变更”：看 `pkg/migrator` 和 `pkg/sqlbuilder/catalog`。
