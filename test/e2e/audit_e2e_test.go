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

func TestE2EAudit_LogQuery(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "audit-test-user",
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

	t.Run("execute query that should be logged", func(t *testing.T) {
		resp, err := client.Post("/api/v1/query", map[string]interface{}{
			"instance": instance,
			"sql":      "SELECT 1 AS test",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("query audit logs", func(t *testing.T) {
		time.Sleep(500 * time.Millisecond)

		endTime := time.Now().Format("2006-01-02")
		startTime := time.Now().Add(-1 * time.Hour).Format("2006-01-02")

		resp, err := client.Get("/api/v1/audit/logs?start_time=" + startTime + "&end_time=" + endTime + "&user_id=audit-test-user")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseJSONResponse(t, resp)
		assert.True(t, result["success"].(bool))
	})
}
