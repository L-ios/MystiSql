package types

// InstanceConfig 定义单个数据库实例的配置
type InstanceConfig struct {
	Name      string            `json:"name" yaml:"name"`
	Type      DatabaseType      `json:"type" yaml:"type"`
	Host      string            `json:"host" yaml:"host"`
	Port      int               `json:"port" yaml:"port"`
	Username  string            `json:"username,omitempty" yaml:"username,omitempty"`
	Password  string            `json:"-" yaml:"password,omitempty"`
	Database  string            `json:"database,omitempty" yaml:"database,omitempty"`
	Labels    map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Role      string            `json:"role,omitempty" yaml:"role,omitempty"`           // primary, replica, readwrite (default)
	ReplicaOf string            `json:"replicaOf,omitempty" yaml:"replicaOf,omitempty"` // Name of primary instance
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

	if c.Role != "" {
		instance.Role = c.Role
	}
	if c.ReplicaOf != "" {
		instance.ReplicaOf = c.ReplicaOf
	}

	return instance
}

// K8sDiscoveryConfig 定义 K8s 服务发现配置
type K8sDiscoveryConfig struct {
	Namespace     string         `json:"namespace" yaml:"namespace"`
	LabelSelector string         `json:"labelSelector" yaml:"labelSelector"`
	PortMapping   map[string]int `json:"portMapping" yaml:"portMapping"`
}

// DiscoveryConfig 定义服务发现配置
type DiscoveryConfig struct {
	Type string             `json:"type" yaml:"type"`
	K8s  K8sDiscoveryConfig `json:"k8s,omitempty" yaml:"k8s,omitempty"`
}

// ServerConfig 定义服务器配置
type ServerConfig struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
	Mode string `json:"mode" yaml:"mode"`
}

// HealthConfig 定义健康监控配置
type HealthConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	Interval string `json:"interval" yaml:"interval"`
}

// PoolConfig 定义连接池配置
type PoolConfig struct {
	MaxConnections    int    `json:"maxConnections" yaml:"maxConnections"`
	MinConnections    int    `json:"minConnections" yaml:"minConnections"`
	MaxIdleTime       string `json:"maxIdleTime" yaml:"maxIdleTime"`
	MaxLifetime       string `json:"maxLifetime" yaml:"maxLifetime"`
	ConnectionTimeout string `json:"connectionTimeout" yaml:"connectionTimeout"`
	PingInterval      string `json:"pingInterval" yaml:"pingInterval"`
}

// TokenConfig 定义 Token 配置
type TokenConfig struct {
	Secret string `json:"secret" yaml:"secret"`
	Expire string `json:"expire" yaml:"expire"`
}

// AuthConfig 定义认证配置
type AuthConfig struct {
	Enabled   bool        `json:"enabled" yaml:"enabled"`
	Token     TokenConfig `json:"token" yaml:"token"`
	Whitelist []string    `json:"whitelist" yaml:"whitelist"`
}

// AuditConfig 定义审计配置
type AuditConfig struct {
	Enabled       bool   `json:"enabled" yaml:"enabled"`
	LogFile       string `json:"logFile" yaml:"logFile"`
	RetentionDays int    `json:"retentionDays" yaml:"retentionDays"`
}

// ValidatorConfig 定义验证器配置
type ValidatorConfig struct {
	Enabled             bool     `json:"enabled" yaml:"enabled"`
	DangerousOperations []string `json:"dangerousOperations" yaml:"dangerousOperations"`
	Whitelist           []string `json:"whitelist" yaml:"whitelist"`
	Blacklist           []string `json:"blacklist" yaml:"blacklist"`
}

// WebSocketConfig 定义 WebSocket 配置
type WebSocketConfig struct {
	Enabled              bool     `json:"enabled" yaml:"enabled"`
	MaxConnections       int      `json:"maxConnections" yaml:"maxConnections"`
	IdleTimeout          string   `json:"idleTimeout" yaml:"idleTimeout"`
	MaxConcurrentQueries int      `json:"maxConcurrentQueries" yaml:"maxConcurrentQueries"`
	AllowedOrigins       []string `json:"allowedOrigins" yaml:"allowedOrigins"`
	MaxMessageSize       int64    `json:"maxMessageSize" yaml:"maxMessageSize"`
}

// WebUIConfig 定义 WebUI 配置
type WebUIConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Mode    string `json:"mode" yaml:"mode"` // embedded | external
}

// CORSConfig 定义 CORS 跨域配置
type CORSConfig struct {
	AllowedOrigins []string `json:"allowedOrigins" yaml:"allowedOrigins"`
}

// Config 定义应用程序的根配置
type Config struct {
	Server    ServerConfig     `json:"server" yaml:"server"`
	Auth      AuthConfig       `json:"auth" yaml:"auth"`
	Audit     AuditConfig      `json:"audit" yaml:"audit"`
	Validator ValidatorConfig  `json:"validator" yaml:"validator"`
	Discovery DiscoveryConfig  `json:"discovery" yaml:"discovery"`
	Health    HealthConfig     `json:"health" yaml:"health"`
	Pool      PoolConfig       `json:"pool" yaml:"pool"`
	WebSocket WebSocketConfig  `json:"websocket" yaml:"websocket"`
	WebUI     WebUIConfig      `json:"webui" yaml:"webui"`
	CORS      CORSConfig       `json:"cors" yaml:"cors"`
	Instances []InstanceConfig `json:"instances" yaml:"instances"`
}

// NewConfig 创建一个带有默认值的配置
func NewConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
			Mode: "release",
		},
		Auth: AuthConfig{
			Enabled: false,
			Token: TokenConfig{
				Secret: "",
				Expire: "24h",
			},
			Whitelist: []string{"/health"},
		},
		Audit: AuditConfig{
			Enabled:       false,
			LogFile:       "/var/log/mystisql/audit.log",
			RetentionDays: 30,
		},
		Validator: ValidatorConfig{
			Enabled:             false,
			DangerousOperations: []string{"DROP", "TRUNCATE", "DELETE_WITHOUT_WHERE", "UPDATE_WITHOUT_WHERE"},
			Whitelist:           []string{},
			Blacklist:           []string{},
		},
		Discovery: DiscoveryConfig{
			Type: "static",
			K8s: K8sDiscoveryConfig{
				Namespace:     "default",
				LabelSelector: "app=database",
				PortMapping: map[string]int{
					"mysql":         3306,
					"postgresql":    5432,
					"oracle":        1521,
					"redis":         6379,
					"sqlite":        0,
					"mssql":         1433,
					"mongodb":       27017,
					"elasticsearch": 9200,
					"clickhouse":    9000,
					"etcd":          2379,
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
		WebSocket: WebSocketConfig{
			Enabled:              true,
			MaxConnections:       100,
			IdleTimeout:          "5m",
			MaxConcurrentQueries: 5,
			AllowedOrigins:       []string{},
			MaxMessageSize:       1048576, // 1MB
		},
		WebUI: WebUIConfig{
			Enabled: true,
			Mode:    "embedded",
		},
		Instances: make([]InstanceConfig, 0),
	}
}
