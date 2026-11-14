package target

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *targetService) GetIdentifierDetails(ctx context.Context, id uuid.UUID, month string) (*IdentifierDetailsDTO, error) {
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

	identsPerf, err := s.ListIdentifiers(ctx, ident.Name, "", month, "")
	if err != nil || len(identsPerf) == 0 {
		return nil, fmt.Errorf("فشل في احتساب أداء المعرف")
	}
	var perf IdentifierPerformanceDTO
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

	driversBreakdown := make([]DriverContributionDTO, 0)
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
		driversBreakdown = append(driversBreakdown, DriverContributionDTO{
			DriverID:    dID,
			DriverName:  dName,
			Orders:      count,
			Percentage:  pct,
			DailyOrders: dMap,
			DaysActive:  daysActive,
		})
	}

	tMonth, _ := time.Parse("2006-01", month)
	daysInMonth := daysIn(tMonth.Month(), tMonth.Year())
	dailyTimeline := make([]DayTrendDTO, 0, daysInMonth)
	for d := 1; d <= daysInMonth; d++ {
		dailyTimeline = append(dailyTimeline, DayTrendDTO{
			Date:   fmt.Sprintf("%s-%02d", month, d),
			Day:    d,
			Orders: dayOrdersMap[d],
			Target: 15,
		})
	}

	return &IdentifierDetailsDTO{
		Performance:        perf,
		DriversBreakdown:   driversBreakdown,
		AppsBreakdown:      appsBreakdown,
		DailyTimeline:      dailyTimeline,
		ActiveDriversCount: len(driversBreakdown),
	}, nil
}

func (s *targetService) CreateIdentifier(ctx context.Context, name, appName, code string, monthlyTarget, dailyTarget int) (*Identifier, error) {
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

	ident := &Identifier{
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

func (s *targetService) UpdateIdentifier(ctx context.Context, id uuid.UUID, name, appName, code string, monthlyTarget, dailyTarget int, isActive bool, accountStatus string) error {
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

	accountStatus = strings.TrimSpace(accountStatus)
	if accountStatus != "" {
		ident.AccountStatus = accountStatus
		if accountStatus == "ACTIVE" {
			ident.IsActive = true
		} else {
			ident.IsActive = false
		}
	} else {
		ident.IsActive = isActive
		if !isActive && (ident.AccountStatus == "" || ident.AccountStatus == "ACTIVE") {
			ident.AccountStatus = "SUSPENDED_TEMP"
		} else if isActive && ident.AccountStatus != "ACTIVE" {
			ident.AccountStatus = "ACTIVE"
		}
	}

	return s.targetRepo.UpdateIdentifier(ctx, ident)
}

func (s *targetService) UpdateIdentifierByIDOrCode(ctx context.Context, idOrCode, name, appName, code string, monthlyTarget, dailyTarget int, isActive bool, accountStatus string) error {
	ident, err := s.targetRepo.FindIdentifierByIDOrCode(ctx, idOrCode)
	if err != nil {
		return err
	}
	if strings.TrimSpace(name) != "" {
		ident.Name = strings.TrimSpace(name)
	}
	if strings.TrimSpace(appName) != "" {
		ident.AppName = strings.TrimSpace(appName)
	}
	if strings.TrimSpace(code) != "" {
		ident.Code = strings.TrimSpace(code)
	}
	if monthlyTarget > 0 {
		ident.MonthlyTarget = monthlyTarget
	}
	if dailyTarget > 0 {
		ident.DailyTarget = dailyTarget
	}

	accountStatus = strings.TrimSpace(accountStatus)
	if accountStatus != "" {
		ident.AccountStatus = accountStatus
		ident.IsActive = (accountStatus == "ACTIVE")
	} else {
		ident.IsActive = isActive
	}

	return s.targetRepo.UpdateIdentifier(ctx, ident)
}

func (s *targetService) LinkEmployeeToIdentifier(ctx context.Context, idOrCode string, employeeID *string) error {
	return s.targetRepo.LinkEmployeeToIdentifierCode(ctx, idOrCode, employeeID)
}

func (s *targetService) DeleteIdentifier(ctx context.Context, id uuid.UUID) error {
	return s.targetRepo.DeleteIdentifier(ctx, id)
}

func (s *targetService) DeleteAllIdentifiers(ctx context.Context) error {
	return s.targetRepo.DeleteAllIdentifiers(ctx)
}

