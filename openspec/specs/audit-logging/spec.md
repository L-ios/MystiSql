# 审计日志规范

## Purpose

定义 MystiSql 的审计日志功能，记录所有 SQL 执行操作的详细日志，包括查询语句、用户信息、执行时间和结果状态，为安全审计、问题排查和合规性检查提供支持。

## Requirements

### Requirement: 审计日志记录
系统 SHALL 记录所有 SQL 执行操作的审计日志。

#### Scenario: 记录 SELECT 查询
- **WHEN** 用户执行 SELECT 查询
- **THEN** 系统记录审计日志，包含 SQL 语句、用户、实例、执行时间、返回行数

#### Scenario: 记录 INSERT/UPDATE/DELETE 操作
- **WHEN** 用户执行数据修改操作
- **THEN** 系统记录审计日志，包含 SQL 语句、用户、实例、执行时间、影响行数

#### Scenario: 记录 DDL 操作
- **WHEN** 用户执行 DDL 操作（CREATE、ALTER、DROP）
- **THEN** 系统记录审计日志并标记为高风险操作

---

### Requirement: 审计日志字段
审计日志 SHALL 包含以下必要字段。

#### Scenario: 日志字段完整性
- **WHEN** 记录审计日志
- **THEN** 日志包含：timestamp, user_id, client_ip, instance, database, query, query_type, rows_affected, execution_time, status

#### Scenario: 时间戳格式
- **WHEN** 记录审计日志的 timestamp 字段
- **THEN** 使用 ISO 8601 格式（如 "2026-03-07T10:00:00Z"）

---

### Requirement: 审计日志存储
系统 SHALL 将审计日志存储到文件，格式为 JSON Lines。

#### Scenario: 日志文件路径
- **WHEN** 配置 `audit.logFile` 为 `/var/log/mystisql/audit.log`
- **THEN** 审计日志写入该文件

#### Scenario: 日志文件格式
- **WHEN** 写入审计日志
- **THEN** 每行一个 JSON 对象，便于日志分析工具处理

#### Scenario: 异步写入
- **WHEN** SQL 执行完成
- **THEN** 审计日志异步写入文件，不阻塞请求响应

---

### Requirement: 日志轮转
系统 SHALL 支持审计日志文件轮转。

#### Scenario: 按天轮转
- **WHEN** 审计日志文件跨天
- **THEN** 创建新的日志文件，旧文件重命名为 `<filename>.YYYY-MM-DD`

#### Scenario: 保留期限
- **WHEN** 日志文件超过保留天数（默认 30 天）
- **THEN** 自动删除过期日志文件

---

### Requirement: 审计日志配置
系统 SHALL 支持通过配置文件管理审计日志参数。

#### Scenario: 启用/禁用审计日志
- **WHEN** 配置 `audit.enabled` 为 `true` 或 `false`
- **THEN** 系统相应地启用或禁用审计日志记录

#### Scenario: 配置日志文件路径
- **WHEN** 配置 `audit.logFile`
- **THEN** 审计日志写入指定路径

#### Scenario: 配置保留天数
- **WHEN** 配置 `audit.retentionDays`
- **THEN** 系统保留指定天数的日志文件

---

### Requirement: 审计日志查询接口
系统 SHALL 提供 RESTful API 查询审计日志。

#### Scenario: 按时间范围查询
- **WHEN** GET `/api/v1/audit/logs?start=2026-03-01&end=2026-03-07`
- **THEN** 系统返回指定时间范围内的审计日志

#### Scenario: 按用户查询
- **WHEN** GET `/api/v1/audit/logs?user_id=admin`
- **THEN** 系统返回指定用户的审计日志

#### Scenario: 按实例查询
- **WHEN** GET `/api/v1/audit/logs?instance=production-mysql`
- **THEN** 系统返回指定实例的审计日志

#### Scenario: 分页查询
- **WHEN** GET `/api/v1/audit/logs?page=1&page_size=100`
- **THEN** 系统返回第 1 页的 100 条日志记录

---

### Requirement: 敏感操作标记
系统 SHALL 在审计日志中标记敏感操作。

#### Scenario: 标记危险 DDL 操作
- **WHEN** 记录 DROP、TRUNCATE 等 DDL 操作
- **THEN** 审计日志标记 `sensitive: true`

#### Scenario: 标记全表删除
- **WHEN** 记录无 WHERE 子句的 DELETE 操作
- **THEN** 审计日志标记 `sensitive: true`
