package static

import (
	"context"

	"MystiSql/internal/discovery"
	"MystiSql/pkg/types"
)

// Discoverer 静态发现实现
type Discoverer struct {
	instances []types.InstanceConfig
}

// NewDiscoverer 创建一个新的静态发现器
func NewDiscoverer(instances []types.InstanceConfig) *Discoverer {
	return &Discoverer{
		instances: instances,
	}
}

// Name 返回发现器的名称
func (d *Discoverer) Name() string {
	return "static"
}

// Discover 从配置中发现数据库实例
func (d *Discoverer) Discover(ctx context.Context) ([]*types.DatabaseInstance, error) {
	// 检查是否已取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 如果没有配置实例，返回空列表
	if len(d.instances) == 0 {
		return []*types.DatabaseInstance{}, nil
	}

	// 转换配置为实例对象
	instances := make([]*types.DatabaseInstance, 0, len(d.instances))
	for _, cfg := range d.instances {
		// 将 InstanceConfig 转换为 DatabaseInstance
		instance := cfg.ToDatabaseInstance()

		// 初始状态设为 Unknown
		instance.SetStatus(types.InstanceStatusUnknown)

		instances = append(instances, instance)
	}

	return instances, nil
}

// Watch 监听实例变化事件（静态发现器不支持）
func (d *Discoverer) Watch(ctx context.Context) (<-chan discovery.DiscoveryEvent, error) {
	// 静态发现器不支持动态变化，返回一个关闭的通道
	ch := make(chan discovery.DiscoveryEvent)
	close(ch)
	return ch, nil
}

// Stop 停止发现器（静态发现器不需要停止）
func (d *Discoverer) Stop() error {
	return nil
}

// Ensure Discoverer implements InstanceDiscoverer interface
var _ discovery.InstanceDiscoverer = (*Discoverer)(nil)
