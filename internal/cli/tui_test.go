package cli

import (
	"testing"
)

func TestTUICmd(t *testing.T) {
	// 测试 TUI 命令是否能正确创建
	tuiCmd := NewTUICmd()
	if tuiCmd == nil {
		t.Errorf("NewTUICmd() returned nil")
	}

	if tuiCmd.Use != "tui" {
		t.Errorf("Expected command name 'tui', got '%s'", tuiCmd.Use)
	}

	if tuiCmd.Short != "启动交互式 TUI 界面" {
		t.Errorf("Expected short description '启动交互式 TUI 界面', got '%s'", tuiCmd.Short)
	}
}
