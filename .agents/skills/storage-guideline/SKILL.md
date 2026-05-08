---
name: storage-guideline
description: 说明如何在其他项目中引入并使用 `github.com/octohelm/storage` 的主要包与集成入口；当 agent 需要把 storage 作为依赖接入宿主项目时使用。
---

# Storage Guideline

用于在宿主项目中接入 `github.com/octohelm/storage`，快速选择合适的包入口和最小集成路径。

## 架构总览

```
sqlfrag ──→ sqlbuilder ──→ sqlpipe ──→ session
SQL 片段     表/列/条件     查询管道      执行上下文
              │              │
              ├─ sqltype     ├─ filter
              │  字段类型     │  可序列化过滤
              │              │
              └─ migrator    └─ sqlpipe/ex
                 结构迁移       便捷执行
```

按需接入，不强制整套——缺哪层接哪层。

## 使用范围

- 宿主项目需要 SQL 组装、模型映射、查询管道、会话执行或迁移。
- 判断当前需求应从哪个包进入。
- 需要具体 API 签名时，优先 `go doc`，不在 skill 中复制手册。

不适用：只维护 storage 仓库自身源码，或宿主项目已确定不使用 storage。

## 读取导航

- 判断应使用哪个公开包 → [references/package-map.md](references/package-map.md)
- 确认接入顺序和最小验证 → [references/integration-workflow.md](references/integration-workflow.md)
- 若需要枚举能力 → 参考 enumeration-guideline skill
- 若需要 gengo 生成约定或统一测试风格 → 参考 gengo-guideline / testing-guideline skill

## 完成标准

- 已判断宿主项目需要接入哪一层公开能力。
- 已选出最小必要包集合。
- 已明确最小验证方式。
