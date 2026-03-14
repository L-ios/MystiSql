# WebSocket 支持实现增量规范

此增量规范定义 WebSocket 端点的具体实现要求，补充 `openspec/specs/websocket-support/spec.md` 规范。

## ADDED Requirements

### Requirement: WebSocket 路由注册
系统 SHALL 在 REST API 服务器中注册 WebSocket 路由。

#### Scenario: 注册 /ws 路由
- **WHEN** 服务器启动且 `websocket.enabled` 为 true
- **THEN** 系统在 `/ws` 路径注册 WebSocket 处理器

#### Scenario: 禁用 WebSocket
- **WHEN** 服务器启动且 `websocket.enabled` 为 false
- **THEN** 系统不注册 WebSocket 路由，返回 404

---

### Requirement: WebSocket 处理器实现
系统 SHALL 实现 WebSocket 连接处理器。

#### Scenario: 处理器结构
- **WHEN** 创建 WebSocket 处理器
- **THEN** 处理器包含以下依赖：
  - `authService`: 用于 Token 验证
  - `engine`: 用于执行 SQL 查询
  - `logger`: 用于日志记录

#### Scenario: 连接升级
- **WHEN** 客户端请求 WebSocket 连接
- **THEN** 系统使用 `gorilla/websocket.Upgrader` 升级 HTTP 连接

---

### Requirement: WebSocket 认证集成
系统 SHALL 在 WebSocket 握手时验证 JWT Token。

#### Scenario: URL 参数认证
- **WHEN** 客户端连接 `ws://host:port/ws?token=<jwt>`
- **THEN** 系统从 URL 查询参数提取 Token 并验证

#### Scenario: Token 无效
- **WHEN** Token 无效或已过期
- **THEN** 系统拒绝连接并返回 HTTP 401

---

### Requirement: WebSocket 消息处理
系统 SHALL 处理 JSON 格式的 WebSocket 消息。

#### Scenario: 处理 query 消息
- **WHEN** 收到 `{"action": "query", "instance": "...", "query": "...", "requestId": "..."}`
- **THEN** 系统执行查询并返回 `{"requestId": "...", "success": true, "columns": [...], "rows": [...]}`

#### Scenario: 处理 ping 消息
- **WHEN** 收到 `{"action": "ping"}`
- **THEN** 系统返回 `{"action": "pong"}`

#### Scenario: 查询执行失败
- **WHEN** SQL 执行失败
- **THEN** 系统返回 `{"requestId": "...", "success": false, "error": "错误描述"}`

---

### Requirement: WebSocket 配置集成
系统 SHALL 从配置文件读取 WebSocket 配置。

#### Scenario: 配置参数
- **WHEN** 初始化 WebSocket 服务
- **THEN** 系统读取以下配置：
  - `websocket.enabled`: 是否启用
  - `websocket.maxConnections`: 最大连接数
  - `websocket.idleTimeout`: 空闲超时
  - `websocket.maxConcurrentQueries`: 最大并发查询数
