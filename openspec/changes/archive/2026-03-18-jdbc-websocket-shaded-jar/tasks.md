## 1. 构建配置更新

- [x] 1.1 在 `jdbc/build.gradle.kts` 中添加 Shadow 插件 (`com.github.johnrengelman.shadow`)
  > **已实现**: 使用 `com.gradleup.shadow` 8.3.5 版本（支持 Java 21）
- [x] 1.2 在 `jdbc/build.gradle.kts` 中添加 Java-WebSocket 依赖 (`org.java-websocket:Java-WebSocket:1.5.4`)
  > **已实现**: `jdbc/build.gradle.kts:29`
- [x] 1.3 配置 Shadow 插件的包重定位规则（okhttp3, okio, jackson, java_websocket）
  > **已实现**: 所有依赖已重定位到 `io.github.mystisql.shaded.*`
- [x] 1.4 配置 Shadow 插件保留 META-INF/services 和 MANIFEST.MF
  > **已实现**: 排除签名文件，保留服务配置
- [x] 1.5 配置 Shadow 插件最小化（移除未使用的类）
  > **已实现**: 使用 `minimize` 配置，排除必要的 WebSocket 和 Jackson 依赖
- [x] 1.6 验证构建生成两种 JAR：标准 JAR 和 shaded JAR（`-all.jar`）
  > **已实现**: 
  > - `mystisql-jdbc-1.1.0.jar` (59KB) - 标准 JAR
  > - `mystisql-jdbc-1.1.0-all.jar` (5.1MB) - Shaded JAR

## 2. 传输层抽象

- [x] 2.1 创建 `Transport.java` 接口，定义 `query()`, `exec()`, `isValid()`, `close()` 方法
  > **已实现**: `jdbc/src/main/java/io/github/mystisql/jdbc/client/Transport.java`
  > 方法名: `executeQuery`, `executeUpdate`, `healthCheck`, `close`, `getTransportType`
- [x] 2.2 重构 `RestClient.java` 实现 `Transport` 接口
  > **已实现**: `RestClient implements Transport`
- [x] 2.3 为 `Transport` 接口编写单元测试
  > **已实现**: `TransportContractTest.java` 契约测试接口

## 3. WebSocket 客户端实现

- [x] 3.1 创建 `WebSocketClient.java` 实现 `Transport` 接口
  > **已实现**: `jdbc/src/main/java/io/github/mystisql/jdbc/client/WebSocketTransport.java`
- [x] 3.2 实现 WebSocket 连接建立（支持 ws:// 和 wss://）
  > **已实现**: `WebSocketTransport.java:50-51` URL 转换
- [x] 3.3 实现 Token 认证（URL 参数方式）
  > **已实现**: `WebSocketTransport.java:52-55` Token 附加到 URL
- [x] 3.4 实现 JSON 消息发送和接收
  > **已实现**: `WebSocketTransport.java:197-218` 使用 Jackson ObjectMapper
- [x] 3.5 实现请求-响应匹配（使用 requestId）
  > **已实现**: `WebSocketTransport.java:200-202, 143-150` 使用 CompletableFuture 和 pendingRequests Map
- [x] 3.6 实现连接复用（单连接模式）
  > **已实现**: `WebSocketTransport.java:64-74` ensureConnected 方法
- [x] 3.7 实现心跳保活机制（每 30s 发送 ping）
  > **已实现**: `WebSocketTransport.java:35, 159-178` HEARTBEAT_INTERVAL = 30000ms
- [x] 3.8 实现自动重连机制（最多 3 次，间隔 1s/2s/4s）
  > **已实现**: `WebSocketTransport.java:33-34, 76-101` RECONNECT_INTERVALS = [1000, 2000, 4000]
- [x] 3.9 实现 `isValid()` 方法（发送 ping 验证连接）
  > **已实现**: `WebSocketTransport.java:267-275` healthCheck 方法
- [x] 3.10 为 WebSocketClient 编写单元测试
  > **已实现**: `WebSocketTransportTest.java`

## 4. Connection 传输选择

- [x] 4.1 在 `MystiSqlConnection.java` 中解析 `transport` URL 参数
  > **已实现**: `MystiSqlConnection.java:39, 53` transportType 参数
- [x] 4.2 实现传输层工厂方法，根据参数创建 RestClient 或 WebSocketClient
  > **已实现**: `MystiSqlConnection.java:55-59` 条件分支创建
- [x] 4.3 更新 `MystiSqlConnection` 使用 Transport 接口而非直接使用 RestClient
  > **已实现**: `MystiSqlConnection.java:25, 79-82` 使用 Transport 接口
- [x] 4.4 确保默认行为为 HTTP 传输（向后兼容）
  > **已修复**: 默认值改为 `http` (MystiSqlConnection.java:53-55)
- [x] 4.5 为传输选择逻辑编写单元测试
  > **已实现**: `MystiSqlConnectionTest.java` 添加了传输选择测试

## 5. 测试与验证

- [ ] 5.1 编写 WebSocket 传输集成测试（需要 Mock WebSocket 服务器）
- [ ] 5.2 编写 Shaded JAR 依赖重定位验证测试
- [ ] 5.3 手动测试 Shaded JAR 在无外部依赖环境下运行
- [ ] 5.4 手动测试 WebSocket 连接到 MystiSql Gateway
- [ ] 5.5 验证 IDE 工具（DataGrip、DBeaver）兼容性

## 6. 文档更新

- [x] 6.1 更新 `jdbc/README.md`，说明 WebSocket 传输使用方式
  > **已实现**: 添加了"传输方式选择"章节
- [x] 6.2 更新 `jdbc/README.md`，说明 Shaded JAR 使用方式
  > **已实现**: 添加了"Shaded JAR 使用"章节
- [x] 6.3 更新 URL 参数文档，添加 `transport` 参数说明
  > **已实现**: 参数表格已更新
- [x] 6.4 更新版本兼容性表格
  > **已实现**: 版本表格已更新至 1.1.x

## 7. 发布准备

- [x] 7.1 更新版本号至 1.1.0
  > **已实现**: `build.gradle.kts` 版本已更新为 `1.1.0`
- [ ] 7.2 验证 Maven 发布配置
- [ ] 7.3 准备 GitHub Release 说明（包含两种 JAR）
