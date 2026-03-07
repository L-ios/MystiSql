package mysql

import (
	"context"
	"testing"

	"MystiSql/pkg/types"
)

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		instance *types.DatabaseInstance
		want     string
	}{
		{
			name: "完整配置",
			instance: &types.DatabaseInstance{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "secret",
				Database: "testdb",
			},
			want: "root:secret@tcp(localhost:3306)/testdb?parseTime=true&loc=Local&charset=utf8mb4&timeout=30s",
		},
		{
			name: "无密码",
			instance: &types.DatabaseInstance{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Database: "testdb",
			},
			want: "root@tcp(localhost:3306)/testdb?parseTime=true&loc=Local&charset=utf8mb4&timeout=30s",
		},
		{
			name: "无用户名",
			instance: &types.DatabaseInstance{
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
			},
			want: "tcp(localhost:3306)/testdb?parseTime=true&loc=Local&charset=utf8mb4&timeout=30s",
		},
		{
			name: "无数据库",
			instance: &types.DatabaseInstance{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "secret",
			},
			want: "root:secret@tcp(localhost:3306)/?parseTime=true&loc=Local&charset=utf8mb4&timeout=30s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDSN(tt.instance)
			if got != tt.want {
				t.Errorf("buildDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConnection(t *testing.T) {
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewConnection(instance)

	if conn == nil {
		t.Fatal("NewConnection() returned nil")
	}

	// 转换为具体类型以访问私有字段
	mysqlConn, ok := conn.(*Connection)
	if !ok {
		t.Fatal("conn 不是 *Connection 类型")
	}

	if mysqlConn.instance != instance {
		t.Error("实例未正确设置")
	}

	if mysqlConn.db != nil {
		t.Error("新连接的 db 应该为 nil")
	}
}

func TestConnection_Close_NotOpen(t *testing.T) {
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewConnection(instance)

	// 测试关闭未打开的连接（幂等操作）
	err := conn.Close()
	if err != nil {
		t.Errorf("关闭未打开的连接应该返回 nil，实际返回: %v", err)
	}
}

func TestConnection_Ping_NotOpen(t *testing.T) {
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewConnection(instance)

	// 测试 ping 未打开的连接
	err := conn.Ping(context.Background())
	if err == nil {
		t.Error("ping 未打开的连接应该返回错误")
	}
}

func TestConnection_Query_NotOpen(t *testing.T) {
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewConnection(instance)

	// 测试查询未打开的连接
	_, err := conn.Query(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("查询未打开的连接应该返回错误")
	}
}

func TestConnection_Exec_NotOpen(t *testing.T) {
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	conn := NewConnection(instance)

	// 测试执行未打开的连接
	_, err := conn.Exec(context.Background(), "INSERT INTO test VALUES (1)")
	if err == nil {
		t.Error("执行未打开的连接应该返回错误")
	}
}

// 注意：实际的连接测试需要真实的 MySQL 数据库
// 以下测试需要测试数据库环境才能运行

func TestConnection_Connect_Real(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的测试")
	}

	// TODO: 添加真实数据库连接测试
	// 需要设置测试数据库环境
}

func TestConnection_Query_Real(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的测试")
	}

	// TODO: 添加真实数据库查询测试
	// 需要设置测试数据库环境
}

func TestConnection_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的测试")
	}

	// TODO: 添加上下文取消测试
	// 需要设置测试数据库环境
}
