//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func NewMySQLConnection(t *testing.T, config *MySQLConfig) *sql.DB {
	t.Helper()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err, "Failed to open MySQL connection")

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	require.NoError(t, err, "Failed to ping MySQL database")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func NewPostgreSQLConnection(t *testing.T, config *PostgreSQLConfig) *sql.DB {
	t.Helper()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.Username, config.Password, config.Host, config.Port, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Failed to open PostgreSQL connection")

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	require.NoError(t, err, "Failed to ping PostgreSQL database")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func CleanupTable(t *testing.T, db *sql.DB, whereClause string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", whereClause))
	if err != nil {
		t.Logf("Warning: failed to cleanup: %v", err)
	}
}

func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}
}
