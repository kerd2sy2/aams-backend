package work

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	
	
	
	
	"delivery-backend/pkg/middleware"
)

type Module struct {
	Service WorkService
	Handler *WorkHandler
}

func NewModule(db *gorm.DB, storage StorageProvider, audit ...IAuditContract) *Module {
	repo := NewWorkRepository(db)
	empAdapter := &gormEmployeeAdapter{db: db}
	vehicleAdapter := &gormVehicleAdapter{db: db}
	svc := NewWorkService(repo, empAdapter, vehicleAdapter, storage)
	var auditContract IAuditContract
	if len(audit) > 0 {
		auditContract = audit[0]
	}
	handler := NewWorkHandler(svc, auditContract)
	return &Module{
		Service: svc,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/work/start", m.Handler.StartWork)
	rg.POST("/work/end", m.Handler.EndWork)
	rg.POST("/work/manual-entry", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"), m.Handler.SaveManualEntries)
	rg.PUT("/work/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.UpdateWorkSession)
	rg.PUT("/work/:id/review", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.ReviewWorkSession)
	rg.PUT("/work/sessions/:id/review", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.ReviewWorkSession)
	rg.GET("/work/sessions/:id", m.Handler.GetSessionByID)
	rg.GET("/work/active", m.Handler.GetActiveSession)
	rg.GET("/work/last-km", m.Handler.GetLastKM)
	rg.POST("/work/scan-plate", m.Handler.ScanPlate)
	rg.GET("/work/today-count", m.Handler.TodayCount)
	rg.GET("/work/check-oil", m.Handler.CheckOilChange)

	// Supervisor shift routes (Strictly restricted to supervisor / management roles)
	supervisorWork := rg.Group("/work/supervisor")
	supervisorWork.Use(middleware.RequireRoles("SUPERVISOR", "ADMIN", "SUPER_ADMIN"))
	{
		supervisorWork.POST("/start-shift", m.Handler.SupervisorStartShift)
		supervisorWork.POST("/end-shift", m.Handler.SupervisorEndShift)
	}
}
