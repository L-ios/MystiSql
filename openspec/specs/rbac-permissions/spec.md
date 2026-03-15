# Capability: RBAC Permissions

## Purpose

提供基于角色的访问控制（RBAC）能力，支持细粒度的权限管理（库级别、表级别、操作级别），实现企业级的数据访问权限管控。

## Requirements

### Requirement: RBAC 角色定义
系统 SHALL 支持定义角色并分配权限。

#### Scenario: 创建角色
- **WHEN** 管理员创建角色 `data_analyst`
- **THEN** 系统保存角色定义

#### Scenario: 为角色分配权限
- **WHEN** 为角色添加权限 `mysql-prod:reports:SELECT`
- **THEN** 该角色的用户获得指定权限

### Requirement: 权限粒度
系统 SHALL 支持库级别、表级别、操作级别的权限控制。

#### Scenario: 库级别权限
- **WHEN** 用户拥有 `mysql-prod:*:*` 权限
- **THEN** 用户可以访问该库的所有表和操作

#### Scenario: 表级别权限
- **WHEN** 用户拥有 `mysql-prod:users:SELECT` 权限
- **THEN** 用户只能查询 users 表

#### Scenario: 操作级别权限
- **WHEN** 用户拥有 `mysql-prod:orders:SELECT,INSERT` 权限
- **THEN** 用户只能查询和插入 orders 表

### Requirement: 角色继承
系统 SHALL 支持角色继承关系。

#### Scenario: 角色继承权限
- **WHEN** 角色 `senior_analyst` 继承 `analyst`
- **THEN** `senior_analyst` 拥有 `analyst` 的所有权限

### Requirement: 权限检查
系统 SHALL 在执行查询前检查用户权限。

#### Scenario: 有权限执行
- **WHEN** 用户有 `mysql-prod:users:SELECT` 权限
- **THEN** 查询被允许执行

#### Scenario: 无权限拒绝
- **WHEN** 用户没有 `mysql-prod:users:DELETE` 权限
- **THEN** 系统拒绝请求并返回 403

### Requirement: 权限管理 API
系统 SHALL 提供权限管理 REST API。

#### Scenario: 查询用户权限
- **WHEN** GET /api/v1/rbac/users/{userId}/permissions
- **THEN** 返回用户的所有权限列表

#### Scenario: 分配角色给用户
- **WHEN** POST /api/v1/rbac/users/{userId}/roles
- **THEN** 用户获得指定角色
