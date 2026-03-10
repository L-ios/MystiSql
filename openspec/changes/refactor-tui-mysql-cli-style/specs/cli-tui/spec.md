## Purpose

定义 MystiSql CLI 的 TUI（文本用户界面）功能规范，提供简洁的 REPL 风格终端交互体验。

## REMOVED Requirements

### Requirement: 界面布局

**Reason**: 全屏多区域布局不适合 REPL 风格界面，已替换为简洁的流式输入输出模式。

**Migration**: 使用新的 `mysql-cli-style-repl` 和 `streaming-output` capability。

### Requirement: SQL 语法高亮

**Reason**: REPL 模式下语法高亮实现复杂度较高，初期版本暂不支持。

**Migration**: 后续版本可通过配置启用，或使用外部工具（如 `highlight` 管道）实现。

## MODIFIED Requirements

### Requirement: TUI 命令接口

系统必须提供 REPL 命令启动交互式终端界面。

#### Scenario: 启动 REPL

- **WHEN** 用户运行 `mystisql` 无参数或 `mystisql repl`
- **THEN** 系统必须启动交互式 REPL 界面
- **AND** 必须显示欢迎信息
- **AND** 必须显示当前实例信息
- **AND** 必须进入命令输入模式

#### Scenario: 指定初始实例

- **WHEN** 用户运行 `mystisql --instance <instance-name>` 或 `mystisql -i <instance-name>`
- **THEN** 系统必须连接到指定实例
- **AND** 提示符必须显示当前实例名称
- **AND** 如果实例不存在，必须显示错误信息并退出

---

### Requirement: 基本操作

系统必须支持基本的 REPL 操作。

#### Scenario: 执行 SQL 命令

- **WHEN** 用户输入以分号结尾的 SQL 后按 Enter 键
- **THEN** 系统必须执行该 SQL 命令
- **AND** 必须在终端显示执行结果
- **AND** 必须显示新的提示符

#### Scenario: 取消当前输入

- **WHEN** 用户按 Ctrl+C
- **THEN** 系统必须取消当前输入
- **AND** 输入缓冲区必须清空
- **AND** 系统必须显示新的提示符

#### Scenario: 退出 REPL

- **WHEN** 用户输入 `exit`、`quit` 或按 Ctrl+D
- **THEN** 系统必须退出 REPL 界面
- **AND** 必须返回到命令行

#### Scenario: 清屏

- **WHEN** 用户输入 `clear` 命令
- **THEN** 系统必须清空终端屏幕
- **AND** 必须保留当前实例连接

---

### Requirement: 命令历史

系统必须支持命令历史功能。

#### Scenario: 保存历史记录

- **WHEN** 用户执行 SQL 命令
- **THEN** 系统必须将该命令保存到历史记录
- **AND** 历史记录必须去重
- **AND** 历史记录必须保存到 `~/.mystisql/history` 文件

#### Scenario: 浏览历史命令

- **WHEN** 用户按上/下箭头键
- **THEN** 系统必须显示历史命令
- **AND** 上箭头必须显示更早的命令
- **AND** 下箭头必须显示更近的命令

#### Scenario: 历史持久化

- **WHEN** REPL 退出时
- **THEN** 系统必须将历史记录保存到本地文件
- **AND** 下次启动时必须加载历史记录
- **AND** 历史文件必须位于 `~/.mystisql/history`

---

### Requirement: 错误处理

系统必须提供清晰的错误显示。

#### Scenario: 显示错误信息

- **WHEN** SQL 执行失败
- **THEN** 系统必须在终端显示错误信息
- **AND** 错误信息必须以 "ERROR" 开头
- **AND** 必须包含错误类型和描述

#### Scenario: 错误恢复

- **WHEN** 发生错误后
- **THEN** 用户必须能够继续输入新的 SQL 命令
- **AND** REPL 界面必须保持正常
- **AND** 连接状态必须保持（除非连接断开）
