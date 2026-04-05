package oracle

import (
	"context"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/base"
	"MystiSql/pkg/types"

	go_ora "github.com/sijms/go-ora/v2"
)

// Connection Oracle 连接实现
type Connection struct {
	*base.SQLConnection
}

// Factory Oracle 连接工厂
type Factory struct{}

// NewFactory 创建一个新的 Oracle 连接工厂
func NewFactory() connection.ConnectionFactory {
	return &Factory{}
}

// CreateConnection 创建一个新的 Oracle 连接
func (f *Factory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	return NewConnection(instance), nil
}

// NewConnection 创建一个新的 Oracle 连接
func NewConnection(instance *types.DatabaseInstance) connection.Connection {
	return &Connection{
		SQLConnection: base.NewSQLConnection(instance, base.DefaultPoolConfig()),
	}
}

// Connect 建立到 Oracle 数据库的连接
func (c *Connection) Connect(ctx context.Context) error {
	inst := c.Instance()
	uri := go_ora.BuildUrl(inst.Host, inst.Port, inst.Database, inst.Username, inst.Password, nil)
	return c.InitDB(ctx, "oracle", uri)
}
