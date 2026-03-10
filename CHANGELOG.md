# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
