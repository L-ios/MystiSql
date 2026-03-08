package auth

import (
	"testing"
	"time"
)

func TestNewTokenGenerator(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		duration  time.Duration
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid configuration",
			secretKey: "test-secret-key",
			duration:  24 * time.Hour,
			wantErr:   false,
		},
		{
			name:      "empty secret key",
			secretKey: "",
			duration:  24 * time.Hour,
			wantErr:   true,
			errMsg:    "secret key cannot be empty",
		},
		{
			name:      "zero duration",
			secretKey: "test-secret-key",
			duration:  0,
			wantErr:   true,
			errMsg:    "token duration must be positive",
		},
		{
			name:      "negative duration",
			secretKey: "test-secret-key",
			duration:  -1 * time.Hour,
			wantErr:   true,
			errMsg:    "token duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := NewTokenGenerator(tt.secretKey, tt.duration)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTokenGenerator() expected error, got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewTokenGenerator() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("NewTokenGenerator() unexpected error: %v", err)
				return
			}
			if generator == nil {
				t.Error("NewTokenGenerator() returned nil generator")
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	generator, err := NewTokenGenerator("test-secret-key", 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	tests := []struct {
		name    string
		userID  string
		role    string
		wantErr bool
	}{
		{
			name:    "valid token generation",
			userID:  "user123",
			role:    "admin",
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			role:    "admin",
			wantErr: true,
		},
		{
			name:    "empty role",
			userID:  "user123",
			role:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := generator.GenerateToken(tt.userID, tt.role)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateToken() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GenerateToken() unexpected error: %v", err)
				return
			}
			if token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	secretKey := "test-secret-key"
	duration := 24 * time.Hour

	generator, err := NewTokenGenerator(secretKey, duration)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	validToken, err := generator.GenerateToken("user123", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
		errType error
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			errType: ErrInvalidToken,
		},
		{
			name:    "invalid token format",
			token:   "invalid-token",
			wantErr: true,
			errType: ErrInvalidToken,
		},
		{
			name:    "token signed with different key",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCIsInJvbGUiOiJhZG1pbiIsImV4cCI6OTk5OTk5OTk5OX0.wrong-signature",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := generator.ValidateToken(tt.token)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateToken() expected error, got nil")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("ValidateToken() error = %v, want %v", err, tt.errType)
				}
				return
			}
			if err != nil {
				t.Errorf("ValidateToken() unexpected error: %v", err)
				return
			}
			if claims == nil {
				t.Error("ValidateToken() returned nil claims")
				return
			}
			if claims.UserID != "user123" {
				t.Errorf("ValidateToken() userID = %v, want user123", claims.UserID)
			}
			if claims.Role != "admin" {
				t.Errorf("ValidateToken() role = %v, want admin", claims.Role)
			}
		})
	}
}

func TestTokenClaims(t *testing.T) {
	generator, err := NewTokenGenerator("test-secret-key", 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	token, err := generator.GenerateToken("user123", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := generator.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("UserID = %v, want user123", claims.UserID)
	}

	if claims.Role != "admin" {
		t.Errorf("Role = %v, want admin", claims.Role)
	}

	if claims.ID == "" {
		t.Error("Token ID is empty")
	}

	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt is nil")
	}

	if claims.IssuedAt == nil {
		t.Error("IssuedAt is nil")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("Token is already expired")
	}
}
