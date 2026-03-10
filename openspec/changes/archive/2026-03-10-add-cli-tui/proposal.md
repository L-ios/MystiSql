## Why

为 MystiSql CLI 添加 TUI 功能，以支持长期运行 SQL 查询和提供类似原生数据库客户端的交互式体验，同时允许用户在不同数据库实例之间切换。

## What Changes

- 添加新的 `tui` 命令到 CLI，启动交互式终端界面
- 支持基本 SQL 执行和结果显示
- 支持长期运行的 SQL 查询，显示实时进度
- 提供类似 MySQL TUI 的交互式会话功能
- 允许在 TUI 中切换不同的数据库实例
- 支持命令历史记录和 SQL 语法高亮

## Capabilities

### New Capabilities
- `cli-tui`: 为 CLI 添加 TUI 界面，支持交互式 SQL 执行和实例切换
- `sql-execution-tui`: 在 TUI 中执行 SQL 查询，支持长期运行的查询
- `instance-switching-tui`: 在 TUI 中切换不同的数据库实例

### Modified Capabilities

## Impact

- 影响 `cmd/mystisql/cli/` 目录下的 CLI 实现
- 需要添加新的依赖库来支持 TUI 界面
- 可能需要修改现有的 CLI 命令结构
- 不影响现有的 REST API 和 WebSocket 功能