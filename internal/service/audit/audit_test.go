package audit

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewAuditLog(t *testing.T) {
	userID := "test-user"
	clientIP := "192.168.1.1"
	instance := "test-instance"
	database := "test-db"
	query := "SELECT * FROM users"

	log := NewAuditLog(userID, clientIP, instance, database, query)

	if log.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, log.UserID)
	}
	if log.ClientIP != clientIP {
		t.Errorf("Expected ClientIP %s, got %s", clientIP, log.ClientIP)
	}
	if log.Instance != instance {
		t.Errorf("Expected Instance %s, got %s", instance, log.Instance)
	}
	if log.Database != database {
		t.Errorf("Expected Database %s, got %s", database, log.Database)
	}
	if log.Query != query {
		t.Errorf("Expected Query %s, got %s", query, log.Query)
	}
	if log.Status != "pending" {
		t.Errorf("Expected Status 'pending', got %s", log.Status)
	}
	if log.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestAuditLog_SetQueryInfo(t *testing.T) {
	log := &AuditLog{}
	queryType := "SELECT"
	rowsAffected := int64(100)
	execTimeMs := int64(50)

	log.SetQueryInfo(queryType, rowsAffected, execTimeMs)

	if log.QueryType != queryType {
		t.Errorf("Expected QueryType %s, got %s", queryType, log.QueryType)
	}
	if log.RowsAffected != rowsAffected {
		t.Errorf("Expected RowsAffected %d, got %d", rowsAffected, log.RowsAffected)
	}
	if log.ExecutionTime != execTimeMs {
		t.Errorf("Expected ExecutionTime %d, got %d", execTimeMs, log.ExecutionTime)
	}
}

func TestAuditLog_SetSuccess(t *testing.T) {
	log := &AuditLog{Status: "pending"}
	log.SetSuccess()

	if log.Status != "success" {
		t.Errorf("Expected Status 'success', got %s", log.Status)
	}
}

func TestAuditLog_SetError(t *testing.T) {
	log := &AuditLog{Status: "pending"}
	errMsg := "connection failed"

	log.SetError(errMsg)

	if log.Status != "error" {
		t.Errorf("Expected Status 'error', got %s", log.Status)
	}
	if log.ErrorMessage != errMsg {
		t.Errorf("Expected ErrorMessage %s, got %s", errMsg, log.ErrorMessage)
	}
}

func TestAuditLog_MarkSensitive(t *testing.T) {
	log := &AuditLog{}
	log.MarkSensitive()

	if !log.Sensitive {
		t.Error("Expected Sensitive to be true")
	}
}

func TestNewLogWriter(t *testing.T) {
	tmpFile := "/tmp/test-audit.log"
	defer os.Remove(tmpFile)

	logger := zap.NewNop()
	writer, err := NewLogWriter(tmpFile, 100, logger)
	if err != nil {
		t.Fatalf("Failed to create log writer: %v", err)
	}
	defer writer.Close()

	if writer == nil {
		t.Fatal("Writer should not be nil")
	}
}

func TestLogWriter_Write(t *testing.T) {
	tmpFile := "/tmp/test-audit-write.log"
	defer os.Remove(tmpFile)

	logger := zap.NewNop()
	writer, err := NewLogWriter(tmpFile, 100, logger)
	if err != nil {
		t.Fatalf("Failed to create log writer: %v", err)
	}
	defer writer.Close()

	log := NewAuditLog("user1", "192.168.1.1", "instance1", "db1", "SELECT 1")
	log.SetQueryInfo("SELECT", 1, 10)
	log.SetSuccess()

	err = writer.Write(log)
	if err != nil {
		t.Errorf("Failed to write log: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

func TestNewAuditService(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		config  *AuditConfig
		wantErr bool
	}{
		{
			name: "disabled audit",
			config: &AuditConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled with valid config",
			config: &AuditConfig{
				Enabled:       true,
				LogFile:       "/tmp/test-audit-service.log",
				BufferSize:    100,
				RetentionDays: 30,
			},
			wantErr: false,
		},
		{
			name: "zero buffer size uses default",
			config: &AuditConfig{
				Enabled:       true,
				LogFile:       "/tmp/test-audit-service2.log",
				BufferSize:    0,
				RetentionDays: 30,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewAuditService(tt.config, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuditService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if service != nil {
				defer service.Close()
			}
		})
	}
}

func TestAuditService_Log(t *testing.T) {
	tmpFile := "/tmp/test-audit-service-log.log"
	defer os.Remove(tmpFile)

	logger := zap.NewNop()
	config := &AuditConfig{
		Enabled:       true,
		LogFile:       tmpFile,
		BufferSize:    10,
		RetentionDays: 30,
	}

	service, err := NewAuditService(config, logger)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer service.Close()

	log := NewAuditLog("user1", "192.168.1.1", "instance1", "db1", "SELECT 1")
	log.SetQueryInfo("SELECT", 1, 10)
	log.SetSuccess()

	err = service.Log(context.Background(), log)
	if err != nil {
		t.Errorf("Failed to log: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
}

func TestAuditService_Disabled(t *testing.T) {
	logger := zap.NewNop()
	config := &AuditConfig{
		Enabled: false,
	}

	service, err := NewAuditService(config, logger)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}

	log := NewAuditLog("user1", "192.168.1.1", "instance1", "db1", "SELECT 1")

	err = service.Log(context.Background(), log)
	if err != nil {
		t.Errorf("Disabled service should not return error: %v", err)
	}
}

func TestAuditService_EnableDisable(t *testing.T) {
	logger := zap.NewNop()
	config := &AuditConfig{
		Enabled: false,
	}

	service, err := NewAuditService(config, logger)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}

	if service.IsEnabled() {
		t.Error("Service should be disabled initially")
	}

	service.Enable()
	if !service.IsEnabled() {
		t.Error("Service should be enabled after Enable()")
	}

	service.Disable()
	if service.IsEnabled() {
		t.Error("Service should be disabled after Disable()")
	}
}
