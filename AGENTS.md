# MystiSql Agent Guidelines

本文件只保留 **AI Agent 必须立即知道** 的规则、约束和索引。
详细说明请按需阅读 `docs/agents/*.md`。

## Project Overview

MystiSql 是一个面向 Kubernetes 集群的数据库访问网关，支持多种数据库，并提供以下接入方式：
- CLI / REPL
- REST API
- WebSocket
- JDBC Driver
- WebUI

**当前阶段**：README 标记为 Phase 3，但 AI Agent **不能只相信 README**，必须以代码实际实现为准。

## Working Rules for Agents

### 1. 先读规则，再动手
- 先看本文件，再按需阅读 `docs/agents/*.md`
- 不要把详细参考文档和必须遵守的规则混在一起理解

### 2. 优先使用 Makefile
- 构建、测试、lint、E2E **优先使用 Makefile 目标**
- 仅在 Makefile 未覆盖的场景下，才直接使用底层命令
- 详细命令见：`docs/agents/build-and-test.md`

### 3. 默认按代码实际情况判断
- 不以 README、目录结构、文件数量判断功能完成度
- 需要逐文件确认“代码存在”还是“功能可用”
- 评估完成度时必须覆盖：**Go 后端、JDBC、WebUI**

### 4. 代码风格要求
- 遵循 **Effective Go** 和 **Go Code Review Comments**
- 函数保持精简，优先组合而非继承
- 错误必须显式处理并附加上下文
- `context.Context` 始终作为第一个参数
- 导出符号必须有文档注释
- goroutine 必须受控，不允许无界并发
- 详细规范见：`docs/agents/code-style.md`

### 5. 交互界面命名
- 项目中的终端交互界面统一称为 **REPL**
- 不要把它描述成 TUI
- REPL/CLI 细节见：`docs/agents/cli.md`

## Project Owner Rules

以下规则为强制规则，优先级最高。

### 1. 审查要深，不要表面检查
- 逐文件读代码，区分“文件存在”和“功能可用”
- 不以文件计数或目录结构作为完成度指标
- 给出每个模块的真实实现深度评估（实质逻辑 vs 桩代码）

### 2. 影响分析要全
- 任何改动方案必须评估对所有接入层的影响：Go 后端、JDBC Client、WebUI
- 不能只看 Go 后端就认为影响分析完成
- 标注每个改动对 JDBC/WebUI 的具体影响（必须改动 / 需要适配 / 无影响）

### 3. 信心驱动决策
- 每个方案/改动必须标注信心值（0-100%）
- **信心不足 90% 不许实施**，必须先讨论分歧点
- 信心值必须附带原因说明
- 实施过程中若发现假设错误导致信心下降，立即停止并讨论

### 4. 不遗漏
- 低信心项也必须纳入规划，不能因为难就砍掉
- 所有发现的问题都要记录，即使暂不处理也要说明原因
- 明确区分“本次不做”和“不需要做”

### 5. 粒度可控
- 大方案必须拆分为可独立推进、独立验证的 proposal
- 每个 proposal 必须有明确 Done 标准和前置条件
- 依赖关系必须显式标注

### 6. 先验证再动手
- 编译修复是一切改动的前置条件，项目必须先能 `go build` 通过
- 不确定的技术选型（新库、新 API）先做 POC 验证
- POC 结果必须记录，并纳入信心评估

### 7. 发现分歧立即停止
- 实施过程中如果发现 proposal 假设有误，先更新 proposal 再继续
- 不在未确认的假设上继续构建
- 分歧点必须先与 Owner 讨论确认

### 8. 中文沟通
- 所有面向 Owner 的输出使用中文
- 代码注释和文档遵循项目现有语言习惯

## Detailed References

按需阅读以下文档：

- `docs/agents/project-overview.md` — 项目定位、核心能力、目录结构、功能概览
- `docs/agents/build-and-test.md` — 构建、测试、lint、E2E 命令详细说明
- `docs/agents/cli.md` — CLI / REPL 交互界面、快捷键、内置命令
- `docs/agents/code-style.md` — 代码风格、安全最佳实践、开发流程、PR Checklist

## Communication Language

- 与 Owner 交互：中文
- 文档：中文优先
- 代码注释：中文或英文，但文件内保持一致
