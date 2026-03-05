package connection

import (
	"context"

	"MystiSql/pkg/types"
)

// Connection 定义数据库连接接口
// 所有数据库连接实现（MySQL、PostgreSQL、Oracle、Redis 等）都必须实现此接口
type Connection interface {
	// Connect 建立到数据库的连接
	// ctx 用于取消操作和超时控制
	Connect(ctx context.Context) error

	// Query 执行查询语句（SELECT）
	// 返回查询结果集
	Query(ctx context.Context, sql string) (*types.QueryResult, error)

	// Exec 执行非查询语句（INSERT、UPDATE、DELETE）
	// 返回受影响的行数和最后插入的ID
	Exec(ctx context.Context, sql string) (*types.ExecResult, error)

	// Ping 检查连接是否仍然存活
	Ping(ctx context.Context) error

	// Close 关闭连接并释放资源
	Close() error
}

// ConnectionFactory 定义连接工厂接口
type ConnectionFactory interface {
	// CreateConnection 创建一个新的数据库连接
	CreateConnection(instance *types.DatabaseInstance) (Connection, error)
}
