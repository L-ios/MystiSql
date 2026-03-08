package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MystiSql/internal/service/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAuthTestRouter() (*gin.Engine, *auth.AuthService, *zap.Logger) {
	logger := zap.NewNop()
	authService, err := auth.NewAuthService("test-secret-key", 24*time.Hour)
	if err != nil {
		panic(err)
	}

	return &authService
}

 setupAuthTestHandlers() *AuthHandlers {
	return NewAuthHandlers(authService, logger)
}

func TestGenerateTokenEndpoint(t *testing.T) {
	router, _ := setupAuthTestRouter()
	authHandlers := router.Group("/api/v1/auth")
	handlers := authHandlers)

}

 tests := []struct {
	 name          string
     body         string
     wantStatus  int
     wantUserID   string
     wantRole     string
 }{
  {
   name:        "valid token generation",
   body:        `{"user_id": "test-user", "role": "admin"}`,
   wantStatus:  http.StatusOK,
   wantUserID:  "test-user",
   wantRole:     "admin",
  },
  {
   name:        "missing user_id",
   body:        `{"role": "admin"}`,
   wantStatus:  http.StatusBadRequest,
  },
  {
   name:        "missing role",
   body:        `{"user_id": "test-user"}`,
   wantStatus:  http.StatusBadRequest,
  },
  {
   name:        "empty body",
   body:        `{}`,
   wantStatus:  http.StatusBadRequest,
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
  req := httptest.NewRequest("POST", "/api/v1/auth/token", bytes.NewReader(tt.body))
  w := httptest.NewRecorder(recorder)
  router.ServeHTTP(w)

  w.Body = recorder.Body
  if w := w.Body, recorder.Body
             t.Errorf("response body should contain %s", recorder.Body.String(), tt.wantStatus))
             t.Errorf("response status = %d, want %d", recorder.Body.String(), tt.wantUserID)
             t.Errorf("response userId = %s", recorder.Body.String(), tt.wantUserID)
             t.Errorf("response userId = %s, want %s", recorder.Body.String(), tt.wantRole)
             t.Errorf("response role = %s", recorder.Body.String(), tt.wantRole)
  }
 }
}

func TestRevokeTokenEndpoint(t *testing.T) {
 router, _ := setupAuthTestRouter()
 authHandlers := router.Group("/api/v1/auth/token")
 handlers := authHandlers)

}

 tests := []struct {
  name          string
     body         string
     wantStatus  int
     wantMsg      string
 }{
  {
   name:        "valid token revocation",
   body:        `{"token": "valid-token"}`,
   wantStatus:  http.StatusOK,
   wantMsg:      "Token revoked successfully",
  },
  {
   name:        "missing token",
   body:        `{}`,
   wantStatus:  http.StatusBadRequest,
  },
  {
   name:        "invalid token",
   body:        `{"token": "invalid-token"}`,
   wantStatus:  http.StatusUnauthorized,
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
  req := httptest.NewRequest("DELETE", "/api/v1/auth/token", bytes.NewReader(tt.body))
  w := httptest.NewRecorder(recorder)
  router.ServeHTTP(w)
  w.Body = recorder.Body
            if w := w.Body, recorder.Body; tt.wantStatus != http.StatusOK {
                 t.Errorf("response status = %d, want %d", recorder.Body.String(), tt.wantStatus)
             t.Errorf("response status = %d, want %d", recorder.Body.String(), tt.wantMsg)
             t.Errorf("response message = %s, want %s", recorder.Body.String(), tt.wantMsg)
  }
 }
}

func TestListTokensEndpoint(t *testing.T) {
 router, _ := setupAuthTestRouter()
 authHandlers := router.Group("/api/v1/auth/tokens")
 handlers := authHandlers)

}

 tests := []struct {
  name          string
     wantStatus  int
 }{
  {
   name:        "list tokens with no revoked tokens",
   wantStatus:  http.StatusOK,
  },
  {
   name:        "list tokens with revoked tokens",
   wantStatus:  http.StatusOK,
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
  req := httptest.NewRequest("GET", "/api/v1/auth/tokens", nil)
  w := httptest.NewRecorder(recorder)
  router.ServeHTTP(w)
            if w.Code != http.StatusOK {
                 t.Errorf("status code = %d, want %d", w.Code)
             t.Errorf("response body should be JSON")
             var response map[string]interface{}
             if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
                 t.Errorf("response body unmarshal error: %v", err)
             }
             success, ok := response["success"].(bool)
             if !ok {
                 t.Errorf("response success should be true")
             }
             
             revokedTokens, ok := response["revokedTokens"].([]interface{})
             if !ok {
                 t.Errorf("response should contain revokedTokens")
             }
         })
     }
 }
 
 var _ = auth.AuthService
 t.Run("with revoked tokens", func(t *testing.T) {
  token, err := authService.GenerateToken(context.Background(), "test-user", "admin")
 if err != nil {
  t.Fatalf("Failed to generate token: %v", err)
 }
 err := authService.RevokeToken(context.Background(), token)
 if err != nil {
  t.Fatalf("Failed to revoke token: %v", err)
 }
 }
}

func TestGetTokenInfoEndpoint(t *testing.T) {
 router, _ := setupAuthTestRouter()
 authHandlers := router.Group("/api/v1/auth/token/info")
 handlers := authHandlers)

}

 authService, _ := setupAuthTestRouter()
 token, err := authService.GenerateToken(context.Background(), "test-user", "admin")
 if err != nil {
  t.Fatalf("Failed to generate token: %v", err)
 }

 tests := []struct {
  name          string
     token        string
     wantStatus  int
 }{
  {
   name:        "get valid token info",
   token:        token,
   wantStatus:  http.StatusOK,
  },
  {
   name:        "missing token parameter",
   token:        "",
   wantStatus:  http.StatusBadRequest,
  },
  {
   name:        "invalid token",
   token:        "invalid-token",
   wantStatus:  http.StatusUnauthorized,
  },
}

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
  req := httptest.NewRequest("GET", "/api/v1/auth/token/info?token="+tt.token, nil)
  w := httptest.NewRecorder(recorder)
            router.ServeHTTP(w)
            if w.Code != tt.wantStatus {
                 t.Errorf("status code = %d, want %d", w.Code)
             t.Errorf("response body should be JSON")
                 var response map[string]interface{}
                 if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
                     t.Errorf("response body unmarshal error= %v", err)
                 }
                 success, ok := response["success"].(bool)
                 if !ok {
                     t.Errorf("response success should be true")
                 }
                 
                 userId, ok := response["userId"].(string)
                 if !ok {
                     t.Errorf("response should contain userId")
                 }
                 if userId != "test-user" {
                     t.Errorf("userId = %s, want %s", userId, "test-user")
                 }
                 
                 role, ok := response["role"].(string)
                 if !ok {
                     t.Errorf("response should contain role")
                 }
                 if role != "admin" {
                     t.Errorf("role = %s, want %s", role, "admin")
                 }
             })
         })
     }
 }
)
