package auth

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewTokenBlacklist(t *testing.T) {
	bl := NewTokenBlacklist()
	if bl == nil {
		t.Fatal("NewTokenBlacklist returned nil")
	}
	if bl.Size() != 0 {
		t.Errorf("new blacklist should be empty, got %d", bl.Size())
	}
}

func TestTokenBlacklist_AddAndContains(t *testing.T) {
	bl := NewTokenBlacklist()
	bl.Add("token1", "test reason")

	if !bl.Contains("token1") {
		t.Error("should contain token1 after Add")
	}
	if bl.Contains("token2") {
		t.Error("should not contain unadded token2")
	}
}

func TestTokenBlacklist_Remove(t *testing.T) {
	bl := NewTokenBlacklist()
	bl.Add("token1", "test reason")
	bl.Remove("token1")

	if bl.Contains("token1") {
		t.Error("should not contain removed token")
	}
}

func TestTokenBlacklist_Remove_Nonexistent(t *testing.T) {
	bl := NewTokenBlacklist()
	bl.Remove("nonexistent")
}

func TestTokenBlacklist_GetAll(t *testing.T) {
	bl := NewTokenBlacklist()
	bl.Add("token1", "reason1")
	bl.Add("token2", "reason2")

	items := bl.GetAll()
	if len(items) != 2 {
		t.Errorf("GetAll count = %d, want %d", len(items), 2)
	}
	if items[0].Token != "token1" && items[1].Token != "token1" {
		t.Error("GetAll should contain added tokens")
	}
}

func TestTokenBlacklist_GetAll_Empty(t *testing.T) {
	bl := NewTokenBlacklist()
	items := bl.GetAll()
	if len(items) != 0 {
		t.Errorf("empty blacklist GetAll count = %d, want %d", len(items), 0)
	}
}

func TestTokenBlacklist_Clear(t *testing.T) {
	bl := NewTokenBlacklist()
	bl.Add("token1", "reason1")
	bl.Add("token2", "reason2")

	bl.Clear()

	if bl.Size() != 0 {
		t.Errorf("Clear should empty blacklist, got %d", bl.Size())
	}
	if bl.Contains("token1") {
		t.Error("should not contain token after Clear")
	}
}

func TestTokenBlacklist_Size(t *testing.T) {
	bl := NewTokenBlacklist()
	if bl.Size() != 0 {
		t.Errorf("initial size = %d, want 0", bl.Size())
	}

	bl.Add("t1", "r1")
	bl.Add("t2", "r2")
	bl.Add("t3", "r3")

	if bl.Size() != 3 {
		t.Errorf("size after 3 adds = %d, want %d", bl.Size(), 3)
	}
}

func TestTokenBlacklist_RevokeTime(t *testing.T) {
	bl := NewTokenBlacklist()
	before := time.Now()

	bl.Add("token1", "test reason")

	item := bl.GetAll()[0]
	if item.RevokedAt.Before(before) {
		t.Error("RevokedAt should be after before time")
	}
	if item.Reason != "test reason" {
		t.Errorf("Reason = %q, want %q", item.Reason, "test reason")
	}
}

func TestAuthService_New(t *testing.T) {
	svc, err := NewAuthService("secret-key-1234567890123456", 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService error: %v", err)
	}
	if svc == nil {
		t.Fatal("NewAuthService returned nil")
	}
}

func TestAuthService_New_BadSecret(t *testing.T) {
	_, err := NewAuthService("", 24*time.Hour)
	if err == nil {
		t.Error("expected error for empty secret")
	}
}

func TestAuthService_GenerateAndValidate(t *testing.T) {
	svc, err := NewAuthService("secret-key-1234567890123456", 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService error: %v", err)
	}

	token, err := svc.GenerateToken(context.Background(), "user1", "admin")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := svc.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}
	if claims.UserID != "user1" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user1")
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want %q", claims.Role, "admin")
	}
}

func TestAuthService_RevokeToken(t *testing.T) {
	svc, err := NewAuthService("secret-key-1234567890123456", 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService error: %v", err)
	}

	token, err := svc.GenerateToken(context.Background(), "user1", "admin")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	err = svc.RevokeToken(context.Background(), token)
	if err != nil {
		t.Fatalf("RevokeToken error: %v", err)
	}

	if !svc.IsTokenRevoked(context.Background(), token) {
		t.Error("token should be revoked")
	}

	_, err = svc.ValidateToken(context.Background(), token)
	if err != ErrTokenRevoked {
		t.Errorf("ValidateToken after revoke should return ErrTokenRevoked, got %v", err)
	}
}

func TestAuthService_IsTokenRevoked(t *testing.T) {
	svc, err := NewAuthService("secret-key-1234567890123456", 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService error: %v", err)
	}

	if svc.IsTokenRevoked(context.Background(), "nonexistent-token") {
		t.Error("nonexistent token should not be revoked")
	}
}

func TestAuthService_GetTokenInfo(t *testing.T) {
	svc, err := NewAuthService("secret-key-1234567890123456", 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService error: %v", err)
	}

	token, err := svc.GenerateToken(context.Background(), "user1", "admin")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	info, err := svc.GetTokenInfo(context.Background(), token)
	if err != nil {
		t.Fatalf("GetTokenInfo error: %v", err)
	}
	if info.UserID != "user1" {
		t.Errorf("TokenInfo.UserID = %q, want %q", info.UserID, "user1")
	}
	if info.Role != "admin" {
		t.Errorf("TokenInfo.Role = %q, want %q", info.Role, "admin")
	}
	if info.IssuedAt.IsZero() {
		t.Error("IssuedAt should not be zero")
	}
	if info.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}
}

func TestAuthService_ListRevokedTokens(t *testing.T) {
	svc, err := NewAuthService("secret-key-1234567890123456", 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService error: %v", err)
	}

	token, err := svc.GenerateToken(context.Background(), "user1", "admin")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	revoked := svc.ListRevokedTokens(context.Background())
	if len(revoked) != 0 {
		t.Error("no tokens should be revoked initially")
	}

	svc.RevokeToken(context.Background(), token)

	revoked = svc.ListRevokedTokens(context.Background())
	if len(revoked) != 1 {
		t.Errorf("revoked count = %d, want %d", len(revoked), 1)
	}
}

func TestAuthService_ValidateInvalidToken(t *testing.T) {
	svc, err := NewAuthService("secret-key-1234567890123456", 24*time.Hour)
	if err != nil {
		t.Fatalf("NewAuthService error: %v", err)
	}

	_, err = svc.ValidateToken(context.Background(), "invalid-token")
	if err == nil {
		t.Error("should fail for invalid token")
	}
}

func TestTokenInfo_Struct(t *testing.T) {
	info := TokenInfo{
		UserID:    "user1",
		Role:      "admin",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IssuedAt:  time.Now(),
		TokenID:   "abc123",
	}

	if info.UserID != "user1" {
		t.Errorf("UserID = %q", info.UserID)
	}
	if info.Role != "admin" {
		t.Errorf("Role = %q", info.Role)
	}
	if info.TokenID != "abc123" {
		t.Errorf("TokenID = %q", info.TokenID)
	}
}

func TestNewAuthServiceWithConfig(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "blacklist.jsonl")
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        tmpFile,
		CleanupInterval: time.Minute,
	}
	svc, err := NewAuthServiceWithConfig("secret-key-1234567890123456", 24*time.Hour, cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("NewAuthServiceWithConfig error: %v", err)
	}
	defer svc.Close()

	token, err := svc.GenerateToken(context.Background(), "user1", "admin")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	claims, err := svc.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}
	if claims.UserID != "user1" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user1")
	}
}

func TestNewAuthServiceWithConfig_BadSecret(t *testing.T) {
	cfg := BlacklistConfig{TTL: time.Hour}
	_, err := NewAuthServiceWithConfig("", 24*time.Hour, cfg, zap.NewNop())
	if err == nil {
		t.Error("expected error for empty secret")
	}
}

func TestNewAuthServiceWithConfig_BadFilePath(t *testing.T) {
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		FilePath:        "/nonexistent/dir/blacklist.jsonl",
		CleanupInterval: time.Minute,
	}
	_, err := NewAuthServiceWithConfig("secret-key-1234567890123456", 24*time.Hour, cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("should succeed even with bad file path: %v", err)
	}
}

func TestAuthService_Close(t *testing.T) {
	cfg := BlacklistConfig{
		TTL:             time.Hour,
		CleanupInterval: time.Minute,
	}
	svc, err := NewAuthServiceWithConfig("secret-key-1234567890123456", 24*time.Hour, cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	token, err := svc.GenerateToken(context.Background(), "user1", "admin")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	svc.RevokeToken(context.Background(), token)
	svc.Close()

	if !svc.IsTokenRevoked(context.Background(), token) {
		t.Error("token should still be revoked after Close")
	}
}

func TestNewTokenGenerator_EmptySecret(t *testing.T) {
	_, err := NewTokenGenerator("", time.Hour)
	if err == nil {
		t.Error("expected error for empty secret")
	}
}

func TestNewTokenGenerator_ZeroDuration(t *testing.T) {
	_, err := NewTokenGenerator("secret", 0)
	if err == nil {
		t.Error("expected error for zero duration")
	}
}

func TestNewTokenGenerator_NegativeDuration(t *testing.T) {
	_, err := NewTokenGenerator("secret", -time.Hour)
	if err == nil {
		t.Error("expected error for negative duration")
	}
}

func TestGenerateToken_EmptyUserID(t *testing.T) {
	gen, err := NewTokenGenerator("secret-key-1234567890123456", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gen.GenerateToken("", "admin")
	if err == nil {
		t.Error("expected error for empty user ID")
	}
}

func TestGenerateToken_EmptyRole(t *testing.T) {
	gen, err := NewTokenGenerator("secret-key-1234567890123456", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gen.GenerateToken("user1", "")
	if err == nil {
		t.Error("expected error for empty role")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	gen, err := NewTokenGenerator("secret-key-1234567890123456", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gen.ValidateToken("")
	if err != ErrInvalidToken {
		t.Errorf("error = %v, want ErrInvalidToken", err)
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	gen, err := NewTokenGenerator("secret-key-1234567890123456", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gen.ValidateToken("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.sig")
	if err == nil {
		t.Error("expected error for wrong signing method")
	}
}
