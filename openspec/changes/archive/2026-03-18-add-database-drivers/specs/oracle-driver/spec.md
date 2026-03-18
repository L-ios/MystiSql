## Overview

Oracle Database 是企业级关系数据库管理系统。本文档定义 MystiSql 对 Oracle 的连接和操作支持。

## Connection

### Requirement: 支持标准 Oracle 连接

Oracle 连接必须支持:
- 用户名/密码认证
- 连接池配置
- 服务名 (Service Name) 或 SID 连接
- TLS 加密

#### Scenario: 连接到 Oracle 数据库
- **Given** 一个配置了 host、 port、 service、 username、 password 的 Oracle 实例
- **When** 系统建立连接
- **then** 连接成功建立并通过查询验证

### Requirement: 支持 Oracle 连接字符串

```
oracle://user:password@host:port/service
```

参数说明:
- `service`: Oracle 服务名 (SID)
- `as`: 连接为 SYSDBA/SYSOPER
- `pooled`: 启用连接池
- `timezone`: 会话时区

## Query Operations

### Requirement: 支持标准 PL/SQL 查询

支持所有 Oracle SQL 操作
- SELECT: 查询数据
- INSERT: 插入数据
- UPDATE: 更新数据
- DELETE: 删除数据
- EXEC: 执行存储过程
- CREATE/ALTER/DROP: DDL 操作
- PL/SQL 块支持

#### Scenario: 查询表数据
- **Given** 一个 SELECT 查询
- **When** 查询执行
- **then** 返回结果集

### Requirement: 支持 Oracle 特有功能

- ROWNUM/ROW_NUMBER: 行号
- DUAL: 双值表
- CONNECT BY: 层级查询
- MERGE: 合并操作
- FLASHBACK: 闪回查询
- 序列操作 (NEXTVAL, CURRVAL)

## Transaction Support

### Requirement: 支持分布式事务

- SET TRANSACTION: 设置事务
- COMMIT: 提交事务
- ROLLBACK: 回滚事务
- SAVEPOINT: 设置保存点
- 隔离级别配置

#### Scenario: 事务操作
- **Given** 开始事务后执行多个操作
- **When** 提交事务
- **then** 所有操作原子性生效

## Implementation Notes
- 使用 `github.com/sijms/go-ora/v2` 驱动 (纯 Go 实现)
- 实现 `Connection` 接口
- 无需 Oracle 客户端库
- 支持连接池配置
