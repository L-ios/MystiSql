## MODIFIED Requirements

### Requirement: 白名单/黑名单匹配
系统 SHALL 支持通过正则表达式模式配置 SQL 白名单和黑名单。

#### Scenario: AST 模式下白名单匹配（未来）
- **WHEN** `validator.useParser` 为 true，且白名单配置了表名模式
- **THEN** 系统 SHALL 继续使用正则匹配（AST 表名级别匹配留后续 change 实现）

#### Scenario: 黑名单匹配行为不变
- **WHEN** SQL 匹配黑名单中的正则模式
- **THEN** 系统 SHALL 拒绝该查询，无论 `useParser` 设置如何
