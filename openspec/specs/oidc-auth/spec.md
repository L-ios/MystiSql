# Capability: OIDC Auth

## Purpose

支持通过 OpenID Connect (OIDC) 协议进行身份认证，兼容 Keycloak、Dex 等主流 OIDC 提供者，实现单点登录和统一身份管理。

## Requirements

### Requirement: OIDC 提供者配置
系统 SHALL 支持配置 OIDC 提供者（如 Keycloak、Dex）。

#### Scenario: 配置 OIDC 提供者
- **WHEN** 配置 `auth.oidc.issuerUrl` 和 `auth.oidc.clientId`
- **THEN** 系统自动发现 OIDC 配置

#### Scenario: 支持多个 OIDC 提供者
- **WHEN** 配置多个 OIDC 提供者
- **THEN** 用户可以选择任意一个登录

### Requirement: OIDC 登录流程
系统 SHALL 支持 Authorization Code Flow 登录。

#### Scenario: 发起 OIDC 登录
- **WHEN** 用户访问 `/api/v1/auth/oidc/login`
- **THEN** 系统返回 OIDC 授权 URL

#### Scenario: OIDC 回调处理
- **WHEN** OIDC 提供者回调并携带 code
- **THEN** 系统交换 code 获取 Token 并创建会话

### Requirement: Token 验证
系统 SHALL 验证 OIDC ID Token。

#### Scenario: 验证 ID Token 签名
- **WHEN** 收到 ID Token
- **THEN** 系统使用 OIDC 提供者的公钥验证签名

#### Scenario: 验证 Token 过期时间
- **WHEN** Token 已过期
- **THEN** 系统拒绝请求并返回 401

### Requirement: 用户信息提取
系统 SHALL 从 ID Token 或 UserInfo 端点提取用户信息。

#### Scenario: 从 Token 提取用户信息
- **WHEN** ID Token 包含 `preferred_username` 和 `email`
- **THEN** 系统提取作为用户名和邮箱

#### Scenario: 从 UserInfo 端点获取信息
- **WHEN** 配置 `auth.oidc.userInfoEnabled = true`
- **THEN** 系统调用 UserInfo 端点获取完整用户信息

### Requirement: OIDC 角色映射
系统 SHALL 将 OIDC 角色/组映射到 MystiSql 权限。

#### Scenario: 从 Token 角色映射
- **WHEN** ID Token 包含 `roles` 或 `groups` 声明
- **THEN** 系统根据映射规则转换为 MystiSql 角色
