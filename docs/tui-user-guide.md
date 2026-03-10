# MystiSql TUI 用户指南

## 概述

MystiSql TUI（文本用户界面）提供了一个简洁、高效的命令行交互界面，让您能够在终端中直接访问和管理 Kubernetes 集群中的数据库实例。

TUI 采用类似 MySQL 命令行的简洁设计，无复杂装饰，专注于提供流畅的数据库操作体验。

## 启动 TUI

### 默认启动

```bash
# 无参数启动，自动进入 TUI 交互模式
./bin/mystisql

# 指定配置文件
./bin/mystisql --config /path/to/config.yaml
```

### 指定初始实例

```bash
# 启动并连接到特定实例
./bin/mystisql --instance local-mysql
```

## 界面布局

TUI 界面包含以下区域：

```
Welcome to the MystiSql monitor. Commands end with Enter.
Your MystiSql connection has 1 instance(s) configured.
Current instance: local-mysql

Type 'help' or '?' for help. Type 'exit' or Ctrl+C to quit.

mystisql@local-mysql> SELECT * FROM users LIMIT 2;
id  name    email
1   Alice   alice@example.com
2   Bob     bob@example.com

2 rows, 0.005s

mystisql@local-mysql> _
```

- **欢迎信息**: 显示配置的实例数量和当前实例
- **输入提示符**: 格式为 `mystisql@<instance-name>`，显示当前连接的实例
- **结果区域**: 显示查询结果或执行状态
- **状态信息**: 显示行数、执行时间等统计信息

## 基本操作

### 执行 SQL 查询

```sql
-- 查询数据
SELECT * FROM users LIMIT 10;

-- 显示表结构
DESC users;

-- 显示数据库列表
SHOW DATABASES;
```

**结果示例**:
```
id  name    email
1   Alice   alice@example.com
2   Bob     bob@example.com

2 rows, 0.005s
```

### 执行数据修改

```sql
-- 插入数据
INSERT INTO users (name, email) VALUES ('Charlie', 'charlie@example.com');

-- 更新数据
UPDATE users SET email = 'newemail@example.com' WHERE id = 1;

-- 删除数据
DELETE FROM users WHERE id = 1;
```

**结果示例**:
```
受影响行数: 1
最后插入ID: 3
执行时间: 0.003s
```

## 快捷键

| 快捷键 | 功能 | 说明 |
|--------|------|------|
| **Enter** | 执行 SQL | 执行当前输入的 SQL 语句 |
| **Tab** | 切换实例 | 在多个数据库实例之间切换 |
| **Ctrl+E** | 导出结果 | 将最后的查询结果导出为 CSV 或 JSON |
| **?** | 显示帮助 | 显示快捷键和命令列表 |
| **Ctrl+C** | 退出 | 退出 TUI |
| **↑** | 上一个历史 | 浏览上一条执行的 SQL |
| **↓** | 下一个历史 | 浏览下一条执行的 SQL |
| **Esc** | 取消操作 | 取消当前选择或关闭弹出窗口 |

## 实例管理

### 查看实例列表

```sql
-- 显示所有配置的实例
show instances
```

**示例输出**:
```
可用实例:
  local-mysql       [MySQL]     健康
  local-postgres    [PostgreSQL] 健康
→ test-oracle      [Oracle]    健康

当前实例: test-oracle
```

### 切换实例

**方式 1: 使用 Tab 键**
- 按 Tab 键循环切换到下一个可用实例

**方式 2: 使用命令**
```sql
-- 切换到指定实例
use local-postgres
```

**切换成功提示**:
```
已切换到实例: local-postgres
连接建立时间: 15ms
```

### 查看实例详情

```sql
-- 显示实例详细信息
show instance local-mysql
```

**示例输出**:
```
实例名称: local-mysql
类型: MySQL
主机: localhost
端口: 3306
状态: 健康
连接池:
  活跃连接: 3
  空闲连接: 7
  最大连接: 10
最后连接时间: 2026-03-10 15:30:45
```

## 命令历史

TUI 自动保存您执行的 SQL 命令历史：

- **浏览历史**: 使用 ↑/↓ 箭头键浏览之前执行的命令
- **历史持久化**: 历史记录保存在 `~/.mystisql_history` 文件中
- **历史限制**: 最多保存 1000 条历史记录

## 结果管理

### 结果分页

当查询结果超过终端高度时，TUI 会自动分页：

- 按 **空格键** 显示下一页
- 按 **q 键** 退出分页模式

### 结果导出

**方式 1: 使用快捷键**
1. 执行查询后，按 **Ctrl+E**
2. 选择导出格式（CSV、JSON、Table）
3. 输入文件名

**方式 2: 使用命令**
```sql
-- 导出为 CSV
export result.csv csv

-- 导出为 JSON
export result.json json
```

**导出成功提示**:
```
已导出 2 行数据到: result.csv
```

### 结果排序和过滤

```sql
-- 在结果中搜索
FILTER name = 'Alice'

-- 按列排序
ORDER BY id DESC
```

## 事务支持

### 基本事务操作

```sql
-- 开始事务
BEGIN;

-- 执行多个操作
INSERT INTO users (name, email) VALUES ('David', 'david@example.com');
UPDATE stats SET user_count = user_count + 1;

-- 提交事务
COMMIT;

-- 或回滚事务
ROLLBACK;
```

**事务状态显示**:
```
[TX] mystisql@local-mysql> _
```

提示符中的 `[TX]` 表示当前有活跃事务。

### 查看事务状态

```sql
SHOW TRANSACTION STATUS
```

**输出示例**:
```
事务状态: 活跃
开始时间: 2026-03-10 15:35:20
隔离级别: REPEATABLE READ
持续时间: 45s
```

### 设置隔离级别

```sql
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;
```

## 性能优化

### 查看执行计划

```sql
-- 使用 EXPLAIN 查看查询执行计划
EXPLAIN SELECT * FROM users WHERE email = 'test@example.com';
```

**示例输出**:
```
id  select_type  table  type  possible_keys  key       key_len  rows  Extra
1   SIMPLE       users  const PRIMARY        PRIMARY   4        1     Using where
```

### 查询性能提示

- 查询执行时间超过 1 秒会显示进度指示
- 查询执行时间超过 5 秒会显示性能警告
- 可以按 **Ctrl+C** 取消长时间运行的查询

## 错误处理

### 语法错误

```sql
SELECT * FORM users;
```

**错误提示**:
```
ERROR: SQL 语法错误: near "FORM"
位置: 第 10 个字符
提示: 您是否想使用 "FROM"?
```

### 连接错误

```
ERROR: 连接失败: 连接超时
实例: local-mysql
建议: 检查网络连接或增加超时时间
```

### 权限错误

```
ERROR: 权限不足: INSERT 命令被拒绝
表: users
建议: 联系管理员获取 INSERT 权限
```

## 批量操作

### 从文件执行 SQL

```sql
-- 执行 SQL 脚本文件
source /path/to/script.sql
```

**脚本文件示例** (script.sql):
```sql
-- 创建用户表
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100),
    email VARCHAR(100)
);

-- 插入测试数据
INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com');
INSERT INTO users (name, email) VALUES ('Bob', 'bob@example.com');
```

### 多语句执行

可以直接粘贴多条 SQL 语句（用分号分隔）：

```sql
INSERT INTO users (name, email) VALUES ('Carol', 'carol@example.com');
UPDATE stats SET user_count = user_count + 1;
SELECT * FROM users WHERE name = 'Carol';
```

## 高级功能

### 实例分组

```sql
-- 创建实例分组
create group production local-mysql, local-postgres

-- 查看分组
show groups

-- 按组列出实例
show instances --group production
```

### 快速访问历史实例

```sql
-- 显示最近使用的实例
show recent-instances

-- 快速切换到上一个实例（快捷键 Ctrl+-）
```

## 常见问题

### Q: 如何退出 TUI？

A: 有三种方式：
1. 输入 `exit` 或 `quit` 命令
2. 按 **Ctrl+C**
3. 按 **Ctrl+D**

### Q: 如何清空屏幕？

A: 按 **Ctrl+L** 或输入 `clear` 命令

### Q: 历史命令保存在哪里？

A: 保存在 `~/.mystisql_history` 文件中

### Q: 如何查看当前连接的实例？

A: 查看提示符中的实例名称，或运行 `show instances` 命令

### Q: 查询超时怎么办？

A: 
- 默认查询超时为 30 秒
- 可以在配置文件中调整超时时间
- 按 **Ctrl+C** 取消长时间运行的查询

### Q: 如何处理大结果集？

A:
- TUI 自动分页显示结果
- 建议在查询中使用 `LIMIT` 子句
- 使用导出功能将结果保存到文件

## 最佳实践

1. **使用 LIMIT**: 查询大表时始终使用 `LIMIT` 避免一次性加载过多数据
2. **定期提交事务**: 长时间运行的事务会锁定资源，建议及时提交或回滚
3. **检查执行计划**: 对慢查询使用 `EXPLAIN` 分析执行计划
4. **导出重要结果**: 重要查询结果及时导出保存
5. **使用实例分组**: 为开发和生产环境创建不同的实例分组
6. **善用历史命令**: 使用上下箭头快速重用之前执行的 SQL

## 获取帮助

- **TUI 内帮助**: 按 **?** 键
- **命令帮助**: 输入 `help` 命令
- **项目文档**: 查看 README.MD 和 AGENTS.md
- **问题反馈**: 提交 Issue 到项目仓库

## 相关资源

- [MystiSql README](../README.MD)
- [API 参考文档](./api-reference.md)
- [部署指南](./phase3-deployment-guide.md)
- [AGENTS.md - 开发者指南](../AGENTS.md)
