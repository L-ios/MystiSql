package connection

import (
	"fmt"
	"sync"

	"MystiSql/pkg/types"
)

// DriverRegistry 驱动注册中心
type DriverRegistry struct {
	mu      sync.RWMutex
	drivers map[types.DatabaseType]ConnectionFactory
}

var (
	registry     *DriverRegistry
	registryOnce sync.Once
)

// GetRegistry 获取驱动注册中心单例
func GetRegistry() *DriverRegistry {
	registryOnce.Do(func() {
		registry = &DriverRegistry{
			drivers: make(map[types.DatabaseType]ConnectionFactory),
		}
	})
	return registry
}

// RegisterDriver 注册数据库驱动
func (r *DriverRegistry) RegisterDriver(driverType types.DatabaseType, factory ConnectionFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if factory == nil {
		return fmt.Errorf("factory cannot be nil for driver type: %s", driverType)
	}

	if _, exists := r.drivers[driverType]; exists {
		return fmt.Errorf("driver already registered: %s", driverType)
	}

	r.drivers[driverType] = factory
	return nil
}

// GetFactory 获取驱动工厂
func (r *DriverRegistry) GetFactory(driverType types.DatabaseType) (ConnectionFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.drivers[driverType]
	if !exists {
		return nil, fmt.Errorf("driver not found: %s", driverType)
	}

	return factory, nil
}

// ListDrivers 列出所有已注册的驱动
func (r *DriverRegistry) ListDrivers() []types.DatabaseType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	drivers := make([]types.DatabaseType, 0, len(r.drivers))
	for driverType := range r.drivers {
		drivers = append(drivers, driverType)
	}
	return drivers
}

// IsDriverRegistered 检查驱动是否已注册
func (r *DriverRegistry) IsDriverRegistered(driverType types.DatabaseType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.drivers[driverType]
	return exists
}

// UnregisterDriver 注销驱动（主要用于测试）
func (r *DriverRegistry) UnregisterDriver(driverType types.DatabaseType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.drivers, driverType)
}

// Clear 清空所有驱动（主要用于测试）
func (r *DriverRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.drivers = make(map[types.DatabaseType]ConnectionFactory)
}
