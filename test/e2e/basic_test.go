//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	config, err := LoadConfig()
	require.NoError(t, err)

	db, err := sql.Open("mysql", config.MySQL.DSN())
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	assert.NoError(t, err, "Should connect to MySQL")

	var result int
	err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestPostgreSQLBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	config, err := LoadConfig()
	require.NoError(t, err)

	db, err := sql.Open("postgres", config.PostgreSQL.DSN())
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	assert.NoError(t, err, "Should connect to PostgreSQL")

	var result int
	err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestMySQLQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	config, err := LoadConfig()
	require.NoError(t, err)

	db, err := sql.Open("mysql", config.MySQL.DSN())
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	rows, err := db.QueryContext(ctx, "SELECT id, username FROM users LIMIT 5")
	require.NoError(t, err)
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var username string
		err := rows.Scan(&id, &username)
		require.NoError(t, err)
		assert.NotEmpty(t, username)
		count++
	}
	assert.GreaterOrEqual(t, count, 1, "Should have at least 1 user")
}

func TestPostgreSQLQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	config, err := LoadConfig()
	require.NoError(t, err)

	db, err := sql.Open("postgres", config.PostgreSQL.DSN())
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	rows, err := db.QueryContext(ctx, "SELECT id, username FROM users LIMIT 5")
	require.NoError(t, err)
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var username string
		err := rows.Scan(&id, &username)
		require.NoError(t, err)
		assert.NotEmpty(t, username)
		count++
	}
	assert.GreaterOrEqual(t, count, 1, "Should have at least 1 user")
}
