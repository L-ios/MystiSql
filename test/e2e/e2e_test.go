package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MystiSql/internal/api/rest"
	"MystiSql/internal/connection"
	"MystiSql/internal/connection/pool"
	"MystiSql/internal/service/audit"
	"MystiSql/internal/service/auth"
	"MystiSql/internal/service/batch"
	"MystiSql/internal/service/query"
	"MystiSql/internal/service/transaction"
	"MystiSql/internal/service/validator"
	"MystiSql/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestE2E_TokenAuthentication tests the complete token authentication flow
func TestE2E_TokenAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	logger := zap.NewNop()
	authService := auth.NewAuthService("test-secret-key", 24*time.Hour, logger)

	// Test: Generate Token
	t.Run("generate_token", func(t *testing.T) {
		req := &auth.GenerateTokenRequest{
			UserID: "admin",
			Role:   "admin",
		}

		resp, err := authService.GenerateToken(context.Background(), req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, "admin", resp.UserID)
		assert.Equal(t, "admin", resp.Role)
		assert.False(t, resp.ExpiresAt.IsZero())

		// Store token for next test
		t.Setenv("TEST_TOKEN", resp.Token)
	})

	// Test: Validate Token
	t.Run("validate_token", func(t *testing.T) {
		token := os.Getenv("TEST_TOKEN")
		if token == "" {
			t.Skip("No token available")
		}

		claims, err := authService.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "admin", claims.UserID)
		assert.Equal(t, "admin", claims.Role)
	})

	// Test: Revoke Token
	t.Run("revoke_token", func(t *testing.T) {
		token := os.Getenv("TEST_TOKEN")
		if token == "" {
			t.Skip("No token available")
		}

		err := authService.RevokeToken(token)
		require.NoError(t, err)

		// Verify token is revoked
		_, err = authService.ValidateToken(token)
		assert.Error(t, err)
	})
}

// TestE2E_AuditLogging tests the complete audit logging flow
func TestE2E_AuditLogging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	logger := zap.NewNop()
	auditService := audit.NewAuditService(logger, &audit.AuditConfig{
		Enabled:       true,
		LogFile:       "/tmp/mystisql-test-audit.log",
		RetentionDays: 1,
	})

	defer auditService.Close()

	// Test: Log SQL execution
	t.Run("log_sql_execution", func(t *testing.T) {
		log := &audit.AuditLog{
			UserID:        "test-user",
			ClientIP:      "127.0.0.1",
			Instance:      "test-mysql",
			Database:      "testdb",
			Query:         "SELECT * FROM users",
			QueryType:     "SELECT",
			ExecutionTime: 10 * time.Millisecond,
			RowsAffected:  100,
			Status:        "success",
			Timestamp:     time.Now(),
		}

		err := auditService.LogExecution(context.Background(), log)
		require.NoError(t, err)
	})

	// Test: Query audit logs
	t.Run("query_audit_logs", func(t *testing.T) {
		logs, err := auditService.QueryLogs(context.Background(), &audit.QueryParams{
			UserID: "test-user",
			Start:  time.Now().Add(-1 * time.Hour),
			End:    time.Now(),
			Page:   1,
			PageSz: 10,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, logs)
	})
}

// TestE2E_SQLValidation tests the complete SQL validation flow
func TestE2E_SQLValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	logger := zap.NewNop()
	validatorService := validator.NewValidatorService(logger, &validator.ValidatorConfig{
		Enabled: true,
		DangerousOperations: []string{
			"DROP",
			"TRUNCATE",
			"DELETE_WITHOUT_WHERE",
		},
		Whitelist: []string{"SELECT * FROM system_config"},
		Blacklist: []string{"DELETE FROM audit_log"},
	})

	// Test: Block dangerous SQL
	t.Run("block_dangerous_sql", func(t *testing.T) {
		tests := []struct {
			name    string
			sql     string
			wantErr bool
		}{
			{"DROP TABLE", "DROP TABLE users", true},
			{"TRUNCATE", "TRUNCATE TABLE logs", true},
			{"DELETE without WHERE", "DELETE FROM users", true},
			{"SELECT safe", "SELECT * FROM users WHERE id = 1", false},
			{"UPDATE with WHERE", "UPDATE users SET name = 'test' WHERE id = 1", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := validatorService.Validate(context.Background(), tt.sql)
				require.NoError(t, err)

				if tt.wantErr {
					assert.False(t, result.Allowed)
					assert.NotEmpty(t, result.Reason)
				} else {
					assert.True(t, result.Allowed)
				}
			})
		}
	})

	// Test: Whitelist bypass
	t.Run("whitelist_bypass", func(t *testing.T) {
		result, err := validatorService.Validate(context.Background(), "SELECT * FROM system_config")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	})

	// Test: Blacklist enforcement
	t.Run("blacklist_enforcement", func(t *testing.T) {
		result, err := validatorService.Validate(context.Background(), "DELETE FROM audit_log WHERE id = 1")
		require.NoError(t, err)
		assert.False(t, result.Allowed)
		assert.Contains(t, result.Reason, "blacklist")
	})
}

// TestE2E_JDBCTransactions tests the complete JDBC transaction flow
func TestE2E_JDBCTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Setup mock connection factory
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

	logger := zap.NewNop()
	tm := transaction.NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	// Test: Begin, Execute, Commit
	t.Run("begin_execute_commit", func(t *testing.T) {
		ctx := context.Background()

		// Begin transaction
		tx, err := tm.BeginTransaction(ctx, "test-mysql", types.IsolationLevelDefault, "test-user")
		require.NoError(t, err)
		assert.NotEmpty(t, tx.ID)
		assert.Equal(t, transaction.StateActive, tx.State)

		// Execute query in transaction
		_, err = tx.Connection.Exec(ctx, "INSERT INTO users (name) VALUES ('test')")
		require.NoError(t, err)

		// Commit transaction
		err = tm.CommitTransaction(ctx, tx.ID)
		require.NoError(t, err)
	})

	// Test: Begin, Execute, Rollback
	t.Run("begin_execute_rollback", func(t *testing.T) {
		ctx := context.Background()

		// Begin transaction
		tx, err := tm.BeginTransaction(ctx, "test-mysql", types.IsolationLevelDefault, "test-user")
		require.NoError(t, err)

		// Execute query in transaction
		_, err = tx.Connection.Exec(ctx, "INSERT INTO users (name) VALUES ('test')")
		require.NoError(t, err)

		// Rollback transaction
		err = tm.RollbackTransaction(ctx, tx.ID)
		require.NoError(t, err)
	})

	// Test: Transaction expiry
	t.Run("transaction_expiry", func(t *testing.T) {
		config := &transaction.TransactionConfig{
			DefaultTimeout:  100 * time.Millisecond,
			MaxTimeout:      1 * time.Second,
			CleanupInterval: 50 * time.Millisecond,
			MaxConcurrent:   10,
		}

		tmExpiry := transaction.NewTransactionManager(poolManager, logger, config)
		defer tmExpiry.Close()

		ctx := context.Background()
		tx, err := tmExpiry.BeginTransaction(ctx, "test-mysql", types.IsolationLevelDefault, "test-user")
		require.NoError(t, err)

		// Wait for expiry
		time.Sleep(150 * time.Millisecond)

		// Transaction should be expired
		_, err = tmExpiry.GetTransaction(tx.ID)
		assert.Error(t, err)
		assert.Equal(t, transaction.ErrTransactionExpired, err)
	})
}

// TestE2E_BatchOperations tests the complete batch operation flow
func TestE2E_BatchOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Setup
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

	logger := zap.NewNop()
	tm := transaction.NewTransactionManager(poolManager, logger, nil)
	defer tm.Close()

	batchService := batch.NewBatchService(tm, nil, logger)

	// Test: Execute batch with transaction
	t.Run("execute_batch_with_transaction", func(t *testing.T) {
		ctx := context.Background()

		// Begin transaction
		tx, err := tm.BeginTransaction(ctx, "test-mysql", types.IsolationLevelDefault, "test-user")
		require.NoError(t, err)

		// Execute batch
		req := &batch.BatchRequest{
			Instance:      "test-mysql",
			Queries:       []string{"INSERT 1", "INSERT 2", "INSERT 3"},
			TransactionID: tx.ID,
			StopOnError:   false,
		}

		resp, err := batchService.ExecuteBatch(ctx, req)
		require.NoError(t, err)
		assert.Len(t, resp.Results, 3)
		assert.Equal(t, 3, resp.SuccessCount)

		// Commit
		err = tm.CommitTransaction(ctx, tx.ID)
		require.NoError(t, err)
	})

	// Test: Execute batch with auto transaction
	t.Run("execute_batch_with_auto_transaction", func(t *testing.T) {
		ctx := context.Background()

		req := &batch.BatchRequest{
			Instance:    "test-mysql",
			Queries:     []string{"INSERT 1", "INSERT 2"},
			StopOnError: true,
		}

		resp, err := batchService.ExecuteBatchWithNewTransaction(ctx, req)
		require.NoError(t, err)
		assert.Len(t, resp.Results, 2)
	})

	// Test: Batch size limit
	t.Run("batch_size_limit", func(t *testing.T) {
		config := &batch.BatchConfig{
			MaxBatchSize: 2,
		}

		limitedService := batch.NewBatchService(tm, config, logger)

		req := &batch.BatchRequest{
			Instance: "test-mysql",
			Queries:  []string{"INSERT 1", "INSERT 2", "INSERT 3"},
		}

		_, err := limitedService.ExecuteBatch(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum limit")
	})
}

// TestE2E_IntegratedFlow tests a complete integrated workflow
func TestE2E_IntegratedFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Setup complete system
	logger := zap.NewNop()

	// 1. Auth service
	authService := auth.NewAuthService("test-secret", 24*time.Hour, logger)

	// 2. Validator service
	validatorService := validator.NewValidatorService(logger, &validator.ValidatorConfig{
		Enabled:             true,
		DangerousOperations: []string{"DROP", "TRUNCATE"},
	})

	// 3. Audit service
	auditService := audit.NewAuditService(logger, &audit.AuditConfig{
		Enabled:       true,
		LogFile:       "/tmp/mystisql-e2e-audit.log",
		RetentionDays: 1,
	})
	defer auditService.Close()

	// Test: Complete authenticated workflow
	t.Run("authenticated_workflow", func(t *testing.T) {
		ctx := context.Background()

		// Step 1: Generate token
		tokenResp, err := authService.GenerateToken(ctx, &auth.GenerateTokenRequest{
			UserID: "testuser",
			Role:   "user",
		})
		require.NoError(t, err)

		// Step 2: Validate token
		claims, err := authService.ValidateToken(tokenResp.Token)
		require.NoError(t, err)
		assert.Equal(t, "testuser", claims.UserID)

		// Step 3: Validate SQL
		validationResult, err := validatorService.Validate(ctx, "SELECT * FROM users")
		require.NoError(t, err)
		assert.True(t, validationResult.Allowed)

		// Step 4: Log execution
		auditLog := &audit.AuditLog{
			UserID:        claims.UserID,
			ClientIP:      "127.0.0.1",
			Instance:      "test-instance",
			Database:      "testdb",
			Query:         "SELECT * FROM users",
			QueryType:     "SELECT",
			ExecutionTime: 5 * time.Millisecond,
			RowsAffected:  10,
			Status:        "success",
			Timestamp:     time.Now(),
		}

		err = auditService.LogExecution(ctx, auditLog)
		require.NoError(t, err)

		// Step 5: Query audit logs
		logs, err := auditService.QueryLogs(ctx, &audit.QueryParams{
			UserID: claims.UserID,
			Start:  time.Now().Add(-1 * time.Hour),
			End:    time.Now(),
			Page:   1,
			PageSz: 10,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, logs)
	})

	// Test: Security enforcement
	t.Run("security_enforcement", func(t *testing.T) {
		ctx := context.Background()

		// Test dangerous SQL is blocked
		result, err := validatorService.Validate(ctx, "DROP TABLE users")
		require.NoError(t, err)
		assert.False(t, result.Allowed)
		assert.Contains(t, result.Reason, "DROP")

		// Test revoked token is rejected
		tokenResp, err := authService.GenerateToken(ctx, &auth.GenerateTokenRequest{
			UserID: "baduser",
			Role:   "user",
		})
		require.NoError(t, err)

		err = authService.RevokeToken(tokenResp.Token)
		require.NoError(t, err)

		_, err = authService.ValidateToken(tokenResp.Token)
		assert.Error(t, err)
	})
}
