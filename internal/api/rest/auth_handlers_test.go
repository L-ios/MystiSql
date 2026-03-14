package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MystiSql/internal/service/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupAuthTestHandlers(t *testing.T) (*AuthHandlers, *auth.AuthService) {
	gin.SetMode(gin.TestMode)

	authService, err := auth.NewAuthService("test-secret-key-for-testing", 24*time.Hour)
	require.NoError(t, err)

	logger := zap.NewNop()

	handlers := NewAuthHandlers(authService, logger)
	return handlers, authService
}

func TestGenerateTokenEndpoint_Success(t *testing.T) {
	handlers, _ := setupAuthTestHandlers(t)

	router := gin.New()
	router.POST("/api/v1/auth/token", handlers.GenerateToken)

	reqBody := map[string]string{
		"user_id": "test-user",
		"role":    "admin",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
}

func TestGenerateTokenEndpoint_MissingFields(t *testing.T) {
	handlers, _ := setupAuthTestHandlers(t)

	router := gin.New()
	router.POST("/api/v1/auth/token", handlers.GenerateToken)

	tests := []struct {
		name    string
		body    map[string]string
		wantErr string
	}{
		{
			name:    "missing user_id",
			body:    map[string]string{"role": "admin"},
			wantErr: "user_id",
		},
		{
			name:    "missing role",
			body:    map[string]string{"user_id": "test-user"},
			wantErr: "role",
		},
		{
			name:    "empty body",
			body:    map[string]string{},
			wantErr: "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestGetTokenInfo_Success(t *testing.T) {
	handlers, authService := setupAuthTestHandlers(t)

	token, err := authService.GenerateToken(context.Background(), "test-user", "admin")
	require.NoError(t, err)

	router := gin.New()
	router.GET("/api/v1/auth/token/info", handlers.GetTokenInfo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/token/info?token="+token, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "test-user", resp["user_id"])
	assert.Equal(t, "admin", resp["role"])
}

func TestGetTokenInfo_InvalidToken(t *testing.T) {
	handlers, _ := setupAuthTestHandlers(t)

	router := gin.New()
	router.GET("/api/v1/auth/token/info", handlers.GetTokenInfo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/token/info?token=invalid-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTokenInfo_MissingToken(t *testing.T) {
	handlers, _ := setupAuthTestHandlers(t)

	router := gin.New()
	router.GET("/api/v1/auth/token/info", handlers.GetTokenInfo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/token/info", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRevokeToken_Success(t *testing.T) {
	handlers, authService := setupAuthTestHandlers(t)

	token, err := authService.GenerateToken(context.Background(), "test-user", "admin")
	require.NoError(t, err)

	router := gin.New()
	router.DELETE("/api/v1/auth/token", handlers.RevokeToken)

	reqBody := map[string]string{"token": token}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListTokens(t *testing.T) {
	handlers, authService := setupAuthTestHandlers(t)

	token, err := authService.GenerateToken(context.Background(), "test-user", "admin")
	require.NoError(t, err)

	_ = authService.RevokeToken(context.Background(), token)

	router := gin.New()
	router.GET("/api/v1/auth/tokens", handlers.ListTokens)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/tokens", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
