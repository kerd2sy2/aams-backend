package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/system/domain"
	"delivery-backend/internal/modules/system/dto"
)

type SystemRepository interface {
	// Audit
	CreateAuditLog(ctx context.Context, log *domain.AuditLog) error
	FindAuditLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error)
	GetLatestAuditLogs(ctx context.Context, limit int, branchID *uuid.UUID) ([]domain.AuditLog, error)
	DeleteAuditLog(ctx context.Context, id uuid.UUID) error
	BulkDeleteAuditLogs(ctx context.Context, ids []uuid.UUID) error
	ClearAuditLogs(ctx context.Context, branchID *uuid.UUID) error

	// Settings
	GetSettings(ctx context.Context) ([]domain.AppSetting, error)
	GetSettingByKey(ctx context.Context, key string) (*domain.AppSetting, error)
	UpsertSetting(ctx context.Context, setting *domain.AppSetting) error

	// Notifications & Broadcasts
	FindNotificationsByAdmin(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID, status string) ([]domain.Notification, error)
	MarkNotificationAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error
	MarkAllNotificationsAsRead(ctx context.Context, adminID uuid.UUID) error
	CreateNotification(ctx context.Context, notif *domain.Notification) error
	CreateBroadcast(ctx context.Context, broadcast *domain.BroadcastNotification) error
	FindBroadcasts(ctx context.Context, branchID *uuid.UUID, limit, offset int) ([]domain.BroadcastNotification, int64, error)
	FindBroadcastByID(ctx context.Context, id uuid.UUID) (*domain.BroadcastNotification, error)
	DeleteBroadcast(ctx context.Context, id uuid.UUID) error
	FindBroadcastsForEmployee(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time, limit int) ([]dto.BroadcastItemDTO, error)
	FindUnreadBroadcastsForEmployee(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) ([]dto.BroadcastItemDTO, error)
	MarkEmployeeBroadcastRead(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID) error
	MarkAllEmployeeBroadcastsRead(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) error
	SubmitVote(ctx context.Context, vote *domain.BroadcastVote) error
	FindVotesForBroadcast(ctx context.Context, broadcastID uuid.UUID) ([]dto.BroadcastVoteItemDTO, error)
	UpdateEmployeePushToken(ctx context.Context, empID uuid.UUID, token string) error

	// Archive
	GetArchivedItems(ctx context.Context, filter dto.ArchiveFilter) ([]dto.ArchivedItemDTO, int64, dto.ArchiveStatsDTO, error)
	RestoreItem(ctx context.Context, itemType string, id uuid.UUID) error
	PermanentDeleteItem(ctx context.Context, itemType string, id uuid.UUID) error
	BulkRestoreItems(ctx context.Context, itemType string, ids []uuid.UUID) error
	BulkPermanentDeleteItems(ctx context.Context, itemType string, ids []uuid.UUID) error

	// Dashboard & Stats
	GetDashboardStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardStatsResponse, error)
	GetDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) (*dto.DailyReportResponse, error)
	GetReports(ctx context.Context, filter dto.ReportFilter) ([]map[string]interface{}, int64, error)
}

type gormSystemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) SystemRepository {
	return &gormSystemRepository{db: db}
}

// ---------------- Audit Logs ----------------

func (r *gormSystemRepository) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormSystemRepository) FindAuditLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AuditLog{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *gormSystemRepository) GetLatestAuditLogs(ctx context.Context, limit int, branchID *uuid.UUID) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	query := r.db.WithContext(ctx).Model(&domain.AuditLog{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if limit <= 0 {
		limit = 10
	}
	err := query.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *gormSystemRepository) DeleteAuditLog(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.AuditLog{}, "id = ?", id).Error
}

func (r *gormSystemRepository) BulkDeleteAuditLogs(ctx context.Context, ids []uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.AuditLog{}, "id IN ?", ids).Error
}

func (r *gormSystemRepository) ClearAuditLogs(ctx context.Context, branchID *uuid.UUID) error {
	query := r.db.WithContext(ctx)
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	return query.Delete(&domain.AuditLog{}).Error
}

// ---------------- Settings ----------------

func (r *gormSystemRepository) GetSettings(ctx context.Context) ([]domain.AppSetting, error) {
	var settings []domain.AppSetting
	err := r.db.WithContext(ctx).Find(&settings).Error
	return settings, err
}

func (r *gormSystemRepository) GetSettingByKey(ctx context.Context, key string) (*domain.AppSetting, error) {
	var setting domain.AppSetting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *gormSystemRepository) UpsertSetting(ctx context.Context, setting *domain.AppSetting) error {
	var existing domain.AppSetting
	err := r.db.WithContext(ctx).Where("key = ?", setting.Key).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.WithContext(ctx).Create(setting).Error
		}
		return err
	}
	existing.Value = setting.Value
	return r.db.WithContext(ctx).Save(&existing).Error
}

// ---------------- Notifications & Broadcasts ----------------

func (r *gormSystemRepository) FindNotificationsByAdmin(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID, status string) ([]domain.Notification, error) {
	var notifs []domain.Notification
	query := r.db.WithContext(ctx)

	if status == "unread" || status == "read" {
		query = query.Where("status = ?", status)
	}

	if branchID != nil {
		query = query.Where("admin_id = ? OR branch_id = ? OR (admin_id IS NULL AND branch_id IS NULL)", adminID, branchID)
	} else {
		query = query.Where("admin_id = ? OR (admin_id IS NULL AND branch_id IS NULL)", adminID)
	}

	if err := query.Order("created_at DESC").Limit(200).Find(&notifs).Error; err != nil {
		return nil, err
	}
	return notifs, nil
}

func (r *gormSystemRepository) MarkNotificationAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Notification{}).Where("id = ?", id).Update("status", "read").Error
}

func (r *gormSystemRepository) MarkAllNotificationsAsRead(ctx context.Context, adminID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Notification{}).Where("status = ?", "unread").Update("status", "read").Error
}

func (r *gormSystemRepository) CreateNotification(ctx context.Context, notif *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}

func (r *gormSystemRepository) CreateBroadcast(ctx context.Context, broadcast *domain.BroadcastNotification) error {
	return r.db.WithContext(ctx).Create(broadcast).Error
}

func (r *gormSystemRepository) FindBroadcasts(ctx context.Context, branchID *uuid.UUID, limit, offset int) ([]domain.BroadcastNotification, int64, error) {
	var broadcasts []domain.BroadcastNotification
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.BroadcastNotification{})
	if branchID != nil {
		query = query.Where("target = 'ALL' OR branch_id = ?", branchID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 50
	}
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&broadcasts).Error; err != nil {
		return nil, 0, err
	}

	return broadcasts, total, nil
}

func (r *gormSystemRepository) FindBroadcastByID(ctx context.Context, id uuid.UUID) (*domain.BroadcastNotification, error) {
	var b domain.BroadcastNotification
	if err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormSystemRepository) DeleteBroadcast(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_ = tx.Delete(&domain.BroadcastVote{}, "broadcast_id = ?", id).Error
		_ = tx.Delete(&domain.BroadcastRead{}, "broadcast_id = ?", id).Error
		return tx.Delete(&domain.BroadcastNotification{}, "id = ?", id).Error
	})
}

func (r *gormSystemRepository) FindBroadcastsForEmployee(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time, limit int) ([]dto.BroadcastItemDTO, error) {
	if limit <= 0 {
		limit = 50
	}

	query := r.db.WithContext(ctx).Model(&domain.BroadcastNotification{})
	if branchID != nil {
		query = query.Where("target = 'ALL' OR branch_id = ?", branchID)
	} else {
		query = query.Where("target = 'ALL'")
	}

	if !registeredAt.IsZero() {
		query = query.Where("created_at >= ?", registeredAt.Add(-2*time.Minute))
	}

	var broadcasts []domain.BroadcastNotification
	if err := query.Order("created_at DESC").Limit(limit).Find(&broadcasts).Error; err != nil {
		return nil, err
	}

	if len(broadcasts) == 0 {
		return []dto.BroadcastItemDTO{}, nil
	}

	var readIDs []uuid.UUID
	r.db.WithContext(ctx).Model(&domain.BroadcastRead{}).Where("employee_id = ?", empID).Pluck("broadcast_id", &readIDs)
	readMap := make(map[uuid.UUID]bool, len(readIDs))
	for _, id := range readIDs {
		readMap[id] = true
	}

	var votes []domain.BroadcastVote
	r.db.WithContext(ctx).Where("employee_id = ?", empID).Find(&votes)
	voteMap := make(map[uuid.UUID]string, len(votes))
	for _, v := range votes {
		voteMap[v.BroadcastID] = v.Response
	}

	result := make([]dto.BroadcastItemDTO, len(broadcasts))
	for i, b := range broadcasts {
		result[i] = dto.BroadcastItemDTO{
			ID:             b.ID,
			Title:          b.Title,
			Body:           b.Body,
			TitleAr:        b.TitleAr,
			TitleEn:        b.TitleEn,
			TitleBn:        b.TitleBn,
			BodyAr:         b.BodyAr,
			BodyEn:         b.BodyEn,
			BodyBn:         b.BodyBn,
			ImageURL:       b.ImageURL,
			Target:         b.Target,
			BranchID:       b.BranchID,
			CreatedBy:      b.CreatedBy,
			SentCount:      b.SentCount,
			HasPoll:        b.HasPoll,
			PollQuestion:   b.PollQuestion,
			PollQuestionAr: b.PollQuestionAr,
			PollQuestionEn: b.PollQuestionEn,
			PollQuestionBn: b.PollQuestionBn,
			AgreeCount:     b.AgreeCount,
			DisagreeCount:  b.DisagreeCount,
			CreatedAt:      b.CreatedAt,
			IsRead:         readMap[b.ID],
			UserVote:       voteMap[b.ID],
		}
	}

	return result, nil
}

func (r *gormSystemRepository) FindUnreadBroadcastsForEmployee(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) ([]dto.BroadcastItemDTO, error) {
	all, err := r.FindBroadcastsForEmployee(ctx, empID, branchID, registeredAt, 100)
	if err != nil {
		return nil, err
	}

	var unread []dto.BroadcastItemDTO
	for _, b := range all {
		if !b.IsRead {
			unread = append(unread, b)
		}
	}
	return unread, nil
}

func (r *gormSystemRepository) MarkEmployeeBroadcastRead(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID) error {
	var count int64
	r.db.WithContext(ctx).Model(&domain.BroadcastRead{}).Where("employee_id = ? AND broadcast_id = ?", empID, broadcastID).Count(&count)
	if count > 0 {
		return nil
	}

	rec := domain.BroadcastRead{
		EmployeeID:  empID,
		BroadcastID: broadcastID,
		ReadAt:      time.Now(),
	}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *gormSystemRepository) MarkAllEmployeeBroadcastsRead(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) error {
	broadcasts, err := r.FindBroadcastsForEmployee(ctx, empID, branchID, registeredAt, 200)
	if err != nil {
		return err
	}

	for _, b := range broadcasts {
		if !b.IsRead {
			_ = r.MarkEmployeeBroadcastRead(ctx, empID, b.ID)
		}
	}
	return nil
}

func (r *gormSystemRepository) SubmitVote(ctx context.Context, vote *domain.BroadcastVote) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing domain.BroadcastVote
		err := tx.Where("broadcast_id = ? AND employee_id = ?", vote.BroadcastID, vote.EmployeeID).First(&existing).Error
		if err == nil {
			return errors.New("تم التصويت مسبقاً")
		}

		if err := tx.Create(vote).Error; err != nil {
			return err
		}

		switch vote.Response {
		case "AGREE":
			return tx.Model(&domain.BroadcastNotification{}).Where("id = ?", vote.BroadcastID).
				Update("agree_count", gorm.Expr("agree_count + 1")).Error
		case "DISAGREE":
			return tx.Model(&domain.BroadcastNotification{}).Where("id = ?", vote.BroadcastID).
				Update("disagree_count", gorm.Expr("disagree_count + 1")).Error
		}
		return nil
	})
}

func (r *gormSystemRepository) FindVotesForBroadcast(ctx context.Context, broadcastID uuid.UUID) ([]dto.BroadcastVoteItemDTO, error) {
	var votes []struct {
		ID             uuid.UUID `json:"id"`
		EmployeeID     uuid.UUID `json:"employee_id"`
		EmployeeName   string    `json:"employee_name"`
		EmployeeNumber string    `json:"employee_number"`
		NationalID     string    `json:"national_id"`
		Phone          string    `json:"phone"`
		Response       string    `json:"response"`
		Reason         string    `json:"reason"`
		CreatedAt      time.Time `json:"created_at"`
	}

	err := r.db.WithContext(ctx).
		Table("broadcast_votes").
		Select("broadcast_votes.id, broadcast_votes.employee_id, employees.name as employee_name, employees.employee_number, employees.national_id, employees.phone, broadcast_votes.response, broadcast_votes.reason, broadcast_votes.created_at").
		Joins("LEFT JOIN employees ON employees.id = broadcast_votes.employee_id").
		Where("broadcast_votes.broadcast_id = ?", broadcastID).
		Order("broadcast_votes.created_at DESC").
		Scan(&votes).Error

	if err != nil {
		return nil, err
	}

	result := make([]dto.BroadcastVoteItemDTO, len(votes))
	for i, v := range votes {
		result[i] = dto.BroadcastVoteItemDTO{
			ID:             v.ID,
			EmployeeID:     v.EmployeeID,
			EmployeeName:   v.EmployeeName,
			EmployeeNumber: v.EmployeeNumber,
			NationalID:     v.NationalID,
			Phone:          v.Phone,
			Response:       v.Response,
			Reason:         v.Reason,
			CreatedAt:      v.CreatedAt,
		}
	}
	return result, nil
}

func (r *gormSystemRepository) UpdateEmployeePushToken(ctx context.Context, empID uuid.UUID, token string) error {
	return r.db.WithContext(ctx).Table("employees").Where("id = ?", empID).Update("push_token", token).Error
}

// ---------------- Archive & Trash ----------------

func (r *gormSystemRepository) GetArchivedItems(ctx context.Context, filter dto.ArchiveFilter) ([]dto.ArchivedItemDTO, int64, dto.ArchiveStatsDTO, error) {
	var items []dto.ArchivedItemDTO
	var total int64
	var stats dto.ArchiveStatsDTO

	// Stats counts
	r.db.WithContext(ctx).Table("employees").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalEmployees)
	r.db.WithContext(ctx).Table("vehicles").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalVehicles)
	r.db.WithContext(ctx).Table("branches").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalBranches)
	r.db.WithContext(ctx).Table("employee_documents").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalDocuments)
	r.db.WithContext(ctx).Table("work_sessions").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalWorkSessions)
	r.db.WithContext(ctx).Table("leave_requests").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalLeaves)
	r.db.WithContext(ctx).Table("maintenance_requests").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalMaintenance)
	r.db.WithContext(ctx).Table("violations").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalViolations)
	r.db.WithContext(ctx).Table("support_tickets").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalTickets)
	stats.GrandTotal = stats.TotalEmployees + stats.TotalVehicles + stats.TotalBranches + stats.TotalDocuments + stats.TotalWorkSessions + stats.TotalLeaves + stats.TotalMaintenance + stats.TotalViolations + stats.TotalTickets

	// Query items based on filter.Type
	typeQuery := filter.Type
	if typeQuery == "" || typeQuery == "all" || typeQuery == "employees" {
		var emps []struct {
			ID        uuid.UUID `gorm:"column:id"`
			Name      string    `gorm:"column:name"`
			NationalID string   `gorm:"column:national_id"`
			Phone     string    `gorm:"column:phone"`
			DeletedAt time.Time `gorm:"column:deleted_at"`
			CreatedAt time.Time `gorm:"column:created_at"`
		}
		q := r.db.WithContext(ctx).Table("employees").Unscoped().Where("deleted_at IS NOT NULL")
		if filter.Search != "" {
			s := "%" + filter.Search + "%"
			q = q.Where("name ILIKE ? OR national_id ILIKE ? OR phone ILIKE ?", s, s, s)
		}
		_ = q.Order("deleted_at DESC").Limit(50).Scan(&emps).Error
		for _, e := range emps {
			items = append(items, dto.ArchivedItemDTO{
				ID:         e.ID,
				Type:       "employees",
				TypeName:   "مندوب",
				Title:      e.Name,
				Subtitle:   e.NationalID,
				Details:    e.Phone,
				ArchivedAt: e.DeletedAt,
				CreatedAt:  e.CreatedAt,
			})
		}
	}

	total = int64(len(items))
	return items, total, stats, nil
}

func (r *gormSystemRepository) RestoreItem(ctx context.Context, itemType string, id uuid.UUID) error {
	table := r.getTableName(itemType)
	if table == "" {
		return errors.New("نوع العنصر غير صالح")
	}
	return r.db.WithContext(ctx).Table(table).Unscoped().Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *gormSystemRepository) PermanentDeleteItem(ctx context.Context, itemType string, id uuid.UUID) error {
	table := r.getTableName(itemType)
	if table == "" {
		return errors.New("نوع العنصر غير صالح")
	}
	return r.db.WithContext(ctx).Table(table).Unscoped().Where("id = ?", id).Delete(nil).Error
}

func (r *gormSystemRepository) BulkRestoreItems(ctx context.Context, itemType string, ids []uuid.UUID) error {
	table := r.getTableName(itemType)
	if table == "" {
		return errors.New("نوع العنصر غير صالح")
	}
	return r.db.WithContext(ctx).Table(table).Unscoped().Where("id IN ?", ids).Update("deleted_at", nil).Error
}

func (r *gormSystemRepository) BulkPermanentDeleteItems(ctx context.Context, itemType string, ids []uuid.UUID) error {
	table := r.getTableName(itemType)
	if table == "" {
		return errors.New("نوع العنصر غير صالح")
	}
	return r.db.WithContext(ctx).Table(table).Unscoped().Where("id IN ?", ids).Delete(nil).Error
}

func (r *gormSystemRepository) getTableName(itemType string) string {
	switch itemType {
	case "employees":
		return "employees"
	case "vehicles":
		return "vehicles"
	case "branches":
		return "branches"
	case "documents":
		return "employee_documents"
	case "work_sessions":
		return "work_sessions"
	case "leaves":
		return "leave_requests"
	case "maintenance":
		return "maintenance_requests"
	case "violations":
		return "violations"
	case "tickets":
		return "support_tickets"
	default:
		return ""
	}
}

// ---------------- Dashboard & Stats ----------------

func (r *gormSystemRepository) GetDashboardStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardStatsResponse, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	resp := &dto.DashboardStatsResponse{
		DistanceChart:    make([]dto.ChartDataPoint, 0),
		OrdersChart:      make([]dto.ChartDataPoint, 0),
		FuelCostChart:    make([]dto.ChartDataPoint, 0),
		LatestActivities: make([]dto.AuditLogResponse, 0),
	}

	// 1. Employee counts
	empQuery := r.db.WithContext(ctx).Table("employees").Where("deleted_at IS NULL")
	if branchID != nil {
		empQuery = empQuery.Where("branch_id = ?", *branchID)
	}
	empQuery.Count(&resp.TotalEmployees)
	empQuery.Where("is_working = ?", true).Count(&resp.ActiveEmployees)

	// 2. Vehicle counts
	vehQuery := r.db.WithContext(ctx).Table("vehicles").Where("deleted_at IS NULL")
	if branchID != nil {
		vehQuery = vehQuery.Where("branch_id = ?", *branchID)
	}
	vehQuery.Count(&resp.TotalMotorcycles)
	vehQuery.Where("status = ?", "WORKING").Count(&resp.ActiveMotorcycles)

	// 3. Work sessions base query helper
	baseQuery := func() *gorm.DB {
		q := r.db.WithContext(ctx).Table("work_sessions")
		if branchID != nil {
			q = q.Joins("JOIN employees ON employees.id = work_sessions.employee_id").
				Where("employees.branch_id = ?", *branchID)
		}
		return q
	}

	// Today unique employees worked
	baseQuery().Where("work_sessions.start_time >= ?", startOfDay).Distinct("work_sessions.employee_id").Count(&resp.TodayEmployees)

	// Currently active working sessions
	baseQuery().Where("work_sessions.status = ?", "ACTIVE").Count(&resp.WorkingEmployees)

	// Finished shifts today
	baseQuery().Where("work_sessions.start_time >= ? AND work_sessions.status = ?", startOfDay, "COMPLETED").Count(&resp.FinishedEmployees)

	// Today's Orders Sum (only reviewed sessions approved by supervisor)
	var ordersSum struct{ Total int64 }
	baseQuery().Select("COALESCE(SUM(work_sessions.orders_count), 0) as total").
		Where("work_sessions.start_time >= ? AND work_sessions.is_reviewed = ?", startOfDay, true).
		Scan(&ordersSum)
	resp.TodayOrders = ordersSum.Total

	// Today's Distance Sum
	var distSum struct{ Total float64 }
	baseQuery().Select("COALESCE(SUM(work_sessions.distance), 0) as total").
		Where("work_sessions.start_time >= ?", startOfDay).
		Scan(&distSum)
	resp.TodayDistance = distSum.Total

	// Today's Fuel Cost Sum
	var fuelSum struct{ Total float64 }
	baseQuery().Select("COALESCE(SUM(work_sessions.fuel_cost), 0) as total").
		Where("work_sessions.start_time >= ?", startOfDay).
		Scan(&fuelSum)
	resp.TodayFuelCost = fuelSum.Total

	// Average Working Hours today
	var completedRows []struct {
		StartTime time.Time  `gorm:"column:start_time"`
		EndTime   *time.Time `gorm:"column:end_time"`
	}
	baseQuery().Select("work_sessions.start_time, work_sessions.end_time").
		Where("work_sessions.start_time >= ? AND work_sessions.status = ?", startOfDay, "COMPLETED").
		Scan(&completedRows)
	if len(completedRows) > 0 {
		var totalHours float64
		for _, s := range completedRows {
			if s.EndTime != nil {
				totalHours += s.EndTime.Sub(s.StartTime).Hours()
			}
		}
		resp.AvgWorkingHours = totalHours / float64(len(completedRows))
	}

	// 4. Chart Data
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	type sessionAgg struct {
		StartTime  time.Time `gorm:"column:start_time"`
		Distance   float64   `gorm:"column:distance"`
		Orders     int       `gorm:"column:orders_count"`
		Fuel       float64   `gorm:"column:fuel_cost"`
		IsReviewed bool      `gorm:"column:is_reviewed"`
	}
	var aggSessions []sessionAgg
	chartQ := r.db.WithContext(ctx).Table("work_sessions").
		Select("work_sessions.start_time, work_sessions.distance, work_sessions.orders_count, work_sessions.fuel_cost, work_sessions.is_reviewed").
		Where("work_sessions.start_time >= ?", firstOfMonth)
	if branchID != nil {
		chartQ = chartQ.Joins("JOIN employees ON employees.id = work_sessions.employee_id").
			Where("employees.branch_id = ?", *branchID)
	}
	chartQ.Scan(&aggSessions)

	type dayTotals struct {
		dist float64
		ord  float64
		fuel float64
	}
	dayMap := make(map[string]*dayTotals, 31)
	for _, s := range aggSessions {
		key := s.StartTime.Format("2006-01-02")
		if dayMap[key] == nil {
			dayMap[key] = &dayTotals{}
		}
		dayMap[key].dist += s.Distance
		if s.IsReviewed {
			dayMap[key].ord += float64(s.Orders)
		}
		dayMap[key].fuel += s.Fuel
	}

	// Last 7 days for distance & fuel
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dayStr := day.Format("2006-01-02")
		totals := dayMap[dayStr]
		var dist, fuel float64
		if totals != nil {
			dist = totals.dist
			fuel = totals.fuel
		}
		resp.DistanceChart = append(resp.DistanceChart, dto.ChartDataPoint{Date: dayStr, Value: dist})
		resp.FuelCostChart = append(resp.FuelCostChart, dto.ChartDataPoint{Date: dayStr, Value: fuel})
	}

	// Full month for orders
	monthDays := lastOfMonth.Day()
	for i := 1; i <= monthDays; i++ {
		day := time.Date(now.Year(), now.Month(), i, 0, 0, 0, 0, now.Location())
		dayStr := day.Format("2006-01-02")
		totals := dayMap[dayStr]
		var ord float64
		if totals != nil {
			ord = totals.ord
		}
		resp.OrdersChart = append(resp.OrdersChart, dto.ChartDataPoint{Date: dayStr, Value: ord})
	}

	// 5. Latest activities
	latestLogs, _ := r.GetLatestAuditLogs(ctx, 10, branchID)
	for _, l := range latestLogs {
		resp.LatestActivities = append(resp.LatestActivities, dto.AuditLogResponse{
			ID:        l.ID,
			AdminName: l.AdminName,
			Action:    l.Action,
			Details:   l.Details,
			IPAddress: l.IPAddress,
			CreatedAt: l.CreatedAt,
		})
	}

	return resp, nil
}

func (r *gormSystemRepository) GetDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) (*dto.DailyReportResponse, error) {
	resp := &dto.DailyReportResponse{
		Date:         dateStr,
		AppSummaries: make([]dto.DailyAppSummary, 0),
		Employees:    make([]dto.DailyEmployeeRow, 0),
	}

	query := r.db.WithContext(ctx).Table("work_sessions").
		Select("work_sessions.*, employees.name as employee_name, employees.key_number, branches.name as branch_name").
		Joins("LEFT JOIN employees ON employees.id = work_sessions.employee_id").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Where("DATE(work_sessions.start_time) = ?", dateStr)

	if branchID != nil {
		query = query.Where("employees.branch_id = ?", *branchID)
	}

	var rows []struct {
		EmployeeID       string    `gorm:"column:employee_id"`
		EmployeeName     string    `gorm:"column:employee_name"`
		BranchName       string    `gorm:"column:branch_name"`
		KeyNumber        string    `gorm:"column:key_number"`
		ApplicationID    string    `gorm:"column:application_id"`
		ApplicationType  string    `gorm:"column:application_type"`
		MotorcycleNumber string    `gorm:"column:motorcycle_number"`
		OrdersCount      int       `gorm:"column:orders_count"`
		Distance         float64   `gorm:"column:distance"`
		FuelCost         float64   `gorm:"column:fuel_cost"`
		StartTime        time.Time `gorm:"column:start_time"`
		EndTime          *time.Time `gorm:"column:end_time"`
		Status           string    `gorm:"column:status"`
		ReviewNotes      string    `gorm:"column:review_notes"`
		IsReviewed       bool      `gorm:"column:is_reviewed"`
	}

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		resp.TotalOrders += row.OrdersCount
		resp.TotalKM += row.Distance
		resp.TotalFuel += row.FuelCost
		resp.TotalSessions++

		dur := "0 ساعة"
		if row.EndTime != nil {
			d := row.EndTime.Sub(row.StartTime)
			dur = fmt.Sprintf("%.1f ساعة", d.Hours())
		}

		resp.Employees = append(resp.Employees, dto.DailyEmployeeRow{
			EmployeeID:       row.EmployeeID,
			EmployeeName:     row.EmployeeName,
			BranchName:       row.BranchName,
			KeyNumber:        row.KeyNumber,
			AppType:          row.ApplicationType,
			AppName:          row.ApplicationID,
			MotorcycleNumber: row.MotorcycleNumber,
			OrdersCount:      row.OrdersCount,
			Distance:         row.Distance,
			FuelCost:         row.FuelCost,
			WorkingDuration:  dur,
			Status:           row.Status,
			ReviewNotes:      row.ReviewNotes,
			IsReviewed:       row.IsReviewed,
		})
	}

	return resp, nil
}

func (r *gormSystemRepository) GetReports(ctx context.Context, filter dto.ReportFilter) ([]map[string]interface{}, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Table("work_sessions").
		Select("work_sessions.*, employees.name as employee_name, employees.national_id, employees.key_number, branches.name as branch_name").
		Joins("LEFT JOIN employees ON employees.id = work_sessions.employee_id").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id")

	if filter.BranchID != nil {
		query = query.Where("employees.branch_id = ?", *filter.BranchID)
	}
	if filter.StartDate != "" {
		query = query.Where("work_sessions.start_time >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("work_sessions.start_time <= ?", filter.EndDate+" 23:59:59")
	}
	if filter.EmployeeID != "" {
		query = query.Where("work_sessions.employee_id = ?", filter.EmployeeID)
	}
	if filter.IsReviewed != nil {
		query = query.Where("work_sessions.is_reviewed = ?", *filter.IsReviewed)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var results []map[string]interface{}
	err := query.Order("work_sessions.start_time DESC").Offset(offset).Limit(limit).Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range results {
		results[i]["working_duration"] = "—"
		startTimeVal := results[i]["start_time"]
		endTimeVal := results[i]["end_time"]
		if startTimeVal != nil {
			var startTime time.Time
			switch t := startTimeVal.(type) {
			case time.Time:
				startTime = t
			case string:
				startTime, _ = time.Parse(time.RFC3339, t)
			}
			if !startTime.IsZero() {
				if endTimeVal != nil {
					var endTime time.Time
					switch t := endTimeVal.(type) {
					case time.Time:
						endTime = t
					case string:
						endTime, _ = time.Parse(time.RFC3339, t)
					}
					if !endTime.IsZero() {
						d := endTime.Sub(startTime)
						hours := int(d.Hours())
						minutes := int(d.Minutes()) % 60
						if hours > 0 && minutes > 0 {
							results[i]["working_duration"] = fmt.Sprintf("%d س و %d د", hours, minutes)
						} else if hours > 0 {
							results[i]["working_duration"] = fmt.Sprintf("%d س", hours)
						} else {
							results[i]["working_duration"] = fmt.Sprintf("%d د", minutes)
						}
					}
				} else {
					results[i]["working_duration"] = "قائم الآن"
				}
			}
		}
	}

	return results, total, nil
}
