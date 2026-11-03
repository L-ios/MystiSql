package clickhouse

import (
	"context"
	"fmt"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/base"
	"MystiSql/pkg/types"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

// Connection 基于 base.SQLConnection 的 ClickHouse 连接实现，
// 仅需实现 Connect（DSN 构造），Query/Exec/Ping/Close 复用 base 层。
type Connection struct {
	*base.SQLConnection
}

type Factory struct{}

func NewFactory() connection.ConnectionFactory {
	return &Factory{}
}

func (f *Factory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	return NewConnection(instance), nil
}

func NewConnection(instance *types.DatabaseInstance) connection.Connection {
	return &Connection{
		SQLConnection: base.NewSQLConnection(instance, base.DefaultPoolConfig()),
	}
}

func (c *Connection) Connect(ctx context.Context) error {
	return c.InitDB(ctx, "clickhouse", c.buildDSN())
}

func (c *Connection) buildDSN() string {
	instance := c.Instance()
	if instance.Database != "" {
		return fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s",
			instance.Username, instance.Password, instance.Host, instance.Port, instance.Database)
	}
	return fmt.Sprintf("clickhouse://%s:%s@%s:%d",
		instance.Username, instance.Password, instance.Host, instance.Port)
}
