package rbac

import (
	"sync"
)

type RBACService struct {
	mu    sync.RWMutex
	roles map[string]Role // role name -> role
}

func NewRBACService() *RBACService {
	return &RBACService{roles: make(map[string]Role)}
}

func (r *RBACService) AddRole(role Role) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roles[role.Name] = role
}

func (r *RBACService) HasPermission(role Role, perm Permission) bool {
	for _, p := range role.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

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

var userRolesStore = struct {
	sync.RWMutex
	data map[string][]string
}{
	data: make(map[string][]string),
}

func (r *RBACService) ListRoles() []Role {
	r.mu.RLock()
	defer r.mu.RUnlock()
	roles := make([]Role, 0, len(r.roles))
	for _, role := range r.roles {
		roles = append(roles, role)
	}
	return roles
}

func (r *RBACService) GetRole(name string) (Role, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	role, exists := r.roles[name]
	return role, exists
}

func (r *RBACService) DeleteRole(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.roles, name)
}

func (r *RBACService) ListUserRoles(userID string) []string {
	userRolesStore.RLock()
	defer userRolesStore.RUnlock()
	roles, exists := userRolesStore.data[userID]
	if !exists {
		return []string{}
	}
	return append([]string{}, roles...)
}

func (r *RBACService) AssignRolesToUser(userID string, roleNames []string) {
	userRolesStore.Lock()
	defer userRolesStore.Unlock()
	userRolesStore.data[userID] = append([]string{}, roleNames...)
}

func (r *RBACService) RemoveUserRoles(userID string) {
	userRolesStore.Lock()
	defer userRolesStore.Unlock()
	delete(userRolesStore.data, userID)
}
