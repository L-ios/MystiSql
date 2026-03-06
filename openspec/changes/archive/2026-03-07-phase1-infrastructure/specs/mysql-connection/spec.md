## ADDED Requirements

### Requirement: MySQL 驱动集成

系统必须集成 go-sql-driver/mysql 驱动以提供 MySQL 数据库连接能力。

#### Scenario: 驱动注册

- **WHEN** 系统初始化 MySQL 连接支持
- **THEN** go-sql-driver/mysql 驱动必须注册到 database/sql
- **AND** 连接必须使用驱动名称 "mysql"

---

### Requirement: 建立连接

系统必须能够使用提供的凭据建立到 MySQL 实例的连接。

#### Scenario: 成功建立连接

- **WHEN** 使用有效的 host、port、username、password 和 database 调用 Connect() 方法
- **THEN** 系统必须成功建立到 MySQL 实例的连接
- **AND** 连接必须准备好执行查询
- **AND** 必须返回 nil 错误

#### Scenario: 连接失败 - 无效凭据

- **WHEN** 使用无效的用户名或密码调用 Connect() 方法
- **THEN** 系统必须返回 ErrConnectionFailed 错误
- **AND** 错误必须包含底层的 MySQL 错误信息
- **AND** 不得建立连接

#### Scenario: 连接失败 - 无法访问的主机

- **WHEN** 使用无法访问的 host 或 port 调用 Connect() 方法
- **THEN** 系统必须返回 ErrConnectionFailed 错误
- **AND** 系统不得无限期挂起（必须超时）
- **AND** 必须在合理的时间内返回（默认 30 秒）

#### Scenario: 连接失败 - 数据库不存在

- **WHEN** 指定的数据库不存在
- **THEN** 系统必须返回 ErrConnectionFailed 错误
- **AND** 错误消息必须明确说明 "数据库不存在"

---

### Requirement: 执行查询

系统必须能够在已建立的 MySQL 连接上执行 SQL 查询。

#### Scenario: 执行 SELECT 查询

- **WHEN** 使用有效的 SELECT 语句调用 Query(ctx, sql) 方法
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
- **AND** 错误必须包含 MySQL 返回的语法错误信息

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
- **AND** 如果连接不健康，必须返回错误

#### Scenario: 关闭已关闭的连接

- **WHEN** 在已关闭的连接上调用 Close() 方法
- **THEN** 系统必须返回 nil（幂等操作）
- **AND** 不得产生错误

---

### Requirement: 连接配置

系统必须支持 MySQL 特定的连接参数。

#### Scenario: 配置连接超时

- **WHEN** 创建带有超时配置的连接
- **THEN** 连接必须遵守配置的超时时间
- **AND** 查询不得超过超时时长
- **AND** 默认超时时间必须是 30 秒

#### Scenario: 配置字符集

- **WHEN** 创建带有字符集参数（如 utf8mb4）的连接
- **THEN** 连接必须使用指定的字符集
- **AND** 字符编码必须正确工作
- **AND** 默认字符集必须是 utf8mb4

#### Scenario: 配置时区

- **WHEN** 创建带有时区参数的连接
- **THEN** 连接必须使用指定的时区
- **AND** 时间值必须正确转换

---

### Requirement: 查询结果处理

系统必须正确处理和格式化查询结果。

#### Scenario: 返回列信息

- **WHEN** 执行 SELECT 查询
- **THEN** QueryResult 必须包含列名
- **AND** 列类型必须正确识别
- **AND** 列顺序必须与 SELECT 语句一致

#### Scenario: 返回行数据

- **WHEN** 执行 SELECT 查询返回多行
- **THEN** QueryResult 必须包含所有行数据
- **AND** 每行数据必须按列名或索引访问
- **AND** NULL 值必须正确处理

#### Scenario: 处理空结果集

- **WHEN** 执行 SELECT 查询返回零行
- **THEN** QueryResult 必须包含空的行数组
- **AND** 列信息必须仍然存在
- **AND** 不得返回错误

---

### Requirement: 错误处理

系统必须提供清晰和有意义的错误消息。

#### Scenario: 包装底层错误

- **WHEN** MySQL 返回错误
- **THEN** 系统必须用上下文包装错误
- **AND** 错误消息必须包含实例名称和操作类型
- **AND** 必须可以使用 errors.Is() 检查错误类型

#### Scenario: 记录连接错误

- **WHEN** 发生连接错误
- **THEN** 系统必须记录错误详情（不包含密码）
- **AND** 必须包含实例名称、主机和端口
- **AND** 必须使用结构化日志格式
