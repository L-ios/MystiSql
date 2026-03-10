# CLI TUI 功能规范

## Purpose

定义 MystiSql CLI 的 TUI（文本用户界面）功能规范，包括界面布局、交互方式和基本操作，为用户提供直观的终端交互体验。

## Requirements

### Requirement: TUI 命令接口

系统必须提供 TUI 命令启动交互式终端界面。

#### Scenario: 启动 TUI

- **WHEN** 用户运行 `mystisql tui`
- **THEN** 系统必须启动交互式 TUI 界面
- **AND** 必须显示界面布局（状态栏、输入区、结果区）
- **AND** 必须进入命令输入模式

#### Scenario: 指定初始实例

- **WHEN** 用户运行 `mystisql tui --instance <instance-name>`
- **THEN** 系统必须连接到指定实例
- **AND** 状态栏必须显示当前实例名称
- **AND** 如果实例不存在，必须显示错误信息

---

### Requirement: 界面布局

系统必须提供清晰的多区域界面布局。

#### Scenario: 顶部状态栏

- **WHEN** TUI 启动后
- **THEN** 顶部必须显示状态栏
- **AND** 状态栏必须包含：当前实例名、数据库名、连接状态
- **AND** 状态栏必须实时更新

#### Scenario: 主输入区域

- **WHEN** TUI 启动后
- **THEN** 必须提供主输入区域用于输入 SQL 命令
- **AND** 输入区域必须支持多行输入
- **AND** 输入区域必须有明显的视觉边界

#### Scenario: 结果显示区域

- **WHEN** TUI 启动后
- **THEN** 必须提供结果显示区域
- **AND** 结果区域必须支持表格形式展示
- **AND** 结果区域必须支持滚动查看

#### Scenario: 底部状态栏

- **WHEN** TUI 启动后
- **THEN** 底部必须显示状态栏
- **AND** 必须显示当前操作状态
- **AND** 必须显示快捷键提示

---

### Requirement: 基本操作

系统必须支持基本的 TUI 操作。

#### Scenario: 执行 SQL 命令

- **WHEN** 用户在输入区域输入 SQL 后按 Enter 键
- **THEN** 系统必须执行该 SQL 命令
- **AND** 必须在结果区域显示执行结果
- **AND** 必须更新状态栏显示执行状态

#### Scenario: 取消当前输入

- **WHEN** 用户按 Ctrl+C
- **THEN** 系统必须取消当前输入
- **AND** 输入区域必须清空
- **AND** 状态栏必须显示"已取消"消息

#### Scenario: 退出 TUI

- **WHEN** 用户按 Ctrl+D 或输入 `exit` 命令
- **THEN** 系统必须退出 TUI 界面
- **AND** 必须返回到命令行

#### Scenario: 清屏

- **WHEN** 用户按 Ctrl+L
- **THEN** 系统必须清空结果显示区域
- **AND** 输入区域必须保持不变

---

### Requirement: 命令历史

系统必须支持命令历史功能。

#### Scenario: 保存历史记录

- **WHEN** 用户执行 SQL 命令
- **THEN** 系统必须将该命令保存到历史记录
- **AND** 历史记录必须包含时间戳
- **AND** 历史记录必须去重

#### Scenario: 浏览历史命令

- **WHEN** 用户按上/下箭头键
- **THEN** 系统必须显示历史命令
- **AND** 上箭头必须显示更早的命令
- **AND** 下箭头必须显示更近的命令

#### Scenario: 历史持久化

- **WHEN** TUI 退出时
- **THEN** 系统必须将历史记录保存到本地文件
- **AND** 下次启动时必须加载历史记录
- **AND** 历史文件必须位于用户主目录（如 ~/.mystisql/history）

---

### Requirement: SQL 语法高亮

系统必须支持 SQL 语法高亮。

#### Scenario: 关键字高亮

- **WHEN** 用户输入 SQL 命令
- **THEN** SQL 关键字必须以不同颜色显示
- **AND** 关键字包括：SELECT、FROM、WHERE、INSERT、UPDATE、DELETE 等
- **AND** 关键字必须不区分大小写

#### Scenario: 函数名高亮

- **WHEN** 用户输入 SQL 函数
- **THEN** 函数名必须以不同颜色显示
- **AND** 函数名必须包括常见数据库函数

#### Scenario: 字符串高亮

- **WHEN** 用户输入字符串字面量
- **THEN** 字符串必须以不同颜色显示
- **AND** 必须支持单引号和双引号字符串

#### Scenario: 多数据库方言支持

- **WHEN** 连接到不同类型的数据库
- **THEN** 语法高亮必须适配对应数据库的 SQL 方言
- **AND** 必须支持 MySQL、PostgreSQL、Oracle 等方言

---

### Requirement: 错误处理

系统必须提供清晰的错误显示。

#### Scenario: 显示错误信息

- **WHEN** SQL 执行失败
- **THEN** 系统必须在结果区域显示错误信息
- **AND** 错误信息必须用红色高亮显示
- **AND** 必须包含错误类型和描述

#### Scenario: 错误定位

- **WHEN** SQL 语法错误
- **THEN** 系统必须指出错误位置
- **AND** 必须高亮显示错误的 SQL 片段
- **AND** 必须显示行号和列号（如果可用）

#### Scenario: 错误恢复

- **WHEN** 发生错误后
- **THEN** 用户必须能够继续输入新的 SQL 命令
- **AND** TUI 界面必须保持正常
- **AND** 连接状态必须保持（除非连接断开）
