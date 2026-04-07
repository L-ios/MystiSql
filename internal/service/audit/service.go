package audit

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type AuditService struct {
	writer  *LogWriter
	rotator *LogRotator
	logger  *zap.Logger
	enabled bool
	mu      sync.RWMutex
}

type AuditConfig struct {
	Enabled       bool
	LogFile       string
	BufferSize    int
	RetentionDays int
}

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

	writer, err := NewLogWriter(config.LogFile, config.BufferSize, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create log writer: %w", err)
	}

	rotator := NewLogRotator(config.LogFile, config.RetentionDays, writer, logger)
	rotator.Start()

	return &AuditService{
		writer:  writer,
		rotator: rotator,
		logger:  logger,
		enabled: true,
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

	return s.writer.Write(log)
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
