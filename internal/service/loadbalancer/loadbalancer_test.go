package loadbalancer

import (
	"testing"

	"MystiSql/pkg/types"
)

func TestRoundRobinLB_Select(t *testing.T) {
	lb := NewRoundRobinLB()

	instances := []*types.DatabaseInstance{
		{Name: "instance-1"},
		{Name: "instance-2"},
		{Name: "instance-3"},
	}

	// Test round robin selection
	instance1, err := lb.Select(instances)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if instance1.Name != "instance-1" {
		t.Errorf("Expected instance-1, got %s", instance1.Name)
	}

	instance2, err := lb.Select(instances)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if instance2.Name != "instance-2" {
		t.Errorf("Expected instance-2, got %s", instance2.Name)
	}

	instance3, err := lb.Select(instances)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if instance3.Name != "instance-3" {
		t.Errorf("Expected instance-3, got %s", instance3.Name)
	}

	// Should wrap around
	instance1Again, err := lb.Select(instances)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if instance1Again.Name != "instance-1" {
		t.Errorf("Expected instance-1 (wrap around), got %s", instance1Again.Name)
	}
}

func TestRoundRobinLB_Select_Empty(t *testing.T) {
	lb := NewRoundRobinLB()

	instance, err := lb.Select([]*types.DatabaseInstance{})
	if err == nil {
		t.Error("Expected error for empty instances, got nil")
	}
	if instance != nil {
		t.Error("Expected nil instance for empty instances, got non-nil")
	}
}

func TestRoundRobinLB_Reset(t *testing.T) {
	lb := NewRoundRobinLB()

	instances := []*types.DatabaseInstance{
		{Name: "instance-1"},
		{Name: "instance-2"},
	}

	// Select once
	_, err := lb.Select(instances)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	// Reset
	lb.Reset()

	// Should start from the beginning again
	instance, err := lb.Select(instances)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if instance.Name != "instance-1" {
		t.Errorf("Expected instance-1 after reset, got %s", instance.Name)
	}
}

func TestWeightedLB_Select(t *testing.T) {
	lb := NewWeightedLB()

	instances := []*types.DatabaseInstance{
		{Name: "instance-1", Weight: 5},
		{Name: "instance-2", Weight: 1},
	}

	// Set weights
	lb.SetWeight("instance-1", 5)
	lb.SetWeight("instance-2", 1)

	// Test selection
	instance, err := lb.Select(instances)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	// Should return one of the instances
	if instance.Name != "instance-1" && instance.Name != "instance-2" {
		t.Errorf("Unexpected instance: %s", instance.Name)
	}
}

func TestWeightedLB_Select_ZeroWeight(t *testing.T) {
	lb := NewWeightedLB()

	instances := []*types.DatabaseInstance{
		{Name: "instance-1", Weight: 0},
		{Name: "instance-2", Weight: 0},
	}

	instance, err := lb.Select(instances)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if instance.Name != "instance-1" && instance.Name != "instance-2" {
		t.Errorf("Unexpected instance: %s", instance.Name)
	}
}

func TestWeightedLB_Select_Empty(t *testing.T) {
	lb := NewWeightedLB()

	instance, err := lb.Select([]*types.DatabaseInstance{})
	if err == nil {
		t.Error("Expected error for empty instances, got nil")
	}
	if instance != nil {
		t.Error("Expected nil instance for empty instances, got non-nil")
	}
}

func TestLeastConnLB_Select(t *testing.T) {
	lb := NewLeastConnLB()

	instances := []*types.DatabaseInstance{
		{Name: "instance-1"},
		{Name: "instance-2"},
	}

	// Test initial selection (all have 0 connections)
	instance, err := lb.Select(instances)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if instance.Name != "instance-1" && instance.Name != "instance-2" {
		t.Errorf("Unexpected instance: %s", instance.Name)
	}
}

func TestLeastConnLB_IncrementDecrement(t *testing.T) {
	lb := NewLeastConnLB()

	lb.Increment("instance-1")
	lb.Increment("instance-1")
	lb.Increment("instance-2")

	if lb.GetConnections("instance-1") != 2 {
		t.Errorf("Expected 2 connections for instance-1, got %d", lb.GetConnections("instance-1"))
	}

	if lb.GetConnections("instance-2") != 1 {
		t.Errorf("Expected 1 connection for instance-2, got %d", lb.GetConnections("instance-2"))
	}

	lb.Decrement("instance-1")
	if lb.GetConnections("instance-1") != 1 {
		t.Errorf("Expected 1 connection after decrement, got %d", lb.GetConnections("instance-1"))
	}

	// Test decrementing below zero
	lb.Decrement("instance-1")
	lb.Decrement("instance-1")
	if lb.GetConnections("instance-1") != 0 {
		t.Errorf("Expected 0 connections after decrementing below zero, got %d", lb.GetConnections("instance-1"))
	}
}

func TestLeastConnLB_Select_WithConnections(t *testing.T) {
	lb := NewLeastConnLB()

	instances := []*types.DatabaseInstance{
		{Name: "instance-1"},
		{Name: "instance-2"},
	}

	// Add connections to instance-1
	lb.Increment("instance-1")
	lb.Increment("instance-1")

	// instance-2 should be selected since it has fewer connections
	instance, err := lb.Select(instances)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if instance.Name != "instance-2" {
		t.Errorf("Expected instance-2 (least connections), got %s", instance.Name)
	}
}

func TestLeastConnLB_Reset(t *testing.T) {
	lb := NewLeastConnLB()

	lb.Increment("instance-1")
	lb.Increment("instance-2")

	lb.Reset()

	if lb.GetConnections("instance-1") != 0 {
		t.Errorf("Expected 0 connections after reset, got %d", lb.GetConnections("instance-1"))
	}

	if lb.GetConnections("instance-2") != 0 {
		t.Errorf("Expected 0 connections after reset, got %d", lb.GetConnections("instance-2"))
	}
}

func TestLoadBalancerFactory_Create(t *testing.T) {
	factory := NewLoadBalancerFactory()

	// Test RoundRobin
	lb := factory.Create(types.ReadStrategyRoundRobin)
	if lb.Name() != "round-robin" {
		t.Errorf("Expected round-robin, got %s", lb.Name())
	}

	// Test Weighted
	lb = factory.Create(types.ReadStrategyWeighted)
	if lb.Name() != "weighted" {
		t.Errorf("Expected weighted, got %s", lb.Name())
	}

	// Test LeastConn
	lb = factory.Create(types.ReadStrategyLeastConn)
	if lb.Name() != "least-conn" {
		t.Errorf("Expected least-conn, got %s", lb.Name())
	}

	// Test default
	lb = factory.Create("invalid")
	if lb.Name() != "round-robin" {
		t.Errorf("Expected round-robin as default, got %s", lb.Name())
	}
}
