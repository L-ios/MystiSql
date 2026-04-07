## ADDED Requirements

### Requirement: Engine 驱动工厂动态注册
Engine SHALL 通过 `connection.DriverRegistry` 按实例类型动态查找 ConnectionFactory，而非硬编码 MySQL/PostgreSQL 工厂。

#### Scenario: Oracle 实例查询可达
- **WHEN** 注册 Oracle 驱动到 DriverRegistry，用户通过 API 查询 Oracle 实例
- **THEN** Engine 成功创建连接并返回查询结果

#### Scenario: Redis 实例查询可达
- **WHEN** 注册 Redis 驱动到 DriverRegistry，用户通过 API 查询 Redis 实例
- **THEN** Engine 成功创建连接并返回查询结果

#### Scenario: 未注册的数据库类型返回明确错误
- **WHEN** 用户查询的实例类型未注册驱动
- **THEN** 返回错误 "不支持的数据库类型: <type>"，而非 panic

#### Scenario: 驱动在启动时统一注册
- **WHEN** serve 命令启动
- **THEN** 所有已实现的驱动（MySQL, PostgreSQL, Oracle, Redis, SQLite, MSSQL）在 Engine 使用前完成注册

### Requirement: Health Monitor 启动
Health Monitor SHALL 在服务启动时实例化并运行，定期检查所有实例的健康状态。

#### Scenario: 服务启动时 Monitor 自动运行
- **WHEN** serve 命令启动
- **THEN** Health Monitor 调用 Start() 开始定期健康检查

#### Scenario: 服务关闭时 Monitor 停止
- **WHEN** 服务收到 shutdown 信号
- **THEN** Health Monitor 调用 Stop() 优雅停止

### Requirement: RBAC 路由注册
RBAC CRUD 路由 SHALL 在 server.go 的 setupRoutes() 中注册到 HTTP 路由。

#### Scenario: 创建角色路由可达
- **WHEN** POST /api/v1/rbac/roles 请求包含角色定义
- **THEN** 返回非 404 状态码（即路由已注册）

#### Scenario: 列出角色路由可达
- **WHEN** GET /api/v1/rbac/roles
- **THEN** 返回角色列表（非 404）

### Requirement: RBAC 中间件从 JWT claims 读角色
RBAC PermissionMiddleware SHALL 从 gin context 读取角色信息，而非从客户端 Header 读取。

#### Scenario: 从 JWT claims 提取角色
- **WHEN** auth middleware 将 JWT claims 中的 role 写入 gin context
- **THEN** RBAC PermissionMiddleware 从 context 读取角色进行权限检查

#### Scenario: 拒绝伪造 Header 角色
- **WHEN** 客户端设置 X-User-Roles: admin 但 JWT claims 中 role 为 viewer
- **THEN** RBAC 使用 viewer 角色进行权限检查，忽略 Header
