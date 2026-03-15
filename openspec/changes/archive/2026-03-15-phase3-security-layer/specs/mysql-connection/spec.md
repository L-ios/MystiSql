## MODIFIED Requirements

### Requirement: 建立连接

系统必须能够使用提供的凭据建立到数据库实例的连接。

#### Scenario: 成功建立 MySQL 连接

- **WHEN** 使用有效的 host、port、username、password 和 database 调用 Connect() 方法（type=mysql）
- **THEN** 系统必须成功建立到 MySQL 实例的连接
- **AND** 连接必须准备好执行查询
- **AND** 必须返回 nil 错误

#### Scenario: 成功建立 PostgreSQL 连接

- **WHEN** 使用有效的 host、port、username、password 和 database 调用 Connect() 方法（type=postgresql）
- **THEN** 系统必须成功建立到 PostgreSQL 实例的连接
- **AND** 连接必须准备好执行查询
- **AND** 必须返回 nil 错误

#### Scenario: 连接失败 - 无效凭据

- **WHEN** 使用无效的用户名或密码调用 Connect() 方法
- **THEN** 系统必须返回 ErrConnectionFailed 错误
- **AND** 错误消息必须包含实例类型和名称

#### Scenario: 连接失败 - 不支持的数据库类型

- **WHEN** 使用不支持的数据库类型（type != mysql && type != postgresql）调用 Connect() 方法
- **THEN** 系统必须返回 ErrUnsupportedDatabaseType 错误
- **AND** 错误消息必须包含支持的数据库类型列表

---

### Requirement: 执行查询

系统必须能够在连接上执行 SQL 查询并返回结果。

#### Scenario: 执行 MySQL SELECT 查询

- **WHEN** 使用有效的 SELECT 语句调用 Query(ctx, sql) 方法（MySQL 实例）
- **THEN** 系统必须执行查询
- **AND** 必须返回包含列和行的 QueryResult 对象
- **AND** 必须正确处理结果集流式传输

#### Scenario: 执行 PostgreSQL SELECT 查询

- **WHEN** 使用有效的 SELECT 语句调用 Query(ctx, sql) 方法（PostgreSQL 实例）
- **THEN** 系统必须执行查询
- **AND** 必须返回包含列和行的 QueryResult 对象
- **AND** 必须正确处理结果集流式传输

#### Scenario: 执行 INSERT/UPDATE/DELETE 查询

- **WHEN** 使用有效的 INSERT、UPDATE 或 DELETE 语句调用 Exec(ctx, sql) 方法
- **THEN** 系统必须执行查询
- **AND** 必须返回受影响的行数
- **AND** 必须返回最后插入的 ID（对于 INSERT）

#### Scenario: 执行带超时的查询

- **WHEN** 使用带有超时的 context 调用 Query() 方法
- **THEN** 如果超过超时时间，系统必须取消查询
- **AND** 必须返回 context.DeadlineExceeded 错误
- **AND** 必须清理相关资源

#### Scenario: 在已关闭的连接上执行查询

- **WHEN** 在已关闭的连接上调用 Query() 方法
- **THEN** 系统必须返回 ErrConnectionClosed 错误
- **AND** 不得尝试执行查询

#### Scenario: 执行无效的 SQL

- **WHEN** 使用无效的 SQL 语法调用 Query() 方法
- **THEN** 系统必须返回 ErrQueryFailed 错误
- **AND** 错误必须包含数据库返回的语法错误信息

---

### Requirement: 连接生命周期管理

系统必须提供基本的连接生命周期管理功能。

#### Scenario: 关闭连接

- **WHEN** 在打开的连接上调用 Close() 方法
- **THEN** 连接必须被正确关闭
- **AND** 相关资源必须被释放
- **AND** 后续操作必须返回 ErrConnectionClosed 错误

#### Scenario: 检查连接健康

- **WHEN** 在连接上调用 Ping() 方法
- **THEN** 系统必须验证连接仍然存活
- **AND** 如果连接健康，必须返回 nil

## ADDED Requirements

### Requirement: 多数据库类型路由

系统必须根据实例类型自动选择对应的数据库驱动。

#### Scenario: 识别 MySQL 实例

- **WHEN** 实例配置中 type 为 "mysql"
- **THEN** 系统必须使用 go-sql-driver/mysql 驱动建立连接

#### Scenario: 识别 PostgreSQL 实例

- **WHEN** 实例配置中 type 为 "postgresql"
- **THEN** 系统必须使用 pgx 驱动建立连接

#### Scenario: 动态驱动选择

- **WHEN** 初始化连接池时
- **THEN** 系统必须根据实例配置的 type 字段动态选择驱动
- **AND** 必须在连接池创建时记录使用的驱动类型

### Requirement: PostgreSQL 特有配置支持

系统必须支持 PostgreSQL 特有的连接配置。

#### Scenario: 配置 SSL 模式

- **WHEN** PostgreSQL 实例配置包含 sslmode 参数
- **THEN** 连接必须使用指定的 SSL 模式（disable、require、verify-ca、verify-full）

#### Scenario: 配置连接超时

- **WHEN** PostgreSQL 实例配置包含 connectTimeout 参数
- **THEN** 连接超时时间必须设置为指定值

### Requirement: PostgreSQL 错误处理

系统必须正确处理 PostgreSQL 特有的错误。

#### Scenario: 处理唯一约束冲突

- **WHEN** INSERT 违反唯一约束（PostgreSQL 实例）
- **THEN** 系统必须返回明确的错误信息
- **AND** 错误消息必须包含冲突的字段名

#### Scenario: 处理外键约束错误

- **WHEN** 操作违反外键约束（PostgreSQL 实例）
- **THEN** 系统必须返回明确的错误信息
- **AND** 错误消息必须包含相关表和外键名
