## Context

MystiSql 当前实现了 MySQL 和 PostgreSQL 两种数据库驱动，采用 `ConnectionFactory` 工厂模式创建连接。现有架构：

```
internal/connection/
├── types.go           # Connection, ConnectionFactory, ConnectionPool 接口定义
├── pool/              # 连接池通用实现
├── mysql/             # MySQL 驱动实现
└── postgresql/        # PostgreSQL 驱动实现
```

`Connection` 接口定义了5个方法：`Connect`, `Query`, `Exec`, `Ping`, `Close`，所有驱动必须实现此接口。

## Goals / Non-Goals

**Goals:**
- 添加 8 种新数据库驱动：Redis, SQLite, MSSQL, Oracle, MongoDB, Elasticsearch, ClickHouse, etcd
- 统一驱动注册机制，支持动态驱动发现
- 保持与现有 `Connection` 接口的兼容性
- 支持各数据库的特有功能（如 Redis 的管道操作
 MongoDB 的聚合管道
 ClickHouse 的列式查询
 etcd 的 Watch 机制）
- 更新 `DatabaseType` 类型定义

**Non-Goals:**
- 不修改现有 MySQL/PostgreSQL 驱动实现
- 不重构连接池核心逻辑
- 不支持数据库特有的管理操作（如创建数据库、用户管理）

## Decisions

### 1. 驱动注册机制

**决策**: 采用显式注册模式，在 `internal/connection/registry.go` 中集中管理驱动工厂。

**理由**:
- 简单直接，易于维护
- 避免反射带来的性能开销
- 编译期检查驱动实现

```go
var drivers = map[types.DatabaseType]connection.ConnectionFactory{
    types.DatabaseTypeMySQL:      mysql.NewFactory(),
    types.DatabaseTypePostgreSQL: postgresql.NewFactory(),
    types.DatabaseTypeRedis:      redis.NewFactory(),
    // ...
}
```

### 2. NoSQL 数据库接口适配

**决策**: Redis/MongoDB/Elasticsearch 实现 `Connection` 接口，但 `Query` 方法接受特定语法的查询字符串。

**理由**:
- 保持接口统一性
- JDBC 驱动需要标准接口
- 用户可通过统一入口访问所有数据库

| 数据库 | Query 方法接受的语法 |
|--------|---------------------|
| Redis | Redis 命令 (GET, SET, HGET...) |
| MongoDB | MongoDB 聚合管道 JSON |
| Elasticsearch | Query DSL JSON |
| etcd | KV 操作 (GET, PUT, DELETE, PREFIX) |

### 3. 纯 Go 驱动优先

**决策**: 所有驱动选择纯 Go 实现，无 CGO 依赖。

| 数据库 | 驱动包 | 纯Go |
|--------|--------|------|
| Redis | github.com/redis/go-redis/v9 | ✅ |
| SQLite | modernc.org/sqlite | ✅ |
| MSSQL | github.com/microsoft/go-mssqldb | ✅ |
| Oracle | github.com/sijms/go-ora/v2 | ✅ |
| MongoDB | go.mongodb.org/mongo-driver/v2 | ✅ |
| Elasticsearch | github.com/elastic/go-elasticsearch/v8 | ✅ |
| ClickHouse | github.com/ClickHouse/clickhouse-go/v2 | ✅ |
| etcd | go.etcd.io/etcd/client/v3 | ✅ |

### 4. 类型定义扩展

**决策**: 扩展 `pkg/types/instance.go` 中的 `DatabaseType` 常量。

```go
const (
    DatabaseTypeMySQL       DatabaseType = "mysql"
    DatabaseTypePostgreSQL  DatabaseType = "postgresql"
    DatabaseTypeOracle      DatabaseType = "oracle"
    DatabaseTypeRedis       DatabaseType = "redis"
    DatabaseTypeSQLite      DatabaseType = "sqlite"
    DatabaseTypeMSSQL       DatabaseType = "mssql"
    DatabaseTypeMongoDB     DatabaseType = "mongodb"
    DatabaseTypeElasticsearch DatabaseType = "elasticsearch"
)
```

### 5. 目录结构

```
internal/connection/
├── registry.go           # 驱动注册中心（新增）
├── redis/
│   ├── connection.go
│   └── connection_test.go
├── sqlite/
│   ├── connection.go
│   └── connection_test.go
├── mssql/
│   ├── connection.go
│   └── connection_test.go
├── oracle/
│   ├── connection.go
│   └── connection_test.go
├── mongodb/
│   ├── connection.go
│   └── connection_test.go
└── elasticsearch/
    ├── connection.go
    └── connection_test.go
└── clickhouse/
    ├── connection.go
    └── connection_test.go
└── etcd/
    ├── connection.go
    └── connection_test.go
```

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|---------|
| NoSQL 查询语法差异大 | 文档明确各数据库支持的查询语法 |
| Oracle 驱动测试需要实际数据库 | 提供 Docker Compose 测试环境 |
| MongoDB/Elasticsearch/ClickHouse/etcd 连接池机制不同 | 使用各自驱动的内置连接池 |
| SQLite 文件路径安全性 | 限制可访问的目录范围 |
| ClickHouse 列式查询语法差异 | 文档说明 ClickHouse SQL 与标准 SQL 的差异 |
| etcd Watch 机制复杂性 | 仅支持基本 KV 操作，Watch 作为扩展功能 |

## Open Questions

1. **Redis 事务支持**: 是否需要支持 MULTI/EXEC 事务？
2. **MongoDB GridFS**: 是否需要支持大文件存储？
3. **Elasticsearch 索引管理**: 是否需要支持索引创建/删除操作？
