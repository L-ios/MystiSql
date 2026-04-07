package types

import "time"

// DatabaseType 定义支持的数据库类型
type DatabaseType string

const (
	DatabaseTypeMySQL         DatabaseType = "mysql"
	DatabaseTypePostgreSQL    DatabaseType = "postgresql"
	DatabaseTypeOracle        DatabaseType = "oracle"
	DatabaseTypeRedis         DatabaseType = "redis"
	DatabaseTypeSQLite        DatabaseType = "sqlite"
	DatabaseTypeMSSQL         DatabaseType = "mssql"
	DatabaseTypeMongoDB       DatabaseType = "mongodb"
	DatabaseTypeElasticsearch DatabaseType = "elasticsearch"
	DatabaseTypeClickHouse    DatabaseType = "clickhouse"
	DatabaseTypeEtcd          DatabaseType = "etcd"
)

// InstanceStatus 定义实例的健康状态
type InstanceStatus string

const (
	InstanceStatusUnknown   InstanceStatus = "unknown"   // 未知状态
	InstanceStatusHealthy   InstanceStatus = "healthy"   // 健康
	InstanceStatusUnhealthy InstanceStatus = "unhealthy" // 不健康
)

// DatabaseInstance 表示一个数据库实例的完整信息
type DatabaseInstance struct {
	// 基本信息
	Name     string       `json:"name" yaml:"name"`         // 实例名称（唯一标识）
	Type     DatabaseType `json:"type" yaml:"type"`         // 数据库类型
	Host     string       `json:"host" yaml:"host"`         // 主机地址
	Port     int          `json:"port" yaml:"port"`         // 端口号
	Database string       `json:"database" yaml:"database"` // 数据库名（可选）

	// 认证信息
	Username string `json:"username,omitempty" yaml:"username,omitempty"` // 用户名（可选）
	Password string `json:"-" yaml:"password,omitempty"`                  // 密码（不会序列化到 JSON）

	// 状态信息
	Status InstanceStatus `json:"status" yaml:"status"` // 实例状态

	// 元数据
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`           // 标签（K8s 风格）
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"` // 注解

	// 高可用配置
	Role   InstanceRole `json:"role" yaml:"role"`                         // 实例角色（master/slave）
	Master string       `json:"master,omitempty" yaml:"master,omitempty"` // 关联的主库名称（从库专用）
	Weight int          `json:"weight,omitempty" yaml:"weight,omitempty"` // 负载均衡权重

	// 时间戳
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"` // 创建时间
	UpdatedAt time.Time `json:"updatedAt" yaml:"updatedAt"` // 更新时间
}

// NewDatabaseInstance 创建一个新的数据库实例
func NewDatabaseInstance(name string, dbType DatabaseType, host string, port int) *DatabaseInstance {
	now := time.Now()
	return &DatabaseInstance{
		Name:      name,
		Type:      dbType,
		Host:      host,
		Port:      port,
		Status:    InstanceStatusUnknown,
		Role:      InstanceRoleMaster,
		Weight:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetCredentials 设置实例的认证凭据
func (i *DatabaseInstance) SetCredentials(username, password string) {
	i.Username = username
	i.Password = password
	i.UpdatedAt = time.Now()
}

// SetStatus 更新实例状态
func (i *DatabaseInstance) SetStatus(status InstanceStatus) {
	i.Status = status
	i.UpdatedAt = time.Now()
}

// SetDatabase 设置数据库名
func (i *DatabaseInstance) SetDatabase(database string) {
	i.Database = database
	i.UpdatedAt = time.Now()
}

// AddLabel 添加一个标签
func (i *DatabaseInstance) AddLabel(key, value string) {
	if i.Labels == nil {
		i.Labels = make(map[string]string)
	}
	i.Labels[key] = value
	i.UpdatedAt = time.Now()
}

// AddAnnotation 添加一个注解
func (i *DatabaseInstance) AddAnnotation(key, value string) {
	if i.Annotations == nil {
		i.Annotations = make(map[string]string)
	}
	i.Annotations[key] = value
	i.UpdatedAt = time.Now()
}
