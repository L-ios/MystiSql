package cli

import (
	"MystiSql/internal/discovery"
	"MystiSql/internal/service/query"
	"MystiSql/pkg/types"
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TUIApp TUI 应用结构
type TUIApp struct {
	app *tea.Program
}

// NewTUIApp 创建 TUI 应用
func NewTUIApp(cfg *types.Config, reg discovery.InstanceRegistry) *TUIApp {
	// 初始化 bubbletea 应用（不使用 AltScreen，保持简单界面）
	app := tea.NewProgram(
		initialModel(cfg, reg),
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
func initialModel(cfg *types.Config, reg discovery.InstanceRegistry) tea.Model {
	// 这里将实现 TUI 模型
	return &model{
		config:   cfg,
		registry: reg,
	}
}

// model TUI 模型
type model struct {
	// 配置和依赖
	config   *types.Config
	registry discovery.InstanceRegistry
	// 界面状态
	width  int
	height int
	// 输入状态
	input string
	// 光标位置
	cursor int
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
	// 命令历史
	history []string
	// 当前历史索引
	historyIndex int
	// 导出格式
	exportFormat string
	// 显示导出选项
	showExportOptions bool
	// 当前选中的导出格式索引
	selectedExportFormat int
	// 显示帮助
	showHelp bool
	// 最小窗口尺寸警告
	minSizeWarning bool
	// 是否已显示欢迎信息
	welcomeShown bool
}

// Init 初始化模型
func (m *model) Init() tea.Cmd {
	// 初始化查询引擎
	m.queryEngine = query.NewEngine(m.registry)

	// 从注册中心获取实例列表
	instances, err := m.registry.ListInstances()
	if err != nil || len(instances) == 0 {
		// 如果获取失败或没有实例，设置默认值
		m.instances = []string{}
		m.instance = ""
		m.errorMsg = "警告: 未配置任何数据库实例，请先在配置文件中添加实例"
	} else {
		// 提取实例名称
		m.instances = make([]string, len(instances))
		for i, inst := range instances {
			m.instances[i] = inst.Name
		}
		// 设置默认实例为第一个
		m.instance = instances[0].Name
		m.selectedInstance = 0
	}

	// 初始化命令历史（如果未设置）
	if m.history == nil {
		m.history = []string{}
	}
	// 初始化历史索引（如果未设置）
	if m.historyIndex == 0 {
		m.historyIndex = -1
	}
	// 初始化导出格式
	m.exportFormat = "csv"
	// 初始化导出选项显示
	m.showExportOptions = false
	// 初始化选中的导出格式索引
	m.selectedExportFormat = 0
	// 初始化光标位置
	m.cursor = 0
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
		// 如果显示帮助，按任意键关闭
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// 处理键盘输入
		switch msg.String() {
		case "ctrl+c":
			// 退出应用
			return m, tea.Quit
		case "tab":
			// 显示/隐藏实例列表
			m.showInstanceList = !m.showInstanceList
		case "ctrl+e":
			// 显示/隐藏导出选项
			m.showExportOptions = !m.showExportOptions
		case "?":
			// 显示/隐藏帮助
			m.showHelp = !m.showHelp
		case "enter":
			// 处理输入
			if m.showInstanceList {
				// 确认选择实例
				if m.selectedInstance >= 0 && m.selectedInstance < len(m.instances) {
					m.instance = m.instances[m.selectedInstance]
					m.results = ""
				}
				m.showInstanceList = false
			} else if m.showExportOptions {
				// 确认选择导出格式
				exportFormats := []string{"csv", "json", "table"}
				if m.selectedExportFormat >= 0 && m.selectedExportFormat < len(exportFormats) {
					m.exportFormat = exportFormats[m.selectedExportFormat]
					m.results = fmt.Sprintf("已设置导出格式: %s", m.exportFormat)
				}
				m.showExportOptions = false
			} else if m.input != "" {
				m.isExecuting = true
				m.errorMsg = ""

				// 执行 SQL
				sqlQuery := m.input

				// 添加到命令历史
				m.history = append(m.history, sqlQuery)
				// 重置历史索引
				m.historyIndex = -1

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
			} else if m.showExportOptions {
				// 选择导出格式
				if m.selectedExportFormat > 0 {
					m.selectedExportFormat--
				}
			} else {
				// 浏览历史命令
				if len(m.history) > 0 {
					if m.historyIndex == -1 {
						m.historyIndex = len(m.history) - 1
					} else if m.historyIndex > 0 {
						m.historyIndex--
					}
					m.input = m.history[m.historyIndex]
				}
			}
		case "down":
			// 下移选择
			if m.showInstanceList {
				if m.selectedInstance < len(m.instances)-1 {
					m.selectedInstance++
				}
			} else if m.showExportOptions {
				// 选择导出格式
				if m.selectedExportFormat < 2 {
					m.selectedExportFormat++
				}
			} else {
				// 浏览历史命令
				if len(m.history) > 0 {
					if m.historyIndex == len(m.history)-1 {
						m.historyIndex = -1
						m.input = ""
					} else if m.historyIndex >= 0 {
						m.historyIndex++
						m.input = m.history[m.historyIndex]
					}
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
	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000"))

	instanceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00bfff"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)
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

// highlightSQL 高亮 SQL 代码
func highlightSQL(sql string) string {
	return sql
}

type layout struct {
	topBar    int
	results   int
	input     int
	bottomBar int
}

func (m *model) calculateLayout() layout {
	l := layout{
		topBar:    1,
		bottomBar: 1,
		input:     5,
	}

	if m.height <= 0 {
		return l
	}

	availableHeight := m.height - l.topBar - l.bottomBar
	if availableHeight > l.input {
		l.results = availableHeight - l.input
	} else {
		l.results = 1
	}

	return l
}

// View 渲染视图
func (m *model) View() string {
	var view strings.Builder

	if !m.welcomeShown {
		view.WriteString("\n")
		view.WriteString(promptStyle.Render("Welcome to the MystiSql monitor."))
		view.WriteString(" Commands end with Enter.")
		view.WriteString("\n")

		instanceCount := len(m.instances)
		view.WriteString(fmt.Sprintf("Your MystiSql connection has %d instance(s) configured.\n", instanceCount))

		if m.instance != "" {
			view.WriteString(hintStyle.Render("Current instance: ") + instanceStyle.Render(m.instance))
			view.WriteString("\n")
		}

		view.WriteString("\n")
		view.WriteString(hintStyle.Render("Type 'help' or '?' for help. Type 'exit' or Ctrl+C to quit."))
		view.WriteString("\n\n")

		m.welcomeShown = true
	}

	if m.showHelp {
		view.WriteString("\n")
		view.WriteString(hintStyle.Render("快捷键:"))
		view.WriteString("\n")
		view.WriteString("  Enter      - 执行 SQL\n")
		view.WriteString("  Tab        - 切换实例\n")
		view.WriteString("  Ctrl+E     - 导出结果\n")
		view.WriteString("  Ctrl+C     - 退出\n")
		view.WriteString("  ↑/↓        - 浏览历史\n")
		view.WriteString("\n")
		view.WriteString(hintStyle.Render("按任意键关闭帮助"))
		view.WriteString("\n\n")
		return view.String()
	}

	if m.showInstanceList {
		view.WriteString("\n")
		view.WriteString(instanceStyle.Render("可用实例:"))
		view.WriteString("\n")
		for i, instance := range m.instances {
			if i == m.selectedInstance {
				view.WriteString(fmt.Sprintf("→ %s\n", instance))
			} else {
				view.WriteString(fmt.Sprintf("  %s\n", instance))
			}
		}
		view.WriteString("\n")
		view.WriteString(hintStyle.Render("按 Enter 选择，按 Esc 取消"))
		view.WriteString("\n\n")
		return view.String()
	}

	if m.showExportOptions {
		view.WriteString("\n")
		view.WriteString(instanceStyle.Render("导出格式:"))
		view.WriteString("\n")
		exportFormats := []string{"csv", "json", "table"}
		for i, format := range exportFormats {
			if i == m.selectedExportFormat {
				view.WriteString(fmt.Sprintf("→ %s\n", format))
			} else {
				view.WriteString(fmt.Sprintf("  %s\n", format))
			}
		}
		view.WriteString("\n")
		view.WriteString(hintStyle.Render("按 Enter 选择，按 Esc 取消"))
		view.WriteString("\n\n")
		return view.String()
	}

	if m.results != "" {
		view.WriteString(m.results)
		view.WriteString("\n\n")
	}

	if m.errorMsg != "" {
		view.WriteString(errorStyle.Render("ERROR: " + m.errorMsg))
		view.WriteString("\n\n")
	}

	if m.isExecuting {
		view.WriteString(hintStyle.Render("执行中..."))
		view.WriteString("\n\n")
	}

	if m.instance != "" {
		view.WriteString(promptStyle.Render("mystisql"))
		view.WriteString("@")
		view.WriteString(instanceStyle.Render(m.instance))
		view.WriteString("> ")
	} else {
		view.WriteString(promptStyle.Render("mystisql"))
		view.WriteString("> ")
	}

	view.WriteString(highlightSQL(m.input))
	view.WriteString("_")

	return view.String()
}
