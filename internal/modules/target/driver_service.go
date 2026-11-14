package target

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *targetService) ListDrivers(ctx context.Context, search, month string, branch string) ([]DriverPerformanceDTO, error) {
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
	driverDailyIdentsMap := make(map[uuid.UUID]map[string]map[string]bool)
	driverDailyShiftsMap := make(map[uuid.UUID]map[string][]DriverDailyShiftDTO)
	driverIdentsMap := make(map[uuid.UUID]map[string]bool)
	driverAppsMap := make(map[uuid.UUID]map[string]bool)
	driverBranchMap := make(map[uuid.UUID]string)

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

		appName := strings.TrimSpace(ord.AppName)
		if appName == "" && ord.Identifier != nil && ord.Identifier.AppName != "" {
			appName = strings.TrimSpace(ord.Identifier.AppName)
		}
		displayApp := appName
		lowerApp := strings.ToLower(appName)
		if lowerApp == "ninja" || appName == "نينجا" {
			displayApp = "نينجا"
		} else if lowerApp == "keeta" || appName == "كيتا" {
			displayApp = "كيتا"
		} else if lowerApp == "toyou" || appName == "تويو" {
			displayApp = "تويو"
		} else if lowerApp == "hungerstation" || appName == "هنقرستيشن" {
			displayApp = "هنقرستيشن"
		} else if lowerApp == "jahez" || appName == "جاهز" {
			displayApp = "جاهز"
		}

		identName := ""
		if ord.Identifier != nil && ord.Identifier.Name != "" {
			identName = strings.TrimSpace(ord.Identifier.Name)
			if driverIdentsMap[ord.DriverID] == nil {
				driverIdentsMap[ord.DriverID] = make(map[string]bool)
			}
			driverIdentsMap[ord.DriverID][identName] = true
		}

		if displayApp != "" {
			if driverAppsMap[ord.DriverID] == nil {
				driverAppsMap[ord.DriverID] = make(map[string]bool)
			}
			driverAppsMap[ord.DriverID][displayApp] = true
		}

		label := ""
		if identName != "" && displayApp != "" {
			label = fmt.Sprintf("%s (%s)", identName, displayApp)
		} else if identName != "" {
			label = identName
		} else if displayApp != "" {
			label = displayApp
		}

		if label != "" {
			if driverDailyIdentsMap[ord.DriverID] == nil {
				driverDailyIdentsMap[ord.DriverID] = make(map[string]map[string]bool)
			}
			if driverDailyIdentsMap[ord.DriverID][ord.OrderDate] == nil {
				driverDailyIdentsMap[ord.DriverID][ord.OrderDate] = make(map[string]bool)
			}
			driverDailyIdentsMap[ord.DriverID][ord.OrderDate][label] = true

			if driverDailyShiftsMap[ord.DriverID] == nil {
				driverDailyShiftsMap[ord.DriverID] = make(map[string][]DriverDailyShiftDTO)
			}
			driverDailyShiftsMap[ord.DriverID][ord.OrderDate] = append(driverDailyShiftsMap[ord.DriverID][ord.OrderDate], DriverDailyShiftDTO{
				IdentifierName: identName,
				AppName:        displayApp,
				OrdersCount:    ord.OrdersCount,
				Label:          label,
			})
		}
	}

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

	result := make([]DriverPerformanceDTO, 0)
	for _, d := range drivers {
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

		dailyIdents := make(map[string][]string)
		if dIdents, ok := driverDailyIdentsMap[d.ID]; ok {
			for dateStr, identSet := range dIdents {
				var idList []string
				for idName := range identSet {
					idList = append(idList, idName)
				}
				dailyIdents[dateStr] = idList
			}
		}

		dailyShifts := driverDailyShiftsMap[d.ID]
		if dailyShifts == nil {
			dailyShifts = make(map[string][]DriverDailyShiftDTO)
		}

		result = append(result, DriverPerformanceDTO{
			ID:               d.ID,
			Name:             d.Name,
			Phone:            d.Phone,
			Branch:           driverBranchMap[d.ID],
			MonthOrders:      driverMonthOrders[d.ID],
			TodayOrders:      driverTodayOrders[d.ID],
			Identifiers:      idents,
			Apps:             apps,
			DailyOrders:      dMap,
			DailyIdentifiers: dailyIdents,
			DailyShifts:      dailyShifts,
			DailyTarget:      18,
			DaysActive:       daysActive,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		rankI := getDriverMasterRank(result[i].Name)
		rankJ := getDriverMasterRank(result[j].Name)
		if rankI != rankJ {
			return rankI < rankJ
		}
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

