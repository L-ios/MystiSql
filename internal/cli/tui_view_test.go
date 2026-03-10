package cli

import (
	"MystiSql/pkg/types"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIViewSimpleStyle(t *testing.T) {
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

	// Mark welcome as shown to test normal view
	m.welcomeShown = true
	view := m.View()

	if strings.Contains(view, "MystiSql TUI | 当前实例:") {
		t.Error("View should not contain top bar")
	}

	if !strings.Contains(view, "mystisql@test-db>") {
		t.Errorf("View should contain simple prompt 'mystisql@test-db>', got:\n%s", view)
	}

	if strings.Contains(view, "┌") || strings.Contains(view, "│") || strings.Contains(view, "└") {
		t.Error("View should not contain box-drawing characters")
	}
}

func TestTUIViewWelcomeMessage(t *testing.T) {
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

	// First view should show welcome message
	view := m.View()

	if !strings.Contains(view, "Welcome to the MystiSql monitor") {
		t.Error("First view should contain welcome message")
	}

	if !strings.Contains(view, "1 instance(s) configured") {
		t.Error("Welcome should show instance count")
	}

	if !strings.Contains(view, "Current instance: test-db") {
		t.Error("Welcome should show current instance")
	}

	// Second view should not show welcome again
	view2 := m.View()
	if strings.Contains(view2, "Welcome to the MystiSql monitor") {
		t.Error("Welcome message should not be shown again")
	}
}

func TestTUIViewWithResults(t *testing.T) {
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
	m.welcomeShown = true

	m.results = "id\tname\n1\tAlice\n\n1 row, 0.005s"

	view := m.View()

	if !strings.Contains(view, "id\tname") {
		t.Error("View should contain results")
	}

	if !strings.Contains(view, "mystisql@test-db>") {
		t.Error("View should contain prompt after results")
	}
}

func TestTUIViewWithErrorMessage(t *testing.T) {
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
	m.welcomeShown = true

	m.errorMsg = "connection failed"

	view := m.View()

	if !strings.Contains(view, "ERROR: connection failed") {
		t.Error("View should contain error message with ERROR prefix")
	}
}

func TestTUIViewHelp(t *testing.T) {
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
	m.welcomeShown = true

	m.showHelp = true
	view := m.View()

	if !strings.Contains(view, "快捷键:") {
		t.Error("Help view should contain shortcuts title")
	}

	if !strings.Contains(view, "Enter") {
		t.Error("Help view should contain Enter key")
	}

	helpMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, _ := m.Update(helpMsg)
	updated := updatedModel.(*model)

	if updated.showHelp {
		t.Error("Help should be closed after pressing any key")
	}
}

func TestTUIViewInstanceList(t *testing.T) {
	mockReg := NewMockRegistry()
	mockReg.AddInstance(&types.DatabaseInstance{
		Name:   "db1",
		Type:   "mysql",
		Host:   "localhost",
		Port:   3306,
		Status: "healthy",
	})
	mockReg.AddInstance(&types.DatabaseInstance{
		Name:   "db2",
		Type:   "postgresql",
		Host:   "localhost",
		Port:   5432,
		Status: "healthy",
	})

	cfg := &types.Config{}
	m := initialModel(cfg, mockReg).(*model)
	m.Init()
	m.welcomeShown = true

	m.showInstanceList = true
	view := m.View()

	if !strings.Contains(view, "可用实例:") {
		t.Error("Instance list view should contain title")
	}

	if !strings.Contains(view, "db1") || !strings.Contains(view, "db2") {
		t.Error("Instance list view should contain instance names")
	}

	if !strings.Contains(view, "→") {
		t.Error("Instance list should show selected indicator")
	}
}
