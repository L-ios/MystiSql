## 1. 构建配置更新

- [ ] 1.1 在 `jdbc/build.gradle.kts` 中添加 Shadow 插件 (`com.github.johnrengelman.shadow`)
- [ ] 1.2 在 `jdbc/build.gradle.kts` 中添加 Java-WebSocket 依赖 (`org.java-websocket:Java-WebSocket:1.5.4`)
- [ ] 1.3 配置 Shadow 插件的包重定位规则（okhttp3, okio, jackson, java_websocket）
- [ ] 1.4 配置 Shadow 插件保留 META-INF/services 和 MANIFEST.MF
- [ ] 1.5 配置 Shadow 插件最小化（移除未使用的类）
- [ ] 1.6 验证构建生成两种 JAR：标准 JAR 和 shaded JAR（`-all.jar`）

## 2. 传输层抽象

- [ ] 2.1 创建 `Transport.java` 接口，定义 `query()`, `exec()`, `isValid()`, `close()` 方法
- [ ] 2.2 重构 `RestClient.java` 实现 `Transport` 接口
- [ ] 2.3 为 `Transport` 接口编写单元测试

## 3. WebSocket 客户端实现

- [ ] 3.1 创建 `WebSocketClient.java` 实现 `Transport` 接口
- [ ] 3.2 实现 WebSocket 连接建立（支持 ws:// 和 wss://）
- [ ] 3.3 实现 Token 认证（URL 参数方式）
- [ ] 3.4 实现 JSON 消息发送和接收
- [ ] 3.5 实现请求-响应匹配（使用 requestId）
- [ ] 3.6 实现连接复用（单连接模式）
- [ ] 3.7 实现心跳保活机制（每 30s 发送 ping）
- [ ] 3.8 实现自动重连机制（最多 3 次，间隔 1s/2s/4s）
- [ ] 3.9 实现 `isValid()` 方法（发送 ping 验证连接）
- [ ] 3.10 为 WebSocketClient 编写单元测试

## 4. Connection 传输选择

- [ ] 4.1 在 `MystiSqlConnection.java` 中解析 `transport` URL 参数
- [ ] 4.2 实现传输层工厂方法，根据参数创建 RestClient 或 WebSocketClient
- [ ] 4.3 更新 `MystiSqlConnection` 使用 Transport 接口而非直接使用 RestClient
- [ ] 4.4 确保默认行为为 HTTP 传输（向后兼容）
- [ ] 4.5 为传输选择逻辑编写单元测试

## 5. 测试与验证

- [ ] 5.1 编写 WebSocket 传输集成测试（需要 Mock WebSocket 服务器）
- [ ] 5.2 编写 Shaded JAR 依赖重定位验证测试
- [ ] 5.3 手动测试 Shaded JAR 在无外部依赖环境下运行
- [ ] 5.4 手动测试 WebSocket 连接到 MystiSql Gateway
- [ ] 5.5 验证 IDE 工具（DataGrip、DBeaver）兼容性

## 6. 文档更新

- [ ] 6.1 更新 `jdbc/README.md`，说明 WebSocket 传输使用方式
- [ ] 6.2 更新 `jdbc/README.md`，说明 Shaded JAR 使用方式
- [ ] 6.3 更新 URL 参数文档，添加 `transport` 参数说明
- [ ] 6.4 更新版本兼容性表格

## 7. 发布准备

- [ ] 7.1 更新版本号至 1.1.0
- [ ] 7.2 验证 Maven 发布配置
- [ ] 7.3 准备 GitHub Release 说明（包含两种 JAR）
