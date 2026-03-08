package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

type LogWriter struct {
	filePath string
	file     *os.File
	writer   *bufio.Writer
	logger   *zap.Logger
	channel  chan *AuditLog
	wg       sync.WaitGroup
	stopCh   chan struct{}
	mu       sync.Mutex
}

func NewLogWriter(filePath string, bufferSize int, logger *zap.Logger) (*LogWriter, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	lw := &LogWriter{
		filePath: filePath,
		file:     file,
		writer:   bufio.NewWriterSize(file, 4096),
		logger:   logger,
		channel:  make(chan *AuditLog, bufferSize),
		stopCh:   make(chan struct{}),
	}

	lw.wg.Add(1)
	go lw.processLogs()

	return lw, nil
}

func (lw *LogWriter) Write(log *AuditLog) error {
	select {
	case lw.channel <- log:
		return nil
	default:
		return fmt.Errorf("audit log channel is full, dropping log entry")
	}
}

func (lw *LogWriter) processLogs() {
	defer lw.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case log := <-lw.channel:
			if err := lw.writeLog(log); err != nil {
				lw.logger.Error("Failed to write audit log", zap.Error(err))
			}
		case <-ticker.C:
			lw.mu.Lock()
			if lw.writer != nil {
				if err := lw.writer.Flush(); err != nil {
					lw.logger.Error("Failed to flush audit log", zap.Error(err))
				}
			}
			lw.mu.Unlock()
		case <-lw.stopCh:
			lw.drain()
			return
		}
	}
}

func (lw *LogWriter) writeLog(log *AuditLog) error {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.writer == nil {
		return fmt.Errorf("log writer is closed")
	}

	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	if _, err := lw.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	if _, err := lw.writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

func (lw *LogWriter) drain() {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	for len(lw.channel) > 0 {
		log := <-lw.channel
		if err := lw.writeLog(log); err != nil {
			lw.logger.Error("Failed to write audit log during drain", zap.Error(err))
		}
	}

	if lw.writer != nil {
		if err := lw.writer.Flush(); err != nil {
			lw.logger.Error("Failed to flush audit log on shutdown", zap.Error(err))
		}
	}
}

func (lw *LogWriter) Close() error {
	close(lw.stopCh)
	lw.wg.Wait()

	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.writer != nil {
		if err := lw.writer.Flush(); err != nil {
			lw.logger.Error("Failed to flush on close", zap.Error(err))
		}
		lw.writer = nil
	}

	if lw.file != nil {
		if err := lw.file.Close(); err != nil {
			return fmt.Errorf("failed to close audit log file: %w", err)
		}
		lw.file = nil
	}

	return nil
}

func (lw *LogWriter) Rotate() error {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.writer != nil {
		if err := lw.writer.Flush(); err != nil {
			lw.logger.Error("Failed to flush before rotation", zap.Error(err))
		}
	}

	if lw.file != nil {
		if err := lw.file.Close(); err != nil {
			return fmt.Errorf("failed to close old log file: %w", err)
		}
	}

	file, err := os.OpenFile(lw.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new audit log file: %w", err)
	}

	lw.file = file
	lw.writer = bufio.NewWriterSize(file, 4096)

	return nil
}
