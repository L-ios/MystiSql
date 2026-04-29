## Why

当前 SQL 验证器使用正则表达式，可被 SQL 注释、子查询、多语句等方式绕过。核心模块（query engine、rbac、health、blacklist）零测试覆盖，回归风险高。这两项改进可以独立于安全加固进行，且 SQL Parser 需要先做 POC 验证。

## What Changes

### SQL Validator AST 增强（POC 前置）
- **前置条件**：先用项目实际 SQL 模式跑通 `github.com/xwb1989/sqlparser` POC
- **POC 验证内容**：中文表名、子查询、多语句、PostgreSQL 特有语法（RETURNING/ON CONFLICT）、SET 语句
- **POC 通过后**：新增 `ast_validator.go` 作为可选验证层，通过 `validator.useParser` 配置开关
- **不替换现有正则验证器**：AST 作为增强层叠加，默认关闭，验证通过后默认开启
- **PG 语法降级**：解析失败时降级到正则 + warn 日志，不阻塞查询
- **白名单/黑名单升级**：从字符串/正则匹配改为基于 AST 的表名级别匹配

### 核心模块测试补齐
- `service/query/engine.go` — 查询路由、超时、验证器集成、审计集成
- `service/rbac/` — 权限检查、角色分配、中间件
- `service/auth/blacklist.go` — TTL 清理、持久化
- `service/health/` — 健康检查、事件通知、状态缓存
- JDBC Transport 接口测试 — WS 模式 DatabaseMetaData 场景

## Capabilities

### New Capabilities
- `sql-parser-validator`: 基于 AST 的 SQL 验证增强层（POC 前置，可选开关）

### Modified Capabilities
- `sql-validator`: AST 增强层叠加
- `sql-whitelist-blacklist`: AST 表名级别匹配

## Impact

### 受影响的代码
- `internal/service/validator/` — 新增 ast_validator.go
- `internal/service/query/` — 新增 engine_test.go
- `internal/service/rbac/` — 新增 rbac_test.go
- `internal/service/auth/blacklist.go` — 新增 blacklist_test.go
- `internal/service/health/` — 新增 monitor_test.go
- `jdbc/src/test/` — Transport 接口测试

### 前置条件
- Gate 0 完成：可编译可测试

### Done 标准
- POC 报告：sqlparser 对项目 SQL 模式的覆盖度 ≥ 90%
- AST 验证器能拦截正则无法检测的绕过（如 `DELETE FROM t -- WHERE 1=1`）
- 所有新增测试通过
- PG 语法降级路径不阻塞查询

### 信心
**65%** — SQL Parser 依赖 POC 结果；降级策略降低风险；测试补齐是确定性工作但可能暴露未知 Bug
