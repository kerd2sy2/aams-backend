package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"delivery-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

// isExecutiveRole checks whether the role string is a top-level system management role.
func isExecutiveRole(role string) bool {
	trimmed := strings.TrimSpace(role)
	return strings.EqualFold(trimmed, "SUPER_ADMIN") ||
		strings.EqualFold(trimmed, "OWNER") ||
		strings.EqualFold(trimmed, "GENERAL_MANAGER")
}

// RequireRoles restricts endpoint access to specified roles (e.g. "ADMIN", "SUPERVISOR")
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("admin_role")
		var userRole string
		if exists {
			userRole, _ = roleVal.(string)
		}

		adminVal, _ := c.Get("admin")
		admin, _ := adminVal.(*domain.Admin)

		if userRole == "" && admin != nil {
			userRole = admin.Role
			if userRole == "" && admin.RoleObj != nil {
				userRole = admin.RoleObj.Name
			}
		}

		if userRole == "" && !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: لا توجد صلاحيات محددة"})
			c.Abort()
			return
		}

		// SUPER_ADMIN, OWNER, and GENERAL_MANAGER have full access to everything
		if isExecutiveRole(userRole) {
			c.Next()
			return
		}

		if admin != nil && admin.RoleObj != nil && isExecutiveRole(admin.RoleObj.Name) {
			c.Next()
			return
		}

		// Wildcard '*' permission grants full access
		if admin != nil && checkAdminPermission(admin, "*") {
			c.Next()
			return
		}

		for _, r := range allowedRoles {
			if strings.EqualFold(userRole, r) {
				c.Next()
				return
			}
			if admin != nil && admin.RoleObj != nil && strings.EqualFold(admin.RoleObj.Name, r) {
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

// RequireRolesOrPermission restricts endpoint access to specified roles OR any of the specified permissions.
func RequireRolesOrPermission(allowedRoles []string, allowedPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("admin_role")
		var userRole string
		if exists {
			userRole, _ = roleVal.(string)
		}

		adminVal, _ := c.Get("admin")
		admin, _ := adminVal.(*domain.Admin)

		if userRole == "" && admin != nil {
			userRole = admin.Role
			if userRole == "" && admin.RoleObj != nil {
				userRole = admin.RoleObj.Name
			}
		}

		// Executive roles have full access
		if isExecutiveRole(userRole) {
			c.Next()
			return
		}

		if admin != nil && admin.RoleObj != nil && isExecutiveRole(admin.RoleObj.Name) {
			c.Next()
			return
		}

		// Wildcard '*' permission grants full access only if account is unconfigured or '*' is the sole permission
		if admin != nil && checkAdminPermission(admin, "*") {
			userPerms := parsePermissionsList(admin.Permissions)
			if !admin.PermissionsConfigured || (len(userPerms) == 1 && strings.TrimSpace(userPerms[0]) == "*") {
				c.Next()
				return
			}
		}

		// When permissions are explicitly configured on this account:
		// If specific permissions are required by this endpoint, the explicit permissions MUST be checked.
		// Generic role names (e.g. ADMIN, SUPERVISOR) must NOT bypass configured permissions.
		if admin != nil && admin.PermissionsConfigured && len(allowedPermissions) > 0 {
			for _, perm := range allowedPermissions {
				if checkAdminPermission(admin, perm) {
					c.Next()
					return
				}
			}
			allRequired := append(allowedRoles, allowedPermissions...)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "غير مصرح: لا تملك الصلاحية المطلوبة (" + strings.Join(allRequired, " أو ") + ")",
			})
			c.Abort()
			return
		}

		// Legacy or unconfigured account: check allowed roles
		for _, r := range allowedRoles {
			if strings.EqualFold(userRole, r) {
				c.Next()
				return
			}
			if admin != nil && admin.RoleObj != nil && strings.EqualFold(admin.RoleObj.Name, r) {
				c.Next()
				return
			}
		}

		// Check allowed permissions
		if admin != nil {
			for _, perm := range allowedPermissions {
				if checkAdminPermission(admin, perm) {
					c.Next()
					return
				}
			}
		}

		allRequired := append(allowedRoles, allowedPermissions...)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "غير مصرح: هذه العملية تتطلب صلاحية (" + strings.Join(allRequired, " أو ") + ")",
		})
		c.Abort()
	}
}

func parsePermissionsList(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "[]" || trimmed == "null" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
		return arr
	}
	// Fallback to comma-separated
	var res []string
	for _, part := range strings.Split(trimmed, ",") {
		cleaned := strings.Trim(strings.TrimSpace(part), `[]"' `)
		if cleaned != "" {
			res = append(res, cleaned)
		}
	}
	return res
}

func checkAdminPermission(admin *domain.Admin, requiredPerm string) bool {
	if admin == nil {
		return false
	}
	roleUpper := strings.ToUpper(strings.TrimSpace(admin.Role))
	if roleUpper == "" && admin.RoleObj != nil {
		roleUpper = strings.ToUpper(strings.TrimSpace(admin.RoleObj.Name))
	}
	if isExecutiveRole(roleUpper) {
		return true
	}
	if admin.RoleObj != nil && isExecutiveRole(admin.RoleObj.Name) {
		return true
	}

	if admin.PermissionsConfigured {
		// When permissions are explicitly configured, only the explicit permissions apply
		userPerms := parsePermissionsList(admin.Permissions)
		if len(userPerms) == 1 && strings.TrimSpace(userPerms[0]) == "*" {
			return true
		}
		for _, p := range userPerms {
			trimmedP := strings.TrimSpace(p)
			if trimmedP != "*" && strings.EqualFold(trimmedP, requiredPerm) {
				return true
			}
		}
		// If configured, empty or missing permission means strictly forbidden (no fallback)
		return false
	}

	// Legacy unconfigured account: fallback to role permissions
	if admin.RoleObj != nil && admin.RoleObj.Permissions != "" {
		rolePerms := parsePermissionsList(admin.RoleObj.Permissions)
		for _, p := range rolePerms {
			trimmedP := strings.TrimSpace(p)
			if trimmedP == "*" || strings.EqualFold(trimmedP, requiredPerm) {
				return true
			}
		}
	}

	// Default fallback for legacy ADMIN role
	if roleUpper == "ADMIN" || (admin.RoleObj != nil && strings.ToUpper(strings.TrimSpace(admin.RoleObj.Name)) == "ADMIN") {
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

