## Overview

本文档列出添加 8 种数据库驱动的实现任务，包括 Redis、SQLite、MSSQL、Oracle、MongoDB、Elasticsearch、ClickHouse 和 etcd。

## References
- Design: design.md
- Specs: specs/**/*.md
- Proposal: proposal.md

---

## Phase 1: 基础设施 (共用组件)

### 1.1 更新 DatabaseType 类型定义
- [x] 在 `pkg/types/instance.go` 中添加新的数据库类型常量
- [x] 添加 `DatabaseTypeSQLite`, `DatabaseTypeMSSQL`, `DatabaseTypeMongoDB`, `DatabaseTypeElasticsearch`, `DatabaseTypeClickHouse`, `DatabaseTypeEtcd`
- [x] 更新 `DefaultPorts` 映射

### 1.2 实现驱动注册机制
- [x] 在 `internal/connection/` 创建 `registry.go`
- [x] 实现 `DriverRegistry` 单例
- [x] 提供 `RegisterDriver(driverType, factory)` 方法
- [x] 提供 `GetFactory(driverType)` 方法
- [x] 单元测试

## Phase 2: Redis 驱动

### 2.1 创建 Redis 连接实现
- [x] 创建 `internal/connection/redis/` 目录
- [x] 实现 `Connection` 接口
- [x] 实现 `Factory`
- [x] 支持 Redis 命令（GET, SET, DEL, KEYS, HGETALL, LRANGE, SMEMBERS, INCR, DECR）
- [x] 单元测试

## Phase 3: SQLite 驱动

### 3.1 创建 SQLite 连接实现
- [x] 创建 `internal/connection/sqlite/` 目录
- [x] 实现 `Connection` 接口
- [x] 实现 `Factory`
- [x] 支持 `file://` URL 格式
- [x] 实现标准 SQL 操作
- [ ] 单元测试

## Phase 4: MSSQL 驱动

### 4.1 创建 MSSQL 连接实现
- [x] 创建 `internal/connection/mssql/` 目录
- [x] 实现 `Connection` 接口
- [x] 实现 `Factory`
- [x] 支持 `sqlserver://` URL 格式
- [x] 实现标准 T-SQL 操作
- [ ] 单元测试

## Phase 5: Oracle 驱动

### 5.1 创建 Oracle 连接实现
- [x] 创建 `internal/connection/oracle/` 目录
- [x] 实现 `Connection` 接口
- [x] 实现 `Factory`
- [x] 支持 `oracle://` URL 格式
- [x] 实现 PL/SQL 支持
- [ ] 单元测试

## Phase 6: MongoDB 驱动

### 6.1 创建 MongoDB 连接实现
- [x] 创建 `internal/connection/mongodb/` 目录
- [x] 实现 `Connection` 接口 (适配 NoSQL 语义)
- [x] 实现 `Factory`
- [x] 支持 `mongodb://` URL 格式
- [x] 实现文档 CRUD 操作
- [ ] 单元测试

## Phase 7: Elasticsearch 驱动

### 7.1 创建 Elasticsearch 连接实现
- [x] 创建 `internal/connection/elasticsearch/` 目录
- [x] 实现 `Connection` 接口 (适配搜索引擎语义)
- [x] 实现 `Factory`
- [x] 支持 `elasticsearch://` URL 格式
- [x] 实现搜索查询
- [ ] 单元测试

## Phase 8: ClickHouse 驱动

### 8.1 创建 ClickHouse 连接实现
- [x] 创建 `internal/connection/clickhouse/` 目录
- [x] 实现 `Connection` 接口 (适配列式数据库语义)
- [x] 实现 `Factory`
- [x] 支持 `clickhouse://` URL 格式
- [x] 实现标准 SQL 查询
- [ ] 单元测试

## Phase 9: etcd 驱动

### 9.1 创建 etcd 连接实现
- [x] 创建 `internal/connection/etcd/` 目录
- [x] 实现 `Connection` 接口（适配 KV 存储语义）
- [x] 实现 `Factory`
- [x] 支持 `etcd://` URL 格式
- [x] 实现基本 KV 操作（GET/PUT/DELETE）
- [x] 实现前缀查询
- [ ] 单元测试

## Phase 10: 文档和配置更新

### 10.1 更新 README.md
- [x] 添加新数据库驱动说明
- [x] 更新支持矩阵
- [x] 添加配置示例

### 10.2 更新配置文件示例
- [x] 添加所有驱动的配置示例
