package operation_video

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/operation_video/controller"
	"delivery-backend/internal/modules/operation_video/repository"
	"delivery-backend/internal/modules/operation_video/service"
)

type Module struct {
	Repo    repository.OperationVideoRepository
	Service service.OperationVideoService
	Handler *controller.OperationVideoHandler
}

func NewModule(db *gorm.DB, uploadBaseDir string) *Module {
	repo := repository.NewOperationVideoRepository(db)
	svc := service.NewOperationVideoService(repo)
	handler := controller.NewOperationVideoHandler(svc, uploadBaseDir)
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
		group.GET("", m.Handler.List)
		group.GET("/:id", m.Handler.GetByID)
		group.GET("/:id/stream", m.Handler.StreamVideo)
		group.PATCH("/:id", m.Handler.Update)
		group.DELETE("/:id", m.Handler.Delete)
	}
}
