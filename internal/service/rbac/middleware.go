package rbac

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *RBACService) PermissionMiddleware(required Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesVal, exists := c.Get("roles")
		if !exists {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		roles, ok := rolesVal.([]string)
		if !ok || len(roles) == 0 {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if s.UserHasPermission(roles, required) {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusForbidden)
	}
}
