package mssql

import (
	"context"
	"fmt"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/base"
	"MystiSql/pkg/types"

	_ "github.com/microsoft/go-mssqldb"
)

// Connection MSSQL 连接实现
type Connection struct {
	*base.SQLConnection
}

// Factory MSSQL 连接工厂
type Factory struct{}

// NewFactory 创建一个新的 MSSQL 连接工厂
func NewFactory() connection.ConnectionFactory {
	return &Factory{}
}

// CreateConnection 创建一个新的 MSSQL 连接
func (f *Factory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	return NewConnection(instance), nil
}

// NewConnection 创建一个新的 MSSQL 连接
func NewConnection(instance *types.DatabaseInstance) connection.Connection {
	return &Connection{
		SQLConnection: base.NewSQLConnection(instance, base.DefaultPoolConfig()),
	}
}

// Connect 建立到 MSSQL 数据库的连接
func (c *Connection) Connect(ctx context.Context) error {
	dsn := c.buildDSN()
	return c.InitDB(ctx, "sqlserver", dsn)
}

// buildDSN 构建 MSSQL 连接字符串
// 格式：sqlserver://username:password@host:port?database=dbname
func (c *Connection) buildDSN() string {
	inst := c.Instance()
	if inst.Database != "" {
		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			inst.Username, inst.Password, inst.Host, inst.Port, inst.Database)
	}
	return fmt.Sprintf("sqlserver://%s:%s@%s:%d",
		inst.Username, inst.Password, inst.Host, inst.Port)
}
