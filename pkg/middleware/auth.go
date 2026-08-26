package middleware

import (
	"context"
	"net/http"
	"strings"

	"delivery-backend/internal/repository"
	"delivery-backend/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(secret string, adminRepo repository.AdminRepository) gin.HandlerFunc {
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

		// Read current branch_id from DB (not JWT) to reflect real-time changes
		admin, err := adminRepo.FindByID(context.Background(), claims.AdminID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "المستخدم غير موجود",
			})
			c.Abort()
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("admin_email", claims.Email)
		c.Set("admin_name", admin.Name)
		c.Set("admin_role", admin.Role)
		c.Set("admin_role_id", admin.RoleID)
		c.Set("branch_id", admin.BranchID)
		c.Set("admin", admin)

		c.Next()

	}
}
