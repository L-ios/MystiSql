package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	go_ora "github.com/sijms/go-ora/v2"
)

type Connection struct {
	instance *types.DatabaseInstance
	db       *sql.DB
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
		instance: instance,
	}
}

func (c *Connection) Connect(ctx context.Context) error {
	uri := go_ora.BuildUrl(c.instance.Host, c.instance.Port, c.instance.Database, c.instance.Username, c.instance.Password, nil)

	db, err := sql.Open("oracle", uri)
	if err != nil {
		return fmt.Errorf("%w: open connection failed: %v", errors.ErrConnectionFailed, err)
	}

	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return fmt.Errorf("%w: ping failed: %v (close also failed: %v)", errors.ErrConnectionFailed, err, closeErr)
		}
		return fmt.Errorf("%w: ping failed: %v", errors.ErrConnectionFailed, err)
	}

	c.db = db
	c.instance.SetStatus(types.InstanceStatusHealthy)
	return nil
}

func (c *Connection) Query(ctx context.Context, query string) (*types.QueryResult, error) {
	if c.db == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrQueryFailed, err)
	}
	defer rows.Close()

	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("%w: get columns failed: %v", errors.ErrQueryFailed, err)
	}

	columnInfos := make([]types.ColumnInfo, len(columns))
	for i, col := range columns {
		columnInfos[i] = types.ColumnInfo{Name: col.Name(), Type: col.DatabaseTypeName()}
	}

	var resultRows []types.Row
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("%w: scan row failed: %v", errors.ErrQueryFailed, err)
		}

		row := make(types.Row, len(values))
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = v
			}
		}
		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: read rows failed: %v", errors.ErrQueryFailed, err)
	}

	return types.NewQueryResult(columnInfos, resultRows, time.Since(start)), nil
}

func (c *Connection) Exec(ctx context.Context, query string) (*types.ExecResult, error) {
	if c.db == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	result, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: exec failed: %v", errors.ErrQueryFailed, err)
	}

	rowsAffected, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()

	return types.NewExecResult(rowsAffected, lastInsertID, time.Since(start)), nil
}

func (c *Connection) Ping(ctx context.Context) error {
	if c.db == nil {
		return errors.ErrConnectionClosed
	}

	if err := c.db.PingContext(ctx); err != nil {
		c.instance.SetStatus(types.InstanceStatusUnhealthy)
		return fmt.Errorf("%w: ping failed: %v", errors.ErrConnectionFailed, err)
	}

	c.instance.SetStatus(types.InstanceStatusHealthy)
	return nil
}

func (c *Connection) Close() error {
	if c.db == nil {
		return nil
	}

	err := c.db.Close()
	c.db = nil
	c.instance.SetStatus(types.InstanceStatusUnknown)

	if err != nil {
		return fmt.Errorf("close connection failed: %v", err)
	}

	return nil
}
