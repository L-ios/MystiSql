## Context

### 当前状态
MystiSql Phase 1 和 Phase 2 已完成，具备以下能力：
- **服务发现**：静态配置发现、K8s API 动态发现
- **数据库连接**：MySQL 连接池管理、健康检查、自动重连
- **查询引擎**：SQL 解析、路由、超时控制、结果集限制
- **接入层**：RESTful API、CLI 工具、JDBC 驱动（基础功能）

### 约束
- Go 语言实现，遵循项目架构规范
- 兼容现有 API 接口，不破坏 Phase 1/2 的功能
- JDBC 驱动必须兼容标准 JDBC 规范和主流连接池（HikariCP）
- 性能要求：认证中间件延迟 < 1ms，审计日志异步写入

### 相关方
- **开发者**：需要在 IDE（DataGrip、DBeaver）中安全访问数据库
- **运维人员**：需要审计日志追溯操作，需要危险操作防护
- **安全团队**：要求认证、授权、审计三要素齐全

## Goals / Non-Goals

**Goals:**
1. 实现基于 JWT Token 的身份认证机制
2. 实现 SQL 执行审计日志，记录所有操作
3. 实现 SQL 安全检查，拦截危险操作（DROP、TRUNCATE 等）
4. 支持 SQL 白名单/黑名单配置
5. 为 REST API 添加认证中间件
6. 为 CLI 添加 Token 认证支持
7. 实现 WebSocket 实时交互接口
8. 支持 PostgreSQL 数据库连接
9. JDBC 驱动支持事务管理（begin、commit、rollback）
10. JDBC 驱动支持批量操作

**Non-Goals:**
- Phase 3 不实现 RBAC 权限模型（Phase 4）
- 不实现 OIDC/LDAP 集成（Phase 4）
- 不实现数据脱敏（Phase 4）
- 不实现读写分离（Phase 4）
- 不实现完整的 WebUI（Phase 5）

## Decisions

### 1. Token 认证方案：JWT vs 自定义 Token

**决策**：使用 JWT (JSON Web Token)

**理由**：
- JWT 是行业标准，有成熟的 Go 库（`golang-jwt/jwt`）
- 无需服务端存储 Session，适合无状态架构
- 支持自定义 Claims，方便扩展用户信息
- Token 自带过期时间，简化过期管理

**替代方案**：
- **Session + Cookie**：需要服务端存储，不适合多实例部署
- **自定义 Token（随机字符串）**：需要额外的存储和查询开销

**实现要点**：
- 使用 HS256 签名算法（对称加密）
- Token 包含：用户 ID、角色、过期时间
- 配置项：`auth.token.secret`（签名密钥）、`auth.token.expire`（过期时间，默认 24h）

### 2. 审计日志存储：文件 vs 数据库

**决策**：Phase 3 使用文件存储，日志格式为 JSON Lines

**理由**：
- 文件存储简单可靠，无需额外依赖
- JSON Lines 格式便于日志分析工具（ELK、Splunk）处理
- 为 Phase 4 实现数据库存储预留接口

**替代方案**：
- **数据库存储**：需要设计表结构，增加数据库依赖，Phase 3 不优先
- **Syslog**：与操作系统耦合，不便于容器化部署

**实现要点**：
- 日志文件路径：`/var/log/mystisql/audit.log`（可配置）
- 日志轮转：按天轮转，保留 30 天
- 异步写入：使用 buffered channel，避免阻塞请求
- 日志字段：timestamp, user_id, instance, query, rows_affected, execution_time, client_ip

### 3. SQL 验证实现：正则 vs SQL 解析器

**决策**：使用 SQL 解析器（复用 Phase 2 的 QueryEngine 解析能力）

**理由**：
- Phase 2 已实现 SQL 解析能力，可直接复用
- 解析器可以准确识别 SQL 语句类型（SELECT、INSERT、UPDATE、DELETE、DDL）
- 比正则表达式更可靠，不易误判

**替代方案**：
- **正则匹配**：简单但不够准确，容易误判（如字段名包含 "DROP"）

**实现要点**：
- 危险操作列表：`DROP TABLE`, `DROP DATABASE`, `TRUNCATE`, `DELETE`（无 WHERE 子句）
- 白名单/黑名单使用正则表达式匹配
- 拦截后返回 403 Forbidden，记录审计日志

### 4. PostgreSQL 驱动：pgx vs pq

**决策**：使用 pgx（`github.com/jackc/pgx/v5`）

**理由**：
- pgx 性能优于 pq，是 PostgreSQL 官方推荐的 Go 驱动
- 支持连接池（内置 pool）
- 支持预处理语句、批量操作
- 活跃维护，社区支持好

**替代方案**：
- **pq**：老牌驱动，但性能和功能不如 pgx，维护活跃度下降

**实现要点**：
- 实现 `ConnectionPool` 接口（与 MySQL 一致）
- 连接字符串格式：`postgres://user:pass@host:port/database`

### 5. WebSocket 实现：gorilla/websocket vs nhooyr/websocket

**决策**：使用 gorilla/websocket

**理由**：
- gorilla/websocket 是 Go 生态中最成熟的 WebSocket 库
- 文档完善，社区使用广泛
- 支持 TLS、压缩、子协议

**替代方案**：
- **nhooyr/websocket**：API 更现代，但社区生态不如 gorilla

**实现要点**：
- 端点：`ws://host:port/ws`
- 认证：握手时通过 URL 参数传递 Token（`?token=xxx`）
- 消息格式：JSON（`{"action": "query", "instance": "xxx", "query": "SELECT ..."}`）

### 6. JDBC 事务实现策略

**决策**：在 Gateway 服务端管理事务，通过连接 ID 关联

**理由**：
- JDBC 标准的 `Connection.setAutoCommit(false)` 需要服务端支持
- 通过连接 ID 关联事务状态，多个 SQL 请求复用同一数据库连接
- 连接池已有连接管理能力，事务状态绑定到连接对象

**实现要点**：
- 新增 `/api/v1/transaction/begin` 接口，返回 `connectionId`
- 后续 SQL 请求携带 `connectionId` 参数
- `/api/v1/transaction/commit` 和 `/api/v1/transaction/rollback` 提交或回滚
- 事务超时自动回滚（默认 5 分钟）

### 7. API 认证中间件实现

**决策**：使用 Gin 中间件，除白名单路径外全局拦截

**理由**：
- Gin 中间件机制简单易用
- 全局拦截确保所有端点都经过认证
- 白名单路径（如 `/health`）无需认证

**实现要点**：
- 中间件提取 `Authorization: Bearer <token>` 或 URL 参数 `?token=xxx`
- 验证 JWT Token，提取用户信息注入到 `gin.Context`
- 认证失败返回 401 Unauthorized
- 白名单路径：`/health`, `/api/v1/auth/login`

## Risks / Trade-offs

### 风险 1：Token 泄露导致未授权访问
- **风险**：JWT Token 一旦泄露，攻击者可以伪造身份访问数据库
- **缓解措施**：
  - 使用 HTTPS 传输，防止中间人攻击
  - Token 设置较短过期时间（默认 24h）
  - 提供 Token 撤销接口（黑名单机制）
  - 记录 Token 使用审计日志

### 风险 2：审计日志影响性能
- **风险**：大量 SQL 操作可能产生大量审计日志，影响系统性能
- **缓解措施**：
  - 异步写入日志，不阻塞请求
  - 日志文件轮转，避免单个文件过大
  - 提供配置开关，允许临时关闭审计日志

### 风险 3：SQL 验证误判
- **风险**：SQL 验证规则可能误判合法 SQL（如字段名包含 "DROP"）
- **缓解措施**：
  - 使用 SQL 解析器而非正则，提高准确性
  - 提供白名单机制，允许特定 SQL 绕过检查
  - 拦截时返回清晰的错误提示，用户可以反馈误判

### 风险 4：WebSocket 连接数限制
- **风险**：大量 WebSocket 长连接可能占用过多资源
- **缓解措施**：
  - 设置最大连接数限制（配置项：`websocket.maxConnections`）
  - 连接空闲超时自动断开（默认 10 分钟）
  - 心跳机制检测连接活性

### 风险 5：JDBC 事务并发冲突
- **风险**：多个请求使用同一个 `connectionId` 可能导致并发冲突
- **缓解措施**：
  - `connectionId` 绑定到客户端 Session，不跨客户端共享
  - 事务超时自动回滚，避免长时间占用连接
  - 连接池监控，告警连接数过多

### 风险 6：PostgreSQL 兼容性问题
- **风险**：PostgreSQL 和 MySQL 的 SQL 语法有差异，可能导致兼容性问题
- **缓解措施**：
  - SQL 解析器支持 PostgreSQL 方言
  - 提供数据库类型感知的路由
  - 文档说明支持的 SQL 语法范围

## Migration Plan

### 阶段 1：准备阶段（不影响现有功能）
1. 部署新代码（默认配置关闭认证和审计）
2. 生成初始 Token，分发给用户
3. 用户配置 Token 环境变量或配置文件

### 阶段 2：灰度阶段（部分启用认证）
1. 启用审计日志（`audit.enabled: true`）
2. 启用 SQL 验证（`validator.enabled: true`）
3. 配置白名单路径，允许部分 API 无需认证

### 阶段 3：全面启用
1. 启用全局认证（`auth.enabled: true`）
2. 所有用户必须使用 Token 访问
3. 监控认证失败率，调整配置

### 回滚策略
- 如果认证导致严重问题，设置 `auth.enabled: false` 即可关闭
- 审计日志和 SQL 验证也可以独立关闭
- JDBC 驱动增强是增量功能，不影响现有使用方式

## Open Questions

### Q1：是否需要实现 Token 刷新机制？
- **背景**：JWT Token 过期后需要重新登录
- **选项**：实现 refresh token 机制，或让用户重新获取 token
- **当前决策**：Phase 3 不实现，Phase 4 集成 OIDC/LDAP 时再考虑

### Q2：审计日志是否需要记录查询结果？
- **背景**：记录查询结果可以完整还原操作，但数据量大
- **选项**：只记录元数据（SQL、行数、时间），或记录部分结果（前 100 行）
- **当前决策**：Phase 3 只记录元数据，Phase 5 可以增加结果记录（可配置）

### Q3：WebSocket 消息格式是否需要支持二进制？
- **背景**：当前设计为 JSON 文本格式
- **选项**：支持二进制消息可以提高性能
- **当前决策**：Phase 3 只支持 JSON，未来按需扩展

### Q4：JDBC 批量操作是否需要支持不同类型的 SQL？
- **背景**：JDBC `addBatch()` 可以混合 INSERT、UPDATE、DELETE
- **选项**：支持混合批处理，或只支持同类型 SQL
- **当前决策**：Phase 3 支持混合批处理，符合 JDBC 规范

### Q5：PostgreSQL 连接池配置是否复用 MySQL 的配置？
- **背景**：PostgreSQL 和 MySQL 的连接池配置可能有差异
- **选项**：统一配置，或分别配置
- **当前决策**：统一配置（`pool.maxOpen`, `pool.maxIdle`, `pool.maxLifetime`），简化管理
