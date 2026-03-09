# jdbc-api-client

## Purpose

TBD: RESTful API client for JDBC driver to communicate with MystiSql Gateway.

## Requirements

### Requirement: RESTful API Client

JDBC驱动SHALL通过HTTP RESTful API与MystiSql Gateway通信。

#### Scenario: Execute query via REST

- **WHEN** JDBC应用执行`Statement.executeQuery("SELECT * FROM users")`
- **THEN** 驱动SHALL发送POST请求到`http://gateway:8080/api/v1/query`
- **AND** 请求体包含：`{"instance": "production-mysql", "query": "SELECT * FROM users"}`
- **AND** 使用OkHttp连接池

#### Scenario: Execute update via REST

- **WHEN** JDBC应用执行`Statement.executeUpdate("UPDATE users SET name='Bob'")`
- **THEN** 驱动SHALL发送POST请求到`http://gateway:8080/api/v1/exec`
- **AND** 返回受影响的行数

#### Scenario: Handle HTTP errors

- **WHEN** Gateway返回HTTP 500或连接失败
- **THEN** 驱动SHALL抛出SQLException
- **AND** SQLException包含HTTP状态码和错误信息

### Requirement: Authentication Integration

驱动SHALL支持MystiSql的Token认证机制。

#### Scenario: Authenticate with token in URL

- **WHEN** 连接URL包含`?token=abc123`
- **THEN** 驱动SHALL在所有HTTP请求头中添加`Authorization: Bearer abc123`

#### Scenario: Authenticate with password as token

- **WHEN** 连接时传入`password="abc123"`
- **THEN** 驱动SHALL将password作为token使用
- **AND** 在HTTP请求头中添加`Authorization: Bearer abc123`

#### Scenario: Token is sent on every request

- **WHEN** 已认证的Connection执行查询
- **THEN** 每个HTTP请求都SHALL包含Authorization头

### Requirement: HTTP Connection Pool

HTTP客户端SHALL使用连接池优化性能。

#### Scenario: Reuse HTTP connections

- **WHEN** 同一个Connection对象执行多次查询
- **THEN** OkHttp连接池SHALL复用TCP连接
- **AND** 减少连接建立开销

#### Scenario: Configure connection pool

- **WHEN** URL包含`?maxConnections=20`
- **THEN** OkHttp连接池SHALL配置为最多20个连接

### Requirement: Request Timeout

驱动SHALL支持配置HTTP请求超时。

#### Scenario: Configure default timeout

- **WHEN** URL包含`?timeout=60`
- **THEN** HTTP请求超时SHALL设置为60秒

#### Scenario: Per-statement timeout override

- **WHEN** 应用调用`Statement.setQueryTimeout(30)`
- **THEN** 该Statement的请求SHALL使用30秒超时
- **AND** 不影响其他Statement

### Requirement: SSL/TLS Support

驱动SHALL支持HTTPS连接。

#### Scenario: Enable SSL via URL

- **WHEN** URL包含`?ssl=true`
- **THEN** HTTP客户端SHALL使用HTTPS协议

#### Scenario: Disable SSL verification (development)

- **WHEN** URL包含`?ssl=true&verifySsl=false`
- **THEN** HTTP客户端SHALL跳过证书验证
- **AND** 在日志中输出安全警告

### Requirement: JSON Response Parsing

驱动SHALL正确解析Gateway的JSON响应。

#### Scenario: Parse query result set

- **WHEN** Gateway返回：
  ```json
  {
    "columns": [
      {"name": "id", "type": "int"},
      {"name": "name", "type": "varchar"}
    ],
    "rows": [[1, "Alice"], [2, "Bob"]],
    "rowCount": 2
  }
  ```
- **THEN** 驱动SHALL创建ResultSet对象
- **AND** `ResultSet.next()`正确遍历所有行
- **AND** `ResultSet.getInt("id")`返回正确值

#### Scenario: Parse execution result

- **WHEN** Gateway返回：
  ```json
  {
    "rowsAffected": 5,
    "lastInsertId": 123
  }
  ```
- **THEN** `Statement.executeUpdate()`返回5
- **AND** `Statement.getGeneratedKeys()`包含lastInsertId

#### Scenario: Parse error response

- **WHEN** Gateway返回错误：
  ```json
  {
    "error": "Table not found",
    "code": "TABLE_NOT_FOUND"
  }
  ```
- **THEN** 驱动SHALL抛出SQLException
- **AND** message包含"Table not found"
- **AND** SQLState根据错误码映射

### Requirement: Error Code Mapping

驱动SHALL将Gateway错误码映射到SQLState。

#### Scenario: Map table not found

- **WHEN** Gateway返回错误码`TABLE_NOT_FOUND`
- **THEN** SQLException的SQLState SHALL为`"42S02"`

#### Scenario: Map syntax error

- **WHEN** Gateway返回错误码`SYNTAX_ERROR`
- **THEN** SQLException的SQLState SHALL为`"42000"`

#### Scenario: Map connection failed

- **WHEN** Gateway返回错误码`CONNECTION_FAILED`
- **THEN** SQLException的SQLState SHALL为`"08001"`

### Requirement: Request Logging

驱动SHALL记录HTTP请求日志（DEBUG级别）。

#### Scenario: Log successful requests

- **WHEN** HTTP请求成功
- **THEN** 驱动SHALL记录（DEBUG级别）：
  - 请求方法和URL
  - 响应状态码
  - 执行时间

#### Scenario: Log failed requests

- **WHEN** HTTP请求失败
- **THEN** 驱动SHALL记录（ERROR级别）：
  - 请求详情
  - 错误信息
  - 异常堆栈
