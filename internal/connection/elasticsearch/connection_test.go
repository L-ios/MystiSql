package elasticsearch

import (
	"context"
	"testing"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Factory 测试 ---

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	require.NotNil(t, factory, "NewFactory() 不应返回 nil")

	// 验证 Factory 实现了 ConnectionFactory 接口
	var _ connection.ConnectionFactory = factory
}

func TestFactory_CreateConnection(t *testing.T) {
	t.Run("成功创建连接", func(t *testing.T) {
		// Arrange
		factory := NewFactory()
		instance := types.NewDatabaseInstance("test-es", types.DatabaseTypeElasticsearch, "localhost", 9200)

		// Act
		conn, err := factory.CreateConnection(instance)

		// Assert
		require.NoError(t, err, "CreateConnection 不应返回错误")
		require.NotNil(t, conn, "CreateConnection 不应返回 nil 连接")

		// 验证实现了 Connection 接口
		var _ connection.Connection = conn
	})
}

// --- NewConnection 测试 ---

func TestNewConnection(t *testing.T) {
	t.Run("返回非 nil Connection", func(t *testing.T) {
		// Arrange
		instance := types.NewDatabaseInstance("test-es", types.DatabaseTypeElasticsearch, "localhost", 9200)
		instance.Database = "test-index"

		// Act
		conn := NewConnection(instance)

		// Assert
		require.NotNil(t, conn, "NewConnection 不应返回 nil")

		// 验证接口实现
		var _ connection.Connection = conn
	})
}

// --- nil client 方法测试（不需要真实 ES 服务器）---

func TestConnection_Query_NilClient(t *testing.T) {
	t.Run("nil client 返回 ErrConnectionClosed", func(t *testing.T) {
		// Arrange
		instance := types.NewDatabaseInstance("test-es", types.DatabaseTypeElasticsearch, "localhost", 9200)
		conn := NewConnection(instance)

		// Act
		result, err := conn.Query(context.Background(), `{"query":{"match_all":{}}}`)

		// Assert
		assert.Nil(t, result, "Query 应返回 nil result")
		assert.ErrorIs(t, err, errors.ErrConnectionClosed, "Query 应返回 ErrConnectionClosed")
	})
}

func TestConnection_Exec_NilClient(t *testing.T) {
	t.Run("nil client 返回 ErrConnectionClosed", func(t *testing.T) {
		// Arrange
		instance := types.NewDatabaseInstance("test-es", types.DatabaseTypeElasticsearch, "localhost", 9200)
		conn := NewConnection(instance)

		// Act
		result, err := conn.Exec(context.Background(), `{"field":"value"}`)

		// Assert
		assert.Nil(t, result, "Exec 应返回 nil result")
		assert.ErrorIs(t, err, errors.ErrConnectionClosed, "Exec 应返回 ErrConnectionClosed")
	})
}

func TestConnection_Ping_NilClient(t *testing.T) {
	t.Run("nil client 返回 ErrConnectionClosed", func(t *testing.T) {
		// Arrange
		instance := types.NewDatabaseInstance("test-es", types.DatabaseTypeElasticsearch, "localhost", 9200)
		conn := NewConnection(instance)

		// Act
		err := conn.Ping(context.Background())

		// Assert
		assert.ErrorIs(t, err, errors.ErrConnectionClosed, "Ping 应返回 ErrConnectionClosed")
	})
}

func TestConnection_Close(t *testing.T) {
	t.Run("关闭未连接的连接不报错", func(t *testing.T) {
		// Arrange
		instance := types.NewDatabaseInstance("test-es", types.DatabaseTypeElasticsearch, "localhost", 9200)
		conn := NewConnection(instance)

		// Act
		err := conn.Close()

		// Assert
		assert.NoError(t, err, "Close 不应返回错误")
	})

	t.Run("关闭后状态变为 unknown", func(t *testing.T) {
		// Arrange
		instance := types.NewDatabaseInstance("test-es", types.DatabaseTypeElasticsearch, "localhost", 9200)
		instance.SetStatus(types.InstanceStatusHealthy)
		conn := NewConnection(instance)

		// Act
		_ = conn.Close()

		// Assert
		assert.Equal(t, types.InstanceStatusUnknown, instance.Status, "关闭后状态应为 unknown")
	})

	t.Run("重复关闭不报错", func(t *testing.T) {
		// Arrange
		instance := types.NewDatabaseInstance("test-es", types.DatabaseTypeElasticsearch, "localhost", 9200)
		conn := NewConnection(instance)

		// Act & Assert
		assert.NoError(t, conn.Close(), "第一次 Close 不应报错")
		assert.NoError(t, conn.Close(), "第二次 Close 也不应报错")
	})
}

func TestConnection_Close_Then_Query(t *testing.T) {
	t.Run("关闭后再查询返回 ErrConnectionClosed", func(t *testing.T) {
		// Arrange
		instance := types.NewDatabaseInstance("test-es", types.DatabaseTypeElasticsearch, "localhost", 9200)
		conn := NewConnection(instance)
		_ = conn.Close()

		// Act
		result, err := conn.Query(context.Background(), `{"query":{"match_all":{}}}`)

		// Assert
		assert.Nil(t, result)
		assert.ErrorIs(t, err, errors.ErrConnectionClosed)
	})
}

// --- parseSearchResponse 测试（核心纯函数测试）---

func TestParseSearchResponse_ValidWithHits(t *testing.T) {
	t.Run("包含多条记录的响应", func(t *testing.T) {
		// Arrange
		body := []byte(`{
			"hits": {
				"total": {"value": 2},
				"hits": [
					{"_source": {"name": "alice", "age": 30}},
					{"_source": {"name": "bob", "age": 25}}
				]
			}
		}`)
		execTime := 10 * time.Millisecond

		// Act
		result, err := parseSearchResponse(body, execTime)

		// Assert
		require.NoError(t, err, "parseSearchResponse 不应返回错误")
		require.NotNil(t, result, "结果不应为 nil")
		assert.Equal(t, 2, result.RowCount, "应有 2 行数据")
		assert.Equal(t, execTime, result.ExecutionTime, "执行时间应匹配")
		// 列数应为 2（name, age），但由于 map 遍历顺序不确定，只检查数量
		assert.Equal(t, 2, len(result.Columns), "应有 2 列")
	})

	t.Run("包含单条记录的响应", func(t *testing.T) {
		// Arrange
		body := []byte(`{
			"hits": {
				"total": {"value": 1},
				"hits": [
					{"_source": {"title": "hello", "content": "world"}}
				]
			}
		}`)

		// Act
		result, err := parseSearchResponse(body, 5*time.Millisecond)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 1, result.RowCount)
		assert.Equal(t, 2, len(result.Columns))
	})
}

func TestParseSearchResponse_EmptyHits(t *testing.T) {
	t.Run("空 hits 数组", func(t *testing.T) {
		// Arrange
		body := []byte(`{
			"hits": {
				"total": {"value": 0},
				"hits": []
			}
		}`)

		// Act
		result, err := parseSearchResponse(body, 1*time.Millisecond)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 0, result.RowCount, "空 hits 应有 0 行")
		assert.Empty(t, result.Columns, "空 hits 应没有列")
		assert.Empty(t, result.Rows, "空 hits 应没有行数据")
	})
}

func TestParseSearchResponse_InvalidJSON(t *testing.T) {
	t.Run("无效 JSON 回退为原始字符串", func(t *testing.T) {
		// Arrange
		body := []byte(`this is not json`)
		execTime := 2 * time.Millisecond

		// Act
		result, err := parseSearchResponse(body, execTime)

		// Assert
		require.NoError(t, err, "无效 JSON 不应返回错误，应回退")
		require.NotNil(t, result)
		assert.Equal(t, 1, result.RowCount, "应返回 1 行回退数据")
		assert.Equal(t, 1, len(result.Columns), "应有 1 列（response）")
		assert.Equal(t, "response", result.Columns[0].Name)
		assert.Equal(t, "json", result.Columns[0].Type)
		assert.Equal(t, "this is not json", result.Rows[0][0])
	})

	t.Run("空 body 回退", func(t *testing.T) {
		// Arrange
		body := []byte(``)

		// Act
		result, err := parseSearchResponse(body, 0)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 1, result.RowCount)
	})
}

func TestParseSearchResponse_DifferentSourceKeys(t *testing.T) {
	t.Run("不同 _source 键的记录", func(t *testing.T) {
		// Arrange — 第二条记录有额外的字段，第一条没有
		body := []byte(`{
			"hits": {
				"total": {"value": 2},
				"hits": [
					{"_source": {"name": "alice"}},
					{"_source": {"name": "bob", "extra": "field"}}
				]
			}
		}`)

		// Act
		result, err := parseSearchResponse(body, 0)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 2, result.RowCount)
		// 列基于第一条记录的 _source，所以只有 name
		assert.Equal(t, 1, len(result.Columns))
		assert.Equal(t, "name", result.Columns[0].Name)
	})
}

func TestParseSearchResponse_ExecutionTime(t *testing.T) {
	t.Run("执行时间正确传递", func(t *testing.T) {
		// Arrange
		body := []byte(`{
			"hits": {"total": {"value": 0}, "hits": []}
		}`)
		execTime := 42 * time.Millisecond

		// Act
		result, err := parseSearchResponse(body, execTime)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, execTime, result.ExecutionTime, "执行时间应精确传递")
	})
}

// --- 表格驱动测试：parseSearchResponse ---

func TestParseSearchResponse_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantRowCount int
		wantColCount int
		wantFallback bool
		wantExecTime time.Duration
	}{
		{
			name:         "多条记录",
			body:         `{"hits":{"total":{"value":3},"hits":[{"_source":{"a":1}},{"_source":{"a":2}},{"_source":{"a":3}}]}}`,
			wantRowCount: 3,
			wantColCount: 1,
			wantExecTime: 10 * time.Millisecond,
		},
		{
			name:         "空 hits",
			body:         `{"hits":{"total":{"value":0},"hits":[]}}`,
			wantRowCount: 0,
			wantColCount: 0,
			wantExecTime: 5 * time.Millisecond,
		},
		{
			name:         "无效 JSON 回退",
			body:         `{broken json`,
			wantRowCount: 1,
			wantColCount: 1,
			wantFallback: true,
			wantExecTime: 1 * time.Millisecond,
		},
		{
			name:         "空 body 回退",
			body:         ``,
			wantRowCount: 1,
			wantColCount: 1,
			wantFallback: true,
			wantExecTime: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSearchResponse([]byte(tt.body), tt.wantExecTime)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.wantRowCount, result.RowCount)
			assert.Equal(t, tt.wantColCount, len(result.Columns))
			assert.Equal(t, tt.wantExecTime, result.ExecutionTime)

			if tt.wantFallback {
				assert.Equal(t, "response", result.Columns[0].Name)
			}
		})
	}
}
