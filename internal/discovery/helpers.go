package discovery

import (
	"context"
	"fmt"
)

// DiscoverAndRegister 使用指定的发现器发现实例并注册到注册中心
func DiscoverAndRegister(ctx context.Context, discoverer InstanceDiscoverer, registry InstanceRegistry) error {
	// 发现实例
	instances, err := discoverer.Discover(ctx)
	if err != nil {
		return fmt.Errorf("发现实例失败: %w", err)
	}

	// 注册所有实例
	registered := 0
	for _, instance := range instances {
		if err := registry.Register(instance); err != nil {
			// 如果实例已存在，跳过（可能是重复注册）
			continue
		}
		registered++
	}

	return nil
}
