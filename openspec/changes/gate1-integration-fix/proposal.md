## Why

项目有大量已实现但不可用的功能：Engine 硬编码仅注册 MySQL/PG 工厂（Oracle/Redis/SQLite/MSSQL 不可达）、RBAC handlers 未注册路由、RBAC 中间件信任客户端 Header、Health Monitor 从未启动。`internal/api/websocket/` 包实现了完整的 WebSocket handler（强类型消息、ConnectionManager、idle 清理、审计上下文注入），但实际运行的是 `rest/websocket_handlers.go` 的简陋实现（弱类型、不注入审计上下文、不传递配置、无优雅关闭）。同时存在 pool.go 递归栈溢出风险、审计日志轮转不生效、Transaction goroutine 泄漏和数据竞争等可靠性问题。

## What Changes

### 集成断裂修复
- **Engine 驱动工厂动态注册**：从硬编码 MySQL/PG 改为通过 `connection.DriverRegistry` 按实例类型动态查找，Oracle/Redis/SQLite/MSSQL 立即可用
- **注册 RBAC 路由**：`server.go` 的 `setupRoutes()` 注册已有 RBAC handlers
- **RBAC 中间件安全修复**：从 JWT claims 提取角色（auth middleware 需先写入 role 到 gin context），移除 `X-User-Roles` Header 读取
- **启动 Health Monitor**：在 `serve.go` 中实例化并 `Start()`
- **WebSocket 正式接入**：用 `internal/api/websocket/` 包替换 `rest/websocket_handlers.go`，接入 server.go 路由、从配置文件读取 WebSocket 参数、服务关闭时优雅关闭 WS 连接；消息协议 JSON tag 统一为 camelCase（`requestId`）与现有客户端兼容

### 可靠性修复
- **pool.go 递归改循环**：for 循环 + maxRetries=3 替代递归调用
- **审计 Rotator/Writer 集成**：Rotator 持有 Writer 引用，轮转时调用 `Writer.Rotate()` 刷出缓冲并重新打开文件
- **CLI auth token nil body 修复**：`auth_cmd.go` 发送 JSON body 而非 nil
- **ExecuteExec 添加审计日志**：与 ExecuteQuery 格式一致
- **Transaction goroutine 泄漏**：添加 stopCh，`Close()` 时通知 cleanup goroutine 退出
- **Transaction 数据竞争**：`GetTransaction` 中 `LastActivityAt` 写操作从 RLock 改为写锁
- **连接池配置外部化**：从配置文件读取 MaxOpen/MaxIdle/MaxLifetime

## Capabilities

### New Capabilities
- `integration-fix`: Engine 工厂动态注册、RBAC 路由注册、Health Monitor 启动
- `websocket`: 用 `internal/api/websocket/` 包替换 `rest/websocket_handlers.go`，接入配置、路由、优雅关闭

### Modified Capabilities
- `rbac-permissions`: 中间件从 JWT claims 读角色，路由注册到 server.go
- `audit-logging`: ExecuteExec 审计 + Rotator/Writer 集成
- `rest-api`: 注册 RBAC 路由，WebSocket handler 替换
- `token-auth`: CLI auth token nil body 修复

## Impact

### 受影响的代码
- `internal/service/query/engine.go` — 工厂注册
- `internal/api/rest/server.go` — 路由注册、WebSocket handler 替换、Shutdown 添加 WS 关闭
- `internal/api/websocket/handler.go` — JSON tag 改 camelCase
- `internal/api/websocket/message.go` — JSON tag 改 camelCase
- `internal/api/rest/websocket_handlers.go` — 删除（被 websocket 包替代）
- `jdbc/src/main/java/io/github/mystisql/jdbc/client/WebSocketTransport.java` — 适配新协议（type 字段替代 action/success）
- `internal/service/rbac/middleware.go` — JWT claims 读角色
- `internal/api/middleware/auth.go` — 写入 role 到 context
- `internal/service/health/monitor.go` — 在 serve.go 中启动
- `internal/connection/pool/pool.go` — 递归改循环
- `internal/service/audit/rotator.go + writer.go` — 生命周期集成
- `internal/service/transaction/manager.go` — stopCh + 锁策略
- `internal/service/query/engine.go` — ExecuteExec 审计
- `internal/cli/auth_cmd.go` — nil body 修复
- `pkg/types/config.go` — 根 Config 添加 WebSocket 字段
- `internal/cli/serve.go` — 传递 WebSocket 配置

### 前置条件
- Gate 0 (build-fix) 必须先完成：`go build ./...` 通过

### Done 标准
- Oracle/Redis/SQLite/MSSQL 实例通过 API 查询可达
- `POST /api/v1/rbac/roles` 返回非 404
- `go test -race ./internal/service/transaction/...` 无竞争报告
- 审计日志轮转后无数据丢失
- `GET /ws?token=<jwt>` 使用 `internal/api/websocket/` 包处理（非 rest 包的旧实现）
- WebSocket 查询的审计日志包含正确 user_id（非 "unknown"）
- WebSocket 配置从 config.yaml 读取，非硬编码
- `Server.Shutdown()` 关闭所有活跃 WebSocket 连接

### 验证状态（2026-03-29）

**前置条件**：`go build ./...` ✅ 通过

#### 逐决策信心

| 决策 | 信心 | 状态 | 备注 |
|------|------|------|------|
| D1 Engine 工厂注册 | 95% | ✅ 已验证 | GetFactory 返回 (factory, error)，接口兼容 |
| D2 RBAC 中间件 | 95% | ✅ 已验证 | auth middleware 确认写 role(string) 到 context |
| D3 RBAC 路由注册 | 95% | ✅ 已验证 | CRUD 方法签名明确 |
| D5 Health Monitor | 95% | ✅ 已验证 | Start/Stop 签名确认 |
| D6 Pool 递归改循环 | 95% | ✅ 已验证 | 2 处递归(line 189, 236)，无 retry counter |
| D7 审计 Rotator/Writer | 90% | ✅ 已验证 | 无 writer 引用、无 WaitGroup、Rotate() 是死代码 |
| D8 Transaction 修复 | 90% | ✅ 已验证 | 三 bug 全确认 |
| D9 ExecuteExec 审计 | 95% | ✅ 已验证 | 缺失确认 |
| D10 CLI nil body | 98% | ✅ 已验证 | 简单 JSON marshal |
| D11 WebSocket 替换 | 90% | ✅ 已验证 | JDBC 适配新协议，websocket 包保持不变 |
| D12 JSON tag 统一 | 90% | ✅ 已验证 | 依赖 D11 方案已确认 |
| D13 Config 接入 | 95% | ✅ 已验证 | string→Duration 标准转换 |

#### 阻塞项（信心 < 90%）

无。

**综合信心：93%** — 所有决策 ≥ 90%，可进入实施
