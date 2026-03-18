## Overview

SQLite 是轻量级嵌入式关系数据库，无需独立服务器进程，适合测试、开发和小型应用。本文档定义 MystiSql 对 SQLite 的连接和操作支持。

## Connection

### Requirement: 支持文件数据库连接

SQLite 连接必须支持：
- 基于文件的数据库路径
- 内存数据库模式 (`:memory:`)
- 只读模式
- WAL 模式配置

#### Scenario: 连接到文件数据库
- **Given** 一个配置了数据库文件路径的 SQLite 实例
- **When** 系统建立连接
- **then** 自动创建数据库文件并建立连接

### Requirement: 支持 SQLite 连接字符串

```
file:/path/to/database.db?cache=shared&mode=rwc
```

参数说明:
- `cache`: shared (默认), | private | shared
- `mode`: rwc (默认) | ro | rw | rwo | memory
- `_busy_timeout`: 设置忙等待超时

## Query Operations

### Requirement: 支持标准 SQL 查询

支持所有标准 SQLite SQL 操作
- SELECT: 查询数据
- INSERT: 插入数据
- UPDATE: 更新数据
- DELETE: 删除数据
- CREATE TABLE: 创建表
- ALTER TABLE: 修改表结构
- DROP TABLE: 删除表

#### Scenario: 查询表数据
- **Given** 一个 SELECT 查询
- **When** 查询执行
- **then** 返回结果集

### Requirement: 支持 SQLite 特有语法

- WITHOUT ROWID: 创建无 ROWID 表
- AUTOINCREMENT: 自增主键
- PRAGMA: 编译指示
- ATTACH: 附加数据库

## Write Operations

### Requirement: 支持事务

SQLite 支持完整的事务支持
- BEGIN TRANSACTION: 开始事务
- COMMIT: 提交事务
- ROLLBACK: 回滚事务

#### Scenario: 事务操作
- **Given** 开始事务后执行多个操作
- **When** 提交事务
- **then** 所有操作原子性生效

## Implementation Notes

- 使用 `modernc.org/sqlite` 驱动（纯 Go 实现，无 CGO)
- 实现 `Connection` 接口
- 使用标准 `database/sql` 接口
- 支持连接池配置
