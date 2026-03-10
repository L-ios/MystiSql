package cli

import (
	"MystiSql/pkg/types"
	"testing"
)

func TestCalculateLayout(t *testing.T) {
	tests := []struct {
		name         string
		width        int
		height       int
		minTopBar    int
		minBottomBar int
		minInput     int
	}{
		{"standard terminal", 100, 30, 1, 1, 3},
		{"small terminal", 80, 24, 1, 1, 3},
		{"large terminal", 120, 40, 1, 1, 3},
		{"minimum size", 80, 20, 1, 1, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &model{
				config:   &types.Config{},
				registry: NewMockRegistry(),
				width:    tt.width,
				height:   tt.height,
			}
			m.Init()

			layout := m.calculateLayout()

			if layout.topBar < tt.minTopBar {
				t.Errorf("topBar %d is less than minimum %d", layout.topBar, tt.minTopBar)
			}
			if layout.bottomBar < tt.minBottomBar {
				t.Errorf("bottomBar %d is less than minimum %d", layout.bottomBar, tt.minBottomBar)
			}
			if layout.input < tt.minInput {
				t.Errorf("input %d is less than minimum %d", layout.input, tt.minInput)
			}

			total := layout.topBar + layout.results + layout.input + layout.bottomBar
			if total > tt.height {
				t.Errorf("total layout height %d exceeds window height %d", total, tt.height)
			}
		})
	}
}

func TestCalculateLayoutZeroSize(t *testing.T) {
	m := &model{
		config:   &types.Config{},
		registry: NewMockRegistry(),
		width:    0,
		height:   0,
	}
	m.Init()

	layout := m.calculateLayout()
	if layout.topBar != 1 {
		t.Errorf("Expected topBar 1 for zero size, got %d", layout.topBar)
	}
	if layout.bottomBar != 1 {
		t.Errorf("Expected bottomBar 1 for zero size, got %d", layout.bottomBar)
	}
}
