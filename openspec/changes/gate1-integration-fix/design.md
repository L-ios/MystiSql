## Context

MystiSql 处于 Phase 3（安全层）开发阶段。代码库中存在大量"已实现但不可用"的功能——代码存在但未接入主流程。具体表现为：

- **Engine 硬编码工厂**：`NewEngine()` 仅注册 MySQL 和 PostgreSQL 两个工厂，绕过了已完整实现的 `DriverRegistry` 系统，导致 Oracle/Redis/SQLite/MSSQL 等 8 种数据库驱动不可达
- **RBAC/OIDC 路由缺失**：`server.go` 的 `setupRoutes()` 从未注册 RBAC 和 OIDC 路由，`RBACService.PermissionMiddleware` 从未应用到中间件链
- **RBAC 中间件不安全**：从客户端 `X-User-Roles` Header 读取角色，任意客户端可伪造
- **可靠性缺陷**：pool.go 递归栈溢出、审计轮转与写入生命周期脱节、Transaction goroutine 泄漏和数据竞争、CLI auth token 发送 nil body、ExecuteExec 绕过验证和审计

前置条件：Gate 0 (build-fix) 已完成，`go build ./...` 通过。

## Goals / Non-Goals

**Goals:**

- Engine 通过 `DriverRegistry` 动态查找工厂，所有已实现数据库驱动可用
- RBAC 路由注册到 HTTP Server，中间件链正确应用
- RBAC 中间件从 JWT claims（gin context）读取角色，移除 Header 信任
- Health Monitor 在服务启动时运行
- `internal/api/websocket/` 包正式接入 server.go，替换 `rest/websocket_handlers.go`
- WebSocket 消息协议 JSON tag 统一为 camelCase，与现有客户端兼容
- WebSocket 配置从 config.yaml 读取，非硬编码
- 服务关闭时优雅关闭所有 WebSocket 连接
- WebSocket 查询正确注入 UserID/ClientIP 到审计上下文
- 消除 pool.go 递归栈溢出风险
- 审计日志轮转后无数据丢失
- CLI `auth token` 命令正确发送 JSON body
- ExecuteExec 与 ExecuteQuery 执行相同的验证和审计逻辑
- Transaction Manager 无 goroutine 泄漏和数据竞争

**Non-Goals:**

- 不新增数据库驱动实现（仅让已有驱动可达）
- 不修改 JWT Token 的签名算法或 Claims 结构（保持向后兼容）
- 不重构 OIDC 流程的核心逻辑（仅修复 Callback 不发 JWT 的问题）
- 不改变 API 路由的 URL 模式
- 不做性能优化（仅修复正确性问题）
- 不实现 WebSocket 的 heartbeat 机制（config.example.yaml 中定义但未实现，后续 change 处理）

## Decisions

### D1: Engine 工厂查找改用 DriverRegistry

**选择**：删除 Engine 内部的 `factories map`，改为在启动时通过 `connection.GetRegistry()` 注册所有驱动，`getConnectionPool()` 通过 `registry.GetFactory(instance.Type)` 查找。

**替代方案**：
- (a) 每个 driver 包用 `init()` 自注册（Go database/sql 模式）—— 需要确保 import 触发，隐式依赖不够可控
- (b) 在 main/serve 层显式调用 `RegisterDriver()` —— **选定方案**，显式、可控、易于调试

**理由**：`DriverRegistry` 已完整实现（单例、线程安全、带测试），只是从未被生产代码使用。显式注册更符合项目的依赖注入风格（Engine 已接收 registry 参数）。

**变更范围**：
- `engine.go`：`NewEngine` 接收 `connection.DriverRegistry` 参数，删除 `factories` 字段，`getConnectionPool` 改用 `registry.GetFactory()`
- `serve.go`：启动时调用 `connection.GetRegistry().RegisterDriver()` 注册所有驱动
- `instances.go`, `query.go`：调整 `NewEngine()` 调用签名

### D2: RBAC 中间件改读 gin context

**选择**：`PermissionMiddleware` 从 `c.Get("roles")` 读取角色列表，而非 `c.GetHeader("X-User-Roles")`。

**前提**：auth middleware 已将 `role` 写入 context（`c.Set("role", claims.Role)`）。当前 `TokenClaims.Role` 是单字符串，RBAC 需要 `[]string`。

**方案**：auth middleware 在写入 `role`（string）的同时，写入 `roles`（`[]string{claims.Role}`）。RBAC 中间件读取 `roles` key。未来扩展多角色时只需修改 auth middleware 的 roles 来源。

**不变更 TokenClaims 结构**：保持 `Role string` 字段不变（JWT 兼容性），运行时在 auth middleware 层扩展为 `[]string`。

### D3: RBAC 路由注册

**选择**：在 `server.go` 的 `setupRoutes()` 中注册 RBAC CRUD 路由，并将 `PermissionMiddleware` 应用到需要权限控制的路由组。

**路由设计**（CRUD）：
- `POST /api/v1/rbac/roles` — 创建角色
- `GET /api/v1/rbac/roles` — 列出角色
- `PUT /api/v1/rbac/roles/:name` — 更新角色
- `DELETE /api/v1/rbac/roles/:name` — 删除角色
- `POST /api/v1/rbac/users/:id/roles` — 分配角色

RBAC 路由本身需要 auth middleware 保护（需登录），但不需要自身权限检查（角色管理暂定 admin-only，通过 `role == "admin"` 检查）。

### D5: Health Monitor 启动

**选择**：在 `serve.go` 的服务启动流程中实例化 `health.Monitor` 并调用 `Start()`，在 shutdown 时调用 `Stop()`。

**实现简单**：Monitor 已完整实现，只需在 `serveCmd` 的启动逻辑中接入。

### D6: pool.go 递归改循环

**选择**：将 `GetConnection()` 的递归重试改为 `for` 循环 + `maxRetries=3`。

**理由**：递归深度等于死连接数，MaxConnections=100 时可能栈溢出。循环是标准重试模式，maxRetries 防止无限循环。

**实现**：
```
for retry := 0; retry < maxRetries; retry++ {
    select {
    case conn := <-idle:
        if ping ok → return conn
        else → close, continue
    default:
        break // try creating new
    }
    // create new connection logic...
}
return error
```

### D7: 审计 Rotator/Writer 生命周期集成

**选择**：Rotator 持有 Writer 引用，轮转时调用 `Writer.Rotate()`；Rotator 添加 `sync.WaitGroup` 确保 `Stop()` 等待 goroutine 退出。

**理由**：Writer 已有正确的 `Rotate()` 实现（刷缓冲 + 关闭旧文件 + 打开新文件），只是从未被调用。最小改动是让 Rotator 持有 Writer 引用。

**变更**：
- `LogRotator` 新增 `writer *LogWriter` 字段
- 轮转时：`os.Rename()` → `writer.Rotate()` → 创建新空文件
- `Stop()` 使用 `sync.WaitGroup.Wait()` 等待 goroutine 退出
- `AuditService.Close()` 调用顺序：`rotator.Stop()`（等 goroutine 退出）→ `writer.Close()`

### D8: Transaction Manager 修复

**三个子修复**：

1. **Goroutine 泄漏**：`TransactionManager` 新增 `stopCh chan struct{}` 和 `sync.WaitGroup`。`cleanupExpiredTransactions` 监听 `stopCh`，`Close()` 关闭 `stopCh` 并等待 WaitGroup。

2. **数据竞争**：`GetTransaction()` 将 `RLock` 改为 `Lock()`（写锁），因为 `tx.LastActivityAt = time.Now()` 是写操作。

3. **长锁持有**：`rollbackTransaction()` 在持有写锁时执行网络 I/O（ROLLBACK）。改为：先复制需要回滚的事务列表，释放锁，再逐个回滚。

### D9: ExecuteExec 添加验证和审计

**选择**：提取 `validateAndAudit()` 辅助方法，`ExecuteQuery` 和 `ExecuteExec` 共用。

**理由**：两方法的验证和审计逻辑相同，提取可避免重复代码和未来不一致。

### D10: CLI auth token nil body 修复

**选择**：将 `userID` 和 `role` 序列化为 JSON body 发送。

**实现**：
```go
body, _ := json.Marshal(map[string]string{"user_id": userID, "role": role})
resp, err := client.Post(url, "application/json", bytes.NewReader(body))
```

### D11: WebSocket 用 `internal/api/websocket/` 包替换 `rest/websocket_handlers.go`

**选择**：用已实现但未接入的 `internal/api/websocket/` 包（`WebSocketHandler` + `ConnectionManager` + 强类型 Message）替换当前 `rest/websocket_handlers.go` 的简陋实现。

**理由**：
- websocket 包有完整的连接生命周期管理（ConnectionManager、idle 清理、stopCh 优雅停止）
- websocket 包正确注入 `query.ContextWithUserID` / `query.ContextWithClientIP` 到审计上下文（当前活跃代码用裸 `context.Background()`，审计日志全是 `"unknown"`）
- websocket 包有强类型消息（`Message`/`QueryResultMessage`/`ErrorMessage`/`PongMessage`），而非 `map[string]interface{}`
- websocket 包有 `Close()` 方法可优雅关闭所有连接（当前活跃代码无此能力）

**替代方案**：修补 `rest/websocket_handlers.go` —— 不选，因为修补量接近重写，且 websocket 包已经写好了更好的版本。

**变更范围**：
- `internal/api/websocket/message.go` — JSON tag 从 snake_case 改为 camelCase（`request_id` → `requestId`），与现有客户端和 E2E 测试保持兼容
- `internal/api/websocket/handler.go` — `ClientConnection.lastActivity` 需在 `ReadMessage`/`SendMessage` 时更新（idle 清理才能生效）
- `pkg/types/config.go` — 根 `Config` struct 添加 `WebSocket WebSocketConfig` 字段
- `internal/api/rest/server.go` — `Server.wsHandlers` 字段改为 `wsHandler *ws.WebSocketHandler`；`Setup()` 中用 `websocket.NewWebSocketHandler()` 替换；`setupRoutes()` 路由改为 `s.wsHandler.Handle`；`Shutdown()` 调用 `s.wsHandler.Close()`
- `internal/cli/serve.go` — 从 `cfg.WebSocket` 读取配置传递给 Server
- `internal/api/rest/websocket_handlers.go` — 删除

### D12: WebSocket 消息协议 JSON tag 统一

**选择**：将 `internal/api/websocket/message.go` 的 JSON tag 从 snake_case 改为 camelCase。

**具体变更**：
- `Message.RequestID`: `json:"request_id,omitempty"` → `json:"requestId,omitempty"`
- `QueryResultMessage.RequestID`: `json:"request_id"` → `json:"requestId"`
- `QueryResultMessage.RowCount`: `json:"row_count"` → `json:"rowCount"`
- `QueryResultMessage.ExecutionTime`: `json:"execution_time_ms"` → `json:"executionTimeMs"`
- `ErrorMessage.RequestID`: `json:"request_id,omitempty"` → `json:"requestId,omitempty"`

**理由**：当前活跃代码（`rest/websocket_handlers.go`）和 E2E 测试均使用 camelCase。现有客户端（JDBC/WebUI）已适配 camelCase 格式。websocket 包的 snake_case 从未在生产环境中使用过，改 camelCase 不会破坏任何现有客户端。

### D13: WebSocket 配置接入

**选择**：`pkg/types/config.go` 根 `Config` 添加 `WebSocket WebSocketConfig` 字段，`serve.go` 读取并传递给 Server。

**配置映射**（yaml → websocket.Config）：
```yaml
websocket:
  enabled: true
  maxConnections: 1000        # → Config.MaxConnections
  idleTimeout: "10m"          # → Config.IdleTimeout (time.ParseDuration)
  maxConcurrentQueries: 5     # 暂不使用（handleConnection 是同步循环）
```

**`websocket.Config` 新增字段映射**：从 `types.WebSocketConfig` 转换时，`ReadBufferSize`/`WriteBufferSize`/`EnableCompression` 使用默认值（1024/1024/false），暂不从配置文件读取。

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| Engine 签名变更影响所有调用方 | 调用方只有 3 处（serve.go, instances.go, query.go），改动明确 |
| RBAC 中间件改动可能影响已有 API 行为 | 当前 RBAC 从未生效，无已有行为需要兼容 |
| pool.go 循环重试可能延迟错误返回 | maxRetries=3 上限，总延迟可控 |
| Transaction `GetTransaction` 改写锁可能影响并发性能 | 写锁持有时间极短（仅赋值一行），影响可忽略 |
| 审计 Rotator 持有 Writer 引入循环依赖 | 通过接口解耦或 AuditService 协调注入 |
| WebSocket JSON tag 变更可能影响客户端 | snake_case 版本从未在生产中使用，现有客户端已适配 camelCase |
| WebSocket 替换后 E2E 测试需验证 | E2E 测试使用 camelCase，与新实现一致，改动量小 |
| `handleConnection` 中 `go h.handleMessage` 无并发限制 | 当前每条消息一个 goroutine，可后续加 semaphore 限流（本次不改） |

## Open Questions

1. **RBAC 路由的权限粒度**：当前 RBAC 路由本身是否需要按角色限制（如只有 admin 可管理角色）？建议暂时硬编码 admin 检查。
2. **连接池配置外部化范围**：当前 pool 配置是否从 config.yaml 读取？还是仅添加常量默认值？建议从 config 读取，提供合理默认值。
3. **WebSocket handleMessage 并发模型**：当前 `go h.handleMessage` 为每条消息启动 goroutine，无并发限制。是否需要在本次 change 中加 semaphore？建议不加（本次聚焦集成），后续 change 处理。
4. **WebSocket config.example.yaml 中的 heartbeat 配置**：定义了但未实现，是否从 example 中移除？建议移除或标注为 planned，避免误导用户。

## 验证结果（2026-03-29）

`go build ./...` ✅ 通过。以下为逐决策代码级验证结论。

### D1 验证 ✅

- `DriverRegistry.GetFactory()` 返回 `(ConnectionFactory, error)` 而非 `(ConnectionFactory, bool)`，与 Engine 当前 map 查找语义不同
- 改动极小：`factory, exists := e.factories[t]` → `factory, err := e.driverRegistry.GetFactory(t)`
- `ConnectionFactory` 接口完全兼容
- Engine 当前无 DriverRegistry 引用，需新增字段或通过 `connection.GetRegistry()` 全局单例获取
- 所有 4 处 `NewEngine()` 调用点（serve.go:55, instances.go:98,132, query.go:48）传 `discovery.InstanceRegistry`，需决定是否新增第二参数

**信心：95%**

### D2 验证 ✅

- auth middleware 精确设置 3 个 context key：`user_id`(string), `role`(string), `token`(string)
- RBAC `PermissionMiddleware` 读取 `X-User-Roles` header（不安全），完全不读 context
- 修复明确：改 `c.GetHeader("X-User-Roles")` → `c.Get("role")`
- `TokenClaims.Role` 是 `string`（单角色），RBAC 需 `[]string` → auth middleware 额外写入 `roles`

**信心：95%**

### D3 验证 ✅

- `RBACService` 方法签名全部明确：`AddRole(Role)`, `ListRoles()[]Role`, `GetRole(string)(Role,bool)`, `DeleteRole(string)`
- `NewRBACService()` 零参数，无依赖注入
- 用户角色管理：`AssignRolesToUser(userID string, roleNames []string)`

**信心：95%**

### D4 验证 ✅

- `NewMonitor(registry, engine, logger, interval)` — 参数类型全部匹配
- `Start()` / `Stop()` 无参数无返回值
- 注意：`Monitor` 缺少 `CheckInstanceHealth(ctx, string)` 方法，不满足 `HealthChecker` 接口（但 server.go 可能不通过接口使用）

**信心：95%**

### D6 验证 ✅

- 精确 2 处递归：pool.go line 189（idle path）和 line 236（wait path）
- 无 retry counter，无 max depth
- 池未满时创建新连接（line 269），满时等待回收（line 216）

**信心：95%**

### D7 验证 ✅

- `LogRotator` struct：无 writer 字段、无 WaitGroup
- `Stop()` 只 `close(stopCh)`，不等 goroutine 退出
- `LogWriter.Rotate()` 存在于 line 154-179，但从未被调用（死代码）
- `AuditService.Close()` 中 `rotator.Stop()` 不等 → `writer.Close()` 有竞争

**信心：90%**

### D8 验证 ✅

- `TransactionManager` struct 无 stopCh/ctx/cancel → goroutine 泄漏
- `GetTransaction()` line 224 用 `RLock`，line 237 写 `LastActivityAt` → 数据竞争
- `rollbackTransaction()` 在 4 个调用方持有的 `Lock()` 内执行网络 I/O
- `Close()` 遍历所有事务逐个 rollback，全程持锁

**信心：90%**

### D11 验证 ⚠️ 严重不兼容

websocket 包的协议结构与 JDBC/E2E 已适配的格式**不只是 JSON tag 不同，而是整个响应形状不同**：

| 维度 | websocket 包（新） | rest 包（旧，JDBC 已适配） |
|---|---|---|
| 成功标志 | 无 `success` 字段，用 `type: "query_result"` | `success: true` |
| 错误格式 | `{type:"error", code:..., message:...}` | `{success:false, error:"..."}` |
| Pong 响应 | `{type:"pong", timestamp:...}` | `{action:"pong"}` |
| 类型区分符 | `type` 字段 | `action` 字段 |

JDBC 客户端 `WebSocketTransport.java`：
- 用 `isSuccess()` 判断成功（依赖 `success` 字段）
- 用 `getAction()=="pong"` 判断心跳（依赖 `action` 字段）
- 用 `getRequestId()` 匹配请求（依赖 `requestId` 字段）

**直接替换会导致 JDBC 客户端完全不可用。**

**信心：60%** — 需决定协议适配方向

### D13 验证 ✅

- `types.WebSocketConfig.IdleTimeout` 是 `string`（yaml: `"10m"`）
- `websocket.Config.IdleTimeout` 是 `time.Duration`
- `types.Config` 根 struct **无** `WebSocket` 字段
- 转换函数：`time.ParseDuration(cfg.WebSocket.IdleTimeout)` — 标准，无风险
- `config.yaml` 无 websocket 段，`config.example.yaml` 有

**信心：95%**
