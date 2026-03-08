package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Connection struct {
	instance *types.DatabaseInstance
	pool     *pgxpool.Pool
	logger   *zap.Logger
}

type Factory struct{}

func NewFactory() connection.ConnectionFactory {
	return &Factory{}
}

func (f *Factory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	return NewConnection(instance, nil)
}

func NewConnection(instance *types.DatabaseInstance, logger *zap.Logger) (connection.Connection, error) {
	conn := &Connection{
		instance: instance,
		logger:   logger,
	}
	return conn, nil
}

func (c *Connection) Connect(ctx context.Context) error {
	if c.pool != nil {
		return errors.ErrConnectionClosed
	}

	dsn, err := c.buildDSN(c.instance)
	if err != nil {
		return fmt.Errorf("failed to build DSN: %w", err)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse pool config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", errors.ErrConnectionFailed, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to validate connection: %w", errors.ErrConnectionFailed, err)
	}

	c.pool = pool
	c.instance.SetStatus(types.InstanceStatusHealthy)
	return nil
}

func (c *Connection) Query(ctx context.Context, query string) (*types.QueryResult, error) {
	if c.pool == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]types.ColumnInfo, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = types.ColumnInfo{
			Name: fd.Name,
			Type: string(fd.DataType.Name),
		}
	}

	var allRows []types.Row
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to scan row values: %w", err)
		}

		row := make(types.Row)
		for _, value := range values {
			row = append(row, value)
		}
		allRows = append(allRows, row)
	}

	return &types.QueryResult{
		Columns:       columns,
		Rows:          allRows,
		RowCount:      len(allRows),
		ExecutionTime: time.Since(start),
	}, nil
}

func (c *Connection) Exec(ctx context.Context, query string) (*types.ExecResult, error) {
	if c.pool == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	result, err := c.pool.Exec(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("exec failed: %w", err)
	}

	return &types.ExecResult{
		RowsAffected:  result.RowsAffected(),
		LastInsertID:  0,
		ExecutionTime: time.Since(start),
	}, nil
}

func (c *Connection) Ping(ctx context.Context) error {
	if c.pool == nil {
		return errors.ErrConnectionClosed
	}
	return c.pool.Ping(ctx)
}

func (c *Connection) Close() error {
	if c.pool == nil {
		return nil
	}
	c.pool.Close()
	c.pool = nil
	c.instance.SetStatus(types.InstanceStatusUnknown)
	return nil
}

func (c *Connection) buildDSN(instance *types.DatabaseInstance) (string, error) {
	var parts []string

	parts = append(parts, fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		instance.Username,
		instance.Password,
		instance.Host,
		instance.Port,
		instance.Database,
	))

	if sslMode, ok := instance.Params["sslmode"]; ok {
		parts = append(parts, fmt.Sprintf("?sslmode=%s", sslMode))
	}

	if connectTimeout, ok := instance.Params["connect_timeout"]; ok {
		parts = append(parts, fmt.Sprintf("&connect_timeout=%s", connectTimeout))
	}

	return strings.Join(parts, ""), nil
}
