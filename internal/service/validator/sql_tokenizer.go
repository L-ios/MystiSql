package validator

import (
	"strings"
	"unicode"
)

// sqlTokenizer provides lightweight SQL tokenization for validation.
// It strips comments and string literals to identify the true statement type,
// preventing bypass via SQL comments, multi-statement injection, or string tricks.
type sqlTokenizer struct{}

// Statement represents a parsed SQL statement with its type and cleaned content.
type Statement struct {
	Type    string // Upper-case first keyword: SELECT, INSERT, UPDATE, DELETE, etc.
	Content string // SQL with comments and string literals removed
}

// newSQLTokenizer creates a new tokenizer instance.
func newSQLTokenizer() *sqlTokenizer {
	return &sqlTokenizer{}
}

// Tokenize splits SQL into individual statements and identifies each statement's type.
// It strips comments, splits by semicolons, removes string literal content,
// and extracts the first keyword as the statement type.
func (t *sqlTokenizer) Tokenize(sql string) []Statement {
	cleaned := t.stripComments(sql)
	rawStatements := t.splitStatements(cleaned)

	statements := make([]Statement, 0, len(rawStatements))
	for _, raw := range rawStatements {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		stripped := t.stripStringLiterals(raw)
		stmtType := t.extractFirstKeyword(stripped)

		statements = append(statements, Statement{
			Type:    stmtType,
			Content: stripped,
		})
	}

	return statements
}

// stripComments removes SQL comments from the input while preserving string literals.
// Handles:
//   - Single-line comments: -- comment (until newline)
//   - Multi-line comments: /* comment */ (can span lines)
func (t *sqlTokenizer) stripComments(sql string) string {
	var result strings.Builder
	result.Grow(len(sql))
	i := 0
	n := len(sql)

	for i < n {
		// Single-line comment: --
		if i+1 < n && sql[i] == '-' && sql[i+1] == '-' {
			for i < n && sql[i] != '\n' {
				i++
			}
			continue
		}

		// Multi-line comment: /* ... */
		if i+1 < n && sql[i] == '/' && sql[i+1] == '*' {
			i += 2
			for i+1 < n && !(sql[i] == '*' && sql[i+1] == '/') {
				i++
			}
			if i+1 < n {
				i += 2 // skip */
			} else {
				i = n // unterminated comment — skip to end
			}
			continue
		}

		// String literal: preserve as-is (don't strip "comments" inside strings)
		if sql[i] == '\'' {
			result.WriteByte(sql[i])
			i++
			for i < n {
				result.WriteByte(sql[i])
				if sql[i] == '\'' {
					i++
					if i < n && sql[i] == '\'' {
						// Escaped quote ''
						result.WriteByte(sql[i])
						i++
						continue
					}
					break
				}
				i++
			}
			continue
		}

		result.WriteByte(sql[i])
		i++
	}

	return result.String()
}

// stripStringLiterals removes string content from the SQL so keyword detection
// works on the SQL structure, not the data. Replaces single-quoted string
// content with an empty string literal, and removes double-quoted/backtick/
// dollar-quoted content entirely.
func (t *sqlTokenizer) stripStringLiterals(sql string) string {
	var result strings.Builder
	result.Grow(len(sql))
	i := 0
	n := len(sql)

	for i < n {
		switch sql[i] {
		case '\'':
			// Replace string content with empty string literal ''
			result.WriteString("''")
			i++
			for i < n {
				if sql[i] == '\'' {
					i++
					if i < n && sql[i] == '\'' {
						// Escaped quote ''
						i++
						continue
					}
					break
				}
				i++
			}

		case '"':
			// Double-quoted identifier: skip content
			i++
			for i < n && sql[i] != '"' {
				i++
			}
			if i < n {
				i++
			}

		case '`':
			// Backtick identifier: skip content
			i++
			for i < n && sql[i] != '`' {
				i++
			}
			if i < n {
				i++
			}

		case '$':
			// Dollar-quoted string (PostgreSQL): $$..$$
			if i+1 < n && sql[i+1] == '$' {
				i += 2
				for i+1 < n && !(sql[i] == '$' && sql[i+1] == '$') {
					i++
				}
				if i+1 < n {
					i += 2
				}
			} else {
				result.WriteByte(sql[i])
				i++
			}

		default:
			result.WriteByte(sql[i])
			i++
		}
	}

	return result.String()
}

// splitStatements splits SQL by semicolons, respecting single-quoted string boundaries.
func (t *sqlTokenizer) splitStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	i := 0
	n := len(sql)

	for i < n {
		if sql[i] == ';' {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
			i++
			continue
		}

		if sql[i] == '\'' {
			current.WriteByte(sql[i])
			i++
			for i < n {
				current.WriteByte(sql[i])
				if sql[i] == '\'' {
					i++
					if i < n && sql[i] == '\'' {
						current.WriteByte(sql[i])
						i++
						continue
					}
					break
				}
				i++
			}
			continue
		}

		current.WriteByte(sql[i])
		i++
	}

	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

// extractFirstKeyword returns the uppercase first keyword of a SQL statement.
// Skips leading whitespace and digits, then collects consecutive letters/underscores.
func (t *sqlTokenizer) extractFirstKeyword(sql string) string {
	sql = strings.TrimSpace(sql)
	var keyword strings.Builder
	for _, r := range sql {
		if unicode.IsLetter(r) || r == '_' {
			keyword.WriteRune(unicode.ToUpper(r))
		} else if keyword.Len() > 0 {
			break
		}
		// Skip leading whitespace, digits, and other non-letter chars
	}
	return keyword.String()
}

// hasKeyword checks if a keyword exists in the SQL (case-insensitive, word boundary).
// A match is valid only when surrounded by non-alphanumeric characters (or string boundaries).
func (t *sqlTokenizer) hasKeyword(sql, keyword string) bool {
	upper := strings.ToUpper(sql)
	target := strings.ToUpper(keyword)
	targetLen := len(target)

	if targetLen == 0 {
		return false
	}

	idx := 0
	for idx <= len(upper)-targetLen {
		pos := strings.Index(upper[idx:], target)
		if pos == -1 {
			return false
		}
		pos += idx

		beforeOK := pos == 0 || !isAlphaNum(rune(upper[pos-1]))
		afterOK := pos+targetLen >= len(upper) || !isAlphaNum(rune(upper[pos+targetLen]))

		if beforeOK && afterOK {
			return true
		}
		idx = pos + 1
	}
	return false
}

// isAlphaNum reports whether the rune is a letter, digit, or underscore.
func isAlphaNum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
