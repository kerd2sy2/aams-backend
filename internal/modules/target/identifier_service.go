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

func (s *targetService) ListIdentifiers(ctx context.Context, search, statusFilter, month string, branch string) ([]IdentifierPerformanceDTO, error) {
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

	idents, err := s.targetRepo.ListIdentifiers(ctx, search, nil)
	if err != nil {
		return nil, fmt.Errorf("فشل في جلب المعرفين: %w", err)
	}

	allMonthOrders, err := s.targetRepo.GetOrdersForMonth(ctx, month)
	if err != nil {
		return nil, err
	}

	branch = strings.TrimSpace(branch)

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
	refDate := now
	if t, err := time.Parse("2006-01-02", todayDate); err == nil {
		refDate = t
	}
	weekStartDate := refDate.AddDate(0, 0, -6).Format("2006-01-02")

	ordersByIdent := make(map[uuid.UUID][]DailyOrder)
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

	result := make([]IdentifierPerformanceDTO, 0)

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

		effectiveElapsedDays := elapsedDays
		if effectiveElapsedDays <= 0 {
			effectiveElapsedDays = 1
		}

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

		recentPace := dailyAverage
		if weekOrders > 0 {
			recentPace = float64(weekOrders) / 7.0
		}

		blendedRate := (0.65 * dailyAverage) + (0.35 * recentPace)
		projected := int(math.Round(float64(monthOrders) + (blendedRate * float64(remainingDays))))
		if monthOrders > projected {
			projected = monthOrders
		}

		expectedToDate := (float64(mTarget) / float64(daysInMonth)) * float64(elapsedDays)

		status := "ON_TRACK"
		if monthOrders >= mTarget {
			status = "TARGET_ACHIEVED"
		} else if projected >= mTarget || (float64(monthOrders) >= expectedToDate && dailyRequired <= 18.0) {
			status = "ON_TRACK"
		} else if projected >= 390 || (dailyRequired <= 20.5 && monthOrders >= 120) {
			status = "AT_RISK"
		} else {
			status = "BEHIND_TARGET"
		}
		isQualified := status == "ON_TRACK" || status == "TARGET_ACHIEVED"

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

		driverName := ""
		for _, d := range ident.Drivers {
			dName := strings.TrimSpace(d.Name)
			if dName != "" && !isAllDigits(dName) && !strings.HasPrefix(dName, "كابتن") {
				driverName = dName
				break
			}
		}
		if driverName == "" && len(ident.Drivers) > 0 {
			driverName = strings.TrimSpace(ident.Drivers[0].Name)
		}
		if driverName == "" && !isAllDigits(ident.Name) {
			driverName = ident.Name
		}

		identCode := ident.Code
		if identCode == "" && isAllDigits(ident.Name) {
			identCode = ident.Name
		}

		dtoItem := IdentifierPerformanceDTO{
			ID:                       ident.ID,
			Name:                     ident.Name,
			AppName:                  appName,
			Branch:                   ident.Branch,
			Code:                     identCode,
			DriverName:               driverName,
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
			AccountStatus:            ident.AccountStatus,
		}
		if dtoItem.AccountStatus == "" {
			if ident.IsActive {
				dtoItem.AccountStatus = "ACTIVE"
			} else {
				dtoItem.AccountStatus = "SUSPENDED_TEMP"
			}
		}

		result = append(result, dtoItem)
	}

	// Deduplicate: If an item has a Code and DriverName, omit text-only items of the same app that match that DriverName
	codedMap := make(map[string]bool)
	for _, it := range result {
		if it.Code != "" && it.DriverName != "" {
			key := strings.ToLower(it.AppName) + "_" + strings.TrimSpace(it.DriverName)
			codedMap[key] = true
		}
	}

	deduped := make([]IdentifierPerformanceDTO, 0, len(result))
	seenCodes := make(map[string]bool)
	for _, it := range result {
		appNorm := strings.ToLower(it.AppName)
		if it.Code != "" {
			codeKey := appNorm + "_" + it.Code
			if seenCodes[codeKey] {
				continue
			}
			seenCodes[codeKey] = true
			deduped = append(deduped, it)
		} else {
			// Uncoded text item: only keep if not already represented in a coded item
			nameKey := appNorm + "_" + strings.TrimSpace(it.Name)
			if codedMap[nameKey] {
				continue
			}
			deduped = append(deduped, it)
		}
	}
	result = deduped

	return result, nil
}

