## Why

`BatchService` 的非事务模式（`executeSingleQuery`，service.go:218）当前返回 `"non-transaction batch execution not yet implemented"`，是 stub 实现。用户发送不含 transaction_id 的 batch 请求时会直接失败。需要实现每条查询独立获取连接、独立提交的模式。

## What Changes

- **实现 `executeSingleQuery`**：每条查询独立从连接池获取连接、执行、释放连接。不开启事务，每条查询自动提交
- **stopOnError 语义**：顺序模式下某条查询失败后，后续查询标记为 "skipped"；并行模式下已在 Gate 1 通过 errgroup 修复
- **结果格式**：每条查询独立返回 success/error/skipped 状态 + affectedRows 或 error message
- **连接泄漏防护**：使用 defer 确保每条查询执行后连接归还连接池
- **批量大小限制**：从配置读取 maxBatchSize（默认 1000），超限拒绝请求

## Capabilities

### Modified Capabilities
- `rest-api`: batch 端点支持无事务模式

## Impact

### 受影响的代码
- `internal/service/batch/service.go` — 实现 executeSingleQuery

### 前置条件
- Gate 0 完成（可编译）

### Done 标准
- 发送不含 transaction_id 的 batch 请求，所有查询独立执行并返回各自结果
- 某条查询失败，stopOnError=true 时后续查询标记 skipped
- 无连接泄漏（`go test -race` 通过）

### 信心
**70%** — 逻辑清晰，风险点在于连接泄漏防护。每条查询独立获取/释放连接的模式在 database/sql 中是标准做法
