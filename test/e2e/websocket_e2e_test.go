//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getWebSocketURL() string {
	if url := os.Getenv("MYSTISQL_WS_URL"); url != "" {
		return url
	}
	return "ws://127.0.0.1:8080"
}

func TestE2EWebSocket_Connection(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "ws-test-user",
			"role":     "admin",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		result := ParseJSONResponse(t, resp)
		data := result["data"].(map[string]interface{})
		token = data["token"].(string)
	})

	t.Run("connect with valid token", func(t *testing.T) {
		wsURL := getWebSocketURL() + "/ws?token=" + token
		dialer := websocket.DefaultDialer

		conn, resp, err := dialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	})

	t.Run("connect without token should fail", func(t *testing.T) {
		wsURL := getWebSocketURL() + "/ws"
		dialer := websocket.DefaultDialer

		_, resp, err := dialer.Dial(wsURL, nil)
		require.Error(t, err)
		if resp != nil {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		}
	})
}

func TestE2EWebSocket_Query(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "ws-query-user",
			"role":     "admin",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		result := ParseJSONResponse(t, resp)
		data := result["data"].(map[string]interface{})
		token = data["token"].(string)
	})

	wsURL := getWebSocketURL() + "/ws?token=" + token
	dialer := websocket.DefaultDialer

	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	instance := os.Getenv("MYSQL_INSTANCE")
	if instance == "" {
		instance = "local-mysql"
	}

	t.Run("execute SELECT query", func(t *testing.T) {
		queryMsg := map[string]interface{}{
			"action":    "query",
			"instance":  instance,
			"query":     "SELECT 1 AS test",
			"requestId": "req-001",
		}

		err := conn.WriteJSON(queryMsg)
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(message, &result)
		require.NoError(t, err)

		assert.Equal(t, "req-001", result["requestId"])
		assert.Equal(t, "query_result", result["type"])
	})

	t.Run("ping pong", func(t *testing.T) {
		pingMsg := map[string]interface{}{
			"action": "ping",
		}

		err := conn.WriteJSON(pingMsg)
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(message, &result)
		require.NoError(t, err)

		assert.Equal(t, "pong", result["type"])
	})
}

func TestE2EWebSocket_Error(t *testing.T) {
	SkipIfShort(t)

	client := NewAPIClient(getAPIBaseURL())

	var token string
	t.Run("setup - get auth token", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/token", map[string]string{
			"user_id": "ws-error-user",
			"role":     "admin",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		result := ParseJSONResponse(t, resp)
		data := result["data"].(map[string]interface{})
		token = data["token"].(string)
	})

	wsURL := getWebSocketURL() + "/ws?token=" + token
	dialer := websocket.DefaultDialer

	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	t.Run("invalid SQL should return error", func(t *testing.T) {
		queryMsg := map[string]interface{}{
			"action":    "query",
			"instance":  "local-mysql",
			"query":     "INVALID SQL STATEMENT",
			"requestId": "req-err-001",
		}

		err := conn.WriteJSON(queryMsg)
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(message, &result)
		require.NoError(t, err)

		assert.Equal(t, "req-err-001", result["requestId"])
		assert.Equal(t, "error", result["type"])
	})
}
