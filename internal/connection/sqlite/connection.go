package sqlite

import (
	"context"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/base"
	"MystiSql/pkg/types"

	_ "modernc.org/sqlite"
)

// Connection SQLite 连接实现
type Connection struct {
	*base.SQLConnection
}

// Factory SQLite 连接工厂
type Factory struct{}

// NewFactory 创建一个新的 SQLite 连接工厂
func NewFactory() connection.ConnectionFactory {
	return &Factory{}
}

// CreateConnection 创建一个新的 SQLite 连接
func (f *Factory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	return NewConnection(instance), nil
}

// NewConnection 创建一个新的 SQLite 连接
func NewConnection(instance *types.DatabaseInstance) connection.Connection {
	// SQLite 使用单连接池（不支持并发写入）
	poolConfig := base.PoolConfig{
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 30 * time.Minute,
	}
	return &Connection{
		SQLConnection: base.NewSQLConnection(instance, poolConfig),
	}
}

// Connect 建立到 SQLite 数据库的连接
func (c *Connection) Connect(ctx context.Context) error {
	dsn := c.buildDSN()
	return c.InitDB(ctx, "sqlite", dsn)
}

// buildDSN 构建 SQLite 连接字符串
func (c *Connection) buildDSN() string {
	inst := c.Instance()
	if inst.Database == "" || inst.Database == ":memory:" {
		return ":memory:"
	}
	return inst.Database
}
