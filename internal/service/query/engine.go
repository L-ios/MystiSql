package query

import (
	"context"
	"fmt"
	"sync"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/mysql"
	"MystiSql/internal/connection/pool"
	"MystiSql/internal/discovery"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"
)

type Engine struct {
	registry  discovery.InstanceRegistry
	parser    SQLParser
	pools     map[string]connection.ConnectionPool
	factories map[types.DatabaseType]connection.ConnectionFactory
	mu        sync.RWMutex
}

func NewEngine(registry discovery.InstanceRegistry) *Engine {
	// 初始化连接工厂
	factories := make(map[types.DatabaseType]connection.ConnectionFactory)
	factories[types.DatabaseTypeMySQL] = mysql.NewFactory()

	return &Engine{
		registry:  registry,
		parser:    NewParser(),
		pools:     make(map[string]connection.ConnectionPool),
		factories: factories,
	}
}

func (e *Engine) ExecuteQuery(ctx context.Context, instanceName, query string) (*types.QueryResult, error) {
	// 解析 SQL 语句
	parseResult, err := e.parser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("解析 SQL 语句失败: %w", err)
	}

	// 验证 SQL 语句
	if err := e.parser.Validate(query); err != nil {
		return nil, fmt.Errorf("验证 SQL 语句失败: %w", err)
	}

	// 获取连接池
	pool, err := e.getConnectionPool(ctx, instanceName)
	if err != nil {
		return nil, err
	}

	// 添加查询超时
	ctx, cancel := WithTimeout(ctx, parseResult.QueryTimeout)
	defer cancel()

	// 获取连接
	conn, err := pool.GetConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}
	defer conn.Close()

	// 执行查询
	result, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("执行查询失败: %w", err)
	}

	// 限制结果集大小
	result = WithResultSizeLimit(result, parseResult.MaxResultSize)

	return result, nil
}

func (e *Engine) ExecuteExec(ctx context.Context, instanceName, query string) (*types.ExecResult, error) {
	// 解析 SQL 语句
	parseResult, err := e.parser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("解析 SQL 语句失败: %w", err)
	}

	// 验证 SQL 语句
	if err := e.parser.Validate(query); err != nil {
		return nil, fmt.Errorf("验证 SQL 语句失败: %w", err)
	}

	// 获取连接池
	pool, err := e.getConnectionPool(ctx, instanceName)
	if err != nil {
		return nil, err
	}

	// 添加查询超时
	ctx, cancel := WithTimeout(ctx, parseResult.QueryTimeout)
	defer cancel()

	// 获取连接
	conn, err := pool.GetConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}
	defer conn.Close()

	// 执行语句
	result, err := conn.Exec(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("执行语句失败: %w", err)
	}

	return result, nil
}

func (e *Engine) PingInstance(ctx context.Context, instanceName string) error {
	// 获取连接池
	pool, err := e.getConnectionPool(ctx, instanceName)
	if err != nil {
		return err
	}

	// 获取连接
	conn, err := pool.GetConnection(ctx)
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}
	defer conn.Close()

	// 执行 ping
	return conn.Ping(ctx)
}

func (e *Engine) getConnectionPool(ctx context.Context, instanceName string) (connection.ConnectionPool, error) {
	e.mu.RLock()
	if p, exists := e.pools[instanceName]; exists {
		e.mu.RUnlock()
		return p, nil
	}
	e.mu.RUnlock()

	// 获取数据库实例
	instance, err := e.registry.GetInstance(instanceName)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrInstanceNotFound, instanceName)
	}

	// 获取连接工厂
	factory, exists := e.factories[instance.Type]
	if !exists {
		return nil, fmt.Errorf("不支持的数据库类型: %s", instance.Type)
	}

	// 创建连接池配置
	config := &connection.PoolConfig{
		MaxConnections:    10,
		MinConnections:    2,
		MaxIdleTime:       "30s",
		MaxLifetime:       "1h",
		ConnectionTimeout: "10s",
		PingInterval:      "30s",
	}

	// 创建连接池
	newPool, err := pool.NewConnectionPool(instance, factory, config)
	if err != nil {
		return nil, fmt.Errorf("创建连接池失败: %w", err)
	}

	// 缓存连接池
	e.mu.Lock()
	e.pools[instanceName] = newPool
	e.mu.Unlock()

	return newPool, nil
}

func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var errs []error
	for name, pool := range e.pools {
		if err := pool.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭连接池 %s 失败: %v", name, err))
		}
	}
	e.pools = make(map[string]connection.ConnectionPool)

	if len(errs) > 0 {
		return fmt.Errorf("关闭连接池时发生错误: %v", errs)
	}

	return nil
}

func (e *Engine) ListInstances() ([]*types.DatabaseInstance, error) {
	return e.registry.ListInstances()
}

func (e *Engine) GetInstanceHealth(ctx context.Context, instanceName string) (types.InstanceStatus, error) {
	instance, err := e.registry.GetInstance(instanceName)
	if err != nil {
		return types.InstanceStatusUnknown, err
	}

	// 获取连接池
	pool, err := e.getConnectionPool(ctx, instanceName)
	if err != nil {
		instance.SetStatus(types.InstanceStatusUnhealthy)
		return types.InstanceStatusUnhealthy, nil
	}

	// 获取连接
	conn, err := pool.GetConnection(ctx)
	if err != nil {
		instance.SetStatus(types.InstanceStatusUnhealthy)
		return types.InstanceStatusUnhealthy, nil
	}
	defer conn.Close()

	// 执行 ping
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		instance.SetStatus(types.InstanceStatusUnhealthy)
		return types.InstanceStatusUnhealthy, nil
	}

	instance.SetStatus(types.InstanceStatusHealthy)
	return types.InstanceStatusHealthy, nil
}

func (e *Engine) GetParser() SQLParser {
	return e.parser
}

func (e *Engine) GetPoolStats(instanceName string) (*connection.PoolStats, error) {
	e.mu.RLock()
	pool, exists := e.pools[instanceName]
	e.mu.RUnlock()

	if !exists {
		_, err := e.getConnectionPool(context.Background(), instanceName)
		if err != nil {
			return nil, err
		}

		e.mu.RLock()
		pool = e.pools[instanceName]
		e.mu.RUnlock()
	}

	return pool.GetStats(), nil
}
