package cli

import (
	"testing"
)

func TestTUIApp(t *testing.T) {
	// 测试 TUI 应用是否能正确创建
	app := NewTUIApp()
	if app == nil {
		t.Errorf("NewTUIApp() returned nil")
	}

	// 测试 TUI 应用是否能正确启动
	// 这里只测试创建，不实际运行，因为会阻塞测试
}
