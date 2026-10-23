package hr_legal

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/hr_legal/contracts"
	"delivery-backend/internal/modules/hr_legal/controller"
	"delivery-backend/internal/modules/hr_legal/repository"
	"delivery-backend/internal/modules/hr_legal/service"
	"delivery-backend/pkg/middleware"
)

type Module struct {
	Handler  *controller.HRLegalHandler
	Service  service.HRLegalService
	Repo     repository.HRLegalRepository
	Contract contracts.IHRLegalContract
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewHRLegalRepository(db)
	svc := service.NewHRLegalService(repo)
	h := controller.NewHRLegalHandler(svc)
	return &Module{
		Handler:  h,
		Service:  svc,
		Repo:     repo,
		Contract: svc,
	}
}

func (m *Module) RegisterPublicRoutes(r *gin.Engine) {
	r.GET("/api/v1/public/doc/:id", m.Handler.GetPublicByID)
	r.GET("/api/v1/public/investigations/:id", m.Handler.GetPublicByID)
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	// 1. Investigations
	rg.POST("/investigations", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.CreateInvestigation)
	rg.GET("/investigations", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetAllInvestigations)
	rg.GET("/investigations/pending-count", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetPendingInvestigationCount)
	rg.GET("/investigations/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetInvestigationByID)
	rg.PUT("/investigations/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.UpdateInvestigation)
	rg.DELETE("/investigations/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeleteInvestigation)
	rg.POST("/investigations/:id/approve", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.ApproveInvestigation)

	// 2. Documents
	rg.GET("/documents", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetAllDocuments)
	rg.GET("/documents/expiring", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetExpiringDocuments)
	rg.GET("/documents/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetDocumentByID)
	rg.POST("/documents", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.CreateDocument)
	rg.PUT("/documents/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.UpdateDocument)
	rg.DELETE("/documents/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeleteDocument)

	// 3. Bank Accounts
	rg.GET("/bank-accounts", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"), m.Handler.GetAllBankAccounts)
	rg.POST("/bank-accounts", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"), m.Handler.CreateBankAccount)
	rg.PUT("/bank-accounts/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"), m.Handler.UpdateBankAccount)
	rg.DELETE("/bank-accounts/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeleteBankAccount)

	// 4. Leave Requests
	rg.GET("/leave-requests", m.Handler.GetAllLeaveRequests)
	rg.POST("/leave-requests", m.Handler.CreateLeaveRequest)
	rg.PUT("/leave-requests/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.UpdateLeaveRequestStatus)
	rg.DELETE("/leave-requests/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeleteLeaveRequest)
	rg.POST("/leave-requests/:id/approve", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.ApproveLeaveRequest)
	rg.POST("/leave-requests/:id/reject", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.RejectLeaveRequest)
	rg.GET("/leave-requests/pending-count", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetPendingLeaveCount)

	// Backward-compatible alias /leaves
	rg.GET("/leaves", m.Handler.GetAllLeaveRequests)
	rg.POST("/leaves", m.Handler.CreateLeaveRequest)
	rg.PUT("/leaves/:id/status", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.UpdateLeaveRequestStatus)
	rg.DELETE("/leaves/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeleteLeaveRequest)

	// 5. Traffic Violations
	rg.GET("/violations", m.Handler.GetAllViolations)
	rg.POST("/violations", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.CreateViolation)
	rg.PUT("/violations/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.UpdateViolation)
	rg.DELETE("/violations/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeleteViolation)

	// 6. Fuel Logs
	rg.GET("/fuel-logs", m.Handler.GetAllFuelLogs)
	rg.POST("/fuel-logs", m.Handler.CreateFuelLog)
	rg.PUT("/fuel-logs/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.UpdateFuelLog)
	rg.DELETE("/fuel-logs/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeleteFuelLog)
}
