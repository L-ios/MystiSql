## ADDED Requirements

### Requirement: WebSocket 使用 internal/api/websocket 包
WebSocket 端点 SHALL 使用 `internal/api/websocket/` 包处理，替换 `rest/websocket_handlers.go`。

#### Scenario: WebSocket 连接使用新 handler
- **WHEN** 客户端连接 ws://host:port/ws?token=<jwt>
- **THEN** 由 `websocket.WebSocketHandler.Handle()` 处理连接（非 rest 包的旧实现）

#### Scenario: WebSocket 查询注入审计上下文
- **WHEN** 通过 WebSocket 执行 SQL 查询
- **THEN** 审计日志中的 user_id 为 JWT claims 中的实际用户 ID（非 "unknown"）

#### Scenario: WebSocket 查询注入客户端 IP
- **WHEN** 通过 WebSocket 执行 SQL 查询
- **THEN** 审计日志中的 client_ip 为实际客户端 IP

### Requirement: WebSocket 优雅关闭
Server.Shutdown() SHALL 关闭所有活跃的 WebSocket 连接。

#### Scenario: 服务关闭时关闭 WebSocket
- **WHEN** 服务收到 shutdown 信号
- **THEN** 调用 WebSocketHandler.Close() 关闭所有活跃连接
- **AND** idle 清理 goroutine 停止

### Requirement: WebSocket 配置从配置文件读取
WebSocket 配置 SHALL 从 config.yaml 读取，而非硬编码。

#### Scenario: 从配置读取最大连接数
- **WHEN** config.yaml 设置 websocket.maxConnections: 500
- **THEN** WebSocket handler 的最大连接数为 500

#### Scenario: 从配置读取空闲超时
- **WHEN** config.yaml 设置 websocket.idleTimeout: "5m"
- **THEN** 空闲连接在 5 分钟后被关闭

#### Scenario: 配置缺失时使用默认值
- **WHEN** config.yaml 未设置 websocket 段
- **THEN** 使用默认值：maxConnections=100, idleTimeout=5m

### Requirement: WebSocket 消息协议 JSON tag 使用 camelCase
WebSocket 消息结构体的 JSON tag SHALL 使用 camelCase 命名。

#### Scenario: 请求消息使用 camelCase
- **WHEN** 客户端发送 WebSocket 消息 {"requestId":"req-1","action":"query","instance":"mysql","query":"SELECT 1"}
- **THEN** 服务端正确解析 requestId 和 action 字段

#### Scenario: 响应消息使用 camelCase
- **WHEN** 服务端返回查询结果
- **THEN** 响应包含 camelCase 字段：requestId, rowCount, executionTimeMs

### Requirement: JDBC 客户端适配 WebSocket 新协议
JDBC WebSocketTransport SHALL 适配新的消息协议格式（type 字段替代 action/success）。

#### Scenario: 查询成功判断改用 type 字段
- **WHEN** JDBC 收到 WebSocket 响应 {"type":"query_result","requestId":"req-1",...}
- **THEN** 判断为查询成功，解析 columns/rows/rowCount

#### Scenario: 错误响应判断改用 type 字段
- **WHEN** JDBC 收到 WebSocket 响应 {"type":"error","code":"...","message":"..."}
- **THEN** 判断为查询失败，提取 error message

#### Scenario: Pong 心跳判断改用 type 字段
- **WHEN** JDBC 收到 WebSocket 响应 {"type":"pong","timestamp":1234567890}
- **THEN** 判断为心跳响应，记录日志
