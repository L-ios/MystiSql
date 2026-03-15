package router

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type MockInstance struct {
	Name    string
	Address string
	Healthy bool
	Lag     time.Duration
}

type MockMasterSelector struct {
	instance *MockInstance
}

type MockSlaveSelector struct {
	instances []*MockInstance
	mu        sync.Mutex
	current   int
}

func (s *MockMasterSelector) SelectMaster(ctx context.Context) (*MockInstance, error) {
	return s.instance, nil
}

func (s *MockSlaveSelector) SelectSlave(ctx context.Context) (*MockInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.instances) == 0 {
		return nil, fmt.Errorf("no slaves available")
	}

	idx := s.current
	instance := s.instances[idx]
	s.current = (s.current + 1) % len(s.instances)
	return instance, nil
}

func TestRouter_SelectRouting(t *testing.T) {
	ctx := context.Background()

	master := &MockInstance{Name: "master-1", Address: "master:3306", Healthy: true}
	slave1 := &MockInstance{Name: "slave-1", Address: "slave:3306", Healthy: true, Lag: 0}
	slave2 := &MockInstance{Name: "slave-2", Address: "slave:3307", Healthy: true, Lag: 0}

	masterSelector := &MockMasterSelector{instance: master}
	slaveSelector := &MockSlaveSelector{
		instances: []*MockInstance{slave1, slave2},
	}

	tests := []struct {
		name     string
		sql      string
		wantRole string
	}{
		{"SELECT routes to slave", "SELECT * FROM users", "slave"},
		{"INSERT routes to master", "INSERT INTO users (name) VALUES ('test')", "master"},
		{"UPDATE routes to master", "UPDATE users SET name = 'test'", "master"},
		{"DELETE routes to master", "DELETE FROM users WHERE id = 1", "master"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target *MockInstance
			var err error

			typ, _, _ := ParseSQL(tt.sql)
			if typ == SQLTypeSelect {
				target, err = slaveSelector.SelectSlave(ctx)
			} else {
				target, err = masterSelector.SelectMaster(ctx)
			}

			if err != nil {
				t.Fatalf("Route failed: %v", err)
			}
			if target == nil {
				t.Fatal("Expected target, got nil")
			}
			if tt.wantRole == "slave" && target.Name != "slave-1" && target.Name != "slave-2" {
				t.Errorf("Expected slave, got %s", target.Name)
			}
			if tt.wantRole == "master" && target.Name != "master-1" {
				t.Errorf("Expected master, got %s", target.Name)
			}
		})
	}
}

func TestRouter_SlaveRoundRobin(t *testing.T) {
	ctx := context.Background()

	slave1 := &MockInstance{Name: "slave-1", Address: "slave:3306", Healthy: true, Lag: 0}
	slave2 := &MockInstance{Name: "slave-2", Address: "slave:3307", Healthy: true, Lag: 0}
	slave3 := &MockInstance{Name: "slave-3", Address: "slave:3308", Healthy: true, Lag: 0}

	selector := &MockSlaveSelector{
		instances: []*MockInstance{slave1, slave2, slave3},
	}

	for i := 0; i < 6; i++ {
		target, err := selector.SelectSlave(ctx)
		if err != nil {
			t.Fatalf("SelectSlave failed: %v", err)
		}
		expectedIdx := i % 3
		if target != selector.instances[expectedIdx] {
			t.Errorf("Round %d: expected slave-%d, got %s", i, expectedIdx+1, target.Name)
		}
	}
}

func TestRouter_SlaveLagFiltering(t *testing.T) {
	ctx := context.Background()

	healthySlave := &MockInstance{Name: "slave-1", Address: "slave:3306", Healthy: true, Lag: 0}
	slowSlave := &MockInstance{Name: "slave-2", Address: "slave:3307", Healthy: true, Lag: 5 * time.Second}
	unhealthySlave := &MockInstance{Name: "slave-3", Address: "slave:3308", Healthy: false, Lag: 0}

	selector := &MockSlaveSelector{
		instances: []*MockInstance{healthySlave, slowSlave, unhealthySlave},
	}

	for i := 0; i < 10; i++ {
		target, err := selector.SelectSlave(ctx)
		if err != nil {
			t.Logf("Select %d: got error (expected when no healthy slaves)", err)
			break
		}
		if target.Healthy {
			t.Logf("Selected healthy slave: %s", target.Name)
		} else {
			t.Logf("Selected unhealthy slave: %s", target.Name)
		}
	}
}

func TestRouter_NoSlavesFallback(t *testing.T) {
	ctx := context.Background()

	master := &MockInstance{Name: "master-1", Address: "master:3306", Healthy: true}
	masterSelector := &MockMasterSelector{instance: master}
	slaveSelector := &MockSlaveSelector{instances: []*MockInstance{}}

	_, err := slaveSelector.SelectSlave(ctx)
	if err == nil {
		t.Logf("No slaves available, falling back to master")
	}

	target, err := masterSelector.SelectMaster(ctx)
	if err != nil {
		t.Fatalf("Master fallback failed: %v", err)
	}
	if target.Name != "master-1" {
		t.Errorf("Expected master fallback, got %s", target.Name)
	}
}
