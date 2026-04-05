package sqlite

import (
	"context"
	"testing"

	"MystiSql/internal/connection"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	require.NotNil(t, factory, "NewFactory() 不应返回 nil")
	var _ connection.ConnectionFactory = factory
}

func TestFactory_CreateConnection(t *testing.T) {
	t.Run("成功创建连接", func(t *testing.T) {
		factory := NewFactory()
		instance := types.NewDatabaseInstance("test-sqlite", types.DatabaseTypeSQLite, "", 0)

		conn, err := factory.CreateConnection(instance)

		require.NoError(t, err)
		require.NotNil(t, conn)
		var _ connection.Connection = conn
	})
}

func TestNewConnection(t *testing.T) {
	t.Run("返回非 nil Connection", func(t *testing.T) {
		instance := types.NewDatabaseInstance("test-sqlite", types.DatabaseTypeSQLite, "", 0)
		conn := NewConnection(instance)
		require.NotNil(t, conn)
		var _ connection.Connection = conn
	})
}

// --- buildDSN 测试 ---

func TestConnection_BuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		database string
		wantDSN  string
	}{
		{
			name:     "内存数据库",
			database: ":memory:",
			wantDSN:  ":memory:",
		},
		{
			name:     "空数据库名",
			database: "",
			wantDSN:  ":memory:",
		},
		{
			name:     "文件路径",
			database: "/tmp/test.db",
			wantDSN:  "/tmp/test.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := types.NewDatabaseInstance("test-sqlite", types.DatabaseTypeSQLite, "", 0)
			instance.SetDatabase(tt.database)
			conn := NewConnection(instance).(*Connection)
			dsn := conn.buildDSN()
			assert.Equal(t, tt.wantDSN, dsn)
		})
	}
}

// --- nil db 方法测试 ---

func TestConnection_Query_NilDB(t *testing.T) {
	t.Run("nil db 返回 ErrConnectionClosed", func(t *testing.T) {
		instance := types.NewDatabaseInstance("test-sqlite", types.DatabaseTypeSQLite, "", 0)
		conn := NewConnection(instance)

		result, err := conn.Query(context.Background(), "SELECT 1")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, errors.ErrConnectionClosed)
	})
}

func TestConnection_Exec_NilDB(t *testing.T) {
	t.Run("nil db 返回 ErrConnectionClosed", func(t *testing.T) {
		instance := types.NewDatabaseInstance("test-sqlite", types.DatabaseTypeSQLite, "", 0)
		conn := NewConnection(instance)

		result, err := conn.Exec(context.Background(), "INSERT INTO t VALUES (1)")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, errors.ErrConnectionClosed)
	})
}

func TestConnection_Ping_NilDB(t *testing.T) {
	t.Run("nil db 返回 ErrConnectionClosed", func(t *testing.T) {
		instance := types.NewDatabaseInstance("test-sqlite", types.DatabaseTypeSQLite, "", 0)
		conn := NewConnection(instance)

		err := conn.Ping(context.Background())

		assert.ErrorIs(t, err, errors.ErrConnectionClosed)
	})
}

func TestConnection_Close(t *testing.T) {
	t.Run("关闭未连接的连接不报错", func(t *testing.T) {
		instance := types.NewDatabaseInstance("test-sqlite", types.DatabaseTypeSQLite, "", 0)
		conn := NewConnection(instance)

		err := conn.Close()

		assert.NoError(t, err)
	})

	t.Run("nil db 关闭时状态不变（early return）", func(t *testing.T) {
		instance := types.NewDatabaseInstance("test-sqlite", types.DatabaseTypeSQLite, "", 0)
		instance.SetStatus(types.InstanceStatusHealthy)
		conn := NewConnection(instance)

		_ = conn.Close()

		// Close 在 db==nil 时直接 return，不会修改 status
		assert.Equal(t, types.InstanceStatusHealthy, instance.Status)
	})

	t.Run("重复关闭不报错", func(t *testing.T) {
		instance := types.NewDatabaseInstance("test-sqlite", types.DatabaseTypeSQLite, "", 0)
		conn := NewConnection(instance)

		assert.NoError(t, conn.Close())
		assert.NoError(t, conn.Close())
	})
}

func TestConnection_Close_Then_Query(t *testing.T) {
	t.Run("关闭后再查询返回 ErrConnectionClosed", func(t *testing.T) {
		instance := types.NewDatabaseInstance("test-sqlite", types.DatabaseTypeSQLite, "", 0)
		conn := NewConnection(instance)
		_ = conn.Close()

		result, err := conn.Query(context.Background(), "SELECT 1")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, errors.ErrConnectionClosed)
	})
}
