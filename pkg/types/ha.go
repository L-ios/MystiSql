package types

import "time"

type InstanceRole string

const (
	InstanceRoleMaster InstanceRole = "master"
	InstanceRoleSlave  InstanceRole = "slave"
)

type ReadStrategy string

const (
	ReadStrategyRoundRobin ReadStrategy = "round-robin"
	ReadStrategyWeighted   ReadStrategy = "weighted"
	ReadStrategyLeastConn  ReadStrategy = "least-conn"
)

type TopologyStatus string

const (
	TopologyStatusHealthy  TopologyStatus = "healthy"
	TopologyStatusDegraded TopologyStatus = "degraded"
	TopologyStatusFailed   TopologyStatus = "failed"
)

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusChecking  HealthStatus = "checking"
)

type ReadWriteSplittingConfig struct {
	Enabled       bool         `json:"enabled" yaml:"enabled"`
	ReadStrategy  ReadStrategy `json:"readStrategy" yaml:"readStrategy"`
	ReadAfterWrite string      `json:"readAfterWrite" yaml:"readAfterWrite"`
}

type HealthCheckConfig struct {
	Enabled          bool          `json:"enabled" yaml:"enabled"`
	Interval         time.Duration `json:"interval" yaml:"interval"`
	Timeout          time.Duration `json:"timeout" yaml:"timeout"`
	FailureThreshold int           `json:"failureThreshold" yaml:"failureThreshold"`
	RecoveryThreshold int          `json:"recoveryThreshold" yaml:"recoveryThreshold"`
}

type FailoverConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	AutoFailover   bool          `json:"autoFailover" yaml:"autoFailover"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
	MaxDelay       time.Duration `json:"maxDelay" yaml:"maxDelay"`
}

type HAConfig struct {
	Enabled           bool                   `json:"enabled" yaml:"enabled"`
	ReadWriteSplitting ReadWriteSplittingConfig `json:"readWriteSplitting" yaml:"readWriteSplitting"`
	HealthCheck       HealthCheckConfig      `json:"healthCheck" yaml:"healthCheck"`
	Failover          FailoverConfig         `json:"failover" yaml:"failover"`
}

type ClusterConfig struct {
	Name        string   `json:"name" yaml:"name"`
	Kubeconfig  string   `json:"kubeconfig" yaml:"kubeconfig"`
	Namespaces  []string `json:"namespaces" yaml:"namespaces"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
}

type MultiClusterConfig struct {
	Enabled  bool            `json:"enabled" yaml:"enabled"`
	Clusters []ClusterConfig `json:"clusters" yaml:"clusters"`
}

type MasterSlaveTopology struct {
	Name        string             `json:"name" yaml:"name"`
	Master      *DatabaseInstance  `json:"master" yaml:"master"`
	Slaves      []*DatabaseInstance `json:"slaves" yaml:"slaves"`
	Status      TopologyStatus     `json:"status" yaml:"status"`
	LastUpdated time.Time          `json:"lastUpdated" yaml:"lastUpdated"`
}

func NewMasterSlaveTopology(name string, master *DatabaseInstance) *MasterSlaveTopology {
	return &MasterSlaveTopology{
		Name:        name,
		Master:      master,
		Slaves:      make([]*DatabaseInstance, 0),
		Status:      TopologyStatusHealthy,
		LastUpdated: time.Now(),
	}
}

func (t *MasterSlaveTopology) AddSlave(slave *DatabaseInstance) {
	t.Slaves = append(t.Slaves, slave)
	t.LastUpdated = time.Now()
}

func (t *MasterSlaveTopology) RemoveSlave(name string) {
	for i, slave := range t.Slaves {
		if slave.Name == name {
			t.Slaves = append(t.Slaves[:i], t.Slaves[i+1:]...)
			t.LastUpdated = time.Now()
			return
		}
	}
}

func (t *MasterSlaveTopology) GetAvailableSlaves() []*DatabaseInstance {
	available := make([]*DatabaseInstance, 0)
	for _, slave := range t.Slaves {
		if slave.Status == InstanceStatusHealthy {
			available = append(available, slave)
		}
	}
	return available
}

func (t *MasterSlaveTopology) UpdateStatus() {
	if t.Master == nil || t.Master.Status == InstanceStatusUnhealthy {
		t.Status = TopologyStatusFailed
		t.LastUpdated = time.Now()
		return
	}
	
	availableSlaves := t.GetAvailableSlaves()
	if len(availableSlaves) == 0 && len(t.Slaves) > 0 {
		t.Status = TopologyStatusDegraded
	} else {
		t.Status = TopologyStatusHealthy
	}
	t.LastUpdated = time.Now()
}

type InstanceHealth struct {
	Name             string        `json:"name" yaml:"name"`
	Status           HealthStatus  `json:"status" yaml:"status"`
	ConsecutiveFails int           `json:"consecutiveFails" yaml:"consecutiveFails"`
	ConsecutiveOKs   int           `json:"consecutiveOKs" yaml:"consecutiveOKs"`
	LastCheck        time.Time     `json:"lastCheck" yaml:"lastCheck"`
	LastError        string        `json:"lastError,omitempty" yaml:"lastError,omitempty"`
	ResponseTime     time.Duration `json:"responseTime" yaml:"responseTime"`
}

func NewInstanceHealth(name string) *InstanceHealth {
	return &InstanceHealth{
		Name:             name,
		Status:           HealthStatusChecking,
		ConsecutiveFails: 0,
		ConsecutiveOKs:   0,
		LastCheck:        time.Now(),
	}
}

type TopologyEvent struct {
	Type      TopologyEventType `json:"type" yaml:"type"`
	Topology  *MasterSlaveTopology `json:"topology" yaml:"topology"`
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
}

type TopologyEventType string

const (
	TopologyEventMasterChanged   TopologyEventType = "master_changed"
	TopologyEventSlaveAdded      TopologyEventType = "slave_added"
	TopologyEventSlaveRemoved    TopologyEventType = "slave_removed"
	TopologyEventStatusChanged   TopologyEventType = "status_changed"
)
