# Token 认证规范

## Purpose

定义 MystiSql 的 Token 认证功能，使用 JWT (JSON Web Token) 实现用户身份验证和授权，支持 Token 的生成、验证、撤销等完整生命周期管理。

## Requirements

### Requirement: JWT Token 生成
系统 SHALL 使用 JWT (JSON Web Token) 生成认证令牌，支持用户身份验证。

#### Scenario: 成功生成 Token
- **WHEN** 管理员调用 Token 生成接口并提供用户信息
- **THEN** 系统返回有效的 JWT Token，包含用户 ID、角色和过期时间

#### Scenario: Token 包含必要信息
- **WHEN** 生成 JWT Token
- **THEN** Token payload 包含 `user_id`、`role`、`exp`（过期时间）字段

---

### Requirement: Token 签名和验证
系统 SHALL 使用 HS256 算法对 Token 进行签名和验证。

#### Scenario: 有效 Token 验证通过
- **WHEN** 客户端提供有效的 JWT Token
- **THEN** 系统验证签名成功并提取用户信息

#### Scenario: 无效 Token 验证失败
- **WHEN** 客户端提供无效或被篡改的 Token
- **THEN** 系统返回 401 Unauthorized 错误

#### Scenario: 过期 Token 验证失败
- **WHEN** 客户端提供已过期的 Token
- **THEN** 系统返回 401 Unauthorized 并提示 Token 已过期

---

### Requirement: Token 配置管理
系统 SHALL 支持通过配置文件管理 Token 相关参数。

#### Scenario: 配置签名密钥
- **WHEN** 配置文件指定 `auth.token.secret`
- **THEN** 系统使用该密钥对 Token 进行签名和验证

#### Scenario: 配置过期时间
- **WHEN** 配置文件指定 `auth.token.expire`（如 "24h"）
- **THEN** 生成的 Token 在指定时间后过期

#### Scenario: 默认过期时间
- **WHEN** 未配置 `auth.token.expire`
- **THEN** Token 默认过期时间为 24 小时

---

### Requirement: Token 在 API 中的传递
系统 SHALL 支持两种 Token 传递方式：HTTP Header 和 URL 参数。

#### Scenario: 通过 Authorization Header 传递
- **WHEN** 客户端在 HTTP Header 中设置 `Authorization: Bearer <token>`
- **THEN** 系统从 Header 中提取并验证 Token

#### Scenario: 通过 URL 参数传递
- **WHEN** 客户端在 URL 中添加 `?token=<token>` 参数
- **THEN** 系统从 URL 参数中提取并验证 Token

#### Scenario: 优先使用 Header Token
- **WHEN** 请求同时包含 Header 和 URL 参数中的 Token
- **THEN** 系统优先使用 Header 中的 Token

---

### Requirement: Token 撤销机制
系统 SHALL 支持通过黑名单机制撤销 Token。

#### Scenario: Token 加入黑名单
- **WHEN** 管理员调用 Token 撤销接口
- **THEN** 该 Token 被加入黑名单，后续请求中被拒绝

#### Scenario: 黑名单 Token 验证失败
- **WHEN** 客户端提供已被撤销的 Token
- **THEN** 系统返回 401 Unauthorized 并提示 Token 已被撤销

---

### Requirement: Token 生成接口
系统 SHALL 提供 RESTful API 接口生成 Token。

#### Scenario: 管理员生成 Token
- **WHEN** POST `/api/v1/auth/token` 请求包含用户信息（user_id, role）
- **THEN** 系统返回 JWT Token

#### Scenario: 未授权用户生成 Token 失败
- **WHEN** 未授权用户尝试调用 Token 生成接口
- **THEN** 系统返回 403 Forbidden

---

### Requirement: Token 列表查询
系统 SHALL 提供接口查询当前有效的 Token 列表。

#### Scenario: 查询有效 Token
- **WHEN** 管理员调用 GET `/api/v1/auth/tokens`
- **THEN** 系统返回所有未过期且未被撤销的 Token 列表
