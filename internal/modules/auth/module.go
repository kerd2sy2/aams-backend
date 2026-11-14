package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	
	
	
	sharedRepo "delivery-backend/internal/repository"
	"delivery-backend/pkg/config"
	"delivery-backend/pkg/middleware"
)

type Module struct {
	AuthService   AuthService
	OTPService    OTPService
	RoleService   RoleService
	AdminService  AdminService
	BranchService BranchService

	AuthHandler   *AuthHandler
	OTPHandler    *OTPHandler
	RoleHandler   *RoleHandler
	AdminHandler  *AdminHandler
	BranchHandler *BranchHandler
}

func NewModule(db *gorm.DB, cfg *config.Config) *Module {
	adminRepo, roleRepo, otpRepo, branchRepo := NewAuthRepository(db)
	empRepo := sharedRepo.NewEmployeeRepository(db)
	authSvc := NewAuthService(adminRepo, branchRepo, cfg)
	otpSvc := NewOTPService(otpRepo, cfg)
	roleSvc := NewRoleService(roleRepo)
	adminSvc := NewAdminService(adminRepo, empRepo)
	branchSvc := NewBranchService(branchRepo)

	return &Module{
		AuthService:   authSvc,
		OTPService:    otpSvc,
		RoleService:   roleSvc,
		AdminService:  adminSvc,
		BranchService: branchSvc,

		AuthHandler:   NewAuthHandler(authSvc),
		OTPHandler:    NewOTPHandler(otpSvc),
		RoleHandler:   NewRoleHandler(roleSvc),
		AdminHandler:  NewAdminHandler(adminSvc),
		BranchHandler: NewBranchHandler(branchSvc),
	}
}

// RegisterPublicRoutes registers unauthenticated auth endpoints (with strict rate limiting and bulkhead isolation).
func (m *Module) RegisterPublicRoutes(r *gin.Engine) {
	pubGroup := r.Group("")
	pubGroup.Use(
		middleware.ModuleIsolationMiddleware("auth"),
		middleware.ModuleTimeoutMiddleware(15*time.Second),
	)

	apiV1 := pubGroup.Group("/api/v1")
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
	pubGroup.POST("/api/refresh", m.AuthHandler.RefreshToken)
	pubGroup.POST("/auth/refresh", m.AuthHandler.RefreshToken)
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

	// User (Admin) management - Executive roles, ADMIN, or users with users.manage / users.view
	adminOnly := rg.Group("/users")
	adminOnly.Use(middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}, "users.manage", "users.view"))
	{
		adminOnly.GET("", m.AdminHandler.GetAll)
		adminOnly.POST("", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}, "users.manage"), m.AdminHandler.Create)
		adminOnly.POST("/promote", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "OWNER"}, "users.manage"), m.AdminHandler.Promote)
		adminOnly.PUT("/:id", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}, "users.manage"), m.AdminHandler.Update)
		adminOnly.DELETE("/:id", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER"}, "users.manage"), m.AdminHandler.Delete)
		adminOnly.POST("/change-password", m.AdminHandler.ChangePassword)
	}

	// Roles & Permissions - Executive roles, ADMIN, or roles.manage / users.manage
	roleOnly := rg.Group("/roles")
	roleOnly.Use(middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}, "roles.manage", "roles.view", "users.manage"))
	{
		roleOnly.GET("", m.RoleHandler.GetAll)
		roleOnly.POST("", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER"}, "roles.manage"), m.RoleHandler.Create)
		roleOnly.GET("/:id", m.RoleHandler.GetByID)
		roleOnly.PUT("/:id", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER"}, "roles.manage"), m.RoleHandler.Update)
		roleOnly.DELETE("/:id", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER"}, "roles.manage"), m.RoleHandler.Delete)
	}
	rg.GET("/permissions", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}, "roles.manage", "roles.view", "users.manage"), m.RoleHandler.GetPermissions)

	// Branch management
	rg.GET("/branches", m.BranchHandler.GetAll)
	rg.GET("/branches/:id", m.BranchHandler.GetByID)
	rg.POST("/branches", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}, "branches.manage"), m.BranchHandler.Create)
	rg.PUT("/branches/:id", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}, "branches.manage"), m.BranchHandler.Update)
	rg.DELETE("/branches/:id", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER"}, "branches.manage"), m.BranchHandler.Delete)

	// OTP Requests & verification audit
	rg.GET("/otp-requests", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}), m.OTPHandler.GetOTPList)
	rg.POST("/otp-requests/:id/cancel", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}), m.OTPHandler.CancelOTP)

	// Trusted Devices Management
	rg.GET("/devices/my-devices", m.OTPHandler.GetMyDevices)
	rg.DELETE("/devices/:uuid", m.OTPHandler.RevokeDevice)
	rg.GET("/auth/devices/my-devices", m.OTPHandler.GetMyDevices)
	rg.DELETE("/auth/devices/:uuid", m.OTPHandler.RevokeDevice)
}
