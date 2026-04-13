package topology

import (
	"context"
	"fmt"
	"sync"
	"time"

	"MystiSql/internal/connection/pool"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	"go.uber.org/zap"
)

type TopologyManager struct {
	mu              sync.RWMutex
	topologies      map[string]*types.MasterSlaveTopology
	pools           map[string]*pool.MasterSlavePool
	config          *types.HAConfig
	logger          *zap.Logger
	eventCh         chan types.TopologyEvent
	ctx             context.Context
	cancel          context.CancelFunc
	running         bool
	failoverHandler FailoverHandler
}

type FailoverHandler interface {
	OnFailover(event *FailoverEvent) error
}

type FailoverEvent struct {
	TopologyName string
	OldMaster    string
	NewMaster    string
	Reason       string
	Timestamp    time.Time
	Success      bool
	Error        error
}

type TopologyManagerOption func(*TopologyManager)

func WithTopologyLogger(logger *zap.Logger) TopologyManagerOption {
	return func(m *TopologyManager) {
		m.logger = logger
	}
}

func WithFailoverHandler(handler FailoverHandler) TopologyManagerOption {
	return func(m *TopologyManager) {
		m.failoverHandler = handler
	}
}

func NewTopologyManager(config *types.HAConfig, opts ...TopologyManagerOption) *TopologyManager {
	ctx, cancel := context.WithCancel(context.Background())

	logger := zap.NewNop()

	m := &TopologyManager{
		topologies: make(map[string]*types.MasterSlaveTopology),
		pools:      make(map[string]*pool.MasterSlavePool),
		config:     config,
		logger:     logger,
		eventCh:    make(chan types.TopologyEvent, 100),
		ctx:        ctx,
		cancel:     cancel,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *TopologyManager) RegisterTopology(topology *types.MasterSlaveTopology) error {
	if topology == nil {
		return fmt.Errorf("topology cannot be nil")
	}
	if topology.Name == "" {
		return fmt.Errorf("topology name cannot be empty")
	}
	if topology.Master == nil {
		return fmt.Errorf("master instance cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.topologies[topology.Name]; exists {
		return fmt.Errorf("topology %s already exists", topology.Name)
	}

	m.topologies[topology.Name] = topology
	topology.LastUpdated = time.Now()

	m.logger.Info("Topology registered",
		zap.String("name", topology.Name),
		zap.String("master", topology.Master.Name),
		zap.Int("slave_count", len(topology.Slaves)))

	return nil
}

func (m *TopologyManager) UnregisterTopology(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.topologies[name]; !exists {
		return fmt.Errorf("topology %s not found", name)
	}

	delete(m.topologies, name)
	delete(m.pools, name)

	m.logger.Info("Topology unregistered", zap.String("name", name))
	return nil
}

func (m *TopologyManager) GetTopology(name string) (*types.MasterSlaveTopology, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	topology, exists := m.topologies[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", errors.ErrInstanceNotFound, name)
	}

	return topology, nil
}

func (m *TopologyManager) GetAllTopologies() map[string]*types.MasterSlaveTopology {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*types.MasterSlaveTopology)
	for k, v := range m.topologies {
		result[k] = v
	}
	return result
}

func (m *TopologyManager) GetMaster(topologyName string) (*types.DatabaseInstance, error) {
	topology, err := m.GetTopology(topologyName)
	if err != nil {
		return nil, err
	}

	return topology.Master, nil
}

func (m *TopologyManager) GetSlaves(topologyName string) ([]*types.DatabaseInstance, error) {
	topology, err := m.GetTopology(topologyName)
	if err != nil {
		return nil, err
	}

	return topology.Slaves, nil
}

func (m *TopologyManager) GetAvailableSlaves(topologyName string) ([]*types.DatabaseInstance, error) {
	topology, err := m.GetTopology(topologyName)
	if err != nil {
		return nil, err
	}

	return topology.GetAvailableSlaves(), nil
}

func (m *TopologyManager) PromoteSlave(topologyName, slaveName string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	topology, exists := m.topologies[topologyName]
	if !exists {
		return fmt.Errorf("topology %s not found", topologyName)
	}

	var targetSlave *types.DatabaseInstance
	var targetIndex int
	for i, slave := range topology.Slaves {
		if slave.Name == slaveName {
			targetSlave = slave
			targetIndex = i
			break
		}
	}

	if targetSlave == nil {
		return fmt.Errorf("slave %s not found in topology %s", slaveName, topologyName)
	}

	oldMaster := topology.Master

	if oldMaster.Status == types.InstanceStatusHealthy && !force {
		return fmt.Errorf("master %s is still healthy, use force=true to promote slave", oldMaster.Name)
	}

	oldMaster.Role = string(types.InstanceRoleSlave)
	oldMaster.Master = targetSlave.Name

	targetSlave.Role = string(types.InstanceRoleMaster)
	targetSlave.Master = ""

	topology.Master = targetSlave
	topology.Slaves = append(topology.Slaves[:targetIndex], topology.Slaves[targetIndex+1:]...)
	topology.Slaves = append(topology.Slaves, oldMaster)
	topology.LastUpdated = time.Now()

	event := &FailoverEvent{
		TopologyName: topologyName,
		OldMaster:    oldMaster.Name,
		NewMaster:    targetSlave.Name,
		Reason:       "manual promotion",
		Timestamp:    time.Now(),
		Success:      true,
	}

	if m.failoverHandler != nil {
		if err := m.failoverHandler.OnFailover(event); err != nil {
			m.logger.Error("Failover handler failed", zap.Error(err))
		}
	}

	m.emitEvent(types.TopologyEventMasterChanged, topology)

	m.logger.Info("Slave promoted to master",
		zap.String("topology", topologyName),
		zap.String("old_master", oldMaster.Name),
		zap.String("new_master", targetSlave.Name))

	return nil
}

func (m *TopologyManager) ForceFailover(topologyName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	topology, exists := m.topologies[topologyName]
	if !exists {
		return fmt.Errorf("topology %s not found", topologyName)
	}

	if len(topology.Slaves) == 0 {
		return fmt.Errorf("no slaves available for failover")
	}

	availableSlaves := topology.GetAvailableSlaves()
	if len(availableSlaves) == 0 {
		return fmt.Errorf("no healthy slaves available for failover")
	}

	bestSlave := m.selectBestSlave(availableSlaves)
	if bestSlave == nil {
		return fmt.Errorf("failed to select best slave for failover")
	}

	return m.promoteSlaveInternal(topology, bestSlave, "auto failover")
}

func (m *TopologyManager) selectBestSlave(slaves []*types.DatabaseInstance) *types.DatabaseInstance {
	if len(slaves) == 0 {
		return nil
	}

	var best *types.DatabaseInstance
	bestWeight := -1

	for _, slave := range slaves {
		if slave.Weight > bestWeight {
			bestWeight = slave.Weight
			best = slave
		}
	}

	if best == nil {
		return slaves[0]
	}

	return best
}

func (m *TopologyManager) promoteSlaveInternal(topology *types.MasterSlaveTopology, slave *types.DatabaseInstance, reason string) error {
	oldMaster := topology.Master

	var targetIndex int
	for i, s := range topology.Slaves {
		if s.Name == slave.Name {
			targetIndex = i
			break
		}
	}

	oldMaster.Role = string(types.InstanceRoleSlave)
	oldMaster.Master = slave.Name

	slave.Role = string(types.InstanceRoleMaster)
	slave.Master = ""

	topology.Master = slave
	topology.Slaves = append(topology.Slaves[:targetIndex], topology.Slaves[targetIndex+1:]...)
	topology.Slaves = append(topology.Slaves, oldMaster)
	topology.LastUpdated = time.Now()

	event := &FailoverEvent{
		TopologyName: topology.Name,
		OldMaster:    oldMaster.Name,
		NewMaster:    slave.Name,
		Reason:       reason,
		Timestamp:    time.Now(),
		Success:      true,
	}

	if m.failoverHandler != nil {
		if err := m.failoverHandler.OnFailover(event); err != nil {
			m.logger.Error("Failover handler failed", zap.Error(err))
		}
	}

	m.emitEvent(types.TopologyEventMasterChanged, topology)

	m.logger.Info("Failover completed",
		zap.String("topology", topology.Name),
		zap.String("old_master", oldMaster.Name),
		zap.String("new_master", slave.Name),
		zap.String("reason", reason))

	return nil
}

func (m *TopologyManager) AddSlave(topologyName string, slave *types.DatabaseInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	topology, exists := m.topologies[topologyName]
	if !exists {
		return fmt.Errorf("topology %s not found", topologyName)
	}

	slave.Role = string(types.InstanceRoleSlave)
	slave.Master = topology.Master.Name

	topology.AddSlave(slave)
	topology.LastUpdated = time.Now()

	m.emitEvent(types.TopologyEventSlaveAdded, topology)

	m.logger.Info("Slave added to topology",
		zap.String("topology", topologyName),
		zap.String("slave", slave.Name))

	return nil
}

func (m *TopologyManager) RemoveSlave(topologyName, slaveName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	topology, exists := m.topologies[topologyName]
	if !exists {
		return fmt.Errorf("topology %s not found", topologyName)
	}

	topology.RemoveSlave(slaveName)
	topology.LastUpdated = time.Now()

	m.emitEvent(types.TopologyEventSlaveRemoved, topology)

	m.logger.Info("Slave removed from topology",
		zap.String("topology", topologyName),
		zap.String("slave", slaveName))

	return nil
}

func (m *TopologyManager) UpdateInstanceStatus(topologyName, instanceName string, status types.InstanceStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	topology, exists := m.topologies[topologyName]
	if !exists {
		return fmt.Errorf("topology %s not found", topologyName)
	}

	if topology.Master.Name == instanceName {
		topology.Master.Status = status
		topology.UpdateStatus()
		topology.LastUpdated = time.Now()

		if status == types.InstanceStatusUnhealthy && m.config.Failover.AutoFailover {
			m.logger.Warn("Master is unhealthy, triggering auto failover",
				zap.String("topology", topologyName),
				zap.String("master", instanceName))

			go func() {
				if err := m.ForceFailover(topologyName); err != nil {
					m.logger.Error("Auto failover failed",
						zap.String("topology", topologyName),
						zap.Error(err))
				}
			}()
		}
	} else {
		for _, slave := range topology.Slaves {
			if slave.Name == instanceName {
				slave.Status = status
				topology.UpdateStatus()
				topology.LastUpdated = time.Now()
				break
			}
		}
	}

	m.emitEvent(types.TopologyEventStatusChanged, topology)

	return nil
}

func (m *TopologyManager) emitEvent(eventType types.TopologyEventType, topology *types.MasterSlaveTopology) {
	event := types.TopologyEvent{
		Type:      eventType,
		Topology:  topology,
		Timestamp: time.Now(),
	}

	select {
	case m.eventCh <- event:
	default:
		m.logger.Warn("Event channel full, dropping event",
			zap.String("type", string(eventType)))
	}
}

func (m *TopologyManager) Watch(ctx context.Context) (<-chan types.TopologyEvent, error) {
	return m.eventCh, nil
}

func (m *TopologyManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}

	m.running = true
	m.logger.Info("Topology manager started")

	return nil
}

func (m *TopologyManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.cancel()
	m.running = false
	close(m.eventCh)

	m.logger.Info("Topology manager stopped")
	return nil
}

func (m *TopologyManager) GetTopologyStatus(topologyName string) (*TopologyStatusInfo, error) {
	topology, err := m.GetTopology(topologyName)
	if err != nil {
		return nil, err
	}

	availableSlaves := topology.GetAvailableSlaves()

	masterStatus := "unknown"
	if topology.Master != nil {
		masterStatus = string(topology.Master.Status)
	}

	return &TopologyStatusInfo{
		Name:            topology.Name,
		MasterName:      topology.Master.Name,
		MasterStatus:    masterStatus,
		TotalSlaves:     len(topology.Slaves),
		AvailableSlaves: len(availableSlaves),
		TopologyStatus:  string(topology.Status),
		LastUpdated:     topology.LastUpdated,
	}, nil
}

func (m *TopologyManager) GetAllTopologyStatus() map[string]*TopologyStatusInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*TopologyStatusInfo)
	for name := range m.topologies {
		status, err := m.GetTopologyStatus(name)
		if err == nil {
			result[name] = status
		}
	}
	return result
}

type TopologyStatusInfo struct {
	Name            string    `json:"name"`
	MasterName      string    `json:"masterName"`
	MasterStatus    string    `json:"masterStatus"`
	TotalSlaves     int       `json:"totalSlaves"`
	AvailableSlaves int       `json:"availableSlaves"`
	TopologyStatus  string    `json:"topologyStatus"`
	LastUpdated     time.Time `json:"lastUpdated"`
}
