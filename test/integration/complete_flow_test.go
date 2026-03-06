package integration

import (
	"context"
	"testing"
	"time"

	"MystiSql/internal/config"
	"MystiSql/internal/discovery"
	"MystiSql/internal/discovery/static"
	"MystiSql/internal/service/query"
	"MystiSql/pkg/types"
)

func TestCompleteFlow(t *testing.T) {
	t.Run("配置加载", func(t *testing.T) {
		cfg := types.NewConfig()
		cfg.Server.Host = "0.0.0.0"
		cfg.Server.Port = 8080
		cfg.Server.Mode = "debug"
		cfg.Instances = []types.InstanceConfig{
			{
				Name:     "test-mysql",
				Type:     types.DatabaseTypeMySQL,
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "root",
				Database: "test",
			},
		}

		if cfg.Server.Host != "0.0.0.0" {
			t.Errorf("期望服务器主机为 0.0.0.0，实际为 %s", cfg.Server.Host)
		}

		if cfg.Server.Port != 8080 {
			t.Errorf("期望服务器端口为 8080，实际为 %d", cfg.Server.Port)
		}

		if len(cfg.Instances) != 1 {
			t.Errorf("期望实例数量为 1，实际为 %d", len(cfg.Instances))
		}

		t.Logf("配置加载成功: %+v", cfg)
	})

	t.Run("实例发现", func(t *testing.T) {
		cfg := &types.Config{
			Discovery: types.DiscoveryConfig{Type: "static"},
			Instances: []types.InstanceConfig{
				{
					Name:     "test-mysql",
					Type:     types.DatabaseTypeMySQL,
					Host:     "localhost",
					Port:     3306,
					Username: "root",
					Password: "root",
					Database: "test",
				},
			},
		}

		discoverer := static.NewDiscoverer(cfg.Instances)
		ctx := context.Background()

		instances, err := discoverer.Discover(ctx)
		if err != nil {
			t.Fatalf("发现实例失败: %v", err)
		}

		if len(instances) != 1 {
			t.Errorf("期望发现 1 个实例，实际为 %d", len(instances))
		}

		instance := instances[0]
		if instance.Name != "test-mysql" {
			t.Errorf("期望实例名为 test-mysql，实际为 %s", instance.Name)
		}

		if instance.Type != types.DatabaseTypeMySQL {
			t.Errorf("期望实例类型为 mysql，实际为 %s", instance.Type)
		}

		t.Logf("实例发现成功: %s@%s:%d", instance.Name, instance.Host, instance.Port)
	})

	t.Run("实例注册", func(t *testing.T) {
		registry := discovery.NewRegistry()

		instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
		instance.SetCredentials("root", "root")
		instance.SetDatabase("test")

		if err := registry.Register(instance); err != nil {
			t.Fatalf("注册实例失败: %v", err)
		}

		registered, err := registry.GetInstance("test-mysql")
		if err != nil {
			t.Fatalf("获取实例失败: %v", err)
		}

		if registered.Name != instance.Name {
			t.Errorf("期望实例名为 %s，实际为 %s", instance.Name, registered.Name)
		}

		t.Logf("实例注册成功: %s", registered.Name)
	})

	t.Run("查询引擎初始化", func(t *testing.T) {
		registry := discovery.NewRegistry()
		instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
		instance.SetCredentials("root", "root")
		instance.SetDatabase("test")

		if err := registry.Register(instance); err != nil {
			t.Fatalf("注册实例失败: %v", err)
		}

		engine := query.NewEngine(registry)
		defer func() { _ = engine.Close() }()

		instances, err := engine.ListInstances()
		if err != nil {
			t.Fatalf("列出实例失败: %v", err)
		}

		if len(instances) != 1 {
			t.Errorf("期望实例数量为 1，实际为 %d", len(instances))
		}

		t.Logf("查询引擎初始化成功")
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("配置文件不存在", func(t *testing.T) {
		_, err := config.LoadFromPath("non-existent-config.yaml")
		if err == nil {
			t.Error("期望加载不存在的配置文件失败，但没有返回错误")
		}
		t.Logf("正确处理配置文件不存在错误: %v", err)
	})

	t.Run("实例不存在", func(t *testing.T) {
		registry := discovery.NewRegistry()

		_, err := registry.GetInstance("non-existent-instance")
		if err == nil {
			t.Error("期望获取不存在的实例失败，但没有返回错误")
		}
		t.Logf("正确处理实例不存在错误: %v", err)
	})

	t.Run("重复注册实例", func(t *testing.T) {
		registry := discovery.NewRegistry()
		instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)

		if err := registry.Register(instance); err != nil {
			t.Fatalf("第一次注册失败: %v", err)
		}

		err := registry.Register(instance)
		if err == nil {
			t.Error("期望重复注册失败，但没有返回错误")
		}
		t.Logf("正确处理重复注册错误: %v", err)
	})

	t.Run("查询引擎实例不存在", func(t *testing.T) {
		registry := discovery.NewRegistry()
		engine := query.NewEngine(registry)
		defer func() { _ = engine.Close() }()

		ctx := context.Background()
		_, err := engine.ExecuteQuery(ctx, "non-existent", "SELECT 1")
		if err == nil {
			t.Error("期望查询不存在的实例失败，但没有返回错误")
		}
		t.Logf("正确处理查询实例不存在错误: %v", err)
	})
}

func TestConfigurationValidation(t *testing.T) {
	t.Run("验证默认配置", func(t *testing.T) {
		cfg := types.NewConfig()

		if cfg.Server.Host != "0.0.0.0" {
			t.Errorf("期望默认主机为 0.0.0.0，实际为 %s", cfg.Server.Host)
		}

		if cfg.Server.Port != 8080 {
			t.Errorf("期望默认端口为 8080，实际为 %d", cfg.Server.Port)
		}

		if cfg.Discovery.Type != "static" {
			t.Errorf("期望默认发现类型为 static，实际为 %s", cfg.Discovery.Type)
		}

		t.Logf("默认配置验证成功")
	})

	t.Run("验证实例配置转换", func(t *testing.T) {
		instanceConfig := types.InstanceConfig{
			Name:     "test",
			Type:     types.DatabaseTypeMySQL,
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "root",
			Database: "test",
			Labels: map[string]string{
				"env": "dev",
			},
		}

		instance := instanceConfig.ToDatabaseInstance()

		if instance.Name != instanceConfig.Name {
			t.Errorf("期望名称为 %s，实际为 %s", instanceConfig.Name, instance.Name)
		}

		if instance.Host != instanceConfig.Host {
			t.Errorf("期望主机为 %s，实际为 %s", instanceConfig.Host, instance.Host)
		}

		if instance.Username != instanceConfig.Username {
			t.Errorf("期望用户名为 %s，实际为 %s", instanceConfig.Username, instance.Username)
		}

		if instance.Labels["env"] != "dev" {
			t.Errorf("期望标签 env 为 dev，实际为 %s", instance.Labels["env"])
		}

		t.Logf("实例配置转换验证成功")
	})
}

func TestTimeoutHandling(t *testing.T) {
	t.Run("上下文超时", func(t *testing.T) {
		registry := discovery.NewRegistry()
		instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
		if err := registry.Register(instance); err != nil {
			t.Fatalf("注册实例失败: %v", err)
		}

		engine := query.NewEngine(registry)
		defer func() { _ = engine.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(1 * time.Millisecond)

		_, err := engine.ExecuteQuery(ctx, "test-mysql", "SELECT 1")
		if err == nil {
			t.Error("期望超时失败，但没有返回错误")
		}
		t.Logf("正确处理超时: %v", err)
	})
}
