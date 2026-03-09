## Context

### 当前状态

项目中存在错误的JDBC实现方向：
- **Go实现的JDBC**（`internal/jdbc/driver.go`，589行）- 违反JDBC是Java接口的基本事实
- Go代码无法被Java程序调用
- README.MD明确要求提供JDBC驱动给Java应用使用

### JDBC规范要求

JDBC（Java Database Connectivity）是Java标准接口，定义在`java.sql`包中：
- 必须用Java实现
- 需要实现`java.sql.Driver`等标准接口
- 通过SPI（Service Provider Interface）自动注册
- 提供JAR文件供Java应用使用

### MystiSql架构定位

```
Java Application (用户代码)
    ↓
JDBC Driver (本次实现，Java)
    ↓
MystiSql Gateway RESTful API (Go实现)
    ↓
Database Instance (MySQL/PostgreSQL等)
```

### 技术约束

- **语言约束**：必须使用Java 8+（JDBC 4.2规范）
- **依赖约束**：必须依赖MystiSql Gateway的RESTful API
- **兼容性约束**：必须兼容主流IDE工具（DataGrip、DBeaver）
- **性能约束**：HTTP通信开销需要通过连接池优化

## Goals / Non-Goals

### Goals

**Phase 2.5 - 核心功能**（本次实现）
1. 删除Go实现的错误JDBC代码
2. 实现Java JDBC驱动核心接口
   - Driver、Connection、Statement、PreparedStatement
   - ResultSet（TYPE_FORWARD_ONLY）
   - DatabaseMetaData（基础元数据查询）
3. 基于RESTful API实现查询执行
4. 支持PreparedStatement参数化查询（防SQL注入）
5. 生成可用的JAR文件
6. 兼容DataGrip、DBeaver基本功能
7. 兼容HikariCP、Druid连接池

**Phase 3 - 增强功能**（后续迭代）
1. WebSocket支持（流式结果集）
2. 事务管理（Connection.commit/rollback）
3. 批量操作（Statement.addBatch）
4. Maven Central发布

### Non-Goals

1. **本次不涉及**：
   - 修改MystiSql Gateway的Go代码（除删除`internal/jdbc/`）
   - 实现完整的JDBC 4.3规范（仅核心方法）
   - 支持所有JDBC高级特性（Array、Blob、Clob、SQLXML等）
   - WebSocket实时推送（Phase 3）
   - 性能压测和优化
   - PostgreSQL/Oracle支持（Phase 3）

2. **明确排除**：
   - 直接实现数据库协议（MySQL Protocol）- 仍通过Gateway代理
   - 服务端逻辑修改 - 仅客户端驱动
   - 连接池实现本身 - 仅提供兼容性

## Decisions

### Decision 1: 项目结构

**选择**：独立Java项目，Gradle构建

**项目布局**：
```
jdbc/
├── build.gradle.kts           # Gradle构建配置
├── settings.gradle.kts        # 项目设置
├── README.md                  # 使用文档
├── USAGE.md                   # IDE配置示例
├── src/
│   ├── main/
│   │   ├── java/
│   │   │   └── io/github/mystisql/jdbc/
│   │   │       ├── MystiSqlDriver.java           # Driver接口实现
│   │   │       ├── MystiSqlConnection.java       # Connection接口实现
│   │   │       ├── MystiSqlStatement.java        # Statement接口实现
│   │   │       ├── MystiSqlPreparedStatement.java # PreparedStatement实现
│   │   │       ├── MystiSqlResultSet.java        # ResultSet接口实现
│   │   │       ├── MystiSqlDatabaseMetaData.java # DatabaseMetaData实现
│   │   │       ├── client/
│   │   │       │   ├── RestClient.java           # HTTP客户端封装
│   │   │       │   └── model/                    # API请求/响应模型
│   │   │       └── util/
│   │   │           └── TypeConverter.java        # 类型转换工具
│   │   └── resources/
│   │       └── META-INF/services/
│   │           └── java.sql.Driver               # SPI注册文件
│   └── test/
│       └── java/
│           └── io/github/mystisql/jdbc/
│               ├── driver/
│               ├── connection/
│               └── integration/
└── gradle.properties           # 版本配置
```

**理由**：
- 独立项目便于版本管理和发布
- Gradle比Maven更灵活，适合现代Java项目
- Kotlin DSL比Groovy更类型安全

**替代方案**：
- Maven：更传统，但配置冗长
- 放在主项目中：耦合度高，不利于独立发布

### Decision 2: 技术栈选型

| 组件 | 选择 | 版本 | 理由 |
|------|------|------|------|
| Java | Java 8 | 1.8 | 最大兼容性，JDBC 4.2支持 |
| 构建工具 | Gradle | 7.x | 灵活、现代、Kotlin DSL支持 |
| HTTP客户端 | OkHttp | 4.x | 性能优秀、连接池内置、API友好 |
| JSON序列化 | Jackson | 2.x | 性能最佳、功能完整 |
| 日志 | SLF4J | 1.7.x | 标准接口，兼容主流实现 |
| 测试 | JUnit 5 | 5.x | 现代测试框架 |

**替代方案**：
- Apache HttpClient：功能强大但API复杂
- Gson：简单但性能不如Jackson
- Log4j 2：性能好但SLF4J更通用

### Decision 3: JDBC URL格式

**格式**：
```
jdbc:mystisql://<gateway-host>:<gateway-port>/<instance-name>?<params>
```

**示例**：
```
jdbc:mystisql://gateway.example.com:8080/production-mysql
jdbc:mystisql://localhost:8080/test-db?timeout=60&ssl=true
jdbc:mystisql://192.168.1.100:8080/mydb?token=abc123
```

**参数说明**：
- `timeout` - 查询超时（秒）
- `ssl` - 是否启用HTTPS
- `verifySsl` - 是否验证证书（开发环境可关闭）
- `token` - 认证令牌（可选，也可通过password传递）

**理由**：
- 遵循JDBC URL标准格式
- 与MySQL JDBC URL相似，便于理解
- 支持灵活的参数配置

### Decision 4: RESTful API集成策略

**API端点映射**：

| JDBC方法 | Gateway API | 说明 |
|----------|-------------|------|
| Statement.executeQuery() | POST /api/v1/query | 执行SELECT |
| Statement.executeUpdate() | POST /api/v1/exec | 执行INSERT/UPDATE/DELETE |
| PreparedStatement.setXxx() | 参数化请求体 | 防SQL注入 |
| DatabaseMetaData.getXxx() | GET /api/v1/metadata/* | 元数据查询 |
| Connection.close() | DELETE /api/v1/auth/token | 撤销token |

**PreparedStatement请求体**：
```json
{
  "instance": "production-mysql",
  "query": "SELECT * FROM users WHERE name = ? AND age > ?",
  "parameters": [
    {"type": "VARCHAR", "value": "Alice"},
    {"type": "INTEGER", "value": 18}
  ]
}
```

**理由**：
- 参数化查询在后端Gateway处理，更安全
- 类型信息传递确保正确绑定
- 与JDBC规范对应

### Decision 5: 类型映射策略

**MySQL → JDBC类型映射**：

| MySQL类型 | JDBC类型 | java.sql.Types常量 |
|-----------|----------|-------------------|
| TINYINT | TINYINT | -6 |
| SMALLINT | SMALLINT | 5 |
| INT | INTEGER | 4 |
| BIGINT | BIGINT | -5 |
| FLOAT | FLOAT | 6 |
| DOUBLE | DOUBLE | 8 |
| DECIMAL | DECIMAL | 3 |
| CHAR | CHAR | 1 |
| VARCHAR | VARCHAR | 12 |
| TEXT | LONGVARCHAR | -1 |
| DATE | DATE | 91 |
| DATETIME | TIMESTAMP | 93 |
| TIMESTAMP | TIMESTAMP | 93 |
| BLOB | BLOB | 2004 |

**实现位置**：
- Gateway侧：负责从MySQL元数据读取类型
- JDBC驱动侧：负责将Gateway返回的类型字符串映射到java.sql.Types

**理由**：
- 遵循JDBC规范
- IDE工具依赖正确的类型映射

### Decision 6: 认证集成策略

**方案**：Token认证

**流程**：
1. JDBC URL或Connection参数中提供token
2. 驱动在所有HTTP请求头中添加：`Authorization: Bearer {token}`
3. Connection.close()时调用Gateway撤销token

**URL示例**：
```java
// 方式1：URL参数
String url = "jdbc:mystisql://gateway:8080/instance?token=abc123";

// 方式2：password参数
Connection conn = DriverManager.getConnection(
    "jdbc:mystisql://gateway:8080/instance",
    "username",
    "token-value"  // 将password作为token使用
);
```

**理由**：
- 与Gateway现有的Token认证机制一致
- 简单且安全（HTTPS传输）
- 兼容JDBC标准的username/password参数

### Decision 7: 错误处理策略

**SQLException映射**：

| Gateway错误码 | SQLState | 说明 |
|--------------|----------|------|
| TABLE_NOT_FOUND | "42S02" | 表不存在 |
| SYNTAX_ERROR | "42000" | SQL语法错误 |
| CONNECTION_FAILED | "08001" | 连接失败 |
| TIMEOUT | "HYT00" | 超时 |
| INTERNAL_ERROR | "HY000" | 内部错误 |

**实现**：
```java
public class MystiSqlException extends SQLException {
    private final String errorCode;
    
    public MystiSqlException(String reason, String sqlState, String errorCode) {
        super(reason, sqlState);
        this.errorCode = errorCode;
    }
}
```

**理由**：
- 遵循SQLState标准
- 便于应用层错误处理
- 保留原始错误码供调试

### Decision 8: 连接池兼容性策略

**HikariCP兼容性要点**：
1. 实现`Connection.isValid(int timeout)`方法
   - 调用Gateway的`/health`端点
   - 超时时间内返回true/false

2. 实现`Connection.close()`幂等性
   - 多次调用不抛异常

3. 支持连接测试查询
   - `Connection.createStatement().executeQuery("SELECT 1")`

**验证方法**：
```java
// HikariCP配置
HikariConfig config = new HikariConfig();
config.setJdbcUrl("jdbc:mystisql://gateway:8080/instance");
config.setUsername("user");
config.setPassword("token");
config.setConnectionTimeout(30000);
config.setMaximumPoolSize(10);

HikariDataSource ds = new HikariDataSource(config);
```

**理由**：
- HikariCP是最流行的连接池
- 兼容性测试确保生产可用

## Risks / Trade-offs

### Risk 1: 性能开销（HTTP vs 原生协议）

**风险**：HTTP通信比原生MySQL协议慢

**缓解方案**：
- OkHttp连接池复用TCP连接
- 启用HTTP Keep-Alive
- Gateway与数据库同集群部署，减少网络延迟
- Phase 3可考虑WebSocket长连接优化

**Trade-off**：牺牲部分性能换取架构简洁性

### Risk 2: 元数据API不完整

**风险**：Gateway未实现所有元数据API，导致IDE工具功能受限

**缓解方案**：
- Phase 2.5优先实现核心元数据（tables, columns, primary-keys, indexes）
- JDBC驱动提供fallback：直接查询`information_schema`
- 在文档中明确说明功能边界

**Trade-off**：部分功能延迟到Phase 3

### Risk 3: PreparedStatement实现复杂度

**风险**：参数化查询的类型绑定和SQL注入防护需要仔细实现

**缓解方案**：
- 严格遵循JDBC规范
- 充分的单元测试和集成测试
- 参考MySQL JDBC驱动的实现

**Trade-off**：开发时间较长，但安全性高

### Risk 4: IDE工具兼容性

**风险**：不同IDE工具对JDBC的使用方式不同，可能遇到兼容性问题

**缓解方案**：
- 重点测试DataGrip和DBeaver（市场占有率最高）
- 提供详细的IDE配置文档
- 建立Issue跟踪兼容性问题

**Trade-off**：无法保证100%兼容所有工具

### Risk 5: 版本管理复杂度

**风险**：JDBC驱动独立版本，与Gateway版本可能不匹配

**缓解方案**：
- 在文档中明确版本兼容性矩阵
- JDBC驱动版本号遵循语义化版本（SemVer）
- 发布Release Notes说明兼容性

**Trade-off**：增加维护成本

## Migration Plan

### Phase 1: 清理（Day 1）

1. **删除Go JDBC代码**
   ```bash
   rm -rf internal/jdbc/
   ```

2. **提交变更**
   ```bash
   git add .
   git commit -m "Remove incorrect Go JDBC implementation"
   ```

### Phase 2: Java项目搭建（Day 1-2）

1. **创建项目结构**
   ```bash
   mkdir -p jdbc/src/{main,test}/java/io/github/mystisql/jdbc
   mkdir -p jdbc/src/main/resources/META-INF/services
   ```

2. **配置Gradle**
   - `build.gradle.kts` - 依赖和构建配置
   - `settings.gradle.kts` - 项目名称

3. **配置SPI**
   ```
   # src/main/resources/META-INF/services/java.sql.Driver
   io.github.mystisql.jdbc.MystiSqlDriver
   ```

### Phase 3: 核心实现（Day 2-7）

1. **Day 2-3**: Driver和Connection
2. **Day 4-5**: Statement和ResultSet
3. **Day 6**: PreparedStatement
4. **Day 7**: DatabaseMetaData

### Phase 4: 测试与验证（Day 8-10）

1. **单元测试**（Day 8）
2. **集成测试**（Day 9）
3. **IDE工具测试**（Day 10）

### Phase 5: 构建与发布（Day 11）

1. **构建JAR**
   ```bash
   cd jdbc
   ./gradlew build
   ```

2. **发布到项目releases**
   - 上传`build/libs/mystisql-jdbc-{version}.jar`
   - 编写Release Notes

### Rollback Plan

**如果Java JDBC驱动有严重问题**：

1. 回退到之前的Git提交（Go JDBC已删除，无法回退）
2. 发布Hotfix版本修复问题
3. 在文档中标注问题版本不可用

**注意**：Go JDBC实现从未正式发布，删除不会影响用户

## Open Questions

1. **元数据查询的SQL方言适配**：
   - Phase 2.5仅支持MySQL，PostgreSQL/Oracle如何支持？
   - **建议**：Phase 3扩展，在Gateway侧实现DatabaseDialect接口

2. **WebSocket的优先级**：
   - 是否在Phase 2.5实现基础WebSocket支持？
   - **建议**：Phase 2.5暂不实现，Phase 3再考虑

3. **连接池验证查询**：
   - `isValid()`方法应该调用哪个Gateway端点？
   - **建议**：调用`/api/v1/instances/{name}/health`，或执行`SELECT 1`

4. **JAR文件命名**：
   - `mystisql-jdbc-{version}.jar` 还是 `mystisql-jdbc-driver-{version}.jar`？
   - **建议**：`mystisql-jdbc-{version}.jar`，更简洁

5. **Maven Central发布**：
   - Phase 2.5是否需要发布到Maven Central？
   - **建议**：Phase 3再考虑，Phase 2.5通过GitHub Releases分发
