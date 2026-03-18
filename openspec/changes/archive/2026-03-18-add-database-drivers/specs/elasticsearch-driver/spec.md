## Overview

Elasticsearch 是分布式搜索和分析引擎。本文档定义 MystiSql 对 Elasticsearch 的连接和操作支持。

## Connection

### Requirement: 支持标准 Elasticsearch 连接

Elasticsearch 连接必须支持:
- 单节点连接
- 集群连接 (多个节点)
- 云服务 (Elastic Cloud)
- API Key 认证
- 基本认证 (用户名/密码)
- TLS 加密

#### Scenario: 连接到 Elasticsearch 集群
- **Given** 一个配置了 addresses 的 Elasticsearch 实例
- **When** 系统建立连接
- **then** 连接成功建立并通过信息 API 验证

### Requirement: 支持 Elasticsearch 连接配置

```
addresses: ["http://localhost:9200"]
username: "elastic"
password: "changeme"
api_key: "your-api-key"
```

## Query Operations

### Requirement: 支持搜索查询

支持基本搜索操作
- Search API: 基本搜索
- Multi-search API: 批量搜索
- Count API: 计数查询
- Exists API: 检查文档存在

#### Scenario: 执行搜索查询
- **Given** 一个搜索请求 `{ "index": "logs", "query": { "match": { "level": "ERROR" } } }`
- **When** 查询执行
- **then** 返回匹配的文档列表

### Requirement: 支持 DSL 查询

支持完整的 Elasticsearch DSL
- bool 查询
- match 查询
- term 查询
- range 查询
- aggregation 聚合
- script 脚本查询

#### Scenario: DSL 聚合查询
- **Given** 一个聚合查询
- **When** 查询执行
- **then** 返回聚合结果

## Index Operations

### Requirement: 支持索引管理

- Create Index: 创建索引
- Delete Index: 删除索引
- Get Index: 获取索引信息
- Update Index Settings: 更新索引设置
- Reindex: 重建索引

### Requirement: 支持映射管理

- Put Mapping: 添加映射
- Get Mapping: 获取映射
- Update Mapping: 更新映射

## Document Operations

### Requirement: 支持文档 CRUD

- Index: 索引文档 (创建/更新)
- Get: 获取文档
- Update: 更新文档
- Delete: 删除文档
- Bulk: 批量操作

#### Scenario: 索引文档
- **Given** 一个文档索引请求
- **When** 请求执行
- **then** 文档被索引并返回确认

## Implementation Notes
- 使用 `github.com/elastic/go-elasticsearch/v8` 驱动
- 实现 `Connection` 接口 (适配搜索引擎语义)
- 使用驱动内置连接池
- 支持 JSON 查询 DSL
