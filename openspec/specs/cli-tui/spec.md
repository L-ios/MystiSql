# CLI TUI 接口规范

## Purpose

定义 MystiSql CLI 的文本用户界面（TUI）功能，提供类似 MySQL 命令行的简洁交互式体验，包括界面布局、基本操作、命令历史和语法高亮等功能。

## Requirements

### Requirement: TUI 启动和初始化

系统必须提供交互式 TUI 界面供用户使用。

#### Scenario: 默认启动 TUI

- **WHEN** 用户运行 `mystisql` 命令且无其他子命令
- **THEN** 系统必须自动启动交互式 TUI 界面
- **AND** 必须显示欢迎消息，包含配置的实例数量
- **AND** 必须自动选择默认实例或第一个可用实例
- **AND** 必须显示提示符格式为 `mystisql@<instance-name>`

#### Scenario: 指定实例启动

- **WHEN** 用户运行 `mystisql --instance <name>` 或 `mystisql tui --instance <name>`
- **THEN** 系统必须连接到指定的数据库实例
- **AND** 如果实例不存在，必须显示错误并退出
- **AND** 如果连接失败，必须显示详细错误信息

#### Scenario: TUI 界面布局

- **WHEN** TUI 启动成功
- **THEN** 必须显示类似 MySQL 的简洁界面
- **AND** 不显示复杂的装饰性元素（如边框、图标等）
- **AND** 必须支持响应式布局，适配不同终端尺寸

---

### Requirement: SQL 执行和结果显示

系统必须支持在 TUI 中执行 SQL 并显示结果。

#### Scenario: 执行单行 SQL

- **WHEN** 用户在提示符后输入 SQL 语句并按 Enter
- **THEN** 系统必须执行该 SQL 语句
- **AND** 必须以表格形式显示查询结果
- **AND** 必须显示结果行数和执行时间
- **AND** 提示符必须重新出现等待下一个命令

#### Scenario: 执行 INSERT/UPDATE/DELETE

- **WHEN** 用户执行数据修改语句（INSERT/UPDATE/DELETE）
- **THEN** 系统必须显示受影响的行数
- **AND** 如果有最后插入 ID，必须显示该 ID
- **AND** 必须显示执行时间

#### Scenario: 显示错误信息

- **WHEN** SQL 执行失败
- **THEN** 系统必须显示清晰的错误消息
- **AND** 错误消息必须包含错误类型和详细描述
- **AND** 必须立即返回提示符，允许用户继续输入

---

### Requirement: 实例切换

系统必须支持在 TUI 中切换数据库实例。

#### Scenario: 使用 Tab 键切换实例

- **WHEN** 用户按 Tab 键
- **THEN** 系统必须切换到下一个可用的数据库实例
- **AND** 提示符必须更新为新实例名称
- **AND** 如果只有一个实例，Tab 键无效

#### Scenario: 显示实例列表

- **WHEN** 用户输入 `show instances` 命令
- **THEN** 系统必须显示所有配置的实例列表
- **AND** 列表必须包含实例名称、类型和状态
- **AND** 必须标记当前激活的实例

---

### Requirement: 命令历史

系统必须维护 SQL 命令历史。

#### Scenario: 浏览历史命令

- **WHEN** 用户按上/下箭头键
- **THEN** 系统必须在命令历史中导航
- **AND** 上键显示更早的命令
- **AND** 下键显示更新的命令

#### Scenario: 持久化历史记录

- **WHEN** 用户执行 SQL 命令
- **THEN** 系统必须将命令保存到历史文件（~/.mystisql_history）
- **AND** 历史文件必须按时间顺序保存
- **AND** 重启 TUI 后必须能访问历史命令

#### Scenario: 历史记录限制

- **WHEN** 历史记录超过 1000 条
- **THEN** 系统必须删除最旧的记录
- **AND** 必须保持历史文件大小合理

---

### Requirement: SQL 语法高亮

系统必须提供 SQL 语法高亮功能。

#### Scenario: 高亮 SQL 关键字

- **WHEN** 用户输入 SQL 语句
- **THEN** SQL 关键字（SELECT, FROM, WHERE 等）必须以不同颜色显示
- **AND** 函数名必须以另一种颜色显示
- **AND** 字符串必须用引号颜色标识

#### Scenario: 支持多种 SQL 方言

- **WHEN** 连接到不同类型的数据库
- **THEN** 语法高亮必须适配该数据库的 SQL 方言
- **AND** MySQL、PostgreSQL、Oracle 的特定关键字必须被正确识别

---

### Requirement: 快捷键和特殊命令

系统必须支持快捷键和特殊命令。

#### Scenario: 退出 TUI

- **WHEN** 用户输入 `exit`、`quit` 或按 Ctrl+C
- **THEN** 系统必须退出 TUI
- **AND** 必须清理资源并关闭连接

#### Scenario: 清屏

- **WHEN** 用户按 Ctrl+L 或输入 `clear`
- **THEN** 系统必须清除屏幕内容
- **AND** 必须保留提示符

#### Scenario: 显示帮助

- **WHEN** 用户输入 `help` 或 `?`
- **THEN** 系统必须显示帮助信息
- **AND** 帮助信息必须包含常用命令和快捷键列表

#### Scenario: 导出结果

- **WHEN** 用户按 Ctrl+E
- **THEN** 系统必须提示用户选择导出格式（CSV、JSON）
- **AND** 必须提示输入文件名
- **AND** 必须将最后的查询结果保存到指定文件

---

### Requirement: 依赖注入和配置

系统必须使用依赖注入方式管理 TUI 组件。

#### Scenario: 使用配置系统

- **WHEN** TUI 初始化
- **THEN** 必须从配置系统读取实例配置
- **AND** 必须使用实例注册中心（InstanceRegistry）获取实例
- **AND** 不得硬编码任何配置

#### Scenario: 无配置时的处理

- **WHEN** 未找到配置文件
- **THEN** 系统必须显示错误消息
- **AND** 必须提示用户如何创建配置文件
- **AND** 必须以非零状态码退出
