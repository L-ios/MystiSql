package topology

import (
	"testing"

	"MystiSql/pkg/types"
)

func TestTopologyManager_RegisterTopology(t *testing.T) {
	config := &types.HAConfig{
		Enabled: true,
		Failover: types.FailoverConfig{
			AutoFailover: false,
		},
	}

	m := NewTopologyManager(config)

	master := &types.DatabaseInstance{
		Name:   "master-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3306,
		Role:   string(types.InstanceRoleMaster),
		Status: types.InstanceStatusHealthy,
	}

	topology := types.NewMasterSlaveTopology("test-topology", master)

	err := m.RegisterTopology(topology)
	if err != nil {
		t.Fatalf("RegisterTopology failed: %v", err)
	}

	retrieved, err := m.GetTopology("test-topology")
	if err != nil {
		t.Fatalf("GetTopology failed: %v", err)
	}

	if retrieved.Name != "test-topology" {
		t.Errorf("Expected topology name 'test-topology', got %q", retrieved.Name)
	}

	if retrieved.Master.Name != "master-1" {
		t.Errorf("Expected master name 'master-1', got %q", retrieved.Master.Name)
	}
}

func TestTopologyManager_AddSlave(t *testing.T) {
	config := &types.HAConfig{Enabled: true}
	m := NewTopologyManager(config)

	master := &types.DatabaseInstance{
		Name:   "master-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3306,
		Role:   string(types.InstanceRoleMaster),
		Status: types.InstanceStatusHealthy,
	}

	topology := types.NewMasterSlaveTopology("test-topology", master)
	_ = m.RegisterTopology(topology)

	slave := &types.DatabaseInstance{
		Name:   "slave-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3307,
		Status: types.InstanceStatusHealthy,
	}

	err := m.AddSlave("test-topology", slave)
	if err != nil {
		t.Fatalf("AddSlave failed: %v", err)
	}

	retrieved, _ := m.GetTopology("test-topology")
	if len(retrieved.Slaves) != 1 {
		t.Errorf("Expected 1 slave, got %d", len(retrieved.Slaves))
	}

	if retrieved.Slaves[0].Name != "slave-1" {
		t.Errorf("Expected slave name 'slave-1', got %q", retrieved.Slaves[0].Name)
	}
}

func TestTopologyManager_PromoteSlave(t *testing.T) {
	config := &types.HAConfig{Enabled: true}
	m := NewTopologyManager(config)

	master := &types.DatabaseInstance{
		Name:   "master-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3306,
		Role:   string(types.InstanceRoleMaster),
		Status: types.InstanceStatusUnhealthy,
	}

	slave := &types.DatabaseInstance{
		Name:   "slave-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3307,
		Role:   string(types.InstanceRoleSlave),
		Status: types.InstanceStatusHealthy,
	}

	topology := types.NewMasterSlaveTopology("test-topology", master)
	topology.AddSlave(slave)
	_ = m.RegisterTopology(topology)

	err := m.PromoteSlave("test-topology", "slave-1", true)
	if err != nil {
		t.Fatalf("PromoteSlave failed: %v", err)
	}

	retrieved, _ := m.GetTopology("test-topology")
	if retrieved.Master.Name != "slave-1" {
		t.Errorf("Expected new master name 'slave-1', got %q", retrieved.Master.Name)
	}

	if retrieved.Master.Role != string(types.InstanceRoleMaster) {
		t.Errorf("Expected new master role 'master', got %q", retrieved.Master.Role)
	}
}

func TestTopologyManager_GetTopologyStatus(t *testing.T) {
	config := &types.HAConfig{Enabled: true}
	m := NewTopologyManager(config)

	master := &types.DatabaseInstance{
		Name:   "master-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3306,
		Role:   string(types.InstanceRoleMaster),
		Status: types.InstanceStatusHealthy,
	}

	topology := types.NewMasterSlaveTopology("test-topology", master)
	_ = m.RegisterTopology(topology)

	status, err := m.GetTopologyStatus("test-topology")
	if err != nil {
		t.Fatalf("GetTopologyStatus failed: %v", err)
	}

	if status.Name != "test-topology" {
		t.Errorf("Expected status name 'test-topology', got %q", status.Name)
	}

	if status.MasterName != "master-1" {
		t.Errorf("Expected master name 'master-1', got %q", status.MasterName)
	}
}

func TestMasterSlaveTopology_GetAvailableSlaves(t *testing.T) {
	master := &types.DatabaseInstance{
		Name:   "master-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3306,
		Role:   string(types.InstanceRoleMaster),
		Status: types.InstanceStatusHealthy,
	}

	topology := types.NewMasterSlaveTopology("test-topology", master)

	healthySlave := &types.DatabaseInstance{
		Name:   "slave-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3307,
		Role:   string(types.InstanceRoleSlave),
		Status: types.InstanceStatusHealthy,
	}

	unhealthySlave := &types.DatabaseInstance{
		Name:   "slave-2",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3308,
		Role:   string(types.InstanceRoleSlave),
		Status: types.InstanceStatusUnhealthy,
	}

	topology.AddSlave(healthySlave)
	topology.AddSlave(unhealthySlave)

	available := topology.GetAvailableSlaves()
	if len(available) != 1 {
		t.Errorf("Expected 1 available slave, got %d", len(available))
	}

	if len(available) > 0 && available[0].Name != "slave-1" {
		t.Errorf("Expected available slave 'slave-1', got %q", available[0].Name)
	}
}

func TestMasterSlaveTopology_UpdateStatus(t *testing.T) {
	master := &types.DatabaseInstance{
		Name:   "master-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3306,
		Role:   string(types.InstanceRoleMaster),
		Status: types.InstanceStatusHealthy,
	}

	topology := types.NewMasterSlaveTopology("test-topology", master)

	topology.UpdateStatus()
	if topology.Status != types.TopologyStatusHealthy {
		t.Errorf("Expected status 'healthy', got %q", topology.Status)
	}

	master.Status = types.InstanceStatusUnhealthy
	topology.UpdateStatus()
	if topology.Status != types.TopologyStatusFailed {
		t.Errorf("Expected status 'failed', got %q", topology.Status)
	}
}

func TestInstanceHealth(t *testing.T) {
	health := types.NewInstanceHealth("test-instance")

	if health.Name != "test-instance" {
		t.Errorf("Expected name 'test-instance', got %q", health.Name)
	}

	if health.Status != types.HealthStatusChecking {
		t.Errorf("Expected initial status 'checking', got %q", health.Status)
	}

	if health.ConsecutiveFails != 0 {
		t.Errorf("Expected initial consecutive fails 0, got %d", health.ConsecutiveFails)
	}

	if health.ConsecutiveOKs != 0 {
		t.Errorf("Expected initial consecutive OKs 0, got %d", health.ConsecutiveOKs)
	}
}
