package cli

import (
	"MystiSql/pkg/types"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEKeyDoesNotTriggerExport(t *testing.T) {
	mockReg := NewMockRegistry()
	mockReg.AddInstance(&types.DatabaseInstance{
		Name:   "test",
		Type:   "mysql",
		Host:   "localhost",
		Port:   3306,
		Status: "healthy",
	})

	cfg := &types.Config{}
	m := initialModel(cfg, mockReg).(*model)
	m.Init()

	m.input = "s"

	eMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	updatedModel, _ := m.Update(eMsg)
	updated := updatedModel.(*model)

	if updated.showExportOptions {
		t.Error("'e' key should not trigger export options")
	}

	if updated.input != "se" {
		t.Errorf("Expected input 'se', got '%s'", updated.input)
	}
}

func TestCtrlEKeyTriggersExport(t *testing.T) {
	mockReg := NewMockRegistry()
	mockReg.AddInstance(&types.DatabaseInstance{
		Name:   "test",
		Type:   "mysql",
		Host:   "localhost",
		Port:   3306,
		Status: "healthy",
	})

	cfg := &types.Config{}
	m := initialModel(cfg, mockReg).(*model)
	m.Init()

	ctrlE := tea.KeyMsg{Type: tea.KeyCtrlE}
	updatedModel, _ := m.Update(ctrlE)
	updated := updatedModel.(*model)

	if !updated.showExportOptions {
		t.Error("Ctrl+E should trigger export options")
	}
}
