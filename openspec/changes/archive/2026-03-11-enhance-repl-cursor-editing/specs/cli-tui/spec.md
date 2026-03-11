# CLI TUI 功能规范 - 变更增量

## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: 行编辑快捷键

系统必须支持完整的行编辑快捷键。

#### Scenario: 光标移动快捷键

- **WHEN** 用户使用光标移动快捷键
- **THEN** 系统必须响应以下快捷键：
  | 快捷键 | 功能 |
  |--------|------|
  | ← / Ctrl+B | 光标左移一个字符 |
  | → / Ctrl+F | 光标右移一个字符 |
  | Home / Ctrl+A | 光标移到行首 |
  | End / Ctrl+E | 光标移到行尾 |

#### Scenario: 删除快捷键

- **WHEN** 用户使用删除快捷键
- **THEN** 系统必须响应以下快捷键：
  | 快捷键 | 功能 |
  |--------|------|
  | Backspace | 删除光标前一个字符 |
  | Delete | 删除光标位置字符 |
  | Ctrl+U | 删除从行首到光标位置 |
  | Ctrl+K | 删除从光标位置到行尾 |

#### Scenario: 字符插入

- **WHEN** 用户在光标位置输入字符
- **THEN** 系统必须在光标位置插入字符
- **AND** 光标后的字符必须向右移动
- **AND** 光标必须移动到新字符之后
