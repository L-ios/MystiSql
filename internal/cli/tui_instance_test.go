package cli

import (
	"MystiSql/pkg/types"
	"testing"
)

func TestModelInstanceToggle(t *testing.T) {
	reg := NewMockRegistry()
	reg.AddInstance(&types.DatabaseInstance{Name: "local-mysql", Type: "mysql"})
	reg.AddInstance(&types.DatabaseInstance{Name: "local-postgres", Type: "postgresql"})

	m := &model{
		config:   &types.Config{},
		registry: reg,
	}
	m.Init()

	if len(m.instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(m.instances))
	}

	if m.selectedInstance != 0 {
		t.Errorf("expected default selected instance index 0, got %d", m.selectedInstance)
	}

	m.selectedInstance = 1
	m.instance = m.instances[m.selectedInstance]
	if m.instance != "local-postgres" {
		t.Errorf("expected instance 'local-postgres', got '%s'", m.instance)
	}
}
