package pool

import (
	"context"
	"fmt"
	"sync"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/monitor"
	"MystiSql/pkg/types"
)

type PooledConnection interface {
	Query(ctx context.Context, sql string) (*types.QueryResult, error)
	Exec(ctx context.Context, sql string) (*types.ExecResult, error)
	Close() error
}

type ManagerOption func(*ConnectionPoolManager)

func WithManagerMetrics(collector *monitor.Collector) ManagerOption {
	return func(m *ConnectionPoolManager) {
		m.metricsCollector = collector
	}
}

type ConnectionPoolManager struct {
	pools            map[string]connection.ConnectionPool
	factory          connection.ConnectionFactory
	config           *connection.PoolConfig
	mu               sync.RWMutex
	metricsCollector *monitor.Collector
}

func NewConnectionPoolManager(factory connection.ConnectionFactory, config *connection.PoolConfig, opts ...ManagerOption) *ConnectionPoolManager {
	if config == nil {
		config = &connection.PoolConfig{
			MaxConnections:    10,
			MinConnections:    2,
			MaxIdleTime:       "10m",
			MaxLifetime:       "1h",
			ConnectionTimeout: "30s",
			PingInterval:      "1m",
		}
	}

	mgr := &ConnectionPoolManager{
		pools:            make(map[string]connection.ConnectionPool),
		factory:          factory,
		config:           config,
		metricsCollector: monitor.DefaultCollector(),
	}

	for _, opt := range opts {
		opt(mgr)
	}

	return mgr
}

func (cpm *ConnectionPoolManager) GetConnection(ctx context.Context, instance string) (PooledConnection, error) {
	cpm.mu.RLock()
	pool, exists := cpm.pools[instance]
	cpm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no connection pool found for instance: %s", instance)
	}

	conn, err := pool.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	return &pooledConnectionWrapper{
		conn:      conn,
		pool:      pool,
		instance:  instance,
		collector: cpm.metricsCollector,
	}, nil
}

func (cpm *ConnectionPoolManager) AddInstance(instance *types.DatabaseInstance) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	if _, exists := cpm.pools[instance.Name]; exists {
		return fmt.Errorf("connection pool already exists for instance: %s", instance.Name)
	}

	opts := []PoolOption{}
	if cpm.metricsCollector != nil {
		opts = append(opts, WithMetricsCollector(cpm.metricsCollector))
	}

	pool, err := NewConnectionPool(instance, cpm.factory, cpm.config, opts...)
	if err != nil {
		return fmt.Errorf("failed to create connection pool for instance %s: %w", instance.Name, err)
	}

	cpm.pools[instance.Name] = pool
	return nil
}

func (cpm *ConnectionPoolManager) RemoveInstance(instance string) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	pool, exists := cpm.pools[instance]
	if !exists {
		return fmt.Errorf("no connection pool found for instance: %s", instance)
	}

	if err := pool.Close(); err != nil {
		return fmt.Errorf("failed to close connection pool for instance %s: %w", instance, err)
	}

	delete(cpm.pools, instance)
	return nil
}

func (cpm *ConnectionPoolManager) Close() error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	var lastErr error
	for instance, pool := range cpm.pools {
		if err := pool.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close connection pool for instance %s: %w", instance, err)
		}
	}

	cpm.pools = make(map[string]connection.ConnectionPool)
	return lastErr
}

func (cpm *ConnectionPoolManager) GetPoolStats(instance string) *connection.PoolStats {
	cpm.mu.RLock()
	pool, exists := cpm.pools[instance]
	cpm.mu.RUnlock()

	if !exists {
		return nil
	}

	return pool.GetStats()
}

func (cpm *ConnectionPoolManager) GetAllPoolStats() map[string]*connection.PoolStats {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	stats := make(map[string]*connection.PoolStats)
	for name, pool := range cpm.pools {
		stats[name] = pool.GetStats()
	}

	return stats
}

func (cpm *ConnectionPoolManager) GetMetricsCollector() *monitor.Collector {
	return cpm.metricsCollector
}

func (cpm *ConnectionPoolManager) GetPool(instance string) connection.ConnectionPool {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()
	return cpm.pools[instance]
}

func (cpm *ConnectionPoolManager) ListInstances() []string {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	instances := make([]string, 0, len(cpm.pools))
	for name := range cpm.pools {
		instances = append(instances, name)
	}
	return instances
}

type pooledConnectionWrapper struct {
	conn      connection.Connection
	pool      connection.ConnectionPool
	instance  string
	collector *monitor.Collector
}

func (w *pooledConnectionWrapper) Query(ctx context.Context, sql string) (*types.QueryResult, error) {
	return w.conn.Query(ctx, sql)
}

func (w *pooledConnectionWrapper) Exec(ctx context.Context, sql string) (*types.ExecResult, error) {
	return w.conn.Exec(ctx, sql)
}

func (w *pooledConnectionWrapper) Close() error {
	w.pool.ReturnConnection(w.conn)
	return nil
}
