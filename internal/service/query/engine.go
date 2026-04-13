package query

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/pool"
	"MystiSql/internal/discovery"
	"MystiSql/internal/service/audit"
	"MystiSql/internal/service/masking"
	"MystiSql/internal/service/router"
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
	maskingService   *masking.MaskingService
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

func (e *Engine) SetMaskingService(service *masking.MaskingService) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.maskingService = service
}

func (e *Engine) execute(ctx context.Context, instanceName, query string, executor func(connection.Connection, context.Context, string) (interface{}, error), resultProcessor func(interface{}, *SQLParseResult) (interface{}, error)) (interface{}, error) {
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
	maskingSvc := e.maskingService
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

	targetInstance, err := e.resolveInstance(ctx, instanceName, query)
	if err != nil {
		return nil, err
	}
	pool, err := e.getConnectionPool(ctx, targetInstance)
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

	result, err := executor(conn, ctx, query)

	if auditSvc != nil {
		auditLog := audit.NewAuditLog(
			getUserIDFromContext(ctx),
			getClientIPFromContext(ctx),
			instanceName,
			"",
			query,
		)
		var affectedRows int64
		if result != nil {
			if qr, ok := result.(*types.QueryResult); ok {
				affectedRows = int64(qr.RowCount)
			} else if er, ok := result.(*types.ExecResult); ok {
				affectedRows = er.RowsAffected
			}
		}
		auditLog.SetQueryInfo(string(parseResult.StatementType), affectedRows, time.Since(startTime).Milliseconds())

		if err != nil {
			auditLog.SetError(err.Error())
		} else {
			auditLog.SetSuccess()
		}

		_ = auditSvc.Log(ctx, auditLog)
	}

	if err != nil {
		return nil, fmt.Errorf("执行语句失败: %w", err)
	}

	result, err = resultProcessor(result, parseResult)
	if err != nil {
		return nil, err
	}

	if maskingSvc != nil {
		if role := getRoleFromContext(ctx); role != "" {
			if qr, ok := result.(*types.QueryResult); ok {
				result = maskingSvc.MaskResult(role, qr)
			}
		}
	}

	return result, nil
}

func (e *Engine) ExecuteQuery(ctx context.Context, instanceName, query string) (*types.QueryResult, error) {
	result, err := e.execute(ctx, instanceName, query,
		func(conn connection.Connection, ctx context.Context, query string) (interface{}, error) {
			return conn.Query(ctx, query)
		},
		func(result interface{}, parseResult *SQLParseResult) (interface{}, error) {
			qr, ok := result.(*types.QueryResult)
			if !ok {
				return nil, fmt.Errorf("invalid query result type")
			}
			return WithResultSizeLimit(qr, parseResult.MaxResultSize), nil
		})

	if err != nil {
		return nil, err
	}
	qr, ok := result.(*types.QueryResult)
	if !ok {
		return nil, fmt.Errorf("invalid query result type")
	}
	return qr, nil
}

func (e *Engine) ExecuteExec(ctx context.Context, instanceName, query string) (*types.ExecResult, error) {
	result, err := e.execute(ctx, instanceName, query,
		func(conn connection.Connection, ctx context.Context, query string) (interface{}, error) {
			return conn.Exec(ctx, query)
		},
		func(result interface{}, parseResult *SQLParseResult) (interface{}, error) {
			er, ok := result.(*types.ExecResult)
			if !ok {
				return nil, fmt.Errorf("invalid exec result type")
			}
			return er, nil
		})

	if err != nil {
		return nil, err
	}
	er, ok := result.(*types.ExecResult)
	if !ok {
		return nil, fmt.Errorf("invalid exec result type")
	}
	return er, nil
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

// resolveInstance determines the target instance based on SQL type and instance roles.
// For write operations on a replica, it routes to the primary.
// For read operations on a primary, it routes to an available replica.
// When no role is configured or role is "readwrite", returns the requested instance unchanged.
func (e *Engine) resolveInstance(ctx context.Context, requestedInstance, query string) (string, error) {
	sqlType, inTxn, _ := router.ParseSQL(query)

	instance, err := e.registry.GetInstance(requestedInstance)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errors.ErrInstanceNotFound, requestedInstance)
	}

	if instance.Role == "" || instance.Role == "readwrite" {
		return requestedInstance, nil
	}

	if inTxn {
		if instance.Role == "replica" {
			if instance.ReplicaOf == "" {
				return "", fmt.Errorf("replica %s has no primary configured for transaction", requestedInstance)
			}
			return instance.ReplicaOf, nil
		}
		return requestedInstance, nil
	}

	if instance.Role == "replica" && sqlType.IsWrite() {
		if instance.ReplicaOf == "" {
			return "", fmt.Errorf("replica %s has no primary configured for write operation", requestedInstance)
		}
		return instance.ReplicaOf, nil
	}

	if instance.Role == "primary" && sqlType == router.SQLTypeSelect {
		replicas := e.findReplicas(requestedInstance)
		if len(replicas) > 0 {
			return replicas[rand.Intn(len(replicas))], nil
		}
	}

	return requestedInstance, nil
}

// findReplicas returns names of all replica instances whose ReplicaOf matches
// the given primary instance name.
func (e *Engine) findReplicas(primaryName string) []string {
	instances, err := e.registry.ListInstances()
	if err != nil {
		return nil
	}
	var replicas []string
	for _, inst := range instances {
		if inst.Role == "replica" && inst.ReplicaOf == primaryName {
			replicas = append(replicas, inst.Name)
		}
	}
	return replicas
}

func (e *Engine) GetPoolStats(ctx context.Context, instanceName string) (*connection.PoolStats, error) {
	e.mu.RLock()
	pool, exists := e.pools[instanceName]
	e.mu.RUnlock()

	if !exists {
		_, err := e.getConnectionPool(ctx, instanceName)
		if err != nil {
			return nil, err
		}

		e.mu.RLock()
		pool = e.pools[instanceName]
		e.mu.RUnlock()
	}

	return pool.GetStats(), nil
}
