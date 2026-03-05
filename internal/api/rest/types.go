package rest

import (
	"time"

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
	Instance string `json:"instance" binding:"required"` // 实例名称（必填）
	SQL      string `json:"sql" binding:"required"`      // SQL 语句（必填）
	Timeout  int    `json:"timeout,omitempty"`           // 超时时间（秒），可选
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

// ExecResultData 执行结果数据
type ExecResultData struct {
	RowsAffected int64 `json:"rowsAffected"` // 受影响的行数
	LastInsertID int64 `json:"lastInsertId"` // 最后插入的 ID
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
