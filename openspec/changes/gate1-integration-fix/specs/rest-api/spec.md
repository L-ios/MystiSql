## MODIFIED Requirements

### Requirement: WebSocket 端点
系统 SHALL 使用 `internal/api/websocket/` 包处理 WebSocket 端点。

#### Scenario: WebSocket 连接
- **WHEN** 连接到 ws://host:port/ws?token=<jwt>
- **THEN** 由 `websocket.WebSocketHandler` 处理连接
- **AND** 通过 JWT token 参数认证（支持 query param 和 Bearer header）
- **AND** 连接数不超过配置的 maxConnections

#### Scenario: WebSocket 消息格式
- **WHEN** 通过 WebSocket 发送消息
- **THEN** 消息必须是 JSON 格式
- **AND** 必须包含：action (query/ping)
- **AND** 响应使用 type 字段区分消息类型（query_result/error/pong）

#### Scenario: WebSocket 优雅关闭
- **WHEN** 服务器 shutdown
- **THEN** 所有活跃 WebSocket 连接被关闭
- **AND** ConnectionManager 的 idle 清理 goroutine 停止

### Requirement: 优雅关闭
服务关闭时 SHALL 同时关闭所有 WebSocket 连接。

#### Scenario: 完整优雅关闭流程
- **WHEN** 收到 SIGTERM 或 SIGINT 信号
- **THEN** 服务器停止接受新 HTTP 请求
- **AND** 关闭所有活跃 WebSocket 连接
- **AND** 等待正在进行的请求完成（最多 30 秒）
