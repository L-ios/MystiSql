package validator

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestSQLValidatorImpl_Validate(t *testing.T) {
	validator := NewSQLValidator()

	tests := []struct {
		name      string
		query     string
		wantAllow bool
		wantRisk  string
	}{
		{
			name:      "valid SELECT query",
			query:     "SELECT * FROM users WHERE id = 1",
			wantAllow: true,
			wantRisk:  "LOW",
		},
		{
			name:      "DROP TABLE - dangerous",
			query:     "DROP TABLE users",
			wantAllow: false,
			wantRisk:  "HIGH",
		},
		{
			name:      "TRUNCATE TABLE - dangerous",
			query:     "TRUNCATE TABLE users",
			wantAllow: false,
			wantRisk:  "HIGH",
		},
		{
			name:      "DELETE without WHERE - dangerous",
			query:     "DELETE FROM users",
			wantAllow: false,
			wantRisk:  "HIGH",
		},
		{
			name:      "DELETE with WHERE - allowed",
			query:     "DELETE FROM users WHERE id = 1",
			wantAllow: true,
			wantRisk:  "LOW",
		},
		{
			name:      "UPDATE without WHERE - dangerous",
			query:     "UPDATE users SET name = 'test'",
			wantAllow: false,
			wantRisk:  "HIGH",
		},
		{
			name:      "UPDATE with WHERE - allowed",
			query:     "UPDATE users SET name = 'test' WHERE id = 1",
			wantAllow: true,
			wantRisk:  "LOW",
		},
		{
			name:      "INSERT - allowed",
			query:     "INSERT INTO users (name) VALUES ('test')",
			wantAllow: true,
			wantRisk:  "LOW",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Validate(context.Background(), "test-instance", tt.query)
			if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
				return
			}

			if result.Allowed != tt.wantAllow {
				t.Errorf("Validate() Allowed = %v, want %v", result.Allowed, tt.wantAllow)
			}

			if result.RiskLevel != tt.wantRisk {
				t.Errorf("Validate() RiskLevel = %v, want %v", result.RiskLevel, tt.wantRisk)
			}
		})
	}
}

func TestSQLValidatorImpl_GetQueryType(t *testing.T) {
	validator := NewSQLValidator()

	tests := []struct {
		query    string
		wantType string
	}{
		{"SELECT * FROM users", "SELECT"},
		{"INSERT INTO users VALUES (1)", "INSERT"},
		{"UPDATE users SET name = 'test'", "UPDATE"},
		{"DELETE FROM users", "DELETE"},
		{"CREATE TABLE test (id INT)", "CREATE"},
		{"ALTER TABLE test ADD COLUMN name VARCHAR(100)", "ALTER"},
		{"DROP TABLE test", "DROP"},
		{"TRUNCATE TABLE test", "TRUNCATE"},
		{"SHOW TABLES", "SELECT"},
		{"UNKNOWN QUERY", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			gotType := validator.GetQueryType(tt.query)
			if gotType != tt.wantType {
				t.Errorf("GetQueryType() = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestValidatorService_Validate(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	t.Run("validate with empty lists", func(t *testing.T) {
		result, err := service.Validate(context.Background(), "test-instance", "SELECT * FROM users")
		if err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
			return
		}

		if !result.Allowed {
			t.Errorf("Validate() should allow safe query, got %v", result.Allowed)
		}
	})

	t.Run("validate with blacklist", func(t *testing.T) {
		err := service.AddToBlacklist("^DROP.*")
		if err != nil {
			t.Errorf("AddToBlacklist() error: %v", err)
			return
		}

		result, err := service.Validate(context.Background(), "test-instance", "DROP TABLE users")
		if err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
			return
		}

		if result.Allowed {
			t.Errorf("Validate() should block blacklisted query")
		}
	})

	t.Run("validate with whitelist", func(t *testing.T) {
		err := service.AddToWhitelist("^SELECT.*FROM users.*")
		if err != nil {
			t.Errorf("AddToWhitelist() error: %v", err)
			return
		}

		result, err := service.Validate(context.Background(), "test-instance", "SELECT * FROM users")
		if err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
			return
		}

		if !result.Allowed {
			t.Errorf("Validate() should allow whitelisted query")
		}
	})
}

func TestWhitelistManager(t *testing.T) {
	manager := NewWhitelistManager()

	t.Run("add and match", func(t *testing.T) {
		err := manager.Add("^SELECT.*")
		if err != nil {
			t.Errorf("Add() error: %v", err)
			return
		}

		if !manager.Match("SELECT * FROM users") {
			t.Errorf("Match() should match SELECT query")
		}

		if manager.Match("DROP TABLE users") {
			t.Errorf("Match() should not match DROP query")
		}
	})

	t.Run("remove", func(t *testing.T) {
		pattern := "^DELETE.*"
		err := manager.Add(pattern)
		if err != nil {
			t.Errorf("Add() error: %v", err)
			return
		}

		err = manager.Remove(pattern)
		if err != nil {
			t.Errorf("Remove() error: %v", err)
			return
		}

		if manager.Match("DELETE FROM users") {
			t.Errorf("Match() should not match after removal")
		}
	})

	t.Run("get all", func(t *testing.T) {
		manager := NewWhitelistManager()
		manager.Add("^SELECT.*")
		manager.Add("^INSERT.*")

		patterns := manager.GetAll()
		if len(patterns) != 2 {
			t.Errorf("GetAll() should return 2 patterns, got %d", len(patterns))
		}
	})
}

func TestBlacklistManager(t *testing.T) {
	manager := NewBlacklistManager()

	t.Run("add and match", func(t *testing.T) {
		err := manager.Add("^DROP.*")
		if err != nil {
			t.Errorf("Add() error: %v", err)
			return
		}

		if !manager.Match("DROP TABLE users") {
			t.Errorf("Match() should match DROP query")
		}

		if manager.Match("SELECT * FROM users") {
			t.Errorf("Match() should not match SELECT query")
		}
	})

	t.Run("priority over whitelist", func(t *testing.T) {
		logger := zap.NewNop()
		service := NewValidatorService(logger)

		// Add to both whitelist and blacklist
		service.AddToWhitelist("^DROP.*")
		service.AddToBlacklist("^DROP.*")

		// Blacklist should take priority
		result, err := service.Validate(context.Background(), "test-instance", "DROP TABLE users")
		if err != nil {
			t.Errorf("Validate() error: %v", err)
			return
		}

		if result.Allowed {
			t.Errorf("Blacklist should take priority over whitelist")
		}
	})
}
