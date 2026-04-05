package base

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"
)

// PoolConfig holds connection pool configuration.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DefaultPoolConfig returns the default pool configuration.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
	}
}

// SQLConnection provides a shared implementation for database/sql-based drivers.
// Drivers embed this struct and only need to implement Connect() (DSN construction)
// and provide the driver name.
type SQLConnection struct {
	instance   *types.DatabaseInstance
	db         *sql.DB
	poolConfig PoolConfig
}

// NewSQLConnection creates a new base SQLConnection.
func NewSQLConnection(instance *types.DatabaseInstance, poolConfig PoolConfig) *SQLConnection {
	return &SQLConnection{
		instance:   instance,
		poolConfig: poolConfig,
	}
}

// InitDB opens a database connection with the given driver name and DSN,
// configures the connection pool, and verifies connectivity.
func (c *SQLConnection) InitDB(ctx context.Context, driverName, dsn string) error {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("%w: open connection failed: %v", errors.ErrConnectionFailed, err)
	}

	db.SetConnMaxLifetime(c.poolConfig.ConnMaxLifetime)
	db.SetMaxOpenConns(c.poolConfig.MaxOpenConns)
	db.SetMaxIdleConns(c.poolConfig.MaxIdleConns)

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

// DB returns the underlying *sql.DB (for drivers that need direct access).
func (c *SQLConnection) DB() *sql.DB {
	return c.db
}

// SetDB sets the underlying *sql.DB (used by Connect implementations).
func (c *SQLConnection) SetDB(db *sql.DB) {
	c.db = db
}

// Instance returns the associated DatabaseInstance.
func (c *SQLConnection) Instance() *types.DatabaseInstance {
	return c.instance
}

// Query executes a query and returns the result.
// Uses ColumnTypes() for better type information (works for all database/sql drivers).
func (c *SQLConnection) Query(ctx context.Context, query string) (*types.QueryResult, error) {
	if c.db == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrQueryFailed, err)
	}
	defer rows.Close()

	// Use ColumnTypes for better type information
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		// Fallback to Columns() if ColumnTypes() is not supported
		columns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("%w: get columns failed: %v", errors.ErrQueryFailed, err)
		}
		columnInfos := make([]types.ColumnInfo, len(columns))
		for i, col := range columns {
			columnInfos[i] = types.ColumnInfo{Name: col, Type: "unknown"}
		}
		return c.scanRows(rows, columnInfos, start)
	}

	columnInfos := make([]types.ColumnInfo, len(colTypes))
	for i, col := range colTypes {
		typeName := col.DatabaseTypeName()
		if typeName == "" {
			typeName = "unknown"
		}
		columnInfos[i] = types.ColumnInfo{Name: col.Name(), Type: typeName}
	}

	return c.scanRows(rows, columnInfos, start)
}

// scanRows reads all rows from the result set and constructs a QueryResult.
func (c *SQLConnection) scanRows(rows *sql.Rows, columnInfos []types.ColumnInfo, start time.Time) (*types.QueryResult, error) {
	var resultRows []types.Row
	for rows.Next() {
		values := make([]interface{}, len(columnInfos))
		valuePtrs := make([]interface{}, len(columnInfos))
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

// Exec executes a non-query statement.
func (c *SQLConnection) Exec(ctx context.Context, query string) (*types.ExecResult, error) {
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

// Ping checks the connection health.
func (c *SQLConnection) Ping(ctx context.Context) error {
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

// Close closes the connection.
func (c *SQLConnection) Close() error {
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
