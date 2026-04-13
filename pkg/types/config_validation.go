package types

import (
	"fmt"
	"strings"
)

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if err := c.validateInstances(); err != nil {
		return err
	}

	if c.HA.Enabled {
		if err := c.validateHA(); err != nil {
			return err
		}
	}

	if c.MultiCluster.Enabled {
		if err := c.validateMultiCluster(); err != nil {
			return err
		}
	}

	return nil
}

// validateInstances 验证实例配置
func (c *Config) validateInstances() error {
	instanceNames := make(map[string]bool)
	masterNames := make(map[string]bool)
	slaveConfigs := make([]InstanceConfig, 0)

	for _, instance := range c.Instances {
		if instance.Name == "" {
			return fmt.Errorf("instance name cannot be empty")
		}

		if instanceNames[instance.Name] {
			return fmt.Errorf("duplicate instance name: %s", instance.Name)
		}
		instanceNames[instance.Name] = true

		if instance.Host == "" {
			return fmt.Errorf("instance %s: host cannot be empty", instance.Name)
		}

		if instance.Port < 1 || instance.Port > 65535 {
			return fmt.Errorf("instance %s: port must be between 1 and 65535", instance.Name)
		}

		if !isValidDatabaseType(instance.Type) {
			return fmt.Errorf("instance %s: unsupported database type: %s", instance.Name, instance.Type)
		}

		if instance.Role == string(InstanceRoleSlave) {
			slaveConfigs = append(slaveConfigs, instance)
		} else if instance.Role == string(InstanceRoleMaster) || instance.Role == "" {
			masterNames[instance.Name] = true
		}
	}

	for _, slave := range slaveConfigs {
		if slave.Master == "" {
			return fmt.Errorf("slave instance %s: master field is required for slave instances", slave.Name)
		}
		if !masterNames[slave.Master] {
			return fmt.Errorf("slave instance %s: master '%s' not found in instances", slave.Name, slave.Master)
		}
	}

	return nil
}

// validateHA 验证高可用配置
func (c *Config) validateHA() error {
	ha := &c.HA

	if ha.ReadWriteSplitting.Enabled {
		if !isValidReadStrategy(ha.ReadWriteSplitting.ReadStrategy) {
			return fmt.Errorf("invalid read strategy: %s, must be one of: round-robin, weighted, least-conn", ha.ReadWriteSplitting.ReadStrategy)
		}
	}

	if ha.HealthCheck.Enabled {
		if ha.HealthCheck.Interval <= 0 {
			return fmt.Errorf("health check interval must be positive")
		}
		if ha.HealthCheck.Timeout <= 0 {
			return fmt.Errorf("health check timeout must be positive")
		}
		if ha.HealthCheck.FailureThreshold < 1 {
			return fmt.Errorf("health check failure threshold must be at least 1")
		}
		if ha.HealthCheck.RecoveryThreshold < 1 {
			return fmt.Errorf("health check recovery threshold must be at least 1")
		}
	}

	if ha.Failover.Enabled {
		if ha.Failover.Timeout <= 0 {
			return fmt.Errorf("failover timeout must be positive")
		}
	}

	return nil
}

// validateMultiCluster 验证多集群配置
func (c *Config) validateMultiCluster() error {
	mc := &c.MultiCluster

	if len(mc.Clusters) == 0 {
		return fmt.Errorf("multi-cluster enabled but no clusters configured")
	}

	clusterNames := make(map[string]bool)
	for _, cluster := range mc.Clusters {
		if cluster.Name == "" {
			return fmt.Errorf("cluster name cannot be empty")
		}

		if clusterNames[cluster.Name] {
			return fmt.Errorf("duplicate cluster name: %s", cluster.Name)
		}
		clusterNames[cluster.Name] = true

		if cluster.Kubeconfig == "" {
			return fmt.Errorf("cluster %s: kubeconfig path cannot be empty", cluster.Name)
		}
	}

	return nil
}

// isValidDatabaseType 检查数据库类型是否有效
func isValidDatabaseType(dbType DatabaseType) bool {
	validTypes := map[DatabaseType]bool{
		DatabaseTypeMySQL:         true,
		DatabaseTypePostgreSQL:    true,
		DatabaseTypeOracle:        true,
		DatabaseTypeRedis:         true,
		DatabaseTypeSQLite:        true,
		DatabaseTypeMSSQL:         true,
		DatabaseTypeMongoDB:       true,
		DatabaseTypeElasticsearch: true,
		DatabaseTypeClickHouse:    true,
		DatabaseTypeEtcd:          true,
	}
	return validTypes[dbType]
}

// isValidReadStrategy 检查读取策略是否有效
func isValidReadStrategy(strategy ReadStrategy) bool {
	validStrategies := map[ReadStrategy]bool{
		ReadStrategyRoundRobin: true,
		ReadStrategyWeighted:   true,
		ReadStrategyLeastConn:  true,
	}
	return validStrategies[strategy] || strategy == ""
}

// GetTopologyGroups 获取拓扑组（主库及其从库）
func (c *Config) GetTopologyGroups() map[string]*TopologyGroup {
	groups := make(map[string]*TopologyGroup)

	for _, instance := range c.Instances {
		if instance.Role == string(InstanceRoleMaster) || instance.Role == "" {
			if _, exists := groups[instance.Name]; !exists {
				groups[instance.Name] = &TopologyGroup{
					Master: instance,
					Slaves: make([]InstanceConfig, 0),
				}
			}
		}
	}

	for _, instance := range c.Instances {
		if instance.Role == string(InstanceRoleSlave) && instance.Master != "" {
			if group, exists := groups[instance.Master]; exists {
				group.Slaves = append(group.Slaves, instance)
			}
		}
	}

	return groups
}

// TopologyGroup 表示一个拓扑组（主库及其从库）
type TopologyGroup struct {
	Master InstanceConfig
	Slaves []InstanceConfig
}

// GetMasterInstances 获取所有主库实例
func (c *Config) GetMasterInstances() []InstanceConfig {
	masters := make([]InstanceConfig, 0)
	for _, instance := range c.Instances {
		if instance.Role == string(InstanceRoleMaster) || instance.Role == "" {
			masters = append(masters, instance)
		}
	}
	return masters
}

// GetSlaveInstances 获取所有从库实例
func (c *Config) GetSlaveInstances() []InstanceConfig {
	slaves := make([]InstanceConfig, 0)
	for _, instance := range c.Instances {
		if instance.Role == string(InstanceRoleSlave) {
			slaves = append(slaves, instance)
		}
	}
	return slaves
}

// GetSlavesByMaster 根据主库名称获取从库列表
func (c *Config) GetSlavesByMaster(masterName string) []InstanceConfig {
	slaves := make([]InstanceConfig, 0)
	for _, instance := range c.Instances {
		if instance.Role == string(InstanceRoleSlave) && instance.Master == masterName {
			slaves = append(slaves, instance)
		}
	}
	return slaves
}

// ParseInstanceName 解析实例名称（支持集群前缀）
func ParseInstanceName(fullName string) (cluster, name string) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", fullName
}

// FormatInstanceName 格式化实例名称（添加集群前缀）
func FormatInstanceName(cluster, name string) string {
	if cluster == "" {
		return name
	}
	return fmt.Sprintf("%s/%s", cluster, name)
}
