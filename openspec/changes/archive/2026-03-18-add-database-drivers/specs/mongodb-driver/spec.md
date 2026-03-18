## Overview

MongoDB 是面向文档的 NoSQL 数据库。本文档定义 MystiSql 对 MongoDB 的连接和操作支持。

## Connection

### Requirement: 支持标准 MongoDB 连接

MongoDB 连接必须支持:
- 直接连接 (单机/副本集)
- 连接字符串 (URI 格式)
- 认证: 无认证 / SCRAM / x.509
- 连接池配置
- TLS/SSL 支持

#### Scenario: 连接到 MongoDB 副本集
- **Given** 一个配置了 URI 或 host/port 的 MongoDB 实例
- **When** 系统建立连接
- **then** 连接成功建立并通过 Ping 验证

### Requirement: 支持 MongoDB 连接字符串

```
mongodb://user:password@host:port/database?options
```

参数说明:
- `replicaSet`: 副本集名称
- `authSource`: 认证源
- `tls`: 启用 TLS
- `maxPoolSize`: 最大连接池大小

## Query Operations

### Requirement: 支持文档查询

支持基本的 CRUD 操作
- find: 查询单个文档
- findMany: 查询多个文档
- insertOne: 插入单个文档
- insertMany: 批量插入
- updateOne: 更新单个文档
- updateMany: 批量更新
- deleteOne: 删除单个文档
- deleteMany: 批量删除

#### Scenario: 查询文档
- **Given** 一个 find 查询 `{ "collection": "users", "filter": { "age": { "$gt": 18 } } }`
- **When** 查询执行
- **then** 返回匹配的文档列表

### Requirement: 支持聚合管道

支持 MongoDB 聚合管道操作
- $match: 过滤文档
- $group: 分组聚合
- $sort: 排序
- $limit/$skip: 分页
- $project: 字段投影
- $lookup: 关联查询

#### Scenario: 聚合查询
- **Given** 一个聚合管道查询
- **When** 查询执行
- **then** 返回聚合结果

## Special Operations

### Requirement: 支持索引操作

- createIndex: 创建索引
- dropIndex: 删除索引
- listIndexes: 列出索引
- ensureIndex: 确保索引存在

### Requirement: 支持集合操作

- createCollection: 创建集合
- dropCollection: 删除集合
- listCollections: 列出集合

## Implementation Notes
- 使用 `go.mongodb.org/mongo-driver/v2` 驱动
- 实现 `Connection` 接口 (适配 NoSQL 语义)
- 使用驱动内置连接池
- 支持 BSON 类型转换
