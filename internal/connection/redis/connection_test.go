package redis

import (
	"context"
	"strconv"
	"testing"

	"MystiSql/pkg/types"

	"github.com/alicebob/miniredis/v2"
)

// testPort 将 miniredis 端口字符串转为 int。
func testPort(server *miniredis.Miniredis) int {
	port, _ := strconv.Atoi(server.Port())
	return port
}

// newTestConnection 启动一个内存 Redis 并返回已连接的 Connection。
func newTestConnection(t *testing.T) *Connection {
	t.Helper()

	server := miniredis.RunT(t)

	instance := types.NewDatabaseInstance("test-redis", types.DatabaseType("redis"), "127.0.0.1", testPort(server))
	conn := NewConnection(instance).(*Connection)

	if err := conn.Connect(context.Background()); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

func TestConnection_Lifecycle(t *testing.T) {
	server := miniredis.RunT(t)

	instance := types.NewDatabaseInstance("test-redis", types.DatabaseType("redis"), "127.0.0.1", testPort(server))
	conn := NewConnection(instance).(*Connection)

	ctx := context.Background()
	if err := conn.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if err := conn.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	if err := conn.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// 关闭后 Query/Exec/Ping 应返回 ErrConnectionClosed
	if _, err := conn.Query(ctx, "GET key"); err == nil {
		t.Error("Query after Close should fail")
	}
	if _, err := conn.Exec(ctx, "SET key value"); err == nil {
		t.Error("Exec after Close should fail")
	}
	if err := conn.Ping(ctx); err == nil {
		t.Error("Ping after Close should fail")
	}
}

func TestConnection_Connect_Refused(t *testing.T) {
	// 使用一个几乎不可能监听的端口
	instance := types.NewDatabaseInstance("bad-redis", types.DatabaseType("redis"), "127.0.0.1", 1)
	conn := NewConnection(instance).(*Connection)

	if err := conn.Connect(context.Background()); err == nil {
		t.Fatal("Connect to unreachable Redis should fail")
	}
}

func TestConnection_SetGet(t *testing.T) {
	conn := newTestConnection(t)
	ctx := context.Background()

	result, err := conn.Exec(ctx, "SET greeting hello world")
	if err != nil {
		t.Fatalf("SET failed: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("RowsAffected = %d, want 1", result.RowsAffected)
	}

	queryResult, err := conn.Query(ctx, "GET greeting")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if len(queryResult.Rows) != 1 || queryResult.Rows[0][0] != "hello world" {
		t.Errorf("GET result = %v, want [hello world]", queryResult.Rows)
	}

	// GET 不存在的 key 应返回一行 NULL
	missing, err := conn.Query(ctx, "GET missing")
	if err != nil {
		t.Fatalf("GET missing failed: %v", err)
	}
	if len(missing.Rows) != 1 || missing.Rows[0][0] != nil {
		t.Errorf("GET missing result = %v, want one row with nil", missing.Rows)
	}
}

func TestConnection_HSetHGetAll(t *testing.T) {
	conn := newTestConnection(t)
	ctx := context.Background()

	if _, err := conn.Exec(ctx, "HSET user:1 name alice"); err != nil {
		t.Fatalf("HSET failed: %v", err)
	}
	if _, err := conn.Exec(ctx, "HSET user:1 role admin"); err != nil {
		t.Fatalf("HSET failed: %v", err)
	}

	result, err := conn.Query(ctx, "HGETALL user:1")
	if err != nil {
		t.Fatalf("HGETALL failed: %v", err)
	}

	fields := map[string]string{}
	for _, row := range result.Rows {
		fields[row[0].(string)] = row[1].(string)
	}
	if fields["name"] != "alice" || fields["role"] != "admin" {
		t.Errorf("HGETALL = %v, want name=alice role=admin", fields)
	}
}

func TestConnection_ListOps(t *testing.T) {
	conn := newTestConnection(t)
	ctx := context.Background()

	if _, err := conn.Exec(ctx, "RPUSH jobs task1"); err != nil {
		t.Fatalf("RPUSH failed: %v", err)
	}
	if _, err := conn.Exec(ctx, "RPUSH jobs task2"); err != nil {
		t.Fatalf("RPUSH failed: %v", err)
	}

	result, err := conn.Query(ctx, "LRANGE jobs 0 -1")
	if err != nil {
		t.Fatalf("LRANGE failed: %v", err)
	}
	if len(result.Rows) != 2 || result.Rows[0][0] != "task1" || result.Rows[1][0] != "task2" {
		t.Errorf("LRANGE = %v, want [task1 task2]", result.Rows)
	}
}

func TestConnection_SetOps(t *testing.T) {
	conn := newTestConnection(t)
	ctx := context.Background()

	if _, err := conn.Exec(ctx, "SADD roles admin"); err != nil {
		t.Fatalf("SADD failed: %v", err)
	}

	result, err := conn.Query(ctx, "SMEMBERS roles")
	if err != nil {
		t.Fatalf("SMEMBERS failed: %v", err)
	}
	if len(result.Rows) != 1 || result.Rows[0][0] != "admin" {
		t.Errorf("SMEMBERS = %v, want [admin]", result.Rows)
	}
}

func TestConnection_IncrDel(t *testing.T) {
	conn := newTestConnection(t)
	ctx := context.Background()

	result, err := conn.Exec(ctx, "INCR counter")
	if err != nil {
		t.Fatalf("INCR failed: %v", err)
	}
	if result.LastInsertID != 1 {
		t.Errorf("INCR LastInsertID = %d, want 1", result.LastInsertID)
	}

	delResult, err := conn.Exec(ctx, "DEL counter")
	if err != nil {
		t.Fatalf("DEL failed: %v", err)
	}
	if delResult.RowsAffected != 1 {
		t.Errorf("DEL RowsAffected = %d, want 1", delResult.RowsAffected)
	}
}

func TestConnection_UnsupportedCommands(t *testing.T) {
	conn := newTestConnection(t)
	ctx := context.Background()

	if _, err := conn.Query(ctx, "UNKNOWN query"); err == nil {
		t.Error("Query with unsupported command should fail")
	}
	if _, err := conn.Exec(ctx, "FLUSHALL"); err == nil {
		t.Error("Exec with unsupported command should fail")
	}
}
