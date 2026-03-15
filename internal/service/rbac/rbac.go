package rbac

// Minimal RBAC core: permissions are expressed as three-part strings
// and stored as typed Permissions. Roles are in-memory for now.

import (
	"sync"
)

// RBACService is a tiny in-memory RBAC implementation.
type RBACService struct {
	mu    sync.RWMutex
	roles map[string]Role // role name -> role
}

// NewRBACService creates a new RBAC service instance.
func NewRBACService() *RBACService {
	return &RBACService{roles: make(map[string]Role)}
}

// AddRole registers a new role with its permissions.
func (r *RBACService) AddRole(role Role) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roles[role.Name] = role
}

// HasPermission checks if a given role contains the specified permission.
func (r *RBACService) HasPermission(role Role, perm Permission) bool {
	for _, p := range role.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// UserHasPermission checks across a set of user roles.
func (r *RBACService) UserHasPermission(userRoles []string, perm Permission) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, name := range userRoles {
		if role, ok := r.roles[name]; ok {
			if r.HasPermission(role, perm) {
				return true
			}
		}
	}
	return false
}
