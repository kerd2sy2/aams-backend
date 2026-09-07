package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/dto"
	"delivery-backend/internal/repository"

	"github.com/google/uuid"
)

type TargetService interface {
	GetDashboardSummary(ctx context.Context, month string, branch string) (*dto.TargetDashboardSummaryDTO, error)
	ListIdentifiers(ctx context.Context, search, status, month string, branch string) ([]dto.IdentifierPerformanceDTO, error)
	GetIdentifierDetails(ctx context.Context, id uuid.UUID, month string) (*dto.IdentifierDetailsDTO, error)
	CreateIdentifier(ctx context.Context, name, appName, code string, monthlyTarget, dailyTarget int) (*domain.Identifier, error)
	UpdateIdentifier(ctx context.Context, id uuid.UUID, name, appName, code string, monthlyTarget, dailyTarget int, isActive bool) error
	DeleteIdentifier(ctx context.Context, id uuid.UUID) error
	DeleteAllIdentifiers(ctx context.Context) error

	ListDrivers(ctx context.Context, search, month string, branch string) ([]dto.DriverPerformanceDTO, error)
	ListAlerts(ctx context.Context, date string, unresolvedOnly bool, branch string) ([]dto.TargetAlertDTO, error)
	ResolveAlert(ctx context.Context, id uuid.UUID) error

	GetTargetSettings(ctx context.Context) (*dto.TargetSettingsDTO, error)
	UpdateTargetSettings(ctx context.Context, monthlyTarget, dailyTarget int) error

	ListImportBatches(ctx context.Context, limit int) ([]domain.ImportBatch, error)
	DeleteImportBatch(ctx context.Context, id uuid.UUID) error
	DeleteSheetByDate(ctx context.Context, orderDate string) error
}

type targetService struct {
	targetRepo repository.TargetRepository
}

func NewTargetService(targetRepo repository.TargetRepository) TargetService {
	return &targetService{targetRepo: targetRepo}
}

func (s *targetService) GetDashboardSummary(ctx context.Context, month string, branch string) (*dto.TargetDashboardSummaryDTO, error) {
	targetDay := ""
	if len(month) >= 10 {
		targetDay = month[:10]
		month = month[:7]
	}
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	// 1. Calculate month date bounds
	tMonth, err := time.Parse("2006-01", month)
	if err != nil {
		tMonth = time.Now()
		month = tMonth.Format("2006-01")
	}
	daysInMonth := daysIn(tMonth.Month(), tMonth.Year())
	now := time.Now()

	// 2. Fetch all orders for this month
	orders, err := s.targetRepo.GetOrdersForMonth(ctx, month)
	if err != nil {
		return nil, fmt.Errorf("فشل في جلب طلبات الشهر: %w", err)
	}

	// Filter by branch if specified
	branch = strings.TrimSpace(branch)
	if branch != "" && branch != "all" && branch != "الكل" {
		var filteredOrders []domain.DailyOrder
		for _, ord := range orders {
			if ord.Branch == branch || (ord.Identifier != nil && ord.Identifier.Branch == branch) {
				filteredOrders = append(filteredOrders, ord)
			}
		}
		orders = filteredOrders
	}

	// Find the latest day with uploaded orders in this month
	maxOrderDay := 0
	maxOrderDate := ""
	for _, ord := range orders {
		if ord.OrdersCount > 0 {
			if ord.OrderDate > maxOrderDate {
				maxOrderDate = ord.OrderDate
			}
			if len(ord.OrderDate) >= 10 {
				if d, err := strconv.Atoi(ord.OrderDate[8:10]); err == nil && d > maxOrderDay {
					maxOrderDay = d
				}
			}
		}
	}

	elapsedDays := now.Day()
	if targetDay != "" {
		if tDay, err := time.Parse("2006-01-02", targetDay); err == nil {
			elapsedDays = tDay.Day()
		}
	} else if month < now.Format("2006-01") {
		elapsedDays = daysInMonth
	} else if month > now.Format("2006-01") {
		elapsedDays = 0
	} else {
		// In current month, shifts finish at 4 AM next day and sheets represent closed shift dates.
		// Elapsed days must match the latest date with uploaded orders (maxOrderDay).
		if maxOrderDay > 0 {
			elapsedDays = maxOrderDay
		} else if now.Day() > 1 {
			elapsedDays = now.Day() - 1
		} else {
			elapsedDays = 1
		}
	}
	if elapsedDays > daysInMonth {
		elapsedDays = daysInMonth
	}
	remainingDays := daysInMonth - elapsedDays
	if remainingDays < 0 {
		remainingDays = 0
	}

	todayDate := now.Format("2006-01-02")
	if targetDay != "" {
		todayDate = targetDay
	} else if maxOrderDate != "" {
		todayDate = maxOrderDate
	}
	totalMonthOrders := 0
	todayTotalOrders := 0
	ordersByDay := make(map[int]int)

	for _, ord := range orders {
		totalMonthOrders += ord.OrdersCount
		if ord.OrderDate == todayDate {
			todayTotalOrders += ord.OrdersCount
		}
		// parse day from YYYY-MM-DD
		if len(ord.OrderDate) >= 10 {
			dayNum, _ := strconv.Atoi(ord.OrderDate[8:10])
			ordersByDay[dayNum] += ord.OrdersCount
		}
	}

	dVal, _ := s.targetRepo.GetTargetSetting(ctx, "DEFAULT_DAILY_TARGET")
	defaultDTarget, _ := strconv.Atoi(dVal)
	if defaultDTarget <= 0 {
		defaultDTarget = 18
	}

	// Build daily trend chart data
	var dailyTrend []dto.DayTrendDTO
	for d := 1; d <= daysInMonth; d++ {
		dateStr := fmt.Sprintf("%s-%02d", month, d)
		dailyTrend = append(dailyTrend, dto.DayTrendDTO{
			Date:   dateStr,
			Day:    d,
			Orders: ordersByDay[d],
			Target: defaultDTarget,
		})
	}

	// 3. Evaluate Identifiers
	allIdents, err := s.ListIdentifiers(ctx, "", "", month, branch)
	if err != nil {
		return nil, err
	}

	achieved := 0
	onTrack := 0
	atRisk := 0
	behind := 0

	for _, ident := range allIdents {
		switch ident.Status {
		case "TARGET_ACHIEVED":
			achieved++
		case "ON_TRACK":
			onTrack++
		case "AT_RISK":
			atRisk++
		case "BEHIND_TARGET":
			behind++
		}
	}

	// 4. Fetch recent alerts
	alerts, _ := s.ListAlerts(ctx, "", false, branch)
	if len(alerts) > 10 {
		alerts = alerts[:10]
	}

	// Top identifiers
	topList := allIdents
	if len(topList) > 5 {
		topList = topList[:5]
	}

	return &dto.TargetDashboardSummaryDTO{
		TotalIdentifiers: len(allIdents),
		TargetAchieved:   achieved,
		OnTrack:          onTrack,
		AtRisk:           atRisk,
		BehindTarget:     behind,
		TotalMonthOrders: totalMonthOrders,
		TodayTotalOrders: todayTotalOrders,
		DailyTrend:       dailyTrend,
		TopIdentifiers:   topList,
		RecentAlerts:     alerts,
		CurrentMonth:     month,
		DaysElapsed:      elapsedDays,
		TotalDaysInMonth: daysInMonth,
		RemainingDays:    remainingDays,
	}, nil
}

func (s *targetService) ListIdentifiers(ctx context.Context, search, statusFilter, month string, branch string) ([]dto.IdentifierPerformanceDTO, error) {
	targetDay := ""
	if len(month) >= 10 {
		targetDay = month[:10]
		month = month[:7]
	}
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	tMonth, err := time.Parse("2006-01", month)
	if err != nil {
		tMonth = time.Now()
		month = tMonth.Format("2006-01")
	}
	daysInMonth := daysIn(tMonth.Month(), tMonth.Year())
	now := time.Now()

	// Fetch all identifiers from DB
	idents, err := s.targetRepo.ListIdentifiers(ctx, search, nil)
	if err != nil {
		return nil, fmt.Errorf("فشل في جلب المعرفين: %w", err)
	}

	// Fetch all orders for this month
	allMonthOrders, err := s.targetRepo.GetOrdersForMonth(ctx, month)
	if err != nil {
		return nil, err
	}

	branch = strings.TrimSpace(branch)

	// Find latest day with uploaded orders in this month
	maxOrderDay := 0
	maxOrderDate := ""
	for _, ord := range allMonthOrders {
		if ord.OrdersCount > 0 {
			if ord.OrderDate > maxOrderDate {
				maxOrderDate = ord.OrderDate
			}
			if len(ord.OrderDate) >= 10 {
				if d, err := strconv.Atoi(ord.OrderDate[8:10]); err == nil && d > maxOrderDay {
					maxOrderDay = d
				}
			}
		}
	}

	elapsedDays := now.Day()
	if targetDay != "" {
		if tDay, err := time.Parse("2006-01-02", targetDay); err == nil {
			elapsedDays = tDay.Day()
		}
	} else if month < now.Format("2006-01") {
		elapsedDays = daysInMonth
	} else if month > now.Format("2006-01") {
		elapsedDays = 0
	} else {
		// Shifts end at 4 AM and sheets reflect closed shift dates.
		// Elapsed days must match the latest date with uploaded orders (maxOrderDay).
		if maxOrderDay > 0 {
			elapsedDays = maxOrderDay
		} else if now.Day() > 1 {
			elapsedDays = now.Day() - 1
		} else {
			elapsedDays = 1
		}
	}
	if elapsedDays > daysInMonth {
		elapsedDays = daysInMonth
	}
	remainingDays := daysInMonth - elapsedDays
	if remainingDays < 0 {
		remainingDays = 0
	}

	// Map orders by identifier
	todayDate := now.Format("2006-01-02")
	if targetDay != "" {
		todayDate = targetDay
	} else if maxOrderDate != "" {
		todayDate = maxOrderDate
	}
	weekStartDate := now.AddDate(0, 0, -7).Format("2006-01-02")

	ordersByIdent := make(map[uuid.UUID][]domain.DailyOrder)
	for _, ord := range allMonthOrders {
		if ord.IdentifierID != nil {
			if branch != "" && branch != "all" && branch != "الكل" {
				if ord.Branch != "" && ord.Branch != branch {
					continue
				}
			}
			ordersByIdent[*ord.IdentifierID] = append(ordersByIdent[*ord.IdentifierID], ord)
		}
	}

	result := make([]dto.IdentifierPerformanceDTO, 0)

	for _, ident := range idents {
		if branch != "" && branch != "all" && branch != "الكل" {
			if ident.Branch != "" && ident.Branch != branch {
				continue
			}
			if ident.Branch == "" && len(ordersByIdent[ident.ID]) == 0 {
				continue
			}
		}

		orders := ordersByIdent[ident.ID]

		monthOrders := 0
		todayOrders := 0
		weekOrders := 0
		uniqueDays := make(map[string]bool)

		for _, ord := range orders {
			monthOrders += ord.OrdersCount
			uniqueDays[ord.OrderDate] = true
			if ord.OrderDate == todayDate {
				todayOrders += ord.OrdersCount
			}
			if ord.OrderDate >= weekStartDate && ord.OrderDate <= todayDate {
				weekOrders += ord.OrdersCount
			}
		}

		mTarget := ident.MonthlyTarget
		if mTarget <= 0 {
			mTarget = 460
		}
		dTarget := ident.DailyTarget
		if dTarget <= 0 {
			dTarget = 18
		}

		achievedPercent := 0.0
		if mTarget > 0 {
			achievedPercent = math.Round((float64(monthOrders)/float64(mTarget))*1000) / 10
		}

		// Effective elapsed days for rate calculation (at least 1 to prevent division by zero)
		effectiveElapsedDays := elapsedDays
		if effectiveElapsedDays <= 0 {
			effectiveElapsedDays = 1
		}

		// Daily average is total month orders divided by elapsed days in the month
		dailyAverage := float64(monthOrders) / float64(effectiveElapsedDays)
		dailyAverage = math.Round(dailyAverage*10) / 10

		remainingTarget := mTarget - monthOrders
		if remainingTarget < 0 {
			remainingTarget = 0
		}

		dailyRequired := 0.0
		if remainingDays > 0 {
			dailyRequired = math.Round(float64(remainingTarget) / float64(remainingDays))
		}

		// Projection calculation: based on actual daily run rate across total month days
		projected := int(math.Round(dailyAverage * float64(daysInMonth)))
		if monthOrders > projected {
			projected = monthOrders
		}
		isQualified := projected >= mTarget

		// Status categorization based explicitly on projected orders:
		// 1. TARGET_ACHIEVED: achieved full monthly target (monthOrders >= mTarget)
		// 2. ON_TRACK: projected 460 or more (projected >= 460) -> يسير بالمعدل
		// 3. AT_RISK: projected 310 to 459 (projected >= 310) -> على وشك المعدل
		// 4. BEHIND_TARGET: projected less than 310 (projected < 310) -> متأخر
		status := "ON_TRACK"
		if monthOrders >= mTarget {
			status = "TARGET_ACHIEVED"
		} else if projected >= 460 {
			status = "ON_TRACK"
		} else if projected >= 310 {
			status = "AT_RISK"
		} else {
			status = "BEHIND_TARGET"
		}

		// Estimated Achievement Date
		estDate := "غير محدد"
		if monthOrders >= mTarget {
			estDate = "تم التحقيق"
		} else if dailyAverage > 0 {
			daysToAchieve := int(math.Ceil(float64(remainingTarget) / dailyAverage))
			estTime := now.AddDate(0, 0, daysToAchieve)
			estDate = estTime.Format("2006-01-02")
		}

		if statusFilter != "" && status != statusFilter {
			continue
		}

		appName := ident.AppName
		if appName == "" && len(orders) > 0 && orders[0].AppName != "" {
			appName = orders[0].AppName
		}

		dtoItem := dto.IdentifierPerformanceDTO{
			ID:                       ident.ID,
			Name:                     ident.Name,
			AppName:                  appName,
			Branch:                   ident.Branch,
			Code:                     ident.Code,
			TodayOrders:              todayOrders,
			WeekOrders:               weekOrders,
			MonthOrders:              monthOrders,
			MonthlyTarget:            mTarget,
			AchievementPercent:       achievedPercent,
			DailyAverage:             dailyAverage,
			DailyRequired:            dailyRequired,
			RemainingDays:            remainingDays,
			Status:                   status,
			ProjectedMonthlyOrders:   projected,
			EstimatedAchievementDate: estDate,
			IsQualified:              isQualified,
			IsActive:                 ident.IsActive,
		}

		result = append(result, dtoItem)
	}

	return result, nil
}

func (s *targetService) GetIdentifierDetails(ctx context.Context, id uuid.UUID, month string) (*dto.IdentifierDetailsDTO, error) {
	if len(month) >= 10 {
		month = month[:7]
	}
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	ident, err := s.targetRepo.FindIdentifierByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("المعرف غير موجود: %w", err)
	}

	// Performance calculation
	identsPerf, err := s.ListIdentifiers(ctx, ident.Name, "", month, "")
	if err != nil || len(identsPerf) == 0 {
		return nil, fmt.Errorf("فشل في احتساب أداء المعرف")
	}
	var perf dto.IdentifierPerformanceDTO
	found := false
	for _, p := range identsPerf {
		if p.ID == ident.ID {
			perf = p
			found = true
			break
		}
	}
	if !found && len(identsPerf) > 0 {
		perf = identsPerf[0]
	}

	// Orders breakdown for this identifier
	orders, err := s.targetRepo.GetOrdersForIdentifierMonth(ctx, id, month)
	if err != nil {
		return nil, err
	}

	driverOrdersMap := make(map[uuid.UUID]int)
	driverNamesMap := make(map[uuid.UUID]string)
	driverDailyMap := make(map[uuid.UUID]map[string]int)
	appsBreakdown := make(map[string]int)
	dayOrdersMap := make(map[int]int)

	totalOrders := 0
	for _, ord := range orders {
		totalOrders += ord.OrdersCount
		driverOrdersMap[ord.DriverID] += ord.OrdersCount
		if ord.Driver != nil {
			driverNamesMap[ord.DriverID] = ord.Driver.Name
		}
		if driverDailyMap[ord.DriverID] == nil {
			driverDailyMap[ord.DriverID] = make(map[string]int)
		}
		driverDailyMap[ord.DriverID][ord.OrderDate] += ord.OrdersCount

		app := ord.AppName
		if app == "" {
			app = "أخرى"
		}
		appsBreakdown[app] += ord.OrdersCount

		if len(ord.OrderDate) >= 10 {
			dNum, _ := strconv.Atoi(ord.OrderDate[8:10])
			dayOrdersMap[dNum] += ord.OrdersCount
		}
	}

	// Drivers breakdown with percentages and daily orders breakdown
	driversBreakdown := make([]dto.DriverContributionDTO, 0)
	for dID, count := range driverOrdersMap {
		pct := 0.0
		if totalOrders > 0 {
			pct = math.Round((float64(count)/float64(totalOrders))*1000) / 10
		}
		dName := driverNamesMap[dID]
		if dName == "" {
			dName = "مندوب"
		}
		dMap := driverDailyMap[dID]
		daysActive := len(dMap)
		driversBreakdown = append(driversBreakdown, dto.DriverContributionDTO{
			DriverID:    dID,
			DriverName:  dName,
			Orders:      count,
			Percentage:  pct,
			DailyOrders: dMap,
			DaysActive:  daysActive,
		})
	}

	// Daily timeline
	tMonth, _ := time.Parse("2006-01", month)
	daysInMonth := daysIn(tMonth.Month(), tMonth.Year())
	dailyTimeline := make([]dto.DayTrendDTO, 0, daysInMonth)
	for d := 1; d <= daysInMonth; d++ {
		dailyTimeline = append(dailyTimeline, dto.DayTrendDTO{
			Date:   fmt.Sprintf("%s-%02d", month, d),
			Day:    d,
			Orders: dayOrdersMap[d],
			Target: 15,
		})
	}

	return &dto.IdentifierDetailsDTO{
		Performance:        perf,
		DriversBreakdown:   driversBreakdown,
		AppsBreakdown:      appsBreakdown,
		DailyTimeline:      dailyTimeline,
		ActiveDriversCount: len(driversBreakdown),
	}, nil
}

func (s *targetService) CreateIdentifier(ctx context.Context, name, appName, code string, monthlyTarget, dailyTarget int) (*domain.Identifier, error) {
	name = strings.TrimSpace(name)
	appName = strings.TrimSpace(appName)
	if name == "" {
		return nil, fmt.Errorf("اسم المعرف مطلوب")
	}

	existing, _ := s.targetRepo.FindIdentifierByNameAndApp(ctx, name, appName)
	if existing != nil {
		if appName != "" {
			return nil, fmt.Errorf("اسم المعرف '%s' للتطبيق '%s' موجود مسبقاً", name, appName)
		}
		return nil, fmt.Errorf("اسم المعرف '%s' موجود مسبقاً", name)
	}

	if monthlyTarget <= 0 {
		monthlyTarget = 460
	}
	if dailyTarget <= 0 {
		dailyTarget = 18
	}

	ident := &domain.Identifier{
		Name:          name,
		AppName:       appName,
		Code:          strings.TrimSpace(code),
		MonthlyTarget: monthlyTarget,
		DailyTarget:   dailyTarget,
		IsActive:      true,
	}

	if err := s.targetRepo.CreateIdentifier(ctx, ident); err != nil {
		return nil, fmt.Errorf("فشل في إنشاء المعرف: %w", err)
	}
	return ident, nil
}

func (s *targetService) UpdateIdentifier(ctx context.Context, id uuid.UUID, name, appName, code string, monthlyTarget, dailyTarget int, isActive bool) error {
	ident, err := s.targetRepo.FindIdentifierByID(ctx, id)
	if err != nil {
		return fmt.Errorf("المعرف غير موجود: %w", err)
	}

	if strings.TrimSpace(name) != "" {
		ident.Name = strings.TrimSpace(name)
	}
	if strings.TrimSpace(appName) != "" {
		ident.AppName = strings.TrimSpace(appName)
	}
	ident.Code = strings.TrimSpace(code)
	if monthlyTarget > 0 {
		ident.MonthlyTarget = monthlyTarget
	}
	if dailyTarget > 0 {
		ident.DailyTarget = dailyTarget
	}
	ident.IsActive = isActive

	return s.targetRepo.UpdateIdentifier(ctx, ident)
}

func (s *targetService) DeleteIdentifier(ctx context.Context, id uuid.UUID) error {
	return s.targetRepo.DeleteIdentifier(ctx, id)
}

func (s *targetService) DeleteAllIdentifiers(ctx context.Context) error {
	return s.targetRepo.DeleteAllIdentifiers(ctx)
}

func (s *targetService) ListDrivers(ctx context.Context, search, month string, branch string) ([]dto.DriverPerformanceDTO, error) {
	targetDay := ""
	if len(month) >= 10 {
		targetDay = month[:10]
		month = month[:7]
	}
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	drivers, err := s.targetRepo.ListDrivers(ctx, search)
	if err != nil {
		return nil, err
	}

	orders, err := s.targetRepo.GetOrdersForMonth(ctx, month)
	if err != nil {
		return nil, err
	}

	branch = strings.TrimSpace(branch)

	// Find the latest order date uploaded
	maxOrderDate := ""
	for _, ord := range orders {
		if ord.OrdersCount > 0 && ord.OrderDate > maxOrderDate {
			maxOrderDate = ord.OrderDate
		}
	}

	todayDate := time.Now().Format("2006-01-02")
	if targetDay != "" {
		todayDate = targetDay
	} else if maxOrderDate != "" {
		todayDate = maxOrderDate
	}
	driverMonthOrders := make(map[uuid.UUID]int)
	driverTodayOrders := make(map[uuid.UUID]int)
	driverDailyMap := make(map[uuid.UUID]map[string]int)
	driverIdentsMap := make(map[uuid.UUID]map[string]bool)
	driverAppsMap := make(map[uuid.UUID]map[string]bool)
	driverBranchMap := make(map[uuid.UUID]string)

	// Preserve the exact sequence of drivers as they appear in the uploaded Excel sheet
	driverSheetOrder := make(map[uuid.UUID]int)
	sheetIndex := 0

	for _, ord := range orders {
		if branch != "" && branch != "all" && branch != "الكل" {
			if ord.Branch != "" && ord.Branch != branch {
				continue
			}
		}

		driverMonthOrders[ord.DriverID] += ord.OrdersCount
		if driverDailyMap[ord.DriverID] == nil {
			driverDailyMap[ord.DriverID] = make(map[string]int)
		}
		driverDailyMap[ord.DriverID][ord.OrderDate] += ord.OrdersCount

		if ord.Branch != "" {
			driverBranchMap[ord.DriverID] = ord.Branch
		}
		if ord.OrderDate == todayDate {
			driverTodayOrders[ord.DriverID] += ord.OrdersCount
			if _, exists := driverSheetOrder[ord.DriverID]; !exists {
				driverSheetOrder[ord.DriverID] = sheetIndex
				sheetIndex++
			}
		}
		if ord.Identifier != nil {
			if driverIdentsMap[ord.DriverID] == nil {
				driverIdentsMap[ord.DriverID] = make(map[string]bool)
			}
			driverIdentsMap[ord.DriverID][ord.Identifier.Name] = true
		}
		if ord.AppName != "" {
			if driverAppsMap[ord.DriverID] == nil {
				driverAppsMap[ord.DriverID] = make(map[string]bool)
			}
			driverAppsMap[ord.DriverID][ord.AppName] = true
		}
	}

	// For drivers who have orders this month but not on today's date, continue sequence
	for _, ord := range orders {
		if branch != "" && branch != "all" && branch != "الكل" {
			if ord.Branch != "" && ord.Branch != branch {
				continue
			}
		}
		if _, exists := driverSheetOrder[ord.DriverID]; !exists {
			driverSheetOrder[ord.DriverID] = sheetIndex
			sheetIndex++
		}
	}

	result := make([]dto.DriverPerformanceDTO, 0)
	for _, d := range drivers {
		// If branch filter is active, only include drivers who have orders in this branch
		if branch != "" && branch != "all" && branch != "الكل" {
			if driverMonthOrders[d.ID] == 0 && driverTodayOrders[d.ID] == 0 {
				continue
			}
		}

		var idents []string
		for idName := range driverIdentsMap[d.ID] {
			idents = append(idents, idName)
		}
		var apps []string
		for appName := range driverAppsMap[d.ID] {
			apps = append(apps, appName)
		}

		dMap := driverDailyMap[d.ID]
		daysActive := len(dMap)

		result = append(result, dto.DriverPerformanceDTO{
			ID:          d.ID,
			Name:        d.Name,
			Phone:       d.Phone,
			Branch:      driverBranchMap[d.ID],
			MonthOrders: driverMonthOrders[d.ID],
			TodayOrders: driverTodayOrders[d.ID],
			Identifiers: idents,
			Apps:        apps,
			DailyOrders: dMap,
			DailyTarget: 15,
			DaysActive:  daysActive,
		})
	}

	// Sort result so drivers appearing in the sheet are ordered exactly as in the sheet!
	sort.SliceStable(result, func(i, j int) bool {
		orderI, hasI := driverSheetOrder[result[i].ID]
		orderJ, hasJ := driverSheetOrder[result[j].ID]
		if hasI && hasJ {
			return orderI < orderJ
		}
		if hasI {
			return true
		}
		if hasJ {
			return false
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (s *targetService) ListAlerts(ctx context.Context, date string, unresolvedOnly bool, branch string) ([]dto.TargetAlertDTO, error) {
	alerts, err := s.targetRepo.ListTargetAlerts(ctx, date, unresolvedOnly)
	if err != nil {
		return nil, err
	}

	branch = strings.TrimSpace(branch)

	result := make([]dto.TargetAlertDTO, 0)
	for _, a := range alerts {
		if branch != "" && branch != "all" && branch != "الكل" {
			if a.Identifier != nil && a.Identifier.Branch != "" && a.Identifier.Branch != branch {
				continue
			}
		}

		identName := "معرف"
		if a.Identifier != nil {
			identName = a.Identifier.Name
		}
		result = append(result, dto.TargetAlertDTO{
			ID:             a.ID,
			IdentifierID:   a.IdentifierID,
			IdentifierName: identName,
			AlertDate:      a.AlertDate,
			TargetOrders:   a.TargetOrders,
			ActualOrders:   a.ActualOrders,
			Deficit:        a.Deficit,
			IsResolved:     a.IsResolved,
			CreatedAt:      a.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	return result, nil
}

func (s *targetService) ResolveAlert(ctx context.Context, id uuid.UUID) error {
	return s.targetRepo.ResolveAlert(ctx, id)
}

func (s *targetService) GetTargetSettings(ctx context.Context) (*dto.TargetSettingsDTO, error) {
	mVal, _ := s.targetRepo.GetTargetSetting(ctx, "DEFAULT_MONTHLY_TARGET")
	dVal, _ := s.targetRepo.GetTargetSetting(ctx, "DEFAULT_DAILY_TARGET")

	mTarget, _ := strconv.Atoi(mVal)
	if mTarget <= 0 {
		mTarget = 460
	}
	dTarget, _ := strconv.Atoi(dVal)
	if dTarget <= 0 {
		dTarget = 18
	}

	return &dto.TargetSettingsDTO{
		DefaultMonthlyTarget: mTarget,
		DefaultDailyTarget:   dTarget,
	}, nil
}

func (s *targetService) UpdateTargetSettings(ctx context.Context, monthlyTarget, dailyTarget int) error {
	if monthlyTarget > 0 {
		_ = s.targetRepo.SetTargetSetting(ctx, "DEFAULT_MONTHLY_TARGET", strconv.Itoa(monthlyTarget))
	}
	if dailyTarget > 0 {
		_ = s.targetRepo.SetTargetSetting(ctx, "DEFAULT_DAILY_TARGET", strconv.Itoa(dailyTarget))
	}
	if monthlyTarget > 0 || dailyTarget > 0 {
		_ = s.targetRepo.UpdateAllIdentifiersTargets(ctx, monthlyTarget, dailyTarget)
	}
	return nil
}

func (s *targetService) ListImportBatches(ctx context.Context, limit int) ([]domain.ImportBatch, error) {
	return s.targetRepo.ListImportBatches(ctx, limit)
}

func (s *targetService) DeleteImportBatch(ctx context.Context, id uuid.UUID) error {
	return s.targetRepo.DeleteImportBatch(ctx, id)
}

func (s *targetService) DeleteSheetByDate(ctx context.Context, orderDate string) error {
	return s.targetRepo.DeleteOrdersByDate(ctx, orderDate)
}

// Helper: days in month
func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
