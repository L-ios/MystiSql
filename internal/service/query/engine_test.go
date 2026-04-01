package query

import (
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
