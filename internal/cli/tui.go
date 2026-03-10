package cli

import (
	"MystiSql/internal/service/query"
	"MystiSql/pkg/types"
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// TUIApp TUI 应用结构
type TUIApp struct {
	app *tea.Program
}

// NewTUIApp 创建 TUI 应用
func NewTUIApp() *TUIApp {
	// 初始化 bubbletea 应用
	app := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	return &TUIApp{
		app: app,
	}
}

// Run 运行 TUI 应用
func (a *TUIApp) Run() error {
	return a.app.Start()
}

// initialModel 创建初始模型
func initialModel() tea.Model {
	// 这里将实现 TUI 模型
	return &model{}
}

// model TUI 模型
type model struct {
	// 界面状态
	width  int
	height int
	// 输入状态
	input string
	// 结果状态
	results string
	// 错误信息
	errorMsg string
	// 当前实例
	instance string
	// 实例列表
	instances []string
	// 显示实例列表
	showInstanceList bool
	// 当前选中的实例索引
	selectedInstance int
	// 执行状态
	isExecuting bool
	// 查询引擎
	queryEngine *query.Engine
}

// Init 初始化模型
func (m *model) Init() tea.Cmd {
	// 初始化查询引擎
	m.queryEngine = query.NewEngine(GetRegistry())
	// 设置默认实例
	m.instance = "local-mysql"
	// 初始化实例列表
	m.instances = []string{"local-mysql", "local-postgres", "local-oracle", "local-redis"}
	// 设置默认选中的实例
	m.selectedInstance = 0
	return nil
}

// Update 更新模型
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// 更新窗口大小
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		// 处理键盘输入
		switch msg.String() {
		case "ctrl+c":
			// 退出应用
			return m, tea.Quit
		case "tab":
			// 显示/隐藏实例列表
			m.showInstanceList = !m.showInstanceList
		case "enter":
			// 处理输入
			if m.showInstanceList {
				// 确认选择实例
				if m.selectedInstance >= 0 && m.selectedInstance < len(m.instances) {
					m.instance = m.instances[m.selectedInstance]
					m.results = fmt.Sprintf("已切换到实例: %s", m.instance)
				}
				m.showInstanceList = false
			} else if m.input != "" {
				m.isExecuting = true
				m.errorMsg = ""

				// 执行 SQL
				sqlQuery := m.input
				m.input = ""

				// 创建带超时的上下文
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				// 判断是查询还是执行语句
				upperQuery := strings.ToUpper(strings.TrimSpace(sqlQuery))
				isQuery := strings.HasPrefix(upperQuery, "SELECT") ||
					strings.HasPrefix(upperQuery, "SHOW") ||
					strings.HasPrefix(upperQuery, "DESC") ||
					strings.HasPrefix(upperQuery, "EXPLAIN")

				if isQuery {
					// 执行查询
					queryResult, execErr := m.queryEngine.ExecuteQuery(ctx, m.instance, sqlQuery)
					if execErr != nil {
						m.errorMsg = fmt.Sprintf("执行查询失败: %v", execErr)
						m.isExecuting = false
						break
					}
					m.results = formatQueryResult(queryResult)
				} else {
					// 执行非查询语句
					execResult, execErr := m.queryEngine.ExecuteExec(ctx, m.instance, sqlQuery)
					if execErr != nil {
						m.errorMsg = fmt.Sprintf("执行语句失败: %v", execErr)
						m.isExecuting = false
						break
					}
					m.results = formatExecResult(execResult)
				}

				m.isExecuting = false
			}
		case "esc":
			// 取消选择实例
			m.showInstanceList = false
		case "up":
			// 上移选择
			if m.showInstanceList {
				if m.selectedInstance > 0 {
					m.selectedInstance--
				}
			}
		case "down":
			// 下移选择
			if m.showInstanceList {
				if m.selectedInstance < len(m.instances)-1 {
					m.selectedInstance++
				}
			}
		case "backspace":
			// 处理退格键
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			// 处理其他按键
			// 只处理单个字符的按键
			if len(msg.String()) == 1 && !m.showInstanceList {
				m.input += msg.String()
			}
		}
	}
	return m, nil
}

// 定义样式
var (
	// 顶部状态栏样式
	topBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#336699")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 2).
			Bold(true)

	// 底部状态栏样式
	bottomBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 2)

	// 输入区域样式
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(1, 2)

	// 结果区域样式
	resultsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(1, 2).
			Background(lipgloss.Color("#f5f5f5")).
			Foreground(lipgloss.Color("#333333"))
)

// formatQueryResult 格式化查询结果
func formatQueryResult(result *types.QueryResult) string {
	if len(result.Columns) == 0 {
		return "查询成功，但没有返回数据"
	}

	var output strings.Builder

	// 打印表头
	for i, col := range result.Columns {
		if i > 0 {
			output.WriteString("\t")
		}
		output.WriteString(col.Name)
	}
	output.WriteString("\n")

	// 打印分隔线
	for i, col := range result.Columns {
		if i > 0 {
			output.WriteString("\t")
		}
		output.WriteString(strings.Repeat("-", len(col.Name)))
	}
	output.WriteString("\n")

	// 打印数据
	for _, row := range result.Rows {
		for i, val := range row {
			if i > 0 {
				output.WriteString("\t")
			}
			if val == nil {
				output.WriteString("NULL")
			} else {
				output.WriteString(fmt.Sprintf("%v", val))
			}
		}
		output.WriteString("\n")
	}

	// 打印统计信息
	output.WriteString(fmt.Sprintf("\n%d 行数据，执行时间: %v", result.RowCount, result.ExecutionTime))

	return output.String()
}

// formatExecResult 格式化执行结果
func formatExecResult(result *types.ExecResult) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("受影响行数: %d\n", result.RowsAffected))
	if result.LastInsertID > 0 {
		output.WriteString(fmt.Sprintf("最后插入ID: %d\n", result.LastInsertID))
	}
	output.WriteString(fmt.Sprintf("执行时间: %v", result.ExecutionTime))

	return output.String()
}

// View 渲染视图
func (m *model) View() string {
	// 计算各区域高度
	contentHeight := m.height - 4 // 减去顶部和底部状态栏
	_ = contentHeight             // 避免未使用的变量警告
	inputHeight := 5
	_ = inputHeight // 避免未使用的变量警告

	// 顶部状态栏
	topBar := topBarStyle.Render(fmt.Sprintf("MystiSql TUI | 当前实例: %s", m.instance))

	// 结果区域
	resultContent := m.results
	if m.errorMsg != "" {
		resultContent = "错误: " + m.errorMsg
	} else if m.isExecuting {
		resultContent = "执行中..."
	}

	// 实例列表
	var instanceList string
	if m.showInstanceList {
		var listContent strings.Builder
		listContent.WriteString("可用实例:\n")
		for i, instance := range m.instances {
			if i == m.selectedInstance {
				listContent.WriteString(fmt.Sprintf("→ %s\n", instance))
			} else {
				listContent.WriteString(fmt.Sprintf("  %s\n", instance))
			}
		}
		listContent.WriteString("\n按 Enter 选择，按 Esc 取消")
		instanceList = resultsStyle.Render(listContent.String())
	}

	// 输入区域
	input := inputStyle.Render(fmt.Sprintf("SQL> %s", m.input))

	// 底部状态栏
	bottomBar := bottomBarStyle.Render("按 Enter 执行 SQL | 按 Tab 切换实例 | 按 Ctrl+C 退出")

	// 组合所有区域
	var mainContent string
	if m.showInstanceList {
		mainContent = lipgloss.JoinVertical(
			lipgloss.Top,
			instanceList,
			input,
		)
	} else {
		mainContent = lipgloss.JoinVertical(
			lipgloss.Top,
			resultsStyle.Render(resultContent),
			input,
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		topBar,
		mainContent,
		bottomBar,
	)
}

// NewTUICmd 创建 TUI 命令
func NewTUICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "启动交互式 TUI 界面",
		Long:  `启动交互式 TUI 界面，支持 SQL 执行、结果显示和实例切换`,
		Run: func(cmd *cobra.Command, args []string) {
			// 启动 TUI 应用
			app := NewTUIApp()
			if err := app.Run(); err != nil {
				cmd.Println("启动 TUI 界面失败:", err)
			}
		},
	}

	// 添加 --instance 参数
	cmd.Flags().StringP("instance", "i", "", "指定初始数据库实例")

	return cmd
}
