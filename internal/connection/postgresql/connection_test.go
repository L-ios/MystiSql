package postgresql

import (
	"context"
	"testing"
	"time"

	"MystiSql/pkg/types"

	"go.uber.org/zap"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	if factory == nil {
		t.Error("NewFactory() returned nil")
	}
}

func TestFactory_CreateConnection(t *testing.T) {
	factory := NewFactory()
	instance := &types.DatabaseInstance{
		Name:     "test-postgres",
		Type:     types.DatabaseTypePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "testuser",
		Password: "testpass",
	}

	conn, err := factory.CreateConnection(instance)
	if err != nil {
		t.Errorf("CreateConnection() error: %v", err)
	}
	if conn == nil {
		t.Error("CreateConnection() returned nil connection")
	}
}

func TestConnection_BuildDSN(t *testing.T) {
	tests := []struct {
		name         string
		instance     *types.DatabaseInstance
		wantContains []string
	}{
		{
			name: "basic DSN",
			instance: &types.DatabaseInstance{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				Password: "testpass",
			},
			wantContains: []string{
				"postgres://testuser:testpass@localhost:5432/testdb",
			},
		},
		{
			name: "DSN with SSL",
			instance: &types.DatabaseInstance{
				Host:     "prod-server",
				Port:     5432,
				Database: "proddb",
				Username: "produser",
				Password: "prodpass",
			},
			wantContains: []string{
				"postgres://produser:prodpass@prod-server:5432/proddb",
				"sslmode=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			conn, err := NewConnection(tt.instance, logger)
			if err != nil {
				t.Fatalf("NewConnection() error: %v", err)
			}

			dsn, err := conn.buildDSN(tt.instance)
			if err != nil {
				t.Fatalf("buildDSN() error: %v", err)
			}

			for _, want := range tt.wantContains {
				if !contains(dsn, want) {
					t.Errorf("DSN = %s, want to contain %s", dsn, want)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s[1:], substr))
}

func TestPostgreSQLError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *PostgreSQLError
		want string
	}{
		{
			name: "error with detail",
			err: &PostgreSQLError{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint",
				Detail:  "Key (id)=(1) already exists.",
			},
			want: "[23505] duplicate key value violates unique constraint: Key (id)=(1) already exists.",
		},
		{
			name: "error without detail",
			err: &PostgreSQLError{
				Code:    "23503",
				Message: "foreign key constraint violation",
			},
			want: "[23503] foreign key constraint violation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUniqueViolation(t *testing.T) {
	err := &PostgreSQLError{Code: "23505"}
	if !IsUniqueViolation(err) {
		t.Error("IsUniqueViolation() should return true for code 23505")
	}

	err2 := &PostgreSQLError{Code: "23503"}
	if IsUniqueViolation(err2) {
		t.Error("IsUniqueViolation() should return false for code 23503")
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	err := &PostgreSQLError{Code: "23503"}
	if !IsForeignKeyViolation(err) {
		t.Error("IsForeignKeyViolation() should return true for code 23503")
	}

	err2 := &PostgreSQLError{Code: "23505"}
	if IsForeignKeyViolation(err2) {
		t.Error("IsForeignKeyViolation() should return false for code 23505")
	}
}

func TestIsCheckViolation(t *testing.T) {
	err := &PostgreSQLError{Code: "23514"}
	if !IsCheckViolation(err) {
		t.Error("IsCheckViolation() should return true for code 23514")
	}
}

func TestIsNotNullViolation(t *testing.T) {
	err := &PostgreSQLError{Code: "23502"}
	if !IsNotNullViolation(err) {
		t.Error("IsNotNullViolation() should return true for code 23502")
	}
}

func TestIsSyntaxError(t *testing.T) {
	err := &PostgreSQLError{Code: "42601"}
	if !IsSyntaxError(err) {
		t.Error("IsSyntaxError() should return true for code 42601")
	}
}

func TestIsUndefinedTable(t *testing.T) {
	err := &PostgreSQLError{Code: "42P01"}
	if !IsUndefinedTable(err) {
		t.Error("IsUndefinedTable() should return true for code 42P01")
	}
}

func TestIsUndefinedColumn(t *testing.T) {
	err := &PostgreSQLError{Code: "42703"}
	if !IsUndefinedColumn(err) {
		t.Error("IsUndefinedColumn() should return true for code 42703")
	}
}

func TestGetErrorDetail(t *testing.T) {
	err := &PostgreSQLError{
		Code:       "23505",
		Message:    "duplicate key value violates unique constraint",
		Detail:     "Key (id)=(1) already exists.",
		Constraint: "users_pkey",
		Table:      "users",
		Column:     "id",
	}

	detail := GetErrorDetail(err)
	if detail == nil {
		t.Fatal("GetErrorDetail() returned nil")
	}

	if detail["code"] != "23505" {
		t.Errorf("Expected code 23505, got %s", detail["code"])
	}

	if detail["constraint"] != "users_pkey" {
		t.Errorf("Expected constraint users_pkey, got %s", detail["constraint"])
	}

	if detail["table"] != "users" {
		t.Errorf("Expected table users, got %s", detail["table"])
	}

	if detail["column"] != "id" {
		t.Errorf("Expected column id, got %s", detail["column"])
	}
}

func TestConnection_Close_WhenNil(t *testing.T) {
	conn := &Connection{pool: nil}
	err := conn.Close()
	if err != nil {
		t.Errorf("Close() on nil pool should not return error, got: %v", err)
	}
}

func TestConnection_Ping_WhenNil(t *testing.T) {
	conn := &Connection{pool: nil}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := conn.Ping(ctx)
	if err == nil {
		t.Error("Ping() on nil pool should return error")
	}
}
