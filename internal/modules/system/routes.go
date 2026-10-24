package system

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/system/controller"
	"delivery-backend/internal/modules/system/repository"
	"delivery-backend/internal/modules/system/service"
	"delivery-backend/pkg/middleware"
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
	// Analytics & Dashboard - Management and Supervisors
	rg.GET("/dashboard", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"), m.DashboardHandler.GetStats)

	// Reports - Management and Supervisors
	reports := rg.Group("/reports")
	reports.Use(middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"))
	{
		reports.GET("", m.ReportHandler.GetReports)
		reports.GET("/export", m.ReportHandler.ExportReports)
		reports.GET("/daily", m.ReportHandler.GetDailyReport)
		reports.GET("/daily/export", m.ReportHandler.ExportDailyReport)
	}

	// Audit Logs - Executive roles, ADMIN (View), or audit_logs.view permission. Clear/Bulk strictly SUPER_ADMIN and OWNER.
	auditLogs := rg.Group("/audit-logs")
	{
		auditLogs.GET("", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}, "audit_logs.view"), m.AuditHandler.GetLogs)
		auditLogs.DELETE("/clear", middleware.RequireRoles("SUPER_ADMIN", "OWNER"), m.AuditHandler.ClearLogs)
		auditLogs.DELETE("/bulk", middleware.RequireRoles("SUPER_ADMIN", "OWNER"), m.AuditHandler.BulkDeleteLogs)
		auditLogs.DELETE("/:id", middleware.RequireRoles("SUPER_ADMIN", "OWNER"), m.AuditHandler.DeleteLog)
	}

	// Settings - Viewing allowed, Updating strictly Executive roles, ADMIN, or settings.manage permission
	rg.GET("/settings", m.SettingHandler.GetSettings)
	rg.PUT("/settings", middleware.RequireRolesOrPermission([]string{"SUPER_ADMIN", "ADMIN", "OWNER", "GENERAL_MANAGER"}, "settings.manage"), m.SettingHandler.UpdateSettings)

	// Notifications - Personal
	rg.GET("/notifications", m.NotificationHandler.GetMyNotifications)
	rg.PUT("/notifications/read-all", m.NotificationHandler.MarkAllAsRead)
	rg.PUT("/notifications/:id/read", m.NotificationHandler.MarkAsRead)

	// Broadcast & Survey Notifications - Sending/deleting restricted to Management/Supervisors
	rg.POST("/notifications/broadcast", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.NotificationHandler.SendBroadcast)
	rg.GET("/notifications/broadcasts", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.NotificationHandler.GetBroadcasts)
	rg.DELETE("/notifications/broadcasts/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.NotificationHandler.DeleteBroadcast)
	rg.GET("/notifications/broadcasts/:id/votes", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.NotificationHandler.GetBroadcastVotes)
	rg.GET("/notifications/employee/broadcasts", m.NotificationHandler.GetEmployeeBroadcasts)
	rg.GET("/notifications/employee/unread", m.NotificationHandler.GetEmployeeUnreadBroadcasts)
	rg.POST("/notifications/employee/read/:id", m.NotificationHandler.MarkEmployeeBroadcastRead)
	rg.POST("/notifications/employee/read-all", m.NotificationHandler.MarkAllEmployeeBroadcastsRead)
	rg.POST("/notifications/employee/vote/:id", m.NotificationHandler.SubmitVote)
	rg.POST("/employees/me/push-token", m.NotificationHandler.SaveEmployeePushToken)

	// Archive & Trash - Strictly SUPER_ADMIN and ADMIN
	archive := rg.Group("/archive")
	archive.Use(middleware.RequireRoles("SUPER_ADMIN", "ADMIN"))
	{
		archive.GET("", m.ArchiveHandler.GetArchived)
		archive.POST("/restore", m.ArchiveHandler.Restore)
		archive.DELETE("/permanent", m.ArchiveHandler.PermanentDelete)
		archive.POST("/restore-bulk", m.ArchiveHandler.BulkRestore)
		archive.DELETE("/permanent-bulk", m.ArchiveHandler.BulkPermanentDelete)
	}
}
