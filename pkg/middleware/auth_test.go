package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/domain"
	"delivery-backend/pkg/middleware"
)

func TestRequirePermission_ConfiguredExplicitZeroPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	role := &domain.Role{
		Name:        "ADMIN",
		Permissions: `["employees.view", "work.view"]`,
	}

	// User has permissions_configured=true but permissions="[]" (explicit zero permissions)
	configuredUser := &domain.Admin{
		ID:                    uuid.New(),
		Name:                  "Test Zero Perms",
		Role:                  "ADMIN",
		RoleObj:               role,
		PermissionsConfigured: true,
		Permissions:           "[]",
	}

	r.GET("/protected", func(c *gin.Context) {
		c.Set("admin", configuredUser)
		c.Next()
	}, middleware.RequirePermission("employees.view"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for explicit empty permissions, got %d", w.Code)
	}
}

func TestRequirePermission_LegacyFallbackToRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	role := &domain.Role{
		Name:        "SUPERVISOR",
		Permissions: `["employees.view"]`,
	}

	// Legacy user has permissions_configured=false -> should fallback to RoleObj permissions
	legacyUser := &domain.Admin{
		ID:                    uuid.New(),
		Name:                  "Test Legacy",
		Role:                  "SUPERVISOR",
		RoleObj:               role,
		PermissionsConfigured: false,
		Permissions:           "",
	}

	r.GET("/protected", func(c *gin.Context) {
		c.Set("admin", legacyUser)
		c.Next()
	}, middleware.RequirePermission("employees.view"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for legacy user falling back to role permissions, got %d", w.Code)
	}
}

func TestRequirePermission_SuperAdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	superAdmin := &domain.Admin{
		ID:                    uuid.New(),
		Name:                  "Super Admin",
		Role:                  "SUPER_ADMIN",
		PermissionsConfigured: true,
		Permissions:           "[]",
	}

	r.GET("/protected", func(c *gin.Context) {
		c.Set("admin", superAdmin)
		c.Next()
	}, middleware.RequirePermission("employees.delete"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for SUPER_ADMIN bypass, got %d", w.Code)
	}
}

func TestRequireRoles_ExecutiveBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rolesToTest := []string{"SUPER_ADMIN", "super_admin", "OWNER", "owner", "GENERAL_MANAGER", "general_manager"}
	for _, role := range rolesToTest {
		t.Run("role_"+role, func(t *testing.T) {
			r := gin.New()
			r.GET("/users-endpoint", func(c *gin.Context) {
				c.Set("admin_role", role)
				c.Next()
			}, middleware.RequireRoles("ADMIN"), func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/users-endpoint", nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected 200 OK for %s in RequireRoles, got %d", role, w.Code)
			}
		})
	}
}

func TestRequireRolesOrPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Allowed role passes", func(t *testing.T) {
		r := gin.New()
		r.GET("/users-endpoint", func(c *gin.Context) {
			c.Set("admin_role", "ADMIN")
			c.Next()
		}, middleware.RequireRolesOrPermission([]string{"ADMIN"}, "users.manage"), func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users-endpoint", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200 OK for ADMIN role, got %d", w.Code)
		}
	})

	t.Run("User with users.manage permission passes even with non-admin role", func(t *testing.T) {
		r := gin.New()
		supervisorWithPerm := &domain.Admin{
			ID:                    uuid.New(),
			Name:                  "Supervisor With Users Manage",
			Role:                  "SUPERVISOR",
			PermissionsConfigured: true,
			Permissions:           `["users.manage"]`,
		}

		r.GET("/users-endpoint", func(c *gin.Context) {
			c.Set("admin_role", "SUPERVISOR")
			c.Set("admin", supervisorWithPerm)
			c.Next()
		}, middleware.RequireRolesOrPermission([]string{"ADMIN"}, "users.manage"), func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users-endpoint", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200 OK for user with users.manage, got %d", w.Code)
		}
	})

	t.Run("User without role or permission is forbidden", func(t *testing.T) {
		r := gin.New()
		regularSupervisor := &domain.Admin{
			ID:                    uuid.New(),
			Name:                  "Regular Supervisor",
			Role:                  "SUPERVISOR",
			PermissionsConfigured: true,
			Permissions:           `["employees.view"]`,
		}

		r.GET("/users-endpoint", func(c *gin.Context) {
			c.Set("admin_role", "SUPERVISOR")
			c.Set("admin", regularSupervisor)
			c.Next()
		}, middleware.RequireRolesOrPermission([]string{"ADMIN"}, "users.manage"), func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users-endpoint", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %d", w.Code)
		}
	})
}
