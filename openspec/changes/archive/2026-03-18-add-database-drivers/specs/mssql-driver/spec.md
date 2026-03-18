## Overview

Microsoft SQL Server 是企业级关系数据库管理系统。本文档定义 MystiSql 对 MSSQL 的连接和操作支持。

## Connection

### Requirement: 支持标准 MSSQL 连接

MSSQL 连接必须支持:
- SQL Server 认证 (用户名/密码)
- Windows 集成认证
- Azure AD 认证
- 连接池配置
- 加密连接 (TLS)

#### Scenario: 连接到 SQL Server
- **Given** 一个配置了 host、 port、 username、 password 的 MSSQL 实例
- **When** 系统建立连接
- **then** 连接成功建立并通过查询验证

### Requirement: 支持 MSSQL 连接字符串

```
sqlserver://user:password@host:port/database?params
```

参数说明:
- `encrypt`: true (启用加密)
- `trustservercertificate`: true (信任服务器证书)
- `connection timeout`: 连接超时
- `app name`: 应用程序名称

## Query Operations

### Requirement: 支持标准 T-SQL 查询

支持所有 T-SQL 操作
- SELECT: 查询数据
- INSERT: 插入数据
- UPDATE: 更新数据
- DELETE: 删除数据
- EXEC: 执行存储过程
- CREATE/ALTER/DROP: DDL 操作

#### Scenario: 查询表数据
- **Given** 一个 SELECT 查询
- **When** 查询执行
- **then** 返回结果集

### Requirement: 支持 MSSQL 特有功能

- TOP/LIMIT: 分页查询
- OFFSET FETCH: 偏移获取
- OUTPUT INSERTED: 获取插入的 ID
- MERGE: 合并操作
- BULK INSERT: 批量插入

## Transaction Support

### Requirement: 支持分布式事务

- BEGIN TRANSACTION: 开始事务
- COMMIT: 提交事务
- ROLLBACK: 回滚事务
- SAVEPOINT: 设置保存点
- 隔离级别配置

#### Scenario: 事务操作
- **Given** 开始事务后执行多个操作
- **When** 提交事务
- **then** 所有操作原子性生效

## Implementation Notes
- 使用 `github.com/microsoft/go-mssqldb` 驱动 (微软官方)
- 实现 `Connection` 接口
- 使用标准 `database/sql` 接口
- 支持连接池配置
