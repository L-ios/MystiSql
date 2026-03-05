package discovery

import (
	"context"

	"MystiSql/pkg/types"
)

// InstanceDiscoverer 定义服务发现接口
// 所有发现实现（静态、K8s、Consul 等）都必须实现此接口
type InstanceDiscoverer interface {
	// Name 返回发现器的名称
	Name() string

	// Discover 发现并返回数据库实例列表
	// ctx 用于取消操作和超时控制
	Discover(ctx context.Context) ([]*types.DatabaseInstance, error)
}

// InstanceRegistry 定义实例注册中心接口
// 用于存储和管理已发现的数据库实例
type InstanceRegistry interface {
	// Register 注册一个数据库实例
	// 如果实例名称已存在，返回 ErrInstanceAlreadyExists 错误
	Register(instance *types.DatabaseInstance) error

	// GetInstance 根据名称获取数据库实例
	// 如果实例不存在，返回 ErrInstanceNotFound 错误
	GetInstance(name string) (*types.DatabaseInstance, error)

	// ListInstances 列出所有已注册的数据库实例
	// 如果没有实例，返回空切片（不是 nil）
	ListInstances() ([]*types.DatabaseInstance, error)

	// Remove 移除一个数据库实例
	// 如果实例不存在，返回 ErrInstanceNotFound 错误
	Remove(name string) error

	// Clear 清空所有实例
	Clear()
}
