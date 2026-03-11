## Purpose

定义 MystiSql CLI 的 MySQL 风格 REPL（Read-Eval-Print-Loop）交互界面规范，提供与 MySQL 命令行客户端高度相似的用户体验。

## ADDED Requirements

### Requirement: REPL 启动与欢迎信息

系统必须提供类似 MySQL CLI 的启动体验。

#### Scenario: 启动 REPL 界面

- **WHEN** 用户运行 `mystisql` 无参数或 `mystisql repl`
- **THEN** 系统必须显示欢迎信息
- **AND** 欢迎信息必须包含产品名称和版本
- **AND** 欢迎信息必须显示当前配置的实例数量
- **AND** 欢迎信息必须显示帮助提示（如 "Type 'help' or '?' for help"）

#### Scenario: 显示当前实例

- **WHEN** REPL 启动后
- **THEN** 系统必须显示当前连接的实例名称
- **AND** 如果没有配置实例，必须显示警告信息

---

### Requirement: 提示符格式

系统必须提供清晰的命令提示符。

#### Scenario: 标准提示符

- **WHEN** REPL 等待用户输入
- **THEN** 系统必须显示格式为 `mystisql@<instance>>` 的提示符
- **AND** 提示符必须使用绿色或醒目颜色显示
- **AND** 实例名称必须使用不同颜色区分

#### Scenario: 无实例提示符

- **WHEN** 没有配置任何实例
- **THEN** 系统必须显示格式为 `mystisql>>` 的提示符
- **AND** 必须在启动时显示警告信息

---

### Requirement: 命令执行循环

系统必须支持标准的 REPL 命令执行流程。

#### Scenario: 执行单行 SQL

- **WHEN** 用户输入以分号结尾的单行 SQL 并按 Enter
- **THEN** 系统必须执行该 SQL 语句
- **AND** 系统必须显示执行结果
- **AND** 系统必须显示新的提示符等待下一条命令

#### Scenario: 空输入处理

- **WHEN** 用户只按 Enter 键不输入任何内容
- **THEN** 系统必须显示新的提示符
- **AND** 系统不得执行任何操作

#### Scenario: 注释处理

- **WHEN** 用户输入以 `--` 或 `#` 开头的注释行
- **THEN** 系统必须忽略该行
- **AND** 系统必须显示新的提示符

---

### Requirement: 内置命令支持

系统必须支持类似 MySQL 的内置命令。

#### Scenario: 退出命令

- **WHEN** 用户输入 `exit`、`quit` 或 `\q`
- **THEN** 系统必须退出 REPL 界面
- **AND** 系统必须显示 "Bye" 或类似告别信息

#### Scenario: 帮助命令

- **WHEN** 用户输入 `help`、`?` 或 `\h`
- **THEN** 系统必须显示帮助信息
- **AND** 帮助信息必须包含可用命令列表
- **AND** 帮助信息必须包含 SQL 语法提示

#### Scenario: 清屏命令

- **WHEN** 用户输入 `clear` 或 `\! clear`
- **THEN** 系统必须清空终端屏幕
- **AND** 系统必须显示新的提示符

#### Scenario: 状态命令

- **WHEN** 用户输入 `status` 或 `\s`
- **THEN** 系统必须显示当前状态信息
- **AND** 状态信息必须包含当前实例名称和类型
- **AND** 状态信息必须包含连接状态

---

### Requirement: 错误处理

系统必须提供清晰的错误提示。

#### Scenario: SQL 执行错误

- **WHEN** SQL 执行失败
- **THEN** 系统必须显示错误信息
- **AND** 错误信息必须以 "ERROR" 开头
- **AND** 错误信息必须包含错误代码（如果可用）
- **AND** 系统必须返回提示符等待新输入

#### Scenario: 连接错误

- **WHEN** 数据库连接失败
- **THEN** 系统必须显示连接错误信息
- **AND** 系统必须提示用户检查实例配置
- **AND** 系统必须保持 REPL 运行不退出

#### Scenario: 未知命令

- **WHEN** 用户输入无法识别的命令
- **THEN** 系统必须显示错误提示
- **AND** 系统必须建议使用 `help` 命令查看帮助
