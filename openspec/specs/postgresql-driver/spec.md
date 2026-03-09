# PostgreSQL 驱动规范

## Purpose

定义 MystiSql 的 PostgreSQL 数据库驱动集成功能，使用 pgx 驱动提供 PostgreSQL 数据库连接、查询执行和连接池管理。

## Requirements

### Requirement: PostgreSQL 连接建立
系统 SHALL 使用 pgx 驱动建立 PostgreSQL 数据库连接。

#### Scenario: 成功连接 PostgreSQL
- **WHEN** 配置 PostgreSQL 实例（type=postgresql）
- **THEN** 系统使用 pgx 驱动建立连接

#### Scenario: 连接字符串格式
- **WHEN** 配置 PostgreSQL 连接
- **THEN** 连接字符串格式为 `postgres://user:password@host:port/database`

---

### Requirement: PostgreSQL 连接池管理
系统 SHALL 为 PostgreSQL 实现连接池管理（复用 MySQL 的 ConnectionPool 接口）。

#### Scenario: 创建连接池
- **WHEN** 初始化 PostgreSQL 实例
- **THEN** 系统创建连接池，复用 `pool.maxOpen`、`pool.maxIdle` 配置

#### Scenario: 连接复用
- **WHEN** 多次查询同一 PostgreSQL 实例
- **THEN** 系统从连接池获取连接，避免重复建立连接

#### Scenario: 连接健康检查
- **WHEN** PostgreSQL 连接池执行健康检查
- **THEN** 系统执行 `SELECT 1` 验证连接可用性

---

### Requirement: PostgreSQL 查询执行
系统 SHALL 支持在 PostgreSQL 上执行 SQL 查询。

#### Scenario: 执行 SELECT 查询
- **WHEN** 用户在 PostgreSQL 实例上执行 SELECT 查询
- **THEN** 系统返回查询结果（columns、rows）

#### Scenario: 执行 INSERT/UPDATE/DELETE
- **WHEN** 用户在 PostgreSQL 实例上执行数据修改操作
- **THEN** 系统返回影响行数和最后插入 ID（如适用）

#### Scenario: 执行 PostgreSQL 特有语法
- **WHEN** 用户执行 PostgreSQL 特有语法（如 `RETURNING` 子句）
- **THEN** 系统正确执行并返回结果

---

### Requirement: 多数据库类型路由
系统 SHALL 根据实例类型自动选择对应的数据库驱动。

#### Scenario: 识别实例类型
- **WHEN** 用户查询实例名为 `production-postgresql`
- **THEN** 系统根据实例配置的 `type: postgresql` 选择 pgx 驱动

#### Scenario: MySQL 实例使用 MySQL 驱动
- **WHEN** 用户查询实例名为 `production-mysql`
- **THEN** 系统根据实例配置的 `type: mysql` 选择 go-sql-driver 驱动

---

### Requirement: PostgreSQL 连接配置
系统 SHALL 支持 PostgreSQL 特有的连接配置。

#### Scenario: 配置 SSL 模式
- **WHEN** 配置 PostgreSQL 实例包含 `sslmode: require`
- **THEN** 连接使用 SSL 加密

#### Scenario: 配置连接超时
- **WHEN** 配置 PostgreSQL 实例包含 `connectTimeout: 10s`
- **THEN** 连接超时时间为 10 秒

---

### Requirement: PostgreSQL 错误处理
系统 SHALL 正确处理 PostgreSQL 特有的错误。

#### Scenario: 处理唯一约束冲突
- **WHEN** INSERT 违反唯一约束
- **THEN** 系统返回明确的错误信息（包含冲突的字段名）

#### Scenario: 处理外键约束错误
- **WHEN** 操作违反外键约束
- **THEN** 系统返回明确的错误信息（包含相关表和外键名）
