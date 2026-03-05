package discovery

import (
	"testing"

	"MystiSql/pkg/types"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	// 测试注册新实例
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	err := registry.Register(instance)
	if err != nil {
		t.Errorf("注册实例失败: %v", err)
	}

	// 测试注册重复实例
	err = registry.Register(instance)
	if err == nil {
		t.Error("期望返回重复实例错误，但注册成功了")
	}
}

func TestRegistry_GetInstance(t *testing.T) {
	registry := NewRegistry()

	// 注册实例
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	_ = registry.Register(instance)

	// 测试获取存在的实例
	got, err := registry.GetInstance("test-mysql")
	if err != nil {
		t.Errorf("获取实例失败: %v", err)
	}
	if got.Name != "test-mysql" {
		t.Errorf("期望实例名称 = 'test-mysql', 实际 = '%s'", got.Name)
	}

	// 测试获取不存在的实例
	_, err = registry.GetInstance("nonexistent")
	if err == nil {
		t.Error("期望返回实例未找到错误，但获取成功了")
	}
}

func TestRegistry_ListInstances(t *testing.T) {
	registry := NewRegistry()

	// 测试空注册中心
	instances, err := registry.ListInstances()
	if err != nil {
		t.Errorf("列出实例失败: %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("期望 0 个实例, 实际 = %d", len(instances))
	}

	// 注册多个实例
	instance1 := types.NewDatabaseInstance("mysql-1", types.DatabaseTypeMySQL, "localhost", 3306)
	instance2 := types.NewDatabaseInstance("mysql-2", types.DatabaseTypeMySQL, "localhost", 3307)
	_ = registry.Register(instance1)
	_ = registry.Register(instance2)

	// 测试列出所有实例
	instances, err = registry.ListInstances()
	if err != nil {
		t.Errorf("列出实例失败: %v", err)
	}
	if len(instances) != 2 {
		t.Errorf("期望 2 个实例, 实际 = %d", len(instances))
	}
}

func TestRegistry_Remove(t *testing.T) {
	registry := NewRegistry()

	// 注册实例
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	_ = registry.Register(instance)

	// 测试移除存在的实例
	err := registry.Remove("test-mysql")
	if err != nil {
		t.Errorf("移除实例失败: %v", err)
	}

	// 验证实例已被移除
	_, err = registry.GetInstance("test-mysql")
	if err == nil {
		t.Error("期望返回实例未找到错误，但实例仍然存在")
	}

	// 测试移除不存在的实例
	err = registry.Remove("nonexistent")
	if err == nil {
		t.Error("期望返回实例未找到错误，但移除成功了")
	}
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	// 注册多个实例
	instance1 := types.NewDatabaseInstance("mysql-1", types.DatabaseTypeMySQL, "localhost", 3306)
	instance2 := types.NewDatabaseInstance("mysql-2", types.DatabaseTypeMySQL, "localhost", 3307)
	_ = registry.Register(instance1)
	_ = registry.Register(instance2)

	// 清空注册中心
	registry.Clear()

	// 验证所有实例已被清空
	instances, _ := registry.ListInstances()
	if len(instances) != 0 {
		t.Errorf("期望 0 个实例, 实际 = %d", len(instances))
	}
}
