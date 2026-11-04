# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2026-11-03

### Fixed

- **CLI auth**: `auth revoke` 改为在 DELETE body 发送 `{"token": "xxx"}`，`auth info` 使用 `?token=` 参数，与服务端 handler 对齐（P1-005/P1-006）
- **JDBC**: PreparedStatement.executeUpdate 后 getUpdateCount()/getGeneratedKeys() 不再返回默认值（P1-004）
- **Transaction**: 隔离级别 SQL 按数据库类型分派：PostgreSQL 用 SET SESSION CHARACTERISTICS、Oracle 用 ALTER SESSION（仅 RC/SERIALIZABLE，其余报错）（P1-003）
- **Batch**: 执行超时 config.Timeout 生效；并行模式支持 StopOnError 取消（P2-005/P2-006）
- **CLI serve**: 补注册 ClickHouse/MSSQL/Elasticsearch/Etcd 四个驱动（P2-013）
- **Elasticsearch**: Close 释放底层 HTTP transport（P3-008）
- **etcd**: Ping 改用原生维护 API，无数据副作用（P3-009）
- **Auth**: generateTokenID 显式处理 crypto/rand 错误（P3-002）
- **CI**: 修复 golangci-lint 全部报错（main 分支 CI 失败根因），.golangci.yml v2 语法排除 web/node_modules

### Added

- **HA API**: 11 个 HA 端点接入 setupRoutes（实例健康/拓扑/故障转移）（P2-001）
- **Validator**: `POST /api/v1/validator/validate` 独立 SQL 验证端点（P2-002）
- **Middleware**: 滑动窗口速率限制中间件（默认 60 req/min，429 RATE_LIMITED）（P2-010）
- **TLS**: ServerConfig 支持 certFile/keyFile，齐备时 HTTPS 启动（P3-003）
- **Audit**: 日志级别过滤 audit.logLevel（all/sensitive/write/ddl）；批量执行接入审计日志（P2-005/P2-007）
- **Auth**: JWT 可选 issuer/audience claims 且验证强制校验（P3-007）
- **Query**: 连接池参数支持外部配置 Engine.SetPoolConfig（P3-010）
- **Discovery**: K8sDynamicDiscoverer 实现 InstanceDiscoverer 接口（P2-011）
- **Query Parser**: Validate 拒绝包含 SQL 注释的语句（P2-008）
- **CSP**: 安全响应头新增 Content-Security-Policy（P3-004）

### Security

- 审计日志文件权限默认 0644 → 0600 且可配置（P2-017）

### Changed

- **WebUI**: 仪表盘在线时长展示真实服务运行时间（/health uptime）（P2-009）
- **ClickHouse**: 驱动重构嵌入 base.SQLConnection，169→48 行（P2-012）
- **WebUI**: eslint 0 error 0 warning（修复 react-hooks/exhaustive-deps 与 ESM 插件兼容）

### Tests

- Redis 驱动补充 miniredis 连接层测试 8 个（P1-001）

## [0.3.0] - 2026-03-10

### Added

#### Security Layer (Phase 3)

**Authentication**
- JWT Token based authentication with HS256 signature
- Token generation API: `POST /api/v1/auth/token`
- Token revocation API: `DELETE /api/v1/auth/token`
- Token info API: `GET /api/v1/auth/token/info`
- Token list API: `GET /api/v1/auth/tokens`
- CLI auth commands: `mystisql auth token`, `mystisql auth revoke`, `mystisql auth info`
- Configurable token expiration (default: 24 hours)
- Token blacklist for revocation support

**Audit Logging**
- Comprehensive SQL execution audit logging
- JSON Lines format for easy log processing
- Log rotation (daily rotation, 30-day retention)
- Async log writing to avoid request blocking
- Audit log query API: `GET /api/v1/audit/logs`
- Sensitive operation marking (DDL, DELETE without WHERE)

**SQL Validation**
- Dangerous operation detection (DROP, TRUNCATE, DELETE/UPDATE without WHERE)
- SQL parser-based validation (not regex)
- Configurable dangerous operation list
- Validation on/off switch
- Clear error messages for blocked operations

**SQL Whitelist/Blacklist**
- SQL whitelist for allowing specific patterns
- SQL blacklist for blocking specific patterns
- Regex pattern support
- Hot reload configuration support
- Priority: blacklist > whitelist > dangerous operation check
- APIs: `PUT /api/v1/validator/whitelist`, `PUT /api/v1/validator/blacklist`

**API Authentication Middleware**
- Global authentication middleware for all API endpoints
- Whitelist path support (e.g., `/health`)
- Token extraction from Authorization header or URL parameter
- User info injection into gin.Context
- < 1ms middleware latency

**CLI Authentication**
- Token configuration via config file, environment variable, or CLI flag
- Priority: `--token` > `MYSTISQL_TOKEN` env > config file
- Authentication for all CLI commands requiring API access
- Clear error messages for auth failures

#### PostgreSQL Support

- PostgreSQL connection using pgx driver
- Connection pool management (reuse MySQL interface)
- SSL mode configuration
- Connection timeout configuration
- PostgreSQL-specific error handling
- Dynamic driver selection based on instance type

#### WebSocket Support

- WebSocket endpoint: `ws://host:port/ws`
- Token authentication via URL parameter
- JSON message format
- Query execution via WebSocket
- Connection management (max connections, idle timeout)
- Heartbeat mechanism (ping/pong)
- Concurrent query limit per connection

#### Transaction Management (JDBC Compatible)

- Transaction begin API: `POST /api/v1/transaction/begin`
- Transaction commit API: `POST /api/v1/transaction/commit`
- Transaction rollback API: `POST /api/v1/transaction/rollback`
- Transaction status API: `GET /api/v1/transaction/{id}`
- Transaction list API: `GET /api/v1/transaction`
- Transaction extend API: `POST /api/v1/transaction/{id}/extend`
- Isolation level configuration
- Automatic timeout rollback (default: 5 minutes)
- Connection ID binding to client

#### Batch Operations

- Batch SQL execution API: `POST /api/v1/batch`
- Support for INSERT, UPDATE, DELETE operations
- Mixed batch support
- Batch size limit (default: 1000)
- Partial success response with detailed results
- Transaction support for batch operations
- Performance optimization using native batch processing

### Changed

- Updated API documentation with all Phase 3 endpoints
- Updated README.md with Phase 3 features
- Updated AGENTS.md with Phase 3 development guidelines

### Fixed

- Various bug fixes and stability improvements

## [0.2.0] - Previous Release

- Core database connectivity
- MySQL support
- Basic REST API
- CLI interface
- Connection pooling

## [0.1.0] - Initial Release

- Project initialization
- Basic structure
