package auth

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/auth/controller"
	"delivery-backend/internal/modules/auth/repository"
	"delivery-backend/internal/modules/auth/service"
	"delivery-backend/pkg/config"
	"delivery-backend/pkg/middleware"
)

type Module struct {
	AuthService   service.AuthService
	OTPService    service.OTPService
	RoleService   service.RoleService
	AdminService  service.AdminService
	BranchService service.BranchService

	AuthHandler   *controller.AuthHandler
	OTPHandler    *controller.OTPHandler
	RoleHandler   *controller.RoleHandler
	AdminHandler  *controller.AdminHandler
	BranchHandler *controller.BranchHandler
}

func NewModule(db *gorm.DB, cfg *config.Config) *Module {
	adminRepo, roleRepo, otpRepo, branchRepo := repository.NewAuthRepository(db)
	authSvc := service.NewAuthService(adminRepo, branchRepo, cfg)
	otpSvc := service.NewOTPService(otpRepo, cfg)
	roleSvc := service.NewRoleService(roleRepo)
	adminSvc := service.NewAdminService(adminRepo)
	branchSvc := service.NewBranchService(branchRepo)

	return &Module{
		AuthService:   authSvc,
		OTPService:    otpSvc,
		RoleService:   roleSvc,
		AdminService:  adminSvc,
		BranchService: branchSvc,

		AuthHandler:   controller.NewAuthHandler(authSvc),
		OTPHandler:    controller.NewOTPHandler(otpSvc),
		RoleHandler:   controller.NewRoleHandler(roleSvc),
		AdminHandler:  controller.NewAdminHandler(adminSvc),
		BranchHandler: controller.NewBranchHandler(branchSvc),
	}
}

// RegisterPublicRoutes registers unauthenticated auth endpoints (with strict rate limiting where appropriate).
func (m *Module) RegisterPublicRoutes(r *gin.Engine) {
	apiV1 := r.Group("/api/v1")
	{
		apiV1.POST("/login", middleware.StrictLoginLimiter(), m.AuthHandler.Login)
		apiV1.POST("/auth/login", middleware.StrictLoginLimiter(), m.AuthHandler.Login)
		apiV1.POST("/refresh", m.AuthHandler.RefreshToken)
		apiV1.POST("/auth/refresh", m.AuthHandler.RefreshToken)
		apiV1.POST("/auth/google/login", middleware.StrictLoginLimiter(), m.AuthHandler.GoogleLogin)
		apiV1.POST("/auth/request-otp", middleware.StrictLoginLimiter(), m.OTPHandler.RequestOTP)
		apiV1.POST("/auth/verify-otp", middleware.StrictLoginLimiter(), m.OTPHandler.VerifyOTP)
	}

	// Compatibility aliases
	r.POST("/api/refresh", m.AuthHandler.RefreshToken)
	r.POST("/auth/refresh", m.AuthHandler.RefreshToken)
}

// RegisterRoutes registers protected routes under the authenticated /api/v1 group.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	// Me, Avatar & Google Auth links
	rg.GET("/me", m.AuthHandler.Me)
	rg.POST("/me/avatar", m.AuthHandler.UpdateAvatar)
	rg.PUT("/me/avatar", m.AuthHandler.UpdateAvatar)
	rg.POST("/auth/me/avatar", m.AuthHandler.UpdateAvatar)
	rg.PUT("/auth/me/avatar", m.AuthHandler.UpdateAvatar)
	rg.POST("/auth/google/link", m.AuthHandler.LinkGoogle)
	rg.POST("/auth/google/unlink", m.AuthHandler.UnlinkGoogle)

	// User (Admin) management - Strictly SUPER_ADMIN and ADMIN
	adminOnly := rg.Group("/users")
	adminOnly.Use(middleware.RequireRoles("SUPER_ADMIN", "ADMIN"))
	{
		adminOnly.GET("", m.AdminHandler.GetAll)
		adminOnly.POST("", m.AdminHandler.Create)
		adminOnly.POST("/promote", middleware.RequireRoles("SUPER_ADMIN"), m.AdminHandler.Promote)
		adminOnly.PUT("/:id", m.AdminHandler.Update)
		adminOnly.DELETE("/:id", m.AdminHandler.Delete)
		adminOnly.POST("/change-password", m.AdminHandler.ChangePassword)
	}

	// Roles & Permissions - Strictly SUPER_ADMIN and ADMIN
	roleOnly := rg.Group("/roles")
	roleOnly.Use(middleware.RequireRoles("SUPER_ADMIN", "ADMIN"))
	{
		roleOnly.GET("", m.RoleHandler.GetAll)
		roleOnly.POST("", m.RoleHandler.Create)
		roleOnly.GET("/:id", m.RoleHandler.GetByID)
		roleOnly.PUT("/:id", m.RoleHandler.Update)
		roleOnly.DELETE("/:id", m.RoleHandler.Delete)
	}
	rg.GET("/permissions", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.RoleHandler.GetPermissions)

	// Branch management
	rg.GET("/branches", m.BranchHandler.GetAll)
	rg.GET("/branches/:id", m.BranchHandler.GetByID)
	rg.POST("/branches", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.BranchHandler.Create)
	rg.PUT("/branches/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.BranchHandler.Update)
	rg.DELETE("/branches/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.BranchHandler.Delete)

	// OTP Requests & verification audit - Strictly SUPER_ADMIN and ADMIN
	rg.GET("/otp-requests", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.OTPHandler.GetOTPList)
	rg.POST("/otp-requests/:id/cancel", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.OTPHandler.CancelOTP)

	// Trusted Devices Management
	rg.GET("/devices/my-devices", m.OTPHandler.GetMyDevices)
	rg.DELETE("/devices/:uuid", m.OTPHandler.RevokeDevice)
	rg.GET("/auth/devices/my-devices", m.OTPHandler.GetMyDevices)
	rg.DELETE("/auth/devices/:uuid", m.OTPHandler.RevokeDevice)
}
