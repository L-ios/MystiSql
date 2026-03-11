## Why

当前 MystiSql REPL 的行编辑功能较为基础，只支持在行尾追加字符和退格删除最后一个字符。用户无法使用左右方向键移动光标到行中间进行插入或修改，这大大降低了 SQL 编辑效率。MySQL 原生命令行客户端支持完整的行编辑功能，MystiSql 应该提供同等体验。

## What Changes

- 新增左右方向键光标移动功能
- 新增在光标位置插入字符功能（而非仅追加到行尾）
- 新增 Delete 键删除光标位置字符功能
- 新增 Home/End 键快速跳转到行首/行尾
- 新增 Ctrl+A/Ctrl+E 快捷键（行首/行尾）
- 新增 Ctrl+U 删除到行首、Ctrl+K 删除到行尾
- 优化光标位置的可视化显示

## Capabilities

### New Capabilities

- `repl-line-editing`: REPL 行编辑增强功能，支持光标移动、插入、删除等完整行编辑能力

### Modified Capabilities

- `cli-tui`: 更新快捷键说明，增加新的行编辑快捷键

## Impact

- **代码影响**: `internal/cli/repl/readline.go` 需要重构以支持光标位置追踪和插入操作
- **API 影响**: 无外部 API 变化
- **依赖影响**: 无新增依赖，继续使用 `golang.org/x/term`
- **兼容性**: 完全向后兼容，现有功能不受影响
