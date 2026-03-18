## Overview

Redis 是高性能键值存储数据库，支持字符串、哈希、列表、集合、有序集合等数据结构。本文档定义 MystiSql 对 Redis 的连接和操作支持。

## Connection

### Requirement: 支持标准 Redis 连接

Redis 连接必须支持：
- 单机模式
- 基于URL 的连接
- 连接池配置
- 密码认证（如果有)

#### Scenario: 连接到单机 Redis
- **Given** 一个配置了 host、 port、 password 的 Redis 实例
- **When** 系统建立连接
- **then** 连接应成功建立并通过 Ping 验证

### Requirement: 支持 Redis URL 格式

```
redis://user:password@host:port/db
rediss://user:password@host:port/db
```

## Query Operations

### Requirement: 支持基本 Redis 埥询

支持以下查询类型:
- GET: 获取单个值
- MGET: 获取多个值
- HGETALL: 获取哈希所有字段
- LRANGE: 获取列表范围
- TYPE: 获取键类型
- KEYS: 获取匹配模式的键

#### Scenario: 获取缓存值
- **Given** 一个 GET 埥询 `GET user:123`
- **When** 查询执行
- **then** 返回对应 key 的值或 null

### Requirement: 支持 Redis 特有命令

- PING: 壀康检查
- INFO: 服务器信息
- TTL: 设置过期时间
- EXISTS: 检查键是否存在

## Write Operations

### Requirement: 支持基本 Redis 写入

- SET: 设置键值
- MSET: 设置多个键值
- HSET: 设置哈希字段
- HMSET: 设置哈希多个字段
- LPUSH/LPOP: 列表操作
- SADD/SREM: 集合操作
- ZADD/ZREM: 有序集合操作

- SETEX: 设置带过期时间的值

#### Scenario: 设置缓存值
- **Given** 一个 SET 查询 `SET user:123 "Alice"`
- **When** 命令执行
- **then** 值被存储并返回成功

## Pipeline Operations

### Requirement: 支持 Redis Pipeline 批量操作

支持通过 Pipeline 批量执行多个命令，减少网络往返。

#### Scenario: 批量设置值
- **Given** 多个 SET 命令
- **When** 使用 Pipeline 执行
- **then** 所有命令在一个请求中发送

## Implementation Notes

- 使用 `github.com/redis/go-redis/v9` 驱动
- 实现 `Connection` 接口
- 连接池使用驱动内置实现
- 支持 context 超时控制
