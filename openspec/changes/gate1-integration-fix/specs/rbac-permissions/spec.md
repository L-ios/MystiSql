## MODIFIED Requirements

### Requirement: 权限检查
系统 SHALL 在执行查询前检查用户权限。

#### Scenario: 有权限执行
- **WHEN** 用户有 `mysql-prod:users:SELECT` 权限
- **THEN** 查询被允许执行

#### Scenario: 无权限拒绝
- **WHEN** 用户没有 `mysql-prod:users:DELETE` 权限
- **THEN** 系统拒绝请求并返回 403

#### Scenario: 角色来源为 JWT claims
- **WHEN** auth middleware 将 JWT claims 中的 role 写入 gin context（c.Set("role", claims.Role)）和 roles（c.Set("roles", []string{claims.Role})）
- **THEN** RBAC PermissionMiddleware 从 gin context 读取 roles（c.Get("roles")），不读取 X-User-Roles Header

#### Scenario: RBAC 路由已注册到 server
- **WHEN** server.go 的 setupRoutes() 执行
- **THEN** RBAC CRUD 路由（POST/GET/PUT/DELETE /api/v1/rbac/roles, POST /api/v1/rbac/users/:id/roles）注册到路由器

### Requirement: 权限管理 API
系统 SHALL 提供权限管理 REST API。

#### Scenario: 查询用户权限
- **WHEN** GET /api/v1/rbac/users/{userId}/permissions
- **THEN** 返回用户的所有权限列表

#### Scenario: 分配角色给用户
- **WHEN** POST /api/v1/rbac/users/{userId}/roles
- **THEN** 用户获得指定角色
