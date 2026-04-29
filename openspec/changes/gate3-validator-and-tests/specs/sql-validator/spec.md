## MODIFIED Requirements

### Requirement: SQL 危险操作验证
系统 SHALL 验证所有 SQL 查询，阻断配置中定义的危险操作（DROP、TRUNCATE、DELETE without WHERE、UPDATE without WHERE）。

#### Scenario: AST 模式下拦截注释绕过的危险操作
- **WHEN** `validator.useParser` 为 true，且 SQL 为 `DROP/*comment*/TABLE users`
- **THEN** 系统 SHALL 通过 AST 解析正确识别 DROP 操作，返回 `Allowed: false`

#### Scenario: AST 模式下拦截多语句注入
- **WHEN** `validator.useParser` 为 true，且 SQL 为 `SELECT 1; DROP TABLE users`
- **THEN** 系统 SHALL 检测到多语句中的 DROP 操作，返回 `Allowed: false`

#### Scenario: tokenizer 模式下行为不变
- **WHEN** `validator.useParser` 为 false，且 SQL 为任意合法查询
- **THEN** 系统 SHALL 使用 tokenizer 验证，行为与当前版本完全一致
