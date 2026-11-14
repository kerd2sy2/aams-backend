package custody

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *custodyService) AddExpense(ctx context.Context, dayID uuid.UUID, branchID *uuid.UUID, req CreateCustodyExpenseRequest, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error) {
	day, err := s.repo.FindDayByID(ctx, dayID)
	if err != nil {
		return nil, errors.New("سجل العهدة غير موجود")
	}

	cleanImg := s.sanitizeImageURL(req.ImageURL)

	expense := &CustodyExpense{
		ID:                uuid.New(),
		CustodyDayID:      dayID,
		Category:          req.Category,
		Amount:            req.Amount,
		RecipientName:     req.RecipientName,
		ExportNumber:      req.ExportNumber,
		ImageURL:          cleanImg,
		CreatedByID:       adminID,
		CreatedByName:     adminName,
		CreatedByUsername: adminUsername,
		CreatedAt:         time.Now(),
	}

	if err := s.repo.CreateExpense(ctx, expense); err != nil {
		return nil, err
	}

	logEntry := &CustodyLog{
		BranchID:      day.BranchID,
		CustodyDayID:  day.ID,
		Date:          day.Date,
		ActionType:    "ADD_EXPENSE",
		Category:      req.Category,
		Amount:        req.Amount,
		Description:   fmt.Sprintf("إضافة مصروف: %s - %.2f", req.Category, req.Amount),
		RecipientName: req.RecipientName,
		ExportNumber:  req.ExportNumber,
		ImageURL:      cleanImg,
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateLog(ctx, logEntry)

	if updatedDay, _ := s.cascadeRecalculateChain(ctx, day.BranchID, day.Date); updatedDay != nil && updatedDay.ID == day.ID {
		day = updatedDay
	} else {
		day.Expenses = append(day.Expenses, *expense)
		s.recomputeClosing(day)
		_ = s.repo.UpdateDay(ctx, day)
	}

	return s.toResponse(day), nil
}

func (s *custodyService) UpdateExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, req UpdateCustodyExpenseRequest, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error) {
	expense, err := s.repo.FindExpenseByID(ctx, expenseID)
	if err != nil {
		return nil, errors.New("المصروف غير موجود")
	}

	day, err := s.repo.FindDayByID(ctx, expense.CustodyDayID)
	if err != nil {
		return nil, errors.New("سجل العهدة غير موجود")
	}

	oldAmount := expense.Amount
	oldRecipient := expense.RecipientName
	oldDate := day.Date

	if req.Category != nil && *req.Category != "" {
		expense.Category = *req.Category
	}
	if req.Amount != nil && *req.Amount >= 0 {
		expense.Amount = *req.Amount
	}
	if req.RecipientName != nil {
		expense.RecipientName = *req.RecipientName
	}
	if req.ImageURL != nil {
		expense.ImageURL = s.sanitizeImageURL(*req.ImageURL)
	}
	if req.ExportNumber != nil {
		expense.ExportNumber = *req.ExportNumber
	}

	dateChanged := req.Date != nil && *req.Date != "" && *req.Date != day.Date
	var targetDay *CustodyDay

	if dateChanged {
		newDate := *req.Date
		targetDay, err = s.repo.FindDayByDate(ctx, day.BranchID, newDate)
		if err != nil || targetDay == nil {
			var openingBalance float64
			prevDay, errPrev := s.repo.FindPreviousDay(ctx, day.BranchID, newDate)
			if errPrev == nil && prevDay != nil {
				s.recomputeClosing(prevDay)
				openingBalance = prevDay.ClosingBalance
			}
			targetDay = &CustodyDay{
				ID:             uuid.New(),
				BranchID:       day.BranchID,
				Date:           newDate,
				OpeningBalance: openingBalance,
				ClosingBalance: openingBalance,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if err := s.repo.CreateDay(ctx, targetDay); err != nil {
				return nil, err
			}
		}
		expense.CustodyDayID = targetDay.ID
	}

	if err := s.repo.UpdateExpense(ctx, expense); err != nil {
		return nil, err
	}

	desc := fmt.Sprintf("تعديل مصروف (%s): من %.2f إلى %.2f (مستلم: %s -> %s)", expense.Category, oldAmount, expense.Amount, oldRecipient, expense.RecipientName)
	if dateChanged {
		desc += fmt.Sprintf(" - نقل التاريخ من %s إلى %s", oldDate, *req.Date)
	}

	logEntry := &CustodyLog{
		BranchID:      day.BranchID,
		CustodyDayID:  expense.CustodyDayID,
		Date:          day.Date,
		ActionType:    "UPDATE_EXPENSE",
		Category:      expense.Category,
		Amount:        expense.Amount,
		Description:   desc,
		RecipientName: expense.RecipientName,
		ExportNumber:  expense.ExportNumber,
		ImageURL:      expense.ImageURL,
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	if dateChanged {
		logEntry.Date = *req.Date
	}
	_ = s.repo.CreateLog(ctx, logEntry)

	refreshedOldDay, err := s.repo.FindDayByID(ctx, day.ID)
	if err == nil && refreshedOldDay != nil {
		s.recomputeClosing(refreshedOldDay)
		_ = s.repo.UpdateDay(ctx, refreshedOldDay)
	}

	earliestDate := day.Date
	if dateChanged && targetDay != nil {
		refreshedTargetDay, err := s.repo.FindDayByID(ctx, targetDay.ID)
		if err == nil && refreshedTargetDay != nil {
			s.recomputeClosing(refreshedTargetDay)
			_ = s.repo.UpdateDay(ctx, refreshedTargetDay)
		}
		if *req.Date < earliestDate {
			earliestDate = *req.Date
		}
	}

	_, _ = s.cascadeRecalculateChain(ctx, day.BranchID, earliestDate)

	updatedReturnDay, err := s.repo.FindDayByID(ctx, expense.CustodyDayID)
	if err == nil && updatedReturnDay != nil {
		return s.toResponse(updatedReturnDay), nil
	}

	return s.toResponse(day), nil
}

func (s *custodyService) DeleteExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error) {
	expense, err := s.repo.FindExpenseByID(ctx, expenseID)
	if err != nil {
		return nil, errors.New("المصروف غير موجود")
	}

	day, err := s.repo.FindDayByID(ctx, expense.CustodyDayID)
	if err != nil {
		return nil, errors.New("سجل العهدة غير موجود")
	}

	if err := s.repo.DeleteExpense(ctx, expenseID); err != nil {
		return nil, err
	}

	logEntry := &CustodyLog{
		BranchID:      day.BranchID,
		CustodyDayID:  day.ID,
		Date:          day.Date,
		ActionType:    "DELETE_EXPENSE",
		Category:      expense.Category,
		Amount:        expense.Amount,
		Description:   fmt.Sprintf("حذف مصروف: %s - %.2f", expense.Category, expense.Amount),
		RecipientName: expense.RecipientName,
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateLog(ctx, logEntry)

	if updatedDay, _ := s.cascadeRecalculateChain(ctx, day.BranchID, day.Date); updatedDay != nil && updatedDay.ID == day.ID {
		day = updatedDay
	} else {
		day, err = s.repo.FindDayByID(ctx, expense.CustodyDayID)
		if err != nil {
			return nil, err
		}
		s.recomputeClosing(day)
		_ = s.repo.UpdateDay(ctx, day)
	}

	return s.toResponse(day), nil
}

func (s *custodyService) GetLogs(ctx context.Context, filter CustodyLogFilter) ([]CustodyLog, int64, error) {
	return s.repo.FindLogs(ctx, filter)
}

func (s *custodyService) DeleteLog(ctx context.Context, id uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) error {
	logEntry, err := s.repo.FindLogByID(ctx, id)
	if err != nil {
		return errors.New("سجل التدقيق غير موجود")
	}

	day, err := s.repo.FindDayByID(ctx, logEntry.CustodyDayID)
	if err == nil && day != nil {
		switch logEntry.ActionType {
		case "ADD_CUSTODY":
			day.AddedAmount -= logEntry.Amount
			if day.AddedAmount < 0 {
				day.AddedAmount = 0
			}
		}
		s.recomputeClosing(day)
		_ = s.repo.UpdateDay(ctx, day)
		_, _ = s.cascadeRecalculateChain(ctx, day.BranchID, day.Date)
	}

	return s.repo.DeleteLog(ctx, id)
}
