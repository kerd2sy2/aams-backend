package fleet

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/pkg/middleware"
)

// Module encapsulates the Fleet & Vehicles bounded context
type Module struct {
	Handler  *VehicleHandler
	Service  VehicleService
	Repo     VehicleRepository
	Contract IVehicleFleetContract
}

// NewModule initializes the Fleet module with clean dependency injection
func NewModule(db *gorm.DB, storage IStorageContract, audit IAuditContract, inv IInventoryContract) *Module {
	repo := NewVehicleRepository(db)
	svc := NewVehicleService(repo, storage, inv)
	h := NewVehicleHandler(svc, audit)
	return &Module{
		Handler:  h,
		Service:  svc,
		Repo:     repo,
		Contract: svc,
	}
}

// RegisterRoutes registers all fleet vehicle endpoints under the provided Gin Router Group
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	vehicles := rg.Group("/vehicles")
	{
		vehicles.GET("", m.Handler.GetAll)
		vehicles.POST("", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.Create)
		vehicles.GET("/check-km", m.Handler.CheckKM)
		vehicles.GET("/:id", m.Handler.GetByID)
		vehicles.PUT("/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.Update)
		vehicles.DELETE("/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.Delete)
		vehicles.POST("/:id/oil-change", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.RecordOilChange)
	}
}
