package cli

import (
	"testing"
)

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"default version", "0.1.0", "0.1.0"},
		{"custom version", "1.0.0", "1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVersion(tt.version, "test", "2024-01-01")
			if got := GetVersion(); got != tt.version {
				t.Errorf("GetVersion() = %v, want %v", got, tt.version)
			}
		})
	}
}

func TestGetFullVersion(t *testing.T) {
	SetVersion("0.2.0", "abc123", "2024-01-01")

	versionInfo := GetFullVersion()

	if versionInfo["version"] != "0.2.0" {
		t.Errorf("Expected version 0.2.0, got %s", versionInfo["version"])
	}

	if versionInfo["gitCommit"] != "abc123" {
		t.Errorf("Expected gitCommit abc123, got %s", versionInfo["gitCommit"])
	}

	if versionInfo["buildDate"] != "2024-01-01" {
		t.Errorf("Expected buildDate 2024-01-01, got %s", versionInfo["buildDate"])
	}
}
