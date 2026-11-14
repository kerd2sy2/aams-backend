package target

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *targetService) GetDashboardSummary(ctx context.Context, month string, branch string) (*TargetDashboardSummaryDTO, error) {
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

	orders, err := s.targetRepo.GetOrdersForMonth(ctx, month)
	if err != nil {
		return nil, fmt.Errorf("فشل في جلب طلبات الشهر: %w", err)
	}

	branch = strings.TrimSpace(branch)
	if branch != "" && branch != "all" && branch != "الكل" {
		var filteredOrders []DailyOrder
		for _, ord := range orders {
			if ord.Branch == branch || (ord.Identifier != nil && ord.Identifier.Branch == branch) {
				filteredOrders = append(filteredOrders, ord)
			}
		}
		orders = filteredOrders
	}

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

	var dailyTrend []DayTrendDTO
	for d := 1; d <= daysInMonth; d++ {
		dateStr := fmt.Sprintf("%s-%02d", month, d)
		dailyTrend = append(dailyTrend, DayTrendDTO{
			Date:   dateStr,
			Day:    d,
			Orders: ordersByDay[d],
			Target: defaultDTarget,
		})
	}

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

	alerts, _ := s.ListAlerts(ctx, "", false, branch)
	if len(alerts) > 10 {
		alerts = alerts[:10]
	}

	topList := allIdents
	if len(topList) > 5 {
		topList = topList[:5]
	}

	return &TargetDashboardSummaryDTO{
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


func (s *targetService) ListImportBatches(ctx context.Context, limit int) ([]ImportBatch, error) {
	return s.targetRepo.ListImportBatches(ctx, limit)
}

func (s *targetService) DeleteImportBatch(ctx context.Context, id uuid.UUID) error {
	return s.targetRepo.DeleteImportBatch(ctx, id)
}

func (s *targetService) DeleteSheetByDate(ctx context.Context, orderDate string) error {
	return s.targetRepo.DeleteOrdersByDate(ctx, orderDate)
}

// ITargetContract implementation
func (s *targetService) GetSummary(ctx context.Context, month string) (*TargetSummaryDTO, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	idents, err := s.targetRepo.ListIdentifiers(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	orders, err := s.targetRepo.GetOrdersForMonth(ctx, month)
	if err != nil {
		return nil, err
	}
	totalOrders := 0
	for _, o := range orders {
		totalOrders += o.OrdersCount
	}
	alerts, _ := s.targetRepo.ListTargetAlerts(ctx, "", true)

	return &TargetSummaryDTO{
		TotalIdentifiers: len(idents),
		TotalOrders:      totalOrders,
		ActiveAlerts:     len(alerts),
	}, nil
}

func (s *targetService) GetIdentifierOrders(ctx context.Context, identifierID uuid.UUID, month string) (int, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	orders, err := s.targetRepo.GetOrdersForIdentifierMonth(ctx, identifierID, month)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, o := range orders {
		total += o.OrdersCount
	}
	return total, nil
}

