package transaction

import (
	"context"
	"sync"
	"testing"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/pool"
	"MystiSql/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockConnection implements connection.Connection interface
type MockConnection struct {
	beginCount    int
	commitCount   int
	rollbackCount int
	execCount     int
	closed        bool
	mu            sync.Mutex
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
	return &types.ExecResult{}, nil
}

func (m *MockConnection) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockConnection) Ping(ctx context.Context) error {
	return nil
}

func (m *MockConnection) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// MockConnectionFactory implements connection.ConnectionFactory
type MockConnectionFactory struct {
	connections []*MockConnection
	callCount   int
	mu          sync.Mutex
}

func NewMockConnectionFactory() *MockConnectionFactory {
	return &MockConnectionFactory{
		connections: make([]*MockConnection, 0),
	}
}

func (f *MockConnectionFactory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	conn := &MockConnection{}
	f.connections = append(f.connections, conn)
	f.callCount++
	return conn, nil
}

func (f *MockConnectionFactory) GetConnection(index int) *MockConnection {
	f.mu.Lock()
	defer f.mu.Unlock()
	if index < len(f.connections) {
		return f.connections[index]
	}
	return nil
}

func TestNewTransactionManager(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()
	config := &connection.PoolConfig{
		MaxConnections:    10,
		MinConnections:    2,
		MaxIdleTime:       "10m",
		MaxLifetime:       "1h",
		ConnectionTimeout: "30s",
		PingInterval:      "1m",
	}

	poolManager := pool.NewConnectionPoolManager(factory, config)
	tmConfig := DefaultTransactionConfig()

	tm := NewTransactionManager(poolManager, logger, tmConfig)

	assert.NotNil(t, tm)
	assert.Equal(t, tmConfig, tm.config)
	assert.NotNil(t, tm.transactions)
	assert.Equal(t, 0, tm.GetActiveTransactionCount())

	// Cleanup
	tm.Close()
}

func TestDefaultTransactionConfig(t *testing.T) {
	config := DefaultTransactionConfig()

	assert.Equal(t, 5*time.Minute, config.DefaultTimeout)
	assert.Equal(t, 30*time.Minute, config.MaxTimeout)
	assert.Equal(t, 1*time.Minute, config.CleanupInterval)
	assert.Equal(t, 100, config.MaxConcurrent)
}

func TestTransactionStates(t *testing.T) {
	tests := []struct {
		name  string
		state TransactionState
		want  string
	}{
		{"active", StateActive, "active"},
		{"committed", StateCommitted, "committed"},
		{"rolled_back", StateRolledBack, "rolled_back"},
		{"expired", StateExpired, "expired"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.state))
		})
	}
}

func TestTransactionManager_BeginTransaction(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()

	poolManager := pool.NewConnectionPoolManager(factory, nil)

	// Add instance
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

	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	ctx := context.Background()
	tx, err := tm.BeginTransaction(ctx, "test-mysql", types.IsolationLevelDefault, "user1")

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.NotEmpty(t, tx.ID)
	assert.NotEmpty(t, tx.ConnectionID)
	assert.Equal(t, StateActive, tx.State)
	assert.Equal(t, "test-mysql", tx.Instance)
	assert.Equal(t, "user1", tx.UserID)
	assert.False(t, tx.CreatedAt.IsZero())
	assert.False(t, tx.ExpiresAt.IsZero())
	assert.True(t, tx.ExpiresAt.After(tx.CreatedAt))

	// Verify transaction is stored
	assert.Equal(t, 1, tm.GetActiveTransactionCount())

	// Verify connection was created and BEGIN was called
	conn := factory.GetConnection(0)
	assert.NotNil(t, conn)
	assert.Equal(t, 1, conn.execCount) // BEGIN only (no isolation level set)
}

func TestTransactionManager_BeginTransaction_WithIsolationLevel(t *testing.T) {
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

	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	ctx := context.Background()
	tx, err := tm.BeginTransaction(ctx, "test-mysql", types.IsolationLevelReadCommitted, "user1")

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, types.IsolationLevelReadCommitted, tx.IsolationLevel)

	// Verify BEGIN and SET TRANSACTION were called
	conn := factory.GetConnection(0)
	assert.NotNil(t, conn)
	assert.Equal(t, 2, conn.execCount) // SET TRANSACTION + BEGIN
}

func TestTransactionManager_BeginTransaction_MaxConcurrent(t *testing.T) {
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

	config := &TransactionConfig{
		DefaultTimeout:  5 * time.Minute,
		MaxTimeout:      30 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxConcurrent:   2,
	}

	tm := NewTransactionManager(poolManager, logger, config)
	defer tm.Close()

	// Create 2 transactions (max limit)
	for i := 0; i < 2; i++ {
		_, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
		require.NoError(t, err)
	}

	// Third transaction should fail
	_, err = tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum concurrent transactions reached")
}

func TestTransactionManager_CommitTransaction(t *testing.T) {
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

	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	// Begin transaction
	tx, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
	require.NoError(t, err)

	// Commit transaction
	err = tm.CommitTransaction(context.Background(), tx.ID)

	assert.NoError(t, err)
	assert.Equal(t, 0, tm.GetActiveTransactionCount())

	// Verify COMMIT was called
	conn := factory.GetConnection(0)
	assert.NotNil(t, conn)
	assert.Equal(t, 2, conn.execCount) // BEGIN + COMMIT
}

func TestTransactionManager_CommitTransaction_NotFound(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()

	poolManager := pool.NewConnectionPoolManager(factory, nil)
	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	err := tm.CommitTransaction(context.Background(), "non-existent-tx")

	assert.Error(t, err)
	assert.Equal(t, ErrTransactionNotFound, err)
}

func TestTransactionManager_CommitTransaction_NotActive(t *testing.T) {
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

	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	// Begin and commit transaction
	tx, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
	require.NoError(t, err)
	err = tm.CommitTransaction(context.Background(), tx.ID)
	require.NoError(t, err)

	// Try to commit again
	err = tm.CommitTransaction(context.Background(), tx.ID)

	assert.Error(t, err)
	assert.Equal(t, ErrTransactionNotFound, err)
}

func TestTransactionManager_RollbackTransaction(t *testing.T) {
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

	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	// Begin transaction
	tx, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
	require.NoError(t, err)

	// Rollback transaction
	err = tm.RollbackTransaction(context.Background(), tx.ID)

	assert.NoError(t, err)
	assert.Equal(t, 0, tm.GetActiveTransactionCount())

	// Verify ROLLBACK was called
	conn := factory.GetConnection(0)
	assert.NotNil(t, conn)
	assert.Equal(t, 2, conn.execCount) // BEGIN + ROLLBACK
	// Note: Connection closed state is managed by pool
}

func TestTransactionManager_RollbackTransaction_NotFound(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()

	poolManager := pool.NewConnectionPoolManager(factory, nil)
	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	err := tm.RollbackTransaction(context.Background(), "non-existent-tx")

	assert.Error(t, err)
	assert.Equal(t, ErrTransactionNotFound, err)
}

func TestTransactionManager_GetTransaction(t *testing.T) {
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

	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	// Begin transaction
	tx, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
	require.NoError(t, err)

	// Get transaction
	retrievedTx, err := tm.GetTransaction(tx.ID)

	assert.NoError(t, err)
	assert.NotNil(t, retrievedTx)
	assert.Equal(t, tx.ID, retrievedTx.ID)
	assert.Equal(t, tx.ConnectionID, retrievedTx.ConnectionID)
	assert.Equal(t, tx.Instance, retrievedTx.Instance)
	assert.Equal(t, tx.UserID, retrievedTx.UserID)
}

func TestTransactionManager_GetTransaction_NotFound(t *testing.T) {
	logger := zap.NewNop()
	factory := NewMockConnectionFactory()

	poolManager := pool.NewConnectionPoolManager(factory, nil)
	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	_, err := tm.GetTransaction("non-existent-tx")

	assert.Error(t, err)
	assert.Equal(t, ErrTransactionNotFound, err)
}

func TestTransactionManager_ExtendTransaction(t *testing.T) {
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

	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	// Begin transaction
	tx, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
	require.NoError(t, err)

	originalExpiry := tx.ExpiresAt

	// Extend transaction
	duration := 10 * time.Minute
	err = tm.ExtendTransaction(tx.ID, duration)

	assert.NoError(t, err)

	// Verify expiry was extended
	retrievedTx, err := tm.GetTransaction(tx.ID)
	require.NoError(t, err)
	assert.True(t, retrievedTx.ExpiresAt.After(originalExpiry))
}

func TestTransactionManager_ExtendTransaction_MaxTimeout(t *testing.T) {
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

	config := &TransactionConfig{
		DefaultTimeout:  5 * time.Minute,
		MaxTimeout:      10 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxConcurrent:   100,
	}

	tm := NewTransactionManager(poolManager, logger, config)
	defer tm.Close()

	// Begin transaction
	tx, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
	require.NoError(t, err)

	// Try to extend beyond MaxTimeout
	duration := 60 * time.Minute
	err = tm.ExtendTransaction(tx.ID, duration)

	assert.NoError(t, err)

	// Verify expiry was capped at MaxTimeout
	retrievedTx, err := tm.GetTransaction(tx.ID)
	require.NoError(t, err)
	expectedExpiry := time.Now().Add(config.MaxTimeout)
	assert.WithinDuration(t, expectedExpiry, retrievedTx.ExpiresAt, 1*time.Second)
}

func TestTransactionManager_ListTransactions(t *testing.T) {
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

	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	// Begin 3 transactions
	for i := 0; i < 3; i++ {
		_, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
		require.NoError(t, err)
	}

	// List transactions
	transactions := tm.ListTransactions()

	assert.Len(t, transactions, 3)
	for _, tx := range transactions {
		assert.NotEmpty(t, tx.ID)
		assert.Equal(t, "test-mysql", tx.Instance)
		assert.Equal(t, StateActive, tx.State)
	}
}

func TestTransactionManager_TransactionExpiry(t *testing.T) {
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

	config := &TransactionConfig{
		DefaultTimeout:  100 * time.Millisecond, // Very short timeout for testing
		MaxTimeout:      30 * time.Minute,
		CleanupInterval: 50 * time.Millisecond,
		MaxConcurrent:   100,
	}

	tm := NewTransactionManager(poolManager, logger, config)
	defer tm.Close()

	// Begin transaction
	tx, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
	require.NoError(t, err)

	// Wait for expiry and cleanup
	time.Sleep(150 * time.Millisecond)

	// Try to get expired transaction (should be cleaned up and return not found)
	_, err = tm.GetTransaction(tx.ID)

	assert.Error(t, err)
	assert.Equal(t, ErrTransactionNotFound, err)
}

func TestTransactionManager_ConcurrentAccess(t *testing.T) {
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

	tm := NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	// Concurrent transaction creation
	var wg sync.WaitGroup
	errChan := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
			errChan <- err
		}()
	}

	wg.Wait()
	close(errChan)

	// Verify all transactions created successfully
	errorCount := 0
	for err := range errChan {
		if err != nil {
			errorCount++
		}
	}

	assert.Equal(t, 0, errorCount)
	assert.Equal(t, 10, tm.GetActiveTransactionCount())
}

func TestTransactionManager_Close(t *testing.T) {
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

	tm := NewTransactionManager(poolManager, logger, nil)

	// Begin 3 transactions
	for i := 0; i < 3; i++ {
		_, err := tm.BeginTransaction(context.Background(), "test-mysql", types.IsolationLevelDefault, "user1")
		require.NoError(t, err)
	}

	// Close should rollback all active transactions
	err = tm.Close()

	assert.NoError(t, err)
	assert.Equal(t, 0, tm.GetActiveTransactionCount())

	// Note: Connection closed state is managed by pool, not directly accessible
	// The key verification is that transaction count is 0 after close
}

func TestTransactionErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"not found", ErrTransactionNotFound, "transaction not found"},
		{"expired", ErrTransactionExpired, "transaction expired"},
		{"not active", ErrTransactionNotActive, "transaction is not active"},
		{"invalid connection ID", ErrInvalidConnectionID, "invalid connection ID"},
		{"already active", ErrTransactionAlreadyActive, "transaction already active"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Contains(t, tt.err.Error(), tt.expected)
		})
	}
}

func TestGenerateIDs(t *testing.T) {
	txID1 := generateTransactionID()
	txID2 := generateTransactionID()
	connID1 := generateConnectionID()
	connID2 := generateConnectionID()

	// IDs should be unique
	assert.NotEqual(t, txID1, txID2)
	assert.NotEqual(t, connID1, connID2)

	// IDs should have proper prefix
	assert.Contains(t, txID1, "tx-")
	assert.Contains(t, connID1, "conn-")

	// IDs should not be empty
	assert.NotEmpty(t, txID1)
	assert.NotEmpty(t, connID1)
}
