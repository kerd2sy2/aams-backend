package employee

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/employee/controller"
	"delivery-backend/internal/modules/employee/repository"
	"delivery-backend/internal/modules/employee/service"
	sharedRepo "delivery-backend/internal/repository"
	"delivery-backend/pkg/middleware"
)

type Module struct {
	Service service.EmployeeService
	Handler *controller.EmployeeHandler
}

func NewModule(db *gorm.DB, storage controller.StorageProvider) *Module {
	repo := repository.NewEmployeeRepository(db)
	adminRepo := sharedRepo.NewAdminRepository(db)
	svc := service.NewEmployeeService(repo, adminRepo)
	handler := controller.NewEmployeeHandler(svc, storage)

	return &Module{
		Service: svc,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	// Management & Supervisor Employee Administration
	rg.POST("/employees", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.Create)
	rg.GET("/employees", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"), m.Handler.GetAll)
	rg.GET("/employees/search", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"), m.Handler.Search)
	rg.GET("/employees/working", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"), m.Handler.GetWorking)
	rg.GET("/employees/:id", m.Handler.GetByID) // Handles self-view if employee, full branch view if supervisor
	rg.PUT("/employees/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.Update)
	rg.DELETE("/employees/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.Delete)
	rg.POST("/employees/batch-oil-setup", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.BatchSetOilChange)
	rg.GET("/employees/locations", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetLocations)
	rg.PUT("/employees/:id/phone", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.SetPhone)
	rg.POST("/employees/:id/reset-password", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.ResetPassword)
	rg.GET("/employees/:id/barcode", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetBarcode)
	rg.GET("/employees/:id/qrcode", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetQRCode)
	rg.GET("/employees/:id/print-card", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetPrintCard)

	// Self-service (Open to authenticated employees / couriers)
	rg.POST("/employees/me/change-password", m.Handler.ChangeMyPassword)
	rg.POST("/employees/me/phone", m.Handler.SetMyPhone)
	rg.POST("/employees/me/location", m.Handler.SetMyLocation)
	rg.POST("/upload", m.Handler.UploadImage)
	rg.POST("/upload-file", m.Handler.UploadFile)
}
