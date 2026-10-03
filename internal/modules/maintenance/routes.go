package maintenance

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/maintenance/controller"
	"delivery-backend/internal/modules/maintenance/repository"
	"delivery-backend/internal/modules/maintenance/service"
)

type Module struct {
	Service service.MaintenanceService
	Handler *controller.MaintenanceHandler
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewMaintenanceRepository(db)
	svc := service.NewMaintenanceService(repo)
	h := controller.NewMaintenanceHandler(svc)
	return &Module{
		Service: svc,
		Handler: h,
	}
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/maintenance/logs", m.Handler.GetAllLogs)
	r.GET("/maintenance/employee-logs", m.Handler.GetEmployeeLogs)

	r.GET("/maintenance-requests", m.Handler.GetAllRequests)
	r.POST("/maintenance-requests", m.Handler.CreateRequest)
	r.PUT("/maintenance-requests/:id", m.Handler.UpdateRequest)
	r.DELETE("/maintenance-requests/:id", m.Handler.DeleteRequest)
}
