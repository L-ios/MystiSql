## Why

ClickHouse、Elasticsearch、etcd 三个驱动在 Gate 0 中被标记为 `//go:build ignore`，因为：(1) go.mod 中缺少对应依赖无法编译；(2) 代码有已知 Bug（Elasticsearch 的 Query 返回已关闭的 `io.ReadCloser`、Exec 使用错误的 Index API；etcd 的 Ping 通过写入健康键来检测导致污染数据）。需要补齐依赖、修复 Bug、添加基础测试使其达到可用状态。

## What Changes

### ClickHouse 驱动
- **补齐依赖**：`go get github.com/ClickHouse/clickhouse-go/v2`
- **移除 `//go:build ignore`**
- **验证**：当前代码从 MySQL 复制而来（95% 相同），需验证 ClickHouse 协议兼容性（HTTP 接口 vs 原生协议）
- **基础测试**：连接、Ping、简单 SELECT

### Elasticsearch 驱动
- **补齐依赖**：`go get github.com/elastic/go-elasticsearch/v8`
- **移除 `//go:build ignore`**
- **Bug 修复**：
  - Query 方法返回已关闭的 `io.ReadCloser` → 改为读取 body 后关闭，返回解析后的结果
  - Exec 方法使用错误的 Index API → 改为正确的 `_doc` 端点或 `_bulk` API
- **基础测试**：连接、Ping、索引文档、搜索

### etcd 驱动
- **补齐依赖**：`go get go.etcd.io/etcd/client/v3`
- **移除 `//go:build ignore`**
- **Bug 修复**：
  - Ping 方法通过写入健康键检测 → 改为 `Get` 一个不存在的键，成功即可（不污染数据）
- **基础测试**：连接、Ping、Get/Put/Delete

### 通用
- 三个驱动均无测试，需各添加基础测试文件
- 如果 gate5-driver-base-refactor 先完成，ClickHouse 和 etcd（如果走 database/sql）可复用 base 包

## Capabilities

### Modified Capabilities
- `directory-structure`: 3 个驱动从 build ignore 恢复为可编译

## Impact

### 受影响的代码
- `internal/connection/clickhouse/connection.go` — 移除 build ignore + 验证
- `internal/connection/elasticsearch/connection.go` — 移除 build ignore + Bug 修复
- `internal/connection/etcd/connection.go` — 移除 build ignore + Bug 修复
- `go.mod` / `go.sum` — 新增 3 个依赖
- 新增 3 个测试文件

### 前置条件
- Gate 0 完成（可编译）
- **需要实际运行环境**（ClickHouse/ES/etcd 实例）才能验证和测试。建议在 E2E 测试环境中进行

### Done 标准
- `go build ./...` 包含 3 个驱动零错误
- 每个驱动至少有 Ping + 基本操作测试
- Elasticsearch 不再返回已关闭的 ReadCloser

### 信心
**45%** — ClickHouse 代码是 MySQL 的复制品，ClickHouse 协议兼容性不确定；Elasticsearch 有严重 Bug 需要重写核心逻辑；三个驱动都需要实际后端服务才能验证。建议在 E2E 环境就绪后再启动
