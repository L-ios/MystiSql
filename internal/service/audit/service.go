package audit

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type AuditService struct {
	writer   *LogWriter
	rotator  *LogRotator
	logger   *zap.Logger
	enabled  bool
	logLevel string
	mu       sync.RWMutex
}

type AuditConfig struct {
	Enabled       bool
	LogFile       string
	BufferSize    int
	RetentionDays int

	// LogLevel 日志级别过滤：
	// "all"（默认）记录全部；"sensitive" 仅记录敏感操作；
	// "write" 记录写操作与敏感操作；"ddl" 仅记录 DDL。
	LogLevel string
}

// 支持的审计日志级别
const (
	AuditLogLevelAll       = "all"
	AuditLogLevelSensitive = "sensitive"
	AuditLogLevelWrite     = "write"
	AuditLogLevelDDL       = "ddl"
)

func NewAuditService(config *AuditConfig, logger *zap.Logger) (*AuditService, error) {
	if !config.Enabled {
		return &AuditService{
			enabled: false,
			logger:  logger,
		}, nil
	}

	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}

	if config.RetentionDays <= 0 {
		config.RetentionDays = 30
	}
	if config.LogLevel == "" {
		config.LogLevel = AuditLogLevelAll
	}

	writer, err := NewLogWriter(config.LogFile, config.BufferSize, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create log writer: %w", err)
	}

	rotator := NewLogRotator(config.LogFile, config.RetentionDays, writer, logger)
	rotator.Start()

	return &AuditService{
		writer:   writer,
		rotator:  rotator,
		logger:   logger,
		enabled:  true,
		logLevel: config.LogLevel,
	}, nil
}

func (s *AuditService) Log(ctx context.Context, log *AuditLog) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.enabled {
		return nil
	}

	if s.writer == nil {
		return fmt.Errorf("audit log writer is not initialized")
	}

	if !s.shouldLog(log) {
		return nil
	}

	return s.writer.Write(log)
}

// shouldLog 根据日志级别判断是否记录该条日志。
func (s *AuditService) shouldLog(log *AuditLog) bool {
	switch s.logLevel {
	case AuditLogLevelSensitive:
		return log.Sensitive
	case AuditLogLevelWrite:
		return log.Sensitive || log.QueryType == "INSERT" || log.QueryType == "UPDATE" ||
			log.QueryType == "DELETE" || log.QueryType == "EXECUTE"
	case AuditLogLevelDDL:
		return log.QueryType == "CREATE" || log.QueryType == "ALTER" ||
			log.QueryType == "DROP" || log.QueryType == "TRUNCATE"
	default: // AuditLogLevelAll
		return true
	}
}

func (s *AuditService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.rotator != nil {
		s.rotator.Stop()
	}

	if s.writer != nil {
		return s.writer.Close()
	}

	return nil
}

func (s *AuditService) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

func (s *AuditService) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
	s.logger.Info("Audit logging enabled")
}

func (s *AuditService) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
	s.logger.Info("Audit logging disabled")
}
