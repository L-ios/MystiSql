## 1. 项目结构与依赖管理 (4 tasks)

- [x] 1.1 创建 `internal/cli/repl/` 目录结构
- [x] 1.2 添加 `golang.org/x/term` 依赖
- [x] 1.3 移除 bubbletea 和 lipgloss 依赖
- [x] 1.4 更新 go.mod 并运行 go mod tidy

## 2. REPL 核心引擎 (5 tasks)

- [x] 2.1 实现 `REPL` 结构体和基本框架 (`internal/cli/repl/repl.go`)
- [x] 2.2 实现提示符显示功能 (`mystisql@instance>`)
- [x] 2.3 实现欢迎信息显示
- [x] 2.4 实现主循环（Read-Eval-Print-Loop）
- [x] 2.5 实现终端原始模式切换（使用 golang.org/x/term）

## 3. 多行输入处理 (6 tasks)

- [x] 3.1 实现 SQL 语句缓冲管理 (`internal/cli/repl/input.go`)
- [x] 3.2 实现分号结束符检测
- [x] 3.3 实现续行提示符（`->`, `'>`, `">`, `` `> ``）
- [x] 3.4 实现字符串和标识符边界检测（单引号、双引号、反引号）
- [x] 3.5 实现 Ctrl+C 取消当前输入功能
- [x] 3.6 实现注释处理（`--`, `#`, `/* */`）

## 4. 输出格式化 (7 tasks)

- [x] 4.1 实现表格格式输出 (`internal/cli/repl/formatter.go`)
- [x] 4.2 实现列宽自动计算
- [x] 4.3 实现 NULL 值显示
- [x] 4.4 实现结果统计信息显示（行数、执行时间）
- [x] 4.5 实现错误输出格式化
- [x] 4.6 实现空结果集显示
- [x] 4.7 实现 \G 垂直输出格式

## 5. 内置命令 (12 tasks)

- [x] 5.1 实现命令解析器 (`internal/cli/repl/commands.go`)
- [x] 5.2 实现 `exit`/`quit`/`\q` 退出命令
- [x] 5.3 实现 `help`/`?`/`\h` 帮助命令
- [x] 5.4 实现 `clear`/`\c` 清除当前输入命令
- [x] 5.5 实现 `print`/`\p` 打印当前输入命令
- [x] 5.6 实现 `edit`/`\e` 编辑当前输入命令（调用 $EDITOR）
- [x] 5.7 实现 `status`/`\s` 状态命令
- [x] 5.8 实现 `ego`/`\G` 垂直输出命令
- [x] 5.9 实现 `use` 实例切换命令 (USE <instance>)
- [x] 5.10 实现 `prompt`/`\R` 自定义提示符命令
- [x] 5.11 实现 `source`/`\.` 执行脚本文件命令
- [x] 5.12 实现 `system`/`\!` 执行系统命令

## 6. 结果导出功能 (3 tasks)

- [x] 6.1 实现 `\o csv` 导出为 CSV 格式
- [x] 6.2 实现 `\o json` 导出为 JSON 格式
- [x] 6.3 实现导出状态提示

## 7. 历史命令管理 (5 tasks)

- [x] 7.1 实现历史记录管理器 (`internal/cli/repl/history.go`)
- [x] 7.2 实现历史文件读写（`~/.mystisql/history`）
- [x] 7.3 实现上下箭头浏览历史 (`internal/cli/repl/readline.go`)
- [x] 7.4 实现历史去重
- [x] 7.5 实现历史持久化

## 8. 错误处理 (4 tasks)

- [x] 8.1 实现 SQL 执行错误处理
- [x] 8.2 实现连接错误处理
- [x] 8.3 实现超时错误处理
- [x] 8.4 实现未知命令错误提示

## 9. 集成与迁移 (2 tasks)

- [x] 9.1 更新 `internal/cli/root.go` 使用新的 REPL
- [x] 9.2 删除旧的 TUI 代码 (`tui.go` 及相关测试)

## 10. 测试 (7 tasks)

- [x] 10.1 编写 REPL 核心功能单元测试
- [x] 10.2 编写多行输入处理测试
- [x] 10.3 编写输出格式化测试
- [x] 10.4 编写内置命令测试
- [x] 10.5 编写历史管理测试
- [x] 10.6 编写集成测试
- [x] 10.7 运行所有测试确保通过

## 11. 文档与清理 (5 tasks)

- [x] 11.1 更新 AGENTS.md 中的 TUI 相关说明
- [x] 11.2 更新 README.md 中的使用说明
- [x] 11.3 删除旧的 TUI 测试文件
- [x] 11.4 运行 golangci-lint 确保代码质量
- [x] 11.5 运行 go fmt 格式化代码
