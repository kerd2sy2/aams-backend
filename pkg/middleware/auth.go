package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/repository"
	"delivery-backend/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(secret string, adminRepo repository.AdminRepository, empRepo repository.EmployeeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "لم يتم التوفير رمز المصادقة (Missing Authorization Header)",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "صيغة رمز المصادقة غير صالحة (Invalid Bearer Token Format)",
			})
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(parts[1], secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "جلسة الدخول منتهية أو غير صالحة (Invalid or Expired Token)",
			})
			c.Abort()
			return
		}

		// 1. Try finding Admin from DB
		admin, err := adminRepo.FindByID(context.Background(), claims.AdminID)
		if err == nil && admin != nil {
			c.Set("admin_id", claims.AdminID)
			c.Set("admin_email", claims.Email)
			c.Set("admin_name", admin.Name)
			c.Set("admin_role", admin.Role)
			c.Set("admin_role_id", admin.RoleID)
			c.Set("branch_id", admin.BranchID)
			c.Set("must_change_password", admin.MustChangePassword)
			c.Set("permissions_configured", admin.PermissionsConfigured)
			c.Set("admin", admin)
			c.Set("is_employee", false)
			c.Next()
			return
		}

		// 2. Try finding Employee (Delegate) from DB
		if empRepo != nil {
			emp, empErr := empRepo.FindByID(context.Background(), claims.AdminID)
			if empErr == nil && emp != nil {
				c.Set("admin_id", claims.AdminID)
				c.Set("employee_id", claims.AdminID)
				c.Set("admin_email", claims.Email)
				c.Set("admin_name", emp.Name)
				c.Set("admin_role", "DRIVER")
				c.Set("branch_id", emp.BranchID)
				c.Set("must_change_password", emp.MustChangePassword)
				c.Set("permissions_configured", false)
				c.Set("employee", emp)
				c.Set("is_employee", true)
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "المستخدم غير موجود",
		})
		c.Abort()
	}
}

// RequireRoles restricts endpoint access to specified roles (e.g. "ADMIN", "SUPERVISOR")
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("admin_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: لا توجد صلاحيات محددة"})
			c.Abort()
			return
		}

		userRole, _ := roleVal.(string)
		// SUPER_ADMIN has access to everything
		if userRole == "SUPER_ADMIN" {
			c.Next()
			return
		}

		for _, r := range allowedRoles {
			if strings.EqualFold(userRole, r) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "غير مصرح: هذه العملية تتطلب صلاحية (" + strings.Join(allowedRoles, " أو ") + ")",
		})
		c.Abort()
	}
}

func checkAdminPermission(admin *domain.Admin, requiredPerm string) bool {
	if admin == nil {
		return false
	}
	roleUpper := strings.ToUpper(strings.TrimSpace(admin.Role))
	if roleUpper == "SUPER_ADMIN" {
		return true
	}

	if admin.PermissionsConfigured {
		// When permissions are explicitly configured, only the explicit permissions apply
		var userPerms []string
		if admin.Permissions != "" {
			_ = json.Unmarshal([]byte(admin.Permissions), &userPerms)
		}
		for _, p := range userPerms {
			if p == "*" || strings.EqualFold(p, requiredPerm) {
				return true
			}
		}
		// If configured, empty or missing permission means strictly forbidden (no fallback)
		return false
	}

	// Legacy unconfigured account: fallback to role permissions
	if admin.RoleObj != nil && admin.RoleObj.Permissions != "" {
		var rolePerms []string
		if err := json.Unmarshal([]byte(admin.RoleObj.Permissions), &rolePerms); err == nil {
			for _, p := range rolePerms {
				if p == "*" || strings.EqualFold(p, requiredPerm) {
					return true
				}
			}
		}
	}

	// Default fallback for legacy ADMIN role
	if roleUpper == "ADMIN" {
		return true
	}

	return false
}

// RequirePermission checks whether the authenticated admin has a specific permission
func RequirePermission(requiredPerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminVal, exists := c.Get("admin")
		if !exists || adminVal == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: هذه العملية مخصصة للمشرفين والمديرين"})
			c.Abort()
			return
		}

		admin, ok := adminVal.(*domain.Admin)
		if !ok || admin == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: خطأ في بيانات المصادقة"})
			c.Abort()
			return
		}

		if !checkAdminPermission(admin, requiredPerm) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "غير مصرح: لا تملك الصلاحية المطلوبة (" + requiredPerm + ")",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission checks whether the authenticated admin has any of the specified permissions
func RequireAnyPermission(requiredPerms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminVal, exists := c.Get("admin")
		if !exists || adminVal == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: هذه العملية مخصصة للمشرفين والمديرين"})
			c.Abort()
			return
		}

		admin, ok := adminVal.(*domain.Admin)
		if !ok || admin == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: خطأ في بيانات المصادقة"})
			c.Abort()
			return
		}

		for _, perm := range requiredPerms {
			if checkAdminPermission(admin, perm) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "غير مصرح: لا تملك أي من الصلاحيات المطلوبة (" + strings.Join(requiredPerms, ", ") + ")",
		})
		c.Abort()
	}
}

