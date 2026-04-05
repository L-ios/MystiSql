package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewTokenBlacklistWithConfig_Basic(t *testing.T) {
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		CleanupInterval: time.Minute,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTokenBlacklistWithConfig error: %v", err)
	}
	defer bl.Stop()

	if bl == nil {
		t.Fatal("expected non-nil blacklist")
	}
	if bl.Size() != 0 {
		t.Errorf("new blacklist should be empty, got %d", bl.Size())
	}
}

func TestNewTokenBlacklistWithConfig_NilLogger(t *testing.T) {
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		CleanupInterval: time.Minute,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, nil)
	if err != nil {
		t.Fatalf("NewTokenBlacklistWithConfig with nil logger error: %v", err)
	}
	defer bl.Stop()

	bl.Add("token1", "reason")
	if !bl.Contains("token1") {
		t.Error("should contain token1")
	}
}

func TestNewTokenBlacklistWithConfig_DefaultTTL(t *testing.T) {
	cfg := BlacklistConfig{
		CleanupInterval: time.Minute,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	if bl.config.TTL != 24*time.Hour {
		t.Errorf("default TTL = %v, want %v", bl.config.TTL, 24*time.Hour)
	}
}

func TestNewTokenBlacklistWithConfig_DefaultCleanupInterval(t *testing.T) {
	cfg := BlacklistConfig{
		TTL: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	if bl.config.CleanupInterval != time.Minute {
		t.Errorf("default CleanupInterval = %v, want %v", bl.config.CleanupInterval, time.Minute)
	}
}

func TestNewTokenBlacklistWithConfig_FilePersistence(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "blacklist.jsonl")
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        tmpFile,
		CleanupInterval: 10 * time.Millisecond,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	bl.Add("token1", "reason1")
	bl.Add("token2", "reason2")

	if !bl.Contains("token1") {
		t.Error("should contain token1")
	}
	if !bl.Contains("token2") {
		t.Error("should contain token2")
	}

	bl.Stop()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if len(data) == 0 {
		t.Error("file should not be empty after adds")
	}

	bl2, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	defer bl2.Stop()

	if !bl2.Contains("token1") {
		t.Error("reloaded blacklist should contain token1")
	}
	if !bl2.Contains("token2") {
		t.Error("reloaded blacklist should contain token2")
	}
	if bl2.Size() != 2 {
		t.Errorf("reloaded size = %d, want 2", bl2.Size())
	}
}

func TestNewTokenBlacklistWithConfig_FileLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        tmpDir,
		CleanupInterval: time.Minute,
	}
	_, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err == nil {
		t.Fatal("expected error when FilePath is a directory")
	}
}

func TestTokenBlacklist_Stop(t *testing.T) {
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		CleanupInterval: 10 * time.Millisecond,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	bl.Add("token1", "reason")
	bl.Stop()

	if !bl.Contains("token1") {
		t.Error("should still contain token1 after Stop")
	}
}

func TestTokenBlacklist_Stop_IdempotentPanic(t *testing.T) {
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		CleanupInterval: time.Minute,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	bl.Stop()
}

func TestTokenBlacklist_CleanupRemovesExpired(t *testing.T) {
	cfg := BlacklistConfig{
		TTL:             50 * time.Millisecond,
		CleanupInterval: 30 * time.Millisecond,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	bl.Add("token1", "expires soon")

	if !bl.Contains("token1") {
		t.Fatal("token1 should be present immediately after Add")
	}

	time.Sleep(80 * time.Millisecond)

	if bl.Contains("token1") {
		t.Error("token1 should be expired")
	}

	time.Sleep(50 * time.Millisecond)

	if bl.Size() != 0 {
		t.Errorf("size after cleanup = %d, want 0", bl.Size())
	}
}

func TestTokenBlacklist_CleanupWithFilePersistence(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "blacklist.jsonl")
	cfg := BlacklistConfig{
		TTL:             50 * time.Millisecond,
		FilePath:        tmpFile,
		CleanupInterval: 30 * time.Millisecond,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	bl.Add("token1", "expires soon")

	time.Sleep(120 * time.Millisecond)

	bl.Stop()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("file should be empty after cleanup, got %q", string(data))
	}
}

func TestTokenBlacklist_Contains_ExpiredToken(t *testing.T) {
	bl := NewTokenBlacklist()

	bl.mu.Lock()
	bl.items["expired-token"] = &BlacklistItem{
		Token:     "expired-token",
		RevokedAt: time.Now().Add(-2 * time.Hour),
		Reason:    "test",
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	bl.mu.Unlock()

	if bl.Contains("expired-token") {
		t.Error("should not contain expired token")
	}
}

func TestTokenBlacklist_Add_WithFilePersistence(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "blacklist.jsonl")
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        tmpFile,
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	bl.Add("file-token", "persisted reason")

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var item BlacklistItem
	if err := json.Unmarshal(data[:len(data)-1], &item); err != nil {
		t.Fatalf("failed to parse JSON line: %v", err)
	}
	if item.Token != "file-token" {
		t.Errorf("file token = %q, want %q", item.Token, "file-token")
	}
	if item.Reason != "persisted reason" {
		t.Errorf("file reason = %q, want %q", item.Reason, "persisted reason")
	}
	if item.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set when TTL > 0")
	}
}

func TestTokenBlacklist_Add_FileWriteError(t *testing.T) {
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        "/nonexistent/dir/blacklist.jsonl",
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	bl.Add("token1", "reason")

	if !bl.Contains("token1") {
		t.Error("token should be in memory even if file write fails")
	}
}

func TestTokenBlacklist_Clear_WithFilePersistence(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "blacklist.jsonl")
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        tmpFile,
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	bl.Add("token1", "reason1")
	bl.Add("token2", "reason2")

	bl.Clear()

	if bl.Size() != 0 {
		t.Errorf("size after Clear = %d, want 0", bl.Size())
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("file should be empty after Clear, got %q", string(data))
	}
}

func TestTokenBlacklist_LoadFromFile_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        filepath.Join(tmpDir, "does-not-exist.jsonl"),
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("should succeed when file doesn't exist: %v", err)
	}
	defer bl.Stop()

	if bl.Size() != 0 {
		t.Errorf("size = %d, want 0", bl.Size())
	}
}

func TestTokenBlacklist_LoadFromFile_MalformedEntry(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "blacklist.jsonl")

	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	validItem1 := BlacklistItem{
		Token:     "valid-token-1",
		RevokedAt: time.Now().Add(-time.Hour),
		Reason:    "test",
		ExpiresAt: time.Now().Add(23 * time.Hour),
	}
	line1, _ := json.Marshal(validItem1)
	f.WriteString(string(line1) + "\n")
	f.WriteString("this-is-not-json\n")
	f.WriteString("\n")

	validItem2 := BlacklistItem{
		Token:     "valid-token-2",
		RevokedAt: time.Now().Add(-time.Hour),
		Reason:    "test",
		ExpiresAt: time.Now().Add(23 * time.Hour),
	}
	line3, _ := json.Marshal(validItem2)
	f.WriteString(string(line3) + "\n")
	f.Close()

	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        tmpFile,
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	if bl.Size() != 2 {
		t.Errorf("size = %d, want 2", bl.Size())
	}
	if !bl.Contains("valid-token-1") {
		t.Error("should contain valid-token-1")
	}
	if !bl.Contains("valid-token-2") {
		t.Error("should contain valid-token-2")
	}
}

func TestTokenBlacklist_LoadFromFile_ExpiredSkipped(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "blacklist.jsonl")

	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	expiredItem := BlacklistItem{
		Token:     "expired-token",
		RevokedAt: time.Now().Add(-48 * time.Hour),
		Reason:    "old",
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}
	line, _ := json.Marshal(expiredItem)
	f.WriteString(string(line) + "\n")

	validItem := BlacklistItem{
		Token:     "valid-token",
		RevokedAt: time.Now().Add(-time.Hour),
		Reason:    "recent",
		ExpiresAt: time.Now().Add(23 * time.Hour),
	}
	line2, _ := json.Marshal(validItem)
	f.WriteString(string(line2) + "\n")
	f.Close()

	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        tmpFile,
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	if bl.Size() != 1 {
		t.Errorf("size = %d, want 1", bl.Size())
	}
	if bl.Contains("expired-token") {
		t.Error("should skip expired token during load")
	}
	if !bl.Contains("valid-token") {
		t.Error("should contain valid token")
	}
}

func TestTokenBlacklist_AppendToFile_Format(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "blacklist.jsonl")
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        tmpFile,
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	bl.Add("tok1", "reason1")
	bl.Add("tok2", "reason2")

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}

	lines := splitLines(string(data))
	if len(lines) != 2 {
		t.Fatalf("line count = %d, want 2", len(lines))
	}

	for i, line := range lines {
		var item BlacklistItem
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestTokenBlacklist_RewriteFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "blacklist.jsonl")
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        tmpFile,
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	bl.Add("tok1", "reason1")
	bl.Add("tok2", "reason2")

	bl.Clear()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("file should be empty after Clear, got %q", string(data))
	}
}

func TestTokenBlacklist_RewriteFile_PreservesValidItems(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "blacklist.jsonl")
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        tmpFile,
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	bl.Add("tok1", "reason1")
	bl.Add("tok2", "reason2")
	bl.Add("tok3", "reason3")

	bl.Stop()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	lines := splitLines(string(data))
	if len(lines) != 3 {
		t.Errorf("line count = %d, want 3", len(lines))
	}
}

func TestTokenBlacklist_RemoveExpired_DirectCall(t *testing.T) {
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	bl.mu.Lock()
	bl.items["expired-token"] = &BlacklistItem{
		Token:     "expired-token",
		RevokedAt: time.Now().Add(-2 * time.Hour),
		Reason:    "test",
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	bl.items["valid-token"] = &BlacklistItem{
		Token:     "valid-token",
		RevokedAt: time.Now(),
		Reason:    "test",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	bl.mu.Unlock()

	bl.removeExpired()

	if bl.Size() != 1 {
		t.Errorf("size after removeExpired = %d, want 1", bl.Size())
	}
	if bl.Contains("expired-token") {
		t.Error("expired token should be removed")
	}
	if !bl.Contains("valid-token") {
		t.Error("valid token should remain")
	}
}

func TestTokenBlacklist_RemoveExpired_NoFilePersistence(t *testing.T) {
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		CleanupInterval: time.Hour,
	}
	bl, err := NewTokenBlacklistWithConfig(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer bl.Stop()

	bl.mu.Lock()
	bl.items["expired-token"] = &BlacklistItem{
		Token:     "expired-token",
		RevokedAt: time.Now().Add(-2 * time.Hour),
		Reason:    "test",
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	bl.mu.Unlock()

	bl.removeExpired()

	if bl.Size() != 0 {
		t.Errorf("size = %d, want 0", bl.Size())
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
