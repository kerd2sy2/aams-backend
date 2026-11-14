package ticket

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/pkg/middleware"
)

// Module encapsulates the Ticket bounded context (Elmansheya Standard)
type Module struct {
	Handler *SupportTicketHandler
	Service SupportTicketService
	Repo    SupportTicketRepository
}

// NewModule initializes the Ticket module with clean dependency injection
func NewModule(db *gorm.DB, empContract IEmployeeContract, notifContract INotificationContract) *Module {
	repo := NewSupportTicketRepository(db)
	svc := NewSupportTicketService(repo, empContract, notifContract)
	h := NewSupportTicketHandler(svc)
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
		tickets.PUT("/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.Update)
		tickets.DELETE("/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.Delete)
	}
}
