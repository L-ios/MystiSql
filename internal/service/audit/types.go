package audit

import (
	"time"
)

type AuditLog struct {
	Timestamp     time.Time `json:"timestamp"`
	UserID        string    `json:"user_id"`
	ClientIP      string    `json:"client_ip"`
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

func (al *AuditLog) SetQueryInfo(queryType string, rowsAffected int64, execTimeMs int64) {
	al.QueryType = queryType
	al.RowsAffected = rowsAffected
	al.ExecutionTime = execTimeMs
}

func (al *AuditLog) SetSuccess() {
	al.Status = "success"
}

func (al *AuditLog) SetError(errMsg string) {
	al.Status = "error"
	al.ErrorMessage = errMsg
}

func (al *AuditLog) MarkSensitive() {
	al.Sensitive = true
}

func (al *AuditLog) IsSensitive() bool {
	return al.Sensitive
}
