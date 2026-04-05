package query

import (
	"context"
	"testing"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/mysql"
	"MystiSql/internal/connection/oracle"
	"MystiSql/internal/connection/postgresql"
	"MystiSql/internal/connection/redis"
	"MystiSql/internal/connection/sqlite"
	"MystiSql/internal/discovery"
	"MystiSql/pkg/types"
)

func TestEngineDriverRegistry(t *testing.T) {
	registry := discovery.NewRegistry()
	driverReg := connection.GetRegistry()

	// 注册所有驱动（与 serve.go 一致）
	driverReg.RegisterDriver(types.DatabaseTypeMySQL, mysql.NewFactory())
	driverReg.RegisterDriver(types.DatabaseTypePostgreSQL, postgresql.NewFactory())
	driverReg.RegisterDriver(types.DatabaseTypeOracle, oracle.NewFactory())
	driverReg.RegisterDriver(types.DatabaseTypeRedis, redis.NewFactory())
	driverReg.RegisterDriver(types.DatabaseTypeSQLite, sqlite.NewFactory())

	engine := NewEngine(registry, driverReg)
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	// 验证所有已注册的驱动类型都能通过 registry 查找到工厂
	tests := []struct {
		name     string
		dbType   types.DatabaseType
		instance string
		host     string
		port     int
	}{
		{"MySQL", types.DatabaseTypeMySQL, "test-mysql", "localhost", 3306},
		{"PostgreSQL", types.DatabaseTypePostgreSQL, "test-pg", "localhost", 5432},
		{"Oracle", types.DatabaseTypeOracle, "test-oracle", "localhost", 1521},
		{"Redis", types.DatabaseTypeRedis, "test-redis", "localhost", 6379},
		{"SQLite", types.DatabaseTypeSQLite, "test-sqlite", "localhost", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证 DriverRegistry 能查到工厂
			factory, err := driverReg.GetFactory(tt.dbType)
			if err != nil {
				t.Fatalf("GetFactory(%s) failed: %v", tt.dbType, err)
			}
			if factory == nil {
				t.Fatalf("GetFactory(%s) returned nil factory", tt.dbType)
			}
		})
	}
}

func TestEngineUnsupportedDriver(t *testing.T) {
	registry := discovery.NewRegistry()
	driverReg := connection.GetRegistry()

	engine := NewEngine(registry, driverReg)

	// 注册一个不存在的实例类型
	instance := types.NewDatabaseInstance("test-unsupported", "unsupported_db_type", "localhost", 1234)
	_ = registry.Register(instance)

	// 尝试查询应该返回不支持的数据库类型错误
	_, err := engine.ExecuteQuery(t.Context(), "test-unsupported", "SELECT 1")
	if err == nil {
		t.Fatal("Expected error for unsupported database type, got nil")
	}
}

// --- Setters ---

func TestEngine_SetAuditService(t *testing.T) {
	engine := newTestEngine(t)
	// Should not panic
	engine.SetAuditService(nil)
}

func TestEngine_SetValidatorService(t *testing.T) {
	engine := newTestEngine(t)
	engine.SetValidatorService(nil)
}

func TestEngine_SetMaskingService(t *testing.T) {
	engine := newTestEngine(t)
	engine.SetMaskingService(nil)
}

// --- GetParser ---

func TestEngine_GetParser(t *testing.T) {
	engine := newTestEngine(t)
	p := engine.GetParser()
	if p == nil {
		t.Fatal("GetParser returned nil")
	}
}

// --- ListInstances ---

func TestEngine_ListInstances_Empty(t *testing.T) {
	engine := newTestEngine(t)
	instances, err := engine.ListInstances()
	if err != nil {
		t.Fatalf("ListInstances error: %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(instances))
	}
}

func TestEngine_ListInstances_WithInstances(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	inst := types.NewDatabaseInstance("mydb", types.DatabaseTypeMySQL, "localhost", 3306)
	_ = registry.Register(inst)

	instances, err := engine.ListInstances()
	if err != nil {
		t.Fatalf("ListInstances error: %v", err)
	}
	if len(instances) != 1 {
		t.Errorf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].Name != "mydb" {
		t.Errorf("instance name = %q, want %q", instances[0].Name, "mydb")
	}
}

// --- Close ---

func TestEngine_Close_Empty(t *testing.T) {
	engine := newTestEngine(t)
	if err := engine.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestEngine_Close_Twice(t *testing.T) {
	engine := newTestEngine(t)
	if err := engine.Close(); err != nil {
		t.Fatalf("first Close error: %v", err)
	}
	if err := engine.Close(); err != nil {
		t.Fatalf("second Close error: %v", err)
	}
}

// --- ExecuteExec ---

func TestEngine_ExecuteExec_UnsupportedDriver(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	inst := types.NewDatabaseInstance("test-unsupported", "unsupported_db_type", "localhost", 1234)
	_ = registry.Register(inst)

	_, err := engine.ExecuteExec(context.Background(), "test-unsupported", "INSERT INTO t VALUES (1)")
	if err == nil {
		t.Fatal("expected error for unsupported driver, got nil")
	}
}

func TestEngine_ExecuteExec_InstanceNotFound(t *testing.T) {
	engine := newTestEngine(t)
	_, err := engine.ExecuteExec(context.Background(), "nonexistent", "INSERT INTO t VALUES (1)")
	if err == nil {
		t.Fatal("expected error for nonexistent instance, got nil")
	}
}

// --- PingInstance ---

func TestEngine_PingInstance_InstanceNotFound(t *testing.T) {
	engine := newTestEngine(t)
	err := engine.PingInstance(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent instance, got nil")
	}
}

func TestEngine_PingInstance_UnsupportedDriver(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	inst := types.NewDatabaseInstance("test-unsupported", "unsupported_db_type", "localhost", 1234)
	_ = registry.Register(inst)

	err := engine.PingInstance(context.Background(), "test-unsupported")
	if err == nil {
		t.Fatal("expected error for unsupported driver, got nil")
	}
}

// --- GetInstanceHealth ---

func TestEngine_GetInstanceHealth_NotFound(t *testing.T) {
	engine := newTestEngine(t)
	_, err := engine.GetInstanceHealth(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent instance, got nil")
	}
}

func TestEngine_GetInstanceHealth_UnsupportedDriver(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	inst := types.NewDatabaseInstance("test-unsupported", "unsupported_db_type", "localhost", 1234)
	_ = registry.Register(inst)

	status, err := engine.GetInstanceHealth(context.Background(), "test-unsupported")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != types.InstanceStatusUnhealthy {
		t.Errorf("status = %q, want %q", status, types.InstanceStatusUnhealthy)
	}
}

// --- GetPoolStats ---

func TestEngine_GetPoolStats_InstanceNotFound(t *testing.T) {
	engine := newTestEngine(t)
	_, err := engine.GetPoolStats(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent instance, got nil")
	}
}

// --- resolveInstance ---

func TestEngine_ResolveInstance_ReadWriteRole(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	inst := types.NewDatabaseInstance("mydb", types.DatabaseTypeMySQL, "localhost", 3306)
	inst.Role = "readwrite"
	_ = registry.Register(inst)

	resolved, err := engine.resolveInstance(context.Background(), "mydb", "SELECT 1")
	if err != nil {
		t.Fatalf("resolveInstance error: %v", err)
	}
	if resolved != "mydb" {
		t.Errorf("resolved = %q, want %q", resolved, "mydb")
	}
}

func TestEngine_ResolveInstance_EmptyRole(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	inst := types.NewDatabaseInstance("mydb", types.DatabaseTypeMySQL, "localhost", 3306)
	inst.Role = ""
	_ = registry.Register(inst)

	resolved, err := engine.resolveInstance(context.Background(), "mydb", "SELECT 1")
	if err != nil {
		t.Fatalf("resolveInstance error: %v", err)
	}
	if resolved != "mydb" {
		t.Errorf("resolved = %q, want %q", resolved, "mydb")
	}
}

func TestEngine_ResolveInstance_ReplicaWriteRedirect(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	primary := types.NewDatabaseInstance("primary-db", types.DatabaseTypeMySQL, "localhost", 3306)
	primary.Role = "primary"
	_ = registry.Register(primary)

	replica := types.NewDatabaseInstance("replica-db", types.DatabaseTypeMySQL, "localhost", 3307)
	replica.Role = "replica"
	replica.ReplicaOf = "primary-db"
	_ = registry.Register(replica)

	// Write on replica should redirect to primary
	resolved, err := engine.resolveInstance(context.Background(), "replica-db", "INSERT INTO t VALUES (1)")
	if err != nil {
		t.Fatalf("resolveInstance error: %v", err)
	}
	if resolved != "primary-db" {
		t.Errorf("resolved = %q, want %q", resolved, "primary-db")
	}
}

func TestEngine_ResolveInstance_ReplicaWriteNoPrimary(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	replica := types.NewDatabaseInstance("replica-db", types.DatabaseTypeMySQL, "localhost", 3307)
	replica.Role = "replica"
	replica.ReplicaOf = ""
	_ = registry.Register(replica)

	_, err := engine.resolveInstance(context.Background(), "replica-db", "INSERT INTO t VALUES (1)")
	if err == nil {
		t.Fatal("expected error for replica with no primary, got nil")
	}
}

func TestEngine_ResolveInstance_InstanceNotFound(t *testing.T) {
	engine := newTestEngine(t)
	_, err := engine.resolveInstance(context.Background(), "nonexistent", "SELECT 1")
	if err == nil {
		t.Fatal("expected error for nonexistent instance, got nil")
	}
}

func TestEngine_ResolveInstance_ReplicaTxnRedirect(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	primary := types.NewDatabaseInstance("primary-db", types.DatabaseTypeMySQL, "localhost", 3306)
	primary.Role = "primary"
	_ = registry.Register(primary)

	replica := types.NewDatabaseInstance("replica-db", types.DatabaseTypeMySQL, "localhost", 3307)
	replica.Role = "replica"
	replica.ReplicaOf = "primary-db"
	_ = registry.Register(replica)

	// Transaction on replica should redirect to primary
	resolved, err := engine.resolveInstance(context.Background(), "replica-db", "BEGIN")
	if err != nil {
		t.Fatalf("resolveInstance error: %v", err)
	}
	if resolved != "primary-db" {
		t.Errorf("resolved = %q, want %q", resolved, "primary-db")
	}
}

func TestEngine_ResolveInstance_ReplicaTxnNoPrimary(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	replica := types.NewDatabaseInstance("replica-db", types.DatabaseTypeMySQL, "localhost", 3307)
	replica.Role = "replica"
	replica.ReplicaOf = ""
	_ = registry.Register(replica)

	_, err := engine.resolveInstance(context.Background(), "replica-db", "BEGIN")
	if err == nil {
		t.Fatal("expected error for replica txn with no primary, got nil")
	}
}

func TestEngine_ResolveInstance_PrimaryTxn(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	primary := types.NewDatabaseInstance("primary-db", types.DatabaseTypeMySQL, "localhost", 3306)
	primary.Role = "primary"
	_ = registry.Register(primary)

	// Transaction on primary stays on primary
	resolved, err := engine.resolveInstance(context.Background(), "primary-db", "BEGIN")
	if err != nil {
		t.Fatalf("resolveInstance error: %v", err)
	}
	if resolved != "primary-db" {
		t.Errorf("resolved = %q, want %q", resolved, "primary-db")
	}
}

func TestEngine_ResolveInstance_PrimaryReadRedirect(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	primary := types.NewDatabaseInstance("primary-db", types.DatabaseTypeMySQL, "localhost", 3306)
	primary.Role = "primary"
	_ = registry.Register(primary)

	replica := types.NewDatabaseInstance("replica-db", types.DatabaseTypeMySQL, "localhost", 3307)
	replica.Role = "replica"
	replica.ReplicaOf = "primary-db"
	_ = registry.Register(replica)

	// Read on primary should redirect to replica
	resolved, err := engine.resolveInstance(context.Background(), "primary-db", "SELECT 1")
	if err != nil {
		t.Fatalf("resolveInstance error: %v", err)
	}
	if resolved != "replica-db" {
		t.Errorf("resolved = %q, want %q", resolved, "replica-db")
	}
}

func TestEngine_ResolveInstance_PrimaryReadNoReplicas(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	primary := types.NewDatabaseInstance("primary-db", types.DatabaseTypeMySQL, "localhost", 3306)
	primary.Role = "primary"
	_ = registry.Register(primary)

	// Read on primary with no replicas stays on primary
	resolved, err := engine.resolveInstance(context.Background(), "primary-db", "SELECT 1")
	if err != nil {
		t.Fatalf("resolveInstance error: %v", err)
	}
	if resolved != "primary-db" {
		t.Errorf("resolved = %q, want %q", resolved, "primary-db")
	}
}

func TestEngine_ResolveInstance_ReplicaRead(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	replica := types.NewDatabaseInstance("replica-db", types.DatabaseTypeMySQL, "localhost", 3307)
	replica.Role = "replica"
	replica.ReplicaOf = "primary-db"
	_ = registry.Register(replica)

	// Read on replica stays on replica
	resolved, err := engine.resolveInstance(context.Background(), "replica-db", "SELECT 1")
	if err != nil {
		t.Fatalf("resolveInstance error: %v", err)
	}
	if resolved != "replica-db" {
		t.Errorf("resolved = %q, want %q", resolved, "replica-db")
	}
}

// --- findReplicas ---

func TestEngine_FindReplicas(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	primary := types.NewDatabaseInstance("primary-db", types.DatabaseTypeMySQL, "localhost", 3306)
	primary.Role = "primary"
	_ = registry.Register(primary)

	replica1 := types.NewDatabaseInstance("replica-1", types.DatabaseTypeMySQL, "localhost", 3307)
	replica1.Role = "replica"
	replica1.ReplicaOf = "primary-db"
	_ = registry.Register(replica1)

	replica2 := types.NewDatabaseInstance("replica-2", types.DatabaseTypeMySQL, "localhost", 3308)
	replica2.Role = "replica"
	replica2.ReplicaOf = "primary-db"
	_ = registry.Register(replica2)

	replicas := engine.findReplicas("primary-db")
	if len(replicas) != 2 {
		t.Fatalf("expected 2 replicas, got %d", len(replicas))
	}
	// Verify replica names are in the result
	found := map[string]bool{}
	for _, r := range replicas {
		found[r] = true
	}
	if !found["replica-1"] || !found["replica-2"] {
		t.Errorf("expected replica-1 and replica-2, got %v", replicas)
	}
}

func TestEngine_FindReplicas_NoReplicas(t *testing.T) {
	registry := discovery.NewRegistry()
	engine := NewEngine(registry, connection.GetRegistry())

	primary := types.NewDatabaseInstance("primary-db", types.DatabaseTypeMySQL, "localhost", 3306)
	primary.Role = "primary"
	_ = registry.Register(primary)

	replicas := engine.findReplicas("primary-db")
	if len(replicas) != 0 {
		t.Errorf("expected 0 replicas, got %d", len(replicas))
	}
}

// --- Helper ---

func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	registry := discovery.NewRegistry()
	driverReg := connection.GetRegistry()
	return NewEngine(registry, driverReg)
}
