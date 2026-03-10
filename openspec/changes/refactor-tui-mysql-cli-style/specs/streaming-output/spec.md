## Purpose

定义 MystiSql REPL 中流式输出功能的规范，提供与 MySQL CLI 一致的查询结果输出格式。

## ADDED Requirements

### Requirement: 表格格式输出

系统必须提供类似 MySQL CLI 的表格格式输出。

#### Scenario: 标准表格输出

- **WHEN** 查询返回数据行
- **THEN** 系统必须以表格形式显示结果
- **AND** 表格必须包含边框线（+---+ 格式）
- **AND** 列标题必须显示在第一行
- **AND** 数据行必须紧跟在标题之后

#### Scenario: 列宽自动调整

- **WHEN** 显示查询结果
- **THEN** 系统必须根据内容自动调整列宽
- **AND** 列宽必须足够容纳最长的值
- **AND** 列宽不得超过终端宽度

#### Scenario: NULL 值显示

- **WHEN** 查询结果包含 NULL 值
- **THEN** 系统必须显示 "NULL" 字符串
- **AND** NULL 必须与普通字符串区分显示

---

### Requirement: 结果统计信息

系统必须显示查询执行统计。

#### Scenario: 查询结果统计

- **WHEN** 查询执行完成
- **THEN** 系统必须显示返回的行数
- **AND** 系统必须显示执行时间
- **AND** 格式必须为 "N rows in set (X.XXX sec)"

#### Scenario: 执行结果统计

- **WHEN** 执行 INSERT/UPDATE/DELETE 语句
- **THEN** 系统必须显示受影响的行数
- **AND** 如果有自增 ID，必须显示最后插入的 ID
- **AND** 格式必须为 "Query OK, N rows affected (X.XXX sec)"

#### Scenario: 空结果集

- **WHEN** 查询返回空结果
- **THEN** 系统必须显示 "Empty set (X.XXX sec)"
- **AND** 系统不得显示表格边框

---

### Requirement: 错误输出格式

系统必须提供清晰的错误输出格式。

#### Scenario: SQL 错误输出

- **WHEN** SQL 执行发生错误
- **THEN** 系统必须显示 "ERROR" 前缀
- **AND** 系统必须显示错误代码（如果可用）
- **AND** 系统必须显示错误消息
- **AND** 格式必须为 "ERROR <code>: <message>"

#### Scenario: 错误位置提示

- **WHEN** SQL 语法错误
- **THEN** 系统必须显示错误位置（如果数据库返回）
- **AND** 系统必须高亮显示错误行

---

### Requirement: 大结果集处理

系统必须能够处理大型结果集。

#### Scenario: 流式输出

- **WHEN** 查询返回大量数据
- **THEN** 系统必须流式输出结果
- **AND** 系统不得等待所有数据加载完成才开始显示
- **AND** 系统内存使用必须保持稳定

#### Scenario: 终端宽度适配

- **WHEN** 结果列数或列宽超过终端宽度
- **THEN** 系统必须自动换行或截断显示
- **AND** 系统必须提示用户结果已被截断

---

### Requirement: 特殊输出格式

系统必须支持特殊的输出格式。

#### Scenario: 单列输出

- **WHEN** 查询只返回一列
- **THEN** 系统可以使用简化的输出格式
- **AND** 每行只显示一个值

#### Scenario: 元数据查询输出

- **WHEN** 执行 SHOW、DESCRIBE、EXPLAIN 等命令
- **THEN** 系统必须以表格形式显示结果
- **AND** 格式必须与 MySQL CLI 一致
