package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_TokenSecurity(t *testing.T) {
	service, err := NewAuthService("test-secret-key-for-security-tests", 1*time.Hour)
	require.NoError(t, err)

	t.Run("token contains expected claims", func(t *testing.T) {
		token, err := service.GenerateToken(context.Background(), "test-user", "admin")
		require.NoError(t, err)

		claims, err := service.ValidateToken(context.Background(), token)
		require.NoError(t, err)

		assert.Equal(t, "test-user", claims.UserID)
		assert.Equal(t, "admin", claims.Role)
		assert.True(t, claims.ExpiresAt.After(time.Now()))
	})

	t.Run("tampered token is rejected", func(t *testing.T) {
		token, _ := service.GenerateToken(context.Background(), "test-user", "admin")

		tamperedToken := token + "tampered"
		_, err := service.ValidateToken(context.Background(), tamperedToken)
		assert.Error(t, err, "Tampered token should be rejected")
	})

	t.Run("token from different secret is rejected", func(t *testing.T) {
		otherService, _ := NewAuthService("different-secret-key", 1*time.Hour)
		token, _ := otherService.GenerateToken(context.Background(), "test-user", "admin")

		_, err := service.ValidateToken(context.Background(), token)
		assert.Error(t, err, "Token from different secret should be rejected")
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		shortLivedService, _ := NewAuthService("short-lived-secret", 1*time.Millisecond)
		token, _ := shortLivedService.GenerateToken(context.Background(), "test-user", "admin")

		time.Sleep(10 * time.Millisecond)

		_, err := shortLivedService.ValidateToken(context.Background(), token)
		assert.Error(t, err, "Expired token should be rejected")
	})

	t.Run("revoked token is rejected", func(t *testing.T) {
		token, _ := service.GenerateToken(context.Background(), "test-user", "admin")

		claims, _ := service.ValidateToken(context.Background(), token)
		require.NotNil(t, claims)

		err := service.RevokeToken(context.Background(), token)
		require.NoError(t, err)

		_, err = service.ValidateToken(context.Background(), token)
		assert.Error(t, err, "Revoked token should be rejected")
	})
}

func TestAuthService_TokenLeakagePrevention(t *testing.T) {
	service, err := NewAuthService("test-secret-key", 1*time.Hour)
	require.NoError(t, err)

	t.Run("different users get different tokens", func(t *testing.T) {
		token1, _ := service.GenerateToken(context.Background(), "user1", "admin")
		token2, _ := service.GenerateToken(context.Background(), "user2", "admin")

		assert.NotEqual(t, token1, token2, "Different users should get different tokens")
	})

	t.Run("same user multiple tokens are independent", func(t *testing.T) {
		token1, _ := service.GenerateToken(context.Background(), "user1", "admin")
		token2, _ := service.GenerateToken(context.Background(), "user1", "admin")

		err := service.RevokeToken(context.Background(), token1)
		require.NoError(t, err)

		_, err = service.ValidateToken(context.Background(), token1)
		assert.Error(t, err, "First token should be revoked")

		_, err = service.ValidateToken(context.Background(), token2)
		assert.NoError(t, err, "Second token should still be valid")
	})
}

func TestAuthService_InputValidation(t *testing.T) {
	service, err := NewAuthService("test-secret-key", 1*time.Hour)
	require.NoError(t, err)

	t.Run("empty user ID is rejected", func(t *testing.T) {
		_, err := service.GenerateToken(context.Background(), "", "admin")
		assert.Error(t, err, "Empty user ID should be rejected")
	})

	t.Run("empty role is rejected", func(t *testing.T) {
		_, err := service.GenerateToken(context.Background(), "test-user", "")
		assert.Error(t, err, "Empty role should be rejected")
	})

	t.Run("empty token is rejected", func(t *testing.T) {
		_, err := service.ValidateToken(context.Background(), "")
		assert.Error(t, err, "Empty token should be rejected")
	})

	t.Run("malformed token is rejected", func(t *testing.T) {
		malformedTokens := []string{
			"not-a-token",
			"a.b",
			"a.b.c.d",
			"....",
			"header.payload.signature.extra",
		}

		for _, token := range malformedTokens {
			_, err := service.ValidateToken(context.Background(), token)
			assert.Error(t, err, "Malformed token should be rejected")
		}
	})
}

func TestAuthService_SQLInjectionPrevention(t *testing.T) {
	service, err := NewAuthService("test-secret-key", 1*time.Hour)
	require.NoError(t, err)

	sqlInjectionPayloads := []string{
		"admin'; DROP TABLE users; --",
		"admin\" OR \"1\"=\"1",
		"admin' OR '1'='1",
		"<script>alert('xss')</script>",
		"admin\x00",
	}

	for _, payload := range sqlInjectionPayloads {
		t.Run("payload: "+payload, func(t *testing.T) {
			token, err := service.GenerateToken(context.Background(), payload, "admin")
			require.NoError(t, err)

			claims, err := service.ValidateToken(context.Background(), token)
			require.NoError(t, err)

			assert.Equal(t, payload, claims.UserID, "User ID should be stored safely without execution")
		})
	}
}
