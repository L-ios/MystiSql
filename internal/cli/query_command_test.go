package cli

import (
	"testing"
)

func TestQueryCommandExists(t *testing.T) {
	queryCmd, _, err := rootCmd.Find([]string{"query"})
	if err != nil {
		t.Fatalf("Failed to find query subcommand: %v", err)
	}

	if queryCmd == nil {
		t.Fatal("query subcommand should exist")
	}

	if queryCmd.Name() != "query" {
		t.Errorf("Expected command name 'query', got '%s'", queryCmd.Name())
	}

	flags := queryCmd.Flags()
	if flags.Lookup("sql") != nil {
		t.Error("query command should not have --sql flag")
	}

	if flags.Lookup("format") == nil {
		t.Error("query command should have --format flag")
	}

	if flags.Lookup("timeout") == nil {
		t.Error("query command should have --timeout flag")
	}
}

func TestTUISubcommandRemoved(t *testing.T) {
	commands := rootCmd.Commands()
	for _, cmd := range commands {
		if cmd.Name() == "tui" {
			t.Error("tui subcommand should be removed")
		}
	}
}
