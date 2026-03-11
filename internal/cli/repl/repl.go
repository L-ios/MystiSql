package repl

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"MystiSql/internal/discovery"
	"MystiSql/internal/service/query"
	"MystiSql/pkg/types"

	"golang.org/x/term"
)

type REPL struct {
	config   *types.Config
	registry discovery.InstanceRegistry
	engine   *query.Engine

	oldState      *term.State
	readline      *ReadLine
	inputBuffer   *InputBuffer
	history       *HistoryManager
	formatter     *Formatter
	commandParser *CommandParser

	currentInstance string
	instances       []*types.DatabaseInstance
	prompt          string
	outputMode      OutputMode
	exportFormat    string

	running bool
}

type OutputMode int

const (
	OutputNormal OutputMode = iota
	OutputVertical
	OutputExport
)

func NewREPL(cfg *types.Config, reg discovery.InstanceRegistry) *REPL {
	r := &REPL{
		config:      cfg,
		registry:    reg,
		engine:      query.NewEngine(reg),
		inputBuffer: NewInputBuffer(),
		history:     NewHistoryManager(),
		formatter:   NewFormatter(),
		prompt:      "mystisql@%i> ",
		outputMode:  OutputNormal,
	}
	r.commandParser = NewCommandParser(r)
	return r
}

func (r *REPL) Run() error {
	fd := int(os.Stdin.Fd())
	var err error
	r.oldState, err = term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer term.Restore(fd, r.oldState)

	r.readline = NewReadLine(os.Stdin, os.Stdout, r.history)
	r.running = true
	defer func() { r.running = false }()

	r.initialize()
	r.showWelcome()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		r.running = false
	}()

	return r.mainLoop()
}

func (r *REPL) initialize() {
	instances, err := r.registry.ListInstances()
	if err == nil && len(instances) > 0 {
		r.instances = instances
		r.currentInstance = instances[0].Name
	} else {
		r.instances = nil
		r.currentInstance = ""
	}
	r.history.Load()
}

func (r *REPL) showWelcome() {
	fmt.Print("\r\n")
	fmt.Print("Welcome to the MystiSql monitor. Commands end with ; or \\g.\r\n")
	fmt.Printf("Your MystiSql connection has %d instance(s) configured.\r\n", len(r.instances))
	if r.currentInstance != "" {
		fmt.Printf("Current instance: %s\r\n", r.currentInstance)
	}
	fmt.Print("\r\n")
	fmt.Print("Type 'help' or '\\h' for help. Type '\\c' to clear current input statement.\r\n")
	fmt.Print("\r\n")
}

func (r *REPL) mainLoop() error {
	for r.running {
		prompt := r.getPrompt()
		line, err := r.readline.ReadLine(prompt)
		if err != nil {
			fmt.Println("Bye")
			return nil
		}
		if err := r.processLine(line); err != nil {
			if err == ErrExit {
				fmt.Println("Bye")
				return nil
			}
			fmt.Fprintf(os.Stderr, "ERROR: %v\r\n", err)
		}
	}
	return nil
}

func (r *REPL) getPrompt() string {
	if r.inputBuffer.IsEmpty() {
		return r.formatPrompt(r.prompt)
	}
	return r.inputBuffer.GetContinuePrompt()
}

func (r *REPL) formatPrompt(template string) string {
	result := strings.ReplaceAll(template, "%i", r.currentInstance)
	result = strings.ReplaceAll(result, "\\i", r.currentInstance)
	return result
}

func (r *REPL) processLine(line string) error {
	if r.inputBuffer.IsEmpty() {
		if cmd, isCommand := r.commandParser.ParseCommand(line); isCommand {
			return r.commandParser.ExecuteCommand(cmd)
		}
	}
	if strings.HasSuffix(strings.TrimSpace(line), "\\G") {
		sql := strings.TrimSuffix(strings.TrimSpace(line), "\\G")
		r.inputBuffer.Append(sql)
		r.outputMode = OutputVertical
		return r.executeCurrentStatement()
	}
	r.inputBuffer.Append(line)
	if r.inputBuffer.IsComplete() {
		return r.executeCurrentStatement()
	}
	return nil
}

func (r *REPL) executeCurrentStatement() error {
	sql := r.inputBuffer.GetSQL()
	r.inputBuffer.Reset()
	if strings.TrimSpace(sql) == "" {
		return nil
	}
	r.history.Add(sql)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	upperSQL := strings.ToUpper(strings.TrimSpace(sql))
	isQuery := strings.HasPrefix(upperSQL, "SELECT") ||
		strings.HasPrefix(upperSQL, "SHOW") ||
		strings.HasPrefix(upperSQL, "DESC") ||
		strings.HasPrefix(upperSQL, "DESCRIBE") ||
		strings.HasPrefix(upperSQL, "EXPLAIN")
	if isQuery {
		result, err := r.engine.ExecuteQuery(ctx, r.currentInstance, sql)
		if err != nil {
			r.outputMode = OutputNormal
			return fmt.Errorf("query failed: %w", err)
		}
		r.formatOutput(result)
	} else {
		result, err := r.engine.ExecuteExec(ctx, r.currentInstance, sql)
		if err != nil {
			r.outputMode = OutputNormal
			return fmt.Errorf("execution failed: %w", err)
		}
		r.formatExecOutput(result)
	}
	r.outputMode = OutputNormal
	return nil
}

func (r *REPL) formatOutput(result *types.QueryResult) {
	switch r.outputMode {
	case OutputVertical:
		fmt.Print(r.formatter.FormatVertical(result))
	case OutputExport:
		r.exportResult(result)
	default:
		fmt.Print(r.formatter.FormatTable(result))
	}
}

func (r *REPL) formatExecOutput(result *types.ExecResult) {
	fmt.Printf("Query OK, %d row(s) affected (%.3f sec)\n", result.RowsAffected, result.ExecutionTime.Seconds())
	if result.LastInsertID > 0 {
		fmt.Printf("Last insert ID: %d\n", result.LastInsertID)
	}
	fmt.Println()
}

func (r *REPL) exportResult(result *types.QueryResult) {
	switch r.exportFormat {
	case "csv":
		fmt.Print(r.formatter.FormatCSV(result))
	case "json":
		fmt.Print(r.formatter.FormatJSON(result))
	default:
		fmt.Print(r.formatter.FormatTable(result))
	}
}

func (r *REPL) SetInstance(name string) error {
	for _, inst := range r.instances {
		if inst.Name == name {
			r.currentInstance = name
			fmt.Printf("Database changed to %s\n", name)
			return nil
		}
	}
	return fmt.Errorf("unknown instance: %s", name)
}

func (r *REPL) GetInstance() string { return r.currentInstance }

func (r *REPL) ListInstances() []*types.DatabaseInstance { return r.instances }

func (r *REPL) SetPrompt(template string) {
	r.prompt = template
	fmt.Printf("PROMPT set to '%s'\n", template)
}

func (r *REPL) GetPrompt() string { return r.prompt }

func (r *REPL) SetOutputMode(mode OutputMode, format string) {
	r.outputMode = mode
	r.exportFormat = format
	if format != "" {
		fmt.Printf("Output format set to %s\n", format)
	}
}

func (r *REPL) CancelInput() {
	r.inputBuffer.Reset()
	fmt.Println()
}

func (r *REPL) PrintCurrentInput() {
	sql := r.inputBuffer.GetSQL()
	if sql == "" {
		fmt.Println("(no input)")
	} else {
		fmt.Println(sql)
	}
}

func (r *REPL) Exit() error {
	r.history.Save()
	return ErrExit
}

var ErrExit = fmt.Errorf("exit")
