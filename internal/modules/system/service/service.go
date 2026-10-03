package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/system/contracts"
	"delivery-backend/internal/modules/system/domain"
	"delivery-backend/internal/modules/system/dto"
	"delivery-backend/internal/modules/system/repository"
)

// ---------------- Dashboard Service ----------------
type DashboardService interface {
	GetStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardStatsResponse, error)
}

type dashboardService struct {
	repo repository.SystemRepository
}

func NewDashboardService(repo repository.SystemRepository) DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) GetStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardStatsResponse, error) {
	return s.repo.GetDashboardStats(ctx, branchID)
}

// ---------------- Report Service ----------------
type ReportService interface {
	GetReports(ctx context.Context, filter dto.ReportFilter) ([]map[string]interface{}, int64, error)
	ExportReports(ctx context.Context, filter dto.ReportFilter) ([]byte, error)
	GetDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) (*dto.DailyReportResponse, error)
	ExportDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) ([]byte, error)
}

type reportService struct {
	repo repository.SystemRepository
}

func NewReportService(repo repository.SystemRepository) ReportService {
	return &reportService{repo: repo}
}

func (s *reportService) GetReports(ctx context.Context, filter dto.ReportFilter) ([]map[string]interface{}, int64, error) {
	return s.repo.GetReports(ctx, filter)
}

func (s *reportService) ExportReports(ctx context.Context, filter dto.ReportFilter) ([]byte, error) {
	filter.Limit = 10000
	rows, _, err := s.repo.GetReports(ctx, filter)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.WriteString("\xEF\xBB\xBF") // UTF-8 BOM for Excel
	w := csv.NewWriter(buf)

	_ = w.Write([]string{"المعرف", "المندوب", "الهوية", "الفرع", "التاريخ", "الطلبات", "المسافة (كم)", "تكلفة الوقود", "مدة العمل", "الحالة"})
	for _, r := range rows {
		_ = w.Write([]string{
			fmt.Sprintf("%v", r["id"]),
			fmt.Sprintf("%v", r["employee_name"]),
			fmt.Sprintf("%v", r["national_id"]),
			fmt.Sprintf("%v", r["branch_name"]),
			fmt.Sprintf("%v", r["start_time"]),
			fmt.Sprintf("%v", r["orders_count"]),
			fmt.Sprintf("%v", r["distance"]),
			fmt.Sprintf("%v", r["fuel_cost"]),
			fmt.Sprintf("%v", r["working_duration"]),
			fmt.Sprintf("%v", r["status"]),
		})
	}
	w.Flush()
	return buf.Bytes(), nil
}

func (s *reportService) GetDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) (*dto.DailyReportResponse, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	return s.repo.GetDailyReport(ctx, dateStr, branchID)
}

func (s *reportService) ExportDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) ([]byte, error) {
	rep, err := s.GetDailyReport(ctx, dateStr, branchID)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(buf)

	_ = w.Write([]string{"المندوب", "الفرع", "التطبيق", "الدباب", "الطلبات", "المسافة (كم)", "الوقود", "مدة العمل", "الحالة"})
	for _, e := range rep.Employees {
		_ = w.Write([]string{
			e.EmployeeName,
			e.BranchName,
			e.AppName,
			e.MotorcycleNumber,
			fmt.Sprintf("%d", e.OrdersCount),
			fmt.Sprintf("%.1f", e.Distance),
			fmt.Sprintf("%.1f", e.FuelCost),
			e.WorkingDuration,
			e.Status,
		})
	}
	w.Flush()
	return buf.Bytes(), nil
}

// ---------------- Audit Service ----------------
type AuditService interface {
	contracts.IAuditLogger
	GetLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error)
	ClearLogs(ctx context.Context, branchID *uuid.UUID) error
	BulkDeleteLogs(ctx context.Context, ids []uuid.UUID) error
	DeleteLog(ctx context.Context, id uuid.UUID) error
	LogAction(ctx context.Context, adminName, action, details, ipAddress string, branchID *uuid.UUID) error
}

type auditService struct {
	repo repository.SystemRepository
}

func NewAuditService(repo repository.SystemRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) Log(ctx context.Context, adminName, action, details, ipAddress string, branchID *uuid.UUID) error {
	return s.LogAction(ctx, adminName, action, details, ipAddress, branchID)
}

func (s *auditService) LogAction(ctx context.Context, adminName, action, details, ipAddress string, branchID *uuid.UUID) error {
	log := &domain.AuditLog{
		AdminName: adminName,
		Action:    action,
		Details:   details,
		IPAddress: ipAddress,
		BranchID:  branchID,
		CreatedAt: time.Now(),
	}
	return s.repo.CreateAuditLog(ctx, log)
}

func (s *auditService) GetLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error) {
	return s.repo.FindAuditLogs(ctx, branchID, page, limit)
}

func (s *auditService) ClearLogs(ctx context.Context, branchID *uuid.UUID) error {
	return s.repo.ClearAuditLogs(ctx, branchID)
}

func (s *auditService) BulkDeleteLogs(ctx context.Context, ids []uuid.UUID) error {
	return s.repo.BulkDeleteAuditLogs(ctx, ids)
}

func (s *auditService) DeleteLog(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteAuditLog(ctx, id)
}

// ---------------- Setting Service ----------------
type SettingService interface {
	GetAllSettings(ctx context.Context) (dto.AppSettingsResponse, error)
	GetSettingByKey(ctx context.Context, key string) (*domain.AppSetting, error)
	UpdateAppSettings(ctx context.Context, req dto.UpdateAppSettingsRequest) error
	UpdateSetting(ctx context.Context, key, value string) error
}

type settingService struct {
	repo repository.SystemRepository
}

func NewSettingService(repo repository.SystemRepository) SettingService {
	return &settingService{repo: repo}
}

func (s *settingService) GetAllSettings(ctx context.Context) (dto.AppSettingsResponse, error) {
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return dto.AppSettingsResponse{}, err
	}

	resp := dto.AppSettingsResponse{
		SiteName: "AAMS Logistics System",
	}

	for _, set := range settings {
		if set.Key == "site_name" {
			resp.SiteName = set.Value
		} else if set.Key == "logo_url" {
			resp.LogoURL = set.Value
		}
	}
	return resp, nil
}

func (s *settingService) GetSettingByKey(ctx context.Context, key string) (*domain.AppSetting, error) {
	return s.repo.GetSettingByKey(ctx, key)
}

func (s *settingService) UpdateAppSettings(ctx context.Context, req dto.UpdateAppSettingsRequest) error {
	if req.SiteName != "" {
		_ = s.repo.UpsertSetting(ctx, &domain.AppSetting{Key: "site_name", Value: req.SiteName})
	}
	if req.LogoURL != "" {
		_ = s.repo.UpsertSetting(ctx, &domain.AppSetting{Key: "logo_url", Value: req.LogoURL})
	}
	return nil
}

func (s *settingService) UpdateSetting(ctx context.Context, key, value string) error {
	return s.repo.UpsertSetting(ctx, &domain.AppSetting{Key: key, Value: value})
}

// ---------------- Notification Service ----------------
type NotificationService interface {
	contracts.INotificationSender
	GetMyNotifications(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID, status string) ([]domain.Notification, error)
	MarkAllAsRead(ctx context.Context, adminID uuid.UUID) error
	MarkAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error
	SendBroadcast(ctx context.Context, req dto.CreateBroadcastRequest, createdBy string) (*domain.BroadcastNotification, error)
	GetBroadcasts(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.BroadcastNotification, int64, error)
	DeleteBroadcast(ctx context.Context, id uuid.UUID) error
	GetBroadcastVotes(ctx context.Context, broadcastID uuid.UUID) ([]dto.BroadcastVoteItemDTO, error)
	GetEmployeeBroadcasts(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time, limit int) ([]dto.BroadcastItemDTO, error)
	GetEmployeeUnreadBroadcasts(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) ([]dto.BroadcastItemDTO, error)
	MarkEmployeeBroadcastRead(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID) error
	MarkAllEmployeeBroadcastsRead(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) error
	SubmitVote(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID, req dto.SubmitPollVoteRequest) error
	SaveEmployeePushToken(ctx context.Context, empID uuid.UUID, req dto.PushTokenRequest) error
}

type notificationService struct {
	repo repository.SystemRepository
}

func NewNotificationService(repo repository.SystemRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) SendNotification(ctx context.Context, branchID, adminID, employeeID *uuid.UUID, title, body, notifType string) error {
	notif := &domain.Notification{
		BranchID:   branchID,
		AdminID:    adminID,
		EmployeeID: employeeID,
		Title:      title,
		Body:       body,
		Type:       notifType,
		Status:     "unread",
		CreatedAt:  time.Now(),
	}
	return s.repo.CreateNotification(ctx, notif)
}

func (s *notificationService) GetMyNotifications(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID, status string) ([]domain.Notification, error) {
	return s.repo.FindNotificationsByAdmin(ctx, adminID, branchID, status)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, adminID uuid.UUID) error {
	return s.repo.MarkAllNotificationsAsRead(ctx, adminID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	return s.repo.MarkNotificationAsRead(ctx, id, adminID)
}

func (s *notificationService) SendBroadcast(ctx context.Context, req dto.CreateBroadcastRequest, createdBy string) (*domain.BroadcastNotification, error) {
	broadcast := &domain.BroadcastNotification{
		Title:          req.Title,
		Body:           req.Body,
		TitleAr:        req.TitleAr,
		TitleEn:        req.TitleEn,
		TitleBn:        req.TitleBn,
		BodyAr:         req.BodyAr,
		BodyEn:         req.BodyEn,
		BodyBn:         req.BodyBn,
		ImageURL:       req.ImageURL,
		Target:         req.Target,
		BranchID:       req.BranchID,
		CreatedBy:      createdBy,
		HasPoll:        req.HasPoll,
		PollQuestion:   req.PollQuestion,
		PollQuestionAr: req.PollQuestionAr,
		PollQuestionEn: req.PollQuestionEn,
		PollQuestionBn: req.PollQuestionBn,
		CreatedAt:      time.Now(),
	}
	if broadcast.Target == "" {
		broadcast.Target = "ALL"
	}
	err := s.repo.CreateBroadcast(ctx, broadcast)
	return broadcast, err
}

func (s *notificationService) GetBroadcasts(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.BroadcastNotification, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	return s.repo.FindBroadcasts(ctx, branchID, limit, offset)
}

func (s *notificationService) DeleteBroadcast(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteBroadcast(ctx, id)
}

func (s *notificationService) GetBroadcastVotes(ctx context.Context, broadcastID uuid.UUID) ([]dto.BroadcastVoteItemDTO, error) {
	return s.repo.FindVotesForBroadcast(ctx, broadcastID)
}

func (s *notificationService) GetEmployeeBroadcasts(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time, limit int) ([]dto.BroadcastItemDTO, error) {
	return s.repo.FindBroadcastsForEmployee(ctx, empID, branchID, registeredAt, limit)
}

func (s *notificationService) GetEmployeeUnreadBroadcasts(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) ([]dto.BroadcastItemDTO, error) {
	return s.repo.FindUnreadBroadcastsForEmployee(ctx, empID, branchID, registeredAt)
}

func (s *notificationService) MarkEmployeeBroadcastRead(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID) error {
	return s.repo.MarkEmployeeBroadcastRead(ctx, empID, broadcastID)
}

func (s *notificationService) MarkAllEmployeeBroadcastsRead(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) error {
	return s.repo.MarkAllEmployeeBroadcastsRead(ctx, empID, branchID, registeredAt)
}

func (s *notificationService) SubmitVote(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID, req dto.SubmitPollVoteRequest) error {
	vote := &domain.BroadcastVote{
		BroadcastID: broadcastID,
		EmployeeID:  empID,
		Response:    req.Response,
		Reason:      req.Reason,
		CreatedAt:   time.Now(),
	}
	return s.repo.SubmitVote(ctx, vote)
}

func (s *notificationService) SaveEmployeePushToken(ctx context.Context, empID uuid.UUID, req dto.PushTokenRequest) error {
	return s.repo.UpdateEmployeePushToken(ctx, empID, req.PushToken)
}

// ---------------- Archive Service ----------------
type ArchiveService interface {
	GetArchivedItems(ctx context.Context, filter dto.ArchiveFilter) (dto.ArchiveResponseDTO, error)
	Restore(ctx context.Context, itemType string, id uuid.UUID) error
	PermanentDelete(ctx context.Context, itemType string, id uuid.UUID) error
	BulkRestore(ctx context.Context, itemType string, ids []uuid.UUID) error
	BulkPermanentDelete(ctx context.Context, itemType string, ids []uuid.UUID) error
}

type archiveService struct {
	repo repository.SystemRepository
}

func NewArchiveService(repo repository.SystemRepository) ArchiveService {
	return &archiveService{repo: repo}
}

func (s *archiveService) GetArchivedItems(ctx context.Context, filter dto.ArchiveFilter) (dto.ArchiveResponseDTO, error) {
	items, total, stats, err := s.repo.GetArchivedItems(ctx, filter)
	if err != nil {
		return dto.ArchiveResponseDTO{}, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	return dto.ArchiveResponseDTO{
		Data:       items,
		Stats:      stats,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *archiveService) Restore(ctx context.Context, itemType string, id uuid.UUID) error {
	return s.repo.RestoreItem(ctx, itemType, id)
}

func (s *archiveService) PermanentDelete(ctx context.Context, itemType string, id uuid.UUID) error {
	return s.repo.PermanentDeleteItem(ctx, itemType, id)
}

func (s *archiveService) BulkRestore(ctx context.Context, itemType string, ids []uuid.UUID) error {
	return s.repo.BulkRestoreItems(ctx, itemType, ids)
}

func (s *archiveService) BulkPermanentDelete(ctx context.Context, itemType string, ids []uuid.UUID) error {
	return s.repo.BulkPermanentDeleteItems(ctx, itemType, ids)
}
