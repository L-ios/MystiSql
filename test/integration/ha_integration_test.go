package integration

import (
	"context"
	"testing"

	"MystiSql/internal/service/loadbalancer"
	"MystiSql/internal/service/router"
	"MystiSql/internal/service/topology"
	"MystiSql/pkg/types"

	"go.uber.org/zap"
)

func TestHASystemIntegration(t *testing.T) {
	// 创建拓扑管理器
	config := &types.HAConfig{
		Enabled: true,
		Failover: types.FailoverConfig{
			AutoFailover: false,
		},
	}
	topologyMgr := topology.NewTopologyManager(config)

	// 创建负载均衡器
	lbFactory := loadbalancer.NewLoadBalancerFactory()
	lb := lbFactory.Create(types.ReadStrategyRoundRobin)

	// 创建读写路由器
	logger := zap.NewNop()
	readWriteRouter := router.NewReadWriteRouter(lb, logger)

	// 测试1: 注册主从拓扑
	master := &types.DatabaseInstance{
		Name:   "master-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3306,
		Role:   string(types.InstanceRoleMaster),
		Status: types.InstanceStatusHealthy,
	}

	topology := types.NewMasterSlaveTopology("test-topology", master)
	err := topologyMgr.RegisterTopology(topology)
	if err != nil {
		t.Fatalf("RegisterTopology failed: %v", err)
	}

	// 测试2: 添加从库
	slave1 := &types.DatabaseInstance{
		Name:   "slave-1",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3307,
		Role:   string(types.InstanceRoleSlave),
		Status: types.InstanceStatusHealthy,
	}

	slave2 := &types.DatabaseInstance{
		Name:   "slave-2",
		Type:   types.DatabaseTypeMySQL,
		Host:   "localhost",
		Port:   3308,
		Role:   string(types.InstanceRoleSlave),
		Status: types.InstanceStatusHealthy,
	}

	err = topologyMgr.AddSlave("test-topology", slave1)
	if err != nil {
		t.Fatalf("AddSlave failed: %v", err)
	}

	err = topologyMgr.AddSlave("test-topology", slave2)
	if err != nil {
		t.Fatalf("AddSlave failed: %v", err)
	}

	// 获取拓扑并设置到路由器
	topologies := make(map[string]*types.MasterSlaveTopology)
	topologies["test-topology"] = topology
	readWriteRouter.SetTopology(topologies)

	// 测试3: 路由读写请求
	t.Run("ReadWriteRouting", func(t *testing.T) {
		ctx := context.Background()
		// 测试读请求路由到从库
		readSQL := "SELECT * FROM users"
		result, err := readWriteRouter.Route(ctx, "test-topology", readSQL, false)
		if err != nil {
			t.Fatalf("Route failed: %v", err)
		}

		// 读请求应该路由到从库
		if result.TargetInstance != "slave-1" && result.TargetInstance != "slave-2" {
			t.Errorf("Expected slave-1 or slave-2 for read, got %s", result.TargetInstance)
		}

		// 测试写请求路由到主库
		writeSQL := "INSERT INTO users (name) VALUES ('test')"
		result, err = readWriteRouter.Route(ctx, "test-topology", writeSQL, false)
		if err != nil {
			t.Fatalf("Route failed: %v", err)
		}

		// 写请求应该路由到主库
		if result.TargetInstance != "master-1" {
			t.Errorf("Expected master-1 for write, got %s", result.TargetInstance)
		}
	})

	// 测试4: 主库故障时的行为
	t.Run("MasterFailure", func(t *testing.T) {
		ctx := context.Background()
		// 将主库标记为不健康
		master.Status = types.InstanceStatusUnhealthy
		topology.UpdateStatus()

		// 写请求应该仍然尝试主库（即使不健康）
		writeSQL := "INSERT INTO users (name) VALUES ('test')"
		result, err := readWriteRouter.Route(ctx, "test-topology", writeSQL, false)
		if err != nil {
			t.Fatalf("Route failed: %v", err)
		}

		if result.TargetInstance != "master-1" {
			t.Errorf("Expected master-1 for write even when unhealthy, got %s", result.TargetInstance)
		}

		// 读请求应该仍然路由到从库
		readSQL := "SELECT * FROM users"
		result, err = readWriteRouter.Route(ctx, "test-topology", readSQL, false)
		if err != nil {
			t.Fatalf("Route failed: %v", err)
		}

		if result.TargetInstance != "slave-1" && result.TargetInstance != "slave-2" {
			t.Errorf("Expected slave-1 or slave-2 for read, got %s", result.TargetInstance)
		}
	})

	// 测试5: 从库故障时的行为
	t.Run("SlaveFailure", func(t *testing.T) {
		ctx := context.Background()
		// 恢复主库健康状态
		master.Status = types.InstanceStatusHealthy
		topology.UpdateStatus()

		// 将一个从库标记为不健康
		slave1.Status = types.InstanceStatusUnhealthy

		// 读请求应该路由到健康的从库
		readSQL := "SELECT * FROM users"
		result, err := readWriteRouter.Route(ctx, "test-topology", readSQL, false)
		if err != nil {
			t.Fatalf("Route failed: %v", err)
		}

		if result.TargetInstance != "slave-2" {
			t.Errorf("Expected slave-2 for read when slave-1 is unhealthy, got %s", result.TargetInstance)
		}
	})

	// 测试6: 主从切换
	t.Run("MasterSlaveSwitch", func(t *testing.T) {
		ctx := context.Background()
		// 执行主从切换
		err := topologyMgr.PromoteSlave("test-topology", "slave-2", true)
		if err != nil {
			t.Fatalf("PromoteSlave failed: %v", err)
		}

		// 验证新的主库
		updatedTopology, err := topologyMgr.GetTopology("test-topology")
		if err != nil {
			t.Fatalf("GetTopology failed: %v", err)
		}

		if updatedTopology.Master.Name != "slave-2" {
			t.Errorf("Expected new master to be slave-2, got %s", updatedTopology.Master.Name)
		}

		if updatedTopology.Master.Role != string(types.InstanceRoleMaster) {
			t.Errorf("Expected new master role to be master, got %s", updatedTopology.Master.Role)
		}

		// 更新路由器的拓扑
		topologies["test-topology"] = updatedTopology
		readWriteRouter.SetTopology(topologies)

		// 写请求应该路由到新的主库
		writeSQL := "INSERT INTO users (name) VALUES ('test')"
		result, err := readWriteRouter.Route(ctx, "test-topology", writeSQL, false)
		if err != nil {
			t.Fatalf("Route failed: %v", err)
		}

		if result.TargetInstance != "slave-2" {
			t.Errorf("Expected slave-2 (new master) for write, got %s", result.TargetInstance)
		}
	})
}

func TestLoadBalancerIntegration(t *testing.T) {
	// 创建不同类型的负载均衡器
	roundRobinLB := loadbalancer.NewRoundRobinLB()
	weightedLB := loadbalancer.NewWeightedLB()
	leastConnLB := loadbalancer.NewLeastConnLB()

	// 创建测试实例
	instances := []*types.DatabaseInstance{
		{Name: "instance-1", Weight: 5},
		{Name: "instance-2", Weight: 3},
		{Name: "instance-3", Weight: 2},
	}

	// 测试轮询负载均衡
	t.Run("RoundRobinLB", func(t *testing.T) {
		// 重置计数器
		roundRobinLB.Reset()

		// 测试轮询选择
		instance1, err := roundRobinLB.Select(instances)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if instance1.Name != "instance-1" {
			t.Errorf("Expected instance-1, got %s", instance1.Name)
		}

		instance2, err := roundRobinLB.Select(instances)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if instance2.Name != "instance-2" {
			t.Errorf("Expected instance-2, got %s", instance2.Name)
		}

		instance3, err := roundRobinLB.Select(instances)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if instance3.Name != "instance-3" {
			t.Errorf("Expected instance-3, got %s", instance3.Name)
		}

		// 测试循环
		instance1Again, err := roundRobinLB.Select(instances)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if instance1Again.Name != "instance-1" {
			t.Errorf("Expected instance-1 (wrap around), got %s", instance1Again.Name)
		}
	})

	// 测试权重负载均衡
	t.Run("WeightedLB", func(t *testing.T) {
		// 设置权重
		weightedLB.SetWeight("instance-1", 5)
		weightedLB.SetWeight("instance-2", 3)
		weightedLB.SetWeight("instance-3", 2)

		// 测试选择
		instance, err := weightedLB.Select(instances)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}

		// 应该返回其中一个实例
		if instance.Name != "instance-1" && instance.Name != "instance-2" && instance.Name != "instance-3" {
			t.Errorf("Unexpected instance: %s", instance.Name)
		}
	})

	// 测试最少连接负载均衡
	t.Run("LeastConnLB", func(t *testing.T) {
		// 重置连接计数
		leastConnLB.Reset()

		// 添加连接
		leastConnLB.Increment("instance-1")
		leastConnLB.Increment("instance-1")
		leastConnLB.Increment("instance-2")

		// 应该选择连接数最少的实例-3
		instance, err := leastConnLB.Select(instances)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}

		if instance.Name != "instance-3" {
			t.Errorf("Expected instance-3 (least connections), got %s", instance.Name)
		}
	})
}
