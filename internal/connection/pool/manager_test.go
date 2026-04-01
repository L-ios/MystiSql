package pool

import (
	"context"
	"errors"
	"testing"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/monitor"
	"MystiSql/pkg/types"
)

type errorFactory struct{}

func (f *errorFactory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	return nil, errors.New("factory error")
}

func TestNewConnectionPoolManagerNilConfig(t *testing.T) {
	factory := &mockFactory{}
	mgr := NewConnectionPoolManager(factory, nil)

	if mgr == nil {
		t.Fatal("expected manager to be created")
	}

	if mgr.config == nil {
		t.Error("expected config to be set")
	}

	if mgr.config.MaxConnections != 10 {
		t.Errorf("expected default MaxConnections=10, got %d", mgr.config.MaxConnections)
	}

	if mgr.config.MinConnections != 2 {
		t.Errorf("expected default MinConnections=2, got %d", mgr.config.MinConnections)
	}
}

func TestGetConnectionNoPool(t *testing.T) {
	factory := &mockFactory{}
	mgr := NewConnectionPoolManager(factory, nil)

	ctx := context.Background()
	_, err := mgr.GetConnection(ctx, "nonexistent")

	if err == nil {
		t.Fatal("expected error for nonexistent pool")
	}

	if !contains(err.Error(), "no connection pool found") {
		t.Errorf("expected 'no connection pool found' error, got: %v", err)
	}
}

func TestRemoveInstanceNoPool(t *testing.T) {
	factory := &mockFactory{}
	mgr := NewConnectionPoolManager(factory, nil)

	err := mgr.RemoveInstance("nonexistent")

	if err == nil {
		t.Fatal("expected error for nonexistent pool")
	}

	if !contains(err.Error(), "no connection pool found") {
		t.Errorf("expected 'no connection pool found' error, got: %v", err)
	}
}

func TestCloseNoPools(t *testing.T) {
	factory := &mockFactory{}
	mgr := NewConnectionPoolManager(factory, nil)

	err := mgr.Close()

	if err != nil {
		t.Errorf("expected nil error when closing empty manager, got: %v", err)
	}
}

func TestGetPoolStatsNoPool(t *testing.T) {
	factory := &mockFactory{}
	mgr := NewConnectionPoolManager(factory, nil)

	stats := mgr.GetPoolStats("nonexistent")

	if stats != nil {
		t.Error("expected nil stats for nonexistent pool")
	}
}

func TestGetAllPoolStatsNoPools(t *testing.T) {
	factory := &mockFactory{}
	mgr := NewConnectionPoolManager(factory, nil)

	stats := mgr.GetAllPoolStats()

	if stats == nil {
		t.Fatal("expected empty map, got nil")
	}

	if len(stats) != 0 {
		t.Errorf("expected empty map, got %d entries", len(stats))
	}
}

func TestGetMetricsCollectorFromManager(t *testing.T) {
	factory := &mockFactory{}
	collector := monitor.NewCollector()
	mgr := NewConnectionPoolManager(factory, nil, WithManagerMetrics(collector))

	result := mgr.GetMetricsCollector()

	if result == nil {
		t.Fatal("expected non-nil collector")
	}

	if result != collector {
		t.Error("expected same collector instance")
	}
}

func TestGetMetricsCollectorDefault(t *testing.T) {
	factory := &mockFactory{}
	mgr := NewConnectionPoolManager(factory, nil)

	result := mgr.GetMetricsCollector()

	if result == nil {
		t.Fatal("expected non-nil default collector")
	}
}

func TestGetPoolNoPool(t *testing.T) {
	factory := &mockFactory{}
	mgr := NewConnectionPoolManager(factory, nil)

	pool := mgr.GetPool("nonexistent")

	if pool != nil {
		t.Error("expected nil pool for nonexistent instance")
	}
}

func TestListInstancesNoPools(t *testing.T) {
	factory := &mockFactory{}
	mgr := NewConnectionPoolManager(factory, nil)

	instances := mgr.ListInstances()

	if instances == nil {
		t.Fatal("expected empty slice, got nil")
	}

	if len(instances) != 0 {
		t.Errorf("expected empty slice, got %d items", len(instances))
	}
}

func TestWithManagerMetrics(t *testing.T) {
	_ = &mockFactory{}
	collector := monitor.NewCollector()

	opt := WithManagerMetrics(collector)
	if opt == nil {
		t.Fatal("expected non-nil option")
	}

	mgr := &ConnectionPoolManager{}
	opt(mgr)

	if mgr.metricsCollector != collector {
		t.Error("expected collector to be set by option")
	}
}

func TestNewConnectionPoolManagerWithOptions(t *testing.T) {
	factory := &mockFactory{}
	collector := monitor.NewCollector()

	mgr := NewConnectionPoolManager(
		factory,
		nil,
		WithManagerMetrics(collector),
	)

	if mgr.metricsCollector != collector {
		t.Error("expected collector to be set via options")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
