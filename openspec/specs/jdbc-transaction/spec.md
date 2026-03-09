# JDBC 事务规范

## Purpose

定义 MystiSql 的 JDBC 事务管理功能，支持通过 JDBC Connection 进行事务的开始、提交、回滚操作，确保数据一致性和事务隔离性。

## Requirements

### Requirement: JDBC 事务开始
系统 SHALL 支持通过 JDBC Connection 开始事务。

#### Scenario: JDBC setAutoCommit(false)
- **WHEN** JDBC 客户端调用 `Connection.setAutoCommit(false)`
- **THEN** 系统创建事务上下文并返回 `connectionId`

#### Scenario: 事务 SQL 请求携带 connectionId
- **WHEN** 事务中的 SQL 请求携带 `connectionId` 参数
- **THEN** 系统使用同一数据库连接执行 SQL

---

### Requirement: JDBC 事务提交
系统 SHALL 支持通过 JDBC Connection 提交事务。

#### Scenario: JDBC commit()
- **WHEN** JDBC 客户端调用 `Connection.commit()`
- **THEN** 系统提交事务并释放数据库连接

#### Scenario: 事务提交后自动关闭连接
- **WHEN** 事务提交成功
- **THEN** 系统关闭 `connectionId` 对应的数据库连接

---

### Requirement: JDBC 事务回滚
系统 SHALL 支持通过 JDBC Connection 回滚事务。

#### Scenario: JDBC rollback()
- **WHEN** JDBC 客户端调用 `Connection.rollback()`
- **THEN** 系统回滚事务并释放数据库连接

#### Scenario: 事务回滚后自动关闭连接
- **WHEN** 事务回滚成功
- **THEN** 系统关闭 `connectionId` 对应的数据库连接

---

### Requirement: 事务超时自动回滚
系统 SHALL 在事务超时后自动回滚。

#### Scenario: 事务超时回滚
- **WHEN** 事务持续时间超过 `transaction.timeout`（默认 5 分钟）
- **THEN** 系统自动回滚事务并释放连接

#### Scenario: 超时后使用 connectionId 失败
- **WHEN** 客户端在事务超时后使用 `connectionId`
- **THEN** 系统返回错误 "事务已超时并回滚"

---

### Requirement: REST API 事务接口
系统 SHALL 提供 RESTful API 管理事务。

#### Scenario: 开始事务
- **WHEN** POST `/api/v1/transaction/begin` 请求包含 `instance` 参数
- **THEN** 系统返回 `{"connectionId": "xxx", "expiresAt": "2026-03-07T10:05:00Z"}`

#### Scenario: 提交事务
- **WHEN** POST `/api/v1/transaction/commit` 请求包含 `connectionId` 参数
- **THEN** 系统提交事务并返回成功响应

#### Scenario: 回滚事务
- **WHEN** POST `/api/v1/transaction/rollback` 请求包含 `connectionId` 参数
- **THEN** 系统回滚事务并返回成功响应

---

### Requirement: 事务隔离级别
系统 SHALL 支持配置事务隔离级别。

#### Scenario: 设置隔离级别
- **WHEN** JDBC 客户端调用 `Connection.setTransactionIsolation(Connection.TRANSACTION_READ_COMMITTED)`
- **THEN** 系统设置事务隔离级别为 READ COMMITTED

---

### Requirement: 事务并发控制
系统 SHALL 防止并发冲突。

#### Scenario: connectionId 绑定到客户端
- **WHEN** 客户端 A 创建 `connectionId`
- **THEN** 客户端 B 无法使用该 `connectionId`

#### Scenario: 单个连接的并发查询
- **WHEN** 同一个 `connectionId` 同时收到多个 SQL 请求
- **THEN** 系统串行执行 SQL（符合 JDBC Connection 规范）
