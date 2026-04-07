package health

import (
	"testing"
	"time"

	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	"go.uber.org/zap"
)

type mockRegistry struct {
	instances []*types.DatabaseInstance
	err       error
}

func (m *mockRegistry) Register(instance *types.DatabaseInstance) error { return nil }
func (m *mockRegistry) GetInstance(name string) (*types.DatabaseInstance, error) {
	return nil, errors.ErrInstanceNotFound
}
func (m *mockRegistry) ListInstances() ([]*types.DatabaseInstance, error) { return m.instances, m.err }
func (m *mockRegistry) Remove(name string) error                          { return nil }
func (m *mockRegistry) Clear()                                            {}

func TestNewMonitor(t *testing.T) {
	logger := zap.NewNop()
	m := NewMonitor(&mockRegistry{}, nil, logger, 30*time.Second)

	if m == nil {
		t.Fatal("NewMonitor returned nil")
	}
	if m.running {
		t.Error("monitor should not be running after creation")
	}
	if m.GetInstanceStatus("any") != types.InstanceStatusUnknown {
		t.Error("unknown instance should return Unknown status")
	}
}

func TestMonitor_StartStop(t *testing.T) {
	logger := zap.NewNop()
	m := NewMonitor(&mockRegistry{}, nil, logger, 1*time.Second)

	m.Start()
	if !m.running {
		t.Error("should be running after Start()")
	}

	m.Start()
	if !m.running {
		t.Error("should still be running after double Start()")
	}

	m.Stop()
	if m.running {
		t.Error("should not be running after Stop()")
	}

	m.Stop()
	if m.running {
		t.Error("should not be running after double Stop()")
	}
}

func TestMonitor_GetEventChannel(t *testing.T) {
	logger := zap.NewNop()
	m := NewMonitor(&mockRegistry{}, nil, logger, 1*time.Second)

	ch := m.GetEventChannel()
	if ch == nil {
		t.Error("event channel should not be nil")
	}
}

func TestMonitor_GetAllStatus(t *testing.T) {
	logger := zap.NewNop()
	m := NewMonitor(&mockRegistry{}, nil, logger, 1*time.Second)

	statuses := m.GetAllStatus()
	if len(statuses) != 0 {
		t.Errorf("empty monitor should have 0 statuses, got %d", len(statuses))
	}
}

func TestHealthEvent(t *testing.T) {
	event := HealthEvent{
		InstanceName: "test-db",
		OldStatus:    types.InstanceStatusUnknown,
		NewStatus:    types.InstanceStatusHealthy,
		Timestamp:    time.Now(),
	}

	if event.InstanceName != "test-db" {
		t.Errorf("InstanceName = %q, want %q", event.InstanceName, "test-db")
	}
	if event.OldStatus != types.InstanceStatusUnknown {
		t.Errorf("OldStatus = %q", event.OldStatus)
	}
	if event.NewStatus != types.InstanceStatusHealthy {
		t.Errorf("NewStatus = %q", event.NewStatus)
	}
}
