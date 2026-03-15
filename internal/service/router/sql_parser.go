// SQL parser for read-write routing decisions in MystiSql.
// Provides a lightweight analysis to identify SQL type and transaction context.
package router

import (
	"regexp"
	"strings"
)

// SQLType represents the type of an SQL statement.
// Exported to allow downstream components (like the read-write router)
// to branch logic based on the statement type.
type SQLType int

const (
	// SQLTypeUnknown indicates the statement type could not be determined
	// or is not one of the supported DML/DDL types.
	SQLTypeUnknown SQLType = iota
	// SQLTypeSelect represents a SELECT statement.
	SQLTypeSelect
	// SQLTypeInsert represents an INSERT statement.
	SQLTypeInsert
	// SQLTypeUpdate represents an UPDATE statement.
	SQLTypeUpdate
	// SQLTypeDelete represents a DELETE statement.
	SQLTypeDelete
)

// String returns a human-friendly representation of the SQLType.
func (t SQLType) String() string {
	switch t {
	case SQLTypeSelect:
		return "SELECT"
	case SQLTypeInsert:
		return "INSERT"
	case SQLTypeUpdate:
		return "UPDATE"
	case SQLTypeDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// ParseSQL analyzes the provided SQL string and returns:
// - the detected SQL type (SELECT/INSERT/UPDATE/DELETE/UNKNOWN)
// - a boolean indicating whether the statement is inside or starts a transaction
// - a boolean indicating whether the statement itself is a transaction boundary
//
// The returned booleans are heuristics suitable for routing decisions in the
// read-write router. They are not a full SQL parser, but cover the common cases
// used in MystiSql routing.
func ParseSQL(sql string) (SQLType, bool, bool) {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	var (
		typ       SQLType
		isTxnStmt bool
		inTxn     bool
	)

	// Determine the SQL type by inspecting the first keyword
	fields := strings.Fields(upper)
	if len(fields) > 0 {
		switch fields[0] {
		case "SELECT":
			typ = SQLTypeSelect
		case "INSERT":
			typ = SQLTypeInsert
		case "UPDATE":
			typ = SQLTypeUpdate
		case "DELETE":
			typ = SQLTypeDelete
		default:
			typ = SQLTypeUnknown
		}

		// Determine if the statement is a transaction boundary statement
		if fields[0] == "BEGIN" || fields[0] == "START" || fields[0] == "COMMIT" || fields[0] == "ROLLBACK" {
			isTxnStmt = true
		}
	}

	// Heuristic: consider inside-transaction if a boundary statement is present
	// or if the text contains a BEGIN marker anywhere.
	if isTxnStmt {
		inTxn = true
	} else {
		re := regexp.MustCompile(`(?i)\bBEGIN\b`)
		if re.MatchString(sql) {
			inTxn = true
		}
	}

	return typ, inTxn, isTxnStmt
}

// IsTransaction reports whether the provided SQL string represents a
// transaction boundary or a transaction-related operation.
// It is a lightweight helper that uses a case-insensitive match on common
// transaction keywords.
func IsTransaction(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	// Quick check on leading keyword
	fields := strings.Fields(upper)
	if len(fields) > 0 {
		switch fields[0] {
		case "BEGIN", "START", "COMMIT", "ROLLBACK":
			return true
		}
	}
	re := regexp.MustCompile(`(?i)\bBEGIN\b`)
	return re.MatchString(sql)
}
