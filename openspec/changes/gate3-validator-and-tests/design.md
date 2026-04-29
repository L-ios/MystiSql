## Context

MystiSql 当前的 SQL 验证层由三套独立的字符串/正则解析器组成：

1. **`validator/sql_tokenizer.go`**（282 行）— 安全验证用，能力最强：剥离注释、字符串字面量、多语句分割、词边界关键字检测
2. **`query/parser.go`**（387 行）— 查询引擎用：识别语句类型、提取表名、估算结果集，纯 `strings.HasPrefix` 实现
3. **`router/sql_parser.go`**（132 行）— 读写路由用：识别 SELECT/INSERT/UPDATE/DELETE + 事务边界检测

**核心问题**：三套解析器互不复用，`sql_tokenizer` 的安全分析能力没有反哺到 `query/parser` 和 `router/sql_parser`。`query/parser.Validate()` 直接用 `strings.Contains` 检测危险操作，可被 SQL 注释绕过（如 `DELETE FROM t -- WHERE 1=1`）。

**测试现状**：proposal 中列出的测试补齐目标大部分已完成：
- `service/query/engine_test.go` ✅ 已有（28 个测试）
- `service/rbac/rbac_test.go` ✅ 已有（13 个测试）
- `service/auth/blacklist_test.go` ✅ 已有（24 个测试）
- `service/health/health_test.go` ⚠️ 已有（5 个），但 `EnhancedHealthChecker`（313 行）未覆盖
- JDBC Transport 测试 ❌ 无

前置条件：Gate 0（build-fix）和 Gate 1（integration-fix）已完成，`go build ./...` 通过。

## Goals / Non-Goals

**Goals:**

- 验证 `github.com/xwb1989/sqlparser` 对项目实际 SQL 模式的覆盖度
- 新增 `ast_validator.go` 作为可选验证增强层，通过配置开关启用
- PG 语法解析失败时降级到现有 tokenizer 验证，不阻塞查询
- 补充 `EnhancedHealthChecker` 单元测试
- 补充 JDBC WebSocket Transport 接口测试

**Non-Goals:**

- 不替换现有 `sql_tokenizer.go`（AST 作为增强层叠加，不破坏已有功能）
- 不修改 `query/parser.go` 和 `router/sql_parser.go`（本次不改它们的实现）
- 不做白名单/黑名单的 AST 表名级别匹配升级（留后续 change）
- 不做 `discovery/k8s`、`connection/redis`、`ha_handlers` 等其他模块的测试补齐（不在本 change 范围内）

## Decisions

### D1: SQL Parser 选型 — `github.com/xwb1989/sqlparser`

**选择**：使用 `github.com/xwb1989/sqlparser`（vitess 的 SQL 解析器独立提取版）。

**替代方案**：
- (a) `github.com/pingcap/parser` — TiDB 的解析器，功能更全但依赖重，引入 ~200 个间接依赖
- (b) `github.com/auxten/postgresql-parser` — 仅支持 PostgreSQL，不适合多数据库项目
- (c) 自研简易 AST — 投入产出比差，`sql_tokenizer` 已经是自研的上限

**理由**：`xwb1989/sqlparser` 是纯 Go 实现，零 CGO 依赖，轻量（~10 个间接依赖），支持 MySQL 语法（项目主要场景），社区成熟。PG 特有语法（RETURNING/ON CONFLICT）不支持时走降级路径。

**POC 验证项**（必须先通过）：
- 中文表名：`SELECT * FROM 用户表`
- 子查询：`DELETE FROM t WHERE id IN (SELECT id FROM t_backup)`
- 多语句：`SELECT 1; DELETE FROM t`
- SET 语句：`SET NAMES utf8`
- 基本增删改查

### D2: AST 验证器架构 — 可选增强层

**选择**：新增 `ASTValidator` 结构体，实现与 `SQLValidatorImpl` 相同的验证接口，作为可选中间层。

**架构**：
```
请求 → Service.Validate()
         ↓
    ASTValidator（如果启用）
      ├─ 解析成功 → 用 AST 信息做精确验证
      └─ 解析失败 → 降级到 SQLValidatorImpl（tokenizer）
         ↓
    SQLValidatorImpl（tokenizer，始终作为兜底）
```

**配置**：
```yaml
validator:
  enabled: true
  useParser: true  # 新增，默认 false
```

**接口设计**：
```go
type ASTValidator struct {
    fallback *SQLValidatorImpl  // 降级目标
    logger   *zap.Logger
}

func (a *ASTValidator) Validate(ctx context.Context, instance, query string) (*ValidationResult, error)
func (a *ASTValidator) GetQueryType(query string) string
```

**理由**：叠加而非替换，保证向后兼容。`useParser: false` 时行为完全不变。

### D3: EnhancedHealthChecker 测试策略

**选择**：为 `enhanced_checker.go`（313 行）编写单元测试，mock 依赖的连接池和健康检查接口。

**测试覆盖**：
- 健康检查流程（正常/超时/连接失败）
- 状态缓存（首次检查 vs 缓存命中）
- 事件通知（状态变化触发回调）
- 并发安全（多个 goroutine 同时检查）

### D4: JDBC WebSocket Transport 测试策略

**选择**：为 JDBC 驱动的 WebSocket Transport 层编写接口级测试。

**范围**：
- `WebSocketTransport.java` 连接建立/断开
- 消息协议匹配（requestId、success 字段）
- 心跳 pong 处理
- 错误重连

**理由**：JDBC 测试不需要运行中的 Gateway，可以 mock WebSocket 连接。

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| `xwb1989/sqlparser` 不支持 PG 特有语法（RETURNING/ON CONFLICT） | 降级到 tokenizer 验证 + warn 日志，不阻塞查询 |
| POC 可能发现 sqlparser 覆盖度不足 90% | 调整方案：仅对 MySQL 语法启用 AST，PG 保持 tokenizer |
| AST 解析有性能开销 | 基准测试对比 tokenizer vs AST 延迟；如果差异显著，增加解析结果缓存 |
| 新增外部依赖增加供应链风险 | sqlparser 是成熟库，vitess 生态，维护活跃 |
| EnhancedHealthChecker 测试可能暴露设计问题 | 如果发现 Bug，记录但不在此 change 中修复（本 change 仅补测试） |

## Open Questions

1. **POC 结果未定**：sqlparser 对中文表名、子查询的支持程度需要实际验证。如果 POC 不通过，AST 增强层方案需要调整。
2. **JDBC 测试环境**：JDBC 测试是否需要 Maven/Gradle 构建环境？当前项目是否有 Java 构建配置？
