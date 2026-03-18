## ADDED Requirements

### Requirement: Shaded JAR 构建

项目 SHALL 构建包含所有依赖的 shaded JAR（uber JAR）。

#### Scenario: 构建 shaded JAR

- **WHEN** 执行 `./gradlew shadowJar`
- **THEN** Gradle SHALL 生成 `build/libs/mystisql-jdbc-<version>-all.jar`
- **AND** JAR 包含所有运行时依赖（OkHttp, Jackson, SLF4J, WebSocket 客户端）
- **AND** JAR 体积不超过 5MB

#### Scenario: 依赖包重定位

- **WHEN** 构建 shaded JAR
- **THEN** 所有依赖包 SHALL 被重定位到 `io.github.mystisql.shaded.*` 命名空间
- **AND** 依赖包重定位规则：
  - `okhttp3.` → `io.github.mystisql.shaded.okhttp3.`
  - `okio.` → `io.github.mystisql.shaded.okio.`
  - `com.fasterxml.jackson.` → `io.github.mystisql.shaded.jackson.`
  - `org.java_websocket.` → `io.github.mystisql.shaded.websocket.`

#### Scenario: 排除未使用的类

- **WHEN** 构建 shaded JAR
- **THEN** Shadow 插件 SHALL 移除未使用的类
- **AND** 最小化 JAR 体积

### Requirement: 双 JAR 分发

项目 SHALL 同时提供标准 JAR 和 shaded JAR 两种分发方式。

#### Scenario: 构建两种 JAR

- **WHEN** 执行 `./gradlew build`
- **THEN** 构建产物 SHALL 包含：
  - `mystisql-jdbc-<version>.jar` - 标准 JAR（需外部依赖）
  - `mystisql-jdbc-<version>-all.jar` - Shaded JAR（零依赖）

#### Scenario: Maven 依赖使用标准 JAR

- **WHEN** 用户通过 Maven 添加依赖：
  ```xml
  <dependency>
      <groupId>io.github.mystisql</groupId>
      <artifactId>mystisql-jdbc</artifactId>
      <version>1.1.0</version>
  </dependency>
  ```
- **THEN** 下载标准 JAR
- **AND** 用户需自行管理传递依赖

#### Scenario: 手动分发使用 shaded JAR

- **WHEN** 用户从 GitHub Releases 下载 JAR
- **THEN** 默认提供 shaded JAR（`-all.jar`）
- **AND** 用户无需添加任何外部依赖

### Requirement: META-INF 保留

Shaded JAR SHALL 保留必要的 META-INF 文件。

#### Scenario: 保留 SPI 配置

- **WHEN** 构建 shaded JAR
- **THEN** `META-INF/services/java.sql.Driver` SHALL 被保留
- **AND** JDBC 驱动可被自动发现和加载

#### Scenario: 保留 MANIFEST.MF

- **WHEN** 构建 shaded JAR
- **THEN** MANIFEST.MF SHALL 包含：
  - `Implementation-Title: MystiSql JDBC Driver`
  - `Implementation-Version: <version>`
  - `Main-Class`（如果需要可执行 JAR）

### Requirement: 依赖版本信息

Shaded JAR SHALL 记录包含的依赖版本。

#### Scenario: 生成依赖报告

- **WHEN** 构建 shaded JAR
- **THEN** 构建过程 SHALL 生成依赖版本报告
- **AND** 报告包含在 JAR 的 `META-INF/DEPENDENCIES` 文件中

### Requirement: Gradle 配置兼容

构建配置 SHALL 兼容 Gradle 7.x 和 Java 11+。

#### Scenario: Gradle 7.6 构建

- **WHEN** 使用 Gradle 7.6 执行构建
- **THEN** 构建 SHALL 成功完成
- **AND** 生成有效的 JAR 文件

#### Scenario: Java 11 运行

- **WHEN** 使用 Java 11+ 运行 shaded JAR
- **THEN** 应用 SHALL 正常运行
- **AND** 无兼容性错误
