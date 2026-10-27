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
		var tokenStr string
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}
		// Security: Restrict query parameter tokens exclusively to media streaming paths
		// where browser/native video players cannot supply custom Authorization headers.
		if tokenStr == "" {
			reqPath := c.Request.URL.Path
			isMediaStreamPath := strings.HasPrefix(reqPath, "/api/v1/operation-videos/") ||
				strings.HasSuffix(reqPath, "/stream")

			if isMediaStreamPath {
				tokenStr = c.Query("token")
				// Redact token from URL query string so downstream loggers do not leak sensitive JWTs
				if tokenStr != "" && c.Request.URL != nil && c.Request.URL.RawQuery != "" {
					q := c.Request.URL.Query()
					if q.Has("token") {
						q.Set("token", "[REDACTED]")
						c.Request.URL.RawQuery = q.Encode()
					}
				}
			}
		}

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "لم يتم التوفير رمز المصادقة (Missing Authorization Token)",
			})
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(tokenStr, secret)
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
			roleName := admin.Role
			if roleName == "" && admin.RoleObj != nil {
				roleName = admin.RoleObj.Name
			}
			c.Set("admin_id", claims.AdminID)
			if admin.EmployeeID != nil {
				c.Set("employee_id", *admin.EmployeeID)
			}
			c.Set("admin_email", claims.Email)
			c.Set("admin_name", admin.Name)
			c.Set("admin_role", roleName)
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
				empRole := strings.TrimSpace(emp.JobRole)
				if empRole == "" {
					if emp.RoleName != "" {
						empRole = strings.TrimSpace(emp.RoleName)
					} else {
						empRole = "DRIVER"
					}
				}

				// Check if there is an Admin linked to this Employee
				var linkedAdmin *domain.Admin
				if adminRepo != nil {
					linkedAdmin, _ = adminRepo.FindByEmployeeID(context.Background(), emp.ID)
				}

				effectivePermissions := emp.Permissions
				permissionsConfigured := emp.Permissions != ""
				if linkedAdmin != nil {
					if linkedAdmin.Permissions != "" {
						effectivePermissions = linkedAdmin.Permissions
						permissionsConfigured = true
					}
					if linkedAdmin.Role != "" {
						empRole = linkedAdmin.Role
					}
				}

				syntheticAdmin := &domain.Admin{
					ID:                    emp.ID,
					Name:                  emp.Name,
					Phone:                 emp.Phone,
					Role:                  empRole,
					RoleID:                emp.RoleID,
					RoleObj:               emp.RoleObj,
					BranchID:              emp.BranchID,
					Branch:                emp.Branch,
					Permissions:           effectivePermissions,
					PermissionsConfigured: permissionsConfigured,
					MustChangePassword:    emp.MustChangePassword,
				}
				if linkedAdmin != nil && linkedAdmin.RoleObj != nil && syntheticAdmin.RoleObj == nil {
					syntheticAdmin.RoleObj = linkedAdmin.RoleObj
				}

				c.Set("admin_id", claims.AdminID)
				c.Set("employee_id", claims.AdminID)
				c.Set("admin_email", claims.Email)
				c.Set("admin_name", emp.Name)
				c.Set("admin_role", empRole)
				c.Set("admin_role_id", emp.RoleID)
				c.Set("branch_id", emp.BranchID)
				c.Set("must_change_password", emp.MustChangePassword)
				c.Set("permissions_configured", permissionsConfigured)
				c.Set("admin", syntheticAdmin)
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

		// Wildcard '*' permission grants full access
		if admin != nil && checkAdminPermission(admin, "*") {
			c.Next()
			return
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
		for _, p := range userPerms {
			trimmedP := strings.TrimSpace(p)
			if trimmedP == "*" || strings.EqualFold(trimmedP, requiredPerm) {
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

