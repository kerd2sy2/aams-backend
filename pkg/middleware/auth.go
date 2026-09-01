package middleware

import (
	"context"
	"net/http"
	"strings"

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
