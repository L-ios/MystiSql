package rest

import (
	"MystiSql/internal/service/rbac"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RBACHandlers struct {
	service *rbac.RBACService
	logger  *zap.Logger
}

func NewRBACHandlers(service *rbac.RBACService, logger *zap.Logger) *RBACHandlers {
	if service == nil {
		return nil
	}
	return &RBACHandlers{
		service: service,
		logger:  logger,
	}
}

func (h *RBACHandlers) ListRoles(c *gin.Context) {
	roles := h.service.ListRoles()
	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

func (h *RBACHandlers) GetRole(c *gin.Context) {
	name := c.Param("name")
	role, exists := h.service.GetRole(name)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "role not found",
		})
		return
	}
	c.JSON(http.StatusOK, role)
}
func (h *RBACHandlers) CreateRole(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	perms := make([]rbac.Permission, 0, len(req.Permissions))
	for _, p := range req.Permissions {
		perm, err := rbac.ParsePermissionFromString(p)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid permission format: " + p})
			return
		}
		perms = append(perms, perm)
	}

	role := rbac.Role{
		Name:        req.Name,
		Permissions: perms,
	}

	h.service.AddRole(role)
	c.JSON(http.StatusCreated, gin.H{
		"message": "role created successfully",
		"role":    role,
	})
}
func (h *RBACHandlers) DeleteRole(c *gin.Context) {
	name := c.Param("name")
	h.service.DeleteRole(name)
	c.JSON(http.StatusOK, gin.H{
		"message": "role deleted successfully",
	})
}
func (h *RBACHandlers) ListUserRoles(c *gin.Context) {
	userID := c.Param("id")
	roles := h.service.ListUserRoles(userID)
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"roles":   roles,
	})
}
func (h *RBACHandlers) AssignRoleToUser(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Roles []string `json:"roles" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.service.AssignRolesToUser(userID, req.Roles)
	c.JSON(http.StatusOK, gin.H{
		"message":        "roles assigned successfully",
		"assigned_roles": req.Roles,
	})
}
