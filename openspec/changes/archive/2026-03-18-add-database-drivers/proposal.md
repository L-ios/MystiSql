## Why

MystiSql 当前仅支持 MySQL 和 PostgreSQL 两种数据库驱动，无法满足企业多样化的数据库访问需求。用户在 K8s 集群中可能同时使用 Redis 缓存、SQLite 嵌入式数据库、MSSQL/Oracle 企业数据库、MongoDB 文档数据库以及 Elasticsearch 搜索引擎。

扩展数据库驱动支持可以：
1. **统一访问入口**：通过单一网关访问所有类型的数据库
2. **简化运维**：无需为每种数据库单独配置访问工具
3. **企业级特性复用**：新增驱动自动获得认证、审计、权限控制等安全特性

## What Changes

### 新增驱动支持

| 数据库 | 驱动包 | 纯Go | 优先级 |
|--------|--------|------|--------|
| Redis | github.com/redis/go-redis/v9 | ✅ | P0 |
| SQLite | modernc.org/sqlite | ✅ | P0 |
| MSSQL | github.com/microsoft/go-mssqldb | ✅ | P1 |
| Oracle | github.com/sijms/go-ora/v2 | ✅ | P1 |
| MongoDB | go.mongodb.org/mongo-driver/v2 | ✅ | P2 |
| Elasticsearch | github.com/elastic/go-elasticsearch/v8 | ✅ | P2 |
| ClickHouse | github.com/ClickHouse/clickhouse-go/v2 | ✅ | P2 |
| etcd | go.etcd.io/etcd/client/v3 | ✅ | P1 |

### 架构调整

1. **扩展 DatabaseType 枚举**：新增 6 种数据库类型
2. **驱动工厂注册机制**：支持动态注册数据库驱动
3. **NoSQL 查询适配**：为 Redis/MongoDB/ES 提供专用查询接口

## Capabilities

### New Capabilities

- `redis-driver`: Redis KV 存储驱动，支持 Key-Value 操作、发布订阅、管道
- `sqlite-driver`: SQLite 嵌入式数据库驱动，支持本地文件数据库
- `mssql-driver`: Microsoft SQL Server 驱动，支持 T-SQL 方言
- `oracle-driver`: Oracle 数据库驱动，支持 PL/SQL 方言
- `mongodb-driver`: MongoDB 文档数据库驱动，支持聚合管道、GridFS
- `elasticsearch-driver`: Elasticsearch 搜索引擎驱动，支持 DSL 查询
- `clickhouse-driver`: ClickHouse 列式数据库驱动，支持高性能 OLAP 查询
- `etcd-driver`: etcd 分布键值存储驱动，支持 K8s 原生组件

### Modified Capabilities

- `mysql-connection`: 扩展连接工厂以支持多驱动注册
- `postgresql-driver`: 统一驱动接口，确保一致性
