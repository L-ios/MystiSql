package validator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLValidator_SQLInjectionPrevention(t *testing.T) {
	validator := NewSQLValidator(DefaultDangerousOperations())

	sqlInjectionAttempts := []string{
		"SELECT * FROM users WHERE id = 1; DROP TABLE users; --",
		"SELECT * FROM users; TRUNCATE table audit_log",
		"INSERT INTO users (name) VALUES ('test'); DELETE FROM users WHERE 1=1; --",
		"SELECT * FROM users WHERE id = 1 OR 1=1; DROP DATABASE testdb; --",
		"/* comment */ DROP TABLE users",
		"-- comment\nDROP TABLE users",
		"SELECT * FROM (SELECT * FROM users); DROP TABLE users; --",
	}

	for _, sql := range sqlInjectionAttempts {
		err := validator.Validate(context.Background(), sql)
		assert.Error(t, err, "SQL injection attempt should be blocked: %s", sql)
	}
}

func TestSQLValidator_DangerousOperations(t *testing.T) {
	validator := NewSQLValidator(DefaultDangerousOperations)

	tests := []struct {
		name       string
		sql        string
		shouldFail bool
	}{
		{"DROP TABLE", "DROP TABLE users", true},
		{"DROP DATABASE", "DROP DATABASE testdb", true},
		{"TRUNCATE", "TRUNCATE TABLE users", true},
		{"DELETE without WHERE", "DELETE FROM users", true},
		{"UPDATE without WHERE", "UPDATE users SET name = 'test'", true},
		{"SELECT with DROP in comment", "SELECT * FROM users /* DROP TABLE users */", false},
		{"Normal SELECT", "SELECT * FROM users WHERE id = 1", false},
		{"Normal INSERT", "INSERT INTO users (name) VALUES ('test')", false},
		{"Normal UPDATE", "UPDATE users SET name = 'test' WHERE id = 1", false},
		{"Normal DELETE", "DELETE FROM users WHERE id = 1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(context.Background(), tt.sql)
			if tt.shouldFail {
				assert.Error(t, err, "Expected dangerous operation to be blocked")
			} else {
				assert.NoError(t, err, "Expected safe operation to pass")
			}
		})
	}
}

func TestSQLValidator_WhitelistBypass(t *testing.T) {
	validator := NewSQLValidator(DefaultDangerousOperations)

	validator.AddToWhitelist("DROP TABLE allowed_table")
	validator.AddToWhitelist("TRUNCATE TABLE audit_log")

	err := validator.Validate(context.Background(), "DROP TABLE allowed_table")
	assert.NoError(t, err, "Whitelisted DROP should be allowed")

	err = validator.Validate(context.Background(), "TRUNCATE TABLE audit_log")
	assert.NoError(t, err, "Whitelisted TRUNCATE should be allowed")

	err = validator.Validate(context.Background(), "DROP TABLE other_table")
	assert.Error(t, err, "Non-whitelisted DROP should be blocked")
}

func TestSQLValidator_BllistBlock(t *testing.T) {
	validator := NewSQLValidator(DefaultDangerousOperations)

	validator.AddToBlacklist("SELECT * FROM sensitive_data")
	validator.AddToBlacklist("INSERT INTO audit_log")

	err := validator.Validate(context.Background(), "SELECT * FROM sensitive_data")
	assert.Error(t, err, "Blacklisted SELECT should be blocked")

	err = validator.Validate(context.Background(), "INSERT INTO audit_log VALUES (1, 'test')")
	assert.Error(t, err, "Blacklisted INSERT should be blocked")

	err = validator.Validate(context.Background(), "SELECT * FROM other_table")
	assert.NoError(t, err, "Non-blacklisted SELECT should be allowed")
}
