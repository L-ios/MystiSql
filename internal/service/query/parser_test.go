package query

import (
	"context"
	"testing"
	"time"

	"MystiSql/pkg/types"
)

func TestParser_Parse(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name         string
		sql          string
		wantType     SQLStatementType
		wantTables   []string
		wantReadOnly bool
		wantEstSize  int
		wantErr      bool
	}{
		{
			name:         "SELECT basic",
			sql:          "SELECT * FROM users",
			wantType:     StatementTypeSelect,
			wantTables:   []string{"users"},
			wantReadOnly: true,
			wantEstSize:  1000,
			wantErr:      false,
		},
		{
			name:         "UPDATE",
			sql:          "UPDATE users SET name = 'test' WHERE id = 1",
			wantType:     StatementTypeUpdate,
			wantTables:   []string{"users"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "UPDATE with leading space",
			sql:          " UPDATE users SET name = 'test'",
			wantType:     StatementTypeUpdate,
			wantTables:   []string{"users"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "UPDATE without WHERE",
			sql:          " UPDATE users SET name = 'test'",
			wantType:     StatementTypeUpdate,
			wantTables:   []string{"users"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "UPDATE",
			sql:          "UPDATE users SET name = 'test' WHERE id = 1",
			wantType:     StatementTypeUpdate,
			wantTables:   []string{"users"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "DELETE",
			sql:          "DELETE FROM users WHERE id = 1",
			wantType:     StatementTypeDelete,
			wantTables:   []string{"users"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "CREATE TABLE",
			sql:          "CREATE TABLE orders (id INT, name VARCHAR(100))",
			wantType:     StatementTypeCreate,
			wantTables:   []string{"orders"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "DROP TABLE",
			sql:          "DROP TABLE users",
			wantType:     StatementTypeDrop,
			wantTables:   []string{"users"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "ALTER TABLE",
			sql:          "ALTER TABLE users ADD COLUMN age INT",
			wantType:     StatementTypeAlter,
			wantTables:   []string{"users"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "Empty string",
			sql:          "",
			wantType:     StatementTypeOther,
			wantTables:   nil,
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      true,
		},
		{
			name:         "Whitespace only",
			sql:          "   ",
			wantType:     StatementTypeOther,
			wantTables:   nil,
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      true,
		},
		{
			name:         "Multi-table SELECT with comma",
			sql:          "SELECT * FROM users, orders",
			wantType:     StatementTypeSelect,
			wantTables:   []string{"users", "orders"},
			wantReadOnly: true,
			wantEstSize:  1000,
			wantErr:      false,
		},
		{
			name:         "SELECT with WHERE clause",
			sql:          "SELECT * FROM users WHERE id = 1",
			wantType:     StatementTypeSelect,
			wantTables:   []string{"users"},
			wantReadOnly: true,
			wantEstSize:  500,
			wantErr:      false,
		},
		{
			name:         "SELECT with JOIN",
			sql:          "SELECT * FROM users JOIN orders ON users.id = orders.user_id",
			wantType:     StatementTypeSelect,
			wantTables:   []string{"users"},
			wantReadOnly: true,
			wantEstSize:  1000,
			wantErr:      false,
		},
		{
			name:         "SELECT with LIMIT",
			sql:          "SELECT * FROM users LIMIT 10",
			wantType:     StatementTypeSelect,
			wantTables:   []string{"users"},
			wantReadOnly: true,
			wantEstSize:  100,
			wantErr:      false,
		},
		{
			name:         "SELECT without WHERE/LIMIT",
			sql:          "SELECT * FROM users",
			wantType:     StatementTypeSelect,
			wantTables:   []string{"users"},
			wantReadOnly: true,
			wantEstSize:  1000,
			wantErr:      false,
		},
		{
			name:         "INSERT without INTO keyword",
			sql:          "INSERT users (name) VALUES ('test')",
			wantType:     StatementTypeInsert,
			wantTables:   nil,
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "UPDATE without WHERE",
			sql:          "UPDATE users SET name = 'test'",
			wantType:     StatementTypeUpdate,
			wantTables:   []string{"users"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "DELETE without WHERE",
			sql:          "DELETE FROM users",
			wantType:     StatementTypeDelete,
			wantTables:   []string{"users"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "CREATE TABLE lowercase",
			sql:          "create table test (id int)",
			wantType:     StatementTypeCreate,
			wantTables:   []string{"test"},
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "SELECT with alias",
			sql:          "SELECT * FROM users u",
			wantType:     StatementTypeSelect,
			wantTables:   []string{"users"},
			wantReadOnly: true,
			wantEstSize:  1000,
			wantErr:      false,
		},
		{
			name:         "SELECT with AS alias",
			sql:          "SELECT * FROM users AS u",
			wantType:     StatementTypeSelect,
			wantTables:   []string{"users"},
			wantReadOnly: true,
			wantEstSize:  1000,
			wantErr:      false,
		},
		{
			name:         "Unknown statement type",
			sql:          "SHOW TABLES",
			wantType:     StatementTypeOther,
			wantTables:   nil,
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
		{
			name:         "GRANT statement",
			sql:          "GRANT SELECT ON users TO app_user",
			wantType:     StatementTypeOther,
			wantTables:   nil,
			wantReadOnly: false,
			wantEstSize:  0,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.sql)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result.StatementType != tt.wantType {
				t.Errorf("StatementType = %v, want %v", result.StatementType, tt.wantType)
			}
			if tt.wantTables == nil {
				if len(result.Tables) != 0 {
					t.Errorf("Tables = %v, want nil", result.Tables)
				}
			} else if len(result.Tables) != len(tt.wantTables) {
				t.Errorf("Tables = %v, want %v", result.Tables, tt.wantTables)
			} else {
				for i := range result.Tables {
					if result.Tables[i] != tt.wantTables[i] {
						t.Errorf("Tables[%d] = %v, want %v", i, result.Tables[i], tt.wantTables[i])
					}
				}
			}
			if result.IsReadOnly != tt.wantReadOnly {
				t.Errorf("IsReadOnly = %v, want %v", result.IsReadOnly, tt.wantReadOnly)
			}
			if result.EstimatedSize != tt.wantEstSize {
				t.Errorf("EstimatedSize = %v, want %v", result.EstimatedSize, tt.wantEstSize)
			}
		})
	}
}

func TestParser_Parse_DefaultValues(t *testing.T) {
	p := NewParser()

	result, err := p.Parse("SELECT * FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.MaxResultSize != 10000 {
		t.Errorf("MaxResultSize = %v, want 10000", result.MaxResultSize)
	}
	if result.QueryTimeout != 30*time.Second {
		t.Errorf("QueryTimeout = %v, want 30s", result.QueryTimeout)
	}
}

func TestParser_Validate(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "Valid SELECT",
			sql:     "SELECT * FROM users",
			wantErr: false,
		},
		{
			name:    "Empty string",
			sql:     "",
			wantErr: true,
		},
		{
			name:    "Contains DROP TABLE",
			sql:     "DROP TABLE users",
			wantErr: true,
		},
		{
			name:    "Contains TRUNCATE",
			sql:     "TRUNCATE TABLE users",
			wantErr: true,
		},
		{
			name:    "Contains DROP DATABASE",
			sql:     "DROP DATABASE mydb",
			wantErr: true,
		},
		{
			name:    "Contains ALTER TABLE",
			sql:     "ALTER TABLE users ADD COLUMN age INT",
			wantErr: true,
		},
		{
			name:    "Safe UPDATE",
			sql:     "UPDATE users SET name = 'test' WHERE id = 1",
			wantErr: false,
		},
		{
			name:    "Safe INSERT",
			sql:     "INSERT INTO users (name) VALUES ('test')",
			wantErr: false,
		},
		{
			name:    "Contains -- comment",
			sql:     "SELECT * FROM users -- comment",
			wantErr: false,
		},
		{
			name:    "Contains /* comment",
			sql:     "SELECT * FROM users /* comment */",
			wantErr: false,
		},
		{
			name:    "DELETE without WHERE",
			sql:     "DELETE FROM users WHERE id = 1",
			wantErr: false,
		},
		{
			name:    "Whitespace only",
			sql:     "   ",
			wantErr: true,
		},
		{
			name:    "DROP TABLE in lowercase",
			sql:     "drop table users",
			wantErr: true,
		},
		{
			name:    "Multiple dangerous keywords - DROP TABLE",
			sql:     "SELECT * FROM users; DROP TABLE users",
			wantErr: true,
		},
		{
			name:    "TRUNCATE in different case",
			sql:     "truncate table users",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Validate(tt.sql)
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParser_SettersAndGetters(t *testing.T) {
	p := NewParser()

	if p.GetMaxResultSize() != 10000 {
		t.Errorf("default MaxResultSize = %v, want 10000", p.GetMaxResultSize())
	}
	if p.GetQueryTimeout() != 30*time.Second {
		t.Errorf("default QueryTimeout = %v, want 30s", p.GetQueryTimeout())
	}

	p.SetMaxResultSize(5000)
	if p.GetMaxResultSize() != 5000 {
		t.Errorf("after SetMaxResultSize(5000), GetMaxResultSize() = %v, want 5000", p.GetMaxResultSize())
	}

	p.SetMaxResultSize(0)
	if p.GetMaxResultSize() != 5000 {
		t.Errorf("after SetMaxResultSize(0), GetMaxResultSize() = %v, want 5000", p.GetMaxResultSize())
	}

	p.SetMaxResultSize(-1)
	if p.GetMaxResultSize() != 5000 {
		t.Errorf("after SetMaxResultSize(-1), GetMaxResultSize() = %v, want 5000", p.GetMaxResultSize())
	}

	p.SetQueryTimeout(10 * time.Second)
	if p.GetQueryTimeout() != 10*time.Second {
		t.Errorf("after SetQueryTimeout(10s), GetQueryTimeout() = %v, want 10s", p.GetQueryTimeout())
	}

	p.SetQueryTimeout(0)
	if p.GetQueryTimeout() != 10*time.Second {
		t.Errorf("after SetQueryTimeout(0), GetQueryTimeout() = %v, want 10s", p.GetQueryTimeout())
	}

	p.SetQueryTimeout(-5 * time.Second)
	if p.GetQueryTimeout() != 10*time.Second {
		t.Errorf("after SetQueryTimeout(-5s), GetQueryTimeout() = %v, want 10s", p.GetQueryTimeout())
	}
}

func TestParser_SetMaxResultSize_AffectsParseResult(t *testing.T) {
	p := NewParser()
	p.SetMaxResultSize(5000)

	result, err := p.Parse("SELECT * FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MaxResultSize != 5000 {
		t.Errorf("Parse result MaxResultSize = %v, want 5000", result.MaxResultSize)
	}
}

func TestParser_SetQueryTimeout_AffectsParseResult(t *testing.T) {
	p := NewParser()
	p.SetQueryTimeout(15 * time.Second)

	result, err := p.Parse("SELECT * FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.QueryTimeout != 15*time.Second {
		t.Errorf("Parse result QueryTimeout = %v, want 15s", result.QueryTimeout)
	}
}

func TestWithTimeout(t *testing.T) {
	ctx, cancel := WithTimeout(context.Background(), 100*time.Millisecond)
	if ctx == nil {
		t.Error("context is nil")
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("context should have a deadline")
	}
	if deadline.Sub(time.Now()) > 200*time.Millisecond {
		t.Error("deadline should be close to 100ms")
	}

	cancel()

	select {
	case <-ctx.Done():
	case <-time.After(10 * time.Millisecond):
		t.Error("context should be done after cancel")
	}

	if ctx.Err() == nil {
		t.Error("context should have error after cancel")
	}
}

func TestWithResultSizeLimit(t *testing.T) {
	smallResult := &types.QueryResult{
		Rows: make([]types.Row, 5),
	}

	result := WithResultSizeLimit(smallResult, 10)
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.Truncated {
		t.Error("should not be truncated when within limit")
	}
	if len(result.Rows) != 5 {
		t.Errorf("Rows length = %v, want 5", len(result.Rows))
	}

	nilResult := WithResultSizeLimit(nil, 10)
	if nilResult != nil {
		t.Error("nil input should return nil")
	}

	largeResult := &types.QueryResult{
		Rows: make([]types.Row, 20),
	}
	truncatedResult := WithResultSizeLimit(largeResult, 10)
	if !truncatedResult.Truncated {
		t.Error("should be truncated when exceeding limit")
	}
	if len(truncatedResult.Rows) != 10 {
		t.Errorf("Rows length = %v, want 10", len(truncatedResult.Rows))
	}

	result1 := &types.QueryResult{
		Rows: make([]types.Row, 20),
	}
	noTruncateZero := WithResultSizeLimit(result1, 0)
	if noTruncateZero.Truncated {
		t.Error("should not truncate when maxSize = 0")
	}

	result2 := &types.QueryResult{
		Rows: make([]types.Row, 20),
	}
	noTruncateNegative := WithResultSizeLimit(result2, -1)
	if noTruncateNegative.Truncated {
		t.Error("should not truncate when maxSize < 0")
	}

	exactLimit := &types.QueryResult{
		Rows: make([]types.Row, 10),
	}
	exactResult := WithResultSizeLimit(exactLimit, 10)
	if exactResult.Truncated {
		t.Error("should not truncate when equal to limit")
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()

	ctxWithUserID := ContextWithUserID(ctx, "user123")
	userID := getUserIDFromContext(ctxWithUserID)
	if userID != "user123" {
		t.Errorf("getUserIDFromContext = %v, want user123", userID)
	}

	ctxWithIP := ContextWithClientIP(ctx, "192.168.1.1")
	clientIP := getClientIPFromContext(ctxWithIP)
	if clientIP != "192.168.1.1" {
		t.Errorf("getClientIPFromContext = %v, want 192.168.1.1", clientIP)
	}

	noUserIDCtx := context.Background()
	missingUserID := getUserIDFromContext(noUserIDCtx)
	if missingUserID != "unknown" {
		t.Errorf("missing user ID = %v, want unknown", missingUserID)
	}

	noIPCtx := context.Background()
	missingIP := getClientIPFromContext(noIPCtx)
	if missingIP != "unknown" {
		t.Errorf("missing client IP = %v, want unknown", missingIP)
	}

	wrongTypeCtx := context.WithValue(ctx, userIDKey, 123)
	wrongTypeUserID := getUserIDFromContext(wrongTypeCtx)
	if wrongTypeUserID != "unknown" {
		t.Errorf("wrong type user ID = %v, want unknown", wrongTypeUserID)
	}

	wrongTypeIPCtx := context.WithValue(ctx, clientIPKey, []byte{1, 2, 3})
	wrongTypeIP := getClientIPFromContext(wrongTypeIPCtx)
	if wrongTypeIP != "unknown" {
		t.Errorf("wrong type IP = %v, want unknown", wrongTypeIP)
	}
}

func TestEstimateResultSize(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name    string
		sql     string
		wantEst int
	}{
		{
			name:    "SELECT with LIMIT",
			sql:     "SELECT * FROM users LIMIT 10",
			wantEst: 100,
		},
		{
			name:    "SELECT with WHERE",
			sql:     "SELECT * FROM users WHERE id = 1",
			wantEst: 500,
		},
		{
			name:    "SELECT without WHERE/LIMIT",
			sql:     "SELECT * FROM users",
			wantEst: 1000,
		},
		{
			name:    "SELECT with both WHERE and LIMIT",
			sql:     "SELECT * FROM users WHERE id = 1 LIMIT 10",
			wantEst: 100,
		},
		{
			name:    "INSERT - non-SELECT",
			sql:     "INSERT INTO users VALUES (1)",
			wantEst: 0,
		},
		{
			name:    "UPDATE - non-SELECT",
			sql:     "UPDATE users SET name = 'test'",
			wantEst: 0,
		},
		{
			name:    "DELETE - non-SELECT",
			sql:     "DELETE FROM users",
			wantEst: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.sql)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.EstimatedSize != tt.wantEst {
				t.Errorf("EstimatedSize = %v, want %v", result.EstimatedSize, tt.wantEst)
			}
		})
	}
}
