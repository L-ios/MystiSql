## Why

根据README.MD的设计，MystiSql需要提供JDBC驱动给Java应用程序使用，让开发者能够在DataGrip、DBeaver等IDE工具中直接连接K8s集群中的数据库实例。**JDBC是Java标准接口，必须用Java实现**，当前项目中用Go实现的JDBC（`internal/jdbc/`）是错误的实现方向，Java程序无法使用Go代码。

本次变更将删除Go实现的JDBC代码，改用Java实现标准的JDBC驱动，基于MystiSql提供的RESTful API和WebSocket API进行封装，最终提供独立的JAR文件供Java生态使用。

## What Changes

### 删除

- **BREAKING** 删除Go实现的JDBC模块 `internal/jdbc/` 目录（589行代码）
  - `internal/jdbc/driver.go` - Go语言无法实现JDBC接口
  - Go的HTTP客户端无法被Java程序调用

### 新增

- 创建Java项目结构 `jdbc/`
  - Gradle构建配置（Kotlin DSL）
  - Java源码目录（`src/main/java/`, `src/test/java/`）

- 实现JDBC Driver核心接口
  - `java.sql.Driver` - 驱动入口，SPI注册
  - `java.sql.Connection` - 连接管理
  - `java.sql.Statement` - SQL执行
  - `java.sql.PreparedStatement` - 预编译语句
  - `java.sql.ResultSet` - 结果集封装
  - `java.sql.DatabaseMetaData` - 元数据查询

- 实现API集成层
  - RESTful API客户端（基于OkHttp）
  - WebSocket客户端（Phase 3，可选）
  - 认证Token管理

- 构建与发布
  - 生成可发布的JAR文件
  - 版本管理（独立于Go主项目）
  - Maven Central发布准备（可选）

### 修改

- 无需修改现有Go代码（JDBC驱动是独立的Java项目）
- 文档更新：README.MD中JDBC使用说明

## Capabilities

### New Capabilities

- `java-jdbc-driver`: Java实现的JDBC驱动核心，实现java.sql包的标准接口
- `jdbc-api-client`: JDBC驱动与MystiSql Gateway的通信层，封装RESTful API调用
- `jdbc-prepared-statement`: PreparedStatement实现，支持参数化查询防止SQL注入
- `jdbc-metadata`: DatabaseMetaData实现，为IDE工具提供数据库元数据查询能力
- `jdbc-build-system`: Gradle构建系统，支持JAR打包和发布

### Modified Capabilities

无（这是全新的Java项目，不修改现有Go代码）

## Impact

### 代码影响

- **删除**：`internal/jdbc/` 目录及其所有文件
- **新增**：`jdbc/` Java项目目录
  - `jdbc/build.gradle.kts` - Gradle构建配置
  - `jdbc/settings.gradle.kts` - 项目设置
  - `jdbc/src/main/java/io/github/mystisql/jdbc/` - Java源码
  - `jdbc/src/test/java/io/github/mystisql/jdbc/` - 测试代码
  - `jdbc/src/main/resources/META-INF/services/java.sql.Driver` - SPI配置

### 技术栈影响

- **新增语言**：Java 8+（JDBC 4.2规范）
- **构建工具**：Gradle 7.x（Kotlin DSL）
- **依赖库**：
  - `com.squareup.okhttp3:okhttp:4.x` - HTTP客户端
  - `com.fasterxml.jackson.core:jackson-databind:2.x` - JSON序列化
  - `org.slf4j:slf4j-api:1.7.x` - 日志接口
  - `org.junit.jupiter:junit-jupiter:5.x` - 测试框架（仅测试依赖）

### 部署影响

- **新增交付物**：`mystisql-jdbc-{version}.jar`
- **发布位置**：
  - 项目releases页面
  - Maven Central（Phase 3，可选）
- **文档**：
  - `jdbc/README.md` - JDBC驱动使用文档
  - `jdbc/USAGE.md` - IDE工具配置示例（DataGrip、DBeaver）

### 兼容性影响

- **向后兼容**：Go实现的JDBC从未正式发布，删除不影响用户
- **API兼容性**：Java JDBC驱动需要MystiSql Gateway的RESTful API支持
  - 现有API端点：`/api/v1/query`, `/api/v1/exec`, `/api/v1/instances`
  - 新增API端点（Phase 2.5）：`/api/v1/metadata/*`（元数据查询）

### 依赖关系

- **依赖**：MystiSql Gateway的RESTful API（必须已部署）
- **独立**：JDBC驱动是独立的Java项目，不依赖Go运行时
- **版本管理**：JDBC驱动版本号需与Gateway版本保持兼容性说明

### 测试影响

- 需要启动MystiSql Gateway进行集成测试
- 需要测试DataGrip、DBeaver等IDE工具的兼容性
- 需要测试HikariCP、Druid等连接池的兼容性
