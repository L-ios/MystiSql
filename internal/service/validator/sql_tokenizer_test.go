package validator

import (
	"context"
	"testing"
)

func TestSQLTokenizer_Tokenize(t *testing.T) {
	tokenizer := newSQLTokenizer()

	tests := []struct {
		name      string
		sql       string
		wantTypes []string
	}{
		{name: "simple SELECT", sql: "SELECT * FROM users", wantTypes: []string{"SELECT"}},
		{name: "simple INSERT", sql: "INSERT INTO t VALUES (1)", wantTypes: []string{"INSERT"}},
		{name: "simple UPDATE", sql: "UPDATE t SET col = 1", wantTypes: []string{"UPDATE"}},
		{name: "simple DELETE", sql: "DELETE FROM t", wantTypes: []string{"DELETE"}},
		{name: "simple DROP", sql: "DROP TABLE t", wantTypes: []string{"DROP"}},
		{name: "simple TRUNCATE", sql: "TRUNCATE TABLE t", wantTypes: []string{"TRUNCATE"}},
		{name: "simple CREATE", sql: "CREATE TABLE t (id INT)", wantTypes: []string{"CREATE"}},
		{name: "simple ALTER", sql: "ALTER TABLE t ADD col INT", wantTypes: []string{"ALTER"}},
		{name: "simple SHOW", sql: "SHOW TABLES", wantTypes: []string{"SHOW"}},

		{name: "two statements", sql: "SELECT 1; DROP TABLE t", wantTypes: []string{"SELECT", "DROP"}},
		{name: "three statements", sql: "SELECT 1; INSERT INTO t VALUES (1); DELETE FROM t", wantTypes: []string{"SELECT", "INSERT", "DELETE"}},

		{name: "single-line comment", sql: "SELECT 1 -- comment", wantTypes: []string{"SELECT"}},
		{name: "comment before statement", sql: "-- comment\nDROP TABLE t", wantTypes: []string{"DROP"}},
		{name: "multi-line comment", sql: "/* comment */ SELECT 1", wantTypes: []string{"SELECT"}},
		{name: "inline comment", sql: "SELECT /* inline */ 1", wantTypes: []string{"SELECT"}},
		{name: "multi-line comment spanning lines", sql: "/* line1\nline2\nline3 */ SELECT 1", wantTypes: []string{"SELECT"}},

		{name: "lowercase select", sql: "select * from users", wantTypes: []string{"SELECT"}},
		{name: "mixed case DROP", sql: "DrOp TaBlE t", wantTypes: []string{"DROP"}},

		{name: "leading whitespace", sql: "   SELECT 1", wantTypes: []string{"SELECT"}},
		{name: "trailing whitespace", sql: "SELECT 1   ", wantTypes: []string{"SELECT"}},

		{name: "empty string", sql: "", wantTypes: []string{}},
		{name: "only whitespace", sql: "   ", wantTypes: []string{}},
		{name: "only comment", sql: "-- just a comment", wantTypes: []string{}},
		{name: "only semicolons", sql: ";;;", wantTypes: []string{}},

		{name: "DROP inside string literal", sql: "SELECT 'DROP TABLE t'", wantTypes: []string{"SELECT"}},
		{name: "DELETE inside string literal", sql: "SELECT 'DELETE FROM t'", wantTypes: []string{"SELECT"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmts := tokenizer.Tokenize(tt.sql)
			if len(stmts) != len(tt.wantTypes) {
				t.Errorf("Tokenize() returned %d statements, want %d", len(stmts), len(tt.wantTypes))
				return
			}
			for i, stmt := range stmts {
				if stmt.Type != tt.wantTypes[i] {
					t.Errorf("Tokenize()[%d].Type = %q, want %q", i, stmt.Type, tt.wantTypes[i])
				}
			}
		})
	}
}

func TestSQLTokenizer_StripComments(t *testing.T) {
	tokenizer := newSQLTokenizer()

	tests := []struct {
		name string
		sql  string
		want string
	}{
		{name: "no comments", sql: "SELECT 1", want: "SELECT 1"},
		{name: "trailing single-line comment", sql: "SELECT 1 -- comment", want: "SELECT 1 "},
		{name: "full-line single-line comment", sql: "-- comment\nSELECT 1", want: "\nSELECT 1"},
		{name: "multi-line comment", sql: "SELECT /* x */ 1", want: "SELECT  1"},
		{name: "multi-line spanning", sql: "/* line1\nline2 */SELECT 1", want: "SELECT 1"},
		{name: "comment inside string preserved", sql: "SELECT '-- not a comment'", want: "SELECT '-- not a comment'"},
		{name: "block comment inside string preserved", sql: "SELECT '/* not a comment */'", want: "SELECT '/* not a comment */'"},
		{name: "escaped quote in string before comment", sql: "SELECT 'it''s' -- real comment", want: "SELECT 'it''s' "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenizer.stripComments(tt.sql)
			if got != tt.want {
				t.Errorf("stripComments() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSQLTokenizer_HasKeyword(t *testing.T) {
	tokenizer := newSQLTokenizer()

	tests := []struct {
		name    string
		sql     string
		keyword string
		want    bool
	}{
		{name: "keyword present", sql: "DELETE FROM t WHERE id = 1", keyword: "WHERE", want: true},
		{name: "keyword absent", sql: "DELETE FROM t", keyword: "WHERE", want: false},
		{name: "keyword as substring rejected", sql: "SELECT WHEREVER", keyword: "WHERE", want: false},
		{name: "keyword at start", sql: "WHERE clause", keyword: "WHERE", want: true},
		{name: "keyword at end", sql: "DELETE FROM t WHERE", keyword: "WHERE", want: true},
		{name: "case insensitive", sql: "delete from t where id = 1", keyword: "WHERE", want: true},
		{name: "keyword inside identifier rejected", sql: "SELECT SOMEWHERE_COLUMN FROM t", keyword: "WHERE", want: false},
		{name: "underscore prefix rejected", sql: "SELECT _WHERE FROM t", keyword: "WHERE", want: false},
		{name: "underscore suffix rejected", sql: "SELECT WHERE_ FROM t", keyword: "WHERE", want: false},
		{name: "keyword with parens", sql: "WHERE(id = 1)", keyword: "WHERE", want: true},
		{name: "empty keyword", sql: "SELECT 1", keyword: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenizer.hasKeyword(tt.sql, tt.keyword)
			if got != tt.want {
				t.Errorf("hasKeyword(%q, %q) = %v, want %v", tt.sql, tt.keyword, got, tt.want)
			}
		})
	}
}

func TestSQLTokenizer_ExtractFirstKeyword(t *testing.T) {
	tokenizer := newSQLTokenizer()

	tests := []struct {
		name string
		sql  string
		want string
	}{
		{name: "SELECT", sql: "SELECT * FROM t", want: "SELECT"},
		{name: "lowercase select", sql: "select * from t", want: "SELECT"},
		{name: "with leading space", sql: "  DROP TABLE t", want: "DROP"},
		{name: "empty string", sql: "", want: ""},
		{name: "only whitespace", sql: "   ", want: ""},
		{name: "starts with number", sql: "123SELECT", want: "SELECT"},
		{name: "SHOW", sql: "SHOW TABLES", want: "SHOW"},
		{name: "mixed case", sql: "InSeRt INTO t", want: "INSERT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenizer.extractFirstKeyword(tt.sql)
			if got != tt.want {
				t.Errorf("extractFirstKeyword(%q) = %q, want %q", tt.sql, got, tt.want)
			}
		})
	}
}

func TestSQLTokenizer_SplitStatements(t *testing.T) {
	tokenizer := newSQLTokenizer()

	tests := []struct {
		name string
		sql  string
		want int
	}{
		{name: "single statement", sql: "SELECT 1", want: 1},
		{name: "two statements", sql: "SELECT 1; SELECT 2", want: 2},
		{name: "trailing semicolon", sql: "SELECT 1;", want: 1},
		{name: "multiple semicolons", sql: "SELECT 1;; SELECT 2", want: 2},
		{name: "semicolon in string", sql: "SELECT 'a;b'", want: 1},
		{name: "empty", sql: "", want: 0},
		{name: "only semicolons", sql: ";;;", want: 0},
		{name: "escaped quote with semicolon", sql: "SELECT 'it''s; ok'; DROP TABLE t", want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmts := tokenizer.splitStatements(tt.sql)
			if len(stmts) != tt.want {
				t.Errorf("splitStatements() returned %d, want %d (stmts: %v)", len(stmts), tt.want, stmts)
			}
		})
	}
}

func TestSQLTokenizer_StripStringLiterals(t *testing.T) {
	tokenizer := newSQLTokenizer()

	tests := []struct {
		name string
		sql  string
		want string
	}{
		{name: "no strings", sql: "SELECT 1", want: "SELECT 1"},
		{name: "single-quoted string", sql: "SELECT 'hello'", want: "SELECT ''"},
		{name: "empty single-quoted", sql: "SELECT ''", want: "SELECT ''"},
		{name: "escaped quote in string", sql: "SELECT 'it''s'", want: "SELECT ''"},
		{name: "double-quoted identifier removed", sql: `SELECT "col" FROM t`, want: "SELECT  FROM t"},
		{name: "backtick identifier removed", sql: "SELECT `col` FROM t", want: "SELECT  FROM t"},
		{name: "dollar-quoted string removed", sql: "SELECT $$hello$$", want: "SELECT "},
		{name: "DROP inside string stripped", sql: "SELECT 'DROP TABLE t'", want: "SELECT ''"},
		{name: "WHERE inside string stripped", sql: "SELECT 'WHERE clause'", want: "SELECT ''"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenizer.stripStringLiterals(tt.sql)
			if got != tt.want {
				t.Errorf("stripStringLiterals() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSQLTokenizer_SecurityBypass(t *testing.T) {
	tokenizer := newSQLTokenizer()

	tests := []struct {
		name      string
		sql       string
		wantAllow bool
		wantRisk  string
	}{
		{name: "DELETE with WHERE in single-line comment", sql: "DELETE FROM t -- WHERE 1=1", wantAllow: false, wantRisk: "HIGH"},
		{name: "DELETE with WHERE in block comment", sql: "DELETE FROM t /* WHERE 1=1 */", wantAllow: false, wantRisk: "HIGH"},
		{name: "DROP hidden in multi-statement", sql: "SELECT 1; DROP TABLE t", wantAllow: false, wantRisk: "HIGH"},
		{name: "DELETE hidden in multi-statement", sql: "SELECT 1; DELETE FROM t", wantAllow: false, wantRisk: "HIGH"},
		{name: "UPDATE hidden in multi-statement no WHERE", sql: "SELECT 1; UPDATE t SET col = 1", wantAllow: false, wantRisk: "HIGH"},

		{name: "DROP inside string literal allowed", sql: "SELECT 'DROP TABLE t'", wantAllow: true, wantRisk: "LOW"},
		{name: "DELETE with WHERE and string containing where", sql: "DELETE FROM t WHERE name = 'no where here'", wantAllow: true, wantRisk: "LOW"},
		{name: "UPDATE with WHERE", sql: "UPDATE t SET col = 'no where here' WHERE id = 1", wantAllow: true, wantRisk: "LOW"},

		{name: "normal SELECT", sql: "SELECT * FROM users", wantAllow: true, wantRisk: "LOW"},
		{name: "INSERT allowed", sql: "INSERT INTO users (name) VALUES ('test')", wantAllow: true, wantRisk: "LOW"},
		{name: "DELETE with WHERE", sql: "DELETE FROM users WHERE id = 1", wantAllow: true, wantRisk: "LOW"},
		{name: "UPDATE with WHERE", sql: "UPDATE users SET name = 'test' WHERE id = 1", wantAllow: true, wantRisk: "LOW"},

		{name: "DROP TABLE blocked", sql: "DROP TABLE users", wantAllow: false, wantRisk: "HIGH"},
		{name: "TRUNCATE TABLE blocked", sql: "TRUNCATE TABLE users", wantAllow: false, wantRisk: "HIGH"},
		{name: "DELETE without WHERE blocked", sql: "DELETE FROM users", wantAllow: false, wantRisk: "HIGH"},
		{name: "UPDATE without WHERE blocked", sql: "UPDATE users SET name = 'test'", wantAllow: false, wantRisk: "HIGH"},
		{name: "lowercase drop blocked", sql: "drop table users", wantAllow: false, wantRisk: "HIGH"},
		{name: "mixed case Truncate blocked", sql: "TrUnCaTe TABLE users", wantAllow: false, wantRisk: "HIGH"},

		{name: "empty string allowed", sql: "", wantAllow: true, wantRisk: "LOW"},
		{name: "only whitespace allowed", sql: "   ", wantAllow: true, wantRisk: "LOW"},
		{name: "only comment allowed", sql: "-- just a comment", wantAllow: true, wantRisk: "LOW"},

		{name: "comment before DROP", sql: "/* sneaky */ DROP TABLE t", wantAllow: false, wantRisk: "HIGH"},
		{name: "UPDATE with WHERE in comment only", sql: "UPDATE t SET col = 1 -- WHERE id = 1", wantAllow: false, wantRisk: "HIGH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &SQLValidatorImpl{tokenizer: tokenizer}
			result, err := v.Validate(context.Background(), "test", tt.sql)
			if err != nil {
				t.Fatalf("Validate() unexpected error: %v", err)
			}
			if result.Allowed != tt.wantAllow {
				t.Errorf("Validate() Allowed = %v, want %v (sql: %q)", result.Allowed, tt.wantAllow, tt.sql)
			}
			if result.RiskLevel != tt.wantRisk {
				t.Errorf("Validate() RiskLevel = %v, want %v", result.RiskLevel, tt.wantRisk)
			}
		})
	}
}
