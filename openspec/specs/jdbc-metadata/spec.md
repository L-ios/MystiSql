# jdbc-metadata

## Purpose

TBD: DatabaseMetaData implementation for JDBC driver to query database metadata.

## Requirements

### Requirement: DatabaseMetaData Interface

MystiSqlDatabaseMetaData SHALL实现`java.sql.DatabaseMetaData`接口，提供数据库元数据查询能力。

#### Scenario: Get database product info

- **WHEN** 应用调用`metaData.getDatabaseProductName()`
- **THEN** SHALL返回"MystiSql Gateway"
- **OR** 返回实际数据库类型（如"MySQL"），如果Gateway配置为透传模式

#### Scenario: Get driver info

- **WHEN** 应用调用`metaData.getDriverName()`和`metaData.getDriverVersion()`
- **THEN** SHALL返回"MystiSql JDBC Driver"和驱动版本号

#### Scenario: Get connection info

- **WHEN** 应用调用`metaData.getUserName()`和`metaData.getURL()`
- **THEN** SHALL返回当前连接的用户名和JDBC URL

### Requirement: Catalog Metadata

DatabaseMetaData SHALL支持查询catalogs（数据库列表）。

#### Scenario: Get all catalogs

- **WHEN** 应用调用`metaData.getCatalogs()`
- **THEN** 驱动SHALL发送GET请求到`/api/v1/metadata/catalogs`
- **AND** 返回ResultSet包含所有catalog名称
- **AND** ResultSet列名为`TABLE_CAT`

#### Scenario: Empty catalog list

- **WHEN** 数据库实例没有catalogs
- **THEN** 返回空ResultSet（next()返回false）

### Requirement: Schema Metadata

DatabaseMetaData SHALL支持查询schemas。

#### Scenario: Get all schemas

- **WHEN** 应用调用`metaData.getSchemas()`
- **THEN** 驱动SHALL发送GET请求到`/api/v1/metadata/schemas`
- **AND** 返回ResultSet包含schema名称和所属catalog
- **AND** ResultSet列名为`TABLE_SCHEM`和`TABLE_CATALOG`

#### Scenario: Filter schemas by catalog

- **WHEN** 应用调用`metaData.getSchemas("production", null)`
- **THEN** 请求SHALL包含参数`?catalog=production`
- **AND** 仅返回production catalog下的schemas

### Requirement: Table Metadata

DatabaseMetaData SHALL支持查询表信息。

#### Scenario: Get all tables

- **WHEN** 应用调用`metaData.getTables(null, null, "%", null)`
- **THEN** 驱动SHALL发送GET请求到`/api/v1/metadata/tables`
- **AND** 返回ResultSet包含表信息
- **AND** ResultSet SHALL包含以下列：
  - TABLE_CAT
  - TABLE_SCHEM
  - TABLE_NAME
  - TABLE_TYPE
  - REMARKS

#### Scenario: Filter tables by type

- **WHEN** 应用调用`metaData.getTables(null, null, "%", new String[]{"TABLE", "VIEW"})`
- **THEN** 请求SHALL包含参数`?types=TABLE,VIEW`
- **AND** 仅返回TABLE和VIEW类型的表

#### Scenario: Filter tables by name pattern

- **WHEN** 应用调用`metaData.getTables(null, "public", "user_%", null)`
- **THEN** 请求SHALL包含参数`?schema=public&pattern=user_%`
- **AND** 仅返回名称匹配"user_%"的表

### Requirement: Column Metadata

DatabaseMetaData SHALL支持查询列信息。

#### Scenario: Get columns of a table

- **WHEN** 应用调用`metaData.getColumns(null, "public", "users", null)`
- **THEN** 驱动SHALL发送GET请求到`/api/v1/metadata/columns?schema=public&table=users`
- **AND** 返回ResultSet包含列信息
- **AND** ResultSet SHALL包含以下列：
  - TABLE_CAT
  - TABLE_SCHEM
  - TABLE_NAME
  - COLUMN_NAME
  - DATA_TYPE (java.sql.Types常量)
  - TYPE_NAME (数据库类型名称)
  - COLUMN_SIZE
  - NULLABLE
  - REMARKS
  - ORDINAL_POSITION

#### Scenario: Map MySQL types to JDBC types

- **WHEN** 数据库列类型为MySQL的"VARCHAR(255)"
- **THEN** DATA_TYPE SHALL返回`java.sql.Types.VARCHAR` (12)
- **AND** TYPE_NAME SHALL返回"VARCHAR"
- **AND** COLUMN_SIZE SHALL返回255

#### Scenario: Handle unknown types

- **WHEN** 数据库列类型无法映射到标准JDBC类型
- **THEN** DATA_TYPE SHALL返回`java.sql.Types.OTHER` (1111)
- **AND** TYPE_NAME SHALL返回原始类型名称

### Requirement: Primary Key Metadata

DatabaseMetaData SHALL支持查询主键信息。

#### Scenario: Get primary keys

- **WHEN** 应用调用`metaData.getPrimaryKeys(null, "public", "users")`
- **THEN** 驱动SHALL发送GET请求到`/api/v1/metadata/primary-keys?schema=public&table=users`
- **AND** 返回ResultSet包含主键列信息
- **AND** ResultSet SHALL包含以下列：
  - TABLE_CAT
  - TABLE_SCHEM
  - TABLE_NAME
  - COLUMN_NAME
  - KEY_SEQ (主键列的顺序)
  - PK_NAME (主键约束名称)

#### Scenario: Composite primary key

- **WHEN** 表有复合主键（user_id, role_id）
- **THEN** ResultSet SHALL返回两行
- **AND** KEY_SEQ SHALL分别为1和2

#### Scenario: No primary key

- **WHEN** 表没有主键
- **THEN** 返回空ResultSet

### Requirement: Index Metadata

DatabaseMetaData SHALL支持查询索引信息。

#### Scenario: Get indexes

- **WHEN** 应用调用`metaData.getIndexInfo(null, "public", "users", false, true)`
- **THEN** 驱动SHALL发送GET请求到`/api/v1/metadata/indexes?schema=public&table=users`
- **AND** 返回ResultSet包含索引信息
- **AND** ResultSet SHALL包含以下列：
  - TABLE_CAT
  - TABLE_SCHEM
  - TABLE_NAME
  - NON_UNIQUE
  - INDEX_NAME
  - TYPE
  - COLUMN_NAME
  - ASC_OR_DESC

#### Scenario: Distinguish unique and non-unique indexes

- **WHEN** 索引为UNIQUE索引
- **THEN** NON_UNIQUE列SHALL返回false
- **WHEN** 索引为普通索引
- **THEN** NON_UNIQUE列SHALL返回true

### Requirement: Metadata Caching

DatabaseMetaData查询SHALL使用缓存，避免频繁查询后端数据库。

#### Scenario: Cache metadata with TTL

- **WHEN** 第一次调用`getTables()`
- **THEN** 驱动SHALL缓存结果，TTL为5分钟
- **AND** 5分钟内再次调用SHALL返回缓存数据
- **AND** 不发送HTTP请求

#### Scenario: Bypass cache with refresh

- **WHEN** 应用调用`connection.refreshMetadata()`（MystiSql扩展方法）
- **THEN** 驱动SHALL清空所有元数据缓存
- **AND** 下一次元数据查询SHALL重新从Gateway获取

#### Scenario: Cache key includes instance name

- **WHEN** 连接不同的数据库实例
- **THEN** 元数据缓存SHALL按实例名称隔离
- **AND** 不同实例的元数据互不影响

### Requirement: Metadata Error Handling

元数据查询SHALL正确处理错误。

#### Scenario: Table not found

- **WHEN** 查询的表不存在
- **THEN** SHALL返回空ResultSet（不抛异常）

#### Scenario: Metadata API not available

- **WHEN** Gateway未实现元数据API（返回404）
- **THEN** 驱动SHALL抛出SQLException
- **AND** 错误信息说明需要升级Gateway版本

#### Scenario: Metadata query timeout

- **WHEN** 元数据查询超时（默认10秒）
- **THEN** SHALL抛出SQLTimeoutException
