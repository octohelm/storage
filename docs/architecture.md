# 架构说明

这个仓库的核心不是单个包，而是几层逐步上升的抽象。

## 分层关系

1. `sqlfrag`
   负责最小 SQL 片段抽象。
   输入是片段和参数，输出是 `Frag(ctx) -> iter.Seq2[string, []any]`。

2. `sqlbuilder`
   在 `sqlfrag` 之上表达表、列、索引、条件、语句和附加子句。
   这一层仍然是“结构化 SQL”，还没有执行语义。

3. `sqlpipe`
   把 `sqlbuilder` 的语句能力组织成“数据源 + 操作符”的管道模型。
   过滤、排序、分页、聚合、投影、插入来源、更新与删除都在这一层组合。

4. `session`
   提供会话抽象，把模型解析到 catalog 和 adapter，并通过 context 传递执行面。

5. `internal/sql/adapter`
   负责具体数据库方言、连接、事务和 catalog 读取。

6. `migrator`
   依赖 `adapter.Dialect` 和 `sqlbuilder.Catalog`，对当前结构和目标结构做差异计算。

## 一条典型查询路径

1. 在 `modelscoped` 或 `sqlbuilder` 中定义表和列。
2. 用 `sqlpipe.From[T]()` 建立起始数据源。
3. 通过 `Where`、`JoinOn`、`Select`、`AscSort`、`Limit` 等操作符组合出目标数据源。
4. 在 `pkg/sqlpipe/ex` 中把数据源包装成可执行对象。
5. `Executor` 调用 `internal/sql/adapter` 执行 SQL。
6. `internal/sql/scanner` 把结果扫描回模型。

## 一条典型写入路径

1. 用 `Value`、`Values`、`InsertFrom` 或 `DoUpdate*` / `DoDelete*` 构造写入源。
2. `sqlpipe/internal.Builder` 把写入源转成 `INSERT`、`UPDATE` 或 `DELETE`。
3. `session` 和 `adapter` 负责事务与执行。

## catalog 与模型

- `sqlbuilder.TableFromModel` 会从模型定义推导表、列和索引。
- `pkg/sqlbuilder/catalog` 负责把一组模型组装成 catalog。
- `session.RegisterCatalog` 让会话能按模型或表名解析到逻辑数据库。

## 为什么复杂度会偏高

这套设计刻意把“字符串 SQL”拆成了多层抽象，带来的好处是：

- 更容易做类型约束和重用。
- 更容易测边界行为。
- 更容易在不同数据库之间复用结构和执行逻辑。

代价是：

- 阅读一条完整 SQL 的生成路径时，需要跨越多个包。
- 上下文 flag、toggle 和 patcher 的存在，提高了灵活性，也提高了理解门槛。

因此，阅读时建议始终从“我现在在看哪一层”入手，而不是直接追所有实现细节。
