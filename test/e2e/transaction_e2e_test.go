//go:build e2e

package e2e

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ETransaction_BasicFlow(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "tx-test-user",
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

	var transactionID string

	t.Run("begin transaction", func(t *testing.T) {
		resp, err := client.Post("/api/v1/transaction/begin", map[string]interface{}{
			"instance":        instance,
			"isolation_level": "READ_COMMITTED",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseJSONResponse(t, resp)
		transactionID = result["transaction_id"].(string)
		assert.NotEmpty(t, transactionID)
	})

	t.Run("execute query in transaction", func(t *testing.T) {
		resp, err := client.Post("/api/v1/query", map[string]interface{}{
			"instance":       instance,
			"sql":            "SELECT 1 AS test",
			"transaction_id": transactionID,
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get transaction status", func(t *testing.T) {
		resp, err := client.Get("/api/v1/transaction/" + transactionID)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseJSONResponse(t, resp)
		assert.Equal(t, transactionID, result["transaction_id"])
		assert.Equal(t, "active", result["state"])
	})

	t.Run("commit transaction", func(t *testing.T) {
		resp, err := client.Post("/api/v1/transaction/commit", map[string]interface{}{
			"transaction_id": transactionID,
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("committed transaction should be inactive", func(t *testing.T) {
		resp, err := client.Get("/api/v1/transaction/" + transactionID)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestE2ETransaction_Rollback(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "tx-rollback-user",
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

	var transactionID string

	t.Run("begin transaction", func(t *testing.T) {
		resp, err := client.Post("/api/v1/transaction/begin", map[string]interface{}{
			"instance": instance,
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		result := ParseJSONResponse(t, resp)
		transactionID = result["transaction_id"].(string)
	})

	t.Run("rollback transaction", func(t *testing.T) {
		resp, err := client.Post("/api/v1/transaction/rollback", map[string]interface{}{
			"transaction_id": transactionID,
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("rolled back transaction should be inactive", func(t *testing.T) {
		resp, err := client.Get("/api/v1/transaction/" + transactionID)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestE2ETransaction_List(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "tx-list-user",
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

	t.Run("list transactions", func(t *testing.T) {
		resp, err := client.Get("/api/v1/transaction")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseJSONResponse(t, resp)
		assert.NotNil(t, result["transactions"])
	})
}
