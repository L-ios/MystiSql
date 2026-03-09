package rest

import (
	"time"

	"MystiSql/internal/connection"
	"MystiSql/pkg/types"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string           `json:"status"`              // 服务状态: "healthy" 或 "unhealthy"
	Timestamp time.Time        `json:"timestamp"`           // 时间戳
	Version   string           `json:"version"`             // 服务版本
	Instances *InstancesHealth `json:"instances,omitempty"` // 实例健康状态（可选）
}

// InstancesHealth 实例健康状态
type InstancesHealth struct {
	Total     int                    `json:"total"`     // 总实例数
	Healthy   int                    `json:"healthy"`   // 健康实例数
	Unhealthy int                    `json:"unhealthy"` // 不健康实例数
	Details   []InstanceHealthDetail `json:"details"`   // 详细信息
}

// InstanceHealthDetail 实例健康详情
type InstanceHealthDetail struct {
	Name   string `json:"name"`   // 实例名称
	Type   string `json:"type"`   // 数据库类型
	Status string `json:"status"` // 状态
}

// InstanceResponse 实例响应（密码脱敏）
type InstanceResponse struct {
	Name     string             `json:"name"`               // 实例名称
	Type     types.DatabaseType `json:"type"`               // 数据库类型
	Host     string             `json:"host"`               // 主机地址
	Port     int                `json:"port"`               // 端口号
	Database string             `json:"database,omitempty"` // 数据库名
	Username string             `json:"username,omitempty"` // 用户名
	Status   string             `json:"status"`             // 实例状态
	Labels   map[string]string  `json:"labels,omitempty"`   // 标签
}

// InstancesListResponse 实例列表响应
type InstancesListResponse struct {
	Total     int                `json:"total"`     // 总数
	Instances []InstanceResponse `json:"instances"` // 实例列表
}

// QueryRequest 查询请求
type QueryRequest struct {
	Instance      string `json:"instance" binding:"required"` // 实例名称（必填）
	SQL           string `json:"sql" binding:"required"`      // SQL 语句（必填）
	Timeout       int    `json:"timeout,omitempty"`           // 超时时间（秒），可选
	TransactionID string `json:"transaction_id,omitempty"`    // 事务 ID（可选，用于事务查询）
}

// QueryResponse 查询响应
type QueryResponse struct {
	Success       bool             `json:"success"`         // 是否成功
	Data          *QueryResultData `json:"data,omitempty"`  // 查询结果数据
	ExecutionTime time.Duration    `json:"executionTime"`   // 执行时间
	Error         *ErrorDetail     `json:"error,omitempty"` // 错误详情
}

// QueryResultData 查询结果数据
type QueryResultData struct {
	Columns  []types.ColumnInfo `json:"columns"`  // 列信息
	Rows     []types.Row        `json:"rows"`     // 行数据
	RowCount int                `json:"rowCount"` // 行数
}

// ExecResponse 执行响应
type ExecResponse struct {
	Success       bool            `json:"success"`         // 是否成功
	Data          *ExecResultData `json:"data,omitempty"`  // 执行结果数据
	ExecutionTime time.Duration   `json:"executionTime"`   // 执行时间
	Error         *ErrorDetail    `json:"error,omitempty"` // 错误详情
}

// ExecRequest 执行请求
type ExecRequest struct {
	Instance      string `json:"instance" binding:"required"` // 实例名称（必填）
	SQL           string `json:"sql" binding:"required"`      // SQL 语句（必填）
	Timeout       int    `json:"timeout,omitempty"`           // 超时时间（秒），可选
	TransactionID string `json:"transaction_id,omitempty"`    // 事务 ID（可选，用于事务执行）
}

// ExecResultData 执行结果数据
type ExecResultData struct {
	AffectedRows int64 `json:"affectedRows"` // 受影响的行数
	LastInsertID int64 `json:"lastInsertId"` // 最后插入的 ID
}

// InstanceHealthResponse 实例健康状态响应
type InstanceHealthResponse struct {
	Instance  string    `json:"instance"`  // 实例名称
	Status    string    `json:"status"`    // 健康状态
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// PoolStatsResponse 连接池统计信息响应
type PoolStatsResponse struct {
	Instance  string               `json:"instance"`  // 实例名称
	Stats     connection.PoolStats `json:"stats"`     // 连接池统计信息
	Timestamp time.Time            `json:"timestamp"` // 时间戳
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Success bool         `json:"success"` // 是否成功（总是 false）
	Error   *ErrorDetail `json:"error"`   // 错误详情
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Code    string `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误消息
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}

// ToInstanceResponse 将 DatabaseInstance 转换为 InstanceResponse（脱敏密码）
func ToInstanceResponse(instance *types.DatabaseInstance) InstanceResponse {
	return InstanceResponse{
		Name:     instance.Name,
		Type:     instance.Type,
		Host:     instance.Host,
		Port:     instance.Port,
		Database: instance.Database,
		Username: instance.Username,
		Status:   string(instance.Status),
		Labels:   instance.Labels,
	}
}

// GenerateTokenRequest 生成 Token 请求
type GenerateTokenRequest struct {
	UserID string `json:"user_id" binding:"required"` // 用户 ID（必填）
	Role   string `json:"role" binding:"required"`    // 角色（必填）
}

// GenerateTokenResponse 生成 Token 响应
type GenerateTokenResponse struct {
	Success bool         `json:"success"`         // 是否成功
	Data    *TokenData   `json:"data,omitempty"`  // Token 数据
	Error   *ErrorDetail `json:"error,omitempty"` // 错误详情
}

// TokenData Token 数据
type TokenData struct {
	Token     string    `json:"token"`     // JWT Token
	TokenID   string    `json:"tokenId"`   // Token ID
	ExpiresAt time.Time `json:"expiresAt"` // 过期时间
	IssuedAt  time.Time `json:"issuedAt"`  // 签发时间
	UserID    string    `json:"userId"`    // 用户 ID
	Role      string    `json:"role"`      // 角色
}

// RevokeTokenRequest 撤销 Token 请求
type RevokeTokenRequest struct {
	Token string `json:"token" binding:"required"` // 要撤销的 Token（必填）
}

// RevokeTokenResponse 撤销 Token 响应
type RevokeTokenResponse struct {
	Success bool         `json:"success"`         // 是否成功
	Message string       `json:"message"`         // 消息
	Error   *ErrorDetail `json:"error,omitempty"` // 错误详情
}

// TokenInfoResponse Token 信息响应
type TokenInfoResponse struct {
	Success   bool         `json:"success"`         // 是否成功
	UserID    string       `json:"userId"`          // 用户 ID
	Role      string       `json:"role"`            // 角色
	TokenID   string       `json:"tokenId"`         // Token ID
	ExpiresAt time.Time    `json:"expiresAt"`       // 过期时间
	IssuedAt  time.Time    `json:"issuedAt"`        // 签发时间
	Error     *ErrorDetail `json:"error,omitempty"` // 错误详情
}

// TokensListResponse Token 列表响应
type TokensListResponse struct {
	Success       bool               `json:"success"`         // 是否成功
	RevokedTokens []RevokedTokenInfo `json:"revokedTokens"`   // 已撤销的 Token 列表
	Error         *ErrorDetail       `json:"error,omitempty"` // 错误详情
}

// RevokedTokenInfo 已撤销的 Token 信息
type RevokedTokenInfo struct {
	Token     string    `json:"token"`     // Token（脱敏）
	Reason    string    `json:"reason"`    // 撤销原因
	RevokedAt time.Time `json:"revokedAt"` // 撤销时间
}
