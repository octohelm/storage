# 集成工作流

用于在宿主项目中把 `storage` 作为依赖接入，而不是把整个仓库的开发流程搬过去。

## 1. 先判断起点

开始前先确认：

- 宿主项目已有模型，还是要新建模型
- 需求是读查询、写入、迁移、过滤还是只做 SQL 组装
- 宿主项目已有数据库会话抽象，还是需要一起建立

如果这些信息缺失，先停在包选型层，不要直接开始写集成代码。

## 2. 最短接入路径

### 只做 SQL 结构化表达

1. 用 `sqlbuilder` 定义表、列、索引或条件。
2. 需要最终 SQL 时，用 `sqlfrag.Collect(...)` 收集 SQL 与参数。
3. 宿主项目若已有自己的执行层，到这里通常就够了。

### 做查询或写入组合

1. 用 `sqlbuilder.TableFromModel(...)` 或现有表定义准备模型映射。
2. 用 `sqlpipe.FromAll[T]()`、`Value(...)`、`Values(...)` 等建立 source。
3. 用 `Where`、`JoinOn`、`Project`、`Limit`、`OnConflictDo...` 等操作符组合行为。
4. 先验证 source 产出的 SQL 片段，再决定是否继续接执行层。

### 做执行会话

1. 先有 `sqlbuilder.Catalog` 或模型表定义。
2. 再接 `session` 注册 catalog，并把会话注入上下文。
3. 真正执行 `sqlpipe` source 时，再接 `pkg/sqlpipe/ex`。
4. 若宿主项目已有自己的事务与连接抽象，优先评估只接 catalog 与上下文注入是否足够。

### 做数据库迁移

1. 先确保目标表结构能由 `sqlbuilder` 稳定表达。
2. 再用 `migrator` 比较目标 catalog 与数据库当前结构。
3. 只有在宿主项目愿意把目标 schema 收敛到 `storage` catalog 表达时，再继续接迁移执行面。

## 2.5 典型接入顺序

### Web API 查询

1. 用 `sqlbuilder.TableFromModel(...)` 或现有模型表定义表达表结构。
2. 用 `sqlpipe.FromAll[T]()` 加 `Where`、`Project`、`Limit`、`Sort` 组织查询。排序参数可通过 `sort.By[E]` 表达可枚举的排序选项（正序/逆序）。
3. 若请求参数需要文本过滤规则，增加 `filter` 和 `sqlpipe/filter`。
4. 最后才接 `session` 与 `pkg/sqlpipe/ex` 执行。

### 后台批量写入

1. 用 `Value(...)`、`Values(...)` 或 `InsertFrom(...)` 建立写入 source。
2. 需要幂等写入时，再加 `OnConflictDoNothing(...)` 或 `OnConflictDoUpdateSet(...)`。
3. 用局部执行测试确认 SQL 和参数，再接事务执行。

### 启动期 schema 管理

1. 先把表、索引和字段定义收敛到 `sqlbuilder.Catalog`。
2. 再让 `migrator` 比较差异。
3. 仅在宿主项目确实接受该差异模型时，才接入自动迁移流程。

## 3. 宿主项目里的落点

- 表和模型相关定义，放在宿主项目自己的模型层。
- 查询和写入组合逻辑，放在宿主项目自己的 repository、store 或 data access 层。
- `session`、连接和 catalog 注册，放在宿主项目的基础设施初始化层。
- 不要把本仓库的测试包结构、内部目录结构原样复制到宿主项目。

## 4. 最小验证

- 只接 `sqlfrag` / `sqlbuilder` 时：
  验证 `Collect(...)` 产出的 SQL 和参数。
- 接 `sqlpipe` 时：
  验证关键查询或写入分支是否产出预期 SQL 片段。
- 接 `session` / `sqlpipe/ex` 时：
  验证会话注入、catalog 解析、读写执行或事务行为。
- 接 `migrator` 时：
  验证目标 schema 与数据库当前结构的差异是否符合预期。

优先级顺序：

1. 先验低层结构与 SQL 片段。
2. 再验执行与事务。
3. 最后才验跨数据库或自动迁移行为。

## 5. 配套技能何时联动

- 过滤规则中需要 enum：联动 `enumeration-guideline`
- 宿主项目也要复用 `gengo` 生成：联动 `gengo-guideline`
- 宿主项目要采用同一套 Go 测试断言风格：联动 `testing-guideline`

## 6. 风险提示

- `storage` 分层较细，误把低层和高层一起接入，常见结果是宿主项目职责混写。
- 直接从执行层开始抄用，而没先确认模型与表定义，常见结果是后续迁移和 SQL 结构难以收敛。
- 只验证最终 SQL 字符串而不验证宿主项目的调用边界，常见结果是集成面可以拼出 SQL，但执行面仍不稳定。
- 把 `storage` 的测试或目录结构直接复制进宿主项目，常见结果是宿主项目出现不必要的仓库耦合。
