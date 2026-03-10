package cli

import (
	"MystiSql/pkg/types"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIWithConfigIntegration(t *testing.T) {
	cfg := &types.Config{
		Instances: []types.InstanceConfig{
			{Name: "test-mysql", Type: "mysql", Host: "localhost", Port: 3306},
			{Name: "test-postgres", Type: "postgresql", Host: "localhost", Port: 5432},
		},
	}

	mockReg := NewMockRegistry()
	for _, instCfg := range cfg.Instances {
		mockReg.AddInstance(&types.DatabaseInstance{
			Name:   instCfg.Name,
			Type:   instCfg.Type,
			Host:   instCfg.Host,
			Port:   instCfg.Port,
			Status: "healthy",
		})
	}

	app := NewTUIApp(cfg, mockReg)
	if app == nil {
		t.Fatal("NewTUIApp returned nil")
	}

	m := initialModel(cfg, mockReg).(*model)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil command")
	}

	if len(m.instances) != 2 {
		t.Errorf("Expected 2 instances, got %d", len(m.instances))
	}

	if m.instance != "test-mysql" {
		t.Errorf("Expected default instance 'test-mysql', got '%s'", m.instance)
	}
}

func TestTUIWithRegistryIntegration(t *testing.T) {
	mockReg := NewMockRegistry()
	mockReg.AddInstance(&types.DatabaseInstance{
		Name:   "integration-test",
		Type:   "mysql",
		Host:   "localhost",
		Port:   3306,
		Status: "healthy",
	})

	cfg := &types.Config{}
	app := NewTUIApp(cfg, mockReg)
	if app == nil {
		t.Fatal("NewTUIApp returned nil")
	}

	m := initialModel(cfg, mockReg).(*model)
	m.Init()

	if len(m.instances) != 1 {
		t.Errorf("Expected 1 instance, got %d", len(m.instances))
	}

	if m.instances[0] != "integration-test" {
		t.Errorf("Expected instance 'integration-test', got '%s'", m.instances[0])
	}
}

func TestTUIInstanceSwitchIntegration(t *testing.T) {
	mockReg := NewMockRegistry()
	mockReg.AddInstance(&types.DatabaseInstance{Name: "db1", Type: "mysql", Host: "localhost", Port: 3306, Status: "healthy"})
	mockReg.AddInstance(&types.DatabaseInstance{Name: "db2", Type: "postgresql", Host: "localhost", Port: 5432, Status: "healthy"})
	mockReg.AddInstance(&types.DatabaseInstance{Name: "db3", Type: "redis", Host: "localhost", Port: 6379, Status: "healthy"})

	cfg := &types.Config{}
	m := initialModel(cfg, mockReg).(*model)
	m.Init()

	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ := m.Update(tabMsg)
	updated := updatedModel.(*model)

	if !updated.showInstanceList {
		t.Error("Tab should show instance list")
	}

	for i := 0; i < 2; i++ {
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ = updated.Update(downMsg)
		updated = updatedModel.(*model)
	}

	if updated.selectedInstance != 2 {
		t.Errorf("Expected selected instance 2, got %d", updated.selectedInstance)
	}

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = updated.Update(enterMsg)
	updated = updatedModel.(*model)

	if updated.instance != "db3" {
		t.Errorf("Expected instance 'db3', got '%s'", updated.instance)
	}

	if updated.showInstanceList {
		t.Error("Instance list should be hidden after selection")
	}
}

func TestTUISQLExecutionIntegration(t *testing.T) {
	mockReg := NewMockRegistry()
	mockReg.AddInstance(&types.DatabaseInstance{
		Name:   "test-db",
		Type:   "mysql",
		Host:   "localhost",
		Port:   3306,
		Status: "healthy",
	})

	cfg := &types.Config{}
	m := initialModel(cfg, mockReg).(*model)
	m.Init()

	m.input = "SELECT 1"

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := m.Update(enterMsg)
	updated := updatedModel.(*model)

	if updated.input != "" {
		t.Error("Input should be cleared after execution")
	}

	if len(updated.history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(updated.history))
	}

	if updated.history[0] != "SELECT 1" {
		t.Errorf("Expected history 'SELECT 1', got '%s'", updated.history[0])
	}
}

func TestTUIErrorHandlingIntegration(t *testing.T) {
	mockReg := NewMockRegistry()

	cfg := &types.Config{}
	m := initialModel(cfg, mockReg).(*model)
	m.Init()

	if m.errorMsg == "" {
		t.Error("Expected error message for empty registry")
	}

	if len(m.instances) != 0 {
		t.Error("Expected empty instances list for empty registry")
	}
}

func TestTUIHistoryNavigationIntegration(t *testing.T) {
	mockReg := NewMockRegistry()
	mockReg.AddInstance(&types.DatabaseInstance{Name: "test", Type: "mysql", Host: "localhost", Port: 3306, Status: "healthy"})

	cfg := &types.Config{}
	m := initialModel(cfg, mockReg).(*model)
	m.Init()

	queries := []string{"SELECT 1", "SELECT 2", "INSERT INTO test VALUES (1)"}
	for _, q := range queries {
		m.input = q
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := m.Update(enterMsg)
		m = updatedModel.(*model)
	}

	if len(m.history) != 3 {
		t.Errorf("Expected 3 history entries, got %d", len(m.history))
	}

	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(upMsg)
	updated := updatedModel.(*model)

	if updated.input != "INSERT INTO test VALUES (1)" {
		t.Errorf("Expected 'INSERT INTO test VALUES (1)', got '%s'", updated.input)
	}

	upMsg = tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = updated.Update(upMsg)
	updated = updatedModel.(*model)

	if updated.input != "SELECT 2" {
		t.Errorf("Expected 'SELECT 2', got '%s'", updated.input)
	}

	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = updated.Update(downMsg)
	updated = updatedModel.(*model)

	if updated.input != "INSERT INTO test VALUES (1)" {
		t.Errorf("Expected 'INSERT INTO test VALUES (1)', got '%s'", updated.input)
	}
}
