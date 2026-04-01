# CLI / REPL 交互界面

MystiSql 默认启动 MySQL 风格的 REPL (Read-Eval-Print-Loop) 交互界面，使用 `golang.org/x/term` 实现终端原始模式。

## REPL 快捷键

| 快捷键 | 功能 |
|--------|------|
| Enter | 执行 SQL 或命令 |
| Tab | 自动补全（未来功能） |
| Ctrl+C | 中断当前输入/退出 |
| ↑/↓ | 浏览历史命令 |
| Ctrl+A | 移动到行首 |
| Ctrl+E | 移动到行尾 |
| Ctrl+U | 删除到行首 |
| Ctrl+K | 删除整行 |
| Esc | 取消当前操作 |

## 提示符

- 新语句: `mystisql@instance>`
- 续行: `    ->`、`    '>`、`    ">`、`    ``>`（反引号）
- 注释后: `    -- ` 或 `#`

## 内置命令

| 命令 | 描述 |
|------|------|
| `exit`/`quit`/`\q` | 退出 REPL |
| `help`/`?`/`\h` | 显示帮助 |
| `clear`/`\c` | 清除当前输入 |
| `status`/`\s` | 显示状态 |
| `print`/`\p` | 打印当前输入 |
| `edit`/`\e` | 编辑当前输入 ($EDITOR) |
| `ego`/`\G` | 查询并垂直显示结果 |
| `go`/`\g` | 执行查询 |
| `use`/`instance` | 切换实例 |
| `prompt`/`\R` | 自定义提示符 |
| `source`/`\.` | 执行脚本文件 |
| `system`/`\!` | 执行系统命令 |
| `output`/`\o` | 设置输出格式 (csv, json) |

## 输出格式

- 表格格式（默认）: ASCII 表格，对齐
- 垂直格式 (`\G`): 每行一列显示
- CSV 格式 (`\o csv`)
- JSON 格式 (`\o json`)

## query 子命令

直接执行 SQL（不进入 REPL）:

```bash
mystisql query --instance local-mysql "SELECT * FROM users"
mystisql query "SELECT 1"  # 使用默认实例
```

## auth 子命令

```bash
# 生成 Token
mystisql auth token --user-id admin --role admin --server http://localhost:8080

# 使用 Token 查询
mystisql query --instance local-mysql "SELECT * FROM users" --token <jwt-token>

# 查看 Token 信息
mystisql auth info --token <jwt-token>

# 撤销 Token
mystisql auth revoke --token <jwt-token>
```

## 代码位置

```
internal/cli/
  root.go        # Cobra 根命令
  repl/           # REPL 核心实现
    repl.go      # REPL 核心引擎
    input.go     # 多行输入处理
    formatter.go # 输出格式化
    commands.go  # 内置命令
    history.go   # 历史管理
    readline.go  # ReadLine 支持
```

## 测试

```bash
go test -v ./internal/cli/repl/...
go test -v ./internal/cli/... -run TestREPL
go test -v ./internal/cli/... -run "TestREPLWith|TestREPLInstance"
```
