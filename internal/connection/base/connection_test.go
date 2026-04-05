package base

import (
	"context"
	"testing"
	"time"

	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"
)

func TestDefaultPoolConfig(t *testing.T) {
	cfg := DefaultPoolConfig()

	if cfg.MaxOpenConns != 10 {
		t.Errorf("MaxOpenConns = %d, want 10", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 30*time.Minute {
		t.Errorf("ConnMaxLifetime = %v, want 30m", cfg.ConnMaxLifetime)
	}
}

func TestNewSQLConnection(t *testing.T) {
	instance := types.NewDatabaseInstance("test", types.DatabaseTypeMySQL, "localhost", 3306)
	cfg := DefaultPoolConfig()

	conn := NewSQLConnection(instance, cfg)

	if conn == nil {
		t.Fatal("NewSQLConnection returned nil")
	}
	if conn.Instance() != instance {
		t.Error("Instance not set correctly")
	}
	if conn.DB() != nil {
		t.Error("New connection DB should be nil")
	}
}

func TestSQLConnection_Query_NotOpen(t *testing.T) {
	instance := types.NewDatabaseInstance("test", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewSQLConnection(instance, DefaultPoolConfig())

	_, err := conn.Query(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("Query on unopened connection should return error")
	}
	if err != errors.ErrConnectionClosed {
		t.Errorf("expected ErrConnectionClosed, got: %v", err)
	}
}

func TestSQLConnection_Exec_NotOpen(t *testing.T) {
	instance := types.NewDatabaseInstance("test", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewSQLConnection(instance, DefaultPoolConfig())

	_, err := conn.Exec(context.Background(), "INSERT INTO test VALUES (1)")
	if err == nil {
		t.Error("Exec on unopened connection should return error")
	}
	if err != errors.ErrConnectionClosed {
		t.Errorf("expected ErrConnectionClosed, got: %v", err)
	}
}

func TestSQLConnection_Ping_NotOpen(t *testing.T) {
	instance := types.NewDatabaseInstance("test", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewSQLConnection(instance, DefaultPoolConfig())

	err := conn.Ping(context.Background())
	if err == nil {
		t.Error("Ping on unopened connection should return error")
	}
	if err != errors.ErrConnectionClosed {
		t.Errorf("expected ErrConnectionClosed, got: %v", err)
	}
}

func TestSQLConnection_Close_NotOpen(t *testing.T) {
	instance := types.NewDatabaseInstance("test", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewSQLConnection(instance, DefaultPoolConfig())

	// Close on unopened connection should be idempotent
	err := conn.Close()
	if err != nil {
		t.Errorf("Close on unopened connection should return nil, got: %v", err)
	}
}

func TestSQLConnection_Close_Twice(t *testing.T) {
	instance := types.NewDatabaseInstance("test", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewSQLConnection(instance, DefaultPoolConfig())

	// First close (no-op since db is nil)
	err := conn.Close()
	if err != nil {
		t.Errorf("First close should return nil, got: %v", err)
	}

	// Second close should also be nil
	err = conn.Close()
	if err != nil {
		t.Errorf("Second close should return nil, got: %v", err)
	}
}

func TestSQLConnection_Instance_Status(t *testing.T) {
	instance := types.NewDatabaseInstance("test", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewSQLConnection(instance, DefaultPoolConfig())

	// After close, status should be unknown
	_ = conn.Close()
	if instance.Status != types.InstanceStatusUnknown {
		t.Errorf("status after close = %v, want unknown", instance.Status)
	}
}

func TestSQLConnection_CustomPoolConfig(t *testing.T) {
	instance := types.NewDatabaseInstance("test", types.DatabaseTypeMySQL, "localhost", 3306)
	cfg := PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 1 * time.Hour,
	}

	conn := NewSQLConnection(instance, cfg)

	if conn == nil {
		t.Fatal("NewSQLConnection returned nil")
	}
	if conn.poolConfig.MaxOpenConns != 25 {
		t.Errorf("MaxOpenConns = %d, want 25", conn.poolConfig.MaxOpenConns)
	}
	if conn.poolConfig.MaxIdleConns != 10 {
		t.Errorf("MaxIdleConns = %d, want 10", conn.poolConfig.MaxIdleConns)
	}
	if conn.poolConfig.ConnMaxLifetime != 1*time.Hour {
		t.Errorf("ConnMaxLifetime = %v, want 1h", conn.poolConfig.ConnMaxLifetime)
	}
}
