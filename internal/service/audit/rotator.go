package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type LogRotator struct {
	filePath      string
	retentionDays int
	logger        *zap.Logger
	stopCh        chan struct{}
}

func NewLogRotator(filePath string, retentionDays int, logger *zap.Logger) *LogRotator {
	return &LogRotator{
		filePath:      filePath,
		retentionDays: retentionDays,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}
}

func (lr *LogRotator) Start() {
	go lr.run()
}

func (lr *LogRotator) Stop() {
	close(lr.stopCh)
}

func (lr *LogRotator) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	lastCheckDate := time.Now().Format("2006-01-02")

	for {
		select {
		case <-ticker.C:
			currentDate := time.Now().Format("2006-01-02")
			if currentDate != lastCheckDate {
				if err := lr.RotateOldLogs(); err != nil {
					lr.logger.Error("Failed to rotate logs", zap.Error(err))
				}
				if err := lr.CleanOldLogs(); err != nil {
					lr.logger.Error("Failed to clean old logs", zap.Error(err))
				}
				lastCheckDate = currentDate
			}
		case <-lr.stopCh:
			return
		}
	}
}

func (lr *LogRotator) RotateOldLogs() error {
	if _, err := os.Stat(lr.filePath); os.IsNotExist(err) {
		return nil
	}

	rotatedPath := lr.filePath + "." + time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	if err := os.Rename(lr.filePath, rotatedPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	lr.logger.Info("Rotated audit log file",
		zap.String("old", lr.filePath),
		zap.String("new", rotatedPath),
	)

	return nil
}

func (lr *LogRotator) CleanOldLogs() error {
	dir := filepath.Dir(lr.filePath)
	baseName := filepath.Base(lr.filePath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	cutoffDate := time.Now().AddDate(0, 0, -lr.retentionDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !lr.isAuditLogFile(entry.Name(), baseName) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			lr.logger.Warn("Failed to get file info",
				zap.String("file", entry.Name()),
				zap.Error(err),
			)
			continue
		}

		if info.ModTime().Before(cutoffDate) {
			filePath := filepath.Join(dir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				lr.logger.Warn("Failed to delete old log file",
					zap.String("file", filePath),
					zap.Error(err),
				)
			} else {
				lr.logger.Info("Deleted old audit log file",
					zap.String("file", filePath),
				)
			}
		}
	}

	return nil
}

func (lr *LogRotator) isAuditLogFile(fileName, baseName string) bool {
	if fileName == baseName {
		return false
	}

	prefix := baseName + "."
	if len(fileName) <= len(prefix) {
		return false
	}

	return fileName[:len(prefix)] == prefix
}
