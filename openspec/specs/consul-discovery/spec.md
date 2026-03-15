# Capability: Consul Discovery

## Purpose

支持通过 HashiCorp Consul 进行数据库实例服务发现，实现动态实例注册与健康状态同步。

## Requirements

### Requirement: Consul 服务发现配置
系统 SHALL 支持通过配置启用 Consul 服务发现，包括地址、命名空间、Token 等参数。

#### Scenario: 启用 Consul 发现
- **WHEN** 配置 `discovery.consul.enabled = true` 并提供 Consul 地址
- **THEN** 系统连接到 Consul 并开始发现数据库实例

#### Scenario: Consul ACL Token 认证
- **WHEN** 配置 `discovery.consul.token`
- **THEN** 系统使用 Token 进行 Consul API 认证

### Requirement: 从 Consul 发现数据库实例
系统 SHALL 从 Consul Catalog 或 KV 存储中发现数据库实例信息。

#### Scenario: 从 Consul Service 发现实例
- **WHEN** Consul 中注册了 tag 为 `database=mysql` 的服务
- **THEN** 系统自动发现并注册该数据库实例

#### Scenario: 从 Consul KV 发现实例
- **WHEN** Consul KV 存储中有 `mystisql/instances/` 前缀的配置
- **THEN** 系统解析 KV 数据并注册实例

### Requirement: Consul 健康检查集成
系统 SHALL 读取 Consul 的健康检查状态来确定实例可用性。

#### Scenario: 实例健康状态同步
- **WHEN** Consul 中服务健康检查状态变更
- **THEN** 系统更新实例状态（healthy/unhealthy）

### Requirement: Consul 服务标签解析
系统 SHALL 解析 Consul 服务标签来识别数据库类型和角色。

#### Scenario: 解析数据库类型标签
- **WHEN** 服务标签包含 `mysql`、`postgresql`、`oracle`
- **THEN** 系统正确设置实例的数据库类型

#### Scenario: 解析主从角色标签
- **WHEN** 服务标签包含 `master` 或 `slave`
- **THEN** 系统设置实例的 role 标签
