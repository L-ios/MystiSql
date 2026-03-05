package query

import (
	"context"
	"fmt"
	"sync"
	"time"

	"MystiSql/internal/connection"
	"MystiSql/internal/connection/mysql"
	"MystiSql/internal/discovery"
	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"
)

type Engine struct {
	registry    discovery.InstanceRegistry
	connections map[string]connection.Connection
	mu          sync.RWMutex
}

func NewEngine(registry discovery.InstanceRegistry) *Engine {
	return &Engine{
		registry:    registry,
		connections: make(map[string]connection.Connection),
	}
}

func (e *Engine) ExecuteQuery(ctx context.Context, instanceName, query string) (*types.QueryResult, error) {
	conn, err := e.getConnection(ctx, instanceName)
	if err != nil {
		return nil, err
	}

	result, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("执行查询失败: %w", err)
	}

	return result, nil
}

func (e *Engine) ExecuteExec(ctx context.Context, instanceName, query string) (*types.ExecResult, error) {
	conn, err := e.getConnection(ctx, instanceName)
	if err != nil {
		return nil, err
	}

	result, err := conn.Exec(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("执行语句失败: %w", err)
	}

	return result, nil
}

func (e *Engine) PingInstance(ctx context.Context, instanceName string) error {
	conn, err := e.getConnection(ctx, instanceName)
	if err != nil {
		return err
	}

	return conn.Ping(ctx)
}

func (e *Engine) getConnection(ctx context.Context, instanceName string) (connection.Connection, error) {
	e.mu.RLock()
	if conn, exists := e.connections[instanceName]; exists {
		e.mu.RUnlock()

		if err := conn.Ping(ctx); err == nil {
			return conn, nil
		}

		e.mu.Lock()
		delete(e.connections, instanceName)
		e.mu.Unlock()
	} else {
		e.mu.RUnlock()
	}

	instance, err := e.registry.GetInstance(instanceName)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrInstanceNotFound, instanceName)
	}

	switch instance.Type {
	case types.DatabaseTypeMySQL:
		conn := mysql.NewConnection(instance)
		if err := conn.Connect(ctx); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrConnectionFailed, err)
		}

		e.mu.Lock()
		e.connections[instanceName] = conn
		e.mu.Unlock()

		return conn, nil

	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", instance.Type)
	}
}

func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var errs []error
	for name, conn := range e.connections {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭连接 %s 失败: %v", name, err))
		}
	}
	e.connections = make(map[string]connection.Connection)

	if len(errs) > 0 {
		return fmt.Errorf("关闭连接时发生错误: %v", errs)
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

	conn, err := e.getConnection(ctx, instanceName)
	if err != nil {
		instance.SetStatus(types.InstanceStatusUnhealthy)
		return types.InstanceStatusUnhealthy, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		instance.SetStatus(types.InstanceStatusUnhealthy)
		return types.InstanceStatusUnhealthy, nil
	}

	instance.SetStatus(types.InstanceStatusHealthy)
	return types.InstanceStatusHealthy, nil
}
