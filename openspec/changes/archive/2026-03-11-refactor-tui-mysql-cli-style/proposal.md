## Why

当前使用 Bubble Tea (tea) 框架实现的 TUI 界面虽然功能完整，但用户体验与标准 MySQL CLI 差异较大。Bubble Tea 的全屏渲染模式和复杂组件模型不适合实现传统 REPL 风格的命令行界面，导致用户需要重新学习操作方式，增加了使用门槛。重构为更接近 MySQL CLI 的简洁风格，可以降低用户学习成本，提供更熟悉的交互体验。

## What Changes

- **BREAKING**: 移除 Bubble Tea 框架依赖，改用标准库 `bufio.Scanner` + `term` 实现 REPL 风格界面
- **BREAKING**: 移除全屏 TUI 渲染模式，改用流式输入输出
- **BREAKING**: 移除复杂的窗口布局（状态栏、结果区、输入区分离），改为简洁的提示符 + 输出模式
- 重构输入处理逻辑，支持真正的多行 SQL 输入（以分号结尾）
- 重构输出格式，完全模仿 MySQL CLI 的表格输出风格
- 保留核心功能：SQL 执行、实例切换、历史命令、结果导出
- 简化快捷键设计，与 MySQL CLI 保持一致

## Capabilities

### New Capabilities

- `mysql-cli-style-repl`: 新的 REPL 风格命令行界面，模仿 MySQL CLI 的交互体验
- `multiline-sql-input`: 支持真正的多行 SQL 输入（以分号作为语句结束符）
- `streaming-output`: 流式输出模式，直接打印结果而非全屏渲染

### Modified Capabilities

- `cli-tui`: 修改界面布局要求，从复杂多区域布局改为简洁 REPL 风格
- `sql-execution-tui`: 修改结果显示方式，从分页表格改为流式表格输出
- `instance-switching-tui`: 修改实例切换交互方式，从列表选择改为命令行参数或 USE 语句

## Impact

**代码变更**:
- `internal/cli/tui.go` - 完全重写，移除 Bubble Tea 依赖
- `internal/cli/tui_*.go` - 测试文件需要相应更新
- `go.mod` - 移除 `github.com/charmbracelet/bubbletea` 和 `github.com/charmbracelet/lipgloss` 依赖

**API 变更**:
- `NewTUIApp()` 函数签名保持不变，但内部实现完全改变
- `TUIApp.Run()` 行为变化：从全屏 TUI 改为 REPL 循环

**依赖变更**:
- 移除: `github.com/charmbracelet/bubbletea`
- 移除: `github.com/charmbracelet/lipgloss`
- 可能新增: `golang.org/x/term` (用于终端原始模式)

**向后兼容性**:
- 所有 CLI 命令行参数保持不变
- 配置文件格式保持不变
- 用户交互方式发生变化（更接近 MySQL CLI）
