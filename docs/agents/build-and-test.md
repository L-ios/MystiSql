# Build/Lint/Test

**优先使用 Makefile 目标，不要直接调用 go 命令。** 运行 `make help` 查看所有可用目标。

## 构建

```bash
make build                    # 构建当前平台二进制
make build-linux              # 构建 Linux AMD64
make build-darwin             # 构建 macOS (AMD64 + ARM64)
make build-windows            # 构建 Windows AMD64
make release                  # 构建所有平台 release 二进制 + checksums
make dev                      # 构建并启动开发服务器
make clean                    # 清理构建产物
```

WebUI 构建支持 build tag：
- 默认构建（不含 WebUI）：`make build`
- 含 WebUI 静态资源：先 `cd web && npm run build`，再 `go build -tags webembed ./cmd/mystisql`

## 测试

```bash
make test                     # 运行所有单元测试
make test-coverage            # 运行测试并生成覆盖率报告 (coverage.html)

# 单独运行特定测试（Makefile 不覆盖此场景，直接用 go test）
go test -v ./path/to/package -run TestFunctionName
go test -race ./internal/connection/mysql
```

## 代码检查

```bash
make fmt                      # 格式化代码 (go fmt)
make vet                      # 静态检查 (go vet)
make lint                     # 运行 golangci-lint
```

## 依赖管理

```bash
go mod tidy                   # 整理依赖（Makefile 未覆盖）
go mod verify                 # 验证依赖完整性
```

## 单元测试（按模块）

```bash
# 认证
go test -v ./internal/service/auth/...

# 审计日志
go test -v ./internal/service/audit/...

# SQL 验证器
go test -v ./internal/service/validator/...

# 事务管理
go test -v ./internal/service/transaction/...

# 批量操作
go test -v ./internal/service/batch/...

# API 中间件
go test -v ./internal/api/middleware/...

# CLI 认证命令
go test -v ./internal/cli/... -run TestAuth
```

## E2E 测试

```bash
make e2e-check                # 检查测试环境
make e2e-setup                # 启动测试环境
make e2e-test                 # 运行 e2e 测试
make e2e-test-coverage        # 运行 e2e 测试并生成覆盖率
make e2e-teardown             # 清理测试环境
make e2e-reset                # 重置测试数据
```

E2E 测试使用 `//go:build e2e` 构建标签，每个测试开头使用 `SkipIfShort(t)`。测试结构：
- `test/e2e/config.go` — 测试配置
- `test/e2e/helper.go` — 辅助函数
- `test/e2e/fixture.go` — 测试数据生成
- `test/e2e/basic_test.go` — 基础连接和查询测试

详见 [test/e2e/README.md](../../test/e2e/README.md)
