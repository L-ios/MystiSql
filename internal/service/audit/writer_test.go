package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestLogWriter_New(t *testing.T) {
	logger := zap.NewNop()
	path := filepath.Join(t.TempDir(), "audit.log")
	lw, err := NewLogWriter(path, 100, logger)
	if err != nil {
		t.Fatalf("NewLogWriter error: %v", err)
	}
	if lw == nil {
		t.Fatal("NewLogWriter returned nil")
	}
	if lw.filePath != path {
		t.Errorf("filePath = %q, want %q", lw.filePath, path)
	}
	lw.Close()
}

func TestNewLogWriter_BadDir(t *testing.T) {
	logger := zap.NewNop()
	_, err := NewLogWriter("/dev/null/impossible/path/audit.log", 100, logger)
	if err == nil {
		t.Error("expected error for impossible directory")
	}
}

func TestLogWriter_WriteSuccess(t *testing.T) {
	logger := zap.NewNop()
	path := filepath.Join(t.TempDir(), "audit.log")
	lw, err := NewLogWriter(path, 100, logger)
	if err != nil {
		t.Fatalf("NewLogWriter error: %v", err)
	}
	defer lw.Close()

	log := NewAuditLog("user", "127.0.0.1", "inst", "db", "SELECT 1")
	err = lw.Write(log)
	if err != nil {
		t.Errorf("Write error: %v", err)
	}
}

func TestLogWriter_WriteChannelFull(t *testing.T) {
	logger := zap.NewNop()
	path := filepath.Join(t.TempDir(), "audit.log")
	// Buffer size 1, but we need to fill it before the goroutine drains it.
	// Use a very large number of writes — at least one will fail.
	lw, err := NewLogWriter(path, 1, logger)
	if err != nil {
		t.Fatalf("NewLogWriter error: %v", err)
	}
	defer lw.Close()

	log := NewAuditLog("user", "127.0.0.1", "inst", "db", "SELECT 1")
	var failed bool
	for i := 0; i < 100; i++ {
		if err := lw.Write(log); err != nil {
			failed = true
			break
		}
	}
	if !failed {
		t.Error("expected at least one Write to fail with full channel")
	}
}

func TestLogWriter_Close(t *testing.T) {
	logger := zap.NewNop()
	path := filepath.Join(t.TempDir(), "audit.log")
	lw, err := NewLogWriter(path, 100, logger)
	if err != nil {
		t.Fatalf("NewLogWriter error: %v", err)
	}

	log := NewAuditLog("user", "127.0.0.1", "inst", "db", "SELECT 1")
	lw.Write(log)

	err = lw.Close()
	if err != nil {
		t.Errorf("Close error: %v", err)
	}

	if lw.writer != nil {
		t.Error("writer should be nil after Close")
	}
	if lw.file != nil {
		t.Error("file should be nil after Close")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if !strings.Contains(string(data), "SELECT 1") {
		t.Errorf("log file should contain query, got: %s", string(data))
	}
}

func TestLogWriter_Close_WriterClosed(t *testing.T) {
	logger := zap.NewNop()
	path := filepath.Join(t.TempDir(), "audit.log")
	lw, err := NewLogWriter(path, 100, logger)
	if err != nil {
		t.Fatalf("NewLogWriter error: %v", err)
	}

	log := NewAuditLog("user", "127.0.0.1", "inst", "db", "SELECT 1")
	lw.Write(log)
	lw.Close()

	time.Sleep(50 * time.Millisecond)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if !strings.Contains(string(data), "SELECT 1") {
		t.Errorf("log written before Close should be flushed, got: %s", string(data))
	}
}

func TestLogWriter_Rotate(t *testing.T) {
	logger := zap.NewNop()
	path := filepath.Join(t.TempDir(), "audit.log")
	lw, err := NewLogWriter(path, 100, logger)
	if err != nil {
		t.Fatalf("NewLogWriter error: %v", err)
	}

	log := NewAuditLog("user", "127.0.0.1", "inst", "db", "SELECT 1")
	lw.Write(log)

	time.Sleep(50 * time.Millisecond)

	err = lw.Rotate()
	if err != nil {
		t.Errorf("Rotate error: %v", err)
	}

	log2 := NewAuditLog("user", "127.0.0.1", "inst", "db", "SELECT 2")
	lw.Write(log2)

	time.Sleep(50 * time.Millisecond)

	lw.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "SELECT 1") || !strings.Contains(content, "SELECT 2") {
		t.Errorf("rotated file should contain both queries, got: %s", content)
	}
}
