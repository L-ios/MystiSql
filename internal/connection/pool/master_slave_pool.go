package pool

import (
	"context"
	"fmt"
	"sync"

	"MystiSql/internal/connection"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	"go.uber.org/zap"
)

type MasterSlavePool struct {
	mu           sync.RWMutex
	masterPool   connection.ConnectionPool
	slavePools   []connection.ConnectionPool
	topology     *types.MasterSlaveTopology
	factory      connection.ConnectionFactory
	config       *connection.PoolConfig
	logger       *zap.Logger
	instanceName string
}

type MasterSlavePoolOption func(*MasterSlavePool)

func WithMasterSlaveLogger(logger *zap.Logger) MasterSlavePoolOption {
	return func(p *MasterSlavePool) {
		p.logger = logger
	}
}

func NewMasterSlavePool(
	topology *types.MasterSlaveTopology,
	factory connection.ConnectionFactory,
	config *connection.PoolConfig,
	opts ...MasterSlavePoolOption,
) (*MasterSlavePool, error) {
	if topology == nil {
		return nil, fmt.Errorf("topology cannot be nil")
	}
	if topology.Master == nil {
		return nil, fmt.Errorf("master instance cannot be nil")
	}

	logger := zap.NewNop()

	pool := &MasterSlavePool{
		topology:     topology,
		factory:      factory,
		config:       config,
		logger:       logger,
		instanceName: topology.Name,
		slavePools:   make([]connection.ConnectionPool, 0),
	}

	for _, opt := range opts {
		opt(pool)
	}

	masterPool, err := NewConnectionPool(
		topology.Master,
		factory,
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create master pool: %w", err)
	}
	pool.masterPool = masterPool

	for _, slave := range topology.Slaves {
		slavePool, err := NewConnectionPool(
			slave,
			factory,
			config,
		)
		if err != nil {
			pool.logger.Warn("failed to create slave pool, skipping",
				zap.String("slave", slave.Name),
				zap.Error(err))
			continue
		}
		pool.slavePools = append(pool.slavePools, slavePool)
	}

	return pool, nil
}

func (p *MasterSlavePool) GetMasterConnection(ctx context.Context) (connection.Connection, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.masterPool == nil {
		return nil, errors.ErrNoMasterAvailable
	}

	return p.masterPool.GetConnection(ctx)
}

func (p *MasterSlavePool) GetSlaveConnection(ctx context.Context) (connection.Connection, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.slavePools) == 0 {
		return p.masterPool.GetConnection(ctx)
	}

	for _, slavePool := range p.slavePools {
		conn, err := slavePool.GetConnection(ctx)
		if err == nil {
			return conn, nil
		}
		p.logger.Debug("slave pool get connection failed, trying next",
			zap.Error(err))
	}

	return p.masterPool.GetConnection(ctx)
}

func (p *MasterSlavePool) GetConnection(ctx context.Context) (connection.Connection, error) {
	return p.GetMasterConnection(ctx)
}

func (p *MasterSlavePool) ReturnConnection(conn connection.Connection) {
	conn.Close()
}

func (p *MasterSlavePool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error

	if p.masterPool != nil {
		if err := p.masterPool.Close(); err != nil {
			errs = append(errs, fmt.Errorf("master pool close failed: %w", err))
		}
	}

	for i, slavePool := range p.slavePools {
		if err := slavePool.Close(); err != nil {
			errs = append(errs, fmt.Errorf("slave pool %d close failed: %w", i, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

func (p *MasterSlavePool) GetStats() *connection.PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	masterStats := p.masterPool.GetStats()

	totalStats := &connection.PoolStats{
		MaxConnections:    masterStats.MaxConnections,
		MinConnections:    masterStats.MinConnections,
		TotalConnections:  masterStats.TotalConnections,
		IdleConnections:   masterStats.IdleConnections,
		ActiveConnections: masterStats.ActiveConnections,
		AcquireCount:      masterStats.AcquireCount,
		AcquireFailed:     masterStats.AcquireFailed,
		ReleaseCount:      masterStats.ReleaseCount,
	}

	for _, slavePool := range p.slavePools {
		slaveStats := slavePool.GetStats()
		totalStats.TotalConnections += slaveStats.TotalConnections
		totalStats.IdleConnections += slaveStats.IdleConnections
		totalStats.ActiveConnections += slaveStats.ActiveConnections
		totalStats.AcquireCount += slaveStats.AcquireCount
		totalStats.AcquireFailed += slaveStats.AcquireFailed
		totalStats.ReleaseCount += slaveStats.ReleaseCount
	}

	return totalStats
}

func (p *MasterSlavePool) SetMaxConnections(max int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.masterPool != nil {
		p.masterPool.SetMaxConnections(max)
	}

	for _, slavePool := range p.slavePools {
		slavePool.SetMaxConnections(max)
	}
}

func (p *MasterSlavePool) SetMinConnections(min int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.masterPool != nil {
		p.masterPool.SetMinConnections(min)
	}

	for _, slavePool := range p.slavePools {
		slavePool.SetMinConnections(min)
	}
}

func (p *MasterSlavePool) SetMaxIdleTime(duration string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.masterPool != nil {
		p.masterPool.SetMaxIdleTime(duration)
	}

	for _, slavePool := range p.slavePools {
		slavePool.SetMaxIdleTime(duration)
	}
}

func (p *MasterSlavePool) SetMaxLifetime(duration string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.masterPool != nil {
		p.masterPool.SetMaxLifetime(duration)
	}

	for _, slavePool := range p.slavePools {
		slavePool.SetMaxLifetime(duration)
	}
}

func (p *MasterSlavePool) UpdateTopology(topology *types.MasterSlaveTopology) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if topology == nil || topology.Master == nil {
		return fmt.Errorf("invalid topology")
	}

	oldMasterPool := p.masterPool
	oldSlavePools := p.slavePools

	newMasterPool, err := NewConnectionPool(
		topology.Master,
		p.factory,
		p.config,
	)
	if err != nil {
		return fmt.Errorf("failed to create new master pool: %w", err)
	}

	newSlavePools := make([]connection.ConnectionPool, 0)
	for _, slave := range topology.Slaves {
		slavePool, err := NewConnectionPool(
			slave,
			p.factory,
			p.config,
		)
		if err != nil {
			p.logger.Warn("failed to create new slave pool",
				zap.String("slave", slave.Name),
				zap.Error(err))
			continue
		}
		newSlavePools = append(newSlavePools, slavePool)
	}

	p.topology = topology
	p.masterPool = newMasterPool
	p.slavePools = newSlavePools

	go func() {
		if oldMasterPool != nil {
			oldMasterPool.Close()
		}
		for _, oldPool := range oldSlavePools {
			oldPool.Close()
		}
	}()

	return nil
}

func (p *MasterSlavePool) GetTopology() *types.MasterSlaveTopology {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.topology
}

func (p *MasterSlavePool) GetMasterPool() connection.ConnectionPool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.masterPool
}

func (p *MasterSlavePool) GetSlavePools() []connection.ConnectionPool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.slavePools
}

func (p *MasterSlavePool) AddSlave(slave *types.DatabaseInstance) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	slavePool, err := NewConnectionPool(
		slave,
		p.factory,
		p.config,
	)
	if err != nil {
		return fmt.Errorf("failed to create slave pool: %w", err)
	}

	p.slavePools = append(p.slavePools, slavePool)
	p.topology.AddSlave(slave)

	return nil
}

func (p *MasterSlavePool) RemoveSlave(slaveName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, pool := range p.slavePools {
		stats := pool.GetStats()
		if stats != nil {
			p.slavePools = append(p.slavePools[:i], p.slavePools[i+1:]...)
			p.topology.RemoveSlave(slaveName)
			go pool.Close()
			return nil
		}
	}

	return fmt.Errorf("slave pool not found: %s", slaveName)
}

func (p *MasterSlavePool) GetAvailableSlaveCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.slavePools)
}

func (p *MasterSlavePool) IsHealthy() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.masterPool == nil {
		return false
	}

	stats := p.masterPool.GetStats()
	return stats != nil && stats.TotalConnections > 0
}
