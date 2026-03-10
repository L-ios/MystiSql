package cli

import (
	"MystiSql/pkg/types"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUITerminalCompatibility(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"标准终端 (80x24)", 80, 24},
		{"宽屏终端 (120x30)", 120, 30},
		{"小终端 (40x12)", 40, 12},
		{"大终端 (200x50)", 200, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &types.Config{
				Instances: []types.InstanceConfig{
					{Name: "test-mysql", Type: "mysql", Host: "localhost", Port: 3306},
				},
			}

			mockReg := NewMockRegistry()
			mockReg.AddInstance(&types.DatabaseInstance{
				Name:   "test-mysql",
				Type:   "mysql",
				Host:   "localhost",
				Port:   3306,
				Status: "healthy",
			})

			app := NewTUIApp(cfg, mockReg)
			if app == nil {
				t.Fatal("NewTUIApp returned nil")
			}

			m := initialModel(cfg, mockReg).(*model)

			windowMsg := tea.WindowSizeMsg{Width: tt.width, Height: tt.height}
			updatedModel, _ := m.Update(windowMsg)
			m = updatedModel.(*model)

			if m.width != tt.width {
				t.Errorf("Expected width %d, got %d", tt.width, m.width)
			}
			if m.height != tt.height {
				t.Errorf("Expected height %d, got %d", tt.height, m.height)
			}

			view := m.View()
			if view == "" {
				t.Error("View should not be empty")
			}
		})
	}
}

func TestTUIResizeHandling(t *testing.T) {
	cfg := &types.Config{
		Instances: []types.InstanceConfig{
			{Name: "test-mysql", Type: "mysql", Host: "localhost", Port: 3306},
		},
	}

	mockReg := NewMockRegistry()
	mockReg.AddInstance(&types.DatabaseInstance{
		Name:   "test-mysql",
		Type:   "mysql",
		Host:   "localhost",
		Port:   3306,
		Status: "healthy",
	})

	m := initialModel(cfg, mockReg).(*model)

	resizeEvents := []tea.WindowSizeMsg{
		{Width: 80, Height: 24},
		{Width: 120, Height: 30},
		{Width: 60, Height: 20},
		{Width: 100, Height: 40},
	}

	for i, resize := range resizeEvents {
		updatedModel, _ := m.Update(resize)
		m = updatedModel.(*model)

		if m.width != resize.Width {
			t.Errorf("Resize %d: Expected width %d, got %d", i, resize.Width, m.width)
		}
		if m.height != resize.Height {
			t.Errorf("Resize %d: Expected height %d, got %d", i, resize.Height, m.height)
		}

		view := m.View()
		if view == "" {
			t.Errorf("Resize %d: View should not be empty", i)
		}
	}
}

func TestTUISmallTerminal(t *testing.T) {
	cfg := &types.Config{
		Instances: []types.InstanceConfig{
			{Name: "test-mysql", Type: "mysql", Host: "localhost", Port: 3306},
		},
	}

	mockReg := NewMockRegistry()
	mockReg.AddInstance(&types.DatabaseInstance{
		Name:   "test-mysql",
		Type:   "mysql",
		Host:   "localhost",
		Port:   3306,
		Status: "healthy",
	})

	m := initialModel(cfg, mockReg).(*model)

	smallWindow := tea.WindowSizeMsg{Width: 20, Height: 10}
	updatedModel, _ := m.Update(smallWindow)
	m = updatedModel.(*model)

	view := m.View()
	if view == "" {
		t.Error("View should not be empty even with small terminal")
	}
}
