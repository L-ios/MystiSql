# SQL 白名单/黑名单规范

## Purpose

定义 MystiSql 的 SQL 白名单和黑名单功能，支持基于规则的 SQL 访问控制，允许管理员灵活配置哪些 SQL 模式可以执行或被禁止，增强系统的安全性和可控性。

## Requirements

### Requirement: SQL 白名单机制
系统 SHALL 支持 SQL 白名单配置，允许特定的 SQL 模式执行。

#### Scenario: 配置白名单规则
- **WHEN** 配置 `validator.whitelist` 为 `["SELECT * FROM users WHERE id = ?", "SHOW TABLES"]`
- **THEN** 匹配白名单的 SQL 允许执行

#### Scenario: 白名单绕过危险检查
- **WHEN** SQL 匹配白名单规则
- **THEN** 即使该操作在危险操作列表中，也允许执行

#### Scenario: 白名单使用正则表达式
- **WHEN** 配置白名单规则为正则表达式（如 `^SELECT .* FROM system_config`)
- **THEN** 系统使用正则匹配判断 SQL 是否在白名单中

---

### Requirement: SQL 黑名单机制
系统 SHALL 支持 SQL 黑名单配置，禁止特定的 SQL 模式执行。

#### Scenario: 配置黑名单规则
- **WHEN** 配置 `validator.blacklist` 为 `["DROP TABLE users", "DELETE FROM audit_log"]`
- **THEN** 匹配黑名单的 SQL 被拒绝执行

#### Scenario: 黑名单返回 403 错误
- **WHEN** SQL 匹配黑名单规则
- **THEN** 系统返回 403 Forbidden 并提示 SQL 在黑名单中

#### Scenario: 黑名单使用正则表达式
- **WHEN** 配置黑名单规则为正则表达式（如 `^DELETE FROM audit`)
- **THEN** 系统使用正则匹配判断 SQL 是否在黑名单中

---

### Requirement: 白名单和黑名单优先级
系统 SHALL 定义白名单和黑名单的优先级规则。

#### Scenario: 黑名单优先于白名单
- **WHEN** SQL 同时匹配白名单和黑名单
- **THEN** 黑名单优先生效，拒绝执行

#### Scenario: 白名单优先于危险操作检查
- **WHEN** SQL 匹配白名单，但属于危险操作
- **THEN** 白名单优先生效，允许执行

---

### Requirement: 配置热更新
系统 SHALL 支持运行时更新白名单和黑名单配置。

#### Scenario: 更新白名单
- **WHEN** 管理员调用 API 更新白名单配置
- **THEN** 新的白名单规则立即生效，无需重启服务

#### Scenario: 更新黑名单
- **WHEN** 管理员调用 API 更新黑名单配置
- **THEN** 新的黑名单规则立即生效，无需重启服务

---

### Requirement: 配置持久化
系统 SHALL 将白名单和黑名单配置持久化到文件。

#### Scenario: 保存配置到文件
- **WHEN** 更新白名单或黑名单配置
- **THEN** 系统将配置保存到 `config/validator.yaml` 文件

#### Scenario: 服务启动时加载配置
- **WHEN** 服务启动
- **THEN** 系统从配置文件加载白名单和黑名单规则
