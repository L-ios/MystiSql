## Context

MystiSql 项目在 Phase 3 开发过程中存在构建阻塞问题：

**WebUI embed 阻塞编译**：`web/handler.go` 中的 `//go:embed all:dist` 要求 `web/dist/` 目录存在，但前端构建产物通常不在 Git 仓库中。没有前端构建环境的开发者（或 CI 环境）无法通过 `go build`。

当前 `go.sum` 已通过 `go mod tidy` 修复，所有 9 个驱动依赖均可编译。

> **注意**：`internal/api/websocket/` 包（567 行）虽暂无外部导入，但这是 Phase 3 计划中的 WebSocket 功能代码，正在等待集成，**不属于死代码**。

## Goals / Non-Goals

**Goals:**
- `go build ./...` 零错误，无需前端构建环境
- `go vet ./...` 零错误
- 保留 WebUI 完整功能（通过 build tag 按需启用）

**Non-Goals:**
- 不删除 `internal/api/websocket/` — 这是未完成的功能代码，等待后续集成
- 不修改 `rest/websocket_handlers.go`
- 不引入新的依赖或工具链
- 不修改前端构建流程

## Decisions

### Decision 1: Build Tag 策略 — 默认排除 embed

**选择**：使用 `//go:build webembed` 和 `//go:build !webembed` 拆分为两个文件。

**理由**：
- 默认构建（`go build`）不包含 WebUI，开发者无需 Node.js 环境
- 生产构建（`go build -tags webembed`）包含完整 WebUI
- `server.go` 已有 nil 检查处理 Handler 为 nil 的情况（line 119-121），无需额外修改

**替代方案**：
- 空目录占位 `web/dist/.gitkeep` → 治标不治本，embed 空目录无意义
- Build-time 脚本生成 → 增加构建复杂度，不符合 Go 惯例

## Risks / Trade-offs

- **[风险] WebUI 功能回归** → `web_handler_embed.go` 保留了完整的 embed + SPA 路由逻辑，通过 `-tags webembed` 可验证。建议在 CI 中添加带 tag 的构建检查。
