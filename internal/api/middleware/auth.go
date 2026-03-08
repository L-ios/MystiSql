package middleware

import (
	"net/http"
	"strings"

	"MystiSql/internal/service/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var defaultWhitelistPaths = map[string]bool{
	"/health":            true,
	"/api/v1/auth/token": true,
}

func AuthMiddleware(authService *auth.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			logger.Warn("Missing authentication token",
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "MISSING_TOKEN",
					"message": "Authentication token is required",
				},
			})
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			logger.Warn("Invalid authentication token",
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.Error(err),
			)

			errorCode := "INVALID_TOKEN"
			if err == auth.ErrTokenExpired {
				errorCode = "TOKEN_EXPIRED"
			} else if err == auth.ErrTokenRevoked {
				errorCode = "TOKEN_REVOKED"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    errorCode,
					"message": "Invalid or expired authentication token",
				},
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("token", token)

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	token := c.Query("token")
	if token != "" {
		return token
	}

	return ""
}

func AuthMiddlewareWithWhitelist(authService *auth.AuthService, logger *zap.Logger, whitelistPaths []string) gin.HandlerFunc {
	whitelist := make(map[string]bool)
	for _, path := range whitelistPaths {
		whitelist[path] = true
	}
	for path := range defaultWhitelistPaths {
		whitelist[path] = true
	}

	return func(c *gin.Context) {
		if isWhitelisted(c.Request.URL.Path, whitelist) {
			c.Next()
			return
		}

		AuthMiddleware(authService, logger)(c)
	}
}

func isWhitelisted(path string, whitelist map[string]bool) bool {
	if whitelist[path] {
		return true
	}

	for whitelistedPath := range whitelist {
		if strings.HasPrefix(path, whitelistedPath) {
			return true
		}
	}

	return false
}

func GetDefaultWhitelistPaths() []string {
	paths := make([]string, 0, len(defaultWhitelistPaths))
	for path := range defaultWhitelistPaths {
		paths = append(paths, path)
	}
	return paths
}
