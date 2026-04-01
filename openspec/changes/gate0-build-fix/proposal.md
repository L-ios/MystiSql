## Why

项目当前无法编译：`web/handler.go` 的 `//go:embed all:dist` 因 `web/dist/` 不存在导致构建失败。`go.sum` 已通过 `go mod tidy` 修复（所有 9 个驱动依赖均已就绪，ClickHouse/Elasticsearch/etcd 均可编译）。

## What Changes

- **WebUI embed 改为可选**：拆分 `web/handler.go` 为两个文件，通过 build tag 控制：
  - `web_handler_embed.go`（`//go:build webembed`）— 包含 `//go:embed all:dist`，构建时加 `-tags webembed`
  - `web_handler_stub.go`（默认，无 build tag）— `NewHandler()` 返回 nil，server.go 已有 nil 检查（line 119-121）

## Capabilities

### New Capabilities
- `build-fixes`: WebUI 可选 embed

## Impact

### 受影响的代码
- `web/handler.go` → 删除，拆分为 `web_handler_embed.go` + `web_handler_stub.go`

### Done 标准
- `go build ./...` 零错误
- `go vet ./...` 零错误

### 信心
**99%** — 改动极简：仅拆一个文件，且实现已完成
