# JDBC 批量操作规范

## Purpose

定义 MystiSql 的 JDBC 批量操作功能，支持通过 JDBC Statement 批量执行 INSERT、UPDATE、DELETE 操作，提供高性能的数据批量处理能力。

## Requirements

### Requirement: JDBC 批量插入
系统 SHALL 支持通过 JDBC Statement 批量插入数据。

#### Scenario: JDBC addBatch() 和 executeBatch()
- **WHEN** JDBC 客户端调用 `PreparedStatement.addBatch()` 多次后调用 `executeBatch()`
- **THEN** 系统批量执行所有 INSERT 语句

#### Scenario: 批量插入返回影响行数
- **WHEN** 批量插入执行完成
- **THEN** 系统返回每个 SQL 的影响行数数组

---

### Requirement: JDBC 批量更新
系统 SHALL 支持通过 JDBC Statement 批量更新数据。

#### Scenario: 批量 UPDATE
- **WHEN** JDBC 客户端批量执行 UPDATE 语句
- **THEN** 系统批量执行所有 UPDATE 语句

#### Scenario: 批量更新返回影响行数
- **WHEN** 批量更新执行完成
- **THEN** 系统返回每个 SQL 的影响行数数组

---

### Requirement: JDBC 批量删除
系统 SHALL 支持通过 JDBC Statement 批量删除数据。

#### Scenario: 批量 DELETE
- **WHEN** JDBC 客户端批量执行 DELETE 语句
- **THEN** 系统批量执行所有 DELETE 语句

---

### Requirement: 混合批处理
系统 SHALL 支持混合批处理（INSERT、UPDATE、DELETE 混合）。

#### Scenario: 混合批处理执行
- **WHEN** JDBC 客户端混合添加 INSERT、UPDATE、DELETE 到批处理
- **THEN** 系统按顺序执行所有操作

#### Scenario: 混合批处理返回结果
- **WHEN** 混合批处理执行完成
- **THEN** 系统返回每个操作的影响行数数组

---

### Requirement: 批处理大小限制
系统 SHALL 限制批处理的最大大小。

#### Scenario: 批处理大小限制
- **WHEN** 批处理中的 SQL 数量超过 `batch.maxSize`（默认 1000）
- **THEN** 系统返回错误 "批处理大小超过限制"

---

### Requirement: REST API 批量操作接口
系统 SHALL 提供 RESTful API 执行批量操作。

#### Scenario: 批量执行 SQL
- **WHEN** POST `/api/v1/batch` 请求包含 `{"instance": "local-mysql", "queries": ["INSERT INTO users (name) VALUES ('Alice')", "INSERT INTO users (name) VALUES ('Bob')"]}`
- **THEN** 系统批量执行所有 SQL 并返回结果

#### Scenario: 批量操作响应格式
- **WHEN** 批量操作执行完成
- **THEN** 响应格式为 `{"results": [{"rowsAffected": 1}, {"rowsAffected": 1}], "totalRowsAffected": 2, "executionTime": 5000000}`

---

### Requirement: 批量操作性能
系统 SHALL 优化批量操作的执行性能。

#### Scenario: 批量操作性能优于单条执行
- **WHEN** 批量执行 100 条 INSERT
- **THEN** 执行时间显著低于逐条执行 100 次 INSERT

---

### Requirement: 批量操作错误处理
系统 SHALL 正确处理批量操作中的错误。

#### Scenario: 部分成功返回详细结果
- **WHEN** 批量操作中第 3 条 SQL 失败
- **THEN** 系统返回每条 SQL 的执行结果（成功或失败），而非整体失败

#### Scenario: 事务中的批量操作
- **WHEN** 批量操作在事务中执行
- **THEN** 任一 SQL 失败时，整个事务回滚
