package system

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SystemRepository interface {
	// Audit
	CreateAuditLog(ctx context.Context, log *AuditLog) error
	FindAuditLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]AuditLog, int64, error)
	GetLatestAuditLogs(ctx context.Context, limit int, branchID *uuid.UUID) ([]AuditLog, error)
	DeleteAuditLog(ctx context.Context, id uuid.UUID) error
	BulkDeleteAuditLogs(ctx context.Context, ids []uuid.UUID) error
	ClearAuditLogs(ctx context.Context, branchID *uuid.UUID) error

	// Settings
	GetSettings(ctx context.Context) ([]AppSetting, error)
	GetSettingByKey(ctx context.Context, key string) (*AppSetting, error)
	UpsertSetting(ctx context.Context, setting *AppSetting) error

	// Notifications & Broadcasts
	FindNotificationsByAdmin(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID, status string) ([]Notification, error)
	MarkNotificationAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error
	MarkAllNotificationsAsRead(ctx context.Context, adminID uuid.UUID) error
	CreateNotification(ctx context.Context, notif *Notification) error
	CreateBroadcast(ctx context.Context, broadcast *BroadcastNotification) error
	FindBroadcasts(ctx context.Context, branchID *uuid.UUID, limit, offset int) ([]BroadcastNotification, int64, error)
	FindBroadcastByID(ctx context.Context, id uuid.UUID) (*BroadcastNotification, error)
	DeleteBroadcast(ctx context.Context, id uuid.UUID) error
	FindBroadcastsForEmployee(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time, limit int) ([]BroadcastItemDTO, error)
	FindUnreadBroadcastsForEmployee(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) ([]BroadcastItemDTO, error)
	MarkEmployeeBroadcastRead(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID) error
	MarkAllEmployeeBroadcastsRead(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) error
	SubmitVote(ctx context.Context, vote *BroadcastVote) error
	FindVotesForBroadcast(ctx context.Context, broadcastID uuid.UUID) ([]BroadcastVoteItemDTO, error)
	UpdateEmployeePushToken(ctx context.Context, empID uuid.UUID, token, language string) error

	// Archive
	GetArchivedItems(ctx context.Context, filter ArchiveFilter) ([]ArchivedItemDTO, int64, ArchiveStatsDTO, error)
	RestoreItem(ctx context.Context, itemType string, id uuid.UUID) error
	PermanentDeleteItem(ctx context.Context, itemType string, id uuid.UUID) error
	BulkRestoreItems(ctx context.Context, itemType string, ids []uuid.UUID) error
	BulkPermanentDeleteItems(ctx context.Context, itemType string, ids []uuid.UUID) error

	// Dashboard & Stats
	GetDashboardStats(ctx context.Context, branchID *uuid.UUID) (*DashboardStatsResponse, error)
	GetDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) (*DailyReportResponse, error)
	GetReports(ctx context.Context, filter ReportFilter) ([]map[string]interface{}, int64, error)
}

type gormSystemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) SystemRepository {
	return &gormSystemRepository{db: db}
}

func (r *gormSystemRepository) likeOperator() string {
	if r.db != nil && r.db.Dialector != nil && r.db.Dialector.Name() == "postgres" {
		return "ILIKE"
	}
	return "LIKE"
}
