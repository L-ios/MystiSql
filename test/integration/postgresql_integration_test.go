//go:build integration
// +build integration

package postgresql_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"MystiSql/internal/connection/postgresql"
	"MystiSql/pkg/types"
)

func TestMain(m *testing.M) {
	if os.Getenv("POSTGRESQL_INTEGRATION_TEST") != "true" {
		fmt.Println("Skipping PostgreSQL integration tests. Set POSTGRESQL_INTEGRATION_TEST=true to run.")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestPostgreSQLConnection_Integration(t *testing.T) {
	config := &types.InstanceConfig{
		Name:     "test-postgresql",
		Type:     "postgresql",
		Host:     getEnvOrDefault("POSTGRESQL_HOST", "localhost"),
		Port:     5432,
		Username: getEnvOrDefault("POSTGRESQL_USER", "postgres"),
		Password: getEnvOrDefault("POSTGRESQL_PASSWORD", "password"),
		Database: getEnvOrDefault("POSTGRESQL_DB", "test"),
	}

	conn, err := postgresql.NewConnection(config)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("Connect", func(t *testing.T) {
		err := conn.PingContext(ctx)
		if err != nil {
			t.Errorf("Failed to ping database: %v", err)
		}
	})

	t.Run("CreateTable", func(t *testing.T) {
		query := `
			CREATE TABLE IF NOT EXISTS test_users (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(100) UNIQUE NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
		_, _, err := conn.Exec(ctx, query)
		if err != nil {
			t.Errorf("Failed to create table: %v", err)
		}
	})

	t.Run("Insert", func(t *testing.T) {
		query := "INSERT INTO test_users (name, email) VALUES ($1, $2) RETURNING id"
		result, err := conn.Query(ctx, query, "Alice", "alice@example.com")
		if err != nil {
			t.Errorf("Failed to insert: %v", err)
		}
		if len(result.Rows) == 0 {
			t.Error("Expected to get returned ID")
		}
	})

	t.Run("Select", func(t *testing.T) {
		query := "SELECT id, name, email FROM test_users WHERE name = $1"
		result, err := conn.Query(ctx, query, "Alice")
		if err != nil {
			t.Errorf("Failed to select: %v", err)
		}
		if len(result.Rows) == 0 {
			t.Error("Expected at least one row")
		}
		if len(result.Columns) != 3 {
			t.Errorf("Expected 3 columns, got %d", len(result.Columns))
		}
	})

	t.Run("Update", func(t *testing.T) {
		query := "UPDATE test_users SET name = $1 WHERE name = $2"
		rowsAffected, _, err := conn.Exec(ctx, query, "Bob", "Alice")
		if err != nil {
			t.Errorf("Failed to update: %v", err)
		}
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		query := "DELETE FROM test_users WHERE name = $1"
		rowsAffected, _, err := conn.Exec(ctx, query, "Bob")
		if err != nil {
			t.Errorf("Failed to delete: %v", err)
		}
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}
	})

	t.Run("DropTable", func(t *testing.T) {
		query := "DROP TABLE IF EXISTS test_users"
		_, _, err := conn.Exec(ctx, query)
		if err != nil {
			t.Errorf("Failed to drop table: %v", err)
		}
	})
}

func TestPostgreSQLConnection_Transaction(t *testing.T) {
	config := &types.InstanceConfig{
		Name:     "test-postgresql-tx",
		Type:     "postgresql",
		Host:     getEnvOrDefault("POSTGRESQL_HOST", "localhost"),
		Port:     5432,
		Username: getEnvOrDefault("POSTGRESQL_USER", "postgres"),
		Password: getEnvOrDefault("POSTGRESQL_PASSWORD", "password"),
		Database: getEnvOrDefault("POSTGRESQL_DB", "test"),
	}

	conn, err := postgresql.NewConnection(config)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	t.Run("CreateTable", func(t *testing.T) {
		query := `
			CREATE TABLE IF NOT EXISTS test_accounts (
				id SERIAL PRIMARY KEY,
				balance DECIMAL(10, 2) NOT NULL
			)
		`
		_, _, err := conn.Exec(ctx, query)
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	})

	t.Run("InsertTestData", func(t *testing.T) {
		_, _, err := conn.Exec(ctx, "INSERT INTO test_accounts (balance) VALUES (100.00), (50.00)")
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	})

	t.Run("BeginTransaction", func(t *testing.T) {
		_, _, err := conn.Exec(ctx, "BEGIN")
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		_, _, err = conn.Exec(ctx, "UPDATE test_accounts SET balance = balance - 10 WHERE id = 1")
		if err != nil {
			t.Errorf("Failed to update account 1: %v", err)
		}

		_, _, err = conn.Exec(ctx, "UPDATE test_accounts SET balance = balance + 10 WHERE id = 2")
		if err != nil {
			t.Errorf("Failed to update account 2: %v", err)
		}

		_, _, err = conn.Exec(ctx, "COMMIT")
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		result, err := conn.Query(ctx, "SELECT balance FROM test_accounts WHERE id = 1")
		if err != nil {
			t.Fatalf("Failed to query balance: %v", err)
		}
		if len(result.Rows) > 0 {
			balance := result.Rows[0][0]
			t.Logf("Account 1 balance after transaction: %v", balance)
		}
	})

	t.Run("Cleanup", func(t *testing.T) {
		_, _, err := conn.Exec(ctx, "DROP TABLE IF EXISTS test_accounts")
		if err != nil {
			t.Errorf("Failed to drop table: %v", err)
		}
	})
}

func TestPostgreSQLConnection_ErrorHandling(t *testing.T) {
	config := &types.InstanceConfig{
		Name:     "test-postgresql-error",
		Type:     "postgresql",
		Host:     getEnvOrDefault("POSTGRESQL_HOST", "localhost"),
		Port:     5432,
		Username: getEnvOrDefault("POSTGRESQL_USER", "postgres"),
		Password: getEnvOrDefault("POSTGRESQL_PASSWORD", "password"),
		Database: getEnvOrDefault("POSTGRESQL_DB", "test"),
	}

	conn, err := postgresql.NewConnection(config)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	t.Run("UniqueConstraintViolation", func(t *testing.T) {
		_, _, err := conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS test_unique (email VARCHAR(100) UNIQUE)")
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		_, _, err = conn.Exec(ctx, "INSERT INTO test_unique (email) VALUES ($1)", "test@example.com")
		if err != nil {
			t.Errorf("First insert should succeed: %v", err)
		}

		_, _, err = conn.Exec(ctx, "INSERT INTO test_unique (email) VALUES ($1)", "test@example.com")
		if err == nil {
			t.Error("Expected unique constraint violation error")
		} else {
			pgErr := postgresql.ParseError(err)
			if pgErr == nil {
				t.Logf("Got error (could not parse as PostgreSQL error): %v", err)
			} else {
				t.Logf("Got PostgreSQL error: code=%s, message=%s", pgErr.Code, pgErr.Message)
			}
		}

		_, _, _ = conn.Exec(ctx, "DROP TABLE IF EXISTS test_unique")
	})

	t.Run("InvalidSQL", func(t *testing.T) {
		_, err := conn.Query(ctx, "SELECT * FROM nonexistent_table_xyz")
		if err == nil {
			t.Error("Expected error for non-existent table")
		}
	})
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
