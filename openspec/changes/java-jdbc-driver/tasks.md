## 1. 清理Go JDBC实现

- [x] 1.1 删除Go实现的JDBC模块
  - 删除`internal/jdbc/`目录
  - 确认没有其他代码依赖此模块
  - 运行测试确保删除后项目编译正常

- [x] 1.2 提交删除变更
  - Git commit说明删除原因
  - 更新相关文档（如有引用）

## 2. Java项目初始化

- [x] 2.1 创建Java项目目录结构
  - 创建`jdbc/`根目录
  - 创建标准Maven/Gradle目录：`src/main/java/`, `src/test/java/`, `src/main/resources/`
  - 创建包结构：`io.github.mystisql.jdbc`

- [x] 2.2 配置Gradle构建文件
  - 创建`build.gradle.kts`（Kotlin DSL）
  - 配置Java 8兼容性
  - 添加依赖：OkHttp 4.x, Jackson 2.x, SLF4J 1.7.x
  - 添加测试依赖：JUnit 5, Mockito
  - 配置JAR打包任务（包含SPI文件）

- [x] 2.3 创建Gradle wrapper
  - 运行`gradle wrapper`生成wrapper文件
  - 配置`gradle.properties`版本信息

- [x] 2.4 配置SPI注册
  - 创建`src/main/resources/META-INF/services/java.sql.Driver`
  - 写入`io.github.mystisql.jdbc.MystiSqlDriver`
  - 配置`META-INF/MANIFEST.MF`（Implementation-Version等）

- [x] 2.5 创建基础文档
  - 创建`jdbc/README.md`（项目说明）
  - 创建`jdbc/USAGE.md`（使用示例）

## 3. JDBC Driver接口实现

- [x] 3.1 实现MystiSqlDriver类
  - 实现`java.sql.Driver`接口的所有必需方法
  - 实现`acceptsURL()`方法，识别`jdbc:mystisql://`前缀
  - 实现`connect()`方法，返回MystiSqlConnection
  - 实现`getPropertyInfo()`方法，返回支持的连接属性
  - 实现`getMajorVersion()`和`getMinorVersion()`
  - 实现`jdbcCompliant()`方法（Phase 2.5返回false）

- [x] 3.2 实现URL解析逻辑
  - 解析`jdbc:mystisql://host:port/instance?params`格式
  - 提取gateway host, port, instance name
  - 解析查询参数（timeout, ssl, token等）
  - 错误处理：无效URL格式抛出SQLException

- [x] 3.3 编写Driver单元测试
  - 测试URL解析正确性
  - 测试SPI自动注册
  - 测试acceptsURL逻辑
  - 测试无效URL错误处理

## 4. RESTful API客户端实现

- [x] 4.1 实现RestClient类
  - 封装OkHttp客户端
  - 配置连接池（默认最多20个连接）
  - 配置超时（默认30秒）
  - 支持HTTPS和证书验证配置

- [x] 4.2 实现认证集成
  - 支持从URL参数提取token
  - 支持将password作为token使用
  - 在所有HTTP请求头中添加`Authorization: Bearer {token}`
  - 连接关闭时调用Gateway撤销token

- [x] 4.3 实现API请求模型类
  - 创建`QueryRequest`类（instance, query, parameters）
  - 创建`ExecRequest`类（instance, sql）
  - 创建`Parameter`类（type, value）

- [x] 4.4 实现API响应模型类
  - 创建`QueryResponse`类（columns, rows, rowCount）
  - 创建`ExecResponse`类（rowsAffected, lastInsertId）
  - 创建`ErrorResponse`类（error, code）

- [x] 4.5 实现JSON序列化/反序列化
  - 使用Jackson进行JSON转换
  - 处理日期时间类型的序列化（ISO 8601格式）
  - 处理NULL值

- [x] 4.6 实现错误码到SQLState的映射
  - 创建`ErrorCodeMapper`类
  - 映射Gateway错误码到标准SQLState
  - 处理未知错误码（使用HY000）

- [x] 4.7 编写RestClient单元测试
  - 使用MockWebServer模拟Gateway
  - 测试各种HTTP请求场景
  - 测试认证token传递
  - 测试错误响应处理

## 5. JDBC Connection接口实现

- [ ] 5.1 实现MystiSqlConnection类
  - 实现`java.sql.Connection`接口的必需方法
  - 维护连接状态（closed, autoCommit等）
  - 持有RestClient实例
  - 持有实例名称、认证token等信息

- [ ] 5.2 实现createStatement()方法
  - 返回MystiSqlStatement对象
  - 关联Connection和Statement

- [ ] 5.3 实现prepareStatement()方法
  - 返回MystiSqlPreparedStatement对象
  - 解析SQL中的`?`占位符

- [ ] 5.4 实现close()方法
  - 标记连接为closed
  - 调用Gateway撤销token（如已认证）
  - 释放RestClient资源

- [ ] 5.5 实现isClosed()方法
  - 返回连接关闭状态

- [ ] 5.6 实现isValid()方法
  - 调用Gateway的`/health`或`/api/v1/instances/{name}/health`
  - 或执行`SELECT 1`查询
  - 返回true/false表示连接有效性

- [ ] 5.7 实现getMetaData()方法
  - 返回MystiSqlDatabaseMetaData对象

- [ ] 5.8 实现auto-commit相关方法
  - `getAutoCommit()`返回当前状态
  - `setAutoCommit(boolean)`设置状态（Phase 2.5仅记录状态）
  - Phase 3实现真正的事务管理

- [ ] 5.9 实现Connection占位方法
  - `commit()`, `rollback()`抛出SQLException（Phase 3实现）
  - `createBlob()`, `createClob()`抛出SQLFeatureNotSupportedException

- [ ] 5.10 编写Connection单元测试
  - 测试Statement和PreparedStatement创建
  - 测试连接关闭和状态检查
  - 测试isValid()逻辑
  - 测试异常处理

## 6. JDBC Statement接口实现

- [ ] 6.1 实现MystiSqlStatement类
  - 实现`java.sql.Statement`接口的必需方法
  - 持有Connection实例
  - 持有当前ResultSet

- [ ] 6.2 实现executeQuery()方法
  - 调用RestClient发送`POST /api/v1/query`
  - 解析响应创建ResultSet
  - 返回ResultSet对象

- [ ] 6.3 实现executeUpdate()方法
  - 调用RestClient发送`POST /api/v1/exec`
  - 返回受影响行数（int）

- [ ] 6.4 实现executeLargeUpdate()方法
  - 返回受影响行数（long）
  - 支持大批量操作

- [ ] 6.5 实现execute()方法
  - 执行SQL，判断是查询还是更新
  - 返回true表示有ResultSet，false表示更新

- [ ] 6.6 实现getResultSet()方法
  - 返回当前ResultSet

- [ ] 6.7 实现getUpdateCount()方法
  - 返回受影响行数

- [ ] 6.8 实现setQueryTimeout()方法
  - 设置Statement级别的超时
  - 覆盖Connection级别的超时设置

- [ ] 6.9 实现close()方法
  - 关闭当前ResultSet（如有）
  - 释放资源

- [ ] 6.10 实现Statement占位方法
  - `addBatch()`, `executeBatch()`抛出SQLException（Phase 3）
  - 其他不常用方法抛出SQLFeatureNotSupportedException

- [ ] 6.11 编写Statement单元测试
  - 测试SELECT查询
  - 测试INSERT/UPDATE/DELETE
  - 测试超时设置
  - 测试错误处理

## 7. JDBC ResultSet接口实现

- [ ] 7.1 实现MystiSqlResultSet类
  - 实现`java.sql.ResultSet`接口的必需方法
  - 持有列信息和行数据
  - 维护当前行索引

- [ ] 7.2 实现next()方法
  - 移动光标到下一行
  - 返回true表示有更多行，false表示结束

- [ ] 7.3 实现getXxx()方法（按列名）
  - `getString(String columnName)`
  - `getInt(String columnName)`
  - `getLong(String columnName)`
  - `getDouble(String columnName)`
  - `getBoolean(String columnName)`
  - `getDate(String columnName)`
  - `getTimestamp(String columnName)`
  - `getBigDecimal(String columnName)`

- [ ] 7.4 实现getXxx()方法（按列索引）
  - 所有getXxx方法的重载版本（int columnIndex）
  - 列索引从1开始

- [ ] 7.5 实现类型转换逻辑
  - JSON值到Java类型的转换
  - 处理类型不匹配情况
  - 处理NULL值

- [ ] 7.6 实现wasNull()方法
  - 检查最后读取的值是否为NULL

- [ ] 7.7 实现getMetaData()方法
  - 返回ResultSetMetaData对象

- [ ] 7.8 实现findColumn()方法
  - 根据列名返回列索引

- [ ] 7.9 实现close()方法
  - 标记ResultSet为closed
  - 释放资源

- [ ] 7.10 实现ResultSet占位方法
  - `getArray()`, `getBlob()`, `getClob()`抛出SQLFeatureNotSupportedException

- [ ] 7.11 编写ResultSet单元测试
  - 测试光标移动
  - 测试各种类型的数据读取
  - 测试NULL值处理
  - 测试类型转换

## 8. JDBC PreparedStatement接口实现

- [ ] 8.1 实现MystiSqlPreparedStatement类
  - 实现`java.sql.PreparedStatement`接口
  - 继承Statement功能
  - 持有SQL语句和参数列表

- [ ] 8.2 实现setXxx()方法
  - `setString(int, String)`
  - `setInt(int, int)`
  - `setLong(int, long)`
  - `setDouble(int, double)`
  - `setBoolean(int, boolean)`
  - `setDate(int, Date)`
  - `setTimestamp(int, Timestamp)`
  - `setBigDecimal(int, BigDecimal)`
  - `setNull(int, int)`

- [ ] 8.3 实现参数存储逻辑
  - 参数按索引存储（索引从1开始）
  - 参数包含类型和值

- [ ] 8.4 实现executeQuery()方法
  - 构建参数化请求体
  - 调用RestClient发送`POST /api/v1/query`
  - 返回ResultSet

- [ ] 8.5 实现executeUpdate()方法
  - 构建参数化请求体
  - 调用RestClient发送`POST /api/v1/exec`
  - 返回受影响行数

- [ ] 8.6 实现clearParameters()方法
  - 清空所有已设置的参数
  - PreparedStatement可复用

- [ ] 8.7 实现getGeneratedKeys()方法
  - 返回包含lastInsertId的ResultSet
  - 用于获取自增ID

- [ ] 8.8 编写PreparedStatement单元测试
  - 测试参数设置
  - 测试参数化查询执行
  - 测试参数清除和复用
  - 测试参数索引验证

## 9. JDBC DatabaseMetaData接口实现

- [ ] 9.1 实现MystiSqlDatabaseMetaData类
  - 实现`java.sql.DatabaseMetaData`接口
  - 持有Connection实例

- [ ] 9.2 实现基本元数据方法
  - `getDatabaseProductName()`
  - `getDatabaseProductVersion()`
  - `getDriverName()`
  - `getDriverVersion()`
  - `getUserName()`
  - `getURL()`

- [ ] 9.3 实现getCatalogs()方法
  - 调用`GET /api/v1/metadata/catalogs`
  - 返回ResultSet

- [ ] 9.4 实现getSchemas()方法
  - 调用`GET /api/v1/metadata/schemas`
  - 支持catalog过滤参数
  - 返回ResultSet

- [ ] 9.5 实现getTables()方法
  - 调用`GET /api/v1/metadata/tables`
  - 支持schema、pattern、type过滤
  - 返回ResultSet包含TABLE_CAT, TABLE_SCHEM, TABLE_NAME, TABLE_TYPE, REMARKS

- [ ] 9.6 实现getColumns()方法
  - 调用`GET /api/v1/metadata/columns`
  - 支持schema、table过滤
  - 返回ResultSet包含完整列信息
  - 将MySQL类型映射到JDBC Types

- [ ] 9.7 实现getPrimaryKeys()方法
  - 调用`GET /api/v1/metadata/primary-keys`
  - 返回ResultSet包含主键信息

- [ ] 9.8 实现getIndexInfo()方法
  - 调用`GET /api/v1/metadata/indexes`
  - 返回ResultSet包含索引信息

- [ ] 9.9 实现元数据缓存
  - 缓存查询结果，TTL 5分钟
  - 按实例名称隔离缓存
  - 支持手动刷新缓存

- [ ] 9.10 编写DatabaseMetaData单元测试
  - 使用MockWebServer模拟Gateway响应
  - 测试各种元数据查询
  - 测试缓存逻辑
  - 测试类型映射

## 10. 类型转换工具实现

- [ ] 10.1 实现TypeConverter工具类
  - MySQL类型字符串到java.sql.Types的映射
  - JSON值到Java类型的转换
  - Java类型到参数类型的转换

- [ ] 10.2 编写类型映射表
  - 创建MySQL → JDBC类型映射Map
  - 处理带长度的类型（VARCHAR(255)）

- [ ] 10.3 编写TypeConverter单元测试
  - 测试所有常用类型的映射
  - 测试边界情况（未知类型）

## 11. 异常处理实现

- [ ] 11.1 创建MystiSqlException类
  - 继承SQLException
  - 添加errorCode字段
  - 提供友好的错误信息

- [ ] 11.2 实现异常工厂
  - 根据Gateway错误响应创建SQLException
  - 正确设置SQLState和errorCode

- [ ] 11.3 编写异常处理测试
  - 测试各种错误场景
  - 测试SQLState映射

## 12. 日志实现

- [ ] 12.1 配置SLF4J日志
  - 在所有关键路径添加日志
  - 使用DEBUG级别记录请求详情
  - 使用ERROR级别记录失败

- [ ] 12.2 敏感信息脱敏
  - 不记录token、password等敏感信息
  - SQL参数记录时脱敏（可配置）

## 13. 集成测试

- [ ] 13.1 搭建集成测试环境
  - 启动MystiSql Gateway（Docker或本地）
  - 准备测试数据库实例（MySQL）
  - 创建测试表和数据

- [ ] 13.2 编写端到端测试
  - 测试完整查询流程
  - 测试PreparedStatement
  - 测试元数据查询
  - 测试错误处理

- [ ] 13.3 编写连接池兼容性测试
  - 测试HikariCP集成
  - 测试Druid集成
  - 测试连接验证和回收

- [ ] 13.4 编写IDE工具兼容性测试
  - 测试DataGrip基本功能（连接、查询、元数据浏览）
  - 测试DBeaver基本功能
  - 记录兼容性问题

## 14. 文档编写

- [ ] 14.1 编写README.md
  - 项目介绍
  - 快速开始
  - 构建说明

- [ ] 14.2 编写USAGE.md
  - JDBC URL格式说明
  - DataGrip配置示例（截图）
  - DBeaver配置示例（截图）
  - 代码示例

- [ ] 14.3 编写API文档
  - 列出已实现和未实现的JDBC方法
  - 说明与标准JDBC的差异
  - 说明已知限制

## 15. 构建与发布

- [ ] 15.1 配置JAR打包
  - 配置Gradle构建任务
  - 包含所有依赖（fat jar）或仅驱动代码
  - 正确设置MANIFEST.MF

- [ ] 15.2 编写发布脚本
  - 构建JAR文件
  - 生成签名和校验和
  - 准备Release Notes

- [ ] 15.3 执行发布
  - 上传JAR到项目releases页面
  - 更新项目主README中的JDBC使用说明

## 16. 清理与优化

- [ ] 16.1 代码审查
  - 检查代码风格一致性
  - 检查异常处理完整性
  - 检查资源释放（close方法）

- [ ] 16.2 性能优化
  - 优化JSON序列化
  - 优化连接池配置
  - 优化元数据缓存

- [ ] 16.3 删除Go JDBC相关引用
  - 检查项目其他文档是否有引用
  - 更新README.MD中关于JDBC的描述

## 17. 验收测试

- [ ] 17.1 功能验收
  - 所有单元测试通过
  - 所有集成测试通过
  - DataGrip基本功能可用
  - HikariCP连接池兼容

- [ ] 17.2 文档验收
  - README和USAGE文档完整
  - 代码示例可运行

- [ ] 17.3 发布验收
  - JAR文件可正常使用
  - 版本号正确
  - Release Notes清晰
