package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

// --- RotateOldLogs ---

func TestLogRotator_RotateOldLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "audit.log")

	// Create the initial log file with some content
	if err := os.WriteFile(logFile, []byte("line1\nline2\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	logger := zap.NewNop()
	writer, err := NewLogWriter(logFile, 100, logger)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	defer writer.Close()

	rotator := NewLogRotator(logFile, 30, writer, logger)
	defer rotator.Stop()

	if err := rotator.RotateOldLogs(); err != nil {
		t.Fatalf("RotateOldLogs error: %v", err)
	}

	// Original file should have been renamed (to yesterday's date backup)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	rotatedPath := logFile + "." + yesterday

	if _, err := os.Stat(rotatedPath); os.IsNotExist(err) {
		t.Errorf("rotated file %q should exist", rotatedPath)
	}

	// Writer.Rotate() should have recreated the log file
	// Give the writer a moment to settle
	time.Sleep(50 * time.Millisecond)
}

func TestLogRotator_RotateOldLogs_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "nonexistent.log")

	logger := zap.NewNop()
	rotator := NewLogRotator(logFile, 30, nil, logger)

	// Should return nil when file doesn't exist
	if err := rotator.RotateOldLogs(); err != nil {
		t.Errorf("RotateOldLogs should return nil for non-existent file, got: %v", err)
	}
}

func TestLogRotator_RotateOldLogs_NilWriter(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "audit.log")

	if err := os.WriteFile(logFile, []byte("data\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	logger := zap.NewNop()
	rotator := NewLogRotator(logFile, 30, nil, logger)

	if err := rotator.RotateOldLogs(); err != nil {
		t.Fatalf("RotateOldLogs with nil writer error: %v", err)
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	rotatedPath := logFile + "." + yesterday
	if _, err := os.Stat(rotatedPath); os.IsNotExist(err) {
		t.Errorf("rotated file %q should exist even with nil writer", rotatedPath)
	}
}

// --- CleanOldLogs ---

func TestLogRotator_CleanOldLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "audit.log")

	// Create the base log file
	if err := os.WriteFile(logFile, []byte("current\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	// Create old audit log files (should be deleted with retention=7)
	oldFiles := []string{
		"audit.log.2024-01-01",
		"audit.log.2024-01-02",
		"audit.log.2024-06-15",
	}
	for _, name := range oldFiles {
		path := filepath.Join(tmpDir, name)
		// Set mod time to 10 days ago so they're older than retention
		if err := os.WriteFile(path, []byte("old\n"), 0644); err != nil {
			t.Fatalf("failed to create old file %s: %v", name, err)
		}
		oldTime := time.Now().AddDate(0, 0, -10)
		if err := os.Chtimes(path, oldTime, oldTime); err != nil {
			t.Fatalf("failed to set mod time for %s: %v", name, err)
		}
	}

	// Create recent audit log files (should NOT be deleted)
	recentFiles := []string{
		"audit.log." + time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		"audit.log." + time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
	}
	for _, name := range recentFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("recent\n"), 0644); err != nil {
			t.Fatalf("failed to create recent file %s: %v", name, err)
		}
	}

	logger := zap.NewNop()
	rotator := NewLogRotator(logFile, 7, nil, logger)

	if err := rotator.CleanOldLogs(); err != nil {
		t.Fatalf("CleanOldLogs error: %v", err)
	}

	// Old files should be deleted
	for _, name := range oldFiles {
		path := filepath.Join(tmpDir, name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("old file %q should have been deleted", name)
		}
	}

	// Recent files should still exist
	for _, name := range recentFiles {
		path := filepath.Join(tmpDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("recent file %q should not have been deleted", name)
		}
	}

	// Base log file should still exist
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("base log file should not be deleted")
	}
}

func TestLogRotator_CleanOldLogs_DirNotExist(t *testing.T) {
	logger := zap.NewNop()
	rotator := NewLogRotator("/nonexistent/dir/audit.log", 7, nil, logger)

	err := rotator.CleanOldLogs()
	if err == nil {
		t.Error("CleanOldLogs should return error for non-existent directory")
	}
}

func TestLogRotator_CleanOldLogs_NonAuditFilesIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "audit.log")

	if err := os.WriteFile(logFile, []byte("current\n"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	// Create a non-audit file that is old
	otherFile := filepath.Join(tmpDir, "other.log.2024-01-01")
	if err := os.WriteFile(otherFile, []byte("other\n"), 0644); err != nil {
		t.Fatalf("failed to create other file: %v", err)
	}
	oldTime := time.Now().AddDate(0, 0, -10)
	if err := os.Chtimes(otherFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set mod time: %v", err)
	}

	logger := zap.NewNop()
	rotator := NewLogRotator(logFile, 7, nil, logger)

	if err := rotator.CleanOldLogs(); err != nil {
		t.Fatalf("CleanOldLogs error: %v", err)
	}

	// Non-audit file should NOT be deleted
	if _, err := os.Stat(otherFile); os.IsNotExist(err) {
		t.Error("non-audit file should not be deleted")
	}
}

// --- isAuditLogFile ---

func TestLogRotator_isAuditLogFile(t *testing.T) {
	logger := zap.NewNop()
	rotator := NewLogRotator("audit.log", 30, nil, logger)

	tests := []struct {
		fileName string
		baseName string
		expected bool
	}{
		{"audit.log.2024-01-01", "audit.log", true},
		{"audit.log.2024-01-01.gz", "audit.log", true},
		{"audit.log", "audit.log", false},            // same as base
		{"other.log.2024-01-01", "audit.log", false}, // different prefix
		{"audit.log.", "audit.log", false},           // too short after prefix
		{"", "audit.log", false},                     // empty
		{"audit", "audit.log", false},                // no dot
		{"AUDIT.LOG.2024-01-01", "audit.log", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			result := rotator.isAuditLogFile(tt.fileName, tt.baseName)
			if result != tt.expected {
				t.Errorf("isAuditLogFile(%q, %q) = %v, want %v",
					tt.fileName, tt.baseName, result, tt.expected)
			}
		})
	}
}
