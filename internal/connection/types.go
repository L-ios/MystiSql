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

// ConnectionPool 定义连接池接口
type ConnectionPool interface {
	// GetConnection 从连接池中获取一个连接
	GetConnection(ctx context.Context) (Connection, error)

	// ReturnConnection 将连接归还到连接池
	ReturnConnection(conn Connection)

	// Close 关闭连接池并释放所有连接
	Close() error

	// GetStats 获取连接池统计信息
	GetStats() *PoolStats

	// SetMaxConnections 设置最大连接数
	SetMaxConnections(max int)

	// SetMinConnections 设置最小连接数
	SetMinConnections(min int)

	// SetMaxIdleTime 设置连接最大空闲时间
	SetMaxIdleTime(duration string)

	// SetMaxLifetime 设置连接最大生命周期
	SetMaxLifetime(duration string)
}

// PoolStats 定义连接池统计信息
type PoolStats struct {
	TotalConnections  int
	IdleConnections   int
	ActiveConnections int
	MaxConnections    int
	MinConnections    int
	AcquireCount      int64
	AcquireFailed     int64
	ReleaseCount      int64

	WaitCount          int64
	WaitDuration       int64
	MaxWaitDuration    int64
	AvgWaitDuration    int64
	AcquireDuration    int64
	MaxAcquireDuration int64
	AvgAcquireDuration int64

	QueryCount    int64
	QueryFailed   int64
	QueryDuration int64

	ExecCount    int64
	ExecFailed   int64
	ExecDuration int64

	HealthCheckCount  int64
	HealthCheckFailed int64

	ConnectionsCreated int64
	ConnectionsClosed  int64

	LastAcquireTime int64
	LastReleaseTime int64
	LastErrorTime   int64
	LastErrorMsg    string
}

type MetricsCollector interface {
	RecordAcquire(instance string, duration int64, success bool)
	RecordRelease(instance string, duration int64)
	RecordWait(instance string, duration int64)
	RecordQuery(instance string, duration int64, success bool)
	RecordExec(instance string, duration int64, success bool)
	RecordHealthCheck(instance string, success bool)
	RecordConnectionCreated(instance string)
	RecordConnectionClosed(instance string)
	UpdatePoolStats(instance string, stats *PoolStats)
	GetAllMetrics() map[string]*PoolStats
}

type MetricsEvent struct {
	Type      string
	Instance  string
	Timestamp int64
	Duration  int64
	Success   bool
	Error     string
	Stats     *PoolStats
}

// PoolConfig 定义连接池配置
type PoolConfig struct {
	// MaxConnections 最大连接数
	MaxConnections int

	// MinConnections 最小连接数
	MinConnections int

	// MaxIdleTime 连接最大空闲时间
	MaxIdleTime string

	// MaxLifetime 连接最大生命周期
	MaxLifetime string

	// ConnectionTimeout 连接超时时间
	ConnectionTimeout string

	// PingInterval 健康检查间隔
	PingInterval string
}
