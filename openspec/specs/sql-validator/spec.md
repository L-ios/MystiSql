# SQL 验证器规范

## Purpose

定义 MystiSql 的 SQL 验证功能，通过检测和拦截危险的 SQL 操作（如 DROP、TRUNCATE、无 WHERE 的 DELETE/UPDATE），防止意外的数据丢失和破坏，保障数据库安全。

## Requirements

### Requirement: 危险操作检测
系统 SHALL 检测并拦截危险的 SQL 操作。

#### Scenario: 拦截 DROP TABLE
- **WHEN** 用户执行 `DROP TABLE table_name`
- **THEN** 系统返回 403 Forbidden 并拒绝执行

#### Scenario: 拦截 DROP DATABASE
- **WHEN** 用户执行 `DROP DATABASE db_name`
- **THEN** 系统返回 403 Forbidden 并拒绝执行

#### Scenario: 拦截 TRUNCATE
- **WHEN** 用户执行 `TRUNCATE TABLE table_name`
- **THEN** 系统返回 403 Forbidden 并拒绝执行

#### Scenario: 拦截无 WHERE 的 DELETE
- **WHEN** 用户执行 `DELETE FROM table_name`（无 WHERE 子句）
- **THEN** 系统返回 403 Forbidden 并拒绝执行

#### Scenario: 拦截无 WHERE 的 UPDATE
- **WHEN** 用户执行 `UPDATE table_name SET column=value`（无 WHERE 子句）
- **THEN** 系统返回 403 Forbidden 并拒绝执行

---

### Requirement: 危险操作配置
系统 SHALL 支持配置哪些操作被视为危险操作。

#### Scenario: 配置危险操作列表
- **WHEN** 配置 `validator.dangerousOperations` 为 `["DROP", "TRUNCATE", "DELETE_WITHOUT_WHERE"]`
- **THEN** 系统拦截配置列表中的操作

#### Scenario: 禁用特定危险操作检查
- **WHEN** 配置中移除某项危险操作
- **THEN** 系统不再拦截该操作

---

### Requirement: SQL 解析器验证
系统 SHALL 使用 SQL 解析器（而非正则表达式）识别 SQL 语句类型。

#### Scenario: 准确识别 SQL 类型
- **WHEN** 用户执行 SQL 语句
- **THEN** 系统通过解析器准确识别语句类型（SELECT、INSERT、UPDATE、DELETE、DDL）

#### Scenario: 不误判字段名
- **WHEN** SQL 语句中包含字符串 "DROP"（如字段名）
- **THEN** 系统不误判为 DROP 操作

---

### Requirement: 验证开关
系统 SHALL 支持启用或禁用 SQL 验证。

#### Scenario: 启用验证
- **WHEN** 配置 `validator.enabled` 为 `true`
- **THEN** 系统对所有 SQL 执行验证

#### Scenario: 禁用验证
- **WHEN** 配置 `validator.enabled` 为 `false`
- **THEN** 系统跳过 SQL 验证

---

### Requirement: 验证错误信息
系统 SHALL 在拦截危险操作时返回清晰的错误信息。

#### Scenario: 返回错误详情
- **WHEN** 拦截危险操作
- **THEN** 返回错误信息包含：被拦截的 SQL、拦截原因、建议操作

---

### Requirement: 白名单绕过
系统 SHALL 支持通过白名单绕过危险操作检查。

#### Scenario: 白名单 SQL 执行
- **WHEN** SQL 匹配白名单规则
- **THEN** 系统允许执行，即使该操作在危险操作列表中
