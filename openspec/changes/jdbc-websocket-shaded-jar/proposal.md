## Why

当前 JDBC 驱动仅支持 HTTP REST API 通信，无法利用 MystiSql Gateway 提供的 WebSocket 实时双向通信能力。同时，JDBC 驱动需要用户手动管理 OkHttp、Jackson 等外部依赖，增加了使用复杂度。用户期望一个"开箱即用"的单文件 JAR 包，无需额外依赖即可运行。

## What Changes

- **新增 WebSocket 传输层**: 为 JDBC 驱动添加 WebSocket 客户端支持，复用 Gateway 已有的 WebSocket API
- **构建 Shaded JAR**: 使用 Gradle Shadow 插件将所有依赖打包成单个 JAR 文件（零依赖分发）
- **传输层抽象**: 重构 RestClient，引入传输层接口，支持 HTTP 和 WebSocket 两种通信方式
- **WebSocket 为默认传输**: 默认使用 WebSocket 进行通信，利用长连接提升性能
- **HTTP 作为备用/强制模式**: 可通过 `transport=http` 参数强制只使用 HTTP RESTful API

## Capabilities

### New Capabilities

- `jdbc-websocket-client`: WebSocket 客户端实现，用于 JDBC 驱动与 Gateway 的实时通信，支持连接复用、心跳保活、自动重连
- `jdbc-shaded-build`: Gradle 构建配置，生成包含所有依赖的 shaded JAR（uber JAR），用户无需添加任何外部依赖

### Modified Capabilities

- `jdbc-api-client`: 重构为支持多种传输方式的客户端抽象层，保持 REST API 客户端作为默认实现，新增 WebSocket 客户端作为可选实现

## Impact

### 代码变更

- `jdbc/build.gradle.kts`: 添加 Shadow 插件配置，新增 WebSocket 客户端依赖（Java-WebSocket 或 Tyrus）
- `jdbc/src/main/java/io/github/mystisql/jdbc/client/`:
  - 新增 `Transport.java` 传输层接口
  - 新增 `WebSocketClient.java` WebSocket 客户端实现
  - 重构 `RestClient.java` 实现 Transport 接口
- `jdbc/src/main/java/io/github/mystisql/jdbc/MystiSqlConnection.java`: 支持传输方式选择

### 依赖变更

- 新增: `org.java-websocket:Java-WebSocket:1.5.4` 或类似 WebSocket 客户端库
- Shaded JAR 将包含: OkHttp, Jackson, SLF4J, WebSocket 客户端

### 兼容性

- 完全向后兼容，默认行为不变（HTTP 传输）
- 新增 URL 参数 `transport=ws` 启用 WebSocket
