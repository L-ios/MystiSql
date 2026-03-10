package cli

import (
	"strings"
	"testing"
)

func TestHighlightSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "SELECT statement",
			input:    "SELECT * FROM users",
			contains: []string{"SELECT", "FROM"},
		},
		{
			name:     "INSERT statement",
			input:    "INSERT INTO users VALUES (1)",
			contains: []string{"INSERT", "INTO", "VALUES"},
		},
		{
			name:     "string literal",
			input:    "SELECT * FROM users WHERE name = 'test'",
			contains: []string{"SELECT", "FROM", "WHERE", "'test'"},
		},
		{
			name:     "mixed case keywords",
			input:    "select * from Users",
			contains: []string{"select", "from"},
		},
		{
			name:     "empty string",
			input:    "",
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := highlightSQL(tt.input)
			if tt.input == "" {
				if result != "" {
					t.Error("highlightSQL('') should return ''")
				}
				return
			}

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("highlightSQL('%s') should contain '%s', got '%s'", tt.input, expected, result)
				}
			}
		})
	}
}

func TestHighlightSQLKeywords(t *testing.T) {
	keywords := []string{
		"SELECT", "FROM", "WHERE", "AND", "OR", "NOT",
		"INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE",
		"CREATE", "TABLE", "DROP", "ALTER",
		"JOIN", "LEFT", "RIGHT", "INNER", "OUTER", "ON",
		"GROUP", "BY", "ORDER", "ASC", "DESC", "LIMIT",
	}

	for _, keyword := range keywords {
		t.Run(keyword, func(t *testing.T) {
			input := keyword + " test"
			result := highlightSQL(input)
			if !strings.Contains(result, keyword) {
				t.Errorf("highlightSQL should preserve keyword '%s'", keyword)
			}
		})
	}
}
