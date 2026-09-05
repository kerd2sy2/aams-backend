package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/dto"
	"delivery-backend/internal/repository"

	"github.com/google/uuid"
)

type TargetService interface {
	GetDashboardSummary(ctx context.Context, month string) (*dto.TargetDashboardSummaryDTO, error)
	ListIdentifiers(ctx context.Context, search, status, month string) ([]dto.IdentifierPerformanceDTO, error)
	GetIdentifierDetails(ctx context.Context, id uuid.UUID, month string) (*dto.IdentifierDetailsDTO, error)
	CreateIdentifier(ctx context.Context, name, code string, monthlyTarget, dailyTarget int) (*domain.Identifier, error)
	UpdateIdentifier(ctx context.Context, id uuid.UUID, name, code string, monthlyTarget, dailyTarget int, isActive bool) error
	DeleteIdentifier(ctx context.Context, id uuid.UUID) error

	ListDrivers(ctx context.Context, search, month string) ([]dto.DriverPerformanceDTO, error)
	ListAlerts(ctx context.Context, date string, unresolvedOnly bool) ([]dto.TargetAlertDTO, error)
	ResolveAlert(ctx context.Context, id uuid.UUID) error

	GetTargetSettings(ctx context.Context) (*dto.TargetSettingsDTO, error)
	UpdateTargetSettings(ctx context.Context, monthlyTarget, dailyTarget int) error
}

type targetService struct {
	targetRepo repository.TargetRepository
}

func NewTargetService(targetRepo repository.TargetRepository) TargetService {
	return &targetService{targetRepo: targetRepo}
}

func (s *targetService) GetDashboardSummary(ctx context.Context, month string) (*dto.TargetDashboardSummaryDTO, error) {
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
	elapsedDays := now.Day()
	if month < now.Format("2006-01") {
		elapsedDays = daysInMonth
	} else if month > now.Format("2006-01") {
		elapsedDays = 0
	}
	remainingDays := daysInMonth - elapsedDays
	if remainingDays < 0 {
		remainingDays = 0
	}

	// 2. Fetch all orders for this month
	orders, err := s.targetRepo.GetOrdersForMonth(ctx, month)
	if err != nil {
		return nil, fmt.Errorf("فشل في جلب طلبات الشهر: %w", err)
	}

	todayDate := now.Format("2006-01-02")
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

	// Build daily trend chart data
	var dailyTrend []dto.DayTrendDTO
	for d := 1; d <= daysInMonth; d++ {
		dateStr := fmt.Sprintf("%s-%02d", month, d)
		dailyTrend = append(dailyTrend, dto.DayTrendDTO{
			Date:   dateStr,
			Day:    d,
			Orders: ordersByDay[d],
			Target: 17,
		})
	}

	// 3. Evaluate Identifiers
	allIdents, err := s.ListIdentifiers(ctx, "", "", month)
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
	alerts, _ := s.ListAlerts(ctx, "", false)
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

func (s *targetService) ListIdentifiers(ctx context.Context, search, statusFilter, month string) ([]dto.IdentifierPerformanceDTO, error) {
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
	elapsedDays := now.Day()
	if month < now.Format("2006-01") {
		elapsedDays = daysInMonth
	} else if month > now.Format("2006-01") {
		elapsedDays = 0
	}
	remainingDays := daysInMonth - elapsedDays
	if remainingDays < 0 {
		remainingDays = 0
	}

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

	// Map orders by identifier
	todayDate := now.Format("2006-01-02")
	weekStartDate := now.AddDate(0, 0, -7).Format("2006-01-02")

	ordersByIdent := make(map[uuid.UUID][]domain.DailyOrder)
	for _, ord := range allMonthOrders {
		ordersByIdent[ord.IdentifierID] = append(ordersByIdent[ord.IdentifierID], ord)
	}

	var result []dto.IdentifierPerformanceDTO

	for _, ident := range idents {
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

		achievedPercent := 0.0
		if mTarget > 0 {
			achievedPercent = math.Round((float64(monthOrders)/float64(mTarget))*1000) / 10
		}

		activeDaysCount := len(uniqueDays)
		if activeDaysCount == 0 && elapsedDays > 0 {
			activeDaysCount = 1
		} else if activeDaysCount == 0 {
			activeDaysCount = 1
		}

		dailyAverage := float64(monthOrders) / float64(activeDaysCount)
		dailyAverage = math.Round(dailyAverage*10) / 10

		remainingTarget := mTarget - monthOrders
		if remainingTarget < 0 {
			remainingTarget = 0
		}

		dailyRequired := 0.0
		if remainingDays > 0 {
			dailyRequired = float64(remainingTarget) / float64(remainingDays)
			dailyRequired = math.Round(dailyRequired*10) / 10
		}

		// Projection calculation
		projected := monthOrders + int(math.Round(dailyAverage*float64(remainingDays)))
		isQualified := projected >= mTarget

		// Status categorization
		status := "ON_TRACK"
		if monthOrders >= mTarget {
			status = "TARGET_ACHIEVED"
		} else if projected >= mTarget {
			status = "ON_TRACK"
		} else if projected >= int(float64(mTarget)*0.8) {
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

		dtoItem := dto.IdentifierPerformanceDTO{
			ID:                       ident.ID,
			Name:                     ident.Name,
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
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	ident, err := s.targetRepo.FindIdentifierByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("المعرف غير موجود: %w", err)
	}

	// Performance calculation
	identsPerf, err := s.ListIdentifiers(ctx, ident.Name, "", month)
	if err != nil || len(identsPerf) == 0 {
		return nil, fmt.Errorf("فشل في احتساب أداء المعرف")
	}
	perf := identsPerf[0]

	// Orders breakdown for this identifier
	orders, err := s.targetRepo.GetOrdersForIdentifierMonth(ctx, id, month)
	if err != nil {
		return nil, err
	}

	driverOrdersMap := make(map[uuid.UUID]int)
	driverNamesMap := make(map[uuid.UUID]string)
	appsBreakdown := make(map[string]int)
	dayOrdersMap := make(map[int]int)

	totalOrders := 0
	for _, ord := range orders {
		totalOrders += ord.OrdersCount
		driverOrdersMap[ord.DriverID] += ord.OrdersCount
		if ord.Driver != nil {
			driverNamesMap[ord.DriverID] = ord.Driver.Name
		}
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

	// Drivers breakdown with percentages
	var driversBreakdown []dto.DriverContributionDTO
	for dID, count := range driverOrdersMap {
		pct := 0.0
		if totalOrders > 0 {
			pct = math.Round((float64(count)/float64(totalOrders))*1000) / 10
		}
		dName := driverNamesMap[dID]
		if dName == "" {
			dName = "مندوب"
		}
		driversBreakdown = append(driversBreakdown, dto.DriverContributionDTO{
			DriverID:   dID,
			DriverName: dName,
			Orders:     count,
			Percentage: pct,
		})
	}

	// Daily timeline
	tMonth, _ := time.Parse("2006-01", month)
	daysInMonth := daysIn(tMonth.Month(), tMonth.Year())
	var dailyTimeline []dto.DayTrendDTO
	for d := 1; d <= daysInMonth; d++ {
		dailyTimeline = append(dailyTimeline, dto.DayTrendDTO{
			Date:   fmt.Sprintf("%s-%02d", month, d),
			Day:    d,
			Orders: dayOrdersMap[d],
			Target: 17,
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

func (s *targetService) CreateIdentifier(ctx context.Context, name, code string, monthlyTarget, dailyTarget int) (*domain.Identifier, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("اسم المعرف مطلوب")
	}

	existing, _ := s.targetRepo.FindIdentifierByName(ctx, name)
	if existing != nil {
		return nil, fmt.Errorf("اسم المعرف '%s' موجود مسبقاً", name)
	}

	if monthlyTarget <= 0 {
		monthlyTarget = 460
	}
	if dailyTarget <= 0 {
		dailyTarget = 17
	}

	ident := domain.Identifier{
		Name:          name,
		Code:          code,
		MonthlyTarget: monthlyTarget,
		DailyTarget:   dailyTarget,
		IsActive:      true,
	}
	if err := s.targetRepo.CreateIdentifier(ctx, &ident); err != nil {
		return nil, err
	}
	return &ident, nil
}

func (s *targetService) UpdateIdentifier(ctx context.Context, id uuid.UUID, name, code string, monthlyTarget, dailyTarget int, isActive bool) error {
	ident, err := s.targetRepo.FindIdentifierByID(ctx, id)
	if err != nil {
		return fmt.Errorf("المعرف غير موجود")
	}

	if name != "" {
		ident.Name = strings.TrimSpace(name)
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

func (s *targetService) ListDrivers(ctx context.Context, search, month string) ([]dto.DriverPerformanceDTO, error) {
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

	todayDate := time.Now().Format("2006-01-02")
	driverMonthOrders := make(map[uuid.UUID]int)
	driverTodayOrders := make(map[uuid.UUID]int)
	driverIdentsMap := make(map[uuid.UUID]map[string]bool)
	driverAppsMap := make(map[uuid.UUID]map[string]bool)

	for _, ord := range orders {
		driverMonthOrders[ord.DriverID] += ord.OrdersCount
		if ord.OrderDate == todayDate {
			driverTodayOrders[ord.DriverID] += ord.OrdersCount
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

	var result []dto.DriverPerformanceDTO
	for _, d := range drivers {
		var idents []string
		for idName := range driverIdentsMap[d.ID] {
			idents = append(idents, idName)
		}
		var apps []string
		for appName := range driverAppsMap[d.ID] {
			apps = append(apps, appName)
		}

		result = append(result, dto.DriverPerformanceDTO{
			ID:          d.ID,
			Name:        d.Name,
			Phone:       d.Phone,
			MonthOrders: driverMonthOrders[d.ID],
			TodayOrders: driverTodayOrders[d.ID],
			Identifiers: idents,
			Apps:        apps,
		})
	}

	return result, nil
}

func (s *targetService) ListAlerts(ctx context.Context, date string, unresolvedOnly bool) ([]dto.TargetAlertDTO, error) {
	alerts, err := s.targetRepo.ListTargetAlerts(ctx, date, unresolvedOnly)
	if err != nil {
		return nil, err
	}

	var result []dto.TargetAlertDTO
	for _, a := range alerts {
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
		dTarget = 17
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
	return nil
}

// Helper: days in month
func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
