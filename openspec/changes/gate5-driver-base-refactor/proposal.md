## Why

5 个 database/sql 驱动（MySQL、PostgreSQL、SQLite、MSSQL、Oracle）的 Query/Exec/Ping/Close 实现约 95% 相同代码，总计约 400+ 行重复。每次修改通用逻辑（如列类型映射、错误处理、连接池参数）需同步修改 5 处。但三者 API 异构：MySQL/SQLite/MSSQL 用 `database/sql` 标准接口，PostgreSQL 用 `pgxpool`（完全不同的 API），Oracle 用 `go-ora`。公共基类只能覆盖 `database/sql` 系的 4 个驱动。

## What Changes

### 提取 connection/base 包
- **`base.SQLConnection` struct**：封装 `*sql.DB`，实现通用的 Query/Exec/Ping/Close 方法
- **列类型映射统一**：`mapColumnType(*sql.ColumnType) string` 统一处理，各驱动仅 override 特殊类型
- **连接池配置统一**：从配置读取 MaxOpen/MaxIdle/MaxLifetime，一处设置全局生效

### 各驱动重构
- **MySQL/SQLite/MSSQL/Oracle**：嵌入 `base.SQLConnection`，仅保留 `Connect()`（DSN 构建）和 `Factory` 方法
- **PostgreSQL 不参与**：`pgxpool` API 与 `database/sql` 完全不同（`pool.Query` vs `db.Query`），强行统一反而增加复杂度。PostgreSQL 保持独立实现
- **Redis 不参与**：非 SQL 驱动，命令式 API

### 逐个迁移策略
- 一次迁移一个驱动，每个迁移后运行该驱动的已有测试确认无破坏
- 顺序：MySQL（最成熟，已有测试）→ SQLite → MSSQL → Oracle（无测试，风险最高）

## Capabilities

### Modified Capabilities
- `mysql-connection`: 嵌入 base.SQLConnection
- `postgresql-driver`: 不变（API 异构）
- `directory-structure`: 新增 `internal/connection/base/` 包

## Impact

### 受影响的代码
- `internal/connection/base/` — 新增（~150 行公共实现）
- `internal/connection/mysql/connection.go` — 重构
- `internal/connection/sqlite/connection.go` — 重构
- `internal/connection/mssql/connection.go` — 重构
- `internal/connection/oracle/connection.go` — 重构
- `internal/connection/postgresql/` — 不变

### 前置条件
- Gate 0 完成（可编译可测试）

### Done 标准
- 4 个 database/sql 驱动通过 base 包工作
- 现有 MySQL/PG 测试全部通过
- 通用逻辑修改只需改 base 一处

### 信心
**55%** — 每个驱动的 DSN 构建和列类型映射有微妙差异，迁移时容易遗漏。Oracle 无测试，回归风险最高。逐个迁移 + 每步验证可降低风险
