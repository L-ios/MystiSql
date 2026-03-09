# WebSocket 支持规范

## Purpose

定义 MystiSql 的 WebSocket 功能，提供实时双向通信能力，支持客户端通过 WebSocket 连接执行 SQL 查询，实现更高效的实时交互体验。

## Requirements

### Requirement: WebSocket 连接建立
系统 SHALL 提供 WebSocket 端点支持实时交互。

#### Scenario: WebSocket 握手
- **WHEN** 客户端连接 `ws://host:port/ws`
- **THEN** 系统完成 WebSocket 握手并建立连接

#### Scenario: WebSocket 使用 Token 认证
- **WHEN** 客户端在 URL 中传递 `ws://host:port/ws?token=<jwt_token>`
- **THEN** 系统验证 Token 并建立连接

#### Scenario: 认证失败拒绝连接
- **WHEN** 客户端提供无效或过期的 Token
- **THEN** 系统拒绝 WebSocket 连接并返回 401 错误

---

### Requirement: WebSocket 消息格式
系统 SHALL 使用 JSON 格式传输 WebSocket 消息。

#### Scenario: 查询请求消息格式
- **WHEN** 客户端发送查询请求
- **THEN** 消息格式为 `{"action": "query", "instance": "local-mysql", "query": "SELECT * FROM users"}`

#### Scenario: 查询响应消息格式
- **WHEN** 服务端返回查询结果
- **THEN** 消息格式为 `{"type": "result", "columns": [...], "rows": [...], "rowCount": 10}`

#### Scenario: 错误响应消息格式
- **WHEN** 查询执行失败
- **THEN** 消息格式为 `{"type": "error", "message": "错误描述", "code": "ERROR_CODE"}`

---

### Requirement: WebSocket 执行查询
系统 SHALL 通过 WebSocket 执行 SQL 查询。

#### Scenario: WebSocket 执行 SELECT
- **WHEN** 客户端发送 `{"action": "query", "instance": "local-mysql", "query": "SELECT * FROM users LIMIT 5"}`
- **THEN** 系统执行查询并返回结果

#### Scenario: WebSocket 执行 INSERT
- **WHEN** 客户端发送 `{"action": "query", "instance": "local-mysql", "query": "INSERT INTO users (name) VALUES ('Alice')"}`
- **THEN** 系统执行插入并返回影响行数

---

### Requirement: WebSocket 连接管理
系统 SHALL 管理 WebSocket 连接的生命周期。

#### Scenario: 最大连接数限制
- **WHEN** WebSocket 连接数达到 `websocket.maxConnections`（默认 1000）
- **THEN** 系统拒绝新的连接并返回 503 错误

#### Scenario: 连接空闲超时
- **WHEN** WebSocket 连接空闲超过 `websocket.idleTimeout`（默认 10 分钟）
- **THEN** 系统自动关闭连接

#### Scenario: 心跳机制
- **WHEN** 客户端发送 `{"action": "ping"}`
- **THEN** 系统返回 `{"type": "pong"}`

---

### Requirement: WebSocket 并发控制
系统 SHALL 限制单个 WebSocket 连接的并发查询数。

#### Scenario: 并发查询限制
- **WHEN** 单个连接同时发送多个查询
- **THEN** 系统限制并发数为 `websocket.maxConcurrentQueries`（默认 5）

---

### Requirement: WebSocket 安全
系统 SHALL 为 WebSocket 连接提供安全保障。

#### Scenario: 支持 TLS
- **WHEN** 配置启用 TLS（`wss://`）
- **THEN** WebSocket 连接使用加密传输

#### Scenario: 验证用户权限
- **WHEN** WebSocket 连接建立后执行查询
- **THEN** 系统验证用户对该实例的访问权限
