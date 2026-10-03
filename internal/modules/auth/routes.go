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
	repo := repository.NewAuthRepository(db)
	authSvc := service.NewAuthService(repo, cfg)
	otpSvc := service.NewOTPService(repo, cfg)
	roleSvc := service.NewRoleService(repo)
	adminSvc := service.NewAdminService(repo)
	branchSvc := service.NewBranchService(repo)

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
	// Me & Google Auth links
	rg.GET("/me", m.AuthHandler.Me)
	rg.POST("/auth/google/link", m.AuthHandler.LinkGoogle)
	rg.POST("/auth/google/unlink", m.AuthHandler.UnlinkGoogle)

	// User (Admin) management
	rg.GET("/users", m.AdminHandler.GetAll)
	rg.POST("/users", m.AdminHandler.Create)
	rg.PUT("/users/:id", m.AdminHandler.Update)
	rg.DELETE("/users/:id", m.AdminHandler.Delete)
	rg.POST("/users/change-password", m.AdminHandler.ChangePassword)

	// Roles & Permissions
	rg.GET("/roles", m.RoleHandler.GetAll)
	rg.POST("/roles", m.RoleHandler.Create)
	rg.GET("/roles/:id", m.RoleHandler.GetByID)
	rg.PUT("/roles/:id", m.RoleHandler.Update)
	rg.DELETE("/roles/:id", m.RoleHandler.Delete)
	rg.GET("/permissions", m.RoleHandler.GetPermissions)

	// Branch management
	rg.GET("/branches", m.BranchHandler.GetAll)
	rg.GET("/branches/:id", m.BranchHandler.GetByID)
	rg.POST("/branches", m.BranchHandler.Create)
	rg.PUT("/branches/:id", m.BranchHandler.Update)
	rg.DELETE("/branches/:id", m.BranchHandler.Delete)

	// OTP Requests & verification audit
	rg.GET("/otp-requests", m.OTPHandler.GetOTPList)
	rg.POST("/otp-requests/:id/cancel", m.OTPHandler.CancelOTP)
}
