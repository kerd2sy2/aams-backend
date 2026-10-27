package custody

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/custody/contracts"
	"delivery-backend/internal/modules/custody/controller"
	"delivery-backend/internal/modules/custody/repository"
	"delivery-backend/internal/modules/custody/service"
	"delivery-backend/pkg/middleware"
)

type Module struct {
	Handler  *controller.CustodyHandler
	Service  service.CustodyService
	Repo     repository.CustodyRepository
	Contract contracts.ICustodyContract
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewCustodyRepository(db)
	svc := service.NewCustodyService(repo)
	h := controller.NewCustodyHandler(svc)
	return &Module{
		Handler:  h,
		Service:  svc,
		Repo:     repo,
		Contract: svc,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	c := rg.Group("/custody")
	c.Use(middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"}, "apps.ohda", "app.ohda", "custody", "custody.access", "custody.view", "custody.manage"))
	{
		c.GET("", m.Handler.List)
		c.POST("", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN"}, "custody.create", "custody.manage"), m.Handler.Create)
		c.POST("/add-amount", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN"}, "custody.create", "custody.manage"), m.Handler.AddAmount)
		c.GET("/logs", m.Handler.GetLogs)
		c.DELETE("/logs/:id", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN"}, "custody.delete", "custody.manage"), m.Handler.DeleteLog)
		c.POST("/:id/expenses", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN"}, "custody.create", "custody.manage"), m.Handler.AddExpense)
		c.PUT("/expenses/:id", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN"}, "custody.manage"), m.Handler.UpdateExpense)
		c.DELETE("/expenses/:id", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN"}, "custody.delete", "custody.manage"), m.Handler.DeleteExpense)
		c.POST("/ledger-image", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN"}, "custody.create", "custody.manage"), m.Handler.SetLedgerImage)
	}
}
