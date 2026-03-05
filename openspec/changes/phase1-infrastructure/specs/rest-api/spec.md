## ADDED Requirements

### Requirement: REST API 框架设置

系统必须使用 Gin 框架提供 RESTful API。

#### Scenario: API 服务器启动

- **WHEN** API 服务器启动
- **THEN** 必须在配置的端口上监听（默认 8080）
- **AND** 必须处理 SIGTERM/SIGINT 的优雅关闭
- **AND** 必须在启动时记录监听地址

#### Scenario: API 服务器配置

- **WHEN** 使用自定义配置启动 API 服务器
- **THEN** 必须使用配置中指定的端口
- **AND** 必须使用配置中指定的主机绑定
- **AND** 默认主机必须是 0.0.0.0（监听所有接口）

#### Scenario: 优雅关闭

- **WHEN** 收到 SIGTERM 或 SIGINT 信号
- **THEN** 服务器必须停止接受新请求
- **AND** 必须等待正在进行的请求完成（最多 30 秒）
- **AND** 必须在超时后强制关闭

---

### Requirement: 健康检查端点

系统必须提供健康检查端点用于监控。

#### Scenario: 健康检查成功

- **WHEN** 调用 GET /health
- **THEN** 必须返回 200 OK 状态
- **AND** 响应必须包含：{ "status": "healthy", "timestamp": "..." }
- **AND** Content-Type 必须是 application/json

#### Scenario: 健康检查包含实例连接性

- **WHEN** 调用 GET /health?check-instances=true
- **THEN** 系统必须测试到所有已注册实例的连接性
- **AND** 响应必须包含每个实例的健康状态
- **AND** 如果任何实例不健康，必须返回 503 状态

#### Scenario: 健康检查响应格式

- **WHEN** 调用健康检查端点
- **THEN** 响应必须是 JSON 格式
- **AND** 必须包含以下字段：
  - status: "healthy" 或 "unhealthy"
  - timestamp: ISO 8601 格式的时间戳
  - version: 服务版本
  - instances: 实例状态数组（如果 check-instances=true）

---

### Requirement: Query 端点

系统必须提供端点来执行 SQL 查询。

#### Scenario: 成功执行查询

- **WHEN** 使用有效的实例名和 SQL 调用 POST /api/v1/query
- **THEN** 必须在指定实例上执行查询
- **AND** 必须以 JSON 格式返回结果，包含 columns 和 rows
- **AND** 必须返回 200 OK 状态
- **AND** 必须在响应中包含查询执行时间

#### Scenario: 请求体格式

- **WHEN** 调用 POST /api/v1/query
- **THEN** 请求体必须是 JSON 格式
- **AND** 必须包含字段：instance（实例名）、sql（查询语句）
- **AND** 可选字段：timeout（超时时间，秒）

#### Scenario: 查询 - 实例不存在

- **WHEN** 使用不存在的实例名调用 POST /api/v1/query
- **THEN** 必须返回 404 Not Found 状态
- **AND** 响应必须包含错误消息："实例未找到：<name>"

#### Scenario: 查询 - 无效 SQL

- **WHEN** 使用无效的 SQL 语法调用 POST /api/v1/query
- **THEN** 必须返回 400 Bad Request 状态
- **AND** 响应必须包含 SQL 错误消息

#### Scenario: 查询 - 超时

- **WHEN** POST /api/v1/query 包含 timeout 参数且查询超时
- **THEN** 必须取消查询
- **AND** 必须返回 408 Request Timeout 状态
- **AND** 响应必须包含错误消息："查询超时"

#### Scenario: 查询结果格式

- **WHEN** 查询成功返回结果
- **THEN** 响应必须是 JSON 格式
- **AND** 必须包含以下字段：
  - columns: 列名数组
  - rows: 行数据数组
  - rowCount: 行数
  - executionTimeMs: 执行时间（毫秒）

---

### Requirement: 实例列表端点

系统必须提供端点列出可用的数据库实例。

#### Scenario: 成功列出实例

- **WHEN** 调用 GET /api/v1/instances
- **THEN** 必须返回所有已注册的实例
- **AND** 每个实例必须包含：名称、类型、主机、端口、状态
- **AND** 必须返回 200 OK 状态

#### Scenario: 列出实例 - 无实例

- **WHEN** 调用 GET /api/v1/instances 但没有注册实例
- **THEN** 必须返回空数组 []
- **AND** 必须返回 200 OK 状态

#### Scenario: 实例信息脱敏

- **WHEN** 返回实例信息
- **THEN** 密码字段必须被脱敏（显示为 "******"）
- **AND** 必须不包含敏感配置信息

---

### Requirement: 错误处理

系统必须提供一致的错误响应。

#### Scenario: 内部服务器错误

- **WHEN** 发生意外错误
- **THEN** 必须返回 500 Internal Server Error 状态
- **AND** 响应必须包含错误消息
- **AND** 系统必须记录错误的详细信息
- **AND** 响应不得包含堆栈跟踪或敏感信息

#### Scenario: 请求验证错误

- **WHEN** 请求包含无效参数或缺少必填字段
- **THEN** 必须返回 400 Bad Request 状态
- **AND** 响应必须包含验证错误详情
- **AND** 必须指出哪个字段无效以及原因

#### Scenario: 错误响应格式

- **WHEN** 返回错误响应
- **THEN** 响应必须是 JSON 格式
- **AND** 必须包含以下字段：
  - error: true
  - message: 错误消息
  - code: 错误代码（可选）
  - details: 详细信息（可选）

#### Scenario: 内容类型错误

- **WHEN** POST 请求的 Content-Type 不是 application/json
- **THEN** 必须返回 415 Unsupported Media Type 状态
- **AND** 响应必须指明需要 application/json

---

### Requirement: 请求/响应格式

系统必须使用 JSON 作为所有 API 请求和响应的格式。

#### Scenario: JSON 请求体

- **WHEN** 发送 POST/PUT 请求
- **THEN** 请求体必须是 JSON 格式
- **AND** Content-Type 必须是 application/json
- **AND** 无效的 JSON 必须返回 400 错误

#### Scenario: JSON 响应格式

- **WHEN** 返回响应
- **THEN** 响应必须是 JSON 格式
- **AND** Content-Type 必须是 application/json
- **AND** 响应必须遵循一致的结构

#### Scenario: 标准响应结构

- **WHEN** 返回成功响应
- **THEN** 响应必须包含以下结构：
  ```json
  {
    "data": { ... },           // 实际数据
    "error": false,             // 是否有错误
    "metadata": {               // 元数据（可选）
      "timestamp": "...",
      "executionTimeMs": 123
    }
  }
  ```

---

### Requirement: CORS 支持

系统必须支持跨域资源共享（CORS）用于浏览器访问。

#### Scenario: CORS 预检请求

- **WHEN** 收到 OPTIONS 请求
- **THEN** 必须返回适当的 CORS 头
- **AND** 必须包括：Access-Control-Allow-Origin、Access-Control-Allow-Methods、Access-Control-Allow-Headers

#### Scenario: CORS 实际请求

- **WHEN** 从浏览器发送请求
- **THEN** 响应必须包括 Access-Control-Allow-Origin 头
- **AND** 默认必须允许所有来源（*）
- **AND** 可以通过配置限制允许的来源

---

### Requirement: 请求日志

系统必须记录所有 API 请求。

#### Scenario: 记录请求信息

- **WHEN** 收到 API 请求
- **THEN** 必须记录：方法、路径、状态码、响应时间
- **AND** 必须使用结构化日志格式
- **AND** 不得记录请求体中的敏感信息（如密码）

#### Scenario: 记录错误请求

- **WHEN** 请求导致错误
- **THEN** 必须记录完整的错误详情
- **AND** 必须包括请求 ID（便于追踪）
- **AND** 必须包括客户端 IP 地址

---

### Requirement: 请求限流

系统必须支持基本的请求限流以防止滥用。

#### Scenario: 启用限流

- **WHEN** 配置了限流参数
- **THEN** 系统必须限制每个 IP 的请求速率
- **AND** 超过限制必须返回 429 Too Many Requests 状态
- **AND** 响应必须包括 Retry-After 头

#### Scenario: 默认限流设置

- **WHEN** 未配置限流参数
- **THEN** 默认限制必须是每个 IP 每秒 100 个请求
- **AND** 可以通过配置调整

---

### Requirement: API 版本控制

系统必须支持 API 版本控制。

#### Scenario: 版本化端点

- **WHEN** 访问 API 端点
- **THEN** 所有端点必须以版本前缀开始（如 /api/v1/）
- **AND** 当前版本必须是 v1

#### Scenario: 未来版本支持

- **WHEN** 引入新的 API 版本
- **THEN** 旧版本必须保持兼容
- **AND** 必须支持多版本共存
