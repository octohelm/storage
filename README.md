# storage

`storage` 是一个面向结构化数据访问的 Go 库仓库，围绕 SQL 组装、类型映射、迁移与会话能力提供可复用组件，当前主要面向 `sqlite3`、`duckdb` 和 `postgres`。

仓库重点不是提供单个可执行程序，而是提供一组可组合的库能力与配套生成、格式化、测试入口。

## 职责与边界

- 负责：结构化数据相关的库能力，包括 SQL 片段、构造器、类型、会话、迁移与过滤等基础组件。
- 负责：仓库内 Go 工具链入口，例如生成、格式化、测试与依赖整理。
- 不负责：在 root README 维护详细协作规则、长篇命令说明或实现细节。

## 文档导航

- [仓库概览](./docs/overview.md)：能力地图、包职责与推荐阅读顺序。
- [架构说明](./docs/architecture.md)：`sqlfrag`、`sqlbuilder`、`sqlpipe`、`session`、`adapter` 的协作关系。
- [开发与验证](./docs/development.md)：本地依赖、常用命令、生成与测试入口。
- [测试说明](./docs/testing.md)：测试分层、命名约定与覆盖率口径。

## 主要目录

- [`pkg/`](./pkg)：
  核心库代码，按能力拆分为 `sqlbuilder`、`sqltype`、`sqlfrag`、`session`、`migrator` 等包。
- [`devpkg/`](./devpkg)：
  仓库内部使用的开发辅助包与生成支持代码。
- [`internal/`](./internal)：
  内部命令、测试辅助与不对外暴露的实现细节。
- [`justfile`](./justfile)：
  仓库级执行入口，用于查看和聚合稳定命令入口。
- [`tool/go/justfile`](./tool/go/justfile)：
  Go 工具链执行面，集中暴露生成、格式化、测试与依赖相关命令。
- [`AGENTS.md`](./AGENTS.md)：
  仓库协作约束与暂停门禁。

## 快速开始

1. 阅读整体能力时，先看 [docs/overview.md](./docs/overview.md)。
2. 理解核心抽象时，再看 [docs/architecture.md](./docs/architecture.md)。
3. 需要本地跑命令时，运行 `just` 或 `just --list --list-submodules`。
4. 需要具体 Go 包时，从 [`pkg/`](./pkg) 进入。
