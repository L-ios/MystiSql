package cli

import (
	"MystiSql/pkg/types"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type MockRegistry struct {
	instances []*types.DatabaseInstance
}

func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		instances: make([]*types.DatabaseInstance, 0),
	}
}

func (m *MockRegistry) AddInstance(instance *types.DatabaseInstance) {
	m.instances = append(m.instances, instance)
}

func (m *MockRegistry) Register(instance *types.DatabaseInstance) error {
	m.instances = append(m.instances, instance)
	return nil
}

func (m *MockRegistry) GetInstance(name string) (*types.DatabaseInstance, error) {
	for _, inst := range m.instances {
		if inst.Name == name {
			return inst, nil
		}
	}
	return nil, nil
}

func (m *MockRegistry) ListInstances() ([]*types.DatabaseInstance, error) {
	return m.instances, nil
}

func (m *MockRegistry) Remove(name string) error {
	for i, inst := range m.instances {
		if inst.Name == name {
			m.instances = append(m.instances[:i], m.instances[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockRegistry) Clear() {
	m.instances = make([]*types.DatabaseInstance, 0)
}

func TestModelUpdateHistoryNavigation(t *testing.T) {
	m := &model{
		config:   &types.Config{},
		registry: NewMockRegistry(),
		history:  []string{"SELECT 1", "SELECT 2", "SELECT 3"},
		input:    "",
		cursor:   0,
	}
	m.Init()

	// Test up arrow - navigate backwards in history
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(upMsg)
	updated := updatedModel.(*model)
	if updated.input != "SELECT 3" {
		t.Errorf("Expected 'SELECT 3', got '%s'", updated.input)
	}

	// Continue navigating up
	updatedModel, _ = updated.Update(upMsg)
	updated = updatedModel.(*model)
	if updated.input != "SELECT 2" {
		t.Errorf("Expected 'SELECT 2', got '%s'", updated.input)
	}

	// Navigate down with down arrow
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = updated.Update(downMsg)
	updated = updatedModel.(*model)
	if updated.input != "SELECT 3" {
		t.Errorf("Expected 'SELECT 3' after down, got '%s'", updated.input)
	}

	// Continue down to wrap around
	updatedModel, _ = updated.Update(downMsg)
	updated = updatedModel.(*model)
	if updated.input != "" {
		t.Errorf("Expected empty input at end of history, got '%s'", updated.input)
	}
}
