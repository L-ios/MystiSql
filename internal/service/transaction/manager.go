package transaction

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"MystiSql/internal/connection/pool"
	"MystiSql/pkg/types"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrTransactionExpired       = errors.New("transaction expired")
	ErrTransactionNotActive     = errors.New("transaction is not active")
	ErrInvalidConnectionID      = errors.New("invalid connection ID")
	ErrTransactionAlreadyActive = errors.New("transaction already active")
)

type TransactionState string

const (
	StateActive     TransactionState = "active"
	StateCommitted  TransactionState = "committed"
	StateRolledBack TransactionState = "rolled_back"
	StateExpired    TransactionState = "expired"
)

type Transaction struct {
	ID             string
	ConnectionID   string
	Connection     pool.PooledConnection
	State          TransactionState
	IsolationLevel types.IsolationLevel
	Instance       string
	CreatedAt      time.Time
	ExpiresAt      time.Time
	LastActivityAt time.Time
	UserID         string
}

type TransactionManager struct {
	transactions map[string]*Transaction
	poolManager  *pool.ConnectionPoolManager
	mu           sync.RWMutex
	logger       *zap.Logger
	config       *TransactionConfig
}

type TransactionConfig struct {
	DefaultTimeout  time.Duration
	MaxTimeout      time.Duration
	CleanupInterval time.Duration
	MaxConcurrent   int
}

func DefaultTransactionConfig() *TransactionConfig {
	return &TransactionConfig{
		DefaultTimeout:  5 * time.Minute,
		MaxTimeout:      30 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxConcurrent:   100,
	}
}

func NewTransactionManager(poolManager *pool.ConnectionPoolManager, logger *zap.Logger, config *TransactionConfig) *TransactionManager {
	if config == nil {
		config = DefaultTransactionConfig()
	}

	tm := &TransactionManager{
		transactions: make(map[string]*Transaction),
		poolManager:  poolManager,
		logger:       logger,
		config:       config,
	}

	// Start cleanup goroutine
	go tm.cleanupExpiredTransactions()

	return tm
}

func (tm *TransactionManager) BeginTransaction(ctx context.Context, instance string, isolationLevel types.IsolationLevel, userID string) (*Transaction, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check concurrent limit
	if len(tm.transactions) >= tm.config.MaxConcurrent {
		return nil, fmt.Errorf("maximum concurrent transactions reached (%d)", tm.config.MaxConcurrent)
	}

	// Get connection from pool
	conn, err := tm.poolManager.GetConnection(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// Generate transaction ID
	txID := generateTransactionID()
	connID := generateConnectionID()

	now := time.Now()
	tx := &Transaction{
		ID:             txID,
		ConnectionID:   connID,
		Connection:     conn,
		State:          StateActive,
		IsolationLevel: isolationLevel,
		Instance:       instance,
		CreatedAt:      now,
		ExpiresAt:      now.Add(tm.config.DefaultTimeout),
		LastActivityAt: now,
		UserID:         userID,
	}

	// Set isolation level
	if isolationLevel != types.IsolationLevelDefault {
		if err := tm.setIsolationLevel(ctx, tx); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to set isolation level: %w", err)
		}
	}

	// Begin transaction
	if err := tm.begin(ctx, tx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	tm.transactions[txID] = tx

	tm.logger.Info("Transaction started",
		zap.String("tx_id", txID),
		zap.String("instance", instance),
		zap.String("user_id", userID),
		zap.Time("expires_at", tx.ExpiresAt),
	)

	return tx, nil
}

func (tm *TransactionManager) CommitTransaction(ctx context.Context, txID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, exists := tm.transactions[txID]
	if !exists {
		return ErrTransactionNotFound
	}

	if tx.State != StateActive {
		return fmt.Errorf("%w: current state is %s", ErrTransactionNotActive, tx.State)
	}

	if time.Now().After(tx.ExpiresAt) {
		tm.rollbackTransaction(tx, StateExpired)
		return ErrTransactionExpired
	}

	// Commit
	if err := tm.commit(ctx, tx); err != nil {
		tm.rollbackTransaction(tx, StateRolledBack)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	tx.State = StateCommitted
	tx.Connection.Close()
	delete(tm.transactions, txID)

	tm.logger.Info("Transaction committed",
		zap.String("tx_id", txID),
		zap.Duration("duration", time.Since(tx.CreatedAt)),
	)

	return nil
}

func (tm *TransactionManager) RollbackTransaction(ctx context.Context, txID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, exists := tm.transactions[txID]
	if !exists {
		return ErrTransactionNotFound
	}

	if tx.State != StateActive {
		return fmt.Errorf("%w: current state is %s", ErrTransactionNotActive, tx.State)
	}

	tm.rollbackTransaction(tx, StateRolledBack)

	tm.logger.Info("Transaction rolled back",
		zap.String("tx_id", txID),
		zap.Duration("duration", time.Since(tx.CreatedAt)),
	)

	return nil
}

func (tm *TransactionManager) rollbackTransaction(tx *Transaction, state TransactionState) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := tm.rollback(ctx, tx); err != nil {
		tm.logger.Error("Failed to rollback transaction",
			zap.String("tx_id", tx.ID),
			zap.Error(err),
		)
	}

	tx.State = state
	tx.Connection.Close()
	delete(tm.transactions, tx.ID)
}

func (tm *TransactionManager) GetTransaction(txID string) (*Transaction, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tx, exists := tm.transactions[txID]
	if !exists {
		return nil, ErrTransactionNotFound
	}

	if time.Now().After(tx.ExpiresAt) {
		return nil, ErrTransactionExpired
	}

	// Update last activity
	tx.LastActivityAt = time.Now()

	return tx, nil
}

func (tm *TransactionManager) ExtendTransaction(txID string, duration time.Duration) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, exists := tm.transactions[txID]
	if !exists {
		return ErrTransactionNotFound
	}

	if tx.State != StateActive {
		return ErrTransactionNotActive
	}

	newExpiry := time.Now().Add(duration)
	if duration > tm.config.MaxTimeout {
		newExpiry = time.Now().Add(tm.config.MaxTimeout)
	}

	tx.ExpiresAt = newExpiry
	tx.LastActivityAt = time.Now()

	tm.logger.Info("Transaction extended",
		zap.String("tx_id", txID),
		zap.Time("new_expires_at", newExpiry),
	)

	return nil
}

func (tm *TransactionManager) cleanupExpiredTransactions() {
	ticker := time.NewTicker(tm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		tm.mu.Lock()
		now := time.Now()
		for txID, tx := range tm.transactions {
			if now.After(tx.ExpiresAt) {
				tm.logger.Warn("Cleaning up expired transaction",
					zap.String("tx_id", txID),
					zap.Time("expired_at", tx.ExpiresAt),
				)
				tm.rollbackTransaction(tx, StateExpired)
			}
		}
		tm.mu.Unlock()
	}
}

func (tm *TransactionManager) begin(ctx context.Context, tx *Transaction) error {
	_, err := tx.Connection.Exec(ctx, "BEGIN")
	return err
}

func (tm *TransactionManager) commit(ctx context.Context, tx *Transaction) error {
	_, err := tx.Connection.Exec(ctx, "COMMIT")
	return err
}

func (tm *TransactionManager) rollback(ctx context.Context, tx *Transaction) error {
	_, err := tx.Connection.Exec(ctx, "ROLLBACK")
	return err
}

func (tm *TransactionManager) setIsolationLevel(ctx context.Context, tx *Transaction) error {
	var levelStr string
	switch tx.IsolationLevel {
	case types.IsolationLevelReadUncommitted:
		levelStr = "READ UNCOMMITTED"
	case types.IsolationLevelReadCommitted:
		levelStr = "READ COMMITTED"
	case types.IsolationLevelRepeatableRead:
		levelStr = "REPEATABLE READ"
	case types.IsolationLevelSerializable:
		levelStr = "SERIALIZABLE"
	default:
		return nil
	}

	query := fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", levelStr)
	_, err := tx.Connection.Exec(ctx, query)
	return err
}

func (tm *TransactionManager) GetActiveTransactionCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.transactions)
}

func (tm *TransactionManager) ListTransactions() []*Transaction {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	transactions := make([]*Transaction, 0, len(tm.transactions))
	for _, tx := range tm.transactions {
		transactions = append(transactions, tx)
	}
	return transactions
}

func (tm *TransactionManager) Close() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for txID, tx := range tm.transactions {
		tm.rollbackTransaction(tx, StateRolledBack)
		tm.logger.Info("Transaction closed during shutdown",
			zap.String("tx_id", txID),
		)
	}

	return nil
}

func generateTransactionID() string {
	return "tx-" + uuid.New().String()
}

func generateConnectionID() string {
	return "conn-" + uuid.New().String()
}
