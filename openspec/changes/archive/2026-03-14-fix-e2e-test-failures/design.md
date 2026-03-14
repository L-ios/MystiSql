## Context

当前 MystiSql 的 E2E 测试存在以下问题：
1. **WebSocket 端点未注册**：`/ws` 路由未在 server.go 中注册，导致 WebSocket 测试返回 404
2. **Batch 测试 SQL 不完整**：INSERT 语句缺少 `password_hash` 必填字段
3. **数据初始化问题**：TestMySQLQuery 期望至少有 1 条用户数据

项目使用 Gin 框架，WebSocket 规范已在 `openspec/specs/websocket-support/spec.md` 中定义。

## Goals / Non-Goals

**Goals:**
- 实现 WebSocket 端点 `/ws`，支持 Token 认证
- 修复 Batch 测试 SQL 语句
- 确保 E2E 测试通过率达到 90%+

**Non-Goals:**
- 不修改 WebSocket 规范（已定义）
- 不重构现有 REST API
- 不修改数据库 schema

## Decisions

### 1. WebSocket 实现方案

**决定**: 使用 `gorilla/websocket` 库，在 `server.go` 中注册 `/ws` 路由

**备选方案**:
- A) 使用 Gin 的 WebSocket 中间件 - 不选择，因为需要额外依赖且灵活性较低
- B) 使用标准库 `golang.org/x/net/websocket` - 不选择，因为 API 不如 gorilla 成熟

**理由**:
- `gorilla/websocket` 是 Go 生态中最成熟的 WebSocket 库
- 项目已依赖此库（`test/e2e/websocket_e2e_test.go` 使用）
- 支持 RFC 6455 全部特性

### 2. WebSocket 处理器架构

**决定**: 创建独立的 `WebSocketHandlers` 结构体

**架构**:
```
Server
  └── websocketHandlers *WebSocketHandlers
        ├── upgrader    *websocket.Upgrader
        ├── authService *auth.AuthService
        ├── engine      *query.Engine
        └── logger      *zap.Logger
```

**理由**:
- 与现有 `AuthHandlers`、`TransactionHandlers` 等保持一致
- 便于测试和依赖注入
- 职责单一

### 3. 认证方式

**决定**: 从 URL 参数 `?token=<jwt>` 获取 Token

**理由**:
- 符合 `websocket-support` 规范定义
- WebSocket 握手时无法添加 Authorization header
- 与 E2E 测试预期一致

### 4. Batch 测试修复

**决定**: 在 INSERT 语句中添加 `password_hash` 字段

**修改**:
```sql
-- 修改前
INSERT INTO users (username, email) VALUES ('batch_test_1', 'batch1@test.com')

-- 修改后
INSERT INTO users (username, email, password_hash) VALUES ('batch_test_1', 'batch1@test.com', 'test_hash')
```

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| WebSocket 连接数过多导致资源耗尽 | 实现 `maxConnections` 限制（配置默认 1000） |
| 空闲连接占用资源 | 实现 `idleTimeout` 自动断开（默认 10 分钟） |
| 并发查询过多 | 实现 `maxConcurrentQueries` 限制（默认 5） |
| Batch 测试数据污染 | 使用唯一前缀 `batch_test__*`，测试后清理 |

## Migration Plan

1. **Phase 1**: 创建 `websocket_handlers.go`
2. **Phase 2**: 在 `server.go` 中注册路由
3. **Phase 3**: 修复 `batch_e2e_test.go`
4. **Phase 4**: 运行 E2E 测试验证

**回滚策略**: 
- WebSocket 路由可通过配置 `websocket.enabled: false` 禁用
- Batch 测试修复是纯添加字段，无需回滚
