package custody

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/custody/contracts"
	"delivery-backend/internal/modules/custody/controller"
	"delivery-backend/internal/modules/custody/repository"
	"delivery-backend/internal/modules/custody/service"
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
	{
		c.GET("", m.Handler.List)
		c.POST("", m.Handler.Create)
		c.POST("/add-amount", m.Handler.AddAmount)
		c.GET("/logs", m.Handler.GetLogs)
		c.DELETE("/logs/:id", m.Handler.DeleteLog)
		c.POST("/:id/expenses", m.Handler.AddExpense)
		c.DELETE("/expenses/:id", m.Handler.DeleteExpense)
	}
}
