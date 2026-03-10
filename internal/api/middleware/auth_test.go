package middleware

import (
	"context"
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

func TestAuthMiddleware_MissingToken(t *testing.T) {
	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("test-secret", 3600*time.Second)

	router := gin.New()
	router.Use(AuthMiddleware(authService, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("test-secret", 3600*time.Second)

	router := gin.New()
	router.Use(AuthMiddleware(authService, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("test-secret", 3600*time.Second)

	token, err := authService.GenerateToken(context.Background(), "test-user", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	router := gin.New()
	router.Use(AuthMiddleware(authService, logger))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			t.Error("user_id not set in context")
		}
		if userID != "test-user" {
			t.Errorf("Expected user_id 'test-user', got %v", userID)
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddlewareWithWhitelist_WhitelistedPath(t *testing.T) {
	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("test-secret", 3600*time.Second)

	router := gin.New()
	router.Use(AuthMiddlewareWithWhitelist(authService, logger, []string{"/health"}))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddlewareWithWhitelist_DefaultWhitelist(t *testing.T) {
	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("test-secret", 3600*time.Second)

	router := gin.New()
	router.Use(AuthMiddlewareWithWhitelist(authService, logger, []string{}))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health endpoint should be whitelisted by default, got status %d", w.Code)
	}
}

func TestAuthMiddlewareWithWhitelist_NonWhitelistedPath(t *testing.T) {
	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("test-secret", 3600*time.Second)

	router := gin.New()
	router.Use(AuthMiddlewareWithWhitelist(authService, logger, []string{"/health"}))
	router.GET("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"users": []string{}})
	})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		queryParam    string
		expectedToken string
	}{
		{
			name:          "Bearer token",
			authHeader:    "Bearer my-token",
			expectedToken: "my-token",
		},
		{
			name:          "Bearer token case insensitive",
			authHeader:    "bearer my-token",
			expectedToken: "my-token",
		},
		{
			name:          "Query parameter",
			queryParam:    "my-token",
			expectedToken: "my-token",
		},
		{
			name:          "Authorization header takes precedence",
			authHeader:    "Bearer header-token",
			queryParam:    "query-token",
			expectedToken: "header-token",
		},
		{
			name:          "No token",
			expectedToken: "",
		},
		{
			name:          "Invalid auth header format",
			authHeader:    "InvalidFormat",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			req := httptest.NewRequest("GET", "/test", nil)

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			if tt.queryParam != "" {
				q := req.URL.Query()
				q.Add("token", tt.queryParam)
				req.URL.RawQuery = q.Encode()
			}

			c.Request = req

			token := extractToken(c)

			if token != tt.expectedToken {
				t.Errorf("Expected token '%s', got '%s'", tt.expectedToken, token)
			}
		})
	}
}

func TestIsWhitelisted(t *testing.T) {
	whitelist := map[string]bool{
		"/health":      true,
		"/api/v1/auth": true,
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/health/check", true},
		{"/api/v1/auth/login", true},
		{"/api/v1/users", false},
		{"/api/v1/query", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isWhitelisted(tt.path, whitelist)
			if result != tt.expected {
				t.Errorf("isWhitelisted(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGetDefaultWhitelistPaths(t *testing.T) {
	paths := GetDefaultWhitelistPaths()

	if len(paths) == 0 {
		t.Error("Default whitelist should not be empty")
	}

	hasHealth := false
	for _, path := range paths {
		if path == "/health" {
			hasHealth = true
			break
		}
	}

	if !hasHealth {
		t.Error("Default whitelist should include /health")
	}
}

func TestAuthMiddleware_RevokedToken(t *testing.T) {
	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("test-secret", 3600*time.Second)

	token, err := authService.GenerateToken(context.Background(), "test-user", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	err = authService.RevokeToken(context.Background(), token)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	router := gin.New()
	router.Use(AuthMiddleware(authService, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Revoked token should return %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
