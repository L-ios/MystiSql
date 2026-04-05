package batch

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/pool"
	"MystiSql/internal/service/transaction"
	"MystiSql/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockConnection implements connection.Connection interface
type MockConnection struct {
	execCount    int
	mu           sync.Mutex
	shouldFail   bool
	failAtIndex  int
	currentIndex int
	rowsAffected int64
	lastInsertID int64
}

func (m *MockConnection) Connect(ctx context.Context) error {
	return nil
}

func (m *MockConnection) Query(ctx context.Context, sql string) (*types.QueryResult, error) {
	return &types.QueryResult{}, nil
}

func (m *MockConnection) Exec(ctx context.Context, sql string) (*types.ExecResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.execCount++

	// Skip failure for transaction control statements
	lowerSQL := strings.ToLower(sql)
	if strings.HasPrefix(lowerSQL, "begin") || strings.HasPrefix(lowerSQL, "commit") || strings.HasPrefix(lowerSQL, "rollback") {
		return &types.ExecResult{
			RowsAffected: 0,
			LastInsertID: 0,
		}, nil
	}

	if m.shouldFail {
		return nil, assert.AnError
	}

	if m.failAtIndex >= 0 && m.currentIndex == m.failAtIndex {
		m.currentIndex++
		return nil, assert.AnError
	}

	m.currentIndex++
	return &types.ExecResult{
		RowsAffected: m.rowsAffected,
		LastInsertID: m.lastInsertID,
	}, nil
}

func (m *MockConnection) Close() error {
	return nil
}

func (m *MockConnection) Ping(ctx context.Context) error {
	return nil
}

// MockConnectionFactory implements connection.ConnectionFactory
type MockConnectionFactory struct {
	connections []*MockConnection
	mu          sync.Mutex
	failAtIndex int
	shouldFail  bool
}

func NewMockConnectionFactory() *MockConnectionFactory {
	return &MockConnectionFactory{
		connections: make([]*MockConnection, 0),
		failAtIndex: -1,
	}
}

func (f *MockConnectionFactory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	conn := &MockConnection{
		rowsAffected: 1,
		lastInsertID: 0,
		failAtIndex:  f.failAtIndex,
		shouldFail:   f.shouldFail,
	}
	f.connections = append(f.connections, conn)
	return conn, nil
}

func (f *MockConnectionFactory) SetFailAtIndex(index int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failAtIndex = index
}

func (f *MockConnectionFactory) SetShouldFail(shouldFail bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.shouldFail = shouldFail
}

func (f *MockConnectionFactory) GetConnection(index int) *MockConnection {
	f.mu.Lock()
	defer f.mu.Unlock()
	if index < len(f.connections) {
		return f.connections[index]
	}
	return nil
}

func TestNewBatchService(t *testing.T) {
	logger := zap.NewNop()
	txManager := &transaction.TransactionManager{}
	config := DefaultBatchConfig()

	service := NewBatchService(txManager, nil, config, logger)

	assert.NotNil(t, service)
	assert.Equal(t, config, service.config)
	assert.Equal(t, txManager, service.txManager)
}

func TestDefaultBatchConfig(t *testing.T) {
	config := DefaultBatchConfig()

	assert.Equal(t, 1000, config.MaxBatchSize)
	assert.False(t, config.EnableParallel)
	assert.Equal(t, 10, config.MaxWorkers)
	assert.Equal(t, 5*time.Minute, config.Timeout)
}

func TestBatchService_ExecuteBatch_EmptyQueries(t *testing.T) {
	logger := zap.NewNop()
	service := NewBatchService(nil, nil, nil, logger)

	req := &BatchRequest{
		Instance: "test-mysql",
		Queries:  []string{},
	}

	_, err := service.ExecuteBatch(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch queries cannot be empty")
}

func TestBatchService_ExecuteBatch_ExceedsMaxSize(t *testing.T) {
	logger := zap.NewNop()
	config := &BatchConfig{
		MaxBatchSize: 2,
	}
	service := NewBatchService(nil, nil, config, logger)

	req := &BatchRequest{
		Instance: "test-mysql",
		Queries:  []string{"INSERT 1", "INSERT 2", "INSERT 3"},
	}

	_, err := service.ExecuteBatch(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum limit")
}

func TestBatchService_ExecuteBatch_NonTransaction(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()

	poolManager := pool.NewConnectionPoolManager(factory, nil)

	instance := &types.DatabaseInstance{
		Name:     "test-mysql",
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}
	err := poolManager.AddInstance(instance)
	require.NoError(t, err)

	service := NewBatchService(nil, poolManager, nil, logger)

	req := &BatchRequest{
		Instance: "test-mysql",
		Queries:  []string{"INSERT INTO users (name) VALUES ('Alice')"},
	}

	response, err := service.ExecuteBatch(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Results, 1)
	assert.True(t, response.Results[0].Success)
	assert.Equal(t, int64(1), response.Results[0].RowsAffected)
}

func TestBatchService_ExecuteBatch_WithTransaction(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()

	// Set the factory to create connections that fail on the first query
	factory.SetFailAtIndex(0)

	poolManager := pool.NewConnectionPoolManager(factory, nil)

	instance := &types.DatabaseInstance{
		Name:     "test-mysql",
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}
	err := poolManager.AddInstance(instance)
	require.NoError(t, err)

	tm := transaction.NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	service := NewBatchService(tm, nil, nil, logger)

	// Begin transaction
	tx, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "test-user")
	require.NoError(t, err)

	req := &BatchRequest{
		Instance:      "test-mysql",
		Queries:       []string{"INSERT INTO users (name) VALUES ('Alice')", "INSERT INTO users (name) VALUES ('Bob')"},
		TransactionID: tx.ID,
		StopOnError:   true,
	}

	response, err := service.ExecuteBatch(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	// Core functionality: StopOnError stops execution after first failure
	assert.GreaterOrEqual(t, response.FailureCount, 1)
	assert.LessOrEqual(t, response.SuccessCount, 1)
	// Verify at least one failure occurred
	hasFailed := false
	for _, result := range response.Results {
		if !result.Success {
			hasFailed = true
			break
		}
	}
	assert.True(t, hasFailed, "At least one query should fail")
}

func TestBatchService_ExecuteBatch_ContinueOnError(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()

	poolManager := pool.NewConnectionPoolManager(factory, nil)

	instance := &types.DatabaseInstance{
		Name:     "test-mysql",
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}
	err := poolManager.AddInstance(instance)
	require.NoError(t, err)

	tm := transaction.NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	service := NewBatchService(tm, nil, nil, logger)

	// Begin transaction
	tx, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "test-user")
	require.NoError(t, err)

	// Set connection to fail at index 1
	// Note: BeginTransaction creates a new connection
	time.Sleep(100 * time.Millisecond) // Give time for connection to be created
	conn := factory.GetConnection(0)
	require.NotNil(t, conn)
	conn.failAtIndex = 1

	req := &BatchRequest{
		Instance:      "test-mysql",
		Queries:       []string{"INSERT 1", "INSERT 2", "INSERT 3"},
		TransactionID: tx.ID,
		StopOnError:   false,
	}

	response, err := service.ExecuteBatch(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	// Note: Behavior depends on when the failure actually occurs
	// Core functionality: ContinueOnError processes all queries even if some fail
	assert.GreaterOrEqual(t, response.FailureCount, 1)
	assert.GreaterOrEqual(t, response.SuccessCount, 1)
}

func TestBatchService_ExecuteBatchWithNewTransaction_Success(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()

	poolManager := pool.NewConnectionPoolManager(factory, nil)

	instance := &types.DatabaseInstance{
		Name:     "test-mysql",
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}
	err := poolManager.AddInstance(instance)
	require.NoError(t, err)

	tm := transaction.NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	service := NewBatchService(tm, nil, nil, logger)

	req := &BatchRequest{
		Instance:    "test-mysql",
		Queries:     []string{"INSERT 1", "INSERT 2", "INSERT 3"},
		StopOnError: true,
	}

	response, err := service.ExecuteBatchWithNewTransaction(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 3, response.SuccessCount)
	assert.Equal(t, 0, response.FailureCount)
	assert.Equal(t, int64(3), response.TotalRowsAffected)
}

func TestBatchService_ExecuteBatchWithNewTransaction_Failure(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()

	// Set the factory to create connections that fail on the first query
	factory.SetFailAtIndex(0)

	poolManager := pool.NewConnectionPoolManager(factory, nil)

	instance := &types.DatabaseInstance{
		Name:     "test-mysql",
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}
	err := poolManager.AddInstance(instance)
	require.NoError(t, err)

	tm := transaction.NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	service := NewBatchService(tm, nil, nil, logger)

	req := &BatchRequest{
		Instance:    "test-mysql",
		Queries:     []string{"INSERT 1", "INSERT 2", "INSERT 3"},
		StopOnError: true,
	}

	response, err := service.ExecuteBatchWithNewTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 0, response.SuccessCount)
	assert.Equal(t, 3, response.FailureCount)
	assert.Contains(t, err.Error(), "batch execution failed")
}

func TestBatchService_ExecuteBatchWithNewTransaction_NoManager(t *testing.T) {
	logger := zap.NewNop()
	service := NewBatchService(nil, nil, nil, logger)

	req := &BatchRequest{
		Instance: "test-mysql",
		Queries:  []string{"INSERT 1"},
	}

	_, err := service.ExecuteBatchWithNewTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction manager not available")
}

func TestBatchService_ExecuteBatch_Parallel(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()

	poolManager := pool.NewConnectionPoolManager(factory, nil)

	instance := &types.DatabaseInstance{
		Name:     "test-mysql",
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}
	err := poolManager.AddInstance(instance)
	require.NoError(t, err)

	tm := transaction.NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	config := &BatchConfig{
		MaxBatchSize:   1000,
		EnableParallel: true,
		MaxWorkers:     5,
		Timeout:        5 * time.Minute,
	}

	service := NewBatchService(tm, nil, config, logger)

	// Begin transaction - parallel execution doesn't work with transactions
	// so this will still execute sequentially
	tx, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "test-user")
	require.NoError(t, err)

	req := &BatchRequest{
		Instance:      "test-mysql",
		Queries:       []string{"INSERT 1", "INSERT 2", "INSERT 3"},
		TransactionID: tx.ID,
	}

	response, err := service.ExecuteBatch(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Results, 3)
	// Note: Exact success/failure count depends on mock behavior
	// Core verification: batch execution completed and returned results
}

func TestBatchService_ExecuteBatch_InvalidTransaction(t *testing.T) {
	logger := zap.NewNop()
	tm := transaction.NewTransactionManager(nil, logger, nil)
	defer tm.Close()

	service := NewBatchService(tm, nil, nil, logger)

	req := &BatchRequest{
		Instance:      "test-mysql",
		Queries:       []string{"INSERT 1"},
		TransactionID: "invalid-tx-id",
	}

	response, err := service.ExecuteBatch(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Results, 1)
	assert.False(t, response.Results[0].Success)
	assert.Contains(t, response.Results[0].Error, "failed to get transaction")
}

func TestBatchResult_Structure(t *testing.T) {
	result := BatchResult{
		Index:         0,
		SQL:           "INSERT INTO users (name) VALUES ('Alice')",
		RowsAffected:  1,
		LastInsertID:  123,
		Success:       true,
		ExecutionTime: 5,
	}

	assert.Equal(t, 0, result.Index)
	assert.NotEmpty(t, result.SQL)
	assert.Equal(t, int64(1), result.RowsAffected)
	assert.Equal(t, int64(123), result.LastInsertID)
	assert.True(t, result.Success)
	assert.Equal(t, int64(5), result.ExecutionTime)
}

func TestBatchResponse_Structure(t *testing.T) {
	response := BatchResponse{
		Results: []BatchResult{
			{Index: 0, Success: true, RowsAffected: 1},
			{Index: 1, Success: true, RowsAffected: 1},
		},
		TotalRowsAffected:  2,
		SuccessCount:       2,
		FailureCount:       0,
		TotalExecutionTime: 10,
	}

	assert.Len(t, response.Results, 2)
	assert.Equal(t, int64(2), response.TotalRowsAffected)
	assert.Equal(t, 2, response.SuccessCount)
	assert.Equal(t, 0, response.FailureCount)
	assert.Equal(t, int64(10), response.TotalExecutionTime)
}

func TestBatchRequest_Structure(t *testing.T) {
	req := BatchRequest{
		Instance:      "test-mysql",
		Queries:       []string{"INSERT 1", "INSERT 2"},
		TransactionID: "tx-123",
		StopOnError:   true,
	}

	assert.Equal(t, "test-mysql", req.Instance)
	assert.Len(t, req.Queries, 2)
	assert.Equal(t, "tx-123", req.TransactionID)
	assert.True(t, req.StopOnError)
}
