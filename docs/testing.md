# 测试说明

## 测试分层

这个仓库的测试大致分成四类：

1. 纯结构层测试
   例如 `sqlfrag`、`sqlbuilder`、`filter`。
   重点验证 SQL 片段、条件组合、语句构造和序列化行为。

2. 管道层测试
   例如 `sqlpipe`。
   重点验证投影、过滤、排序、分页、聚合、插入来源和更新删除分支。

3. 执行层测试
   例如 `session`、`internal/sql/adapter`、`internal/sql/scanner`。
   重点验证事务、连接选择、结果扫描和方言行为。

4. 数据库集成测试
   依赖外部数据库或容器。
   重点验证 catalog、迁移和真实数据库方言差异。

## 文件命名建议

- 优先按主题或对象命名，例如 `typed_column_test.go`、`source_values_test.go`、`order_test.go`。
- 避免持续新增泛化的 `more_test.go` 风格文件。
- 如果一个测试文件已经承担多个分支补洞任务，后续应考虑按主题拆开。

## 覆盖率口径

- 关注手写代码覆盖率。
- 生成文件不计入覆盖率判断。
- 覆盖率达标不代表文档和命名已经足够清晰，仍要关注可维护性。

## 测试写法

仓库内优先使用统一测试辅助和断言风格：

- `github.com/octohelm/x/testing/v2`
- `internal/testutil`
- `pkg/sqlfrag/testutil`

新增测试时，优先表达：

- 输入条件
- 关键分支
- 期望 SQL 或期望行为

而不是把多个无关分支塞进一个很长的测试函数里。
