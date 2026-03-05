package discovery

import (
	"fmt"
	"sync"

	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"
)

// Registry InstanceRegistry 的实现
type Registry struct {
	mu        sync.RWMutex
	instances map[string]*types.DatabaseInstance
}

// NewRegistry 创建一个新的实例注册中心
func NewRegistry() *Registry {
	return &Registry{
		instances: make(map[string]*types.DatabaseInstance),
	}
}

// Register 注册一个数据库实例
func (r *Registry) Register(instance *types.DatabaseInstance) error {
	if instance == nil {
		return fmt.Errorf("%w: 实例不能为 nil", errors.ErrInvalidInstanceConfig)
	}

	if instance.Name == "" {
		return fmt.Errorf("%w: 实例名称不能为空", errors.ErrInvalidInstanceConfig)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查是否已存在
	if _, exists := r.instances[instance.Name]; exists {
		return fmt.Errorf("%w: %s", errors.ErrInstanceAlreadyExists, instance.Name)
	}

	// 添加实例
	r.instances[instance.Name] = instance
	return nil
}

// GetInstance 根据名称获取数据库实例
func (r *Registry) GetInstance(name string) (*types.DatabaseInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instance, exists := r.instances[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", errors.ErrInstanceNotFound, name)
	}

	return instance, nil
}

// ListInstances 列出所有已注册的实例
func (r *Registry) ListInstances() ([]*types.DatabaseInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 返回空数组而不是 nil
	instances := make([]*types.DatabaseInstance, 0, len(r.instances))
	for _, instance := range r.instances {
		instances = append(instances, instance)
	}

	return instances, nil
}

// Remove 移除一个数据库实例
func (r *Registry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.instances[name]; !exists {
		return fmt.Errorf("%w: %s", errors.ErrInstanceNotFound, name)
	}

	delete(r.instances, name)
	return nil
}

// Clear 清空所有实例
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.instances = make(map[string]*types.DatabaseInstance)
}
