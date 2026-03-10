//go:build e2e

package e2e

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getAPIBaseURL() string {
	if url := os.Getenv("MYSTISQL_API_URL"); url != "" {
		return url
	}
	return "http://127.0.0.1:8080"
}

func TestE2EAuth_TokenGeneration(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	t.Run("generate token with valid credentials", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "test-user",
			"role":    "admin",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseJSONResponse(t, resp)
		assert.True(t, result["success"].(bool))

		data := result["data"].(map[string]interface{})
		assert.NotEmpty(t, data["token"])
		assert.NotEmpty(t, data["expiresAt"])
		assert.Equal(t, "test-user", data["userId"])
		assert.Equal(t, "admin", data["role"])
	})

	t.Run("generate token without role", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "test-user",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestE2EAuth_TokenValidation(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - generate token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "validation-test-user",
			"role":    "user",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		result := ParseJSONResponse(t, resp)
		data := result["data"].(map[string]interface{})
		token = data["token"].(string)
	})

	t.Run("access protected endpoint with valid token", func(t *testing.T) {
		client.SetToken(token)
		resp, err := client.Get("/api/v1/instances")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("access protected endpoint without token", func(t *testing.T) {
		noTokenClient := NewAPIClient(getAPIBaseURL())
		resp, err := noTokenClient.Get("/api/v1/instances")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("access protected endpoint with invalid token", func(t *testing.T) {
		invalidTokenClient := NewAPIClient(getAPIBaseURL())
		invalidTokenClient.SetToken("invalid-token-12345")
		resp, err := invalidTokenClient.Get("/api/v1/instances")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestE2EAuth_TokenRevocation(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - generate token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "revoke-test-user",
			"role":    "admin",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		result := ParseJSONResponse(t, resp)
		data := result["data"].(map[string]interface{})
		token = data["token"].(string)
	})

	t.Run("revoke token", func(t *testing.T) {
		client.SetToken(token)
		resp, err := client.Delete("/api/v1/auth/token", map[string]string{
			"token": token,
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("revoked token should be rejected", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)

		revokedClient := NewAPIClient(getAPIBaseURL())
		revokedClient.SetToken(token)
		resp, err := revokedClient.Get("/api/v1/instances")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestE2EAuth_HealthEndpointNoAuth(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	t.Run("health endpoint should not require auth", func(t *testing.T) {
		resp, err := client.Get("/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseJSONResponse(t, resp)
		assert.Equal(t, "healthy", result["status"])
	})
}
