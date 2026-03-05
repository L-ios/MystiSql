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

// Ensure Discoverer implements InstanceDiscoverer interface
var _ discovery.InstanceDiscoverer = (*Discoverer)(nil)
