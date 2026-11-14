package operation_video

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/pkg/middleware"
)

type Module struct {
	Repo    OperationVideoRepository
	Service OperationVideoService
	Handler *OperationVideoHandler
}

func NewModule(db *gorm.DB, uploadBaseDir string) *Module {
	repo := NewOperationVideoRepository(db)
	svc := NewOperationVideoService(repo)
	handler := NewOperationVideoHandler(svc, uploadBaseDir)
	return &Module{
		Repo:    repo,
		Service: svc,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/operation-videos")
	{
		group.POST("/upload", m.Handler.Upload)
		group.GET("", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "MANAGEMENT"), m.Handler.List)
		group.GET("/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "MANAGEMENT"), m.Handler.GetByID)
		group.GET("/:id/stream", m.Handler.StreamVideo)
		group.PATCH("/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "MANAGEMENT"), m.Handler.Update)
		group.DELETE("/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN"), m.Handler.Delete)
	}
}
