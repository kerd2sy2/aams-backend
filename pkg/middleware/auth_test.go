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
