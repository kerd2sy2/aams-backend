package inventory

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/inventory/contracts"
	"delivery-backend/internal/modules/inventory/controller"
	"delivery-backend/internal/modules/inventory/repository"
	"delivery-backend/internal/modules/inventory/service"
	"delivery-backend/pkg/middleware"
)

type Module struct {
	Handler  *controller.InventoryHandler
	Service  service.InventoryService
	Repo     repository.InventoryRepository
	Contract contracts.IInventoryStockContract
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewInventoryRepository(db)
	svc := service.NewInventoryService(repo)
	h := controller.NewInventoryHandler(svc)
	return &Module{
		Handler:  h,
		Service:  svc,
		Repo:     repo,
		Contract: svc, // service implements IInventoryStockContract
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	inv := rg.Group("/inventory")
	inv.Use(middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"))
	{
		inv.GET("/items", m.Handler.GetItems)
		inv.GET("/items/:id", m.Handler.GetItemByID)
		inv.GET("/barcode", m.Handler.FindByBarcode)
		inv.POST("/items", m.Handler.CreateItem)
		inv.PUT("/items/:id", m.Handler.UpdateItem)
		inv.DELETE("/items/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeleteItem)
		inv.POST("/add-stock", m.Handler.AddStock)
		inv.POST("/remove-stock", m.Handler.RemoveStock)
		inv.POST("/dispense-oil", m.Handler.DispenseOil)
		inv.GET("/transactions", m.Handler.GetTransactions)
		inv.DELETE("/transactions", middleware.RequireRoles("SUPER_ADMIN"), m.Handler.DeleteAllTransactions)
		inv.GET("/purchases", m.Handler.GetPurchaseInvoices)
		inv.GET("/purchases/:id", m.Handler.GetPurchaseInvoiceByID)
		inv.POST("/purchases", m.Handler.CreatePurchaseInvoice)
		inv.DELETE("/purchases/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeletePurchaseInvoice)
	}
}
