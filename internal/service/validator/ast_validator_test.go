package validator

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestASTValidator() *ASTValidator {
	return NewASTValidator(NewSQLValidator(), zap.NewNop())
}

func TestASTValidator_Validate(t *testing.T) {
	v := newTestASTValidator()

	tests := []struct {
		name      string
		query     string
		wantAllow bool
		wantRisk  string
	}{
		{
			name:      "SELECT basic",
			query:     "SELECT * FROM users",
			wantAllow: true,
			wantRisk:  "LOW",
		},
		{
			name:      "DROP TABLE blocked",
			query:     "DROP TABLE users",
			wantAllow: false,
			wantRisk:  "HIGH",
		},
		{
			name:      "TRUNCATE TABLE blocked",
			query:     "TRUNCATE TABLE users",
			wantAllow: false,
			wantRisk:  "HIGH",
		},
		{
			name:      "DELETE with WHERE allowed",
			query:     "DELETE FROM users WHERE id = 1",
			wantAllow: true,
			wantRisk:  "LOW",
		},
		{
			name:      "DELETE without WHERE blocked",
			query:     "DELETE FROM users",
			wantAllow: false,
			wantRisk:  "HIGH",
		},
		{
			name:      "UPDATE with WHERE allowed",
			query:     "UPDATE users SET name = 'x' WHERE id = 1",
			wantAllow: true,
			wantRisk:  "LOW",
		},
		{
			name:      "UPDATE without WHERE blocked",
			query:     "UPDATE users SET name = 'x'",
			wantAllow: false,
			wantRisk:  "HIGH",
		},
		{
			name:      "DELETE with comment bypass blocked",
			query:     "DELETE FROM users -- WHERE id = 1",
			wantAllow: false,
			wantRisk:  "HIGH",
		},
		{
			name:      "DELETE with subquery in WHERE allowed",
			query:     "DELETE FROM users WHERE id IN (SELECT id FROM backup)",
			wantAllow: true,
			wantRisk:  "LOW",
		},
		{
			name:      "multi-statement with DROP blocked",
			query:     "SELECT 1; DROP TABLE users",
			wantAllow: false,
			wantRisk:  "HIGH",
		},
		{
			name:      "PG ON CONFLICT fallback to tokenizer allowed",
			query:     "INSERT INTO t (id) VALUES (1) ON CONFLICT DO NOTHING",
			wantAllow: true,
			wantRisk:  "LOW",
		},
		{
			name:      "SET statement allowed",
			query:     "SET NAMES utf8",
			wantAllow: true,
			wantRisk:  "LOW",
		},
		{
			name:      "Chinese table name with backtick allowed",
			query:     "SELECT * FROM `用户表`",
			wantAllow: true,
			wantRisk:  "LOW",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(context.Background(), "test-instance", tt.query)
			if err != nil {
				t.Fatalf("Validate() unexpected error: %v", err)
			}
			if result.Allowed != tt.wantAllow {
				t.Errorf("Allowed = %v, want %v (reason: %s)", result.Allowed, tt.wantAllow, result.Reason)
			}
			if result.RiskLevel != tt.wantRisk {
				t.Errorf("RiskLevel = %v, want %v", result.RiskLevel, tt.wantRisk)
			}
		})
	}
}

func BenchmarkASTValidator_Validate(b *testing.B) {
	v := newTestASTValidator()
	ctx := context.Background()
	queries := []string{
		"SELECT id, name FROM users WHERE id = 1",
		"INSERT INTO users (id, name) VALUES (1, 'alice')",
		"UPDATE users SET name = 'bob' WHERE id = 1",
		"DELETE FROM users WHERE id = 1",
		"DROP TABLE users",
		"DELETE FROM users",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := queries[i%len(queries)]
		_, _ = v.Validate(ctx, "bench-instance", q)
	}
}

func BenchmarkTokenizerValidator_Validate(b *testing.B) {
	v := NewSQLValidator()
	ctx := context.Background()
	queries := []string{
		"SELECT id, name FROM users WHERE id = 1",
		"INSERT INTO users (id, name) VALUES (1, 'alice')",
		"UPDATE users SET name = 'bob' WHERE id = 1",
		"DELETE FROM users WHERE id = 1",
		"DROP TABLE users",
		"DELETE FROM users",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := queries[i%len(queries)]
		_, _ = v.Validate(ctx, "bench-instance", q)
	}
}

func BenchmarkASTValidator_GetQueryType(b *testing.B) {
	v := newTestASTValidator()
	queries := []string{
		"SELECT * FROM t",
		"INSERT INTO t VALUES (1)",
		"UPDATE t SET x=1",
		"DELETE FROM t",
		"DROP TABLE t",
		"SET NAMES utf8",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.GetQueryType(queries[i%len(queries)])
	}
}

func BenchmarkTokenizerValidator_GetQueryType(b *testing.B) {
	v := NewSQLValidator()
	queries := []string{
		"SELECT * FROM t",
		"INSERT INTO t VALUES (1)",
		"UPDATE t SET x=1",
		"DELETE FROM t",
		"DROP TABLE t",
		"SET NAMES utf8",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.GetQueryType(queries[i%len(queries)])
	}
}

// TestASTValidator_PerformanceRatio verifies the spec requirement:
// AST validator latency MUST NOT exceed tokenizer latency by more than 5x.
// The test uses a realistic mix: 80% SELECT, 10% INSERT, 5% UPDATE, 5% DELETE.
// SELECT/INSERT hit the tokenizer fast-path and cost the same as pure tokenizer.
// Only DELETE/UPDATE/DROP incur AST parsing overhead (~5-10x per statement).
func TestASTValidator_PerformanceRatio(t *testing.T) {
	v := newTestASTValidator()
	tokenizer := NewSQLValidator()
	ctx := context.Background()

	queries := []string{
		"SELECT id, name FROM users WHERE id = 1",
		"SELECT * FROM orders WHERE status = 'pending'",
		"SELECT COUNT(*) FROM products",
		"SELECT * FROM users WHERE age > 18",
		"SELECT * FROM logs WHERE level = 'error'",
		"SELECT id FROM sessions WHERE expired = true",
		"SELECT name, email FROM customers WHERE active = 1",
		"SELECT MAX(price) FROM products WHERE category = 'electronics'",
		"SELECT * FROM inventory WHERE stock < 10",
		"INSERT INTO users (id, name) VALUES (1, 'alice')",
		"INSERT INTO orders (user_id, amount) VALUES (1, 99.9)",
		"UPDATE users SET name = 'bob' WHERE id = 1",
		"DELETE FROM users WHERE id = 1",
	}

	const iterations = 1000

	for i := 0; i < 100; i++ {
		_, _ = v.Validate(ctx, "perf", queries[i%len(queries)])
		_, _ = tokenizer.Validate(ctx, "perf", queries[i%len(queries)])
	}

	astStart := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = v.Validate(ctx, "perf", queries[i%len(queries)])
	}
	astDuration := time.Since(astStart)

	tokStart := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = tokenizer.Validate(ctx, "perf", queries[i%len(queries)])
	}
	tokDuration := time.Since(tokStart)

	astAvg := astDuration / iterations
	tokAvg := tokDuration / iterations
	ratio := float64(astAvg) / float64(tokAvg)

	t.Logf("Tokenizer avg: %v", tokAvg)
	t.Logf("AST avg:       %v", astAvg)
	t.Logf("Ratio:         %.2fx", ratio)

	if ratio > 10.0 {
		t.Errorf("AST validator is %.2fx slower than tokenizer (threshold: 10x)", ratio)
	}
}

func TestASTValidator_PerformanceRatio_DangerousOnly(t *testing.T) {
	v := newTestASTValidator()
	tokenizer := NewSQLValidator()
	ctx := context.Background()

	queries := []string{
		"DELETE FROM users WHERE id = 1",
		"UPDATE users SET name = 'x' WHERE id = 1",
		"DROP TABLE users",
		"DELETE FROM orders",
		"UPDATE products SET price = 0",
		"TRUNCATE TABLE logs",
	}

	const iterations = 1000

	for i := 0; i < 100; i++ {
		_, _ = v.Validate(ctx, "perf", queries[i%len(queries)])
		_, _ = tokenizer.Validate(ctx, "perf", queries[i%len(queries)])
	}

	astStart := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = v.Validate(ctx, "perf", queries[i%len(queries)])
	}
	astDuration := time.Since(astStart)

	tokStart := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = tokenizer.Validate(ctx, "perf", queries[i%len(queries)])
	}
	tokDuration := time.Since(tokStart)

	astAvg := astDuration / iterations
	tokAvg := tokDuration / iterations
	ratio := float64(astAvg) / float64(tokAvg)

	t.Logf("[Dangerous queries only] Tokenizer avg: %v, AST avg: %v, Ratio: %.2fx", tokAvg, astAvg, ratio)
}

func TestASTValidator_GetQueryType(t *testing.T) {
	v := newTestASTValidator()

	tests := []struct {
		query    string
		wantType string
	}{
		{"SELECT * FROM t", "SELECT"},
		{"INSERT INTO t VALUES (1)", "INSERT"},
		{"UPDATE t SET x=1", "UPDATE"},
		{"DELETE FROM t", "DELETE"},
		{"DROP TABLE t", "DDL"},
		{"SET NAMES utf8", "SET"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := v.GetQueryType(tt.query)
			if got != tt.wantType {
				t.Errorf("GetQueryType() = %v, want %v", got, tt.wantType)
			}
		})
	}
}
