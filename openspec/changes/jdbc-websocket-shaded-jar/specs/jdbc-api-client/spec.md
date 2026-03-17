## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Transport Layer Abstraction

JDBC 驱动 SHALL 使用传输层抽象支持多种通信方式。

#### Scenario: WebSocket 传输（默认）

- **WHEN** 连接 URL 不指定 transport 参数
- **THEN** 驱动 SHALL 默认使用 WebSocketClient 传输
- **AND** 建立长连接进行通信

#### Scenario: 强制使用 HTTP 传输

- **WHEN** 连接 URL 指定 `transport=http`
- **THEN** 驱动 SHALL 只使用 RestClient（HTTP）传输
- **AND** 行为与现有实现一致

#### Scenario: 传输层切换

- **WHEN** 用户需要切换传输方式
- **THEN** 用户 SHALL 修改连接 URL 参数
- **AND** 无需修改应用代码

### Requirement: Transport Interface

驱动 SHALL 定义 Transport 接口抽象传输层。

#### Scenario: Transport 接口方法

- **WHEN** 定义 Transport 接口
- **THEN** 接口 SHALL 包含以下方法：
  ```java
  public interface Transport extends AutoCloseable {
      QueryResult query(QueryRequest request) throws SQLException;
      ExecResult exec(QueryRequest request) throws SQLException;
      boolean isValid(int timeout);
      void close();
  }
  ```

#### Scenario: RestClient 实现 Transport

- **WHEN** RestClient 实现 Transport 接口
- **THEN** `query()` 方法 SHALL 发送 HTTP POST 到 `/api/v1/query`
- **AND** `exec()` 方法 SHALL 发送 HTTP POST 到 `/api/v1/exec`
- **AND** `isValid()` 方法 SHALL 调用 `/health` 端点

#### Scenario: WebSocketClient 实现 Transport

- **WHEN** WebSocketClient 实现 Transport 接口
- **THEN** `query()` 方法 SHALL 发送 WebSocket 消息 `{"action": "query", ...}`
- **AND** `exec()` 方法 SHALL 发送 WebSocket 消息 `{"action": "query", ...}`
- **AND** `isValid()` 方法 SHALL 发送心跳 `{"action": "ping"}`

### Requirement: Connection Transport Selection

MystiSqlConnection SHALL 根据连接参数选择传输层。

#### Scenario: 解析 transport 参数

- **WHEN** 连接 URL 为 `jdbc:mystisql://host:8080/instance?transport=ws`
- **THEN** Connection SHALL 解析出 `transport=ws`
- **AND** 创建 WebSocketClient 作为传输层

#### Scenario: 默认 HTTP 传输

- **WHEN** 连接 URL 为 `jdbc:mystisql://host:8080/instance`（无 transport 参数）
- **THEN** Connection SHALL 使用 RestClient 作为传输层
- **AND** 保持向后兼容

### Requirement: Authentication Integration

驱动SHALL支持MystiSql的Token认证机制。

#### Scenario: Authenticate with token in URL

- **WHEN** 连接URL包含`?token=abc123`
- **THEN** 驱动SHALL在所有请求中携带认证信息
- **AND** HTTP 传输添加 `Authorization: Bearer abc123` 请求头
- **AND** WebSocket 传输在 URL 中添加 `?token=abc123`

#### Scenario: Authenticate with password as token

- **WHEN** 连接时传入`password="abc123"`
- **THEN** 驱动SHALL将password作为token使用
- **AND** HTTP 传输添加 `Authorization: Bearer abc123` 请求头
- **AND** WebSocket 传输在 URL 中添加 `?token=abc123`

#### Scenario: Token is sent on every request

- **WHEN** 已认证的Connection执行查询
- **THEN** 每个请求都SHALL包含认证信息
