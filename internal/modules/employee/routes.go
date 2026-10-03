package employee

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/employee/controller"
	"delivery-backend/internal/modules/employee/repository"
	"delivery-backend/internal/modules/employee/service"
)

type Module struct {
	Service service.EmployeeService
	Handler *controller.EmployeeHandler
}

func NewModule(db *gorm.DB, storage controller.StorageProvider) *Module {
	repo := repository.NewEmployeeRepository(db)
	svc := service.NewEmployeeService(repo)
	handler := controller.NewEmployeeHandler(svc, storage)

	return &Module{
		Service: svc,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/employees", m.Handler.Create)
	rg.GET("/employees", m.Handler.GetAll)
	rg.GET("/employees/search", m.Handler.Search)
	rg.GET("/employees/working", m.Handler.GetWorking)
	rg.GET("/employees/:id", m.Handler.GetByID)
	rg.PUT("/employees/:id", m.Handler.Update)
	rg.DELETE("/employees/:id", m.Handler.Delete)
	rg.POST("/employees/batch-oil-setup", m.Handler.BatchSetOilChange)
	rg.POST("/employees/me/change-password", m.Handler.ChangeMyPassword)
	rg.POST("/employees/me/phone", m.Handler.SetMyPhone)
	rg.POST("/employees/me/location", m.Handler.SetMyLocation)
	rg.GET("/employees/locations", m.Handler.GetLocations)
	rg.PUT("/employees/:id/phone", m.Handler.SetPhone)
	rg.POST("/employees/:id/reset-password", m.Handler.ResetPassword)
	rg.GET("/employees/:id/barcode", m.Handler.GetBarcode)
	rg.GET("/employees/:id/qrcode", m.Handler.GetQRCode)
	rg.GET("/employees/:id/print-card", m.Handler.GetPrintCard)
	rg.POST("/upload", m.Handler.UploadImage)
	rg.POST("/upload-file", m.Handler.UploadFile)
}
