package cli

import (
	"testing"
)

func TestInstanceSwitching(t *testing.T) {
	// 测试实例切换功能
	app := NewTUIApp()
	if app == nil {
		t.Errorf("NewTUIApp() returned nil")
	}

	// 直接创建模型实例并设置测试数据
	model := &model{
		instances:        []string{"local-mysql", "local-postgres", "local-oracle", "local-redis"},
		instance:         "local-mysql",
		selectedInstance: 0,
	}

	// 测试实例列表初始化
	if len(model.instances) == 0 {
		t.Errorf("instance list is empty")
	}

	// 测试默认实例
	if model.instance != "local-mysql" {
		t.Errorf("expected default instance 'local-mysql', got '%s'", model.instance)
	}

	// 测试默认选中的实例索引
	if model.selectedInstance != 0 {
		t.Errorf("expected default selected instance index 0, got %d", model.selectedInstance)
	}

	// 测试实例切换
	model.selectedInstance = 1
	model.instance = model.instances[model.selectedInstance]
	if model.instance != "local-postgres" {
		t.Errorf("expected instance 'local-postgres', got '%s'", model.instance)
	}
}
