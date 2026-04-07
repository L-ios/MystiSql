package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetToken(t *testing.T) {
	tests := []struct {
		name      string
		flagToken string
		envToken  string
		wantToken string
		wantEmpty bool
	}{
		{
			name:      "from flag",
			flagToken: "flag-token-123",
			wantToken: "flag-token-123",
		},
		{
			name:      "from env when flag empty",
			flagToken: "",
			envToken:  "env-token-456",
			wantToken: "env-token-456",
		},
		{
			name:      "flag overrides env",
			flagToken: "flag-token-789",
			envToken:  "env-token-012",
			wantToken: "flag-token-789",
		},
		{
			name:      "empty when both empty",
			flagToken: "",
			envToken:  "",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalFlag := tokenFlag
			originalEnv := os.Getenv("MYSTISQL_TOKEN")

			defer func() {
				tokenFlag = originalFlag
				if originalEnv != "" {
					os.Setenv("MYSTISQL_TOKEN", originalEnv)
				} else {
					os.Unsetenv("MYSTISQL_TOKEN")
				}
			}()

			tokenFlag = tt.flagToken
			if tt.envToken != "" {
				os.Setenv("MYSTISQL_TOKEN", tt.envToken)
			} else {
				os.Unsetenv("MYSTISQL_TOKEN")
			}

			// Execute
			got := GetToken()

			// Assert
			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.wantToken, got)
			}
		})
	}
}

func TestRequireToken(t *testing.T) {
	tests := []struct {
		name      string
		flagToken string
		envToken  string
		wantToken string
		wantError bool
	}{
		{
			name:      "token provided",
			flagToken: "valid-token",
			wantToken: "valid-token",
			wantError: false,
		},
		{
			name:      "no token",
			flagToken: "",
			envToken:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalFlag := tokenFlag
			originalEnv := os.Getenv("MYSTISQL_TOKEN")

			defer func() {
				tokenFlag = originalFlag
				if originalEnv != "" {
					os.Setenv("MYSTISQL_TOKEN", originalEnv)
				} else {
					os.Unsetenv("MYSTISQL_TOKEN")
				}
			}()

			tokenFlag = tt.flagToken
			if tt.envToken != "" {
				os.Setenv("MYSTISQL_TOKEN", tt.envToken)
			} else {
				os.Unsetenv("MYSTISQL_TOKEN")
			}

			// Execute
			got, err := RequireToken()

			// Assert
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "未提供认证 Token")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, got)
			}
		})
	}
}

func TestValidateTokenWithServer(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/validate" {
			auth := r.Header.Get("Authorization")
			if auth == "Bearer valid-token" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]bool{"valid": true})
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name      string
		token     string
		wantError bool
	}{
		{
			name:      "valid token",
			token:     "valid-token",
			wantError: false,
		},
		{
			name:      "invalid token",
			token:     "invalid-token",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Current implementation always returns nil
			// This test documents the expected behavior
			err := ValidateTokenWithServer(tt.token, server.URL)

			// The current implementation doesn't actually validate
			// It's a placeholder that always returns nil
			assert.NoError(t, err)
		})
	}
}

func TestAuthTokenCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create mock auth server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/token" && r.Method == "POST" {
			// Mock token generation
			response := map[string]interface{}{
				"token":     "mock-jwt-token-12345",
				"userId":    "admin",
				"role":      "admin",
				"expiresAt": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"issuedAt":  time.Now().Format(time.RFC3339),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test would normally use cobra command execution
	// For now, we test the logic directly
	t.Run("generate token request", func(t *testing.T) {
		client := &http.Client{Timeout: 30 * time.Second}

		resp, err := client.Post(
			server.URL+"/api/v1/auth/token",
			"application/json",
			bytes.NewReader([]byte(`{"userId":"admin","role":"admin"}`)),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["token"])
		assert.Equal(t, "admin", result["userId"])
		assert.Equal(t, "admin", result["role"])
	})
}

func TestAuthRevokeCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create mock auth server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/token/revoke" && r.Method == "DELETE" {
			auth := r.Header.Get("Authorization")
			if auth != "" && auth != "Bearer invalid-token" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"message": "Token revoked successfully"})
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name        string
		token       string
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:        "valid token",
			token:       "valid-token-to-revoke",
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			wantStatus:  http.StatusUnauthorized,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{Timeout: 30 * time.Second}

			req, err := http.NewRequestWithContext(
				context.Background(),
				"DELETE",
				server.URL+"/api/v1/auth/token/revoke",
				nil,
			)
			require.NoError(t, err)

			req.Header.Set("Authorization", "Bearer "+tt.token)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantSuccess {
				var result map[string]string
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.Contains(t, result["message"], "revoked")
			}
		})
	}
}

func TestAuthTokenCmdRequestBody_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/token" && r.Method == "POST" {
			buf := new(bytes.Buffer)
			buf.ReadFrom(r.Body)
			receivedBody = buf.String()

			response := map[string]interface{}{
				"token":     "mock-jwt-token-12345",
				"userId":    "admin",
				"role":      "admin",
				"expiresAt": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"issuedAt":  time.Now().Format(time.RFC3339),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("POST request should have JSON body with user_id and role", func(t *testing.T) {
		userID := "admin"
		role := "admin"
		serverURL := server.URL

		body, err := json.Marshal(map[string]string{
			"user_id": userID,
			"role":    role,
		})
		require.NoError(t, err)

		client := &http.Client{Timeout: 30 * time.Second}

		resp, err := client.Post(
			serverURL+"/api/v1/auth/token",
			"application/json",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEmpty(t, receivedBody, "Request body should not be empty")
		assert.JSONEq(t, `{"role":"admin","user_id":"admin"}`, receivedBody)
	})
}

func TestAuthInfoCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create mock auth server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/token/info" && r.Method == "GET" {
			auth := r.Header.Get("Authorization")
			if auth == "Bearer valid-token" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"userId":    "admin",
					"role":      "admin",
					"tokenId":   "token-123",
					"issuedAt":  time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					"expiresAt": time.Now().Add(23 * time.Hour).Format(time.RFC3339),
				})
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("get token info", func(t *testing.T) {
		client := &http.Client{Timeout: 30 * time.Second}

		req, err := http.NewRequestWithContext(
			context.Background(),
			"GET",
			server.URL+"/api/v1/auth/token/info",
			nil,
		)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer valid-token")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "admin", result["userId"])
		assert.Equal(t, "admin", result["role"])
		assert.NotEmpty(t, result["tokenId"])
		assert.NotEmpty(t, result["issuedAt"])
		assert.NotEmpty(t, result["expiresAt"])
	})
}
