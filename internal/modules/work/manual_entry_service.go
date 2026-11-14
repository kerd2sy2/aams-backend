package work

import (
	"context"
	"errors"
	"fmt"

	"strings"
	"time"

	"github.com/google/uuid"

)

func (s *workService) SaveManualEntries(
	ctx context.Context,
	req ManualWorkEntryRequest,
	reviewerID *uuid.UUID,
	reviewerName string,
) ([]WorkSession, error) {
	empUUID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("معرف الموظف غير صالح")
	}

	emp, err := s.empProvider.GetEmployee(ctx, empUUID)
	if err != nil || emp == nil {
		return nil, errors.New("الموظف غير موجود")
	}

	var savedSessions []WorkSession

	err = s.repo.ExecInTx(ctx, func(txRepo WorkRepository) error {
		for _, item := range req.Entries {
			dateStr := strings.TrimSpace(item.Date)
			parsedDay, parseErr := time.ParseInLocation("2006-01-02", dateStr, time.Now().Location())
			if parseErr != nil {
				return fmt.Errorf("صيغة التاريخ غير صالحة (%s)", dateStr)
			}

			startTime := time.Date(parsedDay.Year(), parsedDay.Month(), parsedDay.Day(), 10, 0, 0, 0, parsedDay.Location())
			endTime := time.Date(parsedDay.Year(), parsedDay.Month(), parsedDay.Day(), 22, 0, 0, 0, parsedDay.Location())

			distance := item.Distance
			startKM := item.StartKM
			endKM := item.EndKM

			if distance <= 0 && endKM > startKM {
				distance = endKM - startKM
			} else if distance > 0 && endKM <= 0 && startKM > 0 {
				endKM = startKM + distance
			} else if distance > 0 && startKM <= 0 && endKM <= 0 {
				endKM = distance
			}

			appType := strings.TrimSpace(item.ApplicationType)
			if appType == "" {
				appType = emp.ApplicationType
			}

			noteText := strings.TrimSpace(item.Notes)
			if noteText == "" {
				noteText = "إدخال يدوي من لوحة التحكم"
			}

			existing, _ := txRepo.FindCompletedSessionByEmployeeAndDate(ctx, empUUID, dateStr)
			if existing != nil {
				oldDist := existing.Distance
				oldOrders := existing.OrdersCount
				wasReviewed := existing.IsReviewed

				existing.OrdersCount = item.OrdersCount
				existing.Distance = distance
				existing.FuelCost = item.FuelCost
				if startKM > 0 {
					existing.StartKM = startKM
				}
				if endKM > 0 {
					existing.EndKM = endKM
				}
				if appType != "" {
					existing.ApplicationType = appType
				}
				existing.Notes = noteText
				existing.IsReviewed = true
				existing.ReviewedBy = reviewerID
				existing.IsEditedBySupervisor = true
				existing.EditedByName = reviewerName
				if existing.ReviewNotes == "" {
					existing.ReviewNotes = "معتمد عبر الإدخال اليدوي"
				}

				if err := txRepo.Update(ctx, existing); err != nil {
					return err
				}

				if s.empProvider != nil {
					deltaDist := distance
					deltaOrders := item.OrdersCount
					if wasReviewed {
						deltaDist = distance - oldDist
						deltaOrders = item.OrdersCount - oldOrders
					}
					if deltaDist != 0 || deltaOrders != 0 {
						_ = s.empProvider.UpdateEmployeeOnEndWork(ctx, empUUID, deltaDist, deltaOrders)
					}
				}

				savedSessions = append(savedSessions, *existing)
			} else {
				newSess := &WorkSession{
					ID:                   uuid.New(),
					EmployeeID:           &empUUID,
					StartTime:            startTime,
					EndTime:              &endTime,
					StartKM:              startKM,
					EndKM:                endKM,
					Distance:             distance,
					OrdersCount:          item.OrdersCount,
					FuelCost:             item.FuelCost,
					ApplicationID:        emp.ApplicationID,
					ApplicationType:      appType,
					VehicleType:          emp.VehicleType,
					MotorcycleNumber:     emp.MotorcycleNumber,
					Status:               "COMPLETED",
					IsReviewed:           true,
					ReviewedBy:           reviewerID,
					ReviewNotes:          "معتمد عبر الإدخال اليدوي",
					IsEditedBySupervisor: true,
					EditedByName:         reviewerName,
					Notes:                noteText,
					OriginalStartKM:      startKM,
					OriginalEndKM:        endKM,
					OriginalOrdersCount:  item.OrdersCount,
				}

				if err := txRepo.Create(ctx, newSess); err != nil {
					return err
				}

				if s.empProvider != nil {
					_ = s.empProvider.UpdateEmployeeOnEndWork(ctx, empUUID, distance, item.OrdersCount)
				}

				savedSessions = append(savedSessions, *newSess)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return savedSessions, nil
}


