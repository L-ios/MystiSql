package types

// InstanceConfig 定义单个数据库实例的配置
type InstanceConfig struct {
	Name     string            `json:"name" yaml:"name"`                             // 实例名称
	Type     DatabaseType      `json:"type" yaml:"type"`                             // 数据库类型
	Host     string            `json:"host" yaml:"host"`                             // 主机地址
	Port     int               `json:"port" yaml:"port"`                             // 端口号
	Username string            `json:"username,omitempty" yaml:"username,omitempty"` // 用户名
	Password string            `json:"-" yaml:"password,omitempty"`                  // 密码
	Database string            `json:"database,omitempty" yaml:"database,omitempty"` // 数据库名
	Labels   map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`     // 标签
}

// ToDatabaseInstance 将 InstanceConfig 转换为 DatabaseInstance
func (c *InstanceConfig) ToDatabaseInstance() *DatabaseInstance {
	instance := NewDatabaseInstance(c.Name, c.Type, c.Host, c.Port)
	instance.SetCredentials(c.Username, c.Password)

	if c.Database != "" {
		instance.SetDatabase(c.Database)
	}

	if c.Labels != nil {
		instance.Labels = c.Labels
	}

	return instance
}

// K8sDiscoveryConfig 定义 K8s 服务发现配置
type K8sDiscoveryConfig struct {
	Namespace     string         `json:"namespace" yaml:"namespace"`         // K8s 命名空间
	LabelSelector string         `json:"labelSelector" yaml:"labelSelector"` // 标签选择器
	PortMapping   map[string]int `json:"portMapping" yaml:"portMapping"`     // 端口映射
}

// DiscoveryConfig 定义服务发现配置
type DiscoveryConfig struct {
	Type string             `json:"type" yaml:"type"`                   // 发现类型：static, k8s, consul
	K8s  K8sDiscoveryConfig `json:"k8s,omitempty" yaml:"k8s,omitempty"` // K8s 发现配置
}

// ServerConfig 定义服务器配置
type ServerConfig struct {
	Host string `json:"host" yaml:"host"` // 监听地址
	Port int    `json:"port" yaml:"port"` // 监听端口
	Mode string `json:"mode" yaml:"mode"` // 运行模式：debug, release
}

// HealthConfig 定义健康监控配置
type HealthConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`   // 是否启用健康监控
	Interval string `json:"interval" yaml:"interval"` // 监控间隔
}

// PoolConfig 定义连接池配置
type PoolConfig struct {
	MaxConnections    int    `json:"maxConnections" yaml:"maxConnections"`       // 最大连接数
	MinConnections    int    `json:"minConnections" yaml:"minConnections"`       // 最小连接数
	MaxIdleTime       string `json:"maxIdleTime" yaml:"maxIdleTime"`             // 最大空闲时间
	MaxLifetime       string `json:"maxLifetime" yaml:"maxLifetime"`             // 最大生命周期
	ConnectionTimeout string `json:"connectionTimeout" yaml:"connectionTimeout"` // 连接超时时间
	PingInterval      string `json:"pingInterval" yaml:"pingInterval"`           // 健康检查间隔
}

// Config 定义应用程序的根配置
type Config struct {
	Server    ServerConfig     `json:"server" yaml:"server"`       // 服务器配置
	Discovery DiscoveryConfig  `json:"discovery" yaml:"discovery"` // 发现配置
	Health    HealthConfig     `json:"health" yaml:"health"`       // 健康监控配置
	Pool      PoolConfig       `json:"pool" yaml:"pool"`           // 连接池配置
	Instances []InstanceConfig `json:"instances" yaml:"instances"` // 实例列表
}

// NewConfig 创建一个带有默认值的配置
func NewConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
			Mode: "release",
		},
		Discovery: DiscoveryConfig{
			Type: "static",
			K8s: K8sDiscoveryConfig{
				Namespace:     "default",
				LabelSelector: "app=database",
				PortMapping: map[string]int{
					"mysql":      3306,
					"postgresql": 5432,
					"oracle":     1521,
					"redis":      6379,
				},
			},
		},
		Health: HealthConfig{
			Enabled:  true,
			Interval: "30s",
		},
		Pool: PoolConfig{
			MaxConnections:    10,
			MinConnections:    2,
			MaxIdleTime:       "30s",
			MaxLifetime:       "1h",
			ConnectionTimeout: "10s",
			PingInterval:      "30s",
		},
		Instances: make([]InstanceConfig, 0),
	}
}
