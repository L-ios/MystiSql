//go:build e2e

package e2e

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EBatch_Insert(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "batch-test-user",
			"role":    "admin",
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

	config, _ := LoadConfig()
	db := NewMySQLConnection(t, &config.MySQL)
	defer db.Close()

	t.Run("cleanup test data", func(t *testing.T) {
		CleanupTable(t, db, "users WHERE username LIKE 'batch_test_%'")
	})

	t.Run("batch insert", func(t *testing.T) {
		resp, err := client.Post("/api/v1/batch", map[string]interface{}{
			"instance": instance,
			"queries": []string{
				"INSERT INTO users (username, email, password_hash) VALUES ('batch_test_1', 'batch1@test.com', 'hash1')",
				"INSERT INTO users (username, email, password_hash) VALUES ('batch_test_2', 'batch2@test.com', 'hash2')",
				"INSERT INTO users (username, email, password_hash) VALUES ('batch_test_3', 'batch3@test.com', 'hash3')",
			},
			"stopOnError": false,
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseJSONResponse(t, resp)
		assert.Contains(t, result["message"], "completed")

		batchResult := result["result"].(map[string]interface{})
		assert.Equal(t, float64(3), batchResult["successCount"])
	})

	t.Run("cleanup test data", func(t *testing.T) {
		CleanupTable(t, db, "users WHERE username LIKE 'batch_test_%'")
	})
}

func TestE2EBatch_MixedOperations(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "batch-mixed-user",
			"role":    "admin",
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

	config, _ := LoadConfig()
	db := NewMySQLConnection(t, &config.MySQL)
	defer db.Close()

	t.Run("cleanup and setup test data", func(t *testing.T) {
		CleanupTable(t, db, "users WHERE username LIKE 'batch_mixed_%'")
	})

	t.Run("batch mixed operations", func(t *testing.T) {
		resp, err := client.Post("/api/v1/batch", map[string]interface{}{
			"instance": instance,
			"queries": []string{
				"INSERT INTO users (username, email, password_hash) VALUES ('batch_mixed_1', 'mixed1@test.com', 'hash1')",
				"INSERT INTO users (username, email, password_hash) VALUES ('batch_mixed_2', 'mixed2@test.com', 'hash2')",
				"UPDATE users SET email = 'updated@test.com' WHERE username = 'batch_mixed_1'",
				"DELETE FROM users WHERE username = 'batch_mixed_2'",
			},
			"stopOnError": false,
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("cleanup test data", func(t *testing.T) {
		CleanupTable(t, db, "users WHERE username LIKE 'batch_mixed_%'")
	})
}

func TestE2EBatch_WithTransaction(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "batch-tx-user",
			"role":    "admin",
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

	config, _ := LoadConfig()
	db := NewMySQLConnection(t, &config.MySQL)
	defer db.Close()

	t.Run("cleanup test data", func(t *testing.T) {
		CleanupTable(t, db, "users WHERE username LIKE 'batch_tx_%'")
	})

	t.Run("batch with transaction", func(t *testing.T) {
		resp, err := client.Post("/api/v1/batch", map[string]interface{}{
			"instance":       instance,
			"queries":        []string{"INSERT INTO users (username, email, password_hash) VALUES ('batch_tx_1', 'tx1@test.com', 'hash1')"},
			"useTransaction": true,
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("cleanup test data", func(t *testing.T) {
		CleanupTable(t, db, "users WHERE username LIKE 'batch_tx_%'")
	})
}

func TestE2EBatch_SizeLimit(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "batch-limit-user",
			"role":    "admin",
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

	t.Run("batch size limit exceeded", func(t *testing.T) {
		queries := make([]string, 1001)
		for i := range queries {
			queries[i] = "SELECT 1"
		}

		resp, err := client.Post("/api/v1/batch", map[string]interface{}{
			"instance": instance,
			"queries":  queries,
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
