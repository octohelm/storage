# 开发与验证

## 本地入口

- `just`：查看仓库级入口。
- `just go`：查看 Go 工具链入口。
- `just serve-dbs`：启动测试依赖数据库。

## 常用命令

### 查看入口

```bash
just
just --list --list-submodules
```

### Go 测试

```bash
just go::test
go test ./...
```

按包最小验证时，优先只跑受影响包，例如：

```bash
go test ./pkg/sqlbuilder/...
go test ./pkg/sqlpipe/...
go test ./pkg/session ./internal/sql/adapter/...
```

### 代码生成

```bash
just go::gen
go generate ./...
```

### 依赖整理

```bash
just go::dep
go mod tidy
```

## 与数据库相关的开发

仓库当前主要面向 `sqlite`、`postgres`，部分历史文档里也提到 `duckdb`。实际执行层集中在：

- `internal/sql/adapter/sqlite`
- `internal/sql/adapter/postgres`

如果测试依赖外部数据库，先启动：

```bash
just serve-dbs
```

## 阅读代码时的建议路径

### 查 SQL 是怎么构造出来的

1. `pkg/sqlpipe`
2. `pkg/sqlpipe/internal`
3. `pkg/sqlbuilder`
4. `pkg/sqlfrag`

### 查模型为什么变成这个表结构

1. `pkg/sqlbuilder/TableFromModel`
2. `pkg/sqlbuilder/structs`
3. `pkg/sqlbuilder/internal/columndef`
4. `pkg/sqltype`

### 查执行与扫描

1. `pkg/session`
2. `internal/sql/adapter`
3. `internal/sql/scanner`

## 修改时的建议

- 先做按包最小验证，不要默认跑全仓。
- 生成文件改动和手写代码改动分开看待。
- 涉及 `sqlpipe` 或 `sqlbuilder` 时，优先补充行为测试，而不是只看最终 SQL 字符串。
