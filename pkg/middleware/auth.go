package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/repository"
	"delivery-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type cachedAuthSession struct {
	adminID               uuid.UUID
	employeeID            *uuid.UUID
	email                 string
	name                  string
	role                  string
	roleID                *uuid.UUID
	branchID              *uuid.UUID
	mustChangePassword    bool
	permissionsConfigured bool
	admin                 *domain.Admin
	employee              *domain.Employee
	isEmployee            bool
	cachedAt              time.Time
}

var (
	authSessionCache sync.Map
	authLookupLocks  [64]sync.Mutex
	authCacheTTL     = 5 * time.Minute
)

func init() {
	// Background janitor to sweep stale auth cache entries every 5 minutes
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			now := time.Now()
			authSessionCache.Range(func(key, value interface{}) bool {
				if session, ok := value.(*cachedAuthSession); ok {
					if now.Sub(session.cachedAt) > 5*time.Minute {
						authSessionCache.Delete(key)
					}
				}
				return true
			})
		}
	}()
}

// InvalidateAuthCache clears a user's cached auth session immediately
func InvalidateAuthCache(id uuid.UUID) {
	authSessionCache.Delete(id)
}

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

		// Fast path: Check in-memory session cache to avoid hammering PostgreSQL on every request
		if val, ok := authSessionCache.Load(claims.AdminID); ok {
			if sess, ok := val.(*cachedAuthSession); ok && time.Since(sess.cachedAt) < authCacheTTL {
				applyAuthSessionToContext(c, sess)
				c.Next()
				return
			}
		}

		// Slow path: per-user lock avoids cache stampede when parallel requests arrive right after login or expiry
		lockIdx := claims.AdminID[0] & 63
		authLookupLocks[lockIdx].Lock()

		// Double-check cache after acquiring lock
		if val, ok := authSessionCache.Load(claims.AdminID); ok {
			if sess, ok := val.(*cachedAuthSession); ok && time.Since(sess.cachedAt) < authCacheTTL {
				authLookupLocks[lockIdx].Unlock()
				applyAuthSessionToContext(c, sess)
				c.Next()
				return
			}
		}

		// Resolve user from database using request context with resilient 10-second timeout
		dbCtx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)

		// 1. Try finding Admin from DB
		admin, err := adminRepo.FindByID(dbCtx, claims.AdminID)
		if err == nil && admin != nil {
			roleName := admin.Role
			if roleName == "" && admin.RoleObj != nil {
				roleName = admin.RoleObj.Name
			}
			effectiveBranchID := admin.BranchID
			if effectiveBranchID == nil && admin.EmployeeID != nil && empRepo != nil {
				if linkedEmp, empErr := empRepo.FindByID(dbCtx, *admin.EmployeeID); empErr == nil && linkedEmp != nil && linkedEmp.BranchID != nil {
					effectiveBranchID = linkedEmp.BranchID
					admin.BranchID = linkedEmp.BranchID
					if admin.Branch == nil {
						admin.Branch = linkedEmp.Branch
					}
				}
			}
			if effectiveBranchID == nil && claims.BranchID != nil && *claims.BranchID != uuid.Nil {
				effectiveBranchID = claims.BranchID
			}
			cancel()

			sess := &cachedAuthSession{
				adminID:               claims.AdminID,
				employeeID:            admin.EmployeeID,
				email:                 claims.Email,
				name:                  admin.Name,
				role:                  roleName,
				roleID:                admin.RoleID,
				branchID:              effectiveBranchID,
				mustChangePassword:    admin.MustChangePassword,
				permissionsConfigured: admin.PermissionsConfigured,
				admin:                 admin,
				isEmployee:            false,
				cachedAt:              time.Now(),
			}
			authSessionCache.Store(claims.AdminID, sess)
			authLookupLocks[lockIdx].Unlock()

			applyAuthSessionToContext(c, sess)
			c.Next()
			return
		}

		// 2. Try finding Employee (Delegate) from DB
		if empRepo != nil {
			emp, empErr := empRepo.FindByID(dbCtx, claims.AdminID)
			if empErr == nil && emp != nil {
				// Trim bulky blobs from cached struct so 5,000 sessions occupy minimal RAM
				emp.Barcode = ""
				emp.QRCode = ""

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
					linkedAdmin, _ = adminRepo.FindByEmployeeID(dbCtx, emp.ID)
					if linkedAdmin == nil && emp.Phone != "" {
						linkedAdmin, _ = adminRepo.FindByPhone(dbCtx, emp.Phone)
					}
				}

				effectiveBranchID := emp.BranchID
				effectiveBranch := emp.Branch
				if effectiveBranchID == nil && linkedAdmin != nil && linkedAdmin.BranchID != nil {
					effectiveBranchID = linkedAdmin.BranchID
					effectiveBranch = linkedAdmin.Branch
					emp.BranchID = effectiveBranchID
					if emp.Branch == nil {
						emp.Branch = effectiveBranch
					}
				}
				if effectiveBranchID == nil && claims.BranchID != nil && *claims.BranchID != uuid.Nil {
					effectiveBranchID = claims.BranchID
				}

				effectivePermissions := emp.Permissions
				permissionsConfigured := emp.Permissions != ""
				if linkedAdmin != nil {
					effectivePermissions = linkedAdmin.Permissions
					permissionsConfigured = linkedAdmin.PermissionsConfigured
					if linkedAdmin.Role != "" {
						empRole = linkedAdmin.Role
					}
				}
				if !permissionsConfigured && effectivePermissions == "" && emp.RoleObj != nil && emp.RoleObj.Permissions != "" {
					effectivePermissions = emp.RoleObj.Permissions
				}

				syntheticAdmin := &domain.Admin{
					ID:                    emp.ID,
					Name:                  emp.Name,
					Phone:                 emp.Phone,
					Role:                  empRole,
					RoleID:                emp.RoleID,
					RoleObj:               emp.RoleObj,
					BranchID:              effectiveBranchID,
					Branch:                effectiveBranch,
					Permissions:           effectivePermissions,
					PermissionsConfigured: permissionsConfigured,
					MustChangePassword:    emp.MustChangePassword,
				}
				if linkedAdmin != nil && linkedAdmin.RoleObj != nil && syntheticAdmin.RoleObj == nil {
					syntheticAdmin.RoleObj = linkedAdmin.RoleObj
				}
				cancel()

				sess := &cachedAuthSession{
					adminID:               claims.AdminID,
					employeeID:            &claims.AdminID,
					email:                 claims.Email,
					name:                  emp.Name,
					role:                  empRole,
					roleID:                emp.RoleID,
					branchID:              effectiveBranchID,
					mustChangePassword:    emp.MustChangePassword,
					permissionsConfigured: permissionsConfigured,
					admin:                 syntheticAdmin,
					employee:              emp,
					isEmployee:            true,
					cachedAt:              time.Now(),
				}
				authSessionCache.Store(claims.AdminID, sess)
				authLookupLocks[lockIdx].Unlock()

				applyAuthSessionToContext(c, sess)
				c.Next()
				return
			}
		}
		cancel()
		authLookupLocks[lockIdx].Unlock()

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "المستخدم غير موجود",
		})
		c.Abort()
	}
}

func applyAuthSessionToContext(c *gin.Context, sess *cachedAuthSession) {
	c.Set("admin_id", sess.adminID)
	if sess.employeeID != nil {
		c.Set("employee_id", *sess.employeeID)
	}
	c.Set("admin_email", sess.email)
	c.Set("admin_name", sess.name)
	c.Set("admin_role", sess.role)
	c.Set("admin_role_id", sess.roleID)
	c.Set("branch_id", sess.branchID)
	c.Set("must_change_password", sess.mustChangePassword)
	c.Set("permissions_configured", sess.permissionsConfigured)
	c.Set("admin", sess.admin)
	c.Set("is_employee", sess.isEmployee)
	if sess.isEmployee && sess.employee != nil {
		c.Set("employee", sess.employee)
	}
}

