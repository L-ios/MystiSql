## MODIFIED Requirements

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

#### Scenario: 查询 - 未认证

- **WHEN** 调用 POST /api/v1/query 但未提供有效的认证 Token
- **THEN** 必须返回 401 Unauthorized 状态
- **AND** 响应必须包含错误消息："未提供有效的认证 Token"

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

#### Scenario: 列出实例 - 未认证

- **WHEN** 调用 GET /api/v1/instances 但未提供有效的认证 Token
- **THEN** 必须返回 401 Unauthorized 状态
- **AND** 响应必须包含错误消息："未提供有效的认证 Token"

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

#### Scenario: 健康检查无需认证

- **WHEN** 调用 GET /health
- **THEN** 必须不要求认证 Token
- **AND** 直接返回健康状态
