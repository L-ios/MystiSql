# MystiSql JDBC Driver API 文档

## 概述

MystiSql JDBC Driver 实现了 JDBC 4.2 规范，提供了连接MystiSql Gateway并访问K8s集群中数据库实例的能力。

**版本**: 1.0.0-SNAPSHOT  
**JDBC版本**: 4.2  
**Java版本**: 8+

## 已实现的接口

### 1. Driver 接口 (`java.sql.Driver`)

#### 已实现方法
- ✅ `acceptsURL(String url)` - 接受 `jdbc:mystisql://` 开头的URL
- ✅ `connect(String url, Properties info)` - 创建数据库连接
- ✅ `getMajorVersion()` - 返回主版本号 (1)
- ✅ `getMinorVersion()` - 返回次版本号 (0)
- ✅ `getPropertyInfo(String url, Properties info)` - 返回支持的连接属性
- ✅ `jdbcCompliant()` - 返回 false (Phase 2.5)
- ❌ `getParentLogger()` - 抛出 SQLFeatureNotSupportedException

#### URL格式
```
jdbc:mystisql://<host>:<port>/<instance-name>?<parameters>
```

**参数列表**:
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `token` | String | null | 认证令牌 |
| `timeout` | int | 30 | 查询超时（秒） |
| `ssl` | boolean | false | 是否启用HTTPS |
| `verifySsl` | boolean | true | 是否验证SSL证书 |
| `maxConnections` | int | 20 | HTTP连接池大小 |

### 2. Connection 接口 (`java.sql.Connection`)

#### 已实现方法
- ✅ `createStatement()` - 创建Statement对象
- ✅ `prepareStatement(String sql)` - 创建PreparedStatement对象
- ✅ `close()` - 关闭连接
- ✅ `isClosed()` - 检查连接是否关闭
- ✅ `getMetaData()` - 获取DatabaseMetaData
- ✅ `setAutoCommit(boolean autoCommit)` - 设置自动提交模式
- ✅ `getAutoCommit()` - 获取自动提交模式
- ✅ `commit()` - 抛出SQLException (Phase 3)
- ✅ `rollback()` - 抛出SQLException (Phase 3)
- ✅ `isValid(int timeout)` - 验证连接有效性
- ✅ `getWarnings()` - 返回null
- ✅ `clearWarnings()` - 清空警告
- ✅ `getCatalog()` - 返回null
- ✅ `setCatalog(String catalog)` - 无操作
- ✅ `isReadOnly()` - 返回false
- ✅ `setReadOnly(boolean readOnly)` - 无操作
- ✅ `getTransactionIsolation()` - 返回 TRANSACTION_READ_COMMITTED
- ✅ `setTransactionIsolation(int level)` - 无操作

#### 未实现方法 (抛出 SQLFeatureNotSupportedException)
- ❌ `prepareCall(String sql)` - CallableStatement不支持
- ❌ `createBlob()` - BLOB类型不支持
- ❌ `createClob()` - CLOB类型不支持
- ❌ `createNClob()` - NCLOB类型不支持
- ❌ `createSQLXML()` - SQLXML类型不支持
- ❌ `createArrayOf(...)` - Array类型不支持
- ❌ `createStruct(...)` - Struct类型不支持

### 3. Statement 接口 (`java.sql.Statement`)

#### 已实现方法
- ✅ `executeQuery(String sql)` - 执行SELECT查询
- ✅ `executeUpdate(String sql)` - 执行INSERT/UPDATE/DELETE
- ✅ `executeLargeUpdate(String sql)` - 执行大批量更新
- ✅ `execute(String sql)` - 通用执行方法
- ✅ `getResultSet()` - 获取结果集
- ✅ `getUpdateCount()` - 获取受影响行数
- ✅ `getGeneratedKeys()` - 返回空ResultSet
- ✅ `setQueryTimeout(int seconds)` - 设置查询超时
- ✅ `getQueryTimeout()` - 获取查询超时
- ✅ `close()` - 关闭Statement
- ✅ `isClosed()` - 检查是否关闭
- ✅ `getConnection()` - 获取关联的Connection
- ✅ `setMaxRows(int max)` - 设置最大行数
- ✅ `getMaxRows()` - 获取最大行数

#### 未实现方法 (Phase 3)
- ⏳ `addBatch(String sql)` - 批量操作
- ⏳ `executeBatch()` - 批量执行
- ⏳ `clearBatch()` - 清空批量

### 4. PreparedStatement 接口 (`java.sql.PreparedStatement`)

#### 已实现方法
- ✅ `executeQuery()` - 执行参数化查询
- ✅ `executeUpdate()` - 执行参数化更新
- ✅ `clearParameters()` - 清空参数
- ✅ `getGeneratedKeys()` - 返回空ResultSet

#### 参数设置方法
- ✅ `setString(int parameterIndex, String x)`
- ✅ `setInt(int parameterIndex, int x)`
- ✅ `setLong(int parameterIndex, long x)`
- ✅ `setDouble(int parameterIndex, double x)`
- ✅ `setFloat(int parameterIndex, float x)`
- ✅ `setBoolean(int parameterIndex, boolean x)`
- ✅ `setByte(int parameterIndex, byte x)`
- ✅ `setShort(int parameterIndex, short x)`
- ✅ `setDate(int parameterIndex, Date x)`
- ✅ `setTime(int parameterIndex, Time x)`
- ✅ `setTimestamp(int parameterIndex, Timestamp x)`
- ✅ `setBigDecimal(int parameterIndex, BigDecimal x)`
- ✅ `setBytes(int parameterIndex, byte[] x)`
- ✅ `setNull(int parameterIndex, int sqlType)`
- ✅ `setObject(int parameterIndex, Object x)`

#### 未实现方法
- ❌ `setBlob(...)` - 抛出 SQLFeatureNotSupportedException
- ❌ `setClob(...)` - 抛出 SQLFeatureNotSupportedException
- ❌ `setArray(...)` - 抛出 SQLFeatureNotSupportedException
- ⏳ `addBatch()` - Phase 3

### 5. ResultSet 接口 (`java.sql.ResultSet`)

#### 已实现方法
- ✅ `next()` - 移动到下一行
- ✅ `close()` - 关闭ResultSet
- ✅ `isClosed()` - 检查是否关闭
- ✅ `getMetaData()` - 获取ResultSetMetaData
- ✅ `findColumn(String columnLabel)` - 查找列索引
- ✅ `wasNull()` - 检查最后读取的值是否为NULL
- ✅ `isBeforeFirst()` - 检查是否在第一行之前
- ✅ `isFirst()` - 检查是否在第一行
- ✅ `isLast()` - 检查是否在最后一行
- ✅ `isAfterLast()` - 检查是否在最后一行之后

#### 数据获取方法
- ✅ `getString(int columnIndex)` / `getString(String columnLabel)`
- ✅ `getInt(int columnIndex)` / `getInt(String columnLabel)`
- ✅ `getLong(int columnIndex)` / `getLong(String columnLabel)`
- ✅ `getDouble(int columnIndex)` / `getDouble(String columnLabel)`
- ✅ `getFloat(int columnIndex)` / `getFloat(String columnLabel)`
- ✅ `getBoolean(int columnIndex)` / `getBoolean(String columnLabel)`
- ✅ `getByte(int columnIndex)` / `getByte(String columnLabel)`
- ✅ `getShort(int columnIndex)` / `getShort(String columnLabel)`
- ✅ `getDate(int columnIndex)` / `getDate(String columnLabel)`
- ✅ `getTime(int columnIndex)` / `getTime(String columnLabel)`
- ✅ `getTimestamp(int columnIndex)` / `getTimestamp(String columnLabel)`
- ✅ `getBigDecimal(int columnIndex)` / `getBigDecimal(String columnLabel)`
- ✅ `getBytes(int columnIndex)` / `getBytes(String columnLabel)`
- ✅ `getObject(int columnIndex)` / `getObject(String columnLabel)`

#### 未实现方法 (抛出 SQLFeatureNotSupportedException)
- ❌ `getBlob(...)` - BLOB类型不支持
- ❌ `getClob(...)` - CLOB类型不支持
- ❌ `getArray(...)` - Array类型不支持
- ❌ `getRef(...)` - REF类型不支持
- ❌ `getRowId(...)` - ROWID不支持
- ❌ `getSQLXML(...)` - SQLXML不支持

### 6. DatabaseMetaData 接口 (`java.sql.DatabaseMetaData`)

#### 已实现方法
- ✅ `getDatabaseProductName()` - 返回 "MystiSql Gateway"
- ✅ `getDatabaseProductVersion()` - 返回 "1.0.0"
- ✅ `getDriverName()` - 返回 "MystiSql JDBC Driver"
- ✅ `getDriverVersion()` - 返回 "1.0.0"
- ✅ `getDriverMajorVersion()` - 返回 1
- ✅ `getDriverMinorVersion()` - 返回 0
- ✅ `getUserName()` - 返回用户名
- ✅ `getURL()` - 返回连接URL
- ✅ `getConnection()` - 返回Connection对象

#### 元数据查询方法
- ✅ `getCatalogs()` - 查询所有catalog
- ✅ `getSchemas()` - 查询所有schema
- ✅ `getSchemas(String catalog, String schemaPattern)` - 按条件查询schema
- ✅ `getTables(...)` - 查询表信息
- ✅ `getColumns(...)` - 查询列信息
- ✅ `getPrimaryKeys(...)` - 查询主键信息
- ✅ `getIndexInfo(...)` - 查询索引信息

#### 支持信息方法
- ✅ `supportsSelectForUpdate()` - 返回 false
- ✅ `supportsStoredProcedures()` - 返回 false
- ✅ `supportsTransactions()` - 返回 false (Phase 2.5)
- ✅ `supportsBatchUpdates()` - 返回 false (Phase 2.5)
- ✅ `supportsGetGeneratedKeys()` - 返回 false

## 类型映射

### MySQL → JDBC 类型映射

| MySQL 类型 | JDBC 类型 | java.sql.Types 常量 |
|-----------|----------|---------------------|
| TINYINT | TINYINT | Types.TINYINT |
| SMALLINT | SMALLINT | Types.SMALLINT |
| INT, INTEGER | INTEGER | Types.INTEGER |
| BIGINT | BIGINT | Types.BIGINT |
| FLOAT | FLOAT | Types.FLOAT |
| DOUBLE | DOUBLE | Types.DOUBLE |
| DECIMAL | DECIMAL | Types.DECIMAL |
| NUMERIC | DECIMAL | Types.DECIMAL |
| CHAR | CHAR | Types.CHAR |
| VARCHAR | VARCHAR | Types.VARCHAR |
| TEXT, LONGTEXT | LONGVARCHAR | Types.LONGVARCHAR |
| DATE | DATE | Types.DATE |
| TIME | TIME | Types.TIME |
| DATETIME, TIMESTAMP | TIMESTAMP | Types.TIMESTAMP |
| BLOB, BINARY | BLOB | Types.BLOB |
| BOOLEAN, BOOL | BOOLEAN | Types.BOOLEAN |
| 其他 | OTHER | Types.OTHER |

## 错误处理

### SQLException
所有数据库错误都通过 `SQLException` 抛出，包含：
- **错误消息**: 描述性错误信息
- **SQLState**: 标准SQL状态码
- **ErrorCode**: 错误代码

### SQLFeatureNotSupportedException
未实现的功能抛出 `SQLFeatureNotSupportedException`。

## 已知限制

### Phase 2.5 限制
1. **事务管理**: 不支持 commit/rollback (Phase 3)
2. **批量操作**: 不支持 addBatch/executeBatch (Phase 3)
3. **存储过程**: 不支持 CallableStatement
4. **LOB类型**: 不支持 BLOB/CLOB/NCLOB
5. **ResultSet滚动**: 仅支持 TYPE_FORWARD_ONLY
6. **Savepoint**: 不支持保存点

### 连接池兼容性
- ✅ 支持 HikariCP
- ✅ 支持 Druid
- ✅ 支持 DBCP2
- ✅ 通过 `isValid()` 方法支持连接验证

## 性能特性

- **连接池**: 内置 OkHttp 连接池，默认20个连接
- **超时控制**: 支持连接级和语句级超时
- **类型转换**: 高效的JSON到Java类型转换
- **资源管理**: 自动关闭和资源释放

## 线程安全性

- **Connection**: 非线程安全，每个线程应使用独立连接
- **Statement**: 非线程安全
- **ResultSet**: 非线程安全
- **Driver**: 线程安全，可被多线程并发访问

## 使用示例

### 基本查询
```java
String url = "jdbc:mystisql://gateway:8080/production-mysql";
try (Connection conn = DriverManager.getConnection(url, "user", "token")) {
    try (Statement stmt = conn.createStatement()) {
        ResultSet rs = stmt.executeQuery("SELECT * FROM users");
        while (rs.next()) {
            System.out.println(rs.getString("name"));
        }
    }
}
```

### PreparedStatement
```java
String sql = "SELECT * FROM users WHERE age > ? AND status = ?";
try (PreparedStatement stmt = conn.prepareStatement(sql)) {
    stmt.setInt(1, 18);
    stmt.setString(2, "active");
    ResultSet rs = stmt.executeQuery();
    // 处理结果
}
```

### 连接池 (HikariCP)
```java
HikariConfig config = new HikariConfig();
config.setJdbcUrl("jdbc:mystisql://gateway:8080/instance");
config.setUsername("user");
config.setPassword("token");
config.setMaximumPoolSize(10);

HikariDataSource ds = new HikariDataSource(config);
try (Connection conn = ds.getConnection()) {
    // 使用连接
}
```

## 更新日志

### v1.0.0-SNAPSHOT (Phase 2.5)
- 实现核心 JDBC 4.2 接口
- 支持 SELECT/INSERT/UPDATE/DELETE 操作
- 支持 PreparedStatement 参数化查询
- 支持 DatabaseMetaData 元数据查询
- 支持连接池集成
- 支持IDE工具集成

---

**文档版本**: 1.0  
**最后更新**: 2026-03-09
