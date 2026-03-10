//go:build e2e

package e2e

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EValidator_BlockDangerousOperations(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "validator-test-user",
			"role":     "admin",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		result := ParseJSONResponse(t, resp)
		data := result["data"].(map[string]interface{})
		token = data["token"].(string)
	})

	client.SetToken(token)

	instance := os.Getenv("MYSQL_INSTANCE")
	if instance == "" {
		instance = "local-mysql"
	}

	t.Run("block DROP TABLE", func(t *testing.T) {
		resp, err := client.Post("/api/v1/exec", map[string]interface{}{
			"instance": instance,
			"sql":      "DROP TABLE IF EXISTS nonexistent_table",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("block DROP DATABASE", func(t *testing.T) {
		resp, err := client.Post("/api/v1/exec", map[string]interface{}{
			"instance": instance,
			"sql":      "DROP DATABASE IF EXISTS nonexistent_db",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("block TRUNCATE", func(t *testing.T) {
		resp, err := client.Post("/api/v1/exec", map[string]interface{}{
			"instance": instance,
			"sql":      "TRUNCATE TABLE users",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("block DELETE without WHERE", func(t *testing.T) {
		resp, err := client.Post("/api/v1/exec", map[string]interface{}{
			"instance": instance,
			"sql":      "DELETE FROM users",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("block UPDATE without WHERE", func(t *testing.T) {
		resp, err := client.Post("/api/v1/exec", map[string]interface{}{
			"instance": instance,
			"sql":      "UPDATE users SET name = 'test'",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("allow DELETE with WHERE", func(t *testing.T) {
		resp, err := client.Post("/api/v1/exec", map[string]interface{}{
			"instance": instance,
			"sql":      "DELETE FROM users WHERE id = 999999",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("allow UPDATE with WHERE", func(t *testing.T) {
		resp, err := client.Post("/api/v1/exec", map[string]interface{}{
			"instance": instance,
			"sql":      "UPDATE users SET name = 'test' WHERE id = 999999",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestE2EValidator_WhitelistBlacklist(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "whitelist-test-user",
			"role":     "admin",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		result := ParseJSONResponse(t, resp)
		data := result["data"].(map[string]interface{})
		token = data["token"].(string)
	})

	client.SetToken(token)

	t.Run("update whitelist", func(t *testing.T) {
		resp, err := client.Post("/api/v1/validator/whitelist", map[string]interface{}{
			"patterns": []string{
				"SELECT * FROM system_config",
				"^SHOW TABLES$",
			},
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("update blacklist", func(t *testing.T) {
		resp, err := client.Post("/api/v1/validator/blacklist", map[string]interface{}{
			"patterns": []string{
				"DELETE FROM audit_log",
			},
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
