package health

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"MystiSql/pkg/types"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Mock registry — implements the LOCAL InstanceRegistry interface
// (GetInstance + ListInstances only, NOT discovery.InstanceRegistry).
// ---------------------------------------------------------------------------

type mockEnhancedRegistry struct {
	instances []*types.DatabaseInstance
	listErr   error
	getErr    error
}

func (m *mockEnhancedRegistry) GetInstance(name string) (*types.DatabaseInstance, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, inst := range m.instances {
		if inst.Name == name {
			return inst, nil
		}
	}
	return nil, errors.New("instance not found")
}

func (m *mockEnhancedRegistry) ListInstances() ([]*types.DatabaseInstance, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.instances, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestInstance(name string) *types.DatabaseInstance {
	return types.NewDatabaseInstance(name, types.DatabaseTypeMySQL, "127.0.0.1", 3306)
}

func defaultTestConfig() types.HealthCheckConfig {
	return types.HealthCheckConfig{
		Enabled:           true,
		Interval:          1 * time.Second,
		Timeout:           100 * time.Millisecond,
		FailureThreshold:  3,
		RecoveryThreshold: 2,
	}
}

func newTestChecker(checkFunc func(ctx context.Context, instance *types.DatabaseInstance) error) *EnhancedHealthChecker {
	return newTestCheckerWithRegistry(checkFunc, &mockEnhancedRegistry{})
}

func newTestCheckerWithRegistry(
	checkFunc func(ctx context.Context, instance *types.DatabaseInstance) error,
	registry InstanceRegistry,
) *EnhancedHealthChecker {
	return NewEnhancedHealthChecker(defaultTestConfig(), registry, checkFunc, zap.NewNop())
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestEnhancedHealthChecker_StartStop(t *testing.T) {
	checker := newTestChecker(func(_ context.Context, _ *types.DatabaseInstance) error {
		return nil
	})

	// Start sets running to 1.
	checker.Start()
	if checker.running != 1 {
		t.Fatalf("expected running=1 after Start, got %d", checker.running)
	}

	// Double Start is idempotent.
	checker.Start()
	if checker.running != 1 {
		t.Fatalf("expected running=1 after double Start, got %d", checker.running)
	}

	// Stop sets running to 0.
	checker.Stop()
	if checker.running != 0 {
		t.Fatalf("expected running=0 after Stop, got %d", checker.running)
	}

	// Double Stop is idempotent — must not panic (double close of eventCh).
	// We need a fresh checker because Stop closed eventCh.
	checker2 := newTestChecker(func(_ context.Context, _ *types.DatabaseInstance) error {
		return nil
	})
	checker2.Start()
	checker2.Stop()
	checker2.Stop() // second Stop should be no-op
	if checker2.running != 0 {
		t.Fatalf("expected running=0 after double Stop, got %d", checker2.running)
	}
}

func TestEnhancedHealthChecker_HealthyCheck(t *testing.T) {
	inst := newTestInstance("mysql-1")
	registry := &mockEnhancedRegistry{instances: []*types.DatabaseInstance{inst}}

	checker := newTestCheckerWithRegistry(
		func(_ context.Context, _ *types.DatabaseInstance) error { return nil },
		registry,
	)

	health, err := checker.ForceCheck("mysql-1")
	if err != nil {
		t.Fatalf("ForceCheck returned error: %v", err)
	}

	if health.Status != types.HealthStatusHealthy {
		t.Errorf("expected status %q, got %q", types.HealthStatusHealthy, health.Status)
	}
	if health.ConsecutiveOKs != 1 {
		t.Errorf("expected ConsecutiveOKs=1, got %d", health.ConsecutiveOKs)
	}
	if health.ResponseTime <= 0 {
		t.Error("expected ResponseTime > 0")
	}
	if health.LastCheck.IsZero() {
		t.Error("expected LastCheck to be set")
	}
}

func TestEnhancedHealthChecker_UnhealthyCheck(t *testing.T) {
	inst := newTestInstance("mysql-1")
	registry := &mockEnhancedRegistry{instances: []*types.DatabaseInstance{inst}}
	checkErr := errors.New("connection refused")

	checker := newTestCheckerWithRegistry(
		func(_ context.Context, _ *types.DatabaseInstance) error { return checkErr },
		registry,
	)

	// Simulate FailureThreshold (3) consecutive failures.
	for i := 0; i < 3; i++ {
		health, err := checker.ForceCheck("mysql-1")
		if err != nil {
			t.Fatalf("ForceCheck[%d] returned error: %v", i, err)
		}
		wantStatus := types.HealthStatusChecking
		if i == 2 {
			wantStatus = types.HealthStatusUnhealthy
		}
		if health.Status != wantStatus {
			t.Errorf("check %d: expected status %q, got %q", i, wantStatus, health.Status)
		}
	}

	health, _ := checker.GetHealth("mysql-1")
	if health.ConsecutiveFails != 3 {
		t.Errorf("expected ConsecutiveFails=3, got %d", health.ConsecutiveFails)
	}
	if health.LastError != checkErr.Error() {
		t.Errorf("expected LastError=%q, got %q", checkErr.Error(), health.LastError)
	}
}

func TestEnhancedHealthChecker_RecoveryCheck(t *testing.T) {
	inst := newTestInstance("pg-1")
	registry := &mockEnhancedRegistry{instances: []*types.DatabaseInstance{inst}}

	var checkResult error
	checker := newTestCheckerWithRegistry(
		func(_ context.Context, _ *types.DatabaseInstance) error { return checkResult },
		registry,
	)

	// Phase 1: drive to unhealthy (3 failures).
	checkResult = errors.New("fail")
	for i := 0; i < 3; i++ {
		checker.ForceCheck("pg-1")
	}
	health, _ := checker.GetHealth("pg-1")
	if health.Status != types.HealthStatusUnhealthy {
		t.Fatalf("expected unhealthy after 3 fails, got %q", health.Status)
	}

	// Phase 2: recover (2 successes = RecoveryThreshold).
	checkResult = nil
	for i := 0; i < 2; i++ {
		checker.ForceCheck("pg-1")
	}
	health, _ = checker.GetHealth("pg-1")
	if health.Status != types.HealthStatusHealthy {
		t.Errorf("expected healthy after recovery, got %q", health.Status)
	}
	if health.ConsecutiveOKs != 2 {
		t.Errorf("expected ConsecutiveOKs=2, got %d", health.ConsecutiveOKs)
	}
	if health.ConsecutiveFails != 0 {
		t.Errorf("expected ConsecutiveFails=0, got %d", health.ConsecutiveFails)
	}
}

func TestEnhancedHealthChecker_StatusCaching(t *testing.T) {
	inst := newTestInstance("cache-1")
	registry := &mockEnhancedRegistry{instances: []*types.DatabaseInstance{inst}}

	checker := newTestCheckerWithRegistry(
		func(_ context.Context, _ *types.DatabaseInstance) error { return nil },
		registry,
	)

	// Before any check, GetHealth returns error.
	_, err := checker.GetHealth("cache-1")
	if err == nil {
		t.Error("expected error for unchecked instance")
	}

	// First check stores status.
	_, err = checker.ForceCheck("cache-1")
	if err != nil {
		t.Fatalf("ForceCheck error: %v", err)
	}

	// GetHealth returns cached status.
	health, err := checker.GetHealth("cache-1")
	if err != nil {
		t.Fatalf("GetHealth error: %v", err)
	}
	if health.Status != types.HealthStatusHealthy {
		t.Errorf("expected healthy, got %q", health.Status)
	}

	// GetAllHealth returns all instances.
	all := checker.GetAllHealth()
	if len(all) != 1 {
		t.Fatalf("expected 1 instance in GetAllHealth, got %d", len(all))
	}
	if _, ok := all["cache-1"]; !ok {
		t.Error("expected cache-1 in GetAllHealth")
	}
}

func TestEnhancedHealthChecker_EventNotification(t *testing.T) {
	inst := newTestInstance("event-1")
	registry := &mockEnhancedRegistry{instances: []*types.DatabaseInstance{inst}}

	var checkResult error
	checker := newTestCheckerWithRegistry(
		func(_ context.Context, _ *types.DatabaseInstance) error { return checkResult },
		registry,
	)

	// First check: checking → healthy triggers event.
	checkResult = nil
	checker.ForceCheck("event-1")

	select {
	case evt := <-checker.GetEventChannel():
		if evt.InstanceName != "event-1" {
			t.Errorf("expected InstanceName=event-1, got %q", evt.InstanceName)
		}
		if evt.OldStatus != types.HealthStatusChecking {
			t.Errorf("expected OldStatus=checking, got %q", evt.OldStatus)
		}
		if evt.NewStatus != types.HealthStatusHealthy {
			t.Errorf("expected NewStatus=healthy, got %q", evt.NewStatus)
		}
		if evt.Timestamp.IsZero() {
			t.Error("expected Timestamp to be set")
		}
	case <-time.After(time.Second):
		t.Fatal("expected status-change event, got none")
	}

	// Second healthy check: no status change → no event.
	checker.ForceCheck("event-1")
	select {
	case <-checker.GetEventChannel():
		t.Fatal("unexpected event when status did not change")
	case <-time.After(50 * time.Millisecond):
		// expected
	}

	// Drive to unhealthy: healthy → unhealthy triggers event.
	checkResult = errors.New("fail")
	for i := 0; i < 3; i++ {
		checker.ForceCheck("event-1")
	}

	select {
	case evt := <-checker.GetEventChannel():
		if evt.NewStatus != types.HealthStatusUnhealthy {
			t.Errorf("expected NewStatus=unhealthy, got %q", evt.NewStatus)
		}
		if evt.Error == "" {
			t.Error("expected Error field to be set for unhealthy transition")
		}
	case <-time.After(time.Second):
		t.Fatal("expected unhealthy transition event, got none")
	}
}

func TestEnhancedHealthChecker_ForceCheck(t *testing.T) {
	inst := newTestInstance("force-1")
	registry := &mockEnhancedRegistry{instances: []*types.DatabaseInstance{inst}}

	checker := newTestCheckerWithRegistry(
		func(_ context.Context, _ *types.DatabaseInstance) error { return nil },
		registry,
	)

	health, err := checker.ForceCheck("force-1")
	if err != nil {
		t.Fatalf("ForceCheck error: %v", err)
	}
	if health == nil {
		t.Fatal("expected non-nil health")
	}
	if health.Status != types.HealthStatusHealthy {
		t.Errorf("expected healthy, got %q", health.Status)
	}

	// ForceCheck on non-existent instance returns error.
	_, err = checker.ForceCheck("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent instance")
	}
}

func TestEnhancedHealthChecker_RemoveInstance(t *testing.T) {
	inst := newTestInstance("rm-1")
	registry := &mockEnhancedRegistry{instances: []*types.DatabaseInstance{inst}}

	checker := newTestCheckerWithRegistry(
		func(_ context.Context, _ *types.DatabaseInstance) error { return nil },
		registry,
	)

	checker.ForceCheck("rm-1")
	if _, err := checker.GetHealth("rm-1"); err != nil {
		t.Fatalf("expected instance to exist, got error: %v", err)
	}

	checker.RemoveInstance("rm-1")

	_, err := checker.GetHealth("rm-1")
	if err == nil {
		t.Error("expected error after RemoveInstance")
	}
}

func TestEnhancedHealthChecker_GetStats(t *testing.T) {
	healthyInst := newTestInstance("s-healthy")
	unhealthyInst := newTestInstance("s-unhealthy")
	bothReg := &mockEnhancedRegistry{instances: []*types.DatabaseInstance{healthyInst, unhealthyInst}}

	checker := newTestCheckerWithRegistry(
		func(_ context.Context, inst *types.DatabaseInstance) error {
			if inst.Name == "s-unhealthy" {
				return errors.New("fail")
			}
			return nil
		},
		bothReg,
	)

	checker.ForceCheck("s-healthy")
	for i := 0; i < 3; i++ {
		checker.ForceCheck("s-unhealthy")
	}

	stats := checker.GetStats()
	if stats.TotalInstances != 2 {
		t.Errorf("expected TotalInstances=2, got %d", stats.TotalInstances)
	}
	if stats.HealthyInstances != 1 {
		t.Errorf("expected HealthyInstances=1, got %d", stats.HealthyInstances)
	}
	if stats.UnhealthyInstances != 1 {
		t.Errorf("expected UnhealthyInstances=1, got %d", stats.UnhealthyInstances)
	}
}

func TestEnhancedHealthChecker_IsHealthy(t *testing.T) {
	inst := newTestInstance("ish-1")
	registry := &mockEnhancedRegistry{instances: []*types.DatabaseInstance{inst}}

	var checkResult error
	checker := newTestCheckerWithRegistry(
		func(_ context.Context, _ *types.DatabaseInstance) error { return checkResult },
		registry,
	)

	// Unknown instance → false.
	if checker.IsHealthy("unknown") {
		t.Error("expected IsHealthy=false for unknown instance")
	}

	// Healthy check → true.
	checkResult = nil
	checker.ForceCheck("ish-1")
	if !checker.IsHealthy("ish-1") {
		t.Error("expected IsHealthy=true for healthy instance")
	}

	// Drive to unhealthy → false.
	checkResult = errors.New("fail")
	for i := 0; i < 3; i++ {
		checker.ForceCheck("ish-1")
	}
	if checker.IsHealthy("ish-1") {
		t.Error("expected IsHealthy=false for unhealthy instance")
	}
}

func TestEnhancedHealthChecker_ConcurrentSafety(t *testing.T) {
	instances := []*types.DatabaseInstance{
		newTestInstance("conc-1"),
		newTestInstance("conc-2"),
		newTestInstance("conc-3"),
	}
	registry := &mockEnhancedRegistry{instances: instances}

	checker := newTestCheckerWithRegistry(
		func(_ context.Context, inst *types.DatabaseInstance) error {
			if inst.Name == "conc-2" {
				return errors.New("fail")
			}
			return nil
		},
		registry,
	)

	const goroutines = 20
	names := []string{"conc-1", "conc-2", "conc-3"}
	var wg sync.WaitGroup
	var errCount int64

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				name := names[j%len(names)]
				_, err := checker.ForceCheck(name)
				if err != nil {
					atomic.AddInt64(&errCount, 1)
				}
				_ = checker.IsHealthy(name)
				_ = checker.GetStats()
				_ = checker.GetAllHealth()
			}
		}()
	}

	wg.Wait()

	if atomic.LoadInt64(&errCount) > 0 {
		t.Errorf("unexpected errors during concurrent access: %d", errCount)
	}

	stats := checker.GetStats()
	if stats.TotalInstances != 3 {
		t.Errorf("expected TotalInstances=3, got %d", stats.TotalInstances)
	}
}

func TestEnhancedHealthChecker_ListInstancesError(t *testing.T) {
	registry := &mockEnhancedRegistry{
		listErr: errors.New("registry unavailable"),
	}

	checker := newTestCheckerWithRegistry(
		func(_ context.Context, _ *types.DatabaseInstance) error { return nil },
		registry,
	)

	// ForceCheck uses GetInstance, not ListInstances, so it should still work
	// if we add the instance manually.
	registry.listErr = nil
	registry.instances = []*types.DatabaseInstance{newTestInstance("err-1")}
	_, err := checker.ForceCheck("err-1")
	if err != nil {
		t.Fatalf("ForceCheck should succeed when GetInstance works: %v", err)
	}

	// Now test checkAllInstances with ListInstances error.
	registry.listErr = errors.New("registry unavailable")
	registry.instances = nil

	// checkAllInstances should not panic and should not modify existing statuses.
	// We call it directly to test the error path.
	checker.checkAllInstances()

	// The previously checked instance should still be in the map.
	health, err := checker.GetHealth("err-1")
	if err != nil {
		t.Fatalf("expected existing health status to remain, got error: %v", err)
	}
	if health.Status != types.HealthStatusHealthy {
		t.Errorf("expected status to remain healthy, got %q", health.Status)
	}
}
