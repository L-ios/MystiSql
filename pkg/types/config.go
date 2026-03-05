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

// DiscoveryConfig 定义服务发现配置
type DiscoveryConfig struct {
	Type string `json:"type" yaml:"type"` // 发现类型：static, k8s, consul
}

// ServerConfig 定义服务器配置
type ServerConfig struct {
	Host string `json:"host" yaml:"host"` // 监听地址
	Port int    `json:"port" yaml:"port"` // 监听端口
	Mode string `json:"mode" yaml:"mode"` // 运行模式：debug, release
}

// Config 定义应用程序的根配置
type Config struct {
	Server    ServerConfig     `json:"server" yaml:"server"`       // 服务器配置
	Discovery DiscoveryConfig  `json:"discovery" yaml:"discovery"` // 发现配置
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
		},
		Instances: make([]InstanceConfig, 0),
	}
}
