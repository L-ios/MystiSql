package static

import (
	"context"
	"testing"

	"MystiSql/pkg/types"
)

func TestDiscoverer_Name(t *testing.T) {
	discoverer := NewDiscoverer([]types.InstanceConfig{})
	if discoverer.Name() != "static" {
		t.Errorf("期望名称 = 'static', 实际 = '%s'", discoverer.Name())
	}
}

func TestDiscoverer_Discover(t *testing.T) {
	// 测试空配置
	t.Run("空配置", func(t *testing.T) {
		discoverer := NewDiscoverer([]types.InstanceConfig{})
		instances, err := discoverer.Discover(context.Background())
		if err != nil {
			t.Errorf("发现失败: %v", err)
		}
		if len(instances) != 0 {
			t.Errorf("期望 0 个实例, 实际 = %d", len(instances))
		}
	})

	// 测试单个实例
	t.Run("单个实例", func(t *testing.T) {
		configs := []types.InstanceConfig{
			{
				Name: "test-mysql",
				Type: types.DatabaseTypeMySQL,
				Host: "localhost",
				Port: 3306,
			},
		}
		discoverer := NewDiscoverer(configs)
		instances, err := discoverer.Discover(context.Background())
		if err != nil {
			t.Errorf("发现失败: %v", err)
		}
		if len(instances) != 1 {
			t.Fatalf("期望 1 个实例, 实际 = %d", len(instances))
		}

		instance := instances[0]
		if instance.Name != "test-mysql" {
			t.Errorf("期望名称 = 'test-mysql', 实际 = '%s'", instance.Name)
		}
		if instance.Host != "localhost" {
			t.Errorf("期望 Host = 'localhost', 实际 = '%s'", instance.Host)
		}
		if instance.Port != 3306 {
			t.Errorf("期望 Port = 3306, 实际 = %d", instance.Port)
		}
		if instance.Status != types.InstanceStatusUnknown {
			t.Errorf("期望 Status = 'unknown', 实际 = '%s'", instance.Status)
		}
	})

	// 测试多个实例
	t.Run("多个实例", func(t *testing.T) {
		configs := []types.InstanceConfig{
			{
				Name: "mysql-1",
				Type: types.DatabaseTypeMySQL,
				Host: "localhost",
				Port: 3306,
			},
			{
				Name: "mysql-2",
				Type: types.DatabaseTypeMySQL,
				Host: "localhost",
				Port: 3307,
			},
		}
		discoverer := NewDiscoverer(configs)
		instances, err := discoverer.Discover(context.Background())
		if err != nil {
			t.Errorf("发现失败: %v", err)
		}
		if len(instances) != 2 {
			t.Errorf("期望 2 个实例, 实际 = %d", len(instances))
		}
	})

	// 测试上下文取消
	t.Run("上下文取消", func(t *testing.T) {
		discoverer := NewDiscoverer([]types.InstanceConfig{})
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消

		_, err := discoverer.Discover(ctx)
		if err == nil {
			t.Error("期望返回上下文取消错误，但发现成功了")
		}
	})
}

func TestDiscoverer_InstanceWithCredentials(t *testing.T) {
	configs := []types.InstanceConfig{
		{
			Name:     "test-mysql",
			Type:     types.DatabaseTypeMySQL,
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "secret",
			Database: "testdb",
		},
	}
	discoverer := NewDiscoverer(configs)
	instances, err := discoverer.Discover(context.Background())
	if err != nil {
		t.Fatalf("发现失败: %v", err)
	}

	instance := instances[0]
	if instance.Username != "root" {
		t.Errorf("期望 Username = 'root', 实际 = '%s'", instance.Username)
	}
	if instance.Password != "secret" {
		t.Errorf("期望 Password = 'secret', 实际 = '%s'", instance.Password)
	}
	if instance.Database != "testdb" {
		t.Errorf("期望 Database = 'testdb', 实际 = '%s'", instance.Database)
	}
}
