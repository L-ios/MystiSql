package rbac

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParsePermissionFromString(t *testing.T) {
	tests := []struct {
		input   string
		want    Permission
		wantErr bool
	}{
		{"inst:db:SELECT", Permission{Instance: "inst", Database: "db", Action: "SELECT"}, false},
		{"a:b:c", Permission{Instance: "a", Database: "b", Action: "c"}, false},
		{"too-few", Permission{}, true},
		{"a:b:c:d", Permission{}, true},
		{"", Permission{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePermissionFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePermissionFromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParsePermissionFromString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPermission_String(t *testing.T) {
	p := Permission{Instance: "inst", Database: "db", Action: "SELECT"}
	if got := p.String(); got != "inst:db:SELECT" {
		t.Errorf("Permission.String() = %q, want %q", got, "inst:db:SELECT")
	}
}

func TestNewRBACService(t *testing.T) {
	svc := NewRBACService()
	if svc == nil {
		t.Fatal("NewRBACService returned nil")
	}
	if svc.roles == nil {
		t.Error("roles map should be initialized")
	}
}

func TestRBACService_AddRole(t *testing.T) {
	svc := NewRBACService()
	role := Role{Name: "admin", Permissions: []Permission{{Instance: "*", Database: "*", Action: "*"}}}
	svc.AddRole(role)

	got, ok := svc.GetRole("admin")
	if !ok {
		t.Fatal("role should exist after AddRole")
	}
	if got.Name != "admin" {
		t.Errorf("role.Name = %q, want %q", got.Name, "admin")
	}
}

func TestRBACService_HasPermission(t *testing.T) {
	svc := NewRBACService()
	role := Role{
		Name:        "reader",
		Permissions: []Permission{{Instance: "db1", Database: "app", Action: "SELECT"}},
	}

	if !svc.HasPermission(role, Permission{Instance: "db1", Database: "app", Action: "SELECT"}) {
		t.Error("should have SELECT permission")
	}
	if svc.HasPermission(role, Permission{Instance: "db1", Database: "app", Action: "DELETE"}) {
		t.Error("should NOT have DELETE permission")
	}
}

func TestRBACService_UserHasPermission(t *testing.T) {
	svc := NewRBACService()
	svc.AddRole(Role{Name: "admin", Permissions: []Permission{
		{Instance: "*", Database: "*", Action: "SELECT"},
		{Instance: "*", Database: "*", Action: "DELETE"},
	}})
	svc.AddRole(Role{Name: "viewer", Permissions: []Permission{
		{Instance: "*", Database: "*", Action: "SELECT"},
	}})

	if !svc.UserHasPermission([]string{"admin"}, Permission{Instance: "*", Database: "*", Action: "DELETE"}) {
		t.Error("admin should have DELETE")
	}
	if !svc.UserHasPermission([]string{"viewer"}, Permission{Instance: "*", Database: "*", Action: "SELECT"}) {
		t.Error("viewer should have SELECT")
	}
	if svc.UserHasPermission([]string{"viewer"}, Permission{Instance: "*", Database: "*", Action: "DELETE"}) {
		t.Error("viewer should NOT have DELETE")
	}
	if !svc.UserHasPermission([]string{"guest", "admin"}, Permission{Instance: "*", Database: "*", Action: "SELECT"}) {
		t.Error("user with multiple roles should have permission if any role grants it")
	}
	if svc.UserHasPermission([]string{"nonexistent"}, Permission{Instance: "*", Database: "*", Action: "SELECT"}) {
		t.Error("nonexistent role should not have permission")
	}
	if svc.UserHasPermission(nil, Permission{Instance: "*", Database: "*", Action: "SELECT"}) {
		t.Error("nil roles should not have permission")
	}
}

func TestRBACService_ListRoles(t *testing.T) {
	svc := NewRBACService()
	svc.AddRole(Role{Name: "admin"})
	svc.AddRole(Role{Name: "viewer"})

	roles := svc.ListRoles()
	if len(roles) != 2 {
		t.Errorf("ListRoles count = %d, want %d", len(roles), 2)
	}
}

func TestRBACService_GetRole(t *testing.T) {
	svc := NewRBACService()
	svc.AddRole(Role{Name: "admin"})

	_, ok := svc.GetRole("admin")
	if !ok {
		t.Error("should find admin role")
	}

	_, ok = svc.GetRole("nonexistent")
	if ok {
		t.Error("should not find nonexistent role")
	}
}

func TestRBACService_DeleteRole(t *testing.T) {
	svc := NewRBACService()
	svc.AddRole(Role{Name: "admin"})
	svc.DeleteRole("admin")

	_, ok := svc.GetRole("admin")
	if ok {
		t.Error("role should be deleted")
	}
}

func TestRBACService_DeleteRole_Nonexistent(t *testing.T) {
	svc := NewRBACService()
	svc.DeleteRole("nonexistent")
}

func TestRBACService_AssignAndListUserRoles(t *testing.T) {
	svc := NewRBACService()
	svc.AssignRolesToUser("user1", []string{"admin", "viewer"})

	roles := svc.ListUserRoles("user1")
	if len(roles) != 2 {
		t.Errorf("user roles count = %d, want %d", len(roles), 2)
	}

	roles = svc.ListUserRoles("nonexistent")
	if len(roles) != 0 {
		t.Errorf("nonexistent user roles count = %d, want %d", len(roles), 0)
	}
}

func TestRBACService_RemoveUserRoles(t *testing.T) {
	svc := NewRBACService()
	svc.AssignRolesToUser("user1", []string{"admin"})
	svc.RemoveUserRoles("user1")

	roles := svc.ListUserRoles("user1")
	if len(roles) != 0 {
		t.Errorf("user roles after remove = %d, want %d", len(roles), 0)
	}
}

func TestPermissionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := NewRBACService()
	svc.AddRole(Role{Name: "admin", Permissions: []Permission{{Instance: "*", Database: "*", Action: "SELECT"}}})

	router := gin.New()
	router.GET("/test", svc.PermissionMiddleware(Permission{Instance: "*", Database: "*", Action: "SELECT"}), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	t.Run("with matching permission", func(t *testing.T) {
		router.GET("/auth-test", func(c *gin.Context) {
			c.Set("roles", []string{"admin"})
			c.Next()
		}, svc.PermissionMiddleware(Permission{Instance: "*", Database: "*", Action: "SELECT"}), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		req := httptest.NewRequest("GET", "/auth-test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("without roles context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("with wrong role", func(t *testing.T) {
		router.GET("/wrong-role-test", func(c *gin.Context) {
			c.Set("roles", []string{"guest"})
			c.Next()
		}, svc.PermissionMiddleware(Permission{Instance: "*", Database: "*", Action: "SELECT"}), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		req := httptest.NewRequest("GET", "/wrong-role-test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("with empty roles", func(t *testing.T) {
		router.GET("/empty-role-test", func(c *gin.Context) {
			c.Set("roles", []string{})
			c.Next()
		}, svc.PermissionMiddleware(Permission{Instance: "*", Database: "*", Action: "SELECT"}), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		req := httptest.NewRequest("GET", "/empty-role-test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})
}
