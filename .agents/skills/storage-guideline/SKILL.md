---
name: storage-guideline
description: 说明如何在其他项目中引入并使用 `github.com/octohelm/storage` 的主要包与集成入口；当 agent 需要把 storage 作为依赖接入宿主项目时使用。
metadata:
  primary_pattern: tool-wrapper
---

# Storage Guideline

用于在其他 Go 项目中接入 `github.com/octohelm/storage` 时，快速选择合适的包入口、建模方式、执行路径和最小验证方案。

## 何时使用

- 需要在宿主项目中引入 `storage` 解决 SQL 组装、模型映射、查询管道、会话执行或迁移问题时。
- 需要判断当前需求应该从 `sqlfrag`、`sqlbuilder`、`sqlpipe`、`session`、`migrator` 还是 `filter` 进入时。
- 需要把宿主项目现有模型、数据库连接和测试方式与 `storage` 对齐时。

## 不适用

- 只是维护 `storage` 仓库自身源码、生成器或测试时。
- 只是想了解某个内部实现细节，但不会把 `storage` 当依赖接入宿主项目时。
- 宿主项目已经确定不使用 `storage` 的公开包，只是排查历史生成文件或仓库脚手架时。

## 输入

至少需要：

- 宿主项目当前要解决的数据访问问题
- 目标数据库或执行环境，例如 `postgres`、`sqlite`
- 宿主项目是否已有模型定义、会话封装、迁移入口或 SQL 测试方式

## 使用入口

1. 先读 [`references/package-map.md`](references/package-map.md) 判断应使用哪个公开包。
2. 再读 [`references/integration-workflow.md`](references/integration-workflow.md) 确认接入顺序、最小落地路径和验证方式。
3. 若宿主项目要定义可序列化过滤规则，可联动 [`../enumeration-guideline/SKILL.md`](../enumeration-guideline/SKILL.md) 处理相关枚举。
4. 若宿主项目会直接复用 `storage` 生态内的 `gengo` 生成约定或统一 Go 测试风格，再按需启用：
   - [`../gengo-guideline/SKILL.md`](../gengo-guideline/SKILL.md)
   - [`../testing-guideline/SKILL.md`](../testing-guideline/SKILL.md)

## 关键约定

- 优先从 `pkg/` 下公开包接入，不把 `internal/` 目录当宿主项目依赖面。
- 先选择最小必要层级，再决定是否叠加更高层能力；不要一开始全套接入。
- `sqlbuilder` 负责结构表达，`sqlpipe` 负责查询和写入组合，`session` / `pkg/sqlpipe/ex` 才负责执行。
- 宿主项目若已有稳定的数据访问抽象，应只吸收 `storage` 中缺失的那一层能力，而不是重建整套分层。
- 最小验证应跟接入层级一致：低层看 SQL 与参数，高层再看执行、事务与迁移。

## 工作方式

1. 先判断宿主项目缺的是哪一层能力：片段、结构化 SQL、查询管道、执行会话、迁移还是过滤模型。
2. 按 `references/package-map.md` 选择最小公开包集合，避免一开始同时引入过多层。
3. 再按 `references/integration-workflow.md` 从模型与表定义、查询写入、执行会话、迁移或测试中选最短接入路径。
4. 只有当宿主项目确实需要生成、枚举或统一测试风格时，才继续打开专项 skill。

## 完成标准

- 已判断宿主项目当前需要接入 `storage` 的哪一层公开能力。
- 已选出最小必要包集合与主要入口函数或类型。
- 已明确宿主项目应如何做最小验证，而不是只停留在阅读源码。
- `SKILL.md` 与附属 reference 已能独立充当 `storage` 包使用手册，不依赖仓库内其他文档链接才能执行。
