# MystiSql Agent Guidelines

Guidelines for AI coding agents working on the MystiSql codebase.

## Project Overview

MystiSql is a database access gateway for Kubernetes clusters, supporting MySQL, PostgreSQL, Oracle, and Redis. It provides CLI, WebUI, RESTful API, WebSocket, and JDBC driver interfaces.

**Current Stage**: Early development (Phase 1: Infrastructure layer). See README.md for detailed roadmap and architecture.

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
