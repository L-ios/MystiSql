# java-jdbc-driver

## Purpose

TBD: Java JDBC driver for MystiSql Gateway.

## Requirements

### Requirement: JDBC Driver Registration

MystiSql JDBC驱动SHALL实现`java.sql.Driver`接口，并通过SPI机制自动注册到JDBC DriverManager。

#### Scenario: Driver auto-registration via SPI

- **WHEN** Java应用加载MystiSql驱动
- **THEN** 驱动SHALL通过`META-INF/services/java.sql.Driver`文件自动注册
- **AND** DriverManager能通过`Class.forName()`加载驱动

#### Scenario: Driver accepts MystiSql URL

- **WHEN** 应用调用`DriverManager.getConnection("jdbc:mystisql://host:port/instance")`
- **THEN** MystiSql驱动SHALL接受该URL
- **AND** `Driver.acceptsURL()`返回true
- **AND** 返回有效的Connection对象

#### Scenario: Driver rejects non-MystiSql URL

- **WHEN** 应用调用`DriverManager.getConnection("jdbc:mysql://host:port/db")`
- **THEN** MystiSql驱动SHALL通过`acceptsURL()`返回false
- **AND** 不阻止其他驱动处理该URL

### Requirement: JDBC URL Format

驱动SHALL支持标准JDBC URL格式：`jdbc:mystisql://gateway-host:port/instance-name?params`

#### Scenario: Parse basic URL

- **WHEN** 连接URL为`jdbc:mystisql://gateway.example.com:8080/production-mysql`
- **THEN** 驱动SHALL解析出：
  - Gateway host: `gateway.example.com`
  - Gateway port: `8080`
  - Instance name: `production-mysql`

#### Scenario: Parse URL with parameters

- **WHEN** 连接URL为`jdbc:mystisql://gateway:8080/instance?timeout=60&ssl=true`
- **THEN** 驱动SHALL解析所有查询参数
- **AND** 将参数应用到HTTP客户端配置

#### Scenario: Handle invalid URL format

- **WHEN** 连接URL格式错误（如缺少instance名称）
- **THEN** 驱动SHALL抛出SQLException
- **AND** 错误信息清晰说明URL格式要求

### Requirement: Driver Version Info

驱动SHALL提供正确的版本信息。

#### Scenario: Get driver version

- **WHEN** 应用调用`driver.getMajorVersion()`和`driver.getMinorVersion()`
- **THEN** 驱动SHALL返回与JAR版本号匹配的major和minor版本
- **AND** 版本号从`META-INF/MANIFEST.MF`读取

### Requirement: Connection Interface

MystiSqlConnection SHALL实现`java.sql.Connection`接口，管理到MystiSql Gateway的连接。

#### Scenario: Create statement

- **WHEN** 应用调用`connection.createStatement()`
- **THEN** 驱动SHALL返回MystiSqlStatement对象
- **AND** Statement与Connection关联

#### Scenario: Prepare statement

- **WHEN** 应用调用`connection.prepareStatement("SELECT * FROM users WHERE id = ?")`
- **THEN** 驱动SHALL返回MystiSqlPreparedStatement对象
- **AND** 预编译SQL语句

#### Scenario: Close connection

- **WHEN** 应用调用`connection.close()`
- **THEN** 驱动SHALL释放所有资源
- **AND** 调用Gateway撤销token（如果已认证）
- **AND** 后续调用SHALL抛出SQLException（"Connection closed"）

#### Scenario: Connection is closed check

- **WHEN** 应用调用`connection.isClosed()`
- **THEN** 驱动SHALL返回正确的连接状态

### Requirement: Statement Interface

MystiSqlStatement SHALL实现`java.sql.Statement`接口，执行SQL语句。

#### Scenario: Execute query

- **WHEN** 应用调用`statement.executeQuery("SELECT * FROM users")`
- **THEN** 驱动SHALL发送POST请求到Gateway的`/api/v1/query`
- **AND** 返回MystiSqlResultSet对象

#### Scenario: Execute update

- **WHEN** 应用调用`statement.executeUpdate("UPDATE users SET name='Bob' WHERE id=1")`
- **THEN** 驱动SHALL发送POST请求到Gateway的`/api/v1/exec`
- **AND** 返回受影响的行数

#### Scenario: Execute large update

- **WHEN** 应用调用`statement.executeLargeUpdate()`执行大批量更新
- **THEN** 驱动SHALL返回long类型的受影响行数
- **AND** 支持超过Integer.MAX_VALUE的行数

### Requirement: ResultSet Interface

MystiSqlResultSet SHALL实现`java.sql.ResultSet`接口，封装查询结果。

#### Scenario: Iterate result set

- **WHEN** 应用调用`resultSet.next()`
- **THEN** 驱动SHALL移动光标到下一行
- **AND** 返回true表示有更多行，false表示结束

#### Scenario: Get column values

- **WHEN** 应用调用`resultSet.getString("name")`或`resultSet.getInt(1)`
- **THEN** 驱动SHALL返回当前行对应列的值
- **AND** 支持按列名或列索引访问

#### Scenario: Handle null values

- **WHEN** 列值为NULL
- **THEN** `getString()` SHALL返回null
- **AND** `getInt()` SHALL返回0
- **AND** `wasNull()` SHALL返回true

#### Scenario: Get result set metadata

- **WHEN** 应用调用`resultSet.getMetaData()`
- **THEN** 驱动SHALL返回ResultSetMetaData对象
- **AND** 包含列名、列类型等信息

### Requirement: Connection Auto-Commit

驱动SHALL支持auto-commit模式。

#### Scenario: Default auto-commit mode

- **WHEN** 创建新Connection
- **THEN** auto-commit默认为true
- **AND** 每个SQL语句自动提交

#### Scenario: Disable auto-commit

- **WHEN** 应用调用`connection.setAutoCommit(false)`
- **THEN** 驱动SHALL记录状态
- **AND** Phase 3支持事务管理时，需要实现commit/rollback

### Requirement: Connection Validation

驱动SHALL支持连接有效性验证（用于连接池）。

#### Scenario: Validate connection with timeout

- **WHEN** 应用调用`connection.isValid(5)`
- **THEN** 驱动SHALL在5秒内验证连接有效性
- **AND** 可调用Gateway的`/health`端点或执行`SELECT 1`
- **AND** 返回true/false
