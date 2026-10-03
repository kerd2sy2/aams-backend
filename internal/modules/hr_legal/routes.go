package hr_legal

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/hr_legal/contracts"
	"delivery-backend/internal/modules/hr_legal/controller"
	"delivery-backend/internal/modules/hr_legal/repository"
	"delivery-backend/internal/modules/hr_legal/service"
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
	rg.POST("/investigations", m.Handler.CreateInvestigation)
	rg.GET("/investigations", m.Handler.GetAllInvestigations)
	rg.GET("/investigations/pending-count", m.Handler.GetPendingInvestigationCount)
	rg.GET("/investigations/:id", m.Handler.GetInvestigationByID)
	rg.PUT("/investigations/:id", m.Handler.UpdateInvestigation)
	rg.POST("/investigations/:id/approve", m.Handler.ApproveInvestigation)

	// 2. Documents
	rg.GET("/documents", m.Handler.GetAllDocuments)
	rg.GET("/documents/expiring", m.Handler.GetExpiringDocuments)
	rg.GET("/documents/:id", m.Handler.GetDocumentByID)
	rg.POST("/documents", m.Handler.CreateDocument)
	rg.PUT("/documents/:id", m.Handler.UpdateDocument)
	rg.DELETE("/documents/:id", m.Handler.DeleteDocument)

	// 3. Bank Accounts
	rg.GET("/bank-accounts", m.Handler.GetAllBankAccounts)
	rg.POST("/bank-accounts", m.Handler.CreateBankAccount)
	rg.PUT("/bank-accounts/:id", m.Handler.UpdateBankAccount)
	rg.DELETE("/bank-accounts/:id", m.Handler.DeleteBankAccount)

	// 4. Leave Requests
	rg.GET("/leave-requests", m.Handler.GetAllLeaveRequests)
	rg.POST("/leave-requests", m.Handler.CreateLeaveRequest)
	rg.PUT("/leave-requests/:id", m.Handler.UpdateLeaveRequestStatus)
	rg.DELETE("/leave-requests/:id", m.Handler.DeleteLeaveRequest)
	rg.POST("/leave-requests/:id/approve", m.Handler.ApproveLeaveRequest)
	rg.POST("/leave-requests/:id/reject", m.Handler.RejectLeaveRequest)
	rg.GET("/leave-requests/pending-count", m.Handler.GetPendingLeaveCount)

	// Backward-compatible alias /leaves
	rg.GET("/leaves", m.Handler.GetAllLeaveRequests)
	rg.POST("/leaves", m.Handler.CreateLeaveRequest)
	rg.PUT("/leaves/:id/status", m.Handler.UpdateLeaveRequestStatus)
	rg.DELETE("/leaves/:id", m.Handler.DeleteLeaveRequest)

	// 5. Traffic Violations
	rg.GET("/violations", m.Handler.GetAllViolations)
	rg.POST("/violations", m.Handler.CreateViolation)
	rg.PUT("/violations/:id", m.Handler.UpdateViolation)
	rg.DELETE("/violations/:id", m.Handler.DeleteViolation)

	// 6. Fuel Logs
	rg.GET("/fuel-logs", m.Handler.GetAllFuelLogs)
	rg.POST("/fuel-logs", m.Handler.CreateFuelLog)
	rg.PUT("/fuel-logs/:id", m.Handler.UpdateFuelLog)
	rg.DELETE("/fuel-logs/:id", m.Handler.DeleteFuelLog)
}
