## Why

当前存在多个安全漏洞：Token 生成端点无鉴权（任何人可生成任意角色 JWT）、RBAC 中间件信任客户端 Header（已在 Gate 1 修复）、CORS 全开、WebSocket CheckOrigin 允许任意来源、Token Info 通过 GET 参数泄露 token、DSN 明文密码、Token 黑名单无过期清理。同时 JDBC Client 在 WebSocket 模式下 DatabaseMetaData 和 PreparedStatement 因 `getRestClient()` 返回 null 而 NPE 崩溃。WebUI 登录流程依赖无认证的 Token 端点，需要适配。

## What Changes

### 服务端安全
- **Token 生成端点要求 admin 认证**：`POST /api/v1/auth/token` 需 admin 角色 JWT。首次部署通过 CLI 生成 bootstrap admin token（`mystisql auth bootstrap`）
- **Token Blacklist TTL 清理 + 文件持久化**：每分钟扫描清理过期条目，JSON 文件持久化，重启后自动加载
- **Token Info GET→POST BREAKING**：`GET /auth/token/info?token=xxx` → `POST /auth/token/info`，token 通过 JSON body 传输
- **CORS 可配置白名单**：从配置文件读取允许的 Origin 列表
- **WebSocket CheckOrigin 可配置 + 消息大小限制**（默认 1MB）
- **WebSocket 查询路径集成 SQL 验证器**：WS handler 执行前调用 ValidatorService
- **WebSocket lastActivity 更新**：每次收到消息更新时间戳
- **DSN 密码保护**：MySQL/PG/MSSQL 驱动改用 Config 对象，密码不出现在 DSN 字符串中
- **安全响应头**：X-Content-Type-Options、X-Frame-Options、X-XSS-Protection
- **context.Background 消除**：engine.go 和 websocket_handlers.go 热路径改用调用方 context

### JDBC Client 适配
- **DatabaseMetaData + PreparedStatement NPE 修复**：8 处 `getRestClient()` → `getTransport()`，Transport 接口扩展 `executeMetadataQuery`
- **SQLState 映射更新**：新增 `SQL_BLOCKED`/`VALIDATION_FAILED` 错误码

### WebUI 适配
- **Login 页适配**：添加管理员 Token 登录入口，使用 admin token 换取 session token
- **getTokenInfo GET→POST**
- **types.ts AuditStats/AuditLogsResponse 去重**

## Capabilities

### New Capabilities
- `security-hardening`: Token 端点鉴权 + bootstrap token + CORS 白名单 + WS 安全 + DSN 保护 + 安全头

### Modified Capabilities
- `token-auth`: Token 端点 admin 认证 + bootstrap token + 黑名单 TTL 持久化 + Token Info POST BREAKING
- `sql-validator`: WS 路径集成现有验证器
- `websocket-support`: CheckOrigin 可配置 + 消息大小限制 + lastActivity
- `rest-api`: 安全响应头中间件
- `java-jdbc-driver`: Transport 接口统一，WS NPE 修复
- `webui-interface`: 登录适配 + getTokenInfo POST + types 去重

## Impact

### 受影响的代码
**Go**: auth/, auth_handlers.go, middleware.go, websocket_handlers.go, engine.go, 6 个驱动 DSN 构建, config.yaml
**JDBC**: MystiSqlDatabaseMetaData.java (6处), MystiSqlPreparedStatement.java (2处), Transport.java, RestClient.java, WebSocketTransport.java
**WebUI**: Login.tsx, client.ts, types.ts

### 前置条件
- Gate 0 完成：可编译
- Gate 1 完成：auth middleware 已写入 role 到 context

### Done 标准
- 非管理员无法生成 token
- `GET /auth/token/info` 返回 405
- CORS 拒绝未配置的 Origin
- JDBC `transport=ws` 模式下 `getTables()` 不 NPE
- WebUI 登录流程正常

### 信心
**75%** — 安全改动有标准模式；JDBC NPE 修复路径清晰；WebUI 登录适配需要前后端协同是主要风险点
