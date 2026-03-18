# MystiSql JDBC Driver

JDBC driver for MystiSql - 透明访问K8s集群中的数据库实例

## 概述

MystiSql JDBC Driver是MystiSql项目的Java客户端驱动，允许Java应用程序通过标准JDBC接口访问K8s集群中的数据库实例（MySQL、PostgreSQL、Oracle、Redis）。

### 特性

- ✅ 实现JDBC 4.2规范（Java 8+）
- ✅ 支持 HTTP 和 WebSocket 两种传输方式
- ✅ PreparedStatement支持（防SQL注入）
- ✅ 连接池兼容（HikariCP、Druid）
- ✅ IDE工具支持（DataGrip、DBeaver）
- ✅ 元数据查询支持（DatabaseMetaData）
- ✅ Shaded JAR 零依赖分发

## 快速开始

### 环境要求

- Java 8 或更高版本
- MystiSql Gateway服务已部署

### 安装

从[Releases](https://github.com/mystisql/mystisql/releases)下载最新版本的JAR文件，添加到项目classpath。

### Maven

```xml
<dependency>
    <groupId>io.github.mystisql</groupId>
    <artifactId>mystisql-jdbc</artifactId>
    <version>1.1.0</version>
</dependency>
```

### Gradle

```groovy
implementation 'io.github.mystisql:mystisql-jdbc:1.1.0'
```

### 基本使用

```java
import java.sql.*;

public class Example {
    public static void main(String[] args) {
        String url = "jdbc:mystisql://gateway.example.com:8080/production-mysql";
        String user = "your-username";
        String password = "your-token";
        
        try (Connection conn = DriverManager.getConnection(url, user, password)) {
            // 执行查询
            String sql = "SELECT id, name, email FROM users WHERE age > ?";
            try (PreparedStatement stmt = conn.prepareStatement(sql)) {
                stmt.setInt(1, 18);
                ResultSet rs = stmt.executeQuery();
                
                while (rs.next()) {
                    int id = rs.getInt("id");
                    String name = rs.getString("name");
                    String email = rs.getString("email");
                    System.out.printf("ID: %d, Name: %s, Email: %s%n", id, name, email);
                }
            }
        } catch (SQLException e) {
            e.printStackTrace();
        }
    }
}
```

## JDBC URL格式

```
jdbc:mystisql://<gateway-host>:<gateway-port>/<instance-name>?<parameters>
```

### 参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `timeout` | 查询超时（秒） | 30 |
| `ssl` | 是否启用HTTPS/WSS | false |
| `verifySsl` | 是否验证SSL证书 | true |
| `token` | 认证令牌 | - |
| `maxConnections` | HTTP连接池大小 | 20 |
| `transport` | 传输方式：`http` 或 `ws` | `http` |

### 传输方式选择

驱动支持两种传输方式：

- **HTTP (默认)**: 通过 RESTful API 通信，每次请求独立，适合无状态场景
- **WebSocket (`ws`)**: 长连接通信，支持心跳保活和自动重连，适合高频查询场景

```java
// HTTP 传输（默认）
String url = "jdbc:mystisql://gateway:8080/instance";

// 显式指定 HTTP 传输
String url = "jdbc:mystisql://gateway:8080/instance?transport=http";

// 使用 WebSocket 传输
String url = "jdbc:mystisql://gateway:8080/instance?transport=ws";

// WebSocket over TLS
String url = "jdbc:mystisql://gateway:8080/instance?transport=ws&ssl=true";
```

### 示例

```java
// 基本连接
String url = "jdbc:mystisql://gateway.example.com:8080/production-mysql";

// 带参数的连接
String url = "jdbc:mystisql://gateway.example.com:8080/test-db?timeout=60&ssl=true";

// 使用token认证
String url = "jdbc:mystisql://gateway.example.com:8080/mydb?token=abc123";
```

## 认证

### 方式1：URL参数传递token

```java
String url = "jdbc:mystisql://gateway:8080/instance?token=your-token";
Connection conn = DriverManager.getConnection(url);
```

### 方式2：password参数作为token

```java
String url = "jdbc:mystisql://gateway:8080/instance";
String user = "username";  // 可选
String password = "your-token";  // token
Connection conn = DriverManager.getConnection(url, user, password);
```

## IDE工具配置

### DataGrip

1. 打开 DataGrip
2. 创建新的Data Source
3. 选择 "Driver files" → "Add JAR" → 选择 mystisql-jdbc.jar
4. Class: `io.github.mystisql.jdbc.MystiSqlDriver`
5. URL: `jdbc:mystisql://gateway:8080/instance`
6. User: 任意值（可为空）
7. Password: 你的token

### DBeaver

1. 创建新的数据库连接
2. 选择 "All" → "JDBC" (ODBC)
3. Driver class: `io.github.mystisql.jdbc.MystiSqlDriver`
4. URL template: `jdbc:mystisql://{host}:{port}/{instance}`
5. 填写Host、Port、Instance、Token

## 连接池配置

### HikariCP

```java
HikariConfig config = new HikariConfig();
config.setJdbcUrl("jdbc:mystisql://gateway:8080/instance");
config.setUsername("user");
config.setPassword("your-token");
config.setMaximumPoolSize(10);
config.setConnectionTimeout(30000);

HikariDataSource ds = new HikariDataSource(config);
```

## 构建

```bash
cd jdbc
./gradlew build
```

生成的JAR文件：
- `build/libs/mystisql-jdbc-1.1.0.jar` - 标准JAR（需外部依赖）
- `build/libs/mystisql-jdbc-1.1.0-all.jar` - Shaded JAR（零依赖，包含所有依赖）

### Shaded JAR 使用

Shaded JAR（`-all.jar`）包含所有运行时依赖，无需额外添加任何依赖：

```bash
# 直接运行
java -cp mystisql-jdbc-1.1.0-all.jar com.example.YourApp
```

### 标准 JAR 依赖

使用标准 JAR 时，需要添加以下依赖：

```xml
<dependencies>
    <dependency>
        <groupId>com.squareup.okhttp3</groupId>
        <artifactId>okhttp</artifactId>
        <version>4.12.0</version>
    </dependency>
    <dependency>
        <groupId>org.java-websocket</groupId>
        <artifactId>Java-WebSocket</artifactId>
        <version>1.5.4</version>
    </dependency>
    <dependency>
        <groupId>com.fasterxml.jackson.core</groupId>
        <artifactId>jackson-databind</artifactId>
        <version>2.16.1</version>
    </dependency>
</dependencies>
```

## 版本兼容性

| JDBC Driver版本 | MystiSql Gateway版本 | Java版本 | JDBC版本 | 特性 |
|----------------|---------------------|----------|----------|------|
| 1.0.x          | 1.0.x               | 8+       | 4.2      | HTTP传输 |
| 1.1.x          | 1.0.x+              | 11+      | 4.2      | HTTP/WebSocket传输, Shaded JAR |

## 已知限制

### 1.1.x (当前版本)

- ✅ SELECT、INSERT、UPDATE、DELETE操作
- ✅ PreparedStatement（参数化查询）
- ✅ 元数据查询（tables, columns, primary-keys, indexes）
- ✅ WebSocket传输（长连接、心跳、自动重连）
- ✅ Shaded JAR 零依赖分发
- ❌ 事务管理（commit/rollback） - 计划中
- ❌ 批量操作（addBatch/executeBatch） - 计划中

## 故障排查

### 连接失败

1. 检查Gateway地址和端口是否正确
2. 检查token是否有效
3. 检查网络连通性（`telnet gateway 8080`）
4. 查看Gateway日志

### SQL执行失败

1. 检查SQL语法是否正确
2. 检查表名、列名是否存在
3. 检查用户权限
4. 查看SQLException的错误信息和SQLState

## 许可证

Apache License 2.0

## 联系方式

- GitHub Issues: https://github.com/mystisql/mystisql/issues
- Email: mystisql@example.com
