package cli

import (
	"testing"
)

func TestTUILayout(t *testing.T) {
	// 测试界面布局是否正确
	app := NewTUIApp()
	if app == nil {
		t.Errorf("NewTUIApp() returned nil")
	}

	// 测试模型的 View 方法是否返回非空字符串
	model := initialModel()
	view := model.View()
	if view == "" {
		t.Errorf("View() returned empty string")
	}
}
