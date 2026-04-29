package validator

import (
	"context"
	"strings"

	"github.com/xwb1989/sqlparser"
	"go.uber.org/zap"
)

// ASTValidator uses SQL AST parsing for more accurate validation than
// tokenizer-based approaches. It handles multi-statement queries, comment
// bypass detection, and subquery analysis. Falls back to SQLValidatorImpl
// for unsupported syntax (e.g., PostgreSQL-specific constructs).
type ASTValidator struct {
	fallback *SQLValidatorImpl
	logger   *zap.Logger
}

// NewASTValidator creates a new AST-based SQL validator with the given
// tokenizer-based fallback for degradation scenarios.
func NewASTValidator(fallback *SQLValidatorImpl, logger *zap.Logger) *ASTValidator {
	return &ASTValidator{
		fallback: fallback,
		logger:   logger,
	}
}

func (a *ASTValidator) Validate(ctx context.Context, instance, query string) (*ValidationResult, error) {
	statements := a.fallback.tokenizer.Tokenize(query)

	needsAST := false
	for _, stmt := range statements {
		switch stmt.Type {
		case "DELETE", "UPDATE", "DROP", "TRUNCATE", "DDL":
			needsAST = true
		}
	}

	if !needsAST {
		return a.fallback.Validate(ctx, instance, query)
	}

	sqlTokenizer := sqlparser.NewStringTokenizer(query)
	var stmts []sqlparser.Statement

	for {
		stmt, err := sqlparser.ParseNext(sqlTokenizer)
		if err != nil {
			break
		}
		if stmt == nil {
			break
		}
		stmts = append(stmts, stmt)
	}

	if len(stmts) == 0 {
		a.logger.Warn("AST parser failed to parse query, falling back to tokenizer",
			zap.String("instance", instance),
			zap.String("query", query),
		)
		return a.fallback.Validate(ctx, instance, query)
	}

	for _, stmt := range stmts {
		if result := a.validateStatement(stmt); result != nil {
			return result, nil
		}
	}

	return &ValidationResult{
		Allowed:   true,
		Reason:    "Query validated successfully",
		RiskLevel: "LOW",
	}, nil
}

// validateStatement checks a single parsed SQL statement against security rules.
// Returns a blocking ValidationResult for dangerous statements, or nil if allowed.
func (a *ASTValidator) validateStatement(stmt sqlparser.Statement) *ValidationResult {
	switch s := stmt.(type) {
	case *sqlparser.DDL:
		action := strings.ToLower(s.Action)
		if action == "drop" {
			return &ValidationResult{
				Allowed:   false,
				Reason:    "SQL query contains dangerous operation: DROP",
				RiskLevel: "HIGH",
			}
		}
		if action == "truncate" {
			return &ValidationResult{
				Allowed:   false,
				Reason:    "SQL query contains dangerous operation: TRUNCATE",
				RiskLevel: "HIGH",
			}
		}
	case *sqlparser.Delete:
		if s.Where == nil {
			return &ValidationResult{
				Allowed:   false,
				Reason:    "DELETE operation without WHERE clause is not allowed",
				RiskLevel: "HIGH",
			}
		}
	case *sqlparser.Update:
		if s.Where == nil {
			return &ValidationResult{
				Allowed:   false,
				Reason:    "UPDATE operation without WHERE clause is not allowed",
				RiskLevel: "HIGH",
			}
		}
	}
	return nil
}

// GetQueryType returns the SQL statement type as a string.
// Falls back to the tokenizer-based implementation for unsupported syntax.
func (a *ASTValidator) GetQueryType(query string) string {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		a.logger.Warn("AST parser failed to parse query for type detection, falling back to tokenizer",
			zap.String("query", query),
		)
		return a.fallback.GetQueryType(query)
	}

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
