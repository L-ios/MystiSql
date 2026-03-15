package pool

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/monitor"
	"MystiSql/pkg/types"
)

type mockConnection struct {
	closed   bool
	healthy  bool
	queryErr error
	execErr  error
	pingErr  error
	mu       sync.Mutex
}

func (m *mockConnection) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healthy = true
	return nil
}

func (m *mockConnection) Query(ctx context.Context, sql string) (*types.QueryResult, error) {
	return &types.QueryResult{RowCount: 1}, m.queryErr
}

func (m *mockConnection) Exec(ctx context.Context, sql string) (*types.ExecResult, error) {
	return &types.ExecResult{RowsAffected: 1}, m.execErr
}

func (m *mockConnection) Ping(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("connection closed")
	}
	return m.pingErr
}

func (m *mockConnection) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.healthy = false
	return nil
}

type mockFactory struct {
	connections []*mockConnection
	index       int
}

func (f *mockFactory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	if f.index < len(f.connections) {
		conn := f.connections[f.index]
		f.index++
		return conn, nil
	}
	conn := &mockConnection{healthy: true}
	f.connections = append(f.connections, conn)
	f.index++
	return conn, nil
}

func createTestPool(t *testing.T, maxConn, minConn int) (*ConnectionPoolImpl, *mockFactory) {
	factory := &mockFactory{}
	instance := &types.DatabaseInstance{
		Name:     "test-instance",
		Host:     "localhost",
		Port:     3306,
		Database: "test",
		Username: "root",
	}
	config := &connection.PoolConfig{
		MaxConnections:    maxConn,
		MinConnections:    minConn,
		MaxIdleTime:       "5m",
		MaxLifetime:       "30m",
		ConnectionTimeout: "10s",
		PingInterval:      "1m",
	}

	collector := monitor.NewCollector()
	pool, err := NewConnectionPool(instance, factory, config, WithMetricsCollector(collector))
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	return pool.(*ConnectionPoolImpl), factory
}

func TestNewConnectionPool(t *testing.T) {
	pool, _ := createTestPool(t, 10, 2)
	defer pool.Close()

	stats := pool.GetStats()
	if stats.MaxConnections != 10 {
		t.Errorf("expected MaxConnections=10, got %d", stats.MaxConnections)
	}

	if stats.MinConnections != 2 {
		t.Errorf("expected MinConnections=2, got %d", stats.MinConnections)
	}
}

func TestGetConnection(t *testing.T) {
	pool, _ := createTestPool(t, 5, 0)
	defer pool.Close()

	ctx := context.Background()
	conn, err := pool.GetConnection(ctx)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}

	stats := pool.GetStats()
	if stats.ActiveConnections != 1 {
		t.Errorf("expected ActiveConnections=1, got %d", stats.ActiveConnections)
	}

	conn.Close()

	stats = pool.GetStats()
	if stats.IdleConnections != 1 {
		t.Errorf("expected IdleConnections=1, got %d", stats.IdleConnections)
	}
}

func TestPoolMetrics(t *testing.T) {
	pool, _ := createTestPool(t, 5, 0)
	defer pool.Close()

	ctx := context.Background()
	conn, err := pool.GetConnection(ctx)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}

	_, _ = conn.Query(ctx, "SELECT 1")
	_, _ = conn.Exec(ctx, "INSERT INTO t VALUES (1)")
	conn.Close()

	stats := pool.GetStats()
	if stats.AcquireCount != 1 {
		t.Errorf("expected AcquireCount=1, got %d", stats.AcquireCount)
	}

	if stats.QueryCount != 1 {
		t.Errorf("expected QueryCount=1, got %d", stats.QueryCount)
	}

	if stats.ExecCount != 1 {
		t.Errorf("expected ExecCount=1, got %d", stats.ExecCount)
	}
}

func TestMaxConnections(t *testing.T) {
	pool, _ := createTestPool(t, 2, 0)
	defer pool.Close()

	ctx := context.Background()

	conn1, err := pool.GetConnection(ctx)
	if err != nil {
		t.Fatalf("failed to get connection 1: %v", err)
	}

	conn2, err := pool.GetConnection(ctx)
	if err != nil {
		t.Fatalf("failed to get connection 2: %v", err)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, err = pool.GetConnection(ctxTimeout)
	if err == nil {
		t.Error("expected error when pool is exhausted")
	}

	conn1.Close()
	conn2.Close()
}

func TestReturnConnection(t *testing.T) {
	pool, _ := createTestPool(t, 5, 0)
	defer pool.Close()

	ctx := context.Background()
	conn, err := pool.GetConnection(ctx)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}

	stats := pool.GetStats()
	if stats.ActiveConnections != 1 {
		t.Errorf("expected ActiveConnections=1, got %d", stats.ActiveConnections)
	}

	conn.Close()

	stats = pool.GetStats()
	if stats.ActiveConnections != 0 {
		t.Errorf("expected ActiveConnections=0, got %d", stats.ActiveConnections)
	}

	if stats.IdleConnections != 1 {
		t.Errorf("expected IdleConnections=1, got %d", stats.IdleConnections)
	}
}

func TestClose(t *testing.T) {
	pool, _ := createTestPool(t, 5, 0)

	ctx := context.Background()
	conn, err := pool.GetConnection(ctx)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}

	if err := pool.Close(); err != nil {
		t.Fatalf("failed to close pool: %v", err)
	}

	_, err = pool.GetConnection(ctx)
	if err == nil {
		t.Error("expected error when getting connection from closed pool")
	}

	_ = conn
}

func TestSetMaxConnections(t *testing.T) {
	pool, _ := createTestPool(t, 10, 0)
	defer pool.Close()

	pool.SetMaxConnections(5)

	stats := pool.GetStats()
	if stats.MaxConnections != 5 {
		t.Errorf("expected MaxConnections=5, got %d", stats.MaxConnections)
	}
}

func TestSetMinConnections(t *testing.T) {
	pool, _ := createTestPool(t, 10, 0)
	defer pool.Close()

	pool.SetMinConnections(3)

	stats := pool.GetStats()
	if stats.MinConnections != 3 {
		t.Errorf("expected MinConnections=3, got %d", stats.MinConnections)
	}
}

func TestConnectionPoolManager(t *testing.T) {
	factory := &mockFactory{}
	config := &connection.PoolConfig{
		MaxConnections:    10,
		MinConnections:    2,
		MaxIdleTime:       "10m",
		MaxLifetime:       "1h",
		ConnectionTimeout: "30s",
		PingInterval:      "1m",
	}

	collector := monitor.NewCollector()
	mgr := NewConnectionPoolManager(factory, config, WithManagerMetrics(collector))
	defer mgr.Close()

	instance := &types.DatabaseInstance{
		Name:     "test-mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "test",
		Username: "root",
	}

	if err := mgr.AddInstance(instance); err != nil {
		t.Fatalf("failed to add instance: %v", err)
	}

	instances := mgr.ListInstances()
	if len(instances) != 1 {
		t.Errorf("expected 1 instance, got %d", len(instances))
	}

	ctx := context.Background()
	conn, err := mgr.GetConnection(ctx, "test-mysql")
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}
	defer conn.Close()

	stats := mgr.GetPoolStats("test-mysql")
	if stats == nil {
		t.Fatal("expected stats for instance")
	}

	if stats.ActiveConnections != 1 {
		t.Errorf("expected ActiveConnections=1, got %d", stats.ActiveConnections)
	}

	allStats := mgr.GetAllPoolStats()
	if len(allStats) != 1 {
		t.Errorf("expected 1 instance in all stats, got %d", len(allStats))
	}
}

func TestConnectionPoolManagerRemoveInstance(t *testing.T) {
	factory := &mockFactory{}
	config := &connection.PoolConfig{
		MaxConnections:    10,
		MinConnections:    0,
		MaxIdleTime:       "10m",
		MaxLifetime:       "1h",
		ConnectionTimeout: "30s",
		PingInterval:      "1m",
	}

	mgr := NewConnectionPoolManager(factory, config)
	defer mgr.Close()

	instance := &types.DatabaseInstance{
		Name:     "test-remove",
		Host:     "localhost",
		Port:     3306,
		Database: "test",
		Username: "root",
	}

	if err := mgr.AddInstance(instance); err != nil {
		t.Fatalf("failed to add instance: %v", err)
	}

	if err := mgr.RemoveInstance("test-remove"); err != nil {
		t.Fatalf("failed to remove instance: %v", err)
	}

	instances := mgr.ListInstances()
	if len(instances) != 0 {
		t.Errorf("expected 0 instances after removal, got %d", len(instances))
	}
}

func TestConcurrentGetConnection(t *testing.T) {
	pool, _ := createTestPool(t, 10, 0)
	defer pool.Close()

	var wg sync.WaitGroup
	errors := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			conn, err := pool.GetConnection(ctx)
			if err != nil {
				errors <- err
				return
			}
			time.Sleep(10 * time.Millisecond)
			conn.Close()
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent get connection error: %v", err)
	}
}

func TestGetMetricsCollector(t *testing.T) {
	factory := &mockFactory{}
	collector := monitor.NewCollector()
	mgr := NewConnectionPoolManager(factory, nil, WithManagerMetrics(collector))

	if mgr.GetMetricsCollector() != collector {
		t.Error("expected same collector instance")
	}
}
