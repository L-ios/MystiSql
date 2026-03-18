## Overview

etcd 是分布式键值存储系统，用于共享配置和服务发现。本文档定义 MystiSql 对 etcd 的连接和操作支持。

## Connection

### Requirement: 支持标准 etcd 连接

etcd 连接必须支持:
- 单节点连接
- 集群连接（多个节点）
- TLS 认证
- 用户名/密码认证
- 连接池配置

#### Scenario: 连接到 etcd 集群
- **Given** 一个配置了 endpoints 的 etcd 实例
- **When** 系统建立连接
- **Then** 连接成功建立并通过 Get 操作验证

### Requirement: 支持 etcd 连接字符串

```
etcd://user:password@endpoint1:2379,endpoint2:2379,endpoint3:2379
```

参数说明:
- `dial_timeout`: 连接超时
- `request_timeout`: 请求超时
- `tls`: 启用 TLS
- `cert_file`: 客户端证书
- `key_file`: 客户端密钥
- `ca_file`: CA 证书

## Query Operations

### Requirement: 支持基本 KV 操作

支持以下操作:
- **GET**: 获取单个键
- **PUT**: 设置键值
- **DELETE**: 删除键
- **GET RANGE**: 获取键范围
- **DELETE RANGE**: 删除键范围

#### Scenario: 获取键值
- **Given** 一个 GET 查询 `GET /config/app`
- **When** 查询执行
- **Then** 返回对应键的值或 null

### Requirement: 支持前缀查询

- **GET PREFIX**: 获取前缀匹配的所有键

#### Scenario: 获取前缀所有键
- **Given** 一个 GET PREFIX 查询 `GET PREFIX /config/`
- **When** 查询执行
- **Then** 返回所有以 `/config/` 开头的键值对

## Transaction Support

### Requirement: 支持事务操作

etcd 支持 ACID 事务:
- **BEGIN**: 开始事务
- **PUT**: 事务内写入
- **DELETE**: 事务内删除
- **COMMIT**: 提交事务
- **ROLLBACK**: 回滚事务

#### Scenario: 事务写入
- **Given** 开始事务后执行多个 PUT 操作
- **When** 提交事务
- **Then** 所有操作原子性生效

## Lease Support

### Requirement: 支持租约

etcd 租约用于键的自动过期:
- **GRANT**: 创建租约
- **REVOKE**: 撤销租约
- **KEEPALIVE**: 续约
- **PUT WITH LEASE**: 带租约写入

#### Scenario: 创建带过期时间的键
- **Given** 创建一个 60 秒的租约
- **When** 使用该租约写入键
- **Then** 键在 60 秒后自动删除

## Configuration Examples

```yaml
instances:
  - name: "etcd-cluster"
    type: "etcd"
    endpoints:
      - "http://etcd1:2379"
      - "http://etcd2:2379"
      - "http://etcd3:2379"
    username: "root"
    password: "password"
    tls:
      enabled: true
      cert_file: "/path/to/cert.pem"
      key_file: "/path/to/key.pem"
      ca_file: "/path/to/ca.pem"
```

## Implementation Notes

- 使用 `go.etcd.io/etcd/client/v3` 驱动
- 实现 `Connection` 接口（适配 KV 存储语义）
- 使用驱动内置连接池
- 支持 TLS 双向认证
- 支持 context 超时控制
