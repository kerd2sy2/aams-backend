package system

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/system/controller"
	"delivery-backend/internal/modules/system/repository"
	"delivery-backend/internal/modules/system/service"
)

type Module struct {
	DashboardService    service.DashboardService
	ReportService       service.ReportService
	AuditService        service.AuditService
	SettingService      service.SettingService
	NotificationService service.NotificationService
	ArchiveService      service.ArchiveService

	DashboardHandler    *controller.DashboardHandler
	ReportHandler       *controller.ReportHandler
	AuditHandler        *controller.AuditHandler
	SettingHandler      *controller.SettingHandler
	NotificationHandler *controller.NotificationHandler
	ArchiveHandler      *controller.ArchiveHandler
}

func NewModule(db *gorm.DB) *Module {
	repo := repository.NewSystemRepository(db)

	dashSvc := service.NewDashboardService(repo)
	repSvc := service.NewReportService(repo)
	auditSvc := service.NewAuditService(repo)
	settingSvc := service.NewSettingService(repo)
	notifSvc := service.NewNotificationService(repo)
	archiveSvc := service.NewArchiveService(repo)

	return &Module{
		DashboardService:    dashSvc,
		ReportService:       repSvc,
		AuditService:        auditSvc,
		SettingService:      settingSvc,
		NotificationService: notifSvc,
		ArchiveService:      archiveSvc,

		DashboardHandler:    controller.NewDashboardHandler(dashSvc),
		ReportHandler:       controller.NewReportHandler(repSvc),
		AuditHandler:        controller.NewAuditHandler(auditSvc),
		SettingHandler:      controller.NewSettingHandler(settingSvc),
		NotificationHandler: controller.NewNotificationHandler(notifSvc),
		ArchiveHandler:      controller.NewArchiveHandler(archiveSvc),
	}
}

// RegisterPublicRoutes registers unauthenticated public system routes
func (m *Module) RegisterPublicRoutes(r *gin.Engine) {
	r.GET("/api/v1/settings/public", m.SettingHandler.GetPublicSettings)
}

// RegisterRoutes registers protected routes under the authenticated /api/v1 router group
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	// Analytics & Dashboard
	rg.GET("/dashboard", m.DashboardHandler.GetStats)

	// Reports
	rg.GET("/reports", m.ReportHandler.GetReports)
	rg.GET("/reports/export", m.ReportHandler.ExportReports)
	rg.GET("/reports/daily", m.ReportHandler.GetDailyReport)
	rg.GET("/reports/daily/export", m.ReportHandler.ExportDailyReport)

	// Audit Logs
	rg.GET("/audit-logs", m.AuditHandler.GetLogs)
	rg.DELETE("/audit-logs/clear", m.AuditHandler.ClearLogs)
	rg.DELETE("/audit-logs/bulk", m.AuditHandler.BulkDeleteLogs)
	rg.DELETE("/audit-logs/:id", m.AuditHandler.DeleteLog)

	// Settings
	rg.GET("/settings", m.SettingHandler.GetSettings)
	rg.PUT("/settings", m.SettingHandler.UpdateSettings)

	// Notifications
	rg.GET("/notifications", m.NotificationHandler.GetMyNotifications)
	rg.PUT("/notifications/read-all", m.NotificationHandler.MarkAllAsRead)
	rg.PUT("/notifications/:id/read", m.NotificationHandler.MarkAsRead)

	// Broadcast & Survey Notifications
	rg.POST("/notifications/broadcast", m.NotificationHandler.SendBroadcast)
	rg.GET("/notifications/broadcasts", m.NotificationHandler.GetBroadcasts)
	rg.DELETE("/notifications/broadcasts/:id", m.NotificationHandler.DeleteBroadcast)
	rg.GET("/notifications/broadcasts/:id/votes", m.NotificationHandler.GetBroadcastVotes)
	rg.GET("/notifications/employee/broadcasts", m.NotificationHandler.GetEmployeeBroadcasts)
	rg.GET("/notifications/employee/unread", m.NotificationHandler.GetEmployeeUnreadBroadcasts)
	rg.POST("/notifications/employee/read/:id", m.NotificationHandler.MarkEmployeeBroadcastRead)
	rg.POST("/notifications/employee/read-all", m.NotificationHandler.MarkAllEmployeeBroadcastsRead)
	rg.POST("/notifications/employee/vote/:id", m.NotificationHandler.SubmitVote)
	rg.POST("/employees/me/push-token", m.NotificationHandler.SaveEmployeePushToken)

	// Archive & Trash
	rg.GET("/archive", m.ArchiveHandler.GetArchived)
	rg.POST("/archive/restore", m.ArchiveHandler.Restore)
	rg.DELETE("/archive/permanent", m.ArchiveHandler.PermanentDelete)
	rg.POST("/archive/restore-bulk", m.ArchiveHandler.BulkRestore)
	rg.DELETE("/archive/permanent-bulk", m.ArchiveHandler.BulkPermanentDelete)
}
