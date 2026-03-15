# Capability: WebUI Interface

## Purpose

提供基于 Web 的图形化管理界面，包括登录、实例管理、SQL 查询、审计日志和权限管理等功能。

## Requirements

### Requirement: WebUI 登录页面
系统 SHALL 提供 WebUI 登录页面。

#### Scenario: Token 登录
- **WHEN** 用户输入用户名和 Token
- **THEN** 系统验证并建立会话

#### Scenario: OIDC 登录
- **WHEN** 用户点击 OIDC 登录按钮
- **THEN** 重定向到 OIDC 提供者

#### Scenario: LDAP 登录
- **WHEN** 用户输入 LDAP 用户名和密码
- **THEN** 系统验证 LDAP 凭据

### Requirement: 实例列表页面
系统 SHALL 提供数据库实例列表页面。

#### Scenario: 显示实例列表
- **WHEN** 访问实例页面
- **THEN** 显示用户有权限的所有实例

#### Scenario: 显示实例状态
- **WHEN** 实例列表加载
- **THEN** 显示每个实例的健康状态

#### Scenario: 按集群分组显示
- **WHEN** 配置了多集群
- **THEN** 按集群分组显示实例

### Requirement: SQL 查询页面
系统 SHALL 提供 SQL 查询界面。

#### Scenario: 选择实例
- **WHEN** 用户选择实例
- **THEN** 页面切换到该实例

#### Scenario: 执行 SQL
- **WHEN** 用户输入 SQL 并点击执行
- **THEN** 系统执行并显示结果

### Requirement: 审计日志页面
系统 SHALL 提供审计日志查询页面。

#### Scenario: 查询审计日志
- **WHEN** 用户输入查询条件
- **THEN** 显示匹配的审计日志

### Requirement: 权限管理页面
系统 SHALL 提供权限管理页面（仅管理员）。

#### Scenario: 管理角色
- **WHEN** 管理员访问角色管理页面
- **THEN** 可以创建、编辑、删除角色

#### Scenario: 管理用户权限
- **WHEN** 管理员访问用户权限页面
- **THEN** 可以为用户分配角色
