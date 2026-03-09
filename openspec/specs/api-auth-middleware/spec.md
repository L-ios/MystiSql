# API 认证中间件规范

## Purpose

定义 MystiSql 的 API 认证中间件功能，为所有 API 端点（除白名单路径外）添加统一的认证保护，确保只有经过认证的请求才能访问受保护的资源。

## Requirements

### Requirement: API 全局认证中间件
系统 SHALL 为所有 API 端点添加认证中间件（除白名单路径外）。

#### Scenario: 受保护端点需要认证
- **WHEN** 客户端访问 `/api/v1/query` 且未提供 Token
- **THEN** 系统返回 401 Unauthorized

#### Scenario: 白名单路径无需认证
- **WHEN** 客户端访问 `/health` 端点
- **THEN** 系统不要求认证，直接返回健康状态

#### Scenario: 认证成功继续请求
- **WHEN** 客户端访问 `/api/v1/query` 并提供有效 Token
- **THEN** 系统验证通过并继续处理请求

---

### Requirement: 认证中间件实现
系统 SHALL 使用 Gin 中间件实现认证逻辑。

#### Scenario: 中间件提取 Token
- **WHEN** 请求到达认证中间件
- **THEN** 中间件从 `Authorization` Header 或 URL 参数中提取 Token

#### Scenario: 中间件验证 Token
- **WHEN** 中间件提取到 Token
- **THEN** 中间件验证 Token 签名和有效期

#### Scenario: 中间件注入用户信息
- **WHEN** Token 验证通过
- **THEN** 中间件将用户信息（user_id、role）注入到 `gin.Context`

---

### Requirement: 白名单路径配置
系统 SHALL 支持配置白名单路径（无需认证）。

#### Scenario: 默认白名单路径
- **WHEN** 未配置白名单
- **THEN** 默认白名单为 `/health` 和 `/api/v1/auth/login`

#### Scenario: 自定义白名单路径
- **WHEN** 配置 `auth.whitelist` 为 `["/health", "/metrics", "/api/v1/public/*"]`
- **THEN** 这些路径无需认证

---

### Requirement: 认证失败响应
系统 SHALL 在认证失败时返回标准的错误响应。

#### Scenario: 返回 401 状态码
- **WHEN** 认证失败
- **THEN** 返回 HTTP 状态码 401

#### Scenario: 返回错误详情
- **WHEN** 认证失败
- **THEN** 返回 JSON 错误响应 `{"error": "Unauthorized", "code": "AUTH_FAILED", "message": "Token 无效或已过期"}`

---

### Requirement: 认证日志记录
系统 SHALL 记录认证失败的审计日志。

#### Scenario: 记录认证失败
- **WHEN** 客户端认证失败
- **THEN** 系统记录审计日志，包含 client_ip、失败原因、时间戳

---

### Requirement: 性能要求
认证中间件 SHALL 具有高性能，延迟低于 1ms。

#### Scenario: 中间件延迟
- **WHEN** 认证中间件验证 Token
- **THEN** 处理时间小于 1ms
