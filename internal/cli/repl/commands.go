package repl

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Command struct {
	Name string
	Args string
}

type CommandParser struct {
	repl *REPL
}

func NewCommandParser(repl *REPL) *CommandParser {
	return &CommandParser{repl: repl}
}

func (p *CommandParser) ParseCommand(line string) (*Command, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, false
	}

	if strings.HasPrefix(line, "\\") {
		return p.parseShortCommand(line), true
	}

	lowerLine := strings.ToLower(line)
	switch {
	case lowerLine == "exit" || lowerLine == "quit":
		return &Command{Name: "exit"}, true
	case lowerLine == "help" || lowerLine == "?":
		return &Command{Name: "help"}, true
	case lowerLine == "clear":
		return &Command{Name: "clear"}, true
	case lowerLine == "status":
		return &Command{Name: "status"}, true
	case lowerLine == "print":
		return &Command{Name: "print"}, true
	case lowerLine == "edit":
		return &Command{Name: "edit"}, true
	case strings.HasPrefix(lowerLine, "use "):
		return &Command{Name: "use", Args: strings.TrimSpace(line[4:])}, true
	case strings.HasPrefix(lowerLine, "prompt"):
		args := ""
		if len(line) > 6 {
			args = strings.TrimSpace(line[6:])
		}
		return &Command{Name: "prompt", Args: args}, true
	case strings.HasPrefix(lowerLine, "source "):
		return &Command{Name: "source", Args: strings.TrimSpace(line[7:])}, true
	}

	return nil, false
}

func (p *CommandParser) parseShortCommand(line string) *Command {
	if len(line) < 2 {
		return nil
	}

	cmd := string(line[1])
	args := ""
	if len(line) > 2 {
		args = strings.TrimSpace(line[2:])
	}

	switch cmd {
	case "q":
		return &Command{Name: "exit"}
	case "h", "?":
		return &Command{Name: "help"}
	case "c":
		return &Command{Name: "clear"}
	case "s":
		return &Command{Name: "status"}
	case "p":
		return &Command{Name: "print"}
	case "e":
		return &Command{Name: "edit"}
	case "G":
		return &Command{Name: "ego"}
	case "g":
		return &Command{Name: "go"}
	case "R":
		return &Command{Name: "prompt", Args: args}
	case ".":
		return &Command{Name: "source", Args: args}
	case "!":
		return &Command{Name: "system", Args: args}
	case "o":
		return &Command{Name: "output", Args: args}
	case "u":
		return &Command{Name: "use", Args: args}
	default:
		return &Command{Name: "unknown", Args: line}
	}
}

func (p *CommandParser) ExecuteCommand(cmd *Command) error {
	switch cmd.Name {
	case "exit":
		return p.repl.Exit()

	case "help":
		p.showHelp()
		return nil

	case "clear":
		p.repl.CancelInput()
		return nil

	case "status":
		fmt.Print(p.repl.formatter.FormatStatus(p.repl))
		return nil

	case "print":
		p.repl.PrintCurrentInput()
		return nil

	case "edit":
		return p.editInput()

	case "ego":
		p.repl.outputMode = OutputVertical
		if !p.repl.inputBuffer.IsEmpty() {
			return p.repl.executeCurrentStatement()
		}
		return nil

	case "go":
		if !p.repl.inputBuffer.IsEmpty() {
			return p.repl.executeCurrentStatement()
		}
		return nil

	case "use":
		return p.repl.SetInstance(cmd.Args)

	case "prompt":
		if cmd.Args == "" {
			p.repl.SetPrompt("mystisql@%i> ")
			fmt.Print("Returning to default PROMPT of mystisql@%i>\r\n")
		} else {
			p.repl.SetPrompt(cmd.Args)
		}
		return nil

	case "source":
		return p.sourceFile(cmd.Args)

	case "system":
		return p.executeSystemCommand(cmd.Args)

	case "output":
		return p.setOutput(cmd.Args)

	case "unknown":
		return fmt.Errorf("unknown command: %s", cmd.Args)

	default:
		return fmt.Errorf("unhandled command: %s", cmd.Name)
	}
}

func (p *CommandParser) showHelp() {
	help := `
List of all MystiSql commands:
Note that all text commands must be first on line and end with ';'

?         (\?) Synonym for 'help'.
clear     (\c) Clear the current input statement.
edit      (\e) Edit command with $EDITOR.
ego       (\G) Send command to mystisql server, display result vertically.
exit      (\q) Exit mystisql. Same as quit.
go        (\g) Send command to mystisql server.
help      (\h) Display this help.
output    (\o) Set output format (csv, json). Use without argument to reset.
print     (\p) Print current command.
prompt    (\R) Change your mystisql prompt.
quit      (\q) Quit mystisql.
source    (\.) Execute an SQL script file. Takes a file name as an argument.
status    (\s) Get status information from the server.
system    (\!) Execute a system shell command.
use       (\u) Use another database instance. Takes instance name as argument.

For server side help, type 'help contents'
`
	fmt.Print(strings.ReplaceAll(help, "\n", "\r\n"))
}

func (p *CommandParser) editInput() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	tmpFile, err := os.CreateTemp("", "mystisql-*.sql")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	currentInput := p.repl.inputBuffer.GetSQL()
	if _, err := tmpFile.WriteString(currentInput); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	editedContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	p.repl.inputBuffer.Reset()
	lines := strings.Split(string(editedContent), "\r\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" || len(lines) == 1 {
			p.repl.inputBuffer.Append(line)
		}
	}

	// 回显编辑后的内容
	if !p.repl.inputBuffer.IsEmpty() {
		fmt.Print("Edit result:\r\n")
		fmt.Print(p.repl.inputBuffer.GetSQL() + "\r\n")
	}

	return nil
}

func (p *CommandParser) sourceFile(filename string) error {
	if filename == "" {
		return fmt.Errorf("source command requires a file name")
	}

	expandedPath := os.ExpandEnv(filename)
	if strings.HasPrefix(expandedPath, "~/") {
		homeDir, _ := os.UserHomeDir()
		expandedPath = filepath.Join(homeDir, expandedPath[2:])
	}

	file, err := os.Open(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}

		if cmd, isCommand := p.ParseCommand(line); isCommand {
			if err := p.ExecuteCommand(cmd); err != nil {
				if err == ErrExit {
					return err
				}
				fmt.Fprintf(os.Stderr, "ERROR: %v\r\n", err)
			}
			continue
		}

		p.repl.inputBuffer.Append(line)
		if p.repl.inputBuffer.IsComplete() {
			if err := p.repl.executeCurrentStatement(); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\r\n", err)
			}
		}
	}

	return scanner.Err()
}

func (p *CommandParser) executeSystemCommand(command string) error {
	if command == "" {
		return fmt.Errorf("system command requires an argument")
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (p *CommandParser) setOutput(format string) error {
	if format == "" {
		p.repl.outputMode = OutputNormal
		p.repl.exportFormat = ""
		fmt.Print("Output format reset to normal\r\n")
		return nil
	}

	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "csv":
		p.repl.outputMode = OutputExport
		p.repl.exportFormat = "csv"
		fmt.Print("Output format set to CSV\r\n")
	case "json":
		p.repl.outputMode = OutputExport
		p.repl.exportFormat = "json"
		fmt.Print("Output format set to JSON\r\n")
	default:
		return fmt.Errorf("unknown output format: %s (supported: csv, json)", format)
	}
	return nil
}
