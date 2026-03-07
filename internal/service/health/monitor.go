package health

import (
	"context"
	"sync"
	"time"

	"MystiSql/internal/discovery"
	"MystiSql/internal/service/query"
	"MystiSql/pkg/types"

	"go.uber.org/zap"
)

// Monitor 实例健康监控服务
type Monitor struct {
	// 实例注册表
	registry discovery.InstanceRegistry

	// 查询引擎
	engine *query.Engine

	// 日志器
	logger *zap.Logger

	// 监控间隔
	interval time.Duration

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc

	// 互斥锁
	mu sync.RWMutex

	// 健康状态缓存
	statusCache map[string]types.InstanceStatus

	// 事件通道
	eventCh chan HealthEvent

	// 运行标志
	running bool
}

// HealthEvent 健康状态事件
type HealthEvent struct {
	// 实例名称
	InstanceName string

	// 旧状态
	OldStatus types.InstanceStatus

	// 新状态
	NewStatus types.InstanceStatus

	// 事件时间
	Timestamp time.Time
}

// NewMonitor 创建一个新的健康监控服务
func NewMonitor(registry discovery.InstanceRegistry, engine *query.Engine, logger *zap.Logger, interval time.Duration) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Monitor{
		registry:    registry,
		engine:      engine,
		logger:      logger,
		interval:    interval,
		ctx:         ctx,
		cancel:      cancel,
		statusCache: make(map[string]types.InstanceStatus),
		eventCh:     make(chan HealthEvent, 100),
		running:     false,
	}
}

// Start 启动健康监控服务
func (m *Monitor) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	go m.monitorLoop()
	m.logger.Info("Health monitor started", zap.Duration("interval", m.interval))
}

// Stop 停止健康监控服务
func (m *Monitor) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	m.mu.Unlock()

	m.cancel()
	close(m.eventCh)
	m.logger.Info("Health monitor stopped")
}

// GetEventChannel 获取事件通道
func (m *Monitor) GetEventChannel() <-chan HealthEvent {
	return m.eventCh
}

// GetInstanceStatus 获取实例健康状态
func (m *Monitor) GetInstanceStatus(instanceName string) types.InstanceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if status, exists := m.statusCache[instanceName]; exists {
		return status
	}

	return types.InstanceStatusUnknown
}

// GetAllStatus 获取所有实例的健康状态
func (m *Monitor) GetAllStatus() map[string]types.InstanceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make(map[string]types.InstanceStatus)
	for name, status := range m.statusCache {
		statuses[name] = status
	}

	return statuses
}

// monitorLoop 监控循环
func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// 初始检查
	m.checkAllInstances()

	for {
		select {
		case <-ticker.C:
			m.checkAllInstances()
		case <-m.ctx.Done():
			return
		}
	}
}

// checkAllInstances 检查所有实例的健康状态
func (m *Monitor) checkAllInstances() {
	instances, err := m.registry.ListInstances()
	if err != nil {
		m.logger.Error("Failed to list instances", zap.Error(err))
		return
	}

	for _, instance := range instances {
		m.checkInstance(instance)
	}

	// 清理不存在的实例状态
	m.cleanupStatusCache(instances)
}

// checkInstance 检查单个实例的健康状态
func (m *Monitor) checkInstance(instance *types.DatabaseInstance) {
	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	status, err := m.engine.GetInstanceHealth(ctx, instance.Name)
	if err != nil {
		m.logger.Warn("Failed to check instance health",
			zap.String("instance", instance.Name),
			zap.Error(err),
		)
		status = types.InstanceStatusUnhealthy
	}

	m.updateStatus(instance.Name, status)
}

// updateStatus 更新实例健康状态并发送事件
func (m *Monitor) updateStatus(instanceName string, newStatus types.InstanceStatus) {
	m.mu.Lock()
	oldStatus, exists := m.statusCache[instanceName]
	m.statusCache[instanceName] = newStatus
	m.mu.Unlock()

	// 状态变化时发送事件
	if !exists || oldStatus != newStatus {
		event := HealthEvent{
			InstanceName: instanceName,
			OldStatus:    oldStatus,
			NewStatus:    newStatus,
			Timestamp:    time.Now(),
		}

		select {
		case m.eventCh <- event:
			m.logger.Info("Health status changed",
				zap.String("instance", instanceName),
				zap.String("old_status", string(oldStatus)),
				zap.String("new_status", string(newStatus)),
			)
		default:
			m.logger.Warn("Event channel full, skipping health event",
				zap.String("instance", instanceName),
			)
		}
	}
}

// cleanupStatusCache 清理不存在的实例状态
func (m *Monitor) cleanupStatusCache(instances []*types.DatabaseInstance) {
	// 创建当前实例集合
	currentInstances := make(map[string]bool)
	for _, instance := range instances {
		currentInstances[instance.Name] = true
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 删除不存在的实例状态
	for name := range m.statusCache {
		if !currentInstances[name] {
			delete(m.statusCache, name)
			m.logger.Info("Removed status cache for non-existent instance",
				zap.String("instance", name),
			)
		}
	}
}

// HealthChecker 健康检查接口
type HealthChecker interface {
	// CheckInstanceHealth 检查实例健康状态
	CheckInstanceHealth(ctx context.Context, instanceName string) (types.InstanceStatus, error)

	// GetInstanceStatus 获取实例健康状态
	GetInstanceStatus(instanceName string) types.InstanceStatus

	// GetAllStatus 获取所有实例的健康状态
	GetAllStatus() map[string]types.InstanceStatus

	// Start 启动健康检查
	Start()

	// Stop 停止健康检查
	Stop()

	// GetEventChannel 获取事件通道
	GetEventChannel() <-chan HealthEvent
}
