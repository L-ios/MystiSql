# MystiSql Agent Guidelines

Guidelines for AI coding agents working on the MystiSql codebase.

## Project Overview

MystiSql is a database access gateway for Kubernetes clusters, supporting MySQL, PostgreSQL, Oracle, and Redis. It provides CLI, WebUI, RESTful API, WebSocket, and JDBC driver interfaces.

**Current Stage**: Phase 3 development (Security layer - Authentication, Audit, Validation). See README.md for detailed roadmap and architecture.

## Build/Lint/Test Commands

### Building
```bash
go build -o bin/mystisql ./cmd/mystisql
GOOS=linux GOARCH=amd64 go build -o bin/mystisql-linux-amd64 ./cmd/mystisql
go run ./cmd/mystisql
go run ./cmd/mystisql --config config.yaml
```

### Testing
```bash
go test ./...                                    # Run all tests
go test -v ./...                                 # Verbose output
go test -cover ./...                             # With coverage
go test -coverprofile=coverage.out ./...         # Detailed coverage
go tool cover -html=coverage.out                 # View coverage report

# Run a single test file
go test -v ./path/to/package/package_test.go

# Run a single test function
go test -v ./path/to/package -run TestFunctionName

# Run tests matching a pattern
go test -v ./path/to/package -run "TestPattern.*"

# Run with race detection
go test -race ./...

# Run specific package
go test -v ./internal/connection/mysql
```

### Linting and Formatting
```bash
go fmt ./...                                     # Format code
go vet ./...                                     # Vet for common errors
golangci-lint run                                # Run linter
golangci-lint run ./path/to/package/...          # Lint specific package
golangci-lint run --fix                          # Auto-fix issues
```

### Dependencies
```bash
go mod tidy                                      # Tidy dependencies
go mod verify                                    # Verify dependencies
go get -u ./... && go mod tidy                   # Update dependencies
```

## TUI (Terminal User Interface)

MystiSql 默认启动交互式 TUI 界面，提供类似 MySQL 命令行的简洁体验。

### TUI 特性

- **默认启动**: 运行 `mystisql` 无参数时自动启动 TUI
- **简洁界面**: 类似 MySQL 的命令行风格，无复杂装饰
- **依赖注入**: TUI 使用配置系统和实例注册中心，无硬编码
- **响应式布局**: 动态适配不同终端尺寸
- **多行输入**: 支持多行 SQL 输入
- **历史命令**: 上下键浏览历史命令
- **实例切换**: Tab 键切换数据库实例
- **语法高亮**: SQL 关键字高亮显示

### TUI 界面示例

```
Welcome to the MystiSql monitor. Commands end with Enter.
Your MystiSql connection has 1 instance(s) configured.
Current instance: local-mysql

Type 'help' or '?' for help. Type 'exit' or Ctrl+C to quit.

mystisql@local-mysql> SELECT * FROM users LIMIT 2;
id  name    email
1   Alice   alice@example.com
2   Bob     bob@example.com

2 rows, 0.005s

mystisql@local-mysql> INSERT INTO users (name, email) VALUES ('Charlie', 'charlie@example.com');
受影响行数: 1
最后插入ID: 3
执行时间: 0.003s

mystisql@local-mysql> _
```

### TUI 快捷键

| 快捷键 | 功能 |
|--------|------|
| Enter | 执行 SQL |
| Tab | 切换实例 |
| Ctrl+E | 导出结果 |
| ? | 显示帮助 |
| Ctrl+C | 退出 |
| ↑/↓ | 浏览历史命令 / 导航列表 |
| Esc | 取消当前操作 |

### REPL 交互界面

MystiSql 默认启动 MySQL 鈄式的 REPL (Read-Eval-Print-Loop) 交互界面，使用 `golang.org/x/term` 实现终端原始模式。

**REPL 文** (`internal/cli/repl/`)：
```
repl.go        # REPL 核心引擎
input.go      # 多行输入处理
formatter.go  # 输出格式化
commands.go  # 内置命令
history.go    # 历史管理
readline.go  # ReadLine 支持
```

**快捷键**:
| 快捷键 | 功能 |
|--------|------|
| Enter | 执行 SQL 或命令 |
| Tab | 自动补全（未来功能） |
| Ctrl+C | 中断当前输入/退出 |
| ↑/↓ | 浏览历史命令 |
| Esc | 取消当前操作 |
| Ctrl+D | 删除字符 |
| Ctrl+A | 移动到行首 |
| Ctrl+E | 移动到行尾 |

| Ctrl+U | 删除到行尾 |
| Ctrl+K | 删除整行 |

**提示符**:
- 新语句: `mystisql@instance>`
- 续行: `    ->`、`    '>`、`    ">`、`    `>`（反引号）
- 注释后: `    -- `或 `#`

**内置命令**:
| 命令 | 稡格式 | 描述 |
|------|---------|------|
| `exit`, `quit`, `\q` | 退出 REPL | exit |
| `help`, `?`, `\h` | 显示帮助 | clear, `\c` | 清除当前输入 | status, `\s` | 显示状态 | print, `\p` | 打印当前输入 | edit, `\e` | 编辑当前输入 ($EDITOR) | ego, `\G` | 埥询并垂直显示结果 | go, `\g` | 执行查询 | use, `instance` | 切换实例 | prompt, `\R` | 自定义提示符 | source, `\.` | 执行脚本文件 | system, `\!` | 执行系统命令 | output, `\o` | 设置输出格式 (csv, json) |

**输出格式**:
- 表格格式（默认): ASCII 表格，对齐
- 尴直格式 (`\G`): 每行一列显示
- CSV 格式 (`\o csv`)
- JSON 格式 (`\o json`)
**历史命令**:
- 匁存在 `~/.mystisql/history`
- 上下箭头浏览历史
- 匁久去重
**开发**:
REPL 代码位于 `internal/cli/repl/`，使用 `golang.org/x/term` 庄。
**依赖注入**:
```go
// 创建 REPL 应用时传递配置和注册中心
app := NewREPL(cfg, registry)
app.Run()
```
**测试**:
```bash
# 运行 REPL 单元测试
go test -v ./internal/cli/repl/...
```

**测试**:
```bash
# 运行 TUI 单元测试
go test -v ./internal/cli/... -run TestTUI

# 运行 TUI 集成测试
go test -v ./internal/cli/... -run "TestTUIWith|TestTUIInstance"
```

### query 子命令

使用 `query` 子命令直接执行 SQL（不进入 TUI）:
```bash
mystisql query --instance local-mysql "SELECT * FROM users"
mystisql query "SELECT 1"  # 使用默认实例
```

## Project Structure

```
cmd/mystisql/              # Main CLI entry point
internal/                  # Private application code
  connection/             # Database connection layer (mysql/, postgresql/, oracle/, redis/)
  discovery/              # Service discovery (k8s/, config/, static/)
  service/                # Core service layer (query/, auth/, audit/, cache/)
  api/                    # API layer (rest/, websocket/)
  cli/                    # CLI implementation
pkg/                      # Public library code
  types/                  # Shared types
  errors/                 # Error definitions
config/                   # Configuration files
test/                     # Integration tests
```

## Code Style Guidelines

### General Principles
- Follow [Effective Go](https://golang.org/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write simple, readable code over clever code
- Keep functions small (< 50 lines) and focused
- Keep files small (< 500 lines)
- Prefer composition over inheritance

### Imports
Group imports: standard library, third-party, local packages. Use blank lines between groups.

```go
import (
    "context"
    "fmt"
    
    "github.com/gin-gonic/gin"
    "k8s.io/client-go/kubernetes"
    
    "mystisql/internal/connection"
    "mystisql/pkg/types"
)
```

### Naming Conventions
- **Packages**: lowercase, single word: `connection`, `discovery`, `service`
- **Variables/Functions**: CamelCase (not snake_case)
- **Exported**: `ConnectionPool`, `QueryEngine`
- **Unexported**: `connectionPool`, `queryEngine`
- **Acronyms**: consistent casing: `HTTPServer`, not `HttpServer`
- **Interfaces**: nouns or verbs: `Reader`, `Writer`, `ConnectionPool`
- **Constants**: group related ones

### Error Handling
Always handle errors explicitly with context:

```go
if err != nil {
    return fmt.Errorf("failed to connect to database %s: %w", instanceName, err)
}
```

Define sentinel errors:
```go
var (
    ErrInstanceNotFound = errors.New("instance not found")
    ErrConnectionFailed = errors.New("connection failed")
)
```

### Context Usage
- Always pass `context.Context` as first parameter
- Don't store context in structs
- Use context for cancellation and timeouts

```go
func (s *Service) Query(ctx context.Context, query string) (*Result, error) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    // ...
}
```

### Testing
- File: `connection_test.go` for `connection.go`
- Function: `TestFunctionName`
- Use AAA pattern (Arrange, Act, Assert)
- Prefer table-driven tests

```go
func TestParseConnectionString(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *ConnectionConfig
        wantErr bool
    }{
        {"valid mysql", "mysql://user:pass@host:3306/db", &ConnectionConfig{Host: "host", Port: 3306}, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseConnectionString(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseConnectionString() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Types and Interfaces
Prefer interfaces for abstractions:

```go
type InstanceDiscoverer interface {
    Name() string
    Discover(ctx context.Context) ([]*DatabaseInstance, error)
    Watch(ctx context.Context) (<-chan DiscoveryEvent, error)
}
```

Use type aliases for clarity:
```go
type InstanceID string
type ConnectionString string
```

### Documentation
- Exported functions/types must have documentation comments
- Comment **why**, not what
- Use structured logging (zap)

```go
// QueryEngine handles SQL query parsing and routing.
type QueryEngine struct { /* ... */ }

// Execute runs a SQL query against the specified database instance.
func (e *QueryEngine) Execute(ctx context.Context, instance string, query string) (*Result, error) {
    // ...
}
```

## Architecture Patterns

### Layered Architecture (Bottom-up)
1. **Database Layer**: MySQL, PostgreSQL, Oracle, Redis instances
2. **Discovery Layer**: K8s API / Config / Static discovery
3. **Connection Layer**: Database drivers with connection pooling
4. **Service Layer**: Query engine, auth, audit, cache
5. **Access Layer**: REST API, WebSocket, CLI, JDBC driver

### Key Interfaces
- `InstanceDiscoverer`: Service discovery abstraction
- `ConnectionPool`: Database connection management
- `QueryEngine`: SQL parsing and routing

## Security & Best Practices

- Never log credentials, passwords, or sensitive data
- Use parameterized queries to prevent SQL injection
- Validate all user inputs
- Sanitize data before logging or displaying
- Use TLS for network communication
- Use connection pools and always close connections with `defer`
- Implement query timeouts and result set size limits
- Handle large result sets with streaming

## Concurrency

- Use goroutines judiciously with controlled concurrency
- Always use `defer` for cleanup in goroutines
- Capture loop variables properly in closures
- Use channels or sync primitives for coordination
- Don't spawn unbounded goroutines

```go
sem := make(chan struct{}, 10)
for _, item := range items {
    sem <- struct{}{}
    go func(item Item) {
        defer func() { <-sem }()
        process(item)
    }(item)
}
```

## Development Workflow

1. **Before starting**: Check README.md for current phase and roadmap
2. **Code**: Follow style guidelines, keep functions small
3. **Test**: Write unit tests, aim for high coverage
4. **Lint**: Run `golangci-lint run --fix` before committing
5. **Document**: Add/update comments for exported functions
6. **Review**: Ensure code compiles, tests pass, no linter warnings

## Pull Request Checklist

- [ ] Code compiles without errors
- [ ] All tests pass (`go test ./...`)
- [ ] New code has tests
- [ ] Code follows style guidelines
- [ ] No linter warnings (`golangci-lint run`)
- [ ] Documentation updated if needed
- [ ] No sensitive data in code
- [ ] Breaking changes documented

## Phase 3: Security Layer

Phase 3 adds enterprise-grade security capabilities to MystiSql. When working on Phase 3 features, consider:

### Security Features

**Authentication**:
- All API endpoints require authentication (except whitelisted paths like `/health`)
- Uses JWT tokens with HS256 signature
- Token management via CLI: `mystisql auth token --user-id <id> --role <role>`
- Tokens have configurable expiration (default: 24 hours)

**Audit Logging**:
- All SQL executions are logged with user info, SQL statement, execution time, rows affected
- Logs stored in JSON Lines format for easy processing
- Automatic log rotation (daily rotation, 30-day retention)
- Can be enabled/disabled via configuration

**SQL Validation**:
- Dangerous operations are blocked by default (DROP, TRUNCATE, DELETE without WHERE)
- Uses SQL parser (not regex) for accurate detection
- Whitelist/blacklist support for custom SQL filtering
- Returns 403 Forbidden when validation fails

**WebSocket Support**:
- Real-time query execution via WebSocket at `ws://host:port/ws`
- Authentication via URL parameter: `?token=<jwt>`
- Connection limits and idle timeout enforcement

**PostgreSQL Support**:
- PostgreSQL connections use `pgx` driver
- Connection pool management similar to MySQL
- SSL mode and timeout configuration supported

### Configuration

Phase 3 introduces new configuration options:

```yaml
auth:
  enabled: true
  token:
    secret: "your-secret-key"
    expire: "24h"

audit:
  enabled: true
  logFile: "/var/log/mystisql/audit.log"
  retentionDays: 30

validator:
  enabled: true
  dangerousOperations:
    - DROP
    - TRUNCATE
    - DELETE_WITHOUT_WHERE
  whitelist:
    - "SELECT * FROM system_config"
  blacklist:
    - "DELETE FROM audit_log"

websocket:
  maxConnections: 1000
  idleTimeout: "10m"
  maxConcurrentQueries: 5
```

### Testing Phase 3 Features

**Unit Tests**:
```bash
# Test auth service
go test -v ./internal/service/auth/...

# Test audit logging
go test -v ./internal/service/audit/...

# Test SQL validator
go test -v ./internal/service/validator/...

# Test transaction management
go test -v ./internal/service/transaction/...

# Test batch operations
go test -v ./internal/service/batch/...
```

**Integration Tests**:
```bash
# Test CLI auth commands
go test -v ./internal/cli/... -run TestAuth

# Test API authentication middleware
go test -v ./internal/api/middleware/...
```

### CLI Commands for Phase 3

```bash
# Generate token
mystisql auth token --user-id admin --role admin --server http://localhost:8080

# Use token for queries
mystisql query --instance local-mysql "SELECT * FROM users" --token <jwt-token>

# View token info
mystisql auth info --token <jwt-token>

# Revoke token
mystisql auth revoke --token <jwt-token>
```

### Security Best Practices

When implementing Phase 3 features:

1. **Never log sensitive data**: Don't log passwords, tokens, or query results
2. **Use parameterized queries**: Prevent SQL injection
3. **Validate all inputs**: Check user inputs before processing
4. **Sanitize logs**: Remove sensitive information before logging
5. **Token management**: Tokens should be short-lived and easily revocable
6. **Audit everything**: All SQL executions must be logged for compliance
7. **Block dangerous operations**: Default to safe - block DROP, TRUNCATE, etc.
8. **Connection limits**: Enforce max connections and timeouts
9. **Error handling**: Don't expose internal errors to users

### Database Connection Pooling (Phase 3)

PostgreSQL connection pooling reuses MySQL's ConnectionPool interface:
- Same configuration parameters (MaxOpen, MaxIdle, MaxLifetime)
- Automatic health checking with `SELECT 1`
- Connection recycling on errors

### JDBC Enhancements (Phase 3)

**Transaction Support**:
```bash
# Begin transaction
POST /api/v1/transaction/begin {"instance": "local-mysql"}

# Execute queries with transaction ID
POST /api/v1/query {"instance": "local-mysql", "sql": "INSERT ...", "transactionId": "tx-xxx"}

# Commit or rollback
POST /api/v1/transaction/commit {"transactionId": "tx-xxx"}
POST /api/v1/transaction/rollback {"transactionId": "tx-xxx"}
```

**Batch Operations**:
```bash
# Execute batch
POST /api/v1/batch {
  "instance": "local-mysql",
  "queries": ["INSERT 1", "INSERT 2", "INSERT 3"],
  "stopOnError": true
}
```

## Communication Language

- **Primary Language**: Chinese
- **Documentation**: Chinese preferred
- **Code Comments**: Chinese or English (consistent per file)
- **Communication**: All interactions with users should be in Chinese
- **Generated Documents**: Should be in Chinese whenever possible

This ensures consistent communication with the project maintainers and aligns with this project's target audience.

## E2E Testing

### Quick Start

```bash
# 检查测试环境
make e2e-check

# 启动测试环境
make e2e-setup

# 运行 e2e 测试
make e2e-test

# 清理测试环境
make e2e-teardown

# 重置测试数据
make e2e-reset
```

### E2E Test Commands

```bash
# 运行所有 e2e 测试
go test -v -tags=e2e ./test/e2e/...

# 运行特定测试
go test -v -tags=e2e -run TestMySQLBasic ./test/e2e/...

# 运行并生成覆盖率报告
make e2e-test-coverage
```

### E2E Test Guidelines

When working on E2E tests:

1. **Use build tags**: All e2e tests must use `//go:build e2e` tag
2. **Skip in short mode**: Use `SkipIfShort(t)` at the beginning of each test
3. **Clean up test data**: Always clean up inserted data after tests
4. **Use helper functions**: Leverage existing helper functions in `test/e2e/helper.go`
5. **Use fixtures**: Generate test data using `GenerateTestUser()`, `GenerateTestProduct()`, etc.
6. **Document tests**: Add clear comments explaining what each test validates

### E2E Test Structure

- `test/e2e/config.go` - Test configuration loading
- `test/e2e/helper.go` - Helper functions for database connections and cleanup
- `test/e2e/fixture.go` - Test data generators
- `test/e2e/basic_test.go` - Basic connection and query tests
- `test/e2e/*.sql` - Database initialization scripts

### Running E2E Tests in CI/CD

E2E tests can be optionally run in CI/CD environments:

```yaml
# GitHub Actions example
- name: Run E2E Tests
  run: |
    # Only run if Podman is available
    if command -v podman &> /dev/null; then
      make e2e-setup
      make e2e-test
      make e2e-teardown
    fi
```

For detailed E2E testing documentation, see [test/e2e/README.md](test/e2e/README.md)
