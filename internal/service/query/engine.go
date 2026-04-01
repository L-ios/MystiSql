package query

import (
	"context"
	"fmt"
	"sync"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/pool"
	"MystiSql/internal/discovery"
	"MystiSql/internal/service/audit"
	"MystiSql/internal/service/validator"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"
)

type Engine struct {
	registry         discovery.InstanceRegistry
	parser           SQLParser
	pools            map[string]connection.ConnectionPool
	driverRegistry   *connection.DriverRegistry
	auditService     *audit.AuditService
	validatorService *validator.ValidatorService
	mu               sync.RWMutex
}

func NewEngine(registry discovery.InstanceRegistry, driverReg *connection.DriverRegistry) *Engine {
	return &Engine{
		registry:       registry,
		parser:         NewParser(),
		pools:          make(map[string]connection.ConnectionPool),
		driverRegistry: driverReg,
	}
}

func (e *Engine) SetAuditService(service *audit.AuditService) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.auditService = service
}

func (e *Engine) SetValidatorService(service *validator.ValidatorService) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.validatorService = service
}

func (e *Engine) ExecuteQuery(ctx context.Context, instanceName, query string) (*types.QueryResult, error) {
	startTime := time.Now()

	parseResult, err := e.parser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("解析 SQL 语句失败: %w", err)
	}

	if err := e.parser.Validate(query); err != nil {
		return nil, fmt.Errorf("验证 SQL 语句失败: %w", err)
	}

	e.mu.RLock()
	validatorSvc := e.validatorService
	auditSvc := e.auditService
	e.mu.RUnlock()

	if validatorSvc != nil {
		validationResult, err := validatorSvc.Validate(ctx, instanceName, query)
		if err != nil {
			return nil, fmt.Errorf("SQL 验证失败: %w", err)
		}
		if !validationResult.Allowed {
			return nil, fmt.Errorf("SQL 被拦截: %s", validationResult.Reason)
		}
	}

	pool, err := e.getConnectionPool(ctx, instanceName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := WithTimeout(ctx, parseResult.QueryTimeout)
	defer cancel()

	conn, err := pool.GetConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}
	defer conn.Close()

	result, err := conn.Query(ctx, query)

	if auditSvc != nil {
		auditLog := audit.NewAuditLog(
			getUserIDFromContext(ctx),
			getClientIPFromContext(ctx),
			instanceName,
			"",
			query,
		)
		var rowCount int64
		if result != nil {
			rowCount = int64(result.RowCount)
		}
		auditLog.SetQueryInfo(string(parseResult.StatementType), rowCount, time.Since(startTime).Milliseconds())

		if err != nil {
			auditLog.SetError(err.Error())
		} else {
			auditLog.SetSuccess()
		}

		if err := auditSvc.Log(ctx, auditLog); err != nil {
			// Log the audit failure but don't fail the query
		}
	}

	if err != nil {
		return nil, fmt.Errorf("执行查询失败: %w", err)
	}

	result = WithResultSizeLimit(result, parseResult.MaxResultSize)

	return result, nil
}

func (e *Engine) ExecuteExec(ctx context.Context, instanceName, query string) (*types.ExecResult, error) {
	parseResult, err := e.parser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("解析 SQL 语句失败: %w", err)
	}

	if err := e.parser.Validate(query); err != nil {
		return nil, fmt.Errorf("验证 SQL 语句失败: %w", err)
	}

	e.mu.RLock()
	validatorSvc := e.validatorService
	auditSvc := e.auditService
	e.mu.RUnlock()

	if validatorSvc != nil {
		validationResult, err := validatorSvc.Validate(ctx, instanceName, query)
		if err != nil {
			return nil, fmt.Errorf("SQL 验证失败: %w", err)
		}
		if !validationResult.Allowed {
			return nil, fmt.Errorf("SQL 被拦截: %s", validationResult.Reason)
		}
	}

	pool, err := e.getConnectionPool(ctx, instanceName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := WithTimeout(ctx, parseResult.QueryTimeout)
	defer cancel()

	conn, err := pool.GetConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}
	defer conn.Close()

	startTime := time.Now()
	result, err := conn.Exec(ctx, query)

	if auditSvc != nil {
		auditLog := audit.NewAuditLog(
			getUserIDFromContext(ctx),
			getClientIPFromContext(ctx),
			instanceName,
			"",
			query,
		)
		var rowsAffected int64
		if result != nil {
			rowsAffected = result.RowsAffected
		}
		auditLog.SetQueryInfo(string(parseResult.StatementType), rowsAffected, time.Since(startTime).Milliseconds())

		if err != nil {
			auditLog.SetError(err.Error())
		} else {
			auditLog.SetSuccess()
		}

		if logErr := auditSvc.Log(ctx, auditLog); logErr != nil {
			// 审计日志写入失败不影响查询结果
		}
	}

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
	factory, err := e.driverRegistry.GetFactory(instance.Type)
	if err != nil {
		return nil, fmt.Errorf("不支持的数据库类型: %s: %w", instance.Type, err)
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
