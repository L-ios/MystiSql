## ADDED Requirements

### Requirement: AST-based SQL validation
系统 SHALL 提供基于 AST 解析的 SQL 验证增强层，作为现有 tokenizer 验证的可选替代。

#### Scenario: AST 验证器启用时拦截危险操作
- **WHEN** 配置 `validator.useParser` 为 true，且 SQL 为 `DROP TABLE users`
- **THEN** 系统使用 AST 解析 SQL，返回 `Allowed: false`，Reason 包含 "DROP"

#### Scenario: AST 解析失败时降级到 tokenizer
- **WHEN** 配置 `validator.useParser` 为 true，且 SQL 包含 PostgreSQL 特有语法（如 `INSERT ... ON CONFLICT`）导致 AST 解析失败
- **THEN** 系统 SHALL 自动降级到现有 tokenizer 验证，并记录 warn 级别日志

#### Scenario: AST 验证器关闭时使用 tokenizer
- **WHEN** 配置 `validator.useParser` 为 false（默认）
- **THEN** 系统 SHALL 使用现有 tokenizer 验证，行为与当前完全一致

#### Scenario: 拦截注释绕过的 DELETE
- **WHEN** SQL 为 `DELETE FROM users -- WHERE id = 1`
- **THEN** AST 验证器 SHALL 识别此为无 WHERE 子句的 DELETE，返回 `Allowed: false`

#### Scenario: 拦截子查询绕过的 DELETE
- **WHEN** SQL 为 `DELETE FROM users WHERE id IN (SELECT id FROM backup)`
- **THEN** AST 验证器 SHALL 识别此为有 WHERE 子句的 DELETE，返回 `Allowed: true`

### Requirement: AST 验证器配置
系统 SHALL 通过 `validator.useParser` 配置项控制 AST 验证器的启用状态。

#### Scenario: 默认配置下 AST 验证器关闭
- **WHEN** 配置文件未设置 `validator.useParser`
- **THEN** 系统 SHALL 默认关闭 AST 验证器，使用 tokenizer 验证

#### Scenario: 配置启用 AST 验证器
- **WHEN** 配置文件设置 `validator.useParser: true`
- **THEN** 系统 SHALL 在服务启动时初始化 AST 验证器

### Requirement: AST 验证器性能
AST 验证器 SHALL 在单次验证上的延迟不超过 tokenizer 验证的 5 倍。

#### Scenario: AST 验证器性能基准
- **WHEN** 对标准 SQL 语句运行 1000 次验证
- **THEN** AST 验证器的平均延迟 SHALL 不超过 tokenizer 验证器平均延迟的 5 倍
