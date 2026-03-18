# jdbc-websocket-client

## Purpose
MystiSql JDBC 驱动的 WebSocket 客户端实现，用于支持实时双向通信。

## Requirements
### Requirement: WebSocket 连接建立
JDBC 驱动 SHALL 支持通过 WebSocket 连接到 MystiSql Gateway。
#### Scenario: 建立基本 WebSocket 连接
- **WHEN** 连接 URL 为 `jdbc:mystisql://gateway:8080/instance?transport=ws`
- **THEN** 驱动 SHALL 建立 WebSocket 连接到 `ws://gateway:8080/ws`
- **AND** 连接成功后返回有效的 Connection 对象
#### Scenario: WebSocket 使用 Token 认证
- **WHEN** 连接 URL 包含 token 参数 `jdbc:mystisql://gateway:8080/instance?transport=ws&token=abc123`
- **THEN** 驱动 SHALL 连接到 `ws://gateway:8080/ws?token=abc123`
- **AND** Gateway 验证 Token 后建立连接
#### Scenario: WebSocket 使用 WSS 加密连接
- **WHEN** 连接 URL 为 `jdbc:mystisql://gateway:8080/instance?transport=ws&ssl=true`
- **THEN** 驱动 SHALL 连接到 `wss://gateway:8080/ws`
- **AND** 使用 TLS 加密传输
#### Scenario: WebSocket 连接失败回退
- **WHEN** WebSocket 连接失败（网络错误、认证失败等)
- **THEN** 驱动 SHALL 抛出 SQLException
- **AND** 错误信息包含失败原因
### Requirement: WebSocket 消息格式
驱动 SHALL 使用 Gateway 定义的 JSON 消息格式进行通信。
#### Scenario: 发送查询请求
- **WHEN** 应用执行 `statement.executeQuery("SELECT * FROM users")`
- **THEN** 驱动 SHALL 发送 WebSocket 消息：
  ```json
  {
    "action": "query",
    "instance": "instance-name",
    "query": "SELECT * FROM users",
    "requestId": "uuid"
  }
  ```
#### Scenario: 接收查询结果
- **WHEN** Gateway 返回查询结果
- **THEN** 驱动 SHALL 解析响应消息
- **AND** 创建 ResultSet 对象
#### Scenario: 接收错误响应
- **WHEN** Gateway 返回错误
- **THEN** 驱动 SHALL 抛出 SQLException
- **AND** message 包含错误信息
### Requirement: WebSocket 连接复用
驱动 SHALL 复用 WebSocket 连接以提高性能。
#### Scenario: 复用已建立的连接
- **WHEN** 同一个 Connection 对象执行多次查询
- **THEN** 驱动 SHALL 复用同一个 WebSocket 连接
- **AND** 不重复建立新连接
#### Scenario: 连接关闭后重建
- **WHEN** WebSocket 连接被意外关闭
- **THEN** 下次查询时驱动 SHALL 自动重建连接
- **AND** 重试最多 3 次
### Requirement: WebSocket 心跳保活
驱动 SHALL 发送心跳消息保持连接活跃。
#### Scenario: 定期发送心跳
- **WHEN** WebSocket 连接空闲超过 30 秒
- **THEN** 驱动 SHALL 发送心跳消息 `{"action": "ping"}`
- **AND** 等待响应 `{"action": "pong"}`
#### Scenario: 心跳超时处理
- **WHEN** 心跳响应超时（10 秒内未收到 pong)
- **THEN** 驱动 SHALL 标记连接为不健康
- **AND** 下次查询时重建连接
### Requirement: WebSocket 连接验证
驱动 SHALL 支持 `Connection.isValid()` 方法。
#### Scenario: 验证 WebSocket 连接有效
- **WHEN** 应用调用 `connection.isValid(5)` 且使用 WebSocket 传输
- **THEN** 驱动 SHALL 发送 ping 消息
- **AND** 在 5 秒内收到 pong 响应时返回 true
- **AND** 超时返回 false
### Requirement: WebSocket 自动重连
驱动 SHALL 在连接断开时自动重连。
#### Scenario: 网络抖动自动重连
- **WHEN** WebSocket 连接因网络原因断开
- **THEN** 驱动 SHALL 自动尝试重连
- **AND** 第 1 次重连间隔 1 秒
- **AND** 第 2 次重连间隔 2 秒
- **AND** 第 3 次重连间隔 4 秒
- **AND** 3 次失败后抛出 SQLException
#### Scenario: Token 过期重新认证
- **WHEN** Gateway 返回 Token 过期错误
- **THEN** 驱动 SHALL 尝试使用新 Token 重新连接（如果提供了 Token 刷新回调）
