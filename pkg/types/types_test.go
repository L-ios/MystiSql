package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDatabaseTypeConstants(t *testing.T) {
	tests := []struct {
		dt   DatabaseType
		want string
	}{
		{DatabaseTypeMySQL, "mysql"},
		{DatabaseTypePostgreSQL, "postgresql"},
		{DatabaseTypeOracle, "oracle"},
		{DatabaseTypeRedis, "redis"},
		{DatabaseTypeSQLite, "sqlite"},
		{DatabaseTypeMSSQL, "mssql"},
		{DatabaseTypeMongoDB, "mongodb"},
		{DatabaseTypeElasticsearch, "elasticsearch"},
		{DatabaseTypeClickHouse, "clickhouse"},
		{DatabaseTypeEtcd, "etcd"},
	}
	for _, tt := range tests {
		if string(tt.dt) != tt.want {
			t.Errorf("DatabaseType %q = %q, want %q", tt.dt, string(tt.dt), tt.want)
		}
	}
}

func TestInstanceStatusConstants(t *testing.T) {
	tests := []struct {
		s    InstanceStatus
		want string
	}{
		{InstanceStatusUnknown, "unknown"},
		{InstanceStatusHealthy, "healthy"},
		{InstanceStatusUnhealthy, "unhealthy"},
	}
	for _, tt := range tests {
		if string(tt.s) != tt.want {
			t.Errorf("InstanceStatus = %q, want %q", string(tt.s), tt.want)
		}
	}
}

func TestNewDatabaseInstance(t *testing.T) {
	inst := NewDatabaseInstance("test-db", DatabaseTypeMySQL, "localhost", 3306)

	if inst.Name != "test-db" {
		t.Errorf("Name = %q, want %q", inst.Name, "test-db")
	}
	if inst.Type != DatabaseTypeMySQL {
		t.Errorf("Type = %q, want %q", inst.Type, DatabaseTypeMySQL)
	}
	if inst.Host != "localhost" {
		t.Errorf("Host = %q, want %q", inst.Host, "localhost")
	}
	if inst.Port != 3306 {
		t.Errorf("Port = %d, want %d", inst.Port, 3306)
	}
	if inst.Status != InstanceStatusUnknown {
		t.Errorf("Status = %q, want %q", inst.Status, InstanceStatusUnknown)
	}
	if inst.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if inst.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestDatabaseInstance_SetCredentials(t *testing.T) {
	inst := NewDatabaseInstance("test", DatabaseTypeMySQL, "localhost", 3306)
	before := inst.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	inst.SetCredentials("admin", "secret123")

	if inst.Username != "admin" {
		t.Errorf("Username = %q, want %q", inst.Username, "admin")
	}
	if inst.Password != "secret123" {
		t.Errorf("Password = %q, want %q", inst.Password, "secret123")
	}
	if !inst.UpdatedAt.After(before) {
		t.Error("UpdatedAt should be updated after SetCredentials")
	}
}

func TestDatabaseInstance_SetStatus(t *testing.T) {
	inst := NewDatabaseInstance("test", DatabaseTypeMySQL, "localhost", 3306)
	before := inst.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	inst.SetStatus(InstanceStatusHealthy)

	if inst.Status != InstanceStatusHealthy {
		t.Errorf("Status = %q, want %q", inst.Status, InstanceStatusHealthy)
	}
	if !inst.UpdatedAt.After(before) {
		t.Error("UpdatedAt should be updated after SetStatus")
	}
}

func TestDatabaseInstance_SetDatabase(t *testing.T) {
	inst := NewDatabaseInstance("test", DatabaseTypeMySQL, "localhost", 3306)
	inst.SetDatabase("mydb")

	if inst.Database != "mydb" {
		t.Errorf("Database = %q, want %q", inst.Database, "mydb")
	}
}

func TestDatabaseInstance_AddLabel(t *testing.T) {
	inst := NewDatabaseInstance("test", DatabaseTypeMySQL, "localhost", 3306)

	inst.AddLabel("env", "production")
	inst.AddLabel("team", "backend")

	if inst.Labels == nil {
		t.Fatal("Labels should not be nil after AddLabel")
	}
	if inst.Labels["env"] != "production" {
		t.Errorf("Labels[env] = %q, want %q", inst.Labels["env"], "production")
	}
	if inst.Labels["team"] != "backend" {
		t.Errorf("Labels[team] = %q, want %q", inst.Labels["team"], "backend")
	}

	inst.AddLabel("env", "staging")
	if inst.Labels["env"] != "staging" {
		t.Errorf("Labels[env] = %q after overwrite, want %q", inst.Labels["env"], "staging")
	}
}

func TestDatabaseInstance_AddAnnotation(t *testing.T) {
	inst := NewDatabaseInstance("test", DatabaseTypeMySQL, "localhost", 3306)

	inst.AddAnnotation("owner", "team-a")

	if inst.Annotations == nil {
		t.Fatal("Annotations should not be nil after AddAnnotation")
	}
	if inst.Annotations["owner"] != "team-a" {
		t.Errorf("Annotations[owner] = %q, want %q", inst.Annotations["owner"], "team-a")
	}
}

func TestQueryResult_NewQueryResult(t *testing.T) {
	cols := []ColumnInfo{{Name: "id", Type: "int"}, {Name: "name", Type: "text"}}
	rows := []Row{{1, "alice"}, {2, "bob"}}
	execTime := 10 * time.Millisecond

	qr := NewQueryResult(cols, rows, execTime)

	if qr.RowCount != 2 {
		t.Errorf("RowCount = %d, want %d", qr.RowCount, 2)
	}
	if qr.Truncated {
		t.Error("Truncated should be false for new result")
	}
	if qr.ExecutionTime != execTime {
		t.Errorf("ExecutionTime = %v, want %v", qr.ExecutionTime, execTime)
	}
	if len(qr.Columns) != 2 {
		t.Errorf("len(Columns) = %d, want %d", len(qr.Columns), 2)
	}
}

func TestQueryResult_GetColumnNames(t *testing.T) {
	qr := NewQueryResult(
		[]ColumnInfo{{Name: "id", Type: "int"}, {Name: "name", Type: "text"}},
		[]Row{{1, "alice"}},
		0,
	)

	names := qr.GetColumnNames()
	if len(names) != 2 || names[0] != "id" || names[1] != "name" {
		t.Errorf("GetColumnNames() = %v, want [id, name]", names)
	}
}

func TestQueryResult_GetColumnNames_Empty(t *testing.T) {
	qr := NewQueryResult(nil, nil, 0)
	names := qr.GetColumnNames()
	if len(names) != 0 {
		t.Errorf("GetColumnNames() on empty = %v, want []", names)
	}
}

func TestQueryResult_GetRowByIndex(t *testing.T) {
	qr := NewQueryResult(
		[]ColumnInfo{{Name: "id"}},
		[]Row{{1}, {2}, {3}},
		0,
	)

	row, ok := qr.GetRowByIndex(1)
	if !ok || row[0] != 2 {
		t.Errorf("GetRowByIndex(1) = (%v, %v), want ([2], true)", row, ok)
	}

	_, ok = qr.GetRowByIndex(-1)
	if ok {
		t.Error("GetRowByIndex(-1) should return false")
	}

	_, ok = qr.GetRowByIndex(3)
	if ok {
		t.Error("GetRowByIndex(3) should return false for out of bounds")
	}
}

func TestQueryResult_GetValue(t *testing.T) {
	qr := NewQueryResult(
		[]ColumnInfo{{Name: "id"}, {Name: "name"}},
		[]Row{{1, "alice"}, {2, "bob"}},
		0,
	)

	val, ok := qr.GetValue(0, "name")
	if !ok || val != "alice" {
		t.Errorf("GetValue(0, name) = (%v, %v), want (alice, true)", val, ok)
	}

	val, ok = qr.GetValue(1, "id")
	if !ok || val != 2 {
		t.Errorf("GetValue(1, id) = (%v, %v), want (2, true)", val, ok)
	}

	_, ok = qr.GetValue(0, "nonexistent")
	if ok {
		t.Error("GetValue with nonexistent column should return false")
	}

	_, ok = qr.GetValue(99, "id")
	if ok {
		t.Error("GetValue with out-of-bounds row should return false")
	}
}

func TestExecResult_NewExecResult(t *testing.T) {
	er := NewExecResult(5, 42, 5*time.Millisecond)

	if er.RowsAffected != 5 {
		t.Errorf("RowsAffected = %d, want %d", er.RowsAffected, 5)
	}
	if er.LastInsertID != 42 {
		t.Errorf("LastInsertID = %d, want %d", er.LastInsertID, 42)
	}
	if er.ExecutionTime != 5*time.Millisecond {
		t.Errorf("ExecutionTime = %v, want %v", er.ExecutionTime, 5*time.Millisecond)
	}
}

func TestIsolationLevel_String(t *testing.T) {
	tests := []struct {
		il   IsolationLevel
		want string
	}{
		{IsolationLevelDefault, "DEFAULT"},
		{IsolationLevelReadUncommitted, "READ UNCOMMITTED"},
		{IsolationLevelReadCommitted, "READ COMMITTED"},
		{IsolationLevelRepeatableRead, "REPEATABLE READ"},
		{IsolationLevelSerializable, "SERIALIZABLE"},
	}
	for _, tt := range tests {
		if got := tt.il.String(); got != tt.want {
			t.Errorf("IsolationLevel(%d).String() = %q, want %q", tt.il, got, tt.want)
		}
	}
}

func TestIsolationLevel_MarshalJSON(t *testing.T) {
	il := IsolationLevelReadCommitted
	data, err := json.Marshal(il)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	if string(data) != `"READ COMMITTED"` {
		t.Errorf("MarshalJSON = %s, want %s", string(data), `"READ COMMITTED"`)
	}
}

func TestIsolationLevel_UnmarshalJSON_String(t *testing.T) {
	tests := []struct {
		input string
		want  IsolationLevel
	}{
		{`"READ UNCOMMITTED"`, IsolationLevelReadUncommitted},
		{`"READ COMMITTED"`, IsolationLevelReadCommitted},
		{`"REPEATABLE READ"`, IsolationLevelRepeatableRead},
		{`"SERIALIZABLE"`, IsolationLevelSerializable},
		{`"read committed"`, IsolationLevelReadCommitted},
		{`"readuncommitted"`, IsolationLevelReadUncommitted},
		{`"repeatableread"`, IsolationLevelRepeatableRead},
		{`"unknown"`, IsolationLevelDefault},
	}

	for _, tt := range tests {
		var il IsolationLevel
		if err := json.Unmarshal([]byte(tt.input), &il); err != nil {
			t.Errorf("UnmarshalJSON(%s) error: %v", tt.input, err)
			continue
		}
		if il != tt.want {
			t.Errorf("UnmarshalJSON(%s) = %d, want %d", tt.input, il, tt.want)
		}
	}
}

func TestIsolationLevel_UnmarshalJSON_Int(t *testing.T) {
	var il IsolationLevel
	if err := json.Unmarshal([]byte("2"), &il); err != nil {
		t.Fatalf("UnmarshalJSON int error: %v", err)
	}
	if il != IsolationLevelReadCommitted {
		t.Errorf("UnmarshalJSON(2) = %d, want %d", il, IsolationLevelReadCommitted)
	}
}

func TestIsolationLevel_UnmarshalJSON_Invalid(t *testing.T) {
	var il IsolationLevel
	err := json.Unmarshal([]byte(`"not-json`), &il)
	if err == nil {
		t.Error("UnmarshalJSON with invalid JSON should return error")
	}
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}
	if cfg.Auth.Enabled != false {
		t.Error("Auth.Enabled should be false by default")
	}
	if len(cfg.Auth.Whitelist) != 1 || cfg.Auth.Whitelist[0] != "/health" {
		t.Errorf("Auth.Whitelist = %v, want [/health]", cfg.Auth.Whitelist)
	}
	if cfg.Audit.RetentionDays != 30 {
		t.Errorf("Audit.RetentionDays = %d, want %d", cfg.Audit.RetentionDays, 30)
	}
	if cfg.Pool.MaxConnections != 10 {
		t.Errorf("Pool.MaxConnections = %d, want %d", cfg.Pool.MaxConnections, 10)
	}
	if cfg.Discovery.Type != "static" {
		t.Errorf("Discovery.Type = %q, want %q", cfg.Discovery.Type, "static")
	}
	if cfg.WebSocket.Enabled != true {
		t.Error("WebSocket.Enabled should be true by default")
	}
	if len(cfg.Instances) != 0 {
		t.Errorf("Instances should be empty, got %d", len(cfg.Instances))
	}
}

func TestInstanceConfig_ToDatabaseInstance(t *testing.T) {
	ic := &InstanceConfig{
		Name:     "my-db",
		Type:     DatabaseTypePostgreSQL,
		Host:     "db.example.com",
		Port:     5432,
		Username: "admin",
		Password: "secret",
		Database: "appdb",
		Labels:   map[string]string{"env": "prod"},
	}

	inst := ic.ToDatabaseInstance()

	if inst.Name != "my-db" {
		t.Errorf("Name = %q, want %q", inst.Name, "my-db")
	}
	if inst.Type != DatabaseTypePostgreSQL {
		t.Errorf("Type = %q, want %q", inst.Type, DatabaseTypePostgreSQL)
	}
	if inst.Username != "admin" {
		t.Errorf("Username = %q, want %q", inst.Username, "admin")
	}
	if inst.Password != "secret" {
		t.Errorf("Password = %q, want %q", inst.Password, "secret")
	}
	if inst.Database != "appdb" {
		t.Errorf("Database = %q, want %q", inst.Database, "appdb")
	}
	if inst.Labels["env"] != "prod" {
		t.Errorf("Labels[env] = %q, want %q", inst.Labels["env"], "prod")
	}
}

func TestInstanceConfig_ToDatabaseInstance_NoOptionalFields(t *testing.T) {
	ic := &InstanceConfig{
		Name: "minimal",
		Type: DatabaseTypeMySQL,
		Host: "localhost",
		Port: 3306,
	}

	inst := ic.ToDatabaseInstance()

	if inst.Database != "" {
		t.Errorf("Database should be empty, got %q", inst.Database)
	}
	if inst.Labels != nil {
		t.Errorf("Labels should be nil, got %v", inst.Labels)
	}
}

func TestRBACConfig(t *testing.T) {
	rc := &RBACConfig{Enabled: true, Roles: []string{"admin", "viewer"}}
	if !rc.Enabled {
		t.Error("Enabled should be true")
	}
	if len(rc.Roles) != 2 {
		t.Errorf("len(Roles) = %d, want %d", len(rc.Roles), 2)
	}
}

func TestMaskingConfig(t *testing.T) {
	mc := &MaskingConfig{Enabled: true}
	if !mc.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestOIDCConfig(t *testing.T) {
	oc := &OIDCConfig{
		Providers: []OIDCProviderConfig{
			{
				Name:         "keycloak",
				Issuer:       "https://keycloak.example.com/realms/myrealm",
				ClientID:     "my-client",
				ClientSecret: "my-secret",
				RedirectURL:  "https://mystisql.example.com/callback",
				Scopes:       []string{"openid", "profile"},
				RoleClaim:    "roles",
			},
		},
	}

	if len(oc.Providers) != 1 {
		t.Fatalf("len(Providers) = %d, want %d", len(oc.Providers), 1)
	}
	p := oc.Providers[0]
	if p.Name != "keycloak" {
		t.Errorf("Provider.Name = %q, want %q", p.Name, "keycloak")
	}
	if p.Issuer != "https://keycloak.example.com/realms/myrealm" {
		t.Errorf("Provider.Issuer = %q", p.Issuer)
	}
	if len(p.Scopes) != 2 {
		t.Errorf("len(Scopes) = %d, want %d", len(p.Scopes), 2)
	}
}
