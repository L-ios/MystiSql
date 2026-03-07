package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"

	_ "github.com/go-sql-driver/mysql"
)

// Connection MySQL 连接实现
type Connection struct {
	instance *types.DatabaseInstance
	db       *sql.DB
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
		instance: instance,
	}
}

// Connect 建立到 MySQL 数据库的连接
func (c *Connection) Connect(ctx context.Context) error {
	// 构建连接字符串
	dsn := buildDSN(c.instance)

	// 打开连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("%w: 打开连接失败: %v", errors.ErrConnectionFailed, err)
	}

	// 设置连接参数
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// 验证连接
	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return fmt.Errorf("%w: 连接验证失败: %v (关闭连接也失败: %v)", errors.ErrConnectionFailed, err, closeErr)
		}
		return fmt.Errorf("%w: 连接验证失败: %v", errors.ErrConnectionFailed, err)
	}

	c.db = db
	c.instance.SetStatus(types.InstanceStatusHealthy)
	return nil
}

// Query 执行查询语句
func (c *Connection) Query(ctx context.Context, query string) (*types.QueryResult, error) {
	if c.db == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	// 执行查询
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrQueryFailed, err)
	}
	defer func() {
		_ = rows.Close() // 忽略关闭错误，不影响查询结果
	}()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("%w: 获取列信息失败: %v", errors.ErrQueryFailed, err)
	}

	// 构建列信息
	columnInfos := make([]types.ColumnInfo, len(columns))
	for i, col := range columns {
		columnInfos[i] = types.ColumnInfo{
			Name: col,
			Type: "unknown", // MySQL 驱动不提供准确的类型信息
		}
	}

	// 读取所有行
	var resultRows []types.Row
	for rows.Next() {
		// 创建扫描目标
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行数据
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("%w: 扫描行数据失败: %v", errors.ErrQueryFailed, err)
		}

		// 转换为 Row 类型
		row := make(types.Row, len(values))
		for i, v := range values {
			// 将字节数组转换为字符串
			if b, ok := v.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = v
			}
		}

		resultRows = append(resultRows, row)
	}

	// 检查是否有行读取错误
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: 读取行数据失败: %v", errors.ErrQueryFailed, err)
	}

	execTime := time.Since(start)

	return types.NewQueryResult(columnInfos, resultRows, execTime), nil
}

// Exec 执行非查询语句
func (c *Connection) Exec(ctx context.Context, query string) (*types.ExecResult, error) {
	if c.db == nil {
		return nil, errors.ErrConnectionClosed
	}

	start := time.Now()

	// 执行语句
	result, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: 执行语句失败: %v", errors.ErrQueryFailed, err)
	}

	// 获取受影响的行数
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		rowsAffected = 0 // 某些驱动不支持
	}

	// 获取最后插入的 ID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		lastInsertID = 0 // 某些驱动不支持
	}

	execTime := time.Since(start)

	return types.NewExecResult(rowsAffected, lastInsertID, execTime), nil
}

// Ping 检查连接健康状态
func (c *Connection) Ping(ctx context.Context) error {
	if c.db == nil {
		return errors.ErrConnectionClosed
	}

	if err := c.db.PingContext(ctx); err != nil {
		c.instance.SetStatus(types.InstanceStatusUnhealthy)
		return fmt.Errorf("%w: ping 失败: %v", errors.ErrConnectionFailed, err)
	}

	c.instance.SetStatus(types.InstanceStatusHealthy)
	return nil
}

// Close 关闭连接
func (c *Connection) Close() error {
	if c.db == nil {
		return nil // 幂等操作
	}

	err := c.db.Close()
	c.db = nil
	c.instance.SetStatus(types.InstanceStatusUnknown)

	if err != nil {
		return fmt.Errorf("关闭连接失败: %v", err)
	}

	return nil
}

// buildDSN 构建MySQL连接字符串
// 格式：username:password@tcp(host:port)/database?params
func buildDSN(instance *types.DatabaseInstance) string {
	dsn := ""

	// 添加用户名和密码
	if instance.Username != "" {
		dsn += instance.Username
		if instance.Password != "" {
			dsn += ":" + instance.Password
		}
		dsn += "@"
	}

	// 添加网络协议和地址
	dsn += fmt.Sprintf("tcp(%s:%d)", instance.Host, instance.Port)

	// 添加数据库名
	if instance.Database != "" {
		dsn += "/" + instance.Database
	} else {
		dsn += "/"
	}

	// 添加连接参数
	params := "?parseTime=true&loc=Local&charset=utf8mb4&timeout=30s"
	dsn += params

	return dsn
}
