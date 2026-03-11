package repl

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestInputBuffer_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected bool
	}{
		{"empty buffer", []string{}, true},
		{"single line", []string{"SELECT 1"}, false},
		{"multiple lines", []string{"SELECT *", "FROM users"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewInputBuffer()
			for _, line := range tt.lines {
				buf.Append(line)
			}
			if got := buf.IsEmpty(); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInputBuffer_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected bool
	}{
		{"semicolon end", []string{"SELECT 1;"}, true},
		{"backslash g end", []string{"SELECT 1\\g"}, true},
		{"no end", []string{"SELECT 1"}, false},
		{"multiline with semicolon", []string{"SELECT *", "FROM users", "WHERE id = 1;"}, true},
		{"semicolon in string", []string{"SELECT ';' "}, false},
		{"semicolon in string then end", []string{"SELECT ';';"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewInputBuffer()
			for _, line := range tt.lines {
				buf.Append(line)
			}
			if got := buf.IsComplete(); got != tt.expected {
				t.Errorf("IsComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInputBuffer_GetContinuePrompt(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"normal continue", "SELECT *", "    -> "},
		{"single quote", "SELECT '", "    '> "},
		{"double quote", "SELECT \"", "    \"> "},
		{"backtick", "SELECT `", "    `> "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewInputBuffer()
			buf.Append(tt.line)
			if got := buf.GetContinuePrompt(); got != tt.expected {
				t.Errorf("GetContinuePrompt() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestInputBuffer_GetSQL(t *testing.T) {
	buf := NewInputBuffer()
	buf.Append("SELECT *")
	buf.Append("FROM users")

	expected := "SELECT *\nFROM users"
	if got := buf.GetSQL(); got != expected {
		t.Errorf("GetSQL() = %q, want %q", got, expected)
	}
}

func TestInputBuffer_Reset(t *testing.T) {
	buf := NewInputBuffer()
	buf.Append("SELECT 1")
	buf.Reset()

	if !buf.IsEmpty() {
		t.Error("Reset() did not clear buffer")
	}
}

func TestHistoryManager_Add(t *testing.T) {
	h := NewHistoryManager()

	h.Add("SELECT 1")
	h.Add("SELECT 2")
	h.Add("SELECT 1") // duplicate

	if h.Count() != 2 {
		t.Errorf("Count() = %d, want 2", h.Count())
	}
}

func TestHistoryManager_Get(t *testing.T) {
	h := NewHistoryManager()
	h.Add("SELECT 1")
	h.Add("SELECT 2")

	if got := h.Get(0); got != "SELECT 1" {
		t.Errorf("Get(0) = %q, want %q", got, "SELECT 1")
	}

	if got := h.Get(1); got != "SELECT 2" {
		t.Errorf("Get(1) = %q, want %q", got, "SELECT 2")
	}

	if got := h.Get(999); got != "" {
		t.Errorf("Get(999) = %q, want empty", got)
	}
}

func TestHistoryManager_Clear(t *testing.T) {
	h := NewHistoryManager()
	h.Add("SELECT 1")
	h.Clear()

	if h.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", h.Count())
	}
}

func TestCommandParser_ParseCommand(t *testing.T) {
	r := &REPL{}
	p := NewCommandParser(r)

	tests := []struct {
		line    string
		isCmd   bool
		cmdName string
	}{
		{"exit", true, "exit"},
		{"quit", true, "exit"},
		{"\\q", true, "exit"},
		{"help", true, "help"},
		{"\\h", true, "help"},
		{"?", true, "help"},
		{"clear", true, "clear"},
		{"\\c", true, "clear"},
		{"status", true, "status"},
		{"\\s", true, "status"},
		{"print", true, "print"},
		{"\\p", true, "print"},
		{"edit", true, "edit"},
		{"\\e", true, "edit"},
		{"\\G", true, "ego"},
		{"\\g", true, "go"},
		{"use mysql", true, "use"},
		{"\\u mysql", true, "use"},
		{"prompt test>", true, "prompt"},
		{"\\R test>", true, "prompt"},
		{"source /tmp/test.sql", true, "source"},
		{"\\. /tmp/test.sql", true, "source"},
		{"\\! ls", true, "system"},
		{"\\o csv", true, "output"},
		{"SELECT 1", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			cmd, isCmd := p.ParseCommand(tt.line)
			if isCmd != tt.isCmd {
				t.Errorf("ParseCommand(%q) isCmd = %v, want %v", tt.line, isCmd, tt.isCmd)
			}
			if isCmd && cmd.Name != tt.cmdName {
				t.Errorf("ParseCommand(%q) cmd.Name = %q, want %q", tt.line, cmd.Name, tt.cmdName)
			}
		})
	}
}

func TestFormatter_FormatValue(t *testing.T) {
	f := NewFormatter()

	tests := []struct {
		val      interface{}
		expected string
	}{
		{nil, "NULL"},
		{"test", "test"},
		{123, "123"},
		{true, "true"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := f.formatValue(tt.val); got != tt.expected {
				t.Errorf("formatValue(%v) = %q, want %q", tt.val, got, tt.expected)
			}
		})
	}
}

func TestFormatter_PadRight(t *testing.T) {
	f := NewFormatter()

	if got := f.padRight("test", 10); utf8.RuneCountInString(got) != 10 {
		t.Errorf("padRight rune count = %d, want 10", utf8.RuneCountInString(got))
	}
	if got := f.padRight("testtesttest", 5); utf8.RuneCountInString(got) != 12 {
		t.Errorf("padRight rune count = %d, want 12 (no truncation)", utf8.RuneCountInString(got))
	}
}

func TestFormatter_PadLeft(t *testing.T) {
	f := NewFormatter()

	result := f.padLeft("123", 10)
	if !strings.HasPrefix(result, "       ") {
		t.Errorf("padLeft should add leading spaces, got %q", result)
	}
}
