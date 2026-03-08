## Why

随着 MystiSql Phase 1 和 Phase 2 的完成，系统已具备基本的数据库连接、查询执行和 JDBC 驱动功能。然而，在生产环境中使用仍缺乏必要的安全控制能力：无身份认证导致任何人都可以访问数据库、无审计日志无法追溯操作记录、无危险操作拦截可能导致数据误删除。同时，需要扩展支持 PostgreSQL 数据库并增强 JDBC 驱动功能，以满足更广泛的使用场景。

Phase 3 的目标是实现企业级安全控制能力，让 MystiSql 可以安全地在生产环境中部署使用。

## What Changes

### 安全控制层
- **Token 认证机制**：实现基于 Token 的身份认证，支持 API 和 CLI 认证
- **审计日志**：记录所有 SQL 执行操作，包括操作者、SQL 语句、执行时间、影响行数
- **SQL 安全检查**：检测并拦截危险操作（DROP、TRUNCATE、DELETE 全表等）
- **SQL 白名单/黑名单**：支持配置允许或禁止执行的 SQL 模式

### 接入层增强
- **WebSocket 支持**：提供实时交互能力，支持长连接
- **CLI 认证集成**：CLI 命令支持 Token 认证
- **API 认证中间件**：为所有 API 端点添加认证中间件

### 数据库支持扩展
- **PostgreSQL 驱动**：使用 pgx 驱动支持 PostgreSQL 数据库
- **多数据库类型路由**：根据实例类型自动选择对应的数据库驱动

### JDBC 驱动增强
- **连接池支持**：兼容 HikariCP 等主流连接池
- **事务管理**：支持基础的事务操作（begin、commit、rollback）
- **批量操作**：支持批量 insert/update/delete
- **认证集成**：JDBC 连接支持 Token 传递

## Capabilities

### New Capabilities

- `token-auth`: Token 认证机制，支持 API 和 CLI 认证
- `audit-logging`: SQL 审计日志，记录所有执行操作
- `sql-validator`: SQL 安全检查，拦截危险操作
- `sql-whitelist-blacklist`: SQL 白名单/黑名单机制
- `websocket-support`: WebSocket 实时交互支持
- `cli-auth`: CLI 命令认证集成
- `api-auth-middleware`: REST API 认证中间件
- `postgresql-driver`: PostgreSQL 数据库驱动支持
- `jdbc-transaction`: JDBC 驱动事务管理
- `jdbc-batch-operations`: JDBC 驱动批量操作

### Modified Capabilities

- `rest-api`: 为所有 API 端点添加认证要求（除健康检查外）
- `mysql-connection`: 连接层增加 PostgreSQL 支持，升级为多数据库连接管理
- `cli-interface`: CLI 命令增加认证参数和 token 管理

## Impact

### 代码影响
- `internal/service/auth/` - 新增认证服务模块
- `internal/service/audit/` - 新增审计日志模块
- `internal/service/validator/` - 新增 SQL 验证器
- `internal/connection/postgresql/` - 新增 PostgreSQL 连接实现
- `internal/api/middleware/` - 新增认证中间件
- `internal/api/websocket/` - 新增 WebSocket 处理
- `pkg/jdbc/` - JDBC 驱动增强（事务、批量操作）
- `internal/cli/` - CLI 增加 token 管理和认证参数

### API 影响
- 所有 `/api/v1/*` 端点需要认证（除了 `/health`）
- 新增 `/api/v1/auth/token` - Token 管理接口
- 新增 `/ws` - WebSocket 端点
- 新增 `/api/v1/audit/logs` - 审计日志查询接口

### 配置影响
- 新增 `auth.token.secret` - Token 签名密钥
- 新增 `auth.token.expire` - Token 过期时间
- 新增 `audit.enabled` - 审计日志开关
- 新增 `audit.logFile` - 审计日志文件路径
- 新增 `validator.dangerousOperations` - 危险操作配置
- 新增 `validator.whitelist` - SQL 白名单
- 新增 `validator.blacklist` - SQL 黑名单

### 依赖影响
- 新增 `github.com/golang-jwt/jwt/v5` - JWT Token 处理
- 新增 `github.com/jackc/pgx/v5` - PostgreSQL 驱动
- 新增 `github.com/gorilla/websocket` - WebSocket 支持
