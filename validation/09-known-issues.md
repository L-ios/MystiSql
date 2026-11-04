# 已知缺陷与技术债务

> 从代码探索中识别的所有已知问题，按严重程度分级。
> 修复状态和验证结果由修复人员填写。

## 1. 统计汇总

| 等级 | 数量 | 说明 |
|------|------|------|
| P0（阻塞） | 0 | 无阻塞性缺陷 |
| P1（重要） | 6 | 影响功能或可靠性，建议发布前修复 |
| P2（一般） | 18 | 边缘功能，不影响核心流程，可延后 |
| P3（建议） | 11 | 代码质量、技术债务 |
| **合计** | **35** | |

---

## 2. P1 级缺陷

### P1-001: Redis 连接层无测试文件

- **位置**：`internal/connection/redis/`
- **描述**：Redis 驱动代码 450 行，是所有驱动中代码量最大的，但没有 `connection_test.go` 测试文件。其他 8 个驱动都有对应的测试文件。
- **影响**：回归风险高。任何 Redis 驱动的改动都无法通过自动化测试验证。
- **建议修复方案**：补充 `connection_test.go`，覆盖以下场景：
  - Connect/Ping/Close 基本生命周期
  - SET/GET 基本读写
  - HSET/HGETALL 哈希操作
  - 命令解析边界（不支持命令的错误处理）
- **优先级建议**：发布前修复
- **修复状态**：✅ 已修复（overnight 分支 65b6550：miniredis 测试 8/8 通过）  **验证结果**：✅

### P1-002: RBAC 仅内存存储

- **位置**：`internal/service/rbac/rbac.go`
- **描述**：角色和权限数据使用包级全局变量（`map[string]*Role`）存储，无持久化机制。
- **影响**：服务重启后所有角色/权限/用户角色分配丢失。
- **建议修复方案**：添加文件或数据库持久化（YAML/JSON/SQLite）。
- **优先级建议**：发布前修复或明确标注为 MVP 限制
- **修复状态**：⏸ 待 Owner 决策（存储选型：YAML/JSON/SQLite） | **验证结果**：|

### P1-003: Transaction 隔离级别语法不兼容

- **位置**：`internal/service/transaction/manager.go` 的 `setIsolationLevel()` 方法
- **描述**：使用 MySQL 语法（`SET SESSION TRANSACTION ISOLATION LEVEL ...`）设置隔离级别。PostgreSQL 和 Oracle 使用不同的语法。
- **影响**：PostgreSQL/Oracle 实例的事务隔离级别设置可能失败或静默忽略。
- **建议修复方案**：根据连接类型生成数据库特定的 SQL。
- **优先级建议**：发布前修复或限制隔离级别仅对 MySQL 生效
- **修复状态**：✅ 已修复（4a660c9：按数据库类型分派隔离级别语法）  **验证结果**：✅

### P1-004: JDBC PreparedStatement bug

- **位置**：`jdbc/src/main/java/io/github/mystisql/jdbc/MystiSqlPreparedStatement.java`
- **描述**：`setUpdateCount(int updateCount)` 和 `setLastInsertId(long lastInsertId)` 方法体为空，不会将值传递给父类。
- **影响**：`Statement.getUpdateCount()` 可能始终返回 -1 或默认值，而非实际影响的行数。
- **建议修复方案**：在方法体中调用 `super.setUpdateCount(updateCount)` 和 `super.setLastInsertId(lastInsertId)`。
- **优先级建议**：发布前修复
- **修复状态**：✅ 已修复（95de9b6：父类 protected setter 真实赋值）  **验证结果**：✅

### P1-005: CLI auth revoke 与 API 不匹配

- **位置**：`internal/cli/auth_cmd.go` 的 `authRevokeCmd`
- **描述**：CLI 发送 `DELETE /api/v1/auth/token` 请求，仅将 token 放在 Authorization header 中，但服务端 `RevokeToken` handler 期望 JSON body `{token: "xxx"}`。
- **影响**：撤销 token 命令可能返回 400 Bad Request。
- **建议修复方案**：CLI 端在请求 body 中发送 `{"token": "xxx"}`。
- **优先级建议**：发布前修复
- **修复状态**：✅ 已修复（1fb804a：DELETE body 发送 {"token":"xxx"}）  **验证结果**：✅

### P1-006: CLI auth info 参数传递问题

- **位置**：`internal/cli/auth_cmd.go` 的 `authInfoCmd`
- **描述**：CLI 调用 `GET /api/v1/auth/token/info`，既未在 body 中发送 token，也未在 query 参数中发送 token。服务端支持 `?token=` 参数但 CLI 未使用。
- **影响**：查看 token 信息命令可能返回 400。
- **建议修复方案**：CLI 端使用 `?token=xxx` 参数或在 body 中发送 token。
- **优先级建议**：发布前修复
- **修复状态**：✅ 已修复（1fb804a：使用 ?token= 查询参数）  **验证结果**：✅

---

## 3. P2 级缺陷

### P2-001: HA 路由未注册

- **位置**：`internal/api/rest/ha_handlers.go` + `internal/api/rest/server.go`
- **描述**：`ha_handlers.go` 定义了完整的 `RegisterHARoutes()` 函数（12 个 HA 端点），但 `server.go` 的 `setupRoutes()` 从未调用此函数。
- **影响**：所有 HA 功能（故障转移、拓扑管理）通过 API 不可用。
- **建议修复方案**：在 `server.go` 中添加 `haHandlers` 字段，在 `setupRoutes()` 中调用 `RegisterHARoutes()`。
- **修复状态**：✅ 已修复（1a18d7a：HA 路由接入 setupRoutes）  **验证结果**：✅

### P2-002: ValidateQuery 路由未注册

- **位置**：`internal/api/rest/validator_handlers.go`
- **描述**：`ValidatorHandlers.ValidateQuery` 方法已实现，但未在 `setupRoutes()` 中注册路由。
- **影响**：无法通过 API 独立验证 SQL 而不执行。
- **建议修复方案**：添加 `validatorGroup.POST("/validate", ...)`.
- **修复状态**：✅ 已修复（1a18d7a：POST /validator/validate）  **验证结果**：✅

### P2-003: Cache 服务不存在

- **位置**：`internal/service/cache/` （不存在）
- **描述**：项目中无任何缓存服务实现。文档中提到的 cache 模块实际不存在。
- **影响**：频繁查询无缓存，性能可能受限。
- **建议修复方案**：设计并实现查询结果缓存服务。
- **修复状态**：⏸ 待 Owner 决策（新模块设计） | **验证结果**：|

### P2-004: Audit AuditStats 未填充

- **位置**：`internal/service/audit/types.go`
- **描述**：`AuditStats` 类型已定义（含 TotalQueries, SuccessCount, ErrorCount, AvgExecutionTime, TopUsers, TopInstances 等字段），但无聚合逻辑填充此类型。
- **影响**：审计统计 API (`GET /api/v1/audit/stats`) 可能返回空数据或默认零值。
- **建议修复方案**：在 AuditService 中添加聚合逻辑，定期统计审计日志。
- **修复状态**：✅ 无需修复（calculateStats 聚合已完整实现，含 Top10，文档描述过时）  **验证结果**：✅

### P2-005: Batch Timeout 未执行

- **位置**：`internal/service/batch/service.go`
- **描述**：`BatchConfig.Timeout` 字段已定义（默认 5m），但在 `ExecuteBatch()` 中未设置 `context.WithTimeout`。
- **影响**：批量操作可能无限执行，超时配置无效。
- **建议修复方案**：在 `ExecuteBatch` 开头设置 `ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)`.
- **修复状态**：✅ 已修复（67af707：context.WithTimeout 生效）  **验证结果**：✅

### P2-006: Batch 并行模式忽略 StopOnError

- **位置**：`internal/service/batch/service.go`
- **描述**：`EnableParallel=true` 时，所有查询通过 goroutine 并行启动，`StopOnError` 无法中断已启动的 goroutine。
- **影响**：并行批量操作中失败不会停止其他查询。
- **建议修复方案**：使用 `errgroup` + context 取消实现可中断的并行执行。
- **修复状态**：✅ 已修复（67af707：并行模式 StopOnError 取消）  **验证结果**：✅

### P2-007: Batch 无审计日志

- **位置**：`internal/service/batch/service.go`
- **描述**：批量操作绕过审计服务，每条 SQL 不被单独记录。
- **影响**：通过批量操作执行的 SQL 不在审计日志中可见。
- **建议修复方案**：在批量执行循环中对每条 SQL 调用审计服务。
- **修复状态**：✅ 已修复（c31bda9：BatchService.SetAuditService 每条 SQL 记审计）  **验证结果**：✅

### P2-008: Query parser TODO

- **位置**：`internal/service/query/parser.go:148`
- **描述**：注释检测代码存在但结果被丢弃（`_ = strings.Contains(...)`）。
- **影响**：SQL 解析器无法正确识别注释中的 SQL 类型，可能误判。
- **建议修复方案**：使用 comment detection 结果影响 SQL 类型判断。
- **修复状态**：✅ 已修复（0b7f74c：Validate 拒绝含注释 SQL）  **验证结果**：✅

### P2-009: WebUI Dashboard 在线时长硬编码为 0

- **位置**：`web/src/pages/Dashboard.tsx`
- **描述**：仪表盘"在线时长"统计卡片始终显示 0。
- **影响**：用户体验不佳，无法反映实际运行时间。
- **建议修复方案**：从后端 API 获取实际启动时间，计算在线时长。
- **修复状态**：✅ 已修复（c9d9a73：/health uptime + Dashboard 展示）  **验证结果**：✅

### P2-010: Rate Limiting 中间件缺失

- **位置**：`internal/api/middleware/`
- **描述**：项目中无速率限制中间件。`internal/api/middleware/` 目录下只有 `auth.go` 和 `benchmark_test.go`。
- **影响**：API 无速率限制，易被暴力破解或 DDoS 攻击。
- **建议修复方案**：添加基于令牌桶或滑动窗口的限流中间件。
- **修复状态**：✅ 已修复（d35daa6：滑动窗口限流中间件，尚未全局启用）  **验证结果**：✅

### P2-011: 两个 K8s discoverer 接口不统一

- **位置**：`internal/discovery/k8s/` vs `internal/discovery/dynamic/`
- **描述**：`k8s.Discoverer` 实现了 `discovery.InstanceDiscoverer` 接口，但 `dynamic.K8sDynamicDiscoverer` 未实现此接口。
- **影响**：两个 K8s 发现器无法统一管理，增加使用复杂度。
- **建议修复方案**：统一接口或合并为一个实现。
- **修复状态**：✅ 已修复（e95a20c：补 Name/Discover/Watch 适配 + 接口断言）  **验证结果**：✅

### P2-012: ClickHouse 重复 base 逻辑

- **位置**：`internal/connection/clickhouse/connection.go` (169 行)
- **描述**：ClickHouse 驱动未嵌入 `base.SQLConnection`，而是自建了 Query/Exec/Ping/Close，与 base 层逻辑几乎完全相同。
- **影响**：代码重复，维护成本高。
- **建议修复方案**：重构为嵌入 `base.SQLConnection`，仅实现 `Connect()`。
- **修复状态**：✅ 已修复（3f65f63：嵌入 base.SQLConnection，169→48 行）  **验证结果**：✅

### P2-013: 4 个数据库驱动未注册到 serve 命令

- **位置**：`internal/cli/serve.go`
- **描述**：ClickHouse/MSSQL/Elasticsearch/Etcd 驱动代码完整且可工作，但 `serve` 命令仅注册了 MySQL/PostgreSQL/Oracle/Redis/SQLite 五个驱动。
- **影响**：这四种数据库无法通过标准 `serve` 命令使用。
- **建议修复方案**：在 `serve.go` 中补充注册代码。
- **修复状态**：✅ 已修复（3dc7b5e：4 个驱动注册）  **验证结果**：✅

### P2-014: E2E 测试目录为空

- **位置**：`test/e2e/`
- **描述**：E2E 测试目录存在但无测试文件。Makefile 中定义了 `e2e-setup/e2e-test/e2e-teardown` 目标。
- **影响**：无端到端测试覆盖，完整的用户流程无法自动验证。
- **建议修复方案**：编写关键流程的 E2E 测试（启动服务 -> 认证 -> 查询 -> 验证）。
- **修复状态**：✅ 无需修复（E2E 已迁移到 e2e-test/{backend,frontend}，Makefile 目标已更新；test/e2e 为残留空目录，可删除）  **验证结果**：✅

### P2-015: WebUI 无单元测试

- **位置**：`web/`
- **描述**：WebUI 项目无 `__tests__` 目录或 `*.test.ts`/`*.test.tsx` 文件。
- **影响**：前端代码变更无自动化测试保护。
- **建议修复方案**：为关键组件（API client, auth store）添加 Vitest 单元测试。
- **修复状态**：⏸ 待 Owner 决策（Vitest 框架引入） | **验证结果**：|

### P2-016: WebUI 无 E2E 测试

- **位置**：`web/playwright.config.ts` -> `../e2e-test/frontend/` (不存在)
- **描述**：Playwright E2E 测试配置完整（5 个浏览器项目），但测试文件目录 `e2e-test/frontend/` 不存在。
- **影响**：WebUI 完整用户流程无法自动验证。
- **建议修复方案**：编写关键页面的 E2E 测试（Login -> Query -> Instances -> AuditLogs）。
- **修复状态**：✅ 无需修复（e2e-test/frontend 已存在，含 5 个 Playwright spec 与 test-helpers，文档描述过时）  **验证结果**：✅

### P2-017: Audit 文件权限硬编码

- **位置**：`internal/service/audit/writer.go`
- **描述**：审计日志文件权限硬编码为 0644。
- **影响**：日志文件对系统所有用户可读，可能不符合安全审计要求。
- **建议修复方案**：添加可配置的文件权限，默认 0600。
- **修复状态**：✅ 已修复（73cfec9：默认 0600 可配置）  **验证结果**：✅

### P2-018: WebUI 无用户/Token 管理页面

- **位置**：`web/`
- **描述**：WebUI 无用户管理和 Token 管理页面。管理员无法通过 WebUI 创建/撤销 Token 或管理用户角色。
- **影响**：Token 和 RBAC 管理只能通过 API 或 CLI 进行。
- **建议修复方案**：添加 Token 管理和用户角色管理页面。
- **修复状态**：⏸ 待 Owner 决策（UI 设计） | **验证结果**：|

---

## 4. P3 级缺陷

### P3-001: Validator AST 仅支持 MySQL 语法

- **位置**：`internal/service/validator/ast_validator.go`
- **描述**：AST 验证器使用 `xwb1989/sqlparser`，仅支持 MySQL 语法。PostgreSQL/Oracle 特有语法（如 `RETURNING`、`LIMIT ... OFFSET`）可能导致解析失败。
- **影响**：PostgreSQL/Oracle 的复杂 SQL 可能被误判为高风险。已有 fallback 机制（降级到 tokenizer 模式）。
- **建议修复方案**：引入多方言 SQL 解析器或扩展 AST 规则。
- **修复状态**：⏸ 待 Owner 决策（多方言解析器选型） | **验证结果**：|

### P3-002: Auth generateTokenID 忽略 crypto/rand 错误

- **位置**：`internal/service/auth/token.go`
- **描述**：`generateTokenID()` 调用 `crypto/rand.Read()` 但忽略了返回的错误。
- **影响**：极端情况下（熵源不可用）Token ID 可能不唯一。
- **建议修复方案**：检查错误并返回。
- **修复状态**：✅ 已修复（a281b2f：显式错误处理）  **验证结果**：✅

### P3-003: 无 TLS/HTTPS 配置

- **位置**：全局
- **描述**：代码中未找到 TLS/HTTPS 相关配置。REST 服务器始终以 HTTP 启动。
- **影响**：生产环境必须依赖外部反向代理（Nginx/Caddy）提供 HTTPS。
- **建议修复方案**：在配置中添加 TLS 证书路径和密钥路径。
- **修复状态**：✅ 已修复（05d8fc1：ServerConfig certFile/keyFile + ListenAndServeTLS）  **验证结果**：✅

### P3-004: 无 Content-Security-Policy header

- **位置**：`internal/api/rest/middleware.go`
- **描述**：SecurityHeadersMiddleware 未设置 Content-Security-Policy 头。
- **影响**：浏览器不限制资源加载来源。
- **建议修复方案**：添加 CSP header。
- **修复状态**：✅ 已修复（dcfa889：CSP header）  **验证结果**：✅

### P3-005: Audit 无日志级别过滤

- **位置**：`internal/service/audit/service.go`
- **描述**：所有查询无论类型、敏感度、执行时间都同等记录。
- **影响**：高频查询可能产生大量日志，存储压力大。
- **建议修复方案**：添加日志级别配置（如仅记录 DDL/DELETE/DROP）。
- **修复状态**：✅ 已修复（e8aa19f+010c8f2：audit.logLevel 四级过滤）  **验证结果**：✅

### P3-006: Auth 无 refresh token 机制

- **位置**：`internal/service/auth/`
- **描述**：仅有 access token，无 refresh token 机制。Token 过期后必须重新获取。
- **影响**：长连接应用需要频繁重新认证。
- **建议修复方案**：实现 refresh token 对。
- **修复状态**：⏸ 待 Owner 决策（refresh token 架构） | **验证结果**：|

### P3-007: Auth 无 token audience/issuer claims

- **位置**：`internal/service/auth/token.go`
- **描述**：JWT 中未设置 `aud`（audience）和 `iss`（issuer）claims。
- **影响**：无法区分不同服务的 Token。
- **建议修复方案**：在配置中添加 audience 和 issuer 字段。
- **修复状态**：✅ 已修复（e5f74eb：SetIssuerAudience 可选 iss/aud，验证强制校验）  **验证结果**：✅

### P3-008: Elasticsearch Close 不完整

- **位置**：`internal/connection/elasticsearch/connection.go`
- **描述**：`Close()` 方法仅设置 `c.client = nil`，未调用底层 HTTP transport 的 `Close()` 方法。
- **影响**：HTTP 连接可能未被正确释放。
- **建议修复方案**：调用 `c.client.Transport.(*http.Transport).Close()`。
- **修复状态**：✅ 已修复（96fa07e：Close 释放 transport）  **验证结果**：✅

### P3-009: etcd Ping 有副作用

- **位置**：`internal/connection/etcd/connection.go`
- **描述**：`Ping()` 方法通过执行 `GET health-check-key` 来检查健康，如果 key 不存在会产生错误日志。
- **影响**：每次心跳都会产生一次无意义的 GET 操作。
- **建议修复方案**：使用 etcd 原生的 KV 端点健康检查或 `client.Close()` + `client.Connect()` 方式。
- **修复状态**：✅ 已修复（1ed68e2：改用 etcd 原生 Status 维护 API）  **验证结果**：✅

### P3-010: Query pool config 硬编码

- **位置**：`internal/service/query/engine.go` 的 `getConnectionPool()` 方法
- **描述**：连接池参数（MaxConnections=10, MinConnections=2 等）在代码中硬编码，不可通过外部配置修改。
- **影响**：无法根据实际负载调整连接池大小。
- **建议修复方案**：从全局配置文件读取连接池参数。
- **修复状态**：✅ 已修复（fd83d5d：Engine.SetPoolConfig + serve 注入全局配置）  **验证结果**：✅

### P3-011: 无 Discovery Consul/etcd/DNS 发现

- **位置**：`internal/discovery/`
- **描述**：仅有 static/k8s/dynamic/multicluster 四种发现器，无 Consul、etcd 或 DNS-based 服务发现。
- **影响**：非 Kubernetes 环境只能使用静态配置。
- **建议修复方案**：根据需求添加 Consul 或 DNS 发现器。
- **修复状态**：⏸ 待 Owner 决策（Consul/DNS 需求确认） | **验证结果**：|

---

## 5. 修复优先级建议

### 发布前必须修复（P1）

| 编号 | 问题 | 预估工作量 |
|------|------|-----------|
| P1-001 | Redis 无测试 | 中（补充测试） |
| P1-002 | RBAC 无持久化 | 中（文件持久化） |
| P1-003 | Transaction 语法不兼容 | 中（数据库类型判断） |
| P1-004 | JDBC PreparedStatement bug | 小（2 行修复） |
| P1-005 | CLI auth revoke 不匹配 | 小（参数调整） |
| P1-006 | CLI auth info 不匹配 | 小（参数调整） |

### 建议在下一迭代修复（P2）

优先级排序：
1. P2-010 Rate Limiting（安全优先）
2. P2-013 驱动注册（功能完善）
3. P2-001 HA 路由注册（功能完善）
4. P2-005 Batch Timeout（正确性）
5. P2-007 Batch 审计（安全审计）
6. 其余 P2 项按需处理

### 可延后处理（P3）

所有 P3 项不阻塞发布，可按需排期。
