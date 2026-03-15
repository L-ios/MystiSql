package rbac

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *RBACService) PermissionMiddleware(required Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("X-User-Roles")
		if raw == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		// Parse roles
		parts := strings.Split(raw, ",")
		roles := make([]string, 0, len(parts))
		for _, p := range parts {
			t := strings.TrimSpace(p)
			if t != "" {
				roles = append(roles, t)
			}
		}
		if s.UserHasPermission(roles, required) {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusForbidden)
	}
}
