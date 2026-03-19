package audit

import (
	"time"
)

// AuditLog 审计日志结构
type AuditLog struct {
	Timestamp     time.Time `json:"timestamp"`
	UserID        string    `json:"user_id"`
	ClientIP      string    `json:"client_ip"`
	SessionID     string    `json:"session_id,omitempty"`
	UserAgent     string    `json:"user_agent,omitempty"`
	Instance      string    `json:"instance"`
	Database      string    `json:"database"`
	Query         string    `json:"query"`
	QueryType     string    `json:"query_type"`
	RowsAffected  int64     `json:"rows_affected"`
	ExecutionTime int64     `json:"execution_time_ms"`
	Status        string    `json:"status"`
	Sensitive     bool      `json:"sensitive,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// NewAuditLog 创建新的审计日志
func NewAuditLog(userID, clientIP, instance, database, query string) *AuditLog {
	return &AuditLog{
		Timestamp: time.Now(),
		UserID:    userID,
		ClientIP:  clientIP,
		Instance:  instance,
		Database:  database,
		Query:     query,
		Status:    "pending",
	}
}

// SetQueryInfo 设置查询信息
func (al *AuditLog) SetQueryInfo(queryType string, rowsAffected int64, execTimeMs int64) {
	al.QueryType = queryType
	al.RowsAffected = rowsAffected
	al.ExecutionTime = execTimeMs
}

// SetSuccess 设置成功状态
func (al *AuditLog) SetSuccess() {
	al.Status = "success"
}

// SetError 设置错误状态
func (al *AuditLog) SetError(errMsg string) {
	al.Status = "error"
	al.ErrorMessage = errMsg
}

// MarkSensitive 标记为敏感操作
func (al *AuditLog) MarkSensitive() {
	al.Sensitive = true
}

// IsSensitive 检查是否为敏感操作
func (al *AuditLog) IsSensitive() bool {
	return al.Sensitive
}

// SetSessionID 设置会话ID
func (al *AuditLog) SetSessionID(sessionID string) {
	al.SessionID = sessionID
}

// SetUserAgent 设置用户代理
func (al *AuditLog) SetUserAgent(userAgent string) {
	al.UserAgent = userAgent
}

// AuditStats 审计统计信息
type AuditStats struct {
	TotalQueries     int64            `json:"totalQueries"`
	SuccessCount     int64            `json:"successCount"`
	ErrorCount       int64            `json:"errorCount"`
	SensitiveCount   int64            `json:"sensitiveCount"`
	AvgExecutionTime int64            `json:"avgExecutionTimeMs"`
	TopUsers         []UserStats      `json:"topUsers"`
	TopInstances     []InstanceStats  `json:"topInstances"`
	QueryTypeDist    map[string]int64 `json:"queryTypeDistribution"`
}

// UserStats 用户统计
type UserStats struct {
	UserID string `json:"userId"`
	Count  int64  `json:"count"`
}

// InstanceStats 实例统计
type InstanceStats struct {
	Instance string `json:"instance"`
	Count    int64  `json:"count"`
}
