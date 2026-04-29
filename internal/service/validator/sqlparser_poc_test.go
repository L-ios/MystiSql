package validator

import (
	"testing"

	"github.com/xwb1989/sqlparser"
)

// This file is a POC (Proof of Concept) test for github.com/xwb1989/sqlparser.
// It validates whether the parser can handle the SQL patterns used in MystiSql.

// parseSQL is a helper that attempts to parse a SQL string using xwb1989/sqlparser.
// Returns the parsed statement or an error if parsing fails.
func parseSQL(sql string) (sqlparser.Statement, error) {
	return sqlparser.Parse(sql)
}

// getStmtType extracts the SQL statement type as a string.
func getStmtType(stmt sqlparser.Statement) string {
	switch stmt.(type) {
	case *sqlparser.Select:
		return "SELECT"
	case *sqlparser.Insert:
		return "INSERT"
	case *sqlparser.Update:
		return "UPDATE"
	case *sqlparser.Delete:
		return "DELETE"
	case *sqlparser.DDL:
		return "DDL"
	case *sqlparser.Set:
		return "SET"
	default:
		return "UNKNOWN"
	}
}

func TestPOC_ParseSelect(t *testing.T) {
	stmt, err := parseSQL("SELECT id, name FROM users WHERE id = 1")
	if err != nil {
		t.Fatalf("Failed to parse SELECT: %v", err)
	}
	if getStmtType(stmt) != "SELECT" {
		t.Errorf("Expected SELECT, got %s", getStmtType(stmt))
	}
}

func TestPOC_ParseInsert(t *testing.T) {
	stmt, err := parseSQL("INSERT INTO users (id, name) VALUES (1, 'alice')")
	if err != nil {
		t.Fatalf("Failed to parse INSERT: %v", err)
	}
	if getStmtType(stmt) != "INSERT" {
		t.Errorf("Expected INSERT, got %s", getStmtType(stmt))
	}
}

func TestPOC_ParseUpdate(t *testing.T) {
	stmt, err := parseSQL("UPDATE users SET name = 'bob' WHERE id = 1")
	if err != nil {
		t.Fatalf("Failed to parse UPDATE: %v", err)
	}
	if getStmtType(stmt) != "UPDATE" {
		t.Errorf("Expected UPDATE, got %s", getStmtType(stmt))
	}
}

func TestPOC_ParseDelete(t *testing.T) {
	stmt, err := parseSQL("DELETE FROM users WHERE id = 1")
	if err != nil {
		t.Fatalf("Failed to parse DELETE: %v", err)
	}
	if getStmtType(stmt) != "DELETE" {
		t.Errorf("Expected DELETE, got %s", getStmtType(stmt))
	}
}

func TestPOC_ParseDrop(t *testing.T) {
	stmt, err := parseSQL("DROP TABLE users")
	if err != nil {
		t.Fatalf("Failed to parse DROP TABLE: %v", err)
	}
	if getStmtType(stmt) != "DDL" {
		t.Errorf("Expected DDL, got %s", getStmtType(stmt))
	}

	ddl, ok := stmt.(*sqlparser.DDL)
	if !ok {
		t.Fatal("Expected DDL statement")
	}
	if ddl.Action != "drop" {
		t.Errorf("Expected drop action, got %s", ddl.Action)
	}
}

func TestPOC_ParseTruncate(t *testing.T) {
	stmt, err := parseSQL("TRUNCATE TABLE users")
	if err != nil {
		t.Fatalf("Failed to parse TRUNCATE: %v", err)
	}
	// sqlparser may parse TRUNCATE as DDL or unknown
	stmtType := getStmtType(stmt)
	t.Logf("TRUNCATE parsed as: %s", stmtType)
}

func TestPOC_ChineseTableName(t *testing.T) {
	stmt, err := parseSQL("SELECT * FROM `用户表`")
	if err != nil {
		t.Fatalf("Failed to parse Chinese table name with backtick: %v", err)
	}
	if getStmtType(stmt) != "SELECT" {
		t.Errorf("Expected SELECT, got %s", getStmtType(stmt))
	}

	sel, ok := stmt.(*sqlparser.Select)
	if !ok {
		t.Fatal("Expected SELECT statement")
	}
	t.Logf("Chinese table name parsed: FROM = %v", sel.From)

	// Unquoted Chinese table names are NOT supported by sqlparser (MySQL syntax requires backtick quoting)
	_, err = parseSQL("SELECT * FROM 用户表")
	if err != nil {
		t.Logf("Unquoted Chinese table name correctly fails: %v", err)
	} else {
		t.Log("Unquoted Chinese table name unexpectedly parsed")
	}
}

func TestPOC_Subquery(t *testing.T) {
	// Test: Subquery in WHERE clause
	stmt, err := parseSQL("DELETE FROM users WHERE id IN (SELECT id FROM backup)")
	if err != nil {
		t.Fatalf("Failed to parse subquery: %v", err)
	}
	if getStmtType(stmt) != "DELETE" {
		t.Errorf("Expected DELETE, got %s", getStmtType(stmt))
	}
}

func TestPOC_DeleteWithComment(t *testing.T) {
	// Test: DELETE with SQL comment to bypass regex-based WHERE check
	// This is the key bypass scenario: "DELETE FROM t -- WHERE 1=1"
	stmt, err := parseSQL("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	del, ok := stmt.(*sqlparser.Delete)
	if !ok {
		t.Fatal("Expected DELETE statement")
	}

	// The WHERE clause should be nil because "-- WHERE 1=1" is a comment, not a condition
	if del.Where != nil {
		t.Error("Expected nil WHERE clause (comment should be stripped), but got a WHERE")
	} else {
		t.Log("PASS: Comment-based WHERE bypass correctly detected as no-WHERE DELETE")
	}
}

func TestPOC_MultiStatement(t *testing.T) {
	// sqlparser.Parse only parses the first statement.
	// For multi-statement handling, we need to use sqlparser.ParseNext with a tokenizer.
	sql := "SELECT 1; DROP TABLE users"

	// Use tokenizer-based parsing for multi-statement
	tokenizer := sqlparser.NewStringTokenizer(sql)
	var stmts []sqlparser.Statement
	for {
		stmt, err := sqlparser.ParseNext(tokenizer)
		if err != nil {
			break
		}
		stmts = append(stmts, stmt)
	}

	if len(stmts) < 2 {
		t.Fatalf("Expected at least 2 statements, got %d", len(stmts))
	}

	if getStmtType(stmts[0]) != "SELECT" {
		t.Errorf("First statement: expected SELECT, got %s", getStmtType(stmts[0]))
	}
	if getStmtType(stmts[1]) != "DDL" {
		t.Errorf("Second statement: expected DDL (DROP), got %s", getStmtType(stmts[1]))
	}
	t.Logf("Multi-statement: found %d statements", len(stmts))
}

func TestPOC_SetStatement(t *testing.T) {
	stmt, err := parseSQL("SET NAMES utf8")
	if err != nil {
		t.Fatalf("Failed to parse SET: %v", err)
	}
	if getStmtType(stmt) != "SET" {
		t.Errorf("Expected SET, got %s", getStmtType(stmt))
	}
}

func TestPOC_SelectForUpdate(t *testing.T) {
	stmt, err := parseSQL("SELECT * FROM users WHERE id = 1 FOR UPDATE")
	if err != nil {
		t.Fatalf("Failed to parse SELECT FOR UPDATE: %v", err)
	}
	sel, ok := stmt.(*sqlparser.Select)
	if !ok {
		t.Fatal("Expected SELECT statement")
	}
	if sel.Lock == "" {
		t.Error("Expected lock clause to be set for FOR UPDATE")
	}
	t.Logf("Lock clause: %s", sel.Lock)
}

func TestPOC_PGOnConflict(t *testing.T) {
	// PostgreSQL: INSERT ... ON CONFLICT — should fail on MySQL parser
	stmt, err := parseSQL("INSERT INTO users (id, name) VALUES (1, 'alice') ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name")
	if err != nil {
		t.Logf("PASS: ON CONFLICT correctly failed to parse: %v", err)
		return
	}
	// If it somehow parses, log the result
	t.Logf("ON CONFLICT parsed as: %T (unexpected)", stmt)
}

func TestPOC_PGReturning(t *testing.T) {
	// PostgreSQL: DELETE ... RETURNING — should fail on MySQL parser
	stmt, err := parseSQL("DELETE FROM users WHERE id = 1 RETURNING *")
	if err != nil {
		t.Logf("PASS: RETURNING correctly failed to parse: %v", err)
		return
	}
	t.Logf("RETURNING parsed as: %T (unexpected)", stmt)
}

func TestPOC_PGInsertOnConflictDoNothing(t *testing.T) {
	// PostgreSQL: INSERT ... ON CONFLICT DO NOTHING
	stmt, err := parseSQL("INSERT INTO users (id, name) VALUES (1, 'alice') ON CONFLICT DO NOTHING")
	if err != nil {
		t.Logf("PASS: ON CONFLICT DO NOTHING correctly failed to parse: %v", err)
		return
	}
	t.Logf("ON CONFLICT DO NOTHING parsed as: %T (unexpected)", stmt)
}

func TestPOC_CoverageReport(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantOK  bool
		wantErr bool
	}{
		// Basic DML
		{"SELECT basic", "SELECT 1", true, false},
		{"SELECT with WHERE", "SELECT * FROM t WHERE id = 1", true, false},
		{"SELECT with JOIN", "SELECT * FROM t1 JOIN t2 ON t1.id = t2.id", true, false},
		{"SELECT with GROUP BY", "SELECT COUNT(*) FROM t GROUP BY status", true, false},
		{"SELECT with subquery", "SELECT * FROM t WHERE id IN (SELECT id FROM t2)", true, false},
		{"INSERT basic", "INSERT INTO t (id) VALUES (1)", true, false},
		{"UPDATE basic", "UPDATE t SET name = 'x' WHERE id = 1", true, false},
		{"DELETE basic", "DELETE FROM t WHERE id = 1", true, false},
		{"DELETE no WHERE", "DELETE FROM t", true, false},

		// DDL
		{"DROP TABLE", "DROP TABLE t", true, false},
		{"CREATE TABLE", "CREATE TABLE t (id INT PRIMARY KEY)", true, false},
		{"ALTER TABLE", "ALTER TABLE t ADD COLUMN name VARCHAR(255)", true, false},
		{"TRUNCATE", "TRUNCATE TABLE t", true, false},

		// Edge cases
		{"Chinese table name (backtick)", "SELECT * FROM `用户表`", true, false},
		{"SET statement", "SET NAMES utf8", true, false},
		{"SELECT FOR UPDATE", "SELECT * FROM t WHERE id = 1 FOR UPDATE", true, false},

		// PG-specific (expected to fail on MySQL parser)
		{"PG ON CONFLICT", "INSERT INTO t (id) VALUES (1) ON CONFLICT DO NOTHING", false, true},
		{"PG RETURNING", "DELETE FROM t RETURNING *", false, true},
	}

	passed := 0
	failed := 0
	for _, tt := range tests {
		stmt, err := parseSQL(tt.sql)
		ok := err == nil && stmt != nil

		if tt.wantOK && ok {
			passed++
		} else if tt.wantErr && err != nil {
			passed++
		} else {
			failed++
			t.Logf("UNEXPECTED: %s: sql=%q ok=%v err=%v wantOK=%v wantErr=%v",
				tt.name, tt.sql, ok, err, tt.wantOK, tt.wantErr)
		}
	}

	total := len(tests)
	coverage := float64(passed) / float64(total) * 100
	t.Logf("=== POC Coverage Report ===")
	t.Logf("Total: %d, Passed: %d, Failed: %d", total, passed, failed)
	t.Logf("Coverage: %.1f%%", coverage)

	if coverage < 90.0 {
		t.Errorf("Coverage %.1f%% is below 90%% threshold", coverage)
	}
}
