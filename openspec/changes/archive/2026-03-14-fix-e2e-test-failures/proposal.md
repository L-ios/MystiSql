## Why

当前 E2E 测试存在 3 类失败问题，阻塞了 CI/CD 流程：
1. WebSocket 端点未实现，导致 WebSocket 相关测试全部失败（404 错误）
2. Batch 测试 SQL 语句缺少必填字段 `password_hash`，导致插入失败
3. MySQLQuery 测试期望至少有 1 条用户数据，但数据库可能未正确初始化

这些问题导致 E2E 测试通过率仅为 59%（13/22），需要修复以恢复测试稳定性。

## What Changes

### 1. WebSocket 端点实现
- 在 `server.go` 中注册 WebSocket 路由 `/ws`
- 实现 WebSocket 连接处理器（握手、认证、消息处理）
- 支持 Token 认证（URL 参数 `?token=<jwt>`）
- 实现查询执行和心跳机制

### 2. Batch 测试修复
- 修改 `batch_e2e_test.go` 中的 INSERT 语句，添加 `password_hash` 字段

### 3. 数据初始化检查
- 确保 E2E 测试前数据库已正确初始化
- 检查 `TestMySQLQuery` 测试逻辑是否合理

## Capabilities

### New Capabilities
无新增能力（WebSocket 能力已在 `websocket-support` 规范中定义）

### Modified Capabilities
- `websocket-support`: 需要实现之前定义但未实现的功能（WebSocket 端点注册）

## Impact

- **代码影响**:
  - `internal/api/rest/server.go` - 添加 WebSocket 路由注册
  - `internal/api/rest/websocket_handlers.go` - 新增 WebSocket 处理器
  - `internal/cli/serve.go` - 可能需要初始化 WebSocket 服务

- **测试影响**:
  - `test/e2e/batch_e2e_test.go` - 修复 INSERT 语句
  - E2E 测试通过率预计从 59% 提升到 90%+

- **配置影响**:
  - `e2e-test-config.yaml` - WebSocket 配置已存在，可能需要启用
