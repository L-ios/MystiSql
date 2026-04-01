package validator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestValidatorService_RemoveFromWhitelist(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	err := service.AddToWhitelist("^DELETE.*")
	if err != nil {
		t.Fatalf("AddToWhitelist() error: %v", err)
	}

	err = service.RemoveFromWhitelist("^DELETE.*")
	if err != nil {
		t.Fatalf("RemoveFromWhitelist() error: %v", err)
	}

	patterns := service.GetWhitelist()
	if len(patterns) != 0 {
		t.Errorf("GetWhitelist() after removal = %v, want empty", patterns)
	}
}

func TestValidatorService_RemoveFromBlacklist(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	err := service.AddToBlacklist("^DROP.*")
	if err != nil {
		t.Fatalf("AddToBlacklist() error: %v", err)
	}

	err = service.RemoveFromBlacklist("^DROP.*")
	if err != nil {
		t.Fatalf("RemoveFromBlacklist() error: %v", err)
	}

	patterns := service.GetBlacklist()
	if len(patterns) != 0 {
		t.Errorf("GetBlacklist() after removal = %v, want empty", patterns)
	}
}

func TestValidatorService_GetWhitelist(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	service.AddToWhitelist("^SELECT.*")
	service.AddToWhitelist("^INSERT.*")

	patterns := service.GetWhitelist()
	if len(patterns) != 2 {
		t.Errorf("GetWhitelist() count = %d, want 2", len(patterns))
	}
}

func TestValidatorService_GetBlacklist(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	service.AddToBlacklist("^DROP.*")
	service.AddToBlacklist("^TRUNCATE.*")

	patterns := service.GetBlacklist()
	if len(patterns) != 2 {
		t.Errorf("GetBlacklist() count = %d, want 2", len(patterns))
	}
}

func TestValidatorService_UpdateWhitelist(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	service.AddToWhitelist("^SELECT.*")

	err := service.UpdateWhitelist([]string{"^INSERT.*", "^UPDATE.*"})
	if err != nil {
		t.Fatalf("UpdateWhitelist() error: %v", err)
	}

	patterns := service.GetWhitelist()
	if len(patterns) != 2 {
		t.Errorf("GetWhitelist() count = %d, want 2", len(patterns))
	}
}

func TestValidatorService_UpdateBlacklist(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	service.AddToBlacklist("^DROP.*")

	err := service.UpdateBlacklist([]string{"^TRUNCATE.*", "^DELETE.*"})
	if err != nil {
		t.Fatalf("UpdateBlacklist() error: %v", err)
	}

	patterns := service.GetBlacklist()
	if len(patterns) != 2 {
		t.Errorf("GetBlacklist() count = %d, want 2", len(patterns))
	}
}

func TestValidatorService_UpdateWhitelist_InvalidPattern(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	err := service.UpdateWhitelist([]string{"[invalid("})
	if err == nil {
		t.Error("UpdateWhitelist() should return error for invalid pattern")
	}
}

func TestValidatorService_UpdateBlacklist_InvalidPattern(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	err := service.UpdateBlacklist([]string{"[invalid("})
	if err == nil {
		t.Error("UpdateBlacklist() should return error for invalid pattern")
	}
}

func TestTruncateQuery(t *testing.T) {
	shortQuery := "SELECT * FROM users WHERE id = 1"
	result := truncateQuery(shortQuery)
	if result != shortQuery {
		t.Errorf("truncateQuery() short = %v, want %v", result, shortQuery)
	}

	longQuery := "SELECT * FROM users WHERE id = 1 AND name = 'test' AND status = 'active' AND created_at > '2024-01-01'"
	if len(longQuery) <= 100 {
		t.Fatal("Test requires long query > 100 chars")
	}
	result = truncateQuery(longQuery)
	if len(result) != 103 {
		t.Errorf("truncateQuery() long length = %d, want 103", len(result))
	}
	if result[len(result)-3:] != "..." {
		t.Errorf("truncateQuery() should end with ...")
	}
}

func TestValidatorService_Validate_BlacklistBlocksAllowedQuery(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	err := service.AddToBlacklist("^DELETE.*")
	if err != nil {
		t.Fatalf("AddToBlacklist() error: %v", err)
	}

	result, err := service.Validate(context.Background(), "test-instance", "DELETE FROM users WHERE id = 1")
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	if result.Allowed {
		t.Error("Validate() should block DELETE with WHERE when blacklisted")
	}
}

func TestValidatorService_Validate_WhitelistAllowsBlockedQuery(t *testing.T) {
	logger := zap.NewNop()
	service := NewValidatorService(logger)

	err := service.AddToWhitelist("^DELETE.*")
	if err != nil {
		t.Fatalf("AddToWhitelist() error: %v", err)
	}

	result, err := service.Validate(context.Background(), "test-instance", "DELETE FROM users")
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	if !result.Allowed {
		t.Error("Validate() should allow DELETE without WHERE when whitelisted")
	}
}

func TestConfigPersistence_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	logger := zap.NewNop()
	persistence := NewConfigPersistence(configPath, logger)

	config := &ValidatorConfig{
		Whitelist: []string{"^SELECT.*", "^INSERT.*"},
		Blacklist: []string{"^DROP.*", "^TRUNCATE.*"},
	}

	err := persistence.Save(config)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	if len(data) == 0 {
		t.Error("Saved file should not be empty")
	}

	expectedContents := []string{"whitelist:", "blacklist:", "^SELECT.*", "^DROP.*"}
	for _, expected := range expectedContents {
		if !contains(string(data), expected) {
			t.Errorf("Saved file should contain %q", expected)
		}
	}
}

func TestConfigPersistence_Load(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	logger := zap.NewNop()
	persistence := NewConfigPersistence(configPath, logger)

	yamlContent := `whitelist:
  - "^SELECT.*"
  - "^INSERT.*"
blacklist:
  - "^DROP.*"
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	config, err := persistence.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(config.Whitelist) != 2 {
		t.Errorf("Load() whitelist count = %d, want 2", len(config.Whitelist))
	}
	if len(config.Blacklist) != 1 {
		t.Errorf("Load() blacklist count = %d, want 1", len(config.Blacklist))
	}
}

func TestConfigPersistence_Load_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")
	logger := zap.NewNop()
	persistence := NewConfigPersistence(configPath, logger)

	config, err := persistence.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(config.Whitelist) != 0 {
		t.Errorf("Load() whitelist = %v, want empty", config.Whitelist)
	}
	if len(config.Blacklist) != 0 {
		t.Errorf("Load() blacklist = %v, want empty", config.Blacklist)
	}
}

func TestConfigPersistence_Load_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	logger := zap.NewNop()
	persistence := NewConfigPersistence(configPath, logger)

	err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err = persistence.Load()
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}

func TestConfigPersistence_Save_InvalidDirectory(t *testing.T) {
	logger := zap.NewNop()
	persistence := NewConfigPersistence("/proc/invalid/config.yaml", logger)

	config := &ValidatorConfig{
		Whitelist: []string{"^SELECT.*"},
		Blacklist: []string{},
	}

	err := persistence.Save(config)
	if err == nil {
		t.Error("Save() should return error for invalid directory")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
