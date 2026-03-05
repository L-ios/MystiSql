package cli

import (
	"testing"
	"time"

	"MystiSql/pkg/types"
)

func TestOutputQueryResultTable(t *testing.T) {
	columns := []types.ColumnInfo{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
	}

	rows := []types.Row{
		{1, "Alice", "alice@example.com"},
		{2, "Bob", "bob@example.com"},
		{3, nil, "charlie@example.com"},
	}

	result := types.NewQueryResult(columns, rows, 10*time.Millisecond)

	err := outputQueryResultTable(result)
	if err != nil {
		t.Errorf("outputQueryResultTable failed: %v", err)
	}
}

func TestOutputQueryResultJSON(t *testing.T) {
	columns := []types.ColumnInfo{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
	}

	rows := []types.Row{
		{1, "Alice"},
	}

	result := types.NewQueryResult(columns, rows, 5*time.Millisecond)

	err := outputQueryResultJSON(result, true)
	if err != nil {
		t.Errorf("outputQueryResultJSON failed: %v", err)
	}
}

func TestOutputQueryResultCSV(t *testing.T) {
	columns := []types.ColumnInfo{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
	}

	rows := []types.Row{
		{1, "Alice"},
		{2, "Bob"},
	}

	result := types.NewQueryResult(columns, rows, 3*time.Millisecond)

	err := outputQueryResultCSV(result)
	if err != nil {
		t.Errorf("outputQueryResultCSV failed: %v", err)
	}
}

func TestOutputExecResultTable(t *testing.T) {
	result := types.NewExecResult(10, 123, 5*time.Millisecond)

	err := outputExecResultTable(result)
	if err != nil {
		t.Errorf("outputExecResultTable failed: %v", err)
	}
}

func TestOutputExecResultCSV(t *testing.T) {
	result := types.NewExecResult(5, 0, 2*time.Millisecond)

	err := outputExecResultCSV(result)
	if err != nil {
		t.Errorf("outputExecResultCSV failed: %v", err)
	}
}

func TestOutputQueryResultEmptyTable(t *testing.T) {
	columns := []types.ColumnInfo{}
	rows := []types.Row{}

	result := types.NewQueryResult(columns, rows, 1*time.Millisecond)

	err := outputQueryResultTable(result)
	if err != nil {
		t.Errorf("outputQueryResultTable with empty result failed: %v", err)
	}
}
