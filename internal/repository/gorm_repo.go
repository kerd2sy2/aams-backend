package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormRoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &gormRoleRepository{db: db}
}

func (r *gormRoleRepository) FindAll(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	if err := r.db.WithContext(ctx).Order("is_system DESC, created_at ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *gormRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).Where("LOWER(name) = LOWER(?)", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *gormRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *gormRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, "id = ?", id).Error
}

func (r *gormRoleRepository) CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Admin{}).Where("role_id = ?", roleID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

type gormAdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &gormAdminRepository{db: db}
}

func (r *gormAdminRepository) FindByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(email) = LOWER(?)", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").First(&admin, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(username) = LOWER(?)", username).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByPhone(ctx context.Context, phone string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("phone = ?", phone).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByLogin(ctx context.Context, login string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where(
		"LOWER(email) = LOWER(?) OR LOWER(username) = LOWER(?) OR phone = ?",
		login, login, login,
	).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) Create(ctx context.Context, admin *domain.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *gormAdminRepository) Update(ctx context.Context, admin *domain.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

func (r *gormAdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Admin{}, "id = ?", id).Error
}

func (r *gormAdminRepository) FindByGoogleEmail(ctx context.Context, email string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(google_email) = LOWER(?)", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("google_id = ?", googleID).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) UpdateGoogleLink(ctx context.Context, adminID uuid.UUID, googleEmail, googleID, googleAvatar string) error {
	return r.db.WithContext(ctx).Model(&domain.Admin{}).Where("id = ?", adminID).Updates(map[string]interface{}{
		"google_email":  googleEmail,
		"google_id":     googleID,
		"google_avatar": googleAvatar,
	}).Error
}

func (r *gormAdminRepository) UnlinkGoogle(ctx context.Context, adminID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Admin{}).Where("id = ?", adminID).Updates(map[string]interface{}{
		"google_email":  "",
		"google_id":     "",
		"google_avatar": "",
	}).Error
}

func (r *gormAdminRepository) FindAll(ctx context.Context) ([]domain.Admin, error) {
	var admins []domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Order("created_at DESC").Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}


// GORM Employee Repository
type gormEmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &gormEmployeeRepository{db: db}
}

func (r *gormEmployeeRepository) Create(ctx context.Context, emp *domain.Employee) error {
	return r.db.WithContext(ctx).Create(emp).Error
}

func (r *gormEmployeeRepository) Update(ctx context.Context, emp *domain.Employee) error {
	return r.db.WithContext(ctx).Save(emp).Error
}

func (r *gormEmployeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Employee{}, "id = ?", id).Error
}

func (r *gormEmployeeRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	var emp domain.Employee
	if err := r.db.WithContext(ctx).Preload("Branch").First(&emp, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *gormEmployeeRepository) FindByNationalID(ctx context.Context, nationalID string) (*domain.Employee, error) {
	var emp domain.Employee
	if err := r.db.WithContext(ctx).First(&emp, "national_id = ?", nationalID).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *gormEmployeeRepository) FindBySearchTerm(ctx context.Context, term string, branchID *uuid.UUID) ([]domain.Employee, error) {
	var employees []domain.Employee
	term = strings.TrimSpace(term)
	if term == "" {
		return employees, nil
	}

	query := r.db.WithContext(ctx).Preload("Branch")
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}

	// Only return couriers/drivers for work sessions (exclude supervisors and management)
	query = query.Where("job_role = ? OR job_role = '' OR job_role IS NULL", "DRIVER")

	// Try UUID parse first
	parsedUUID, err := uuid.Parse(term)
	if err == nil {
		query.Where("id = ?", parsedUUID).Find(&employees)
		if len(employees) > 0 {
			return employees, nil
		}
	}

	likeTerm := "%" + term + "%"
	err = query.Where(
		"name LIKE ? OR employee_number LIKE ? OR national_id LIKE ? OR motorcycle_number LIKE ? OR key_number LIKE ? OR application_id LIKE ?",
		likeTerm, likeTerm, likeTerm, likeTerm, likeTerm, likeTerm,
	).Limit(15).Find(&employees).Error
	return employees, err
}

func (r *gormEmployeeRepository) CountAll(ctx context.Context, branchID *uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Employee{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *gormEmployeeRepository) FindAll(ctx context.Context, filter dto.EmployeeFilter) ([]domain.Employee, int64, error) {
	var employees []domain.Employee
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Employee{}).Preload("Branch")

	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}

	if filter.Search != "" {
		likeTerm := "%" + strings.TrimSpace(filter.Search) + "%"
		parsedUUID, err := uuid.Parse(strings.TrimSpace(filter.Search))
		if err == nil {
			query = query.Where("id = ? OR name LIKE ? OR employee_number LIKE ? OR national_id LIKE ? OR application_id LIKE ?", parsedUUID, likeTerm, likeTerm, likeTerm, likeTerm)
		} else {
			query = query.Where("name LIKE ? OR employee_number LIKE ? OR national_id LIKE ? OR motorcycle_number LIKE ? OR key_number LIKE ? OR application_id LIKE ?", likeTerm, likeTerm, likeTerm, likeTerm, likeTerm, likeTerm)
		}
	}

	if filter.ApplicationID != "" {
		query = query.Where("application_id = ?", filter.ApplicationID)
	}

	if filter.ApplicationType != "" {
		query = query.Where("application_type = ?", filter.ApplicationType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := filter.Order
	if strings.ToLower(order) != "asc" && strings.ToLower(order) != "desc" {
		order = "desc"
	}

	sortBy := filter.SortBy
	allowedSorts := map[string]bool{"created_at": true, "name": true, "national_id": true, "application_id": true, "key_number": true, "motorcycle_number": true, "employee_number": true}
	if !allowedSorts[sortBy] {
		sortBy = "created_at"
	}

	// Use CAST for numeric columns to sort correctly (e.g. 1,2,3... not 1,10,11...)
	offset := (filter.Page - 1) * filter.Limit
	if sortBy == "key_number" || sortBy == "motorcycle_number" {
		err := query.Order(fmt.Sprintf("CAST(NULLIF(%s, '') AS INTEGER) %s", sortBy, strings.ToUpper(order))).Offset(offset).Limit(filter.Limit).Find(&employees).Error
		return employees, total, err
	}
	err := query.Order(fmt.Sprintf("%s %s", sortBy, order)).Offset(offset).Limit(filter.Limit).Find(&employees).Error
	return employees, total, err
}

// GORM Work Repository
type gormWorkRepository struct {
	db *gorm.DB
}

func NewWorkRepository(db *gorm.DB) WorkRepository {
	return &gormWorkRepository{db: db}
}

func (r *gormWorkRepository) CreateSession(ctx context.Context, session *domain.WorkSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *gormWorkRepository) UpdateSession(ctx context.Context, session *domain.WorkSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *gormWorkRepository) FindActiveSessionByEmployeeID(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error) {
	var session domain.WorkSession
	err := r.db.WithContext(ctx).
		Preload("Employee").
		Where("employee_id = ? AND status = ?", empID, domain.StatusActive).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) FindActiveSessionForToday(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var session domain.WorkSession
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND start_time >= ?", empID, startOfDay).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.WorkSession{}).
		Where("employee_id = ? AND start_time >= ?", empID, startOfDay).
		Count(&count).Error
	return count, err
}

func (r *gormWorkRepository) FindLastCompletedSession(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error) {
	var session domain.WorkSession
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND status = ?", empID, domain.StatusCompleted).
		Order("end_time desc").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) FindSessionByID(ctx context.Context, id uuid.UUID) (*domain.WorkSession, error) {
	var session domain.WorkSession
	err := r.db.WithContext(ctx).Preload("Employee").First(&session, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) GetReports(ctx context.Context, filter dto.ReportFilter) ([]domain.WorkSession, int64, error) {
	var sessions []domain.WorkSession
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.WorkSession{}).
		Joins("JOIN employees ON employees.id = work_sessions.employee_id").
		Preload("Employee.Branch")

	if filter.BranchID != nil {
		query = query.Where("employees.branch_id = ?", *filter.BranchID)
	}

	if filter.StartDate != "" {
		startTime, err := time.ParseInLocation("2006-01-02", filter.StartDate, time.Local)
		if err == nil {
			query = query.Where("work_sessions.start_time >= ?", startTime)
		}
	}

	if filter.EndDate != "" {
		endTime, err := time.ParseInLocation("2006-01-02", filter.EndDate, time.Local)
		if err == nil {
			endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.Local)
			query = query.Where("work_sessions.start_time <= ?", endTime)
		}
	}

	if filter.EmployeeID != "" {
		if parsedUUID, err := uuid.Parse(filter.EmployeeID); err == nil {
			query = query.Where("work_sessions.employee_id = ?", parsedUUID)
		} else {
			query = query.Where("CAST(work_sessions.employee_id AS TEXT) LIKE ? OR employees.national_id = ? OR employees.employee_number = ?", "%"+filter.EmployeeID+"%", filter.EmployeeID, filter.EmployeeID)
		}
	}

	if filter.ApplicationID != "" {
		query = query.Where("work_sessions.application_id = ? OR employees.application_id = ?", filter.ApplicationID, filter.ApplicationID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	orderClause := "work_sessions.start_time DESC, work_sessions.created_at DESC"
	err := query.Order(orderClause).Offset(offset).Limit(filter.Limit).Find(&sessions).Error
	return sessions, total, err
}

func (r *gormWorkRepository) GetDashboardStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardResponse, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	resp := &dto.DashboardResponse{}

	// Helper: returns a fresh base query with optional branch join
	baseQuery := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&domain.WorkSession{})
		if branchID != nil {
			q = q.Joins("JOIN employees ON employees.id = work_sessions.employee_id").
				Where("employees.branch_id = ?", *branchID)
		}
		return q
	}

	// Total employees worked today (unique employees)
	baseQuery().Where("start_time >= ?", startOfDay).Distinct("employee_id").Count(&resp.TodayEmployees)

	// Currently active working employees
	baseQuery().Where("status = ?", domain.StatusActive).Count(&resp.WorkingEmployees)

	// Finished employees today
	baseQuery().Where("start_time >= ? AND status = ?", startOfDay, domain.StatusCompleted).Count(&resp.FinishedEmployees)

	// Today's Orders Sum - only count reviewed / approved shifts
	var ordersSum struct{ Total int64 }
	baseQuery().Select("COALESCE(SUM(orders_count), 0) as total").
		Where("start_time >= ? AND is_reviewed = true", startOfDay).
		Scan(&ordersSum)
	resp.TodayOrders = ordersSum.Total

	// Today's Distance Sum - only count reviewed / approved shifts
	var distSum struct{ Total float64 }
	baseQuery().Select("COALESCE(SUM(distance), 0) as total").
		Where("start_time >= ? AND is_reviewed = true", startOfDay).
		Scan(&distSum)
	resp.TodayDistance = distSum.Total

	// Today's Fuel Cost Sum - only count reviewed / approved shifts
	var fuelSum struct{ Total float64 }
	baseQuery().Select("COALESCE(SUM(fuel_cost), 0) as total").
		Where("start_time >= ? AND is_reviewed = true", startOfDay).
		Scan(&fuelSum)
	resp.TodayFuelCost = fuelSum.Total

	// Average Working Hours today
	var completedSessions []domain.WorkSession
	baseQuery().Where("start_time >= ? AND status = ?", startOfDay, domain.StatusCompleted).Find(&completedSessions)
	if len(completedSessions) > 0 {
		var totalHours float64
		for _, s := range completedSessions {
			if s.EndTime != nil {
				totalHours += s.EndTime.Sub(s.StartTime).Hours()
			}
		}
		resp.AvgWorkingHours = totalHours / float64(len(completedSessions))
	}

	// Chart Data - distance/fuel: Last 7 Days, orders: 1st of month to today (only reviewed orders)
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1) // Last day of the month
	monthDayCount := lastOfMonth.Day()            // 28, 29, 30, or 31

	resp.DistanceChart = make([]dto.ChartDataPoint, 0, 7)
	resp.OrdersChart = make([]dto.ChartDataPoint, 0, monthDayCount)
	resp.FuelCostChart = make([]dto.ChartDataPoint, 0, 7)

	type sessionAgg struct {
		StartTime  time.Time `gorm:"column:start_time"`
		Distance   float64   `gorm:"column:distance"`
		Orders     int       `gorm:"column:orders_count"`
		Fuel       float64   `gorm:"column:fuel_cost"`
		IsReviewed bool      `gorm:"column:is_reviewed"`
	}

	var aggSessions []sessionAgg
	chartQ := r.db.WithContext(ctx).
		Model(&domain.WorkSession{}).
		Select("start_time, distance, orders_count, fuel_cost, is_reviewed").
		Where("start_time >= ?", firstOfMonth)

	if branchID != nil {
		chartQ = chartQ.Joins("JOIN employees ON employees.id = work_sessions.employee_id").
			Where("employees.branch_id = ?", *branchID)
	}
	chartQ.Scan(&aggSessions)

	// Aggregate by date in Go
	type dayTotals struct {
		dist float64
		ord  float64
		fuel float64
	}
	dayMap := make(map[string]*dayTotals, 30)
	for _, s := range aggSessions {
		key := s.StartTime.Format("2006-01-02")
		if dayMap[key] == nil {
			dayMap[key] = &dayTotals{}
		}
		if s.IsReviewed {
			dayMap[key].dist += s.Distance
			dayMap[key].ord += float64(s.Orders)
			dayMap[key].fuel += s.Fuel
		}
	}

	// Fill last 7 days for distance & fuel
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

	// Fill full month for orders (day 1 to last day of month)
	dayCount := lastOfMonth.Day()
	for i := 1; i <= dayCount; i++ {
		day := time.Date(now.Year(), now.Month(), i, 0, 0, 0, 0, now.Location())
		dayStr := day.Format("2006-01-02")
		totals := dayMap[dayStr]
		var ord float64
		if totals != nil {
			ord = totals.ord
		}
		resp.OrdersChart = append(resp.OrdersChart, dto.ChartDataPoint{Date: dayStr, Value: ord})
	}

	return resp, nil
}

func (r *gormWorkRepository) FindAllActiveEmployeeIDs(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Raw(
		"SELECT DISTINCT employee_id FROM work_sessions WHERE status = ? AND employee_id IS NOT NULL",
		domain.StatusActive,
	).Scan(&ids).Error
	return ids, err
}

func (r *gormWorkRepository) GetDailyReport(ctx context.Context, branchID *uuid.UUID, date time.Time) ([]dto.DailyEmployeeRow, error) {
	// Saudi Arabia is UTC+3 with no DST. Build day boundaries in Riyadh time so
	// shifts starting just after midnight local time (e.g. 00:30 on the 22nd)
	// are counted on the correct calendar day instead of the previous UTC day.
	riyadh := time.FixedZone("Asia/Riyadh", 3*60*60)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, riyadh)
	endOfDay := startOfDay.Add(24 * time.Hour)

	type row struct {
		EmployeeID   string
		EmployeeName string
		BranchName   string
		KeyNumber    string
		AppType      string `gorm:"column:application_type"`
		SessionCount int
		TotalKM      float64
		TotalOrders  int
		TotalFuel    float64
	}

	query := r.db.WithContext(ctx).
		Table("work_sessions").
		Select("work_sessions.employee_id, employees.name as employee_name, COALESCE(branches.name, '') as branch_name, employees.key_number as key_number, COALESCE(NULLIF(work_sessions.application_type, ''), employees.application_type, '') as application_type, COUNT(*) as session_count, COALESCE(SUM(CASE WHEN work_sessions.is_reviewed = true THEN work_sessions.distance ELSE 0 END), 0) as total_km, COALESCE(SUM(CASE WHEN work_sessions.is_reviewed = true THEN work_sessions.orders_count ELSE 0 END), 0) as total_orders, COALESCE(SUM(CASE WHEN work_sessions.is_reviewed = true THEN work_sessions.fuel_cost ELSE 0 END), 0) as total_fuel").
		Joins("JOIN employees ON employees.id = work_sessions.employee_id").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Where("work_sessions.start_time >= ? AND work_sessions.start_time < ?", startOfDay, endOfDay).
		Group("work_sessions.employee_id, employees.name, branches.name, employees.key_number, COALESCE(NULLIF(work_sessions.application_type, ''), employees.application_type, '')")

	if branchID != nil {
		query = query.Where("employees.branch_id = ?", *branchID)
	}

	query = query.Order("COALESCE((regexp_match(employees.key_number, '^[0-9]+'))[1], '0')::bigint ASC, employees.key_number ASC")

	var rows []row
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	appMap := map[string]string{
		"ninja":  "نينجا",
		"keeta":  "كيتا",
		"hunger": "هنجر",
		"toyou":  "تو يو",
	}

	result := make([]dto.DailyEmployeeRow, 0, len(rows))
	for _, r := range rows {
		appName := appMap[r.AppType]
		if appName == "" {
			appName = r.AppType
		}
		result = append(result, dto.DailyEmployeeRow{
			EmployeeID:    r.EmployeeID,
			EmployeeName:  r.EmployeeName,
			BranchName:    r.BranchName,
			KeyNumber:     r.KeyNumber,
			AppType:       r.AppType,
			AppName:       appName,
			SessionsCount: r.SessionCount,
			TotalKM:       r.TotalKM,
			TotalOrders:   r.TotalOrders,
			TotalFuel:     r.TotalFuel,
		})
	}

	return result, nil
}

// GORM Audit Repository
type gormAuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &gormAuditRepository{db: db}
}

func (r *gormAuditRepository) CreateLog(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormAuditRepository) GetLatestLogs(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *gormAuditRepository) GetLatestLogsByBranch(ctx context.Context, branchID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.WithContext(ctx).
		Where("branch_id = ?", branchID).
		Order("created_at desc").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *gormAuditRepository) FindAll(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AuditLog{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *gormAuditRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.AuditLog{}, "id = ?", id).Error
}

func (r *gormAuditRepository) DeleteBulk(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Delete(&domain.AuditLog{}, "id IN ?", ids).Error
}

func (r *gormAuditRepository) ClearAll(ctx context.Context, branchID *uuid.UUID) error {
	query := r.db.WithContext(ctx)
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	} else {
		query = query.Where("1 = 1")
	}
	return query.Delete(&domain.AuditLog{}).Error
}

// GORM Branch Repository
type gormBranchRepository struct {
	db *gorm.DB
}

func NewBranchRepository(db *gorm.DB) BranchRepository {
	return &gormBranchRepository{db: db}
}

func (r *gormBranchRepository) FindAll(ctx context.Context) ([]domain.Branch, error) {
	var branches []domain.Branch
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&branches).Error; err != nil {
		return nil, err
	}
	return branches, nil
}

func (r *gormBranchRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error) {
	var branch domain.Branch
	if err := r.db.WithContext(ctx).First(&branch, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

func (r *gormBranchRepository) Create(ctx context.Context, branch *domain.Branch) error {
	return r.db.WithContext(ctx).Create(branch).Error
}

func (r *gormBranchRepository) Update(ctx context.Context, branch *domain.Branch) error {
	return r.db.WithContext(ctx).Save(branch).Error
}

func (r *gormBranchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Branch{}, "id = ?", id).Error
}

// GORM Inventory Repository
type gormInventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &gormInventoryRepository{db: db}
}

func (r *gormInventoryRepository) CreateItem(ctx context.Context, item *domain.InventoryItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *gormInventoryRepository) UpdateItem(ctx context.Context, item *domain.InventoryItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *gormInventoryRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.InventoryItem{}, "id = ?", id).Error
}

func (r *gormInventoryRepository) FindItemByID(ctx context.Context, id uuid.UUID) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormInventoryRepository) FindByBarcode(ctx context.Context, barcode string) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	if err := r.db.WithContext(ctx).Where("barcode = ?", barcode).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormInventoryRepository) FindAllItems(ctx context.Context, itemType string) ([]domain.InventoryItem, error) {
	var items []domain.InventoryItem
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if itemType != "" {
		query = query.Where("type = ?", itemType)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// GetStockByBranch calculates per-branch quantity for all items from transactions.
// Returns map[itemID]quantity for the given branch.
func (r *gormInventoryRepository) GetStockByBranch(ctx context.Context, branchID *uuid.UUID) (map[uuid.UUID]int, error) {
	type row struct {
		ItemID   uuid.UUID
		Quantity int
	}
	var rows []row

	query := r.db.WithContext(ctx).
		Model(&domain.InventoryTransaction{}).
		Select("item_id, SUM(CASE WHEN type = 'in' THEN quantity ELSE -quantity END) as quantity")

	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	query = query.Group("item_id")

	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		result[r.ItemID] = r.Quantity
	}
	return result, nil
}

// GetItemStock calculates available quantity for a specific item in a branch
func (r *gormInventoryRepository) GetItemStock(ctx context.Context, itemID uuid.UUID, branchID *uuid.UUID) (int, error) {
	var total int64

	query := r.db.WithContext(ctx).
		Model(&domain.InventoryTransaction{}).
		Select("COALESCE(SUM(CASE WHEN type = 'in' THEN quantity ELSE -quantity END), 0)").
		Where("item_id = ?", itemID)

	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0, err
	}
	return int(total), nil
}

// FindOilItemsWithStock returns oil items with positive stock for a specific branch
func (r *gormInventoryRepository) FindOilItemsWithStock(ctx context.Context, branchID *uuid.UUID) ([]domain.InventoryItem, error) {
	// Get all oil items
	var oilItems []domain.InventoryItem
	if err := r.db.WithContext(ctx).Where("type = ?", "oil").Find(&oilItems).Error; err != nil {
		return nil, err
	}

	// Get stock for this branch
	stock, err := r.GetStockByBranch(ctx, branchID)
	if err != nil {
		return nil, err
	}

	// Filter to only items with positive stock in this branch
	var result []domain.InventoryItem
	for _, item := range oilItems {
		if qty, ok := stock[item.ID]; ok && qty > 0 {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *gormInventoryRepository) CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *gormInventoryRepository) FindTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]domain.InventoryTransaction, int64, error) {
	var txs []domain.InventoryTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.InventoryTransaction{}).Preload("Item").Preload("Employee")
	if itemID != nil {
		query = query.Where("item_id = ?", *itemID)
	}
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&txs).Error
	return txs, total, err
}

func (r *gormInventoryRepository) DeleteAllTransactions(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&domain.InventoryTransaction{}).Error
}

func (r *gormInventoryRepository) CreatePurchaseInvoice(ctx context.Context, invoice *domain.PurchaseInvoice, stockTxs []*domain.InventoryTransaction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(invoice).Error; err != nil {
			return err
		}

		for _, stx := range stockTxs {
			if err := tx.Create(stx).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *gormInventoryRepository) FindPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]domain.PurchaseInvoice, int64, error) {
	var invoices []domain.PurchaseInvoice
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).
		Preload("Branch").
		Preload("Items.Item")

	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("invoice_number LIKE ? OR supplier_name LIKE ? OR notes LIKE ?", s, s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("invoice_date DESC, created_at DESC").Offset(offset).Limit(limit).Find(&invoices).Error
	return invoices, total, err
}

func (r *gormInventoryRepository) FindPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseInvoice, error) {
	var invoice domain.PurchaseInvoice
	if err := r.db.WithContext(ctx).
		Preload("Branch").
		Preload("Items.Item").
		First(&invoice, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *gormInventoryRepository) DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var invoice domain.PurchaseInvoice
		if err := tx.Preload("Items").First(&invoice, "id = ?", id).Error; err != nil {
			return err
		}

		// Delete purchase invoice items
		if err := tx.Where("invoice_id = ?", id).Delete(&domain.PurchaseInvoiceItem{}).Error; err != nil {
			return err
		}

		// Delete any inventory stock-in transactions created by this invoice
		if invoice.InvoiceNumber != "" {
			pattern := "%فاتورة رقم: " + invoice.InvoiceNumber + "%"
			_ = tx.Where("notes LIKE ?", pattern).Delete(&domain.InventoryTransaction{}).Error
		}

		// Delete the invoice itself
		if err := tx.Delete(&domain.PurchaseInvoice{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

// GORM Maintenance Repository
type gormMaintenanceRepository struct {
	db *gorm.DB
}

func NewMaintenanceRepository(db *gorm.DB) MaintenanceRepository {
	return &gormMaintenanceRepository{db: db}
}

func (r *gormMaintenanceRepository) CreateLog(ctx context.Context, log *domain.MaintenanceLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormMaintenanceRepository) FindByEmployeeID(ctx context.Context, empID uuid.UUID, limit int) ([]domain.MaintenanceLog, error) {
	var logs []domain.MaintenanceLog
	err := r.db.WithContext(ctx).
		Preload("Employee").
		Where("employee_id = ?", empID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *gormMaintenanceRepository) FindAll(ctx context.Context, page, limit int) ([]domain.MaintenanceLog, int64, error) {
	var logs []domain.MaintenanceLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.MaintenanceLog{}).Preload("Employee")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *gormMaintenanceRepository) FindLastOilChange(ctx context.Context, empID uuid.UUID) (*domain.MaintenanceLog, error) {
	var log domain.MaintenanceLog
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND type = ?", empID, "oil_change").
		Order("created_at DESC").
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// Investigation Repository
type gormInvestigationRepository struct {
	db *gorm.DB
}

func NewInvestigationRepository(db *gorm.DB) InvestigationRepository {
	return &gormInvestigationRepository{db: db}
}

func (r *gormInvestigationRepository) Create(ctx context.Context, investigation *domain.Investigation) error {
	return r.db.WithContext(ctx).Create(investigation).Error
}

func (r *gormInvestigationRepository) Update(ctx context.Context, investigation *domain.Investigation) error {
	return r.db.WithContext(ctx).Save(investigation).Error
}

func (r *gormInvestigationRepository) FindAll(ctx context.Context) ([]domain.Investigation, error) {
	var investigations []domain.Investigation
	err := r.db.WithContext(ctx).
		Preload("Employee").
		Preload("Supervisor").
		Order("created_at DESC").
		Find(&investigations).Error
	return investigations, err
}

func (r *gormInvestigationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Investigation, error) {
	var investigation domain.Investigation
	err := r.db.WithContext(ctx).
		Preload("Employee").
		Preload("Supervisor").
		Where("id = ?", id).
		First(&investigation).Error
	if err != nil {
		return nil, err
	}
	return &investigation, nil
}

func (r *gormInvestigationRepository) CountPendingApprovals(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Investigation{}).
		Where("type IN ?", []string{"advance", "internet_advance"}).
		Where("status = ? OR status IS NULL OR status = ''", "pending").
		Count(&count).Error
	return count, err
}

// GORM Attendance Repository
type gormAttendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &gormAttendanceRepository{db: db}
}

func (r *gormAttendanceRepository) Upsert(ctx context.Context, attendance *domain.Attendance) error {
	var existing domain.Attendance
	err := r.db.WithContext(ctx).Where("employee_id = ? AND date = ?", attendance.EmployeeID, attendance.Date).First(&existing).Error
	if err != nil {
		// Not found, create new
		return r.db.WithContext(ctx).Create(attendance).Error
	}
	// Update existing
	existing.Status = attendance.Status
	existing.Note = attendance.Note
	return r.db.WithContext(ctx).Save(&existing).Error
}

func (r *gormAttendanceRepository) FindByDate(ctx context.Context, date string) ([]domain.Attendance, error) {
	var records []domain.Attendance
	err := r.db.WithContext(ctx).Preload("Employee").Where("date = ?", date).Find(&records).Error
	return records, err
}

func (r *gormAttendanceRepository) FindByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date string) (*domain.Attendance, error) {
	var record domain.Attendance
	err := r.db.WithContext(ctx).Where("employee_id = ? AND date = ?", employeeID, date).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *gormAttendanceRepository) DeleteByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date string) error {
	return r.db.WithContext(ctx).Where("employee_id = ? AND date = ?", employeeID, date).Delete(&domain.Attendance{}).Error
}

// GORM Custody Repository
type gormCustodyRepository struct {
	db *gorm.DB
}

func NewCustodyRepository(db *gorm.DB) CustodyRepository {
	return &gormCustodyRepository{db: db}
}

func (r *gormCustodyRepository) CreateDay(ctx context.Context, day *domain.CustodyDay) error {
	return r.db.WithContext(ctx).Create(day).Error
}

func (r *gormCustodyRepository) UpdateDay(ctx context.Context, day *domain.CustodyDay) error {
	return r.db.WithContext(ctx).Model(&domain.CustodyDay{}).
		Where("id = ?", day.ID).
		Updates(map[string]interface{}{
			"opening_balance": day.OpeningBalance,
			"added_amount":    day.AddedAmount,
			"closing_balance": day.ClosingBalance,
		}).Error
}

func (r *gormCustodyRepository) FindDayByID(ctx context.Context, id uuid.UUID) (*domain.CustodyDay, error) {
	var day domain.CustodyDay
	err := r.db.WithContext(ctx).
		Preload("Expenses").
		First(&day, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *gormCustodyRepository) FindDayByDate(ctx context.Context, branchID *uuid.UUID, date string) (*domain.CustodyDay, error) {
	var day domain.CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses").Where("date = ?", date)
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.First(&day).Error; err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *gormCustodyRepository) FindLastDay(ctx context.Context, branchID *uuid.UUID) (*domain.CustodyDay, error) {
	var day domain.CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses")
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Order("date DESC").First(&day).Error; err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *gormCustodyRepository) FindAll(ctx context.Context, branchID *uuid.UUID) ([]domain.CustodyDay, error) {
	var days []domain.CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses")
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Order("date DESC").Find(&days).Error; err != nil {
		return nil, err
	}
	return days, nil
}

func (r *gormCustodyRepository) CreateExpense(ctx context.Context, expense *domain.CustodyExpense) error {
	return r.db.WithContext(ctx).Create(expense).Error
}

func (r *gormCustodyRepository) DeleteExpense(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.CustodyExpense{}, "id = ?", id).Error
}

func (r *gormCustodyRepository) FindExpenseByID(ctx context.Context, id uuid.UUID) (*domain.CustodyExpense, error) {
	var expense domain.CustodyExpense
	if err := r.db.WithContext(ctx).First(&expense, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *gormCustodyRepository) CreateLog(ctx context.Context, log *domain.CustodyLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormCustodyRepository) FindLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error) {
	var logs []domain.CustodyLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.CustodyLog{}).Preload("Branch")

	if filter.BranchID != "" {
		if bid, err := uuid.Parse(filter.BranchID); err == nil {
			query = query.Where("branch_id = ?", bid)
		}
	}
	if filter.Date != "" {
		query = query.Where("date = ?", filter.Date)
	}
	if filter.StartDate != "" {
		query = query.Where("date >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("date <= ?", filter.EndDate)
	}
	if filter.ActionType != "" {
		query = query.Where("action_type = ?", filter.ActionType)
	}
	if filter.CreatedBy != "" {
		searchTerm := "%" + filter.CreatedBy + "%"
		query = query.Where("admin_name LIKE ? OR admin_username LIKE ?", searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *gormCustodyRepository) FindLogByID(ctx context.Context, id uuid.UUID) (*domain.CustodyLog, error) {
	var log domain.CustodyLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *gormCustodyRepository) DeleteLog(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.CustodyLog{}, "id = ?", id).Error
}

// GORM Setting Repository
type gormSettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &gormSettingRepository{db: db}
}

func (r *gormSettingRepository) GetAll(ctx context.Context) ([]domain.AppSetting, error) {
	var settings []domain.AppSetting
	if err := r.db.WithContext(ctx).Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *gormSettingRepository) GetByKey(ctx context.Context, key string) (*domain.AppSetting, error) {
	var setting domain.AppSetting
	if err := r.db.WithContext(ctx).Where("\"key\" = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *gormSettingRepository) Create(ctx context.Context, setting *domain.AppSetting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *gormSettingRepository) Update(ctx context.Context, setting *domain.AppSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

func (r *gormSettingRepository) Upsert(ctx context.Context, setting *domain.AppSetting) error {
	existing, err := r.GetByKey(ctx, setting.Key)
	if err == nil && existing != nil {
		existing.Value = setting.Value
		return r.db.WithContext(ctx).Save(existing).Error
	}
	return r.db.WithContext(ctx).Create(setting).Error
}

// ------------------------------------------------------------------
// Vehicle Repository Implementation (الدبابات والمركبات)
// ------------------------------------------------------------------

type gormVehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) VehicleRepository {
	return &gormVehicleRepository{db: db}
}

func (r *gormVehicleRepository) Create(ctx context.Context, vehicle *domain.Vehicle) error {
	return r.db.WithContext(ctx).Create(vehicle).Error
}

func (r *gormVehicleRepository) Update(ctx context.Context, vehicle *domain.Vehicle) error {
	return r.db.WithContext(ctx).Save(vehicle).Error
}

func (r *gormVehicleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Vehicle{}, "id = ?", id).Error
}

func (r *gormVehicleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Vehicle, error) {
	var vehicle domain.Vehicle
	err := r.db.WithContext(ctx).Preload("Branch").First(&vehicle, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *gormVehicleRepository) FindByPlateNumber(ctx context.Context, plateNumber string) (*domain.Vehicle, error) {
	var vehicle domain.Vehicle
	err := r.db.WithContext(ctx).Preload("Branch").First(&vehicle, "plate_number = ?", plateNumber).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

// FindByPlateNumberUnscoped finds a vehicle by plate number including soft-deleted records
func (r *gormVehicleRepository) FindByPlateNumberUnscoped(ctx context.Context, plateNumber string) (*domain.Vehicle, error) {
	var vehicle domain.Vehicle
	err := r.db.WithContext(ctx).Unscoped().Preload("Branch").First(&vehicle, "plate_number = ?", plateNumber).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

// RestoreVehicle restores a soft-deleted vehicle and updates its fields
func (r *gormVehicleRepository) RestoreVehicle(ctx context.Context, vehicle *domain.Vehicle) error {
	return r.db.WithContext(ctx).Unscoped().Model(vehicle).Updates(map[string]interface{}{
		"deleted_at":        nil,
		"brand":             vehicle.Brand,
		"model_year":        vehicle.ModelYear,
		"key_number":        vehicle.KeyNumber,
		"current_km":        vehicle.CurrentKM,
		"last_oil_change_km": vehicle.LastOilChangeKM,
		"status":            vehicle.Status,
		"branch_id":         vehicle.BranchID,
		"notes":             vehicle.Notes,
		"vehicle_type":      vehicle.VehicleType,
	}).Error
}


func (r *gormVehicleRepository) FindAll(ctx context.Context, filter dto.VehicleFilter) ([]domain.Vehicle, int64, error) {
	var vehicles []domain.Vehicle
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Vehicle{}).Preload("Branch")

	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.VehicleType != "" {
		query = query.Where("vehicle_type = ?", filter.VehicleType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("plate_number ILIKE ? OR brand ILIKE ? OR key_number ILIKE ? OR notes ILIKE ?", search, search, search, search)
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

	err := query.Order("plate_number ASC").Offset(offset).Limit(limit).Find(&vehicles).Error
	if err != nil {
		return nil, 0, err
	}

	return vehicles, total, nil
}

func (r *gormVehicleRepository) CountAll(ctx context.Context, branchID *uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Vehicle{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *gormVehicleRepository) UpdateOdometer(ctx context.Context, plateNumber string, newKM float64, distance float64) error {
	var vehicle domain.Vehicle
	err := r.db.WithContext(ctx).First(&vehicle, "plate_number = ?", plateNumber).Error
	if err != nil {
		// Auto-register vehicle if not in DB yet
		newVehicle := domain.Vehicle{
			PlateNumber:     plateNumber,
			VehicleType:     "motorcycle",
			CurrentKM:       newKM,
			LastOilChangeKM: 0,
			TotalDistance:   distance,
			Status:          domain.VehicleStatusAvailable,
		}
		return r.db.WithContext(ctx).Create(&newVehicle).Error
	}

	updates := map[string]interface{}{
		"current_km":     newKM,
		"total_distance": gorm.Expr("total_distance + ?", distance),
		"status":         domain.VehicleStatusAvailable,
		"updated_at":     time.Now(),
	}
	return r.db.WithContext(ctx).Model(&vehicle).Updates(updates).Error
}

func (r *gormVehicleRepository) RecordOilChange(ctx context.Context, id uuid.UUID, currentKM float64) error {
	return r.db.WithContext(ctx).Model(&domain.Vehicle{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_oil_change_km": currentKM,
		"updated_at":         time.Now(),
	}).Error
}

func (r *gormVehicleRepository) FindLatestVehicleKM(ctx context.Context, plateNumber string) (float64, error) {
	// First check vehicle table
	var vehicle domain.Vehicle
	if err := r.db.WithContext(ctx).First(&vehicle, "plate_number = ?", plateNumber).Error; err == nil && vehicle.CurrentKM > 0 {
		return vehicle.CurrentKM, nil
	}

	// Fallback to work_sessions table for this motorcycle number
	var lastEndKM float64
	err := r.db.WithContext(ctx).
		Table("work_sessions").
		Select("end_km").
		Where("motorcycle_number = ? AND status = ?", plateNumber, domain.StatusCompleted).
		Order("end_time DESC, updated_at DESC").
		Limit(1).
		Scan(&lastEndKM).Error
	if err != nil || lastEndKM == 0 {
		return 0, err
	}
	return lastEndKM, nil
}

func (r *gormVehicleRepository) SetVehicleStatus(ctx context.Context, plateNumber string, status string) error {
	return r.db.WithContext(ctx).Model(&domain.Vehicle{}).Where("plate_number = ?", plateNumber).Update("status", status).Error
}

// ------------------------------------------------------------------
// 1. FuelLog Repository Implementation
// ------------------------------------------------------------------
type gormFuelLogRepository struct {
	db *gorm.DB
}

func NewFuelLogRepository(db *gorm.DB) FuelLogRepository {
	return &gormFuelLogRepository{db: db}
}

func (r *gormFuelLogRepository) Create(ctx context.Context, log *domain.FuelLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormFuelLogRepository) Update(ctx context.Context, log *domain.FuelLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *gormFuelLogRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.FuelLog{}, "id = ?", id).Error
}

func (r *gormFuelLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.FuelLog, error) {
	var log domain.FuelLog
	err := r.db.WithContext(ctx).Preload("Employee").Preload("Branch").First(&log, "id = ?", id).Error
	return &log, err
}

func (r *gormFuelLogRepository) FindAll(ctx context.Context, filter dto.FuelLogFilter) ([]domain.FuelLog, int64, error) {
	var logs []domain.FuelLog
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.FuelLog{}).Preload("Employee").Preload("Branch")

	if filter.BranchID != nil {
		q = q.Where("branch_id = ?", filter.BranchID)
	}
	if filter.EmployeeID != nil {
		q = q.Where("employee_id = ?", filter.EmployeeID)
	}
	if filter.Plate != "" {
		q = q.Where("vehicle_plate LIKE ?", "%"+filter.Plate+"%")
	}
	if filter.StartDate != "" {
		q = q.Where("fuel_date >= ?", filter.StartDate+" 00:00:00")
	}
	if filter.EndDate != "" {
		q = q.Where("fuel_date <= ?", filter.EndDate+" 23:59:59")
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Joins("LEFT JOIN employees ON employees.id = fuel_logs.employee_id").
			Where("fuel_logs.vehicle_plate LIKE ? OR employees.name LIKE ? OR fuel_logs.station_name LIKE ?", s, s, s)
	}

	q.Count(&total)

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	err := q.Order("fuel_date DESC, created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *gormFuelLogRepository) GetFuelStats(ctx context.Context, branchID *uuid.UUID, startDate, endDate string) (float64, float64, int64, error) {
	type Result struct {
		TotalCost   float64
		TotalLiters float64
		TotalLogs   int64
	}
	var res Result
	q := r.db.WithContext(ctx).Model(&domain.FuelLog{}).
		Select("COALESCE(SUM(amount), 0) as total_cost, COALESCE(SUM(liters), 0) as total_liters, COUNT(*) as total_logs")

	if branchID != nil {
		q = q.Where("branch_id = ?", branchID)
	}
	if startDate != "" {
		q = q.Where("fuel_date >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		q = q.Where("fuel_date <= ?", endDate+" 23:59:59")
	}

	err := q.Scan(&res).Error
	return res.TotalCost, res.TotalLiters, res.TotalLogs, err
}

// ------------------------------------------------------------------
// 2. TrafficViolation Repository Implementation
// ------------------------------------------------------------------
type gormTrafficViolationRepository struct {
	db *gorm.DB
}

func NewTrafficViolationRepository(db *gorm.DB) TrafficViolationRepository {
	return &gormTrafficViolationRepository{db: db}
}

func (r *gormTrafficViolationRepository) Create(ctx context.Context, v *domain.TrafficViolation) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *gormTrafficViolationRepository) Update(ctx context.Context, v *domain.TrafficViolation) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *gormTrafficViolationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.TrafficViolation{}, "id = ?", id).Error
}

func (r *gormTrafficViolationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.TrafficViolation, error) {
	var v domain.TrafficViolation
	err := r.db.WithContext(ctx).Preload("Employee").Preload("Branch").First(&v, "id = ?", id).Error
	return &v, err
}

func (r *gormTrafficViolationRepository) FindAll(ctx context.Context, filter dto.TrafficViolationFilter) ([]domain.TrafficViolation, int64, error) {
	var list []domain.TrafficViolation
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.TrafficViolation{}).Preload("Employee").Preload("Branch")

	if filter.BranchID != nil {
		q = q.Where("branch_id = ?", filter.BranchID)
	}
	if filter.EmployeeID != nil {
		q = q.Where("employee_id = ?", filter.EmployeeID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.StartDate != "" {
		q = q.Where("violation_date >= ?", filter.StartDate+" 00:00:00")
	}
	if filter.EndDate != "" {
		q = q.Where("violation_date <= ?", filter.EndDate+" 23:59:59")
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Joins("LEFT JOIN employees ON employees.id = traffic_violations.employee_id").
			Where("traffic_violations.violation_number LIKE ? OR traffic_violations.vehicle_plate LIKE ? OR traffic_violations.reason LIKE ? OR employees.name LIKE ?", s, s, s, s)
	}

	q.Count(&total)

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	err := q.Order("violation_date DESC, created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormTrafficViolationRepository) GetViolationStats(ctx context.Context, branchID *uuid.UUID) (float64, float64, int64, error) {
	type Result struct {
		TotalAmount    float64
		DeductedAmount float64
		TotalCount     int64
	}
	var res Result
	q := r.db.WithContext(ctx).Model(&domain.TrafficViolation{}).
		Select("COALESCE(SUM(amount), 0) as total_amount, COALESCE(SUM(CASE WHEN status IN ('DEDUCTED', 'PAID') THEN amount ELSE 0 END), 0) as deducted_amount, COUNT(*) as total_count")

	if branchID != nil {
		q = q.Where("branch_id = ?", branchID)
	}

	err := q.Scan(&res).Error
	return res.TotalAmount, res.DeductedAmount, res.TotalCount, err
}

// ------------------------------------------------------------------
// 3. MaintenanceRequest Repository Implementation
// ------------------------------------------------------------------
type gormMaintenanceRequestRepository struct {
	db *gorm.DB
}

func NewMaintenanceRequestRepository(db *gorm.DB) MaintenanceRequestRepository {
	return &gormMaintenanceRequestRepository{db: db}
}

func (r *gormMaintenanceRequestRepository) Create(ctx context.Context, req *domain.MaintenanceRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *gormMaintenanceRequestRepository) Update(ctx context.Context, req *domain.MaintenanceRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *gormMaintenanceRequestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.MaintenanceRequest{}, "id = ?", id).Error
}

func (r *gormMaintenanceRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.MaintenanceRequest, error) {
	var req domain.MaintenanceRequest
	err := r.db.WithContext(ctx).Preload("Employee").Preload("Branch").First(&req, "id = ?", id).Error
	return &req, err
}

func (r *gormMaintenanceRequestRepository) FindAll(ctx context.Context, filter dto.MaintenanceRequestFilter) ([]domain.MaintenanceRequest, int64, error) {
	var list []domain.MaintenanceRequest
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.MaintenanceRequest{}).Preload("Employee").Preload("Branch")

	if filter.BranchID != nil {
		q = q.Where("branch_id = ?", filter.BranchID)
	}
	if filter.Plate != "" {
		q = q.Where("vehicle_plate LIKE ?", "%"+filter.Plate+"%")
	}
	if filter.Priority != "" {
		q = q.Where("priority = ?", filter.Priority)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Joins("LEFT JOIN employees ON employees.id = maintenance_requests.employee_id").
			Where("maintenance_requests.vehicle_plate LIKE ? OR maintenance_requests.issue_description LIKE ? OR maintenance_requests.workshop_name LIKE ? OR employees.name LIKE ?", s, s, s, s)
	}

	q.Count(&total)

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// ------------------------------------------------------------------
// 4. EmployeeDocument Repository Implementation
// ------------------------------------------------------------------
type gormEmployeeDocumentRepository struct {
	db *gorm.DB
}

func NewEmployeeDocumentRepository(db *gorm.DB) EmployeeDocumentRepository {
	return &gormEmployeeDocumentRepository{db: db}
}

func (r *gormEmployeeDocumentRepository) Create(ctx context.Context, doc *domain.EmployeeDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *gormEmployeeDocumentRepository) Update(ctx context.Context, doc *domain.EmployeeDocument) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *gormEmployeeDocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.EmployeeDocument{}, "id = ?", id).Error
}

func (r *gormEmployeeDocumentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDocument, error) {
	var doc domain.EmployeeDocument
	err := r.db.WithContext(ctx).Preload("Employee").First(&doc, "id = ?", id).Error
	return &doc, err
}

func (r *gormEmployeeDocumentRepository) FindAll(ctx context.Context, filter dto.EmployeeDocumentFilter) ([]domain.EmployeeDocument, int64, error) {
	var list []domain.EmployeeDocument
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.EmployeeDocument{}).Preload("Employee")

	if filter.EmployeeID != nil {
		q = q.Where("employee_id = ?", filter.EmployeeID)
	}
	if filter.DocType != "" {
		q = q.Where("doc_type = ?", filter.DocType)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Joins("LEFT JOIN employees ON employees.id = employee_documents.employee_id").
			Where("employee_documents.title LIKE ? OR employee_documents.doc_number LIKE ? OR employees.name LIKE ?", s, s, s)
	}

	q.Count(&total)

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormEmployeeDocumentRepository) FindExpiringSoon(ctx context.Context, days int) ([]domain.EmployeeDocument, error) {
	var list []domain.EmployeeDocument
	futureDate := time.Now().AddDate(0, 0, days)
	err := r.db.WithContext(ctx).
		Preload("Employee").
		Where("expiry_date IS NOT NULL AND expiry_date <= ? AND expiry_date >= ?", futureDate, time.Now()).
		Order("expiry_date ASC").
		Find(&list).Error
	return list, err
}

// ------------------------------------------------------------------
// 5. EmployeeBankAccount Repository Implementation
// ------------------------------------------------------------------
type gormEmployeeBankAccountRepository struct {
	db *gorm.DB
}

func NewEmployeeBankAccountRepository(db *gorm.DB) EmployeeBankAccountRepository {
	return &gormEmployeeBankAccountRepository{db: db}
}

func (r *gormEmployeeBankAccountRepository) Create(ctx context.Context, acc *domain.EmployeeBankAccount) error {
	return r.db.WithContext(ctx).Create(acc).Error
}

func (r *gormEmployeeBankAccountRepository) Update(ctx context.Context, acc *domain.EmployeeBankAccount) error {
	return r.db.WithContext(ctx).Save(acc).Error
}

func (r *gormEmployeeBankAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.EmployeeBankAccount{}, "id = ?", id).Error
}

func (r *gormEmployeeBankAccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeBankAccount, error) {
	var acc domain.EmployeeBankAccount
	err := r.db.WithContext(ctx).Preload("Employee").First(&acc, "id = ?", id).Error
	return &acc, err
}

func (r *gormEmployeeBankAccountRepository) FindAll(ctx context.Context, filter dto.EmployeeBankAccountFilter) ([]domain.EmployeeBankAccount, int64, error) {
	var list []domain.EmployeeBankAccount
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.EmployeeBankAccount{}).Preload("Employee")

	if filter.EmployeeID != nil {
		q = q.Where("employee_id = ?", filter.EmployeeID)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Joins("LEFT JOIN employees ON employees.id = employee_bank_accounts.employee_id").
			Where("employee_bank_accounts.bank_name LIKE ? OR employee_bank_accounts.iban LIKE ? OR employee_bank_accounts.account_owner_name LIKE ? OR employees.name LIKE ?", s, s, s, s)
	}

	q.Count(&total)

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// ------------------------------------------------------------------
// 6. LeaveRequest Repository Implementation
// ------------------------------------------------------------------
type gormLeaveRequestRepository struct {
	db *gorm.DB
}

func NewLeaveRequestRepository(db *gorm.DB) LeaveRequestRepository {
	return &gormLeaveRequestRepository{db: db}
}

func (r *gormLeaveRequestRepository) Create(ctx context.Context, leave *domain.LeaveRequest) error {
	return r.db.WithContext(ctx).Create(leave).Error
}

func (r *gormLeaveRequestRepository) Update(ctx context.Context, leave *domain.LeaveRequest) error {
	return r.db.WithContext(ctx).Save(leave).Error
}

func (r *gormLeaveRequestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.LeaveRequest{}, "id = ?", id).Error
}

func (r *gormLeaveRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error) {
	var leave domain.LeaveRequest
	err := r.db.WithContext(ctx).Preload("Employee").First(&leave, "id = ?", id).Error
	return &leave, err
}

func (r *gormLeaveRequestRepository) FindAll(ctx context.Context, filter dto.LeaveRequestFilter) ([]domain.LeaveRequest, int64, error) {
	var list []domain.LeaveRequest
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.LeaveRequest{}).Preload("Employee")

	if filter.EmployeeID != nil {
		q = q.Where("employee_id = ?", filter.EmployeeID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.LeaveType != "" {
		q = q.Where("leave_type = ?", filter.LeaveType)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Joins("LEFT JOIN employees ON employees.id = leave_requests.employee_id").
			Where("employees.name LIKE ? OR leave_requests.reason LIKE ?", s, s)
	}

	q.Count(&total)

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	err := q.Order("start_date DESC, created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// ------------------------------------------------------------------
// 7. SupportTicket Repository Implementation
// ------------------------------------------------------------------
type gormSupportTicketRepository struct {
	db *gorm.DB
}

func NewSupportTicketRepository(db *gorm.DB) SupportTicketRepository {
	return &gormSupportTicketRepository{db: db}
}

func (r *gormSupportTicketRepository) Create(ctx context.Context, ticket *domain.SupportTicket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *gormSupportTicketRepository) Update(ctx context.Context, ticket *domain.SupportTicket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *gormSupportTicketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.SupportTicket{}, "id = ?", id).Error
}

func (r *gormSupportTicketRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.SupportTicket, error) {
	var ticket domain.SupportTicket
	err := r.db.WithContext(ctx).Preload("Employee").Preload("Branch").First(&ticket, "id = ?", id).Error
	return &ticket, err
}

func (r *gormSupportTicketRepository) FindAll(ctx context.Context, filter dto.SupportTicketFilter) ([]domain.SupportTicket, int64, error) {
	var list []domain.SupportTicket
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.SupportTicket{}).Preload("Employee").Preload("Branch")

	if filter.BranchID != nil {
		q = q.Where("branch_id = ?", filter.BranchID)
	}
	if filter.EmployeeID != nil {
		q = q.Where("employee_id = ?", filter.EmployeeID)
	}
	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.Priority != "" {
		q = q.Where("priority = ?", filter.Priority)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Joins("LEFT JOIN employees ON employees.id = support_tickets.employee_id").
			Where("support_tickets.ticket_number LIKE ? OR support_tickets.subject LIKE ? OR support_tickets.description LIKE ? OR employees.name LIKE ?", s, s, s, s)
	}

	q.Count(&total)

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}



type gormNotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &gormNotificationRepository{db: db}
}

func (r *gormNotificationRepository) FindUnreadByAdmin(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID) ([]domain.Notification, error) {
	return r.FindAllByAdmin(ctx, adminID, branchID, "unread")
}

func (r *gormNotificationRepository) FindAllByAdmin(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID, status string) ([]domain.Notification, error) {
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

func (r *gormNotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	// For simplicity, we just mark it read in the DB for everyone. In a real system, there would be a pivot table for reads.
	// But since this is a small dashboard, it's fine.
	return r.db.WithContext(ctx).Model(&domain.Notification{}).Where("id = ?", id).Update("status", "read").Error
}

func (r *gormNotificationRepository) MarkAllAsRead(ctx context.Context, adminID uuid.UUID) error {
	// Mark all unread as read. Same caveat as above.
	return r.db.WithContext(ctx).Model(&domain.Notification{}).Where("status = ?", "unread").Update("status", "read").Error
}

func (r *gormNotificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}

func (r *gormNotificationRepository) FindByEmployeeAndTypeAndDate(ctx context.Context, empID uuid.UUID, notifType string, date string) (*domain.Notification, error) {
	var notif domain.Notification
	// check if a notification was already created today (date is YYYY-MM-DD string)
	if err := r.db.WithContext(ctx).Where("employee_id = ? AND type = ? AND DATE(created_at) = ?", empID, notifType, date).First(&notif).Error; err != nil {
		return nil, err
	}
	return &notif, nil
}

// ------------------------------------------------------------------
// 8. gormArchiveRepository (سجل الأرشيف والمحذوفات)
// ------------------------------------------------------------------
type gormArchiveRepository struct {
	db *gorm.DB
}

func NewArchiveRepository(db *gorm.DB) ArchiveRepository {
	return &gormArchiveRepository{db: db}
}

func (r *gormArchiveRepository) GetArchivedItems(ctx context.Context, filter dto.ArchiveFilter) ([]dto.ArchivedItemDTO, int64, dto.ArchiveStatsDTO, error) {
	var items []dto.ArchivedItemDTO
	var stats dto.ArchiveStatsDTO

	// Calculate counts for each category
	r.db.WithContext(ctx).Unscoped().Model(&domain.Employee{}).Where("deleted_at IS NOT NULL").Count(&stats.TotalEmployees)
	r.db.WithContext(ctx).Unscoped().Model(&domain.Vehicle{}).Where("deleted_at IS NOT NULL").Count(&stats.TotalVehicles)
	r.db.WithContext(ctx).Unscoped().Model(&domain.Branch{}).Where("deleted_at IS NOT NULL").Count(&stats.TotalBranches)
	r.db.WithContext(ctx).Unscoped().Model(&domain.EmployeeDocument{}).Where("deleted_at IS NOT NULL").Count(&stats.TotalDocuments)
	r.db.WithContext(ctx).Unscoped().Model(&domain.WorkSession{}).Where("deleted_at IS NOT NULL").Count(&stats.TotalWorkSessions)
	r.db.WithContext(ctx).Unscoped().Model(&domain.LeaveRequest{}).Where("deleted_at IS NOT NULL").Count(&stats.TotalLeaves)
	r.db.WithContext(ctx).Unscoped().Model(&domain.MaintenanceRequest{}).Where("deleted_at IS NOT NULL").Count(&stats.TotalMaintenance)
	r.db.WithContext(ctx).Unscoped().Model(&domain.TrafficViolation{}).Where("deleted_at IS NOT NULL").Count(&stats.TotalViolations)
	r.db.WithContext(ctx).Unscoped().Model(&domain.SupportTicket{}).Where("deleted_at IS NOT NULL").Count(&stats.TotalTickets)
	stats.GrandTotal = stats.TotalEmployees + stats.TotalVehicles + stats.TotalBranches + stats.TotalDocuments + stats.TotalWorkSessions + stats.TotalLeaves + stats.TotalMaintenance + stats.TotalViolations + stats.TotalTickets

	fetchType := strings.ToLower(strings.TrimSpace(filter.Type))
	if fetchType == "" {
		fetchType = "all"
	}
	search := "%" + strings.TrimSpace(filter.Search) + "%"
	hasSearch := strings.TrimSpace(filter.Search) != ""

	// 1. Employees
	if fetchType == "all" || fetchType == "employees" {
		var emps []domain.Employee
		q := r.db.WithContext(ctx).Unscoped().Preload("Branch").Where("deleted_at IS NOT NULL")
		if filter.BranchID != nil {
			q = q.Where("branch_id = ?", *filter.BranchID)
		}
		if hasSearch {
			q = q.Where("name LIKE ? OR national_id LIKE ? OR employee_number LIKE ? OR motorcycle_number LIKE ?", search, search, search, search)
		}
		if err := q.Find(&emps).Error; err == nil {
			for _, e := range emps {
				branchName := ""
				if e.Branch != nil {
					branchName = e.Branch.Name
				}
				subtitle := fmt.Sprintf("هوية: %s | وظيفة: %s", e.NationalID, e.JobRole)
				if e.MotorcycleNumber != "" {
					subtitle += fmt.Sprintf(" | دباب: %s", e.MotorcycleNumber)
				}
				archivedAt := time.Time{}
				if e.DeletedAt.Valid {
					archivedAt = e.DeletedAt.Time
				}
				items = append(items, dto.ArchivedItemDTO{
					ID:         e.ID,
					Type:       "employees",
					TypeName:   "مندوب / موظف",
					Title:      e.Name,
					Subtitle:   subtitle,
					Details:    fmt.Sprintf("جوال: %s | مفتاح: %s", e.EmployeeNumber, e.KeyNumber),
					BranchID:   e.BranchID,
					BranchName: branchName,
					ArchivedAt: archivedAt,
					CreatedAt:  e.CreatedAt,
				})
			}
		}
	}

	// 2. Vehicles
	if fetchType == "all" || fetchType == "vehicles" {
		var vehs []domain.Vehicle
		q := r.db.WithContext(ctx).Unscoped().Preload("Branch").Where("deleted_at IS NOT NULL")
		if filter.BranchID != nil {
			q = q.Where("branch_id = ?", *filter.BranchID)
		}
		if hasSearch {
			q = q.Where("plate_number LIKE ? OR brand LIKE ? OR model_year LIKE ?", search, search, search)
		}
		if err := q.Find(&vehs).Error; err == nil {
			for _, v := range vehs {
				branchName := ""
				if v.Branch != nil {
					branchName = v.Branch.Name
				}
				archivedAt := time.Time{}
				if v.DeletedAt.Valid {
					archivedAt = v.DeletedAt.Time
				}
				items = append(items, dto.ArchivedItemDTO{
					ID:         v.ID,
					Type:       "vehicles",
					TypeName:   "مركبة / دباب",
					Title:      fmt.Sprintf("لوحة: %s", v.PlateNumber),
					Subtitle:   fmt.Sprintf("نوع: %s | موديل: %s %s", v.VehicleType, v.Brand, v.ModelYear),
					Details:    fmt.Sprintf("العداد: %.0f كم | مفتاح: %s", v.CurrentKM, v.KeyNumber),
					BranchID:   v.BranchID,
					BranchName: branchName,
					ArchivedAt: archivedAt,
					CreatedAt:  v.CreatedAt,
				})
			}
		}
	}

	// 3. Branches
	if fetchType == "all" || fetchType == "branches" {
		var branches []domain.Branch
		q := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL")
		if hasSearch {
			q = q.Where("name LIKE ?", search)
		}
		if err := q.Find(&branches).Error; err == nil {
			for _, b := range branches {
				archivedAt := time.Time{}
				if b.DeletedAt.Valid {
					archivedAt = b.DeletedAt.Time
				}
				items = append(items, dto.ArchivedItemDTO{
					ID:         b.ID,
					Type:       "branches",
					TypeName:   "فرع",
					Title:      b.Name,
					Subtitle:   "فرع تشغيلي",
					Details:    "",
					ArchivedAt: archivedAt,
					CreatedAt:  b.CreatedAt,
				})
			}
		}
	}

	// 4. Documents
	if fetchType == "all" || fetchType == "documents" {
		var docs []domain.EmployeeDocument
		q := r.db.WithContext(ctx).Unscoped().Preload("Employee").Where("deleted_at IS NOT NULL")
		if hasSearch {
			q = q.Where("title LIKE ? OR doc_number LIKE ? OR notes LIKE ?", search, search, search)
		}
		if err := q.Find(&docs).Error; err == nil {
			for _, d := range docs {
				empName := "غير محدد"
				if d.Employee != nil {
					empName = d.Employee.Name
				}
				archivedAt := time.Time{}
				if d.DeletedAt.Valid {
					archivedAt = d.DeletedAt.Time
				}
				items = append(items, dto.ArchivedItemDTO{
					ID:         d.ID,
					Type:       "documents",
					TypeName:   "مستند / وثيقة",
					Title:      d.Title,
					Subtitle:   fmt.Sprintf("نوع: %s | المندوب: %s", d.DocType, empName),
					Details:    fmt.Sprintf("رقم الوثيقة: %s | الحالة: %s", d.DocNumber, d.Status),
					ArchivedAt: archivedAt,
					CreatedAt:  d.CreatedAt,
				})
			}
		}
	}

	// 5. Work Sessions
	if fetchType == "all" || fetchType == "work_sessions" {
		var sessions []domain.WorkSession
		q := r.db.WithContext(ctx).Unscoped().Preload("Employee").Where("deleted_at IS NOT NULL")
		if hasSearch {
			q = q.Where("motorcycle_number LIKE ? OR application_id LIKE ?", search, search)
		}
		if err := q.Find(&sessions).Error; err == nil {
			for _, s := range sessions {
				empName := "غير محدد"
				if s.Employee != nil {
					empName = s.Employee.Name
				}
				archivedAt := time.Time{}
				if s.DeletedAt.Valid {
					archivedAt = s.DeletedAt.Time
				}
				items = append(items, dto.ArchivedItemDTO{
					ID:         s.ID,
					Type:       "work_sessions",
					TypeName:   "شفت عمل",
					Title:      fmt.Sprintf("شفت %s (%s)", empName, s.StartTime.Format("2006-01-02")),
					Subtitle:   fmt.Sprintf("طلبات: %d | دباب: %s | عداد: %.0f كم", s.OrdersCount, s.MotorcycleNumber, s.EndKM-s.StartKM),
					Details:    fmt.Sprintf("حالة: %s | تكلفة وقود: %.2f", s.Status, s.FuelCost),
					ArchivedAt: archivedAt,
					CreatedAt:  s.CreatedAt,
				})
			}
		}
	}

	// 6. Leave Requests (طلبات الإجازات)
	if fetchType == "all" || fetchType == "leave_requests" || fetchType == "leaves" {
		var leaves []domain.LeaveRequest
		q := r.db.WithContext(ctx).Unscoped().Preload("Employee").Where("deleted_at IS NOT NULL")
		if hasSearch {
			q = q.Where("leave_type LIKE ? OR reason LIKE ?", search, search)
		}
		if err := q.Find(&leaves).Error; err == nil {
			for _, l := range leaves {
				empName := "غير محدد"
				if l.Employee != nil {
					empName = l.Employee.Name
				}
				archivedAt := time.Time{}
				if l.DeletedAt.Valid {
					archivedAt = l.DeletedAt.Time
				}
				items = append(items, dto.ArchivedItemDTO{
					ID:         l.ID,
					Type:       "leave_requests",
					TypeName:   "طلب إجازة",
					Title:      fmt.Sprintf("طلب إجازة: %s (%s)", empName, l.LeaveType),
					Subtitle:   fmt.Sprintf("الفترة: من %s إلى %s (%d أيام)", l.StartDate, l.EndDate, l.DaysCount),
					Details:    fmt.Sprintf("السبب: %s | الحالة: %s", l.Reason, l.Status),
					ArchivedAt: archivedAt,
					CreatedAt:  l.CreatedAt,
				})
			}
		}
	}

	// 7. Maintenance Requests (طلبات الصيانة)
	if fetchType == "all" || fetchType == "maintenance" || fetchType == "maintenance_requests" {
		var reqs []domain.MaintenanceRequest
		q := r.db.WithContext(ctx).Unscoped().Preload("Employee").Preload("Branch").Where("deleted_at IS NOT NULL")
		if hasSearch {
			q = q.Where("vehicle_plate LIKE ? OR issue_description LIKE ? OR workshop_name LIKE ?", search, search, search)
		}
		if err := q.Find(&reqs).Error; err == nil {
			for _, m := range reqs {
				empName := "غير محدد"
				if m.Employee != nil {
					empName = m.Employee.Name
				}
				branchName := ""
				if m.Branch != nil {
					branchName = m.Branch.Name
				}
				archivedAt := time.Time{}
				if m.DeletedAt.Valid {
					archivedAt = m.DeletedAt.Time
				}
				items = append(items, dto.ArchivedItemDTO{
					ID:         m.ID,
					Type:       "maintenance",
					TypeName:   "طلب صيانة",
					Title:      fmt.Sprintf("صيانة مركبة: %s", m.VehiclePlate),
					Subtitle:   fmt.Sprintf("المندوب: %s | الورشة: %s", empName, m.WorkshopName),
					Details:    fmt.Sprintf("العطل: %s | التكلفة: %.2f", m.IssueDescription, m.ActualCost),
					BranchID:   m.BranchID,
					BranchName: branchName,
					ArchivedAt: archivedAt,
					CreatedAt:  m.CreatedAt,
				})
			}
		}
	}

	// 8. Traffic Violations (المخالفات المرورية)
	if fetchType == "all" || fetchType == "violations" || fetchType == "traffic_violations" {
		var viols []domain.TrafficViolation
		q := r.db.WithContext(ctx).Unscoped().Preload("Employee").Preload("Branch").Where("deleted_at IS NOT NULL")
		if hasSearch {
			q = q.Where("violation_number LIKE ? OR vehicle_plate LIKE ? OR reason LIKE ?", search, search, search)
		}
		if err := q.Find(&viols).Error; err == nil {
			for _, v := range viols {
				empName := "غير محدد"
				if v.Employee != nil {
					empName = v.Employee.Name
				}
				branchName := ""
				if v.Branch != nil {
					branchName = v.Branch.Name
				}
				archivedAt := time.Time{}
				if v.DeletedAt.Valid {
					archivedAt = v.DeletedAt.Time
				}
				items = append(items, dto.ArchivedItemDTO{
					ID:         v.ID,
					Type:       "violations",
					TypeName:   "مخالفة مرورية",
					Title:      fmt.Sprintf("مخالفة #%s (%s)", v.ViolationNumber, v.VehiclePlate),
					Subtitle:   fmt.Sprintf("المندوب: %s | المبلغ: %.2f ريال", empName, v.Amount),
					Details:    fmt.Sprintf("السبب: %s | الحالة: %s", v.Reason, v.Status),
					BranchID:   v.BranchID,
					BranchName: branchName,
					ArchivedAt: archivedAt,
					CreatedAt:  v.CreatedAt,
				})
			}
		}
	}

	// 9. Support Tickets (تذاكر الدعم)
	if fetchType == "all" || fetchType == "tickets" || fetchType == "support_tickets" {
		var tickets []domain.SupportTicket
		q := r.db.WithContext(ctx).Unscoped().Preload("Employee").Preload("Branch").Where("deleted_at IS NOT NULL")
		if hasSearch {
			q = q.Where("ticket_number LIKE ? OR subject LIKE ? OR description LIKE ?", search, search, search)
		}
		if err := q.Find(&tickets).Error; err == nil {
			for _, t := range tickets {
				empName := "غير محدد"
				if t.Employee != nil {
					empName = t.Employee.Name
				}
				branchName := ""
				if t.Branch != nil {
					branchName = t.Branch.Name
				}
				archivedAt := time.Time{}
				if t.DeletedAt.Valid {
					archivedAt = t.DeletedAt.Time
				}
				items = append(items, dto.ArchivedItemDTO{
					ID:         t.ID,
					Type:       "tickets",
					TypeName:   "تذكرة دعم",
					Title:      fmt.Sprintf("تذكرة #%s: %s", t.TicketNumber, t.Subject),
					Subtitle:   fmt.Sprintf("المندوب: %s | الأولوية: %s", empName, t.Priority),
					Details:    fmt.Sprintf("الحالة: %s | الوصف: %s", t.Status, t.Description),
					BranchID:   t.BranchID,
					BranchName: branchName,
					ArchivedAt: archivedAt,
					CreatedAt:  t.CreatedAt,
				})
			}
		}
	}

	// Sort by ArchivedAt DESC
	sort.Slice(items, func(i, j int) bool {
		return items[i].ArchivedAt.After(items[j].ArchivedAt)
	})

	total := int64(len(items))

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 50
	}
	start := (page - 1) * limit
	if start > len(items) {
		start = len(items)
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	paginated := items[start:end]
	return paginated, total, stats, nil
}

func (r *gormArchiveRepository) Restore(ctx context.Context, itemType string, id uuid.UUID) error {
	switch strings.ToLower(itemType) {
	case "employees":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.Employee{}).Where("id = ?", id).Update("deleted_at", nil).Error
	case "vehicles":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.Vehicle{}).Where("id = ?", id).Update("deleted_at", nil).Error
	case "branches":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.Branch{}).Where("id = ?", id).Update("deleted_at", nil).Error
	case "documents":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.EmployeeDocument{}).Where("id = ?", id).Update("deleted_at", nil).Error
	case "work_sessions":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.WorkSession{}).Where("id = ?", id).Update("deleted_at", nil).Error
	case "leave_requests", "leaves":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.LeaveRequest{}).Where("id = ?", id).Update("deleted_at", nil).Error
	case "maintenance", "maintenance_requests":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.MaintenanceRequest{}).Where("id = ?", id).Update("deleted_at", nil).Error
	case "violations", "traffic_violations":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.TrafficViolation{}).Where("id = ?", id).Update("deleted_at", nil).Error
	case "tickets", "support_tickets":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.SupportTicket{}).Where("id = ?", id).Update("deleted_at", nil).Error
	default:
		return fmt.Errorf("نوع العنصر غير مدعوم: %s", itemType)
	}
}

func (r *gormArchiveRepository) PermanentDelete(ctx context.Context, itemType string, id uuid.UUID) error {
	switch strings.ToLower(itemType) {
	case "employees":
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Delete or detach all dependent records first
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.Attendance{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.WorkSession{})
			tx.Unscoped().Where("employee_id = ? OR supervisor_id = ?", id, id).Delete(&domain.Investigation{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.MaintenanceLog{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.MaintenanceRequest{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.TrafficViolation{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.FuelLog{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.EmployeeDocument{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.EmployeeBankAccount{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.LeaveRequest{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.SupportTicket{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.Notification{})
			tx.Unscoped().Where("employee_id = ?", id).Delete(&domain.InventoryTransaction{})
			// Finally delete the employee record
			return tx.Unscoped().Delete(&domain.Employee{}, "id = ?", id).Error
		})
	case "vehicles":
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			tx.Unscoped().Where("vehicle_id = ?", id).Delete(&domain.MaintenanceRequest{})
			tx.Unscoped().Where("vehicle_id = ?", id).Delete(&domain.TrafficViolation{})
			tx.Unscoped().Where("vehicle_id = ?", id).Delete(&domain.FuelLog{})
			tx.Unscoped().Where("vehicle_id = ?", id).Delete(&domain.MaintenanceLog{})
			return tx.Unscoped().Delete(&domain.Vehicle{}, "id = ?", id).Error
		})
	case "branches":
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Detach branch from related records instead of hard deleting them
			tx.Unscoped().Model(&domain.Employee{}).Where("branch_id = ?", id).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.Admin{}).Where("branch_id = ?", id).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.Vehicle{}).Where("branch_id = ?", id).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.SupportTicket{}).Where("branch_id = ?", id).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.MaintenanceRequest{}).Where("branch_id = ?", id).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.CustodyDay{}).Where("branch_id = ?", id).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.AuditLog{}).Where("branch_id = ?", id).Update("branch_id", nil)
			return tx.Unscoped().Delete(&domain.Branch{}, "id = ?", id).Error
		})
	case "documents":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.EmployeeDocument{}, "id = ?", id).Error
	case "work_sessions":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.WorkSession{}, "id = ?", id).Error
	case "leave_requests", "leaves":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.LeaveRequest{}, "id = ?", id).Error
	case "maintenance", "maintenance_requests":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.MaintenanceRequest{}, "id = ?", id).Error
	case "violations", "traffic_violations":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.TrafficViolation{}, "id = ?", id).Error
	case "tickets", "support_tickets":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.SupportTicket{}, "id = ?", id).Error
	default:
		return fmt.Errorf("نوع العنصر غير مدعوم: %s", itemType)
	}
}

func (r *gormArchiveRepository) BulkRestore(ctx context.Context, itemType string, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	switch strings.ToLower(itemType) {
	case "employees":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.Employee{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
	case "vehicles":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.Vehicle{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
	case "branches":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.Branch{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
	case "documents":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.EmployeeDocument{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
	case "work_sessions":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.WorkSession{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
	case "leave_requests", "leaves":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.LeaveRequest{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
	case "maintenance", "maintenance_requests":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.MaintenanceRequest{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
	case "violations", "traffic_violations":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.TrafficViolation{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
	case "tickets", "support_tickets":
		return r.db.WithContext(ctx).Unscoped().Model(&domain.SupportTicket{}).Where("id IN ?", ids).Update("deleted_at", nil).Error
	default:
		return fmt.Errorf("نوع العنصر غير مدعوم: %s", itemType)
	}
}

func (r *gormArchiveRepository) BulkPermanentDelete(ctx context.Context, itemType string, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	switch strings.ToLower(itemType) {
	case "employees":
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.Attendance{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.WorkSession{})
			tx.Unscoped().Where("employee_id IN ? OR supervisor_id IN ?", ids, ids).Delete(&domain.Investigation{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.MaintenanceLog{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.MaintenanceRequest{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.TrafficViolation{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.FuelLog{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.EmployeeDocument{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.EmployeeBankAccount{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.LeaveRequest{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.SupportTicket{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.Notification{})
			tx.Unscoped().Where("employee_id IN ?", ids).Delete(&domain.InventoryTransaction{})
			return tx.Unscoped().Delete(&domain.Employee{}, "id IN ?", ids).Error
		})
	case "vehicles":
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			tx.Unscoped().Where("vehicle_id IN ?", ids).Delete(&domain.MaintenanceRequest{})
			tx.Unscoped().Where("vehicle_id IN ?", ids).Delete(&domain.TrafficViolation{})
			tx.Unscoped().Where("vehicle_id IN ?", ids).Delete(&domain.FuelLog{})
			tx.Unscoped().Where("vehicle_id IN ?", ids).Delete(&domain.MaintenanceLog{})
			return tx.Unscoped().Delete(&domain.Vehicle{}, "id IN ?", ids).Error
		})
	case "branches":
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			tx.Unscoped().Model(&domain.Employee{}).Where("branch_id IN ?", ids).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.Admin{}).Where("branch_id IN ?", ids).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.Vehicle{}).Where("branch_id IN ?", ids).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.SupportTicket{}).Where("branch_id IN ?", ids).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.MaintenanceRequest{}).Where("branch_id IN ?", ids).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.CustodyDay{}).Where("branch_id IN ?", ids).Update("branch_id", nil)
			tx.Unscoped().Model(&domain.AuditLog{}).Where("branch_id IN ?", ids).Update("branch_id", nil)
			return tx.Unscoped().Delete(&domain.Branch{}, "id IN ?", ids).Error
		})
	case "documents":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.EmployeeDocument{}, "id IN ?", ids).Error
	case "work_sessions":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.WorkSession{}, "id IN ?", ids).Error
	case "leave_requests", "leaves":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.LeaveRequest{}, "id IN ?", ids).Error
	case "maintenance", "maintenance_requests":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.MaintenanceRequest{}, "id IN ?", ids).Error
	case "violations", "traffic_violations":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.TrafficViolation{}, "id IN ?", ids).Error
	case "tickets", "support_tickets":
		return r.db.WithContext(ctx).Unscoped().Delete(&domain.SupportTicket{}, "id IN ?", ids).Error
	default:
		return fmt.Errorf("نوع العنصر غير مدعوم: %s", itemType)
	}
}

// ------------------------------------------------------------------
// 9. OTP Repository Implementation
// ------------------------------------------------------------------
type gormOTPRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) OTPRepository {
	return &gormOTPRepository{db: db}
}

func (r *gormOTPRepository) Create(ctx context.Context, otp *domain.OTPRequest) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

func (r *gormOTPRepository) FindActiveByNationalID(ctx context.Context, nationalID string) (*domain.OTPRequest, error) {
	var otp domain.OTPRequest
	err := r.db.WithContext(ctx).
		Preload("Employee").
		Where("national_id = ? AND status = 'PENDING' AND expires_at > ?", nationalID, time.Now()).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormOTPRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.OTPRequest, error) {
	var otp domain.OTPRequest
	err := r.db.WithContext(ctx).
		Preload("Employee").
		First(&otp, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormOTPRepository) FindAll(ctx context.Context, query dto.OTPListQuery) ([]domain.OTPRequest, int64, error) {
	var list []domain.OTPRequest
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.OTPRequest{}).Preload("Employee")

	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}
	if query.Search != "" {
		searchTerm := "%" + query.Search + "%"
		q = q.Where("national_id LIKE ? OR employee_name LIKE ? OR otp_code LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

func (r *gormOTPRepository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.OTPRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "VERIFIED",
			"verified_at": &now,
			"updated_at":  now,
		}).Error
}

func (r *gormOTPRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.OTPRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "CANCELLED",
			"updated_at": time.Now(),
		}).Error
}

func (r *gormOTPRepository) InvalidatePreviousPending(ctx context.Context, nationalID string) error {
	return r.db.WithContext(ctx).Model(&domain.OTPRequest{}).
		Where("national_id = ? AND status = 'PENDING'", nationalID).
		Updates(map[string]interface{}{
			"status":     "EXPIRED",
			"updated_at": time.Now(),
		}).Error
}


