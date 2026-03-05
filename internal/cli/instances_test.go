package cli

import (
	"testing"

	"MystiSql/pkg/types"
)

func TestOutputInstancesTable(t *testing.T) {
	instances := []*types.DatabaseInstance{
		types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306),
		types.NewDatabaseInstance("test-postgres", types.DatabaseTypePostgreSQL, "localhost", 5432),
	}

	instances[0].SetDatabase("testdb")
	instances[0].SetCredentials("root", "password")

	err := outputInstancesTable(instances)
	if err != nil {
		t.Errorf("outputInstancesTable failed: %v", err)
	}
}

func TestOutputInstancesJSON(t *testing.T) {
	instances := []*types.DatabaseInstance{
		types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306),
	}

	err := outputInstancesJSON(instances)
	if err != nil {
		t.Errorf("outputInstancesJSON failed: %v", err)
	}
}

func TestOutputInstancesCSV(t *testing.T) {
	instances := []*types.DatabaseInstance{
		types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306),
	}

	err := outputInstancesCSV(instances)
	if err != nil {
		t.Errorf("outputInstancesCSV failed: %v", err)
	}
}

func TestOutputInstanceDetail(t *testing.T) {
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance.SetDatabase("testdb")
	instance.SetCredentials("root", "password")
	instance.AddLabel("env", "test")

	err := outputInstanceDetail(instance)
	if err != nil {
		t.Errorf("outputInstanceDetail failed: %v", err)
	}
}
