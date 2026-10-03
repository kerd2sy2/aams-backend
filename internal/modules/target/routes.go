package target

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/target/controller"
	"delivery-backend/internal/modules/target/repository"
	"delivery-backend/internal/modules/target/service"
	"delivery-backend/pkg/middleware"
)

type Module struct {
	TargetService      service.TargetService
	ExcelImportService service.ExcelImportService
	Handler            *controller.TargetHandler
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewTargetRepository(db)
	targetSvc := service.NewTargetService(repo)
	excelSvc := service.NewExcelImportService(repo)
	handler := controller.NewTargetHandler(targetSvc, excelSvc)

	return &Module{
		TargetService:      targetSvc,
		ExcelImportService: excelSvc,
		Handler:            handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	// 10. Identifier Target & Excel Import System
	targetRoutes := rg.Group("/target")
	{
		// Read & Dashboard access for Admin and Supervisor
		targetRoutes.GET("/dashboard", m.Handler.GetDashboardSummary)
		targetRoutes.GET("/identifiers", m.Handler.ListIdentifiers)
		targetRoutes.GET("/identifiers/:id", m.Handler.GetIdentifierDetails)
		targetRoutes.GET("/drivers", m.Handler.ListDrivers)
		targetRoutes.GET("/alerts", m.Handler.ListAlerts)
		targetRoutes.PATCH("/alerts/resolve-all", m.Handler.ResolveAllAlerts)
		targetRoutes.PATCH("/alerts/:id/resolve", m.Handler.ResolveAlert)
		targetRoutes.GET("/settings", m.Handler.GetTargetSettings)
		targetRoutes.GET("/batches", m.Handler.ListImportBatches)

		// Admin-only management routes
		adminTarget := targetRoutes.Group("")
		adminTarget.Use(middleware.RequireRoles("ADMIN", "SUPER_ADMIN"))
		{
			adminTarget.POST("/import/preview", m.Handler.PreviewExcelImport)
			adminTarget.POST("/import/confirm", m.Handler.ConfirmExcelImport)
			adminTarget.DELETE("/batches/:id", m.Handler.DeleteImportBatch)
			adminTarget.DELETE("/batches/date/:orderDate", m.Handler.DeleteSheetByDate)
			adminTarget.POST("/identifiers", m.Handler.CreateIdentifier)
			adminTarget.PUT("/identifiers/:id", m.Handler.UpdateIdentifier)
			adminTarget.DELETE("/identifiers", m.Handler.DeleteAllIdentifiers)
			adminTarget.DELETE("/identifiers/wipe-all", m.Handler.DeleteAllIdentifiers)
			adminTarget.DELETE("/identifiers/:id", m.Handler.DeleteIdentifier)
			adminTarget.PUT("/settings", m.Handler.UpdateTargetSettings)
		}
	}

	// Alias for /api/v1/admin/target/import
	adminImport := rg.Group("/admin/target/import")
	adminImport.Use(middleware.RequireRoles("ADMIN", "SUPER_ADMIN"))
	{
		adminImport.POST("/preview", m.Handler.PreviewExcelImport)
		adminImport.POST("/confirm", m.Handler.ConfirmExcelImport)
		adminImport.GET("/batches", m.Handler.ListImportBatches)
		adminImport.DELETE("/batches/:id", m.Handler.DeleteImportBatch)
		adminImport.DELETE("/batches/date/:orderDate", m.Handler.DeleteSheetByDate)
	}
}
