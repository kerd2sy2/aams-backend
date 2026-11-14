package system

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cachedDashboardEntry struct {
	data      *DashboardStatsResponse
	expiresAt time.Time
}

var (
	dashboardCacheMu sync.RWMutex
	dashboardCache   = make(map[string]cachedDashboardEntry)
)

// getOperationalWindow returns the start and end of the current operational business day.
// The operational day cut-off is 10:00 AM (Asia/Riyadh timezone).
func getOperationalWindow(t time.Time) (time.Time, time.Time) {
	riyadh := time.FixedZone("Asia/Riyadh", 3*60*60)
	now := t.In(riyadh)
	cutoffHour := 10

	var start time.Time
	var end time.Time

	if now.Hour() < cutoffHour {
		yesterday := now.AddDate(0, 0, -1)
		start = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, riyadh)
		end = time.Date(now.Year(), now.Month(), now.Day(), cutoffHour, 0, 0, 0, riyadh)
	} else {
		tomorrow := now.AddDate(0, 0, 1)
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, riyadh)
		end = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), cutoffHour, 0, 0, 0, riyadh)
	}
	return start, end
}

func (r *gormSystemRepository) GetDashboardStats(ctx context.Context, branchID *uuid.UUID) (*DashboardStatsResponse, error) {
	cacheKey := "global"
	if branchID != nil {
		cacheKey = branchID.String()
	}
	dashboardCacheMu.RLock()
	if entry, ok := dashboardCache[cacheKey]; ok && time.Now().Before(entry.expiresAt) && entry.data != nil {
		dashboardCacheMu.RUnlock()
		return entry.data, nil
	}
	dashboardCacheMu.RUnlock()

	now := time.Now()
	startOfDay, endOfDay := getOperationalWindow(now)

	resp := &DashboardStatsResponse{
		DistanceChart:    make([]ChartDataPoint, 0),
		OrdersChart:      make([]ChartDataPoint, 0),
		FuelCostChart:    make([]ChartDataPoint, 0),
		LatestActivities: make([]AuditLogResponse, 0),
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

	baseQuery().Where("work_sessions.status = ?", "ACTIVE").Distinct("work_sessions.employee_id").Count(&resp.WorkingEmployees)

	baseQuery().Where("(work_sessions.start_time >= ? AND work_sessions.start_time < ?) OR work_sessions.status = ?", startOfDay, endOfDay, "ACTIVE").
		Distinct("work_sessions.employee_id").Count(&resp.TodayEmployees)

	baseQuery().
		Where("work_sessions.start_time >= ? AND work_sessions.start_time < ? AND work_sessions.status = ?", startOfDay, endOfDay, "COMPLETED").
		Where("work_sessions.employee_id NOT IN (SELECT ws2.employee_id FROM work_sessions ws2 WHERE ws2.status = 'ACTIVE' AND ws2.employee_id IS NOT NULL)").
		Distinct("work_sessions.employee_id").
		Count(&resp.FinishedEmployees)

	if resp.TodayEmployees < resp.WorkingEmployees+resp.FinishedEmployees {
		resp.TodayEmployees = resp.WorkingEmployees + resp.FinishedEmployees
	}

	var daySums struct {
		Orders   int64   `gorm:"column:orders"`
		Distance float64 `gorm:"column:distance"`
		Fuel     float64 `gorm:"column:fuel"`
	}
	baseQuery().Select("COALESCE(SUM(work_sessions.orders_count), 0) as orders, COALESCE(SUM(work_sessions.distance), 0) as distance, COALESCE(SUM(work_sessions.fuel_cost), 0) as fuel").
		Where("work_sessions.start_time >= ? AND work_sessions.start_time < ? AND work_sessions.is_reviewed = ?", startOfDay, endOfDay, true).
		Scan(&daySums)
	resp.TodayOrders = daySums.Orders
	resp.TodayDistance = daySums.Distance
	resp.TodayFuelCost = daySums.Fuel

	var completedRows []struct {
		StartTime time.Time  `gorm:"column:start_time"`
		EndTime   *time.Time `gorm:"column:end_time"`
	}
	baseQuery().Select("work_sessions.start_time, work_sessions.end_time").
		Where("work_sessions.start_time >= ? AND work_sessions.start_time < ? AND work_sessions.status = ?", startOfDay, endOfDay, "COMPLETED").
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
		if s.IsReviewed {
			dayMap[key].dist += s.Distance
			dayMap[key].ord += float64(s.Orders)
			dayMap[key].fuel += s.Fuel
		}
	}

	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dayStr := day.Format("2006-01-02")
		totals := dayMap[dayStr]
		var dist, fuel float64
		if totals != nil {
			dist = totals.dist
			fuel = totals.fuel
		}
		resp.DistanceChart = append(resp.DistanceChart, ChartDataPoint{Date: dayStr, Value: dist})
		resp.FuelCostChart = append(resp.FuelCostChart, ChartDataPoint{Date: dayStr, Value: fuel})
	}

	monthDays := lastOfMonth.Day()
	for i := 1; i <= monthDays; i++ {
		day := time.Date(now.Year(), now.Month(), i, 0, 0, 0, 0, now.Location())
		dayStr := day.Format("2006-01-02")
		totals := dayMap[dayStr]
		var ord float64
		if totals != nil {
			ord = totals.ord
		}
		resp.OrdersChart = append(resp.OrdersChart, ChartDataPoint{Date: dayStr, Value: ord})
	}

	// 5. Latest activities
	latestLogs, _ := r.GetLatestAuditLogs(ctx, 10, branchID)
	for _, l := range latestLogs {
		resp.LatestActivities = append(resp.LatestActivities, AuditLogResponse{
			ID:        l.ID,
			AdminName: l.AdminName,
			Action:    l.Action,
			Details:   l.Details,
			IPAddress: l.IPAddress,
			CreatedAt: l.CreatedAt,
		})
	}

	dashboardCacheMu.Lock()
	dashboardCache[cacheKey] = cachedDashboardEntry{
		data:      resp,
		expiresAt: time.Now().Add(5 * time.Second),
	}
	dashboardCacheMu.Unlock()

	return resp, nil
}

func (r *gormSystemRepository) GetDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) (*DailyReportResponse, error) {
	resp := &DailyReportResponse{
		Date:         dateStr,
		AppSummaries: make([]DailyAppSummary, 0),
		Employees:    make([]DailyEmployeeRow, 0),
		Rows:         make([]DailyEmployeeRow, 0),
	}

	query := r.db.WithContext(ctx).Table("work_sessions").
		Select("work_sessions.*, employees.name as employee_name, employees.key_number, branches.name as branch_name").
		Joins("LEFT JOIN employees ON employees.id = work_sessions.employee_id").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id")

	riyadh := time.FixedZone("Asia/Riyadh", 3*60*60)
	if t, err := time.ParseInLocation("2006-01-02", dateStr, riyadh); err == nil {
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, riyadh)
		endOfDay := startOfDay.Add(24 * time.Hour)
		query = query.Where("work_sessions.start_time >= ? AND work_sessions.start_time < ?", startOfDay, endOfDay)
	} else {
		query = query.Where("DATE(work_sessions.start_time) = ?", dateStr)
	}

	if branchID != nil {
		query = query.Where("employees.branch_id = ?", *branchID)
	}

	var rows []struct {
		EmployeeID       string     `gorm:"column:employee_id"`
		EmployeeName     string     `gorm:"column:employee_name"`
		BranchName       string     `gorm:"column:branch_name"`
		KeyNumber        string     `gorm:"column:key_number"`
		ApplicationID    string     `gorm:"column:application_id"`
		ApplicationType  string     `gorm:"column:application_type"`
		MotorcycleNumber string     `gorm:"column:motorcycle_number"`
		OrdersCount      int        `gorm:"column:orders_count"`
		Distance         float64    `gorm:"column:distance"`
		FuelCost         float64    `gorm:"column:fuel_cost"`
		StartTime        time.Time  `gorm:"column:start_time"`
		EndTime          *time.Time `gorm:"column:end_time"`
		Status           string     `gorm:"column:status"`
		ReviewNotes      string     `gorm:"column:review_notes"`
		IsReviewed       bool       `gorm:"column:is_reviewed"`
	}

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	appSummaryMap := make(map[string]*DailyAppSummary)
	appDriversMap := make(map[string]map[string]bool)
	appMap := map[string]string{
		"ninja":  "نينجا",
		"keeta":  "كيتا",
		"hunger": "هنجر",
		"toyou":  "تو يو",
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

		appType := strings.ToLower(strings.TrimSpace(row.ApplicationType))
		appName := appMap[appType]
		if appName == "" {
			appName = row.ApplicationID
		}
		if appName == "" {
			appName = row.ApplicationType
		}

		empRow := DailyEmployeeRow{
			EmployeeID:       row.EmployeeID,
			EmployeeName:     row.EmployeeName,
			BranchName:       row.BranchName,
			KeyNumber:        row.KeyNumber,
			AppType:          row.ApplicationType,
			AppName:          appName,
			MotorcycleNumber: row.MotorcycleNumber,
			SessionsCount:    1,
			OrdersCount:      row.OrdersCount,
			TotalOrders:      row.OrdersCount,
			Distance:         row.Distance,
			TotalKM:          row.Distance,
			FuelCost:         row.FuelCost,
			TotalFuel:        row.FuelCost,
			WorkingDuration:  dur,
			Status:           row.Status,
			ReviewNotes:      row.ReviewNotes,
			IsReviewed:       row.IsReviewed,
		}
		resp.Employees = append(resp.Employees, empRow)

		if appType == "" {
			appType = "other"
		}
		if _, ok := appSummaryMap[appType]; !ok {
			appSummaryMap[appType] = &DailyAppSummary{
				AppType: appType,
				AppName: appName,
			}
			appDriversMap[appType] = make(map[string]bool)
		}
		s := appSummaryMap[appType]
		s.TotalOrders += row.OrdersCount
		s.TotalKM += row.Distance
		s.TotalFuel += row.FuelCost
		if row.EmployeeID != "" && !appDriversMap[appType][row.EmployeeID] {
			appDriversMap[appType][row.EmployeeID] = true
			s.Count++
		}
	}

	for _, s := range appSummaryMap {
		resp.AppSummaries = append(resp.AppSummaries, *s)
	}
	resp.Rows = resp.Employees

	return resp, nil
}

func (r *gormSystemRepository) GetReports(ctx context.Context, filter ReportFilter) ([]map[string]interface{}, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Table("work_sessions").
		Select("work_sessions.*, employees.name as employee_name, employees.national_id, employees.key_number, branches.name as branch_name").
		Joins("LEFT JOIN employees ON employees.id = work_sessions.employee_id").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Where("work_sessions.deleted_at IS NULL")

	if filter.BranchID != nil {
		query = query.Where("employees.branch_id = ?", *filter.BranchID)
	}
	riyadh := time.FixedZone("Asia/Riyadh", 3*60*60)
	if filter.StartDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", filter.StartDate, riyadh); err == nil {
			query = query.Where("work_sessions.start_time >= ?", time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, riyadh))
		} else {
			query = query.Where("work_sessions.start_time >= ?", filter.StartDate)
		}
	}
	if filter.EndDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", filter.EndDate, riyadh); err == nil {
			endExclusive := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, riyadh).Add(24 * time.Hour)
			query = query.Where("work_sessions.start_time < ?", endExclusive)
		} else {
			query = query.Where("work_sessions.start_time <= ?", filter.EndDate+" 23:59:59")
		}
	}
	if filter.EmployeeID != "" {
		query = query.Where("work_sessions.employee_id = ?", filter.EmployeeID)
	}
	if filter.IsReviewed != nil {
		query = query.Where("work_sessions.is_reviewed = ?", *filter.IsReviewed)
	}
	if filter.Status != "" {
		query = query.Where("work_sessions.status = ?", filter.Status)
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
		for _, imgCol := range []string{"start_km_image", "start_plate_image", "end_km_image"} {
			if s, ok := results[i][imgCol].(string); ok && strings.HasPrefix(s, "data:") && len(s) > 4096 {
				results[i][imgCol] = ""
			}
		}
		results[i]["working_duration"] = "—"
		startTime := parseReportTimestamp(results[i]["start_time"])
		if !startTime.IsZero() {
			endTime := parseReportTimestamp(results[i]["end_time"])
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
			} else {
				results[i]["working_duration"] = "قائم الآن"
			}
		}
	}

	return results, total, nil
}

func parseReportTimestamp(v interface{}) time.Time {
	switch t := v.(type) {
	case time.Time:
		return t
	case *time.Time:
		if t != nil {
			return *t
		}
	case string:
		for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05"} {
			if parsed, err := time.Parse(layout, t); err == nil {
				return parsed
			}
		}
	}
	return time.Time{}
}
