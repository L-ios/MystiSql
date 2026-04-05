package mysql

import (
	"context"
	"fmt"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/base"
	"MystiSql/pkg/types"

	_ "github.com/go-sql-driver/mysql"
)

// Connection MySQL 连接实现
type Connection struct {
	*base.SQLConnection
}

// Factory MySQL 连接工厂
type Factory struct{}

// NewFactory 创建一个新的 MySQL 连接工厂
func NewFactory() connection.ConnectionFactory {
	return &Factory{}
}

// CreateConnection 创建一个新的 MySQL 连接
func (f *Factory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	return NewConnection(instance), nil
}

// NewConnection 创建一个新的 MySQL 连接
func NewConnection(instance *types.DatabaseInstance) connection.Connection {
	return &Connection{
		SQLConnection: base.NewSQLConnection(instance, base.DefaultPoolConfig()),
	}
}

// Connect 建立到 MySQL 数据库的连接
func (c *Connection) Connect(ctx context.Context) error {
	dsn := buildDSN(c.Instance())
	return c.InitDB(ctx, "mysql", dsn)
}

// buildDSN 构建MySQL连接字符串
// 格式：username:password@tcp(host:port)/database?params
func buildDSN(instance *types.DatabaseInstance) string {
	dsn := ""

	if instance.Username != "" {
		dsn += instance.Username
		if instance.Password != "" {
			dsn += ":" + instance.Password
		}
		dsn += "@"
	}

	dsn += fmt.Sprintf("tcp(%s:%d)", instance.Host, instance.Port)

	if instance.Database != "" {
		dsn += "/" + instance.Database
	} else {
		dsn += "/"
	}

	params := "?parseTime=true&loc=Local&charset=utf8mb4&timeout=30s"
	dsn += params

	return dsn
}
