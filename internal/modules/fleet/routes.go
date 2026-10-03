package fleet

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/fleet/contracts"
	"delivery-backend/internal/modules/fleet/controller"
	"delivery-backend/internal/modules/fleet/repository"
	"delivery-backend/internal/modules/fleet/service"
)

// Module encapsulates the Fleet & Vehicles bounded context
type Module struct {
	Handler  *controller.VehicleHandler
	Service  service.VehicleService
	Repo     repository.VehicleRepository
	Contract contracts.IVehicleFleetContract
}

// NewModule initializes the Fleet module with clean dependency injection
func NewModule(db *gorm.DB, storage contracts.IStorageContract, audit contracts.IAuditContract) *Module {
	repo := repository.NewVehicleRepository(db)
	svc := service.NewVehicleService(repo, storage)
	h := controller.NewVehicleHandler(svc, audit)
	return &Module{
		Handler:  h,
		Service:  svc,
		Repo:     repo,
		Contract: svc, // service implements IVehicleFleetContract
	}
}

// RegisterRoutes registers all fleet vehicle endpoints under the provided Gin Router Group
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	vehicles := rg.Group("/vehicles")
	{
		vehicles.GET("", m.Handler.GetAll)
		vehicles.POST("", m.Handler.Create)
		vehicles.GET("/check-km", m.Handler.CheckKM)
		vehicles.GET("/:id", m.Handler.GetByID)
		vehicles.PUT("/:id", m.Handler.Update)
		vehicles.DELETE("/:id", m.Handler.Delete)
		vehicles.POST("/:id/oil-change", m.Handler.RecordOilChange)
	}
}
