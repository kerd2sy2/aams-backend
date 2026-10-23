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
	c.Use(middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"))
	{
		c.GET("", m.Handler.List)
		c.POST("", m.Handler.Create)
		c.POST("/add-amount", m.Handler.AddAmount)
		c.GET("/logs", m.Handler.GetLogs)
		c.DELETE("/logs/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeleteLog)
		c.POST("/:id/expenses", m.Handler.AddExpense)
		c.PUT("/expenses/:id", m.Handler.UpdateExpense)
		c.DELETE("/expenses/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.DeleteExpense)
		c.POST("/ledger-image", m.Handler.SetLedgerImage)
	}
}
