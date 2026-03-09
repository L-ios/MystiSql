# jdbc-prepared-statement

## Purpose

TBD: PreparedStatement implementation for JDBC driver supporting parameterized queries.

## Requirements

### Requirement: PreparedStatement Interface

MystiSqlPreparedStatement SHALL实现`java.sql.PreparedStatement`接口，支持参数化查询。

#### Scenario: Create prepared statement

- **WHEN** 应用调用`connection.prepareStatement("SELECT * FROM users WHERE name = ? AND age > ?")`
- **THEN** 驱动SHALL返回PreparedStatement对象
- **AND** 解析SQL中的`?`占位符

#### Scenario: Set string parameter

- **WHEN** 应用调用`preparedStatement.setString(1, "Alice")`
- **THEN** 驱动SHALL记录第一个参数为VARCHAR类型的"Alice"

#### Scenario: Set int parameter

- **WHEN** 应用调用`preparedStatement.setInt(2, 18)`
- **THEN** 驱动SHALL记录第二个参数为INTEGER类型的18

#### Scenario: Set null parameter

- **WHEN** 应用调用`preparedStatement.setNull(1, Types.VARCHAR)`
- **THEN** 驱动SHALL记录第一个参数为NULL

### Requirement: Parameterized Query Execution

PreparedStatement SHALL使用参数化查询执行，防止SQL注入。

#### Scenario: Execute parameterized query

- **WHEN** 应用调用`preparedStatement.executeQuery()`
- **THEN** 驱动SHALL发送POST请求到`/api/v1/query`
- **AND** 请求体包含：
  ```json
  {
    "instance": "production-mysql",
    "query": "SELECT * FROM users WHERE name = ? AND age > ?",
    "parameters": [
      {"type": "VARCHAR", "value": "Alice"},
      {"type": "INTEGER", "value": 18}
    ]
  }
  ```
- **AND** 参数化处理在Gateway侧完成，确保安全

#### Scenario: Prevent SQL injection

- **WHEN** 应用执行`setString(1, "admin' OR '1'='1")`
- **THEN** 参数值SHALL原样传递到Gateway
- **AND** Gateway使用PreparedStatement机制防止注入
- **AND** SQL注入攻击SHALL被阻止

### Requirement: Parameter Data Types

PreparedStatement SHALL支持所有常用JDBC数据类型。

#### Scenario: Set VARCHAR parameter

- **WHEN** 调用`setString(index, value)`
- **THEN** 参数类型为VARCHAR
- **AND** 请求体中`type: "VARCHAR"`

#### Scenario: Set INTEGER parameter

- **WHEN** 调用`setInt(index, value)`
- **THEN** 参数类型为INTEGER
- **AND** 请求体中`type: "INTEGER"`

#### Scenario: Set BIGINT parameter

- **WHEN** 调用`setLong(index, value)`
- **THEN** 参数类型为BIGINT

#### Scenario: Set DOUBLE parameter

- **WHEN** 调用`setDouble(index, value)`
- **THEN** 参数类型为DOUBLE

#### Scenario: Set BOOLEAN parameter

- **WHEN** 调用`setBoolean(index, value)`
- **THEN** 参数类型为BOOLEAN

#### Scenario: Set DATE parameter

- **WHEN** 调用`setDate(index, value)`
- **THEN** 参数类型为DATE
- **AND** 日期格式为ISO 8601

#### Scenario: Set TIMESTAMP parameter

- **WHEN** 调用`setTimestamp(index, value)`
- **THEN** 参数类型为TIMESTAMP

#### Scenario: Set DECIMAL parameter

- **WHEN** 调用`setBigDecimal(index, value)`
- **THEN** 参数类型为DECIMAL

### Requirement: PreparedStatement Execution Methods

PreparedStatement SHALL支持多种执行方法。

#### Scenario: Execute query

- **WHEN** 调用`preparedStatement.executeQuery()`
- **THEN** 返回ResultSet
- **AND** 用于SELECT语句

#### Scenario: Execute update

- **WHEN** 调用`preparedStatement.executeUpdate()`
- **THEN** 返回受影响行数（int）
- **AND** 用于INSERT/UPDATE/DELETE

#### Scenario: Execute large update

- **WHEN** 调用`preparedStatement.executeLargeUpdate()`
- **THEN** 返回受影响行数（long）
- **AND** 支持大批量操作

#### Scenario: Execute generic

- **WHEN** 调用`preparedStatement.execute()`
- **THEN** 返回true表示有ResultSet，false表示更新
- **AND** 通过`getResultSet()`或`getUpdateCount()`获取结果

### Requirement: Clear Parameters

PreparedStatement SHALL支持清除参数。

#### Scenario: Clear parameters

- **WHEN** 调用`preparedStatement.clearParameters()`
- **THEN** 所有已设置的参数SHALL被清除
- **AND** PreparedStatement可重新设置参数并执行

#### Scenario: Reuse prepared statement

- **WHEN** 执行查询后，调用`clearParameters()`并设置新参数
- **THEN** PreparedStatement SHALL可复用
- **AND** 性能优于创建新PreparedStatement

### Requirement: Parameter Index Validation

PreparedStatement SHALL验证参数索引的有效性。

#### Scenario: Invalid parameter index (zero)

- **WHEN** 调用`setString(0, "value")`
- **THEN** SHALL抛出SQLException
- **AND** 错误信息说明索引从1开始

#### Scenario: Invalid parameter index (too large)

- **WHEN** 调用`setString(100, "value")`但SQL只有2个占位符
- **THEN** MAY抛出SQLException（或延迟到执行时）

### Requirement: Batch Operations (Phase 3)

PreparedStatement SHALL支持批量操作（Phase 3实现）。

#### Scenario: Add batch

- **WHEN** 调用`preparedStatement.addBatch()`
- **THEN** 当前参数设置SHALL添加到批处理队列

#### Scenario: Execute batch

- **WHEN** 调用`preparedStatement.executeBatch()`
- **THEN** 驱动SHALL发送POST请求到`/api/v1/batch`
- **AND** 返回每条语句的更新计数数组

#### Scenario: Clear batch

- **WHEN** 调用`preparedStatement.clearBatch()`
- **THEN** 批处理队列SHALL被清空
