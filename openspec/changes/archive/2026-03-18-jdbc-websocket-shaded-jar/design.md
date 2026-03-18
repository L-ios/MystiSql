## Context

MystiSql JDBC 驱动目前仅通过 HTTP REST API 与 Gateway 通信，使用 OkHttp 作为 HTTP 客户端。Gateway 已实现 WebSocket 支持（Phase 3），但 JDBC 驱动端尚未集成。同时，当前构建产物为标准 JAR，用户需手动添加 OkHttp、Jackson 等依赖，增加了使用门槛。

### 当前状态

- **通信方式**: 仅 HTTP REST API
- **依赖管理**: 用户需手动添加外部依赖
- **构建工具**: Gradle 7.6 + Java 11
- **当前依赖**: OkHttp 4.12.0, Jackson 2.16.1, SLF4J 1.7.36

### 约束条件

- 必须保持与 JDBC 4.2 规范兼容
- 必须向后兼容现有用户代码（默认行为不变）
- Shaded JAR 体积需控制在合理范围（< 5MB）

## Goals / Non-Goals

**Goals:**

1. 为 JDBC 驱动添加 WebSocket 传输层，支持实时双向通信
2. 构建包含所有依赖的 shaded JAR，实现"零依赖"分发
3. 重构客户端代码，引入传输层抽象，支持 HTTP 和 WebSocket 切换
4. 保持完全向后兼容，现有代码无需修改

**Non-Goals:**

1. 不修改 Gateway 端 WebSocket 实现
2. 不支持 WebSocket 事务管理（Phase 3 后续功能）
3. 不实现 WebSocket 批量操作（Phase 3 后续功能）
4. 不支持多 WebSocket 连接池（单连接复用足够）

## Decisions

### D1: WebSocket 客户端库选择

**决策**: 使用 `org.java-websocket:Java-WebSocket:1.5.4`

**理由**:
- 轻量级，无额外依赖（仅 100KB）
- 成熟稳定，社区活跃
- 与 OkHttp 风格一致，易于集成
- 支持 SSL/TLS、代理等特性

**备选方案**:
- Tyrus (GlassFish): 依赖复杂，不适合独立 JAR
- Netty: 过于重量级，引入不必要的依赖
- OkHttp WebSocket: 需要维护两个连接（HTTP + WS），复杂度高

### D2: 传输层抽象设计

**决策**: 引入 `Transport` 接口，`RestClient` 和 `WebSocketClient` 分别实现

```java
public interface Transport extends AutoCloseable {
    QueryResult query(QueryRequest request) throws SQLException;
    ExecResult exec(QueryRequest request) throws SQLException;
    boolean isValid(int timeout);
    void close();
}
```

**理由**:
- 解耦传输实现与 JDBC 逻辑
- 便于测试（可 mock Transport）
- 未来可扩展其他传输方式（如 gRPC）

### D3: Shaded JAR 构建策略

**决策**: 使用 Gradle Shadow 插件，重定位依赖包

**重定位规则**:
```
okhttp3. → io.github.mystisql.shaded.okhttp3.
okio. → io.github.mystisql.shaded.okio.
com.fasterxml.jackson. → io.github.mystisql.shaded.jackson.
org.java_websocket. → io.github.mystisql.shaded.websocket.
```

**理由**:
- 避免与用户应用中的依赖冲突
- 保持 API 兼容性
- 最小化 JAR 体积（排除未使用的类）

### D4: 传输方式选择机制

**决策**: 通过 URL 参数 `transport` 选择，默认 HTTP

```
jdbc:mystisql://host:8080/instance                    # HTTP（默认）
jdbc:mystisql://host:8080/instance?transport=ws      # WebSocket
jdbc:mystisql://host:8080/instance?transport=http    # 显式 HTTP
```

**理由**:
- 简单直观，用户无需修改代码即可切换
- 向后兼容（默认 HTTP）
- 便于调试和故障排查

### D5: WebSocket 连接管理

**决策**: 单连接复用 + 自动重连

- 首次查询时建立连接，后续复用
- 连接断开时自动重连（最多 3 次，间隔 1s/2s/4s）
- 空闲超时发送心跳（每 30s 发送 ping）

**理由**:
- 单连接足够满足大多数场景
- 自动重连提升用户体验
- 心跳保活符合 Gateway 配置

## Risks / Trade-offs

### R1: Shaded JAR 体积增大

- **风险**: Shaded JAR 体积约 2-3MB，可能影响下载和加载速度
- **缓解**: 
  - 使用 Shadow 插件的 minimize 功能移除未使用类
  - 同时提供标准 JAR 和 Shaded JAR 两种分发方式

### R2: WebSocket 连接稳定性

- **风险**: 长连接可能因网络波动断开
- **缓解**: 
  - 实现自动重连机制
  - 重连失败时回退到 HTTP 传输

### R3: 依赖冲突

- **风险**: 用户应用可能使用不同版本的 OkHttp/Jackson
- **缓解**: 
  - Shaded JAR 重定位所有依赖包
  - 文档中明确说明两种 JAR 的使用场景

### R4: WebSocket 认证

- **风险**: WebSocket 连接需要 Token，Token 过期后连接失效
- **缓解**: 
  - Token 过期时自动重新连接（使用新 Token）
  - 提供 Token 刷新回调接口

## Migration Plan

### 部署步骤

1. 发布新版本 JDBC 驱动（1.1.0）
   - `mystisql-jdbc-1.1.0.jar` - 标准 JAR
   - `mystisql-jdbc-1.1.0-all.jar` - Shaded JAR
2. 更新 Maven Central 和 GitHub Releases
3. 更新文档，说明两种 JAR 的使用场景

### 回滚策略

- 如发现问题，用户可回退到 1.0.x 版本
- WebSocket 为可选功能，禁用即可回退到纯 HTTP 模式

## Open Questions

1. **WebSocket 批量操作**: 是否在 Phase 3 后续版本支持批量操作？→ 待定
2. **连接池支持**: 是否需要支持多 WebSocket 连接池？→ 当前版本不需要，单连接足够
3. **WSS 支持**: WebSocket over TLS 的证书验证策略？→ 复用 HTTP 的 SSL 配置
