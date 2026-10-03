package ticket

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/ticket/contracts"
	"delivery-backend/internal/modules/ticket/controller"
	"delivery-backend/internal/modules/ticket/repository"
	"delivery-backend/internal/modules/ticket/service"
)

// Module encapsulates the Ticket bounded context
type Module struct {
	Handler *controller.SupportTicketHandler
	Service service.SupportTicketService
	Repo    repository.SupportTicketRepository
}

// NewModule initializes the Ticket module with clean dependency injection
func NewModule(db *gorm.DB, empContract contracts.IEmployeeContract, notifContract contracts.INotificationContract) *Module {
	repo := repository.NewSupportTicketRepository(db)
	svc := service.NewSupportTicketService(repo, empContract, notifContract)
	h := controller.NewSupportTicketHandler(svc)
	return &Module{
		Handler: h,
		Service: svc,
		Repo:    repo,
	}
}

// RegisterRoutes registers all ticket endpoints under the provided Gin Router Group
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	tickets := rg.Group("/tickets")
	{
		tickets.GET("", m.Handler.GetAll)
		tickets.POST("", m.Handler.Create)
		tickets.PUT("/:id", m.Handler.Update)
		tickets.DELETE("/:id", m.Handler.Delete)
	}
}
