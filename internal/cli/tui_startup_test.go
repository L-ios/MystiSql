package cli

import (
	"testing"
)

func TestTUIDefaultStartup(t *testing.T) {
	if rootCmd.RunE == nil {
		t.Error("rootCmd.RunE should not be nil for default TUI startup")
	}
}

func TestSubcommandCompatibility(t *testing.T) {
	commands := rootCmd.Commands()
	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name()] = true
	}

	expectedCommands := []string{"query", "version"}
	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}

	if commandNames["tui"] {
		t.Error("'tui' subcommand should be removed")
	}
}

func TestQuerySubcommandExists(t *testing.T) {
	queryCmd, _, err := rootCmd.Find([]string{"query"})
	if err != nil {
		t.Errorf("Failed to find query subcommand: %v", err)
	}

	if queryCmd == nil {
		t.Error("query subcommand should exist")
	}

	if queryCmd.Name() != "query" {
		t.Errorf("Expected command name 'query', got '%s'", queryCmd.Name())
	}
}
