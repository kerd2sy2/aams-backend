package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/custody/contracts"
	"delivery-backend/internal/modules/custody/domain"
	"delivery-backend/internal/modules/custody/dto"
	"delivery-backend/internal/modules/custody/repository"
)

type CustodyService interface {
	List(ctx context.Context, branchID *uuid.UUID) ([]dto.CustodyDayResponse, error)
	Create(ctx context.Context, req dto.CreateCustodyDayRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error)
	AddAmount(ctx context.Context, req dto.AddCustodyAmountRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error)
	AddExpense(ctx context.Context, dayID uuid.UUID, branchID *uuid.UUID, req dto.CreateCustodyExpenseRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error)
	UpdateExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, req dto.UpdateCustodyExpenseRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error)
	DeleteExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error)
	GetLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error)
	DeleteLog(ctx context.Context, id uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) error
	SetLedgerImage(ctx context.Context, branchID *uuid.UUID, req dto.SetLedgerImageRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error)

	// Contract implementations
	GetDaySummary(ctx context.Context, branchID *uuid.UUID, date string) (*contracts.CustodySummaryDTO, error)
}

type custodyService struct {
	repo repository.CustodyRepository
}

func NewCustodyService(repo repository.CustodyRepository) CustodyService {
	return &custodyService{repo: repo}
}

func (s *custodyService) toResponse(day *domain.CustodyDay) *dto.CustodyDayResponse {
	resp := &dto.CustodyDayResponse{
		ID:             day.ID,
		BranchID:       day.BranchID,
		Date:           day.Date,
		OpeningBalance: day.OpeningBalance,
		AddedAmount:    day.AddedAmount,
		CustodyValue:   day.OpeningBalance + day.AddedAmount,
		ClosingBalance: day.ClosingBalance,
		LedgerImageURL: day.LedgerImageURL,
		Expenses:       make([]dto.CustodyExpenseResponse, 0, len(day.Expenses)),
	}

	var totalExp float64
	for _, e := range day.Expenses {
		totalExp += e.Amount
		switch e.Category {
		case "fuel":
			resp.Totals.Fuel += e.Amount
		case "internet_advance":
			resp.Totals.InternetAdvance += e.Amount
		case "advance":
			resp.Totals.Advance += e.Amount
		case "license":
			resp.Totals.License += e.Amount
		case "spare_parts":
			resp.Totals.SpareParts += e.Amount
		default:
			resp.Totals.Other += e.Amount
		}
		resp.Expenses = append(resp.Expenses, dto.CustodyExpenseResponse{
			ID:            e.ID,
			CustodyDayID:  e.CustodyDayID,
			Category:      e.Category,
			Amount:        e.Amount,
			RecipientName: e.RecipientName,
			ExportNumber:  e.ExportNumber,
			ImageURL:      e.ImageURL,
			CreatedAt:     e.CreatedAt,
		})
	}
	resp.TotalExpenses = totalExp
	resp.ClosingBalance = resp.CustodyValue - totalExp
	resp.CreatedAt = day.CreatedAt

	return resp
}

func (s *custodyService) sumExpenses(day *domain.CustodyDay) float64 {
	var total float64
	for _, e := range day.Expenses {
		total += e.Amount
	}
	return total
}

func (s *custodyService) recomputeClosing(day *domain.CustodyDay) {
	day.ClosingBalance = (day.OpeningBalance + day.AddedAmount) - s.sumExpenses(day)
}

func (s *custodyService) List(ctx context.Context, branchID *uuid.UUID) ([]dto.CustodyDayResponse, error) {
	days, err := s.repo.FindAll(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if len(days) == 0 {
		return []dto.CustodyDayResponse{}, nil
	}

	// 1. Group and recalculate chain in ascending date order
	// so opening balance of each day strictly matches previous day's closing balance
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date < days[j].Date
	})

	// When listing for a specific branch or single operational branch, maintain a single chronological timeline
	if branchID != nil {
		var prevClosing float64
		hasPrev := false

		for i := range days {
			day := &days[i]
			if !hasPrev {
				prev, errPrev := s.repo.FindPreviousDay(ctx, branchID, day.Date)
				if errPrev == nil && prev != nil {
					s.recomputeClosing(prev)
					if day.OpeningBalance != prev.ClosingBalance {
						day.OpeningBalance = prev.ClosingBalance
						_ = s.repo.UpdateDay(ctx, day)
					}
				}
				s.recomputeClosing(day)
				prevClosing = day.ClosingBalance
				hasPrev = true
			} else {
				needsUpdate := false
				if day.OpeningBalance != prevClosing {
					day.OpeningBalance = prevClosing
					needsUpdate = true
				}
				s.recomputeClosing(day)
				if needsUpdate {
					_ = s.repo.UpdateDay(ctx, day)
				}
				prevClosing = day.ClosingBalance
			}
		}
	} else {
		// Multi-branch fallback grouping
		branchLastClosing := make(map[string]float64)
		branchHasPrev := make(map[string]bool)

		// Determine dominant branch key for any records missing branch_id
		defaultBranchKey := "0b52c83c-4a37-4950-8040-fcb00fa718b3"
		for _, d := range days {
			if d.BranchID != nil {
				defaultBranchKey = d.BranchID.String()
				break
			}
		}

		for i := range days {
			day := &days[i]
			bKey := defaultBranchKey
			if day.BranchID != nil {
				bKey = day.BranchID.String()
			}

			if !branchHasPrev[bKey] {
				prev, errPrev := s.repo.FindPreviousDay(ctx, day.BranchID, day.Date)
				if errPrev == nil && prev != nil {
					s.recomputeClosing(prev)
					if day.OpeningBalance != prev.ClosingBalance {
						day.OpeningBalance = prev.ClosingBalance
						_ = s.repo.UpdateDay(ctx, day)
					}
				}
				s.recomputeClosing(day)
				branchLastClosing[bKey] = day.ClosingBalance
				branchHasPrev[bKey] = true
			} else {
				prevClosing := branchLastClosing[bKey]
				needsUpdate := false
				if day.OpeningBalance != prevClosing {
					day.OpeningBalance = prevClosing
					needsUpdate = true
				}
				s.recomputeClosing(day)
				if needsUpdate {
					_ = s.repo.UpdateDay(ctx, day)
				}
				branchLastClosing[bKey] = day.ClosingBalance
			}
		}
	}

	// 2. Sort newest first for client display
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date > days[j].Date
	})

	result := make([]dto.CustodyDayResponse, 0, len(days))
	for i := range days {
		result = append(result, *s.toResponse(&days[i]))
	}
	return result, nil
}

func (s *custodyService) cascadeRecalculateChain(ctx context.Context, branchID *uuid.UUID, fromDate string) error {
	days, err := s.repo.FindDaysFrom(ctx, branchID, fromDate)
	if err != nil {
		return err
	}
	if len(days) == 0 {
		return nil
	}

	for i := range days {
		day := &days[i]
		if i == 0 {
			// Find previous day strictly before day.Date to know opening balance
			prev, err := s.repo.FindPreviousDay(ctx, branchID, day.Date)
			if err == nil && prev != nil {
				s.recomputeClosing(prev)
				day.OpeningBalance = prev.ClosingBalance
			}
		} else {
			prevDay := &days[i-1]
			day.OpeningBalance = prevDay.ClosingBalance
		}
		s.recomputeClosing(day)
		_ = s.repo.UpdateDay(ctx, day)
	}
	return nil
}

func (s *custodyService) Create(ctx context.Context, req dto.CreateCustodyDayRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
	if req.BranchID == nil {
		primaryID := uuid.MustParse("0b52c83c-4a37-4950-8040-fcb00fa718b3")
		req.BranchID = &primaryID
	}
	existing, err := s.repo.FindDayByDate(ctx, req.BranchID, req.Date)
	if err == nil && existing != nil {
		return nil, errors.New("يوجد سجل عهدة مسجل بالفعل لهذا اليوم")
	}

	var openingBalance float64
	prevDay, err := s.repo.FindPreviousDay(ctx, req.BranchID, req.Date)
	if err == nil && prevDay != nil {
		s.recomputeClosing(prevDay)
		openingBalance = prevDay.ClosingBalance
	}

	day := &domain.CustodyDay{
		ID:             uuid.New(),
		BranchID:       req.BranchID,
		Date:           req.Date,
		OpeningBalance: openingBalance,
		AddedAmount:    req.AddedAmount,
		ClosingBalance: openingBalance + req.AddedAmount,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateDay(ctx, day); err != nil {
		return nil, err
	}

	if req.AddedAmount > 0 {
		logEntry := &domain.CustodyLog{
			BranchID:      day.BranchID,
			CustodyDayID:  day.ID,
			Date:          day.Date,
			ActionType:    "ADD_CUSTODY",
			Category:      "custody",
			Amount:        req.AddedAmount,
			Description:   fmt.Sprintf("إضافة عهدة افتتاحية بقيمة %.2f", req.AddedAmount),
			AdminID:       adminID,
			AdminName:     adminName,
			AdminUsername: adminUsername,
			CreatedAt:     time.Now(),
		}
		_ = s.repo.CreateLog(ctx, logEntry)
	}

	// Recalculate subsequent days if this day was inserted retroactively
	_ = s.cascadeRecalculateChain(ctx, req.BranchID, req.Date)

	updatedDay, err := s.repo.FindDayByID(ctx, day.ID)
	if err == nil && updatedDay != nil {
		day = updatedDay
	}

	return s.toResponse(day), nil
}

func (s *custodyService) AddAmount(ctx context.Context, req dto.AddCustodyAmountRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
	day, err := s.repo.FindDayByID(ctx, req.CustodyDayID)
	if err != nil {
		return nil, errors.New("سجل العهدة غير موجود")
	}

	day.AddedAmount += req.AddedAmount
	s.recomputeClosing(day)

	if err := s.repo.UpdateDay(ctx, day); err != nil {
		return nil, err
	}

	logEntry := &domain.CustodyLog{
		BranchID:      day.BranchID,
		CustodyDayID:  day.ID,
		Date:          day.Date,
		ActionType:    "ADD_CUSTODY",
		Category:      "custody",
		Amount:        req.AddedAmount,
		Description:   fmt.Sprintf("إضافة عهدة بقيمة %.2f", req.AddedAmount),
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateLog(ctx, logEntry)

	_ = s.cascadeRecalculateChain(ctx, day.BranchID, day.Date)

	updatedDay, err := s.repo.FindDayByID(ctx, day.ID)
	if err == nil && updatedDay != nil {
		day = updatedDay
	}

	return s.toResponse(day), nil
}

func (s *custodyService) AddExpense(ctx context.Context, dayID uuid.UUID, branchID *uuid.UUID, req dto.CreateCustodyExpenseRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
	day, err := s.repo.FindDayByID(ctx, dayID)
	if err != nil {
		return nil, errors.New("سجل العهدة غير موجود")
	}

	expense := &domain.CustodyExpense{
		ID:                uuid.New(),
		CustodyDayID:      dayID,
		Category:          req.Category,
		Amount:            req.Amount,
		RecipientName:     req.RecipientName,
		ExportNumber:      req.ExportNumber,
		ImageURL:          req.ImageURL,
		CreatedByID:       adminID,
		CreatedByName:     adminName,
		CreatedByUsername: adminUsername,
		CreatedAt:         time.Now(),
	}

	if err := s.repo.CreateExpense(ctx, expense); err != nil {
		return nil, err
	}

	day.Expenses = append(day.Expenses, *expense)
	s.recomputeClosing(day)
	_ = s.repo.UpdateDay(ctx, day)

	logEntry := &domain.CustodyLog{
		BranchID:      day.BranchID,
		CustodyDayID:  day.ID,
		Date:          day.Date,
		ActionType:    "ADD_EXPENSE",
		Category:      req.Category,
		Amount:        req.Amount,
		Description:   fmt.Sprintf("إضافة مصروف: %s - %.2f", req.Category, req.Amount),
		RecipientName: req.RecipientName,
		ExportNumber:  req.ExportNumber,
		ImageURL:      req.ImageURL,
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateLog(ctx, logEntry)

	_ = s.cascadeRecalculateChain(ctx, day.BranchID, day.Date)

	updatedDay, err := s.repo.FindDayByID(ctx, day.ID)
	if err == nil && updatedDay != nil {
		day = updatedDay
	}

	return s.toResponse(day), nil
}

func (s *custodyService) UpdateExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, req dto.UpdateCustodyExpenseRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
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
		expense.ImageURL = *req.ImageURL
	}
	if req.ExportNumber != nil {
		expense.ExportNumber = *req.ExportNumber
	}

	if err := s.repo.UpdateExpense(ctx, expense); err != nil {
		return nil, err
	}

	logEntry := &domain.CustodyLog{
		BranchID:      day.BranchID,
		CustodyDayID:  day.ID,
		Date:          day.Date,
		ActionType:    "UPDATE_EXPENSE",
		Category:      expense.Category,
		Amount:        expense.Amount,
		Description:   fmt.Sprintf("تعديل مصروف (%s): من %.2f إلى %.2f (مستلم: %s -> %s)", expense.Category, oldAmount, expense.Amount, oldRecipient, expense.RecipientName),
		RecipientName: expense.RecipientName,
		ExportNumber:  expense.ExportNumber,
		ImageURL:      expense.ImageURL,
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateLog(ctx, logEntry)

	// Fetch refreshed day with expenses
	day, err = s.repo.FindDayByID(ctx, expense.CustodyDayID)
	if err != nil {
		return nil, err
	}
	s.recomputeClosing(day)
	_ = s.repo.UpdateDay(ctx, day)

	_ = s.cascadeRecalculateChain(ctx, day.BranchID, day.Date)

	updatedDay, err := s.repo.FindDayByID(ctx, day.ID)
	if err == nil && updatedDay != nil {
		day = updatedDay
	}

	return s.toResponse(day), nil
}

func (s *custodyService) DeleteExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
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

	logEntry := &domain.CustodyLog{
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

	day, err = s.repo.FindDayByID(ctx, expense.CustodyDayID)
	if err != nil {
		return nil, err
	}
	s.recomputeClosing(day)
	_ = s.repo.UpdateDay(ctx, day)

	_ = s.cascadeRecalculateChain(ctx, day.BranchID, day.Date)

	updatedDay, err := s.repo.FindDayByID(ctx, day.ID)
	if err == nil && updatedDay != nil {
		day = updatedDay
	}

	return s.toResponse(day), nil
}

func (s *custodyService) GetLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error) {
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
		_ = s.cascadeRecalculateChain(ctx, day.BranchID, day.Date)
	}

	return s.repo.DeleteLog(ctx, id)
}

// Contract implementation
func (s *custodyService) GetDaySummary(ctx context.Context, branchID *uuid.UUID, date string) (*contracts.CustodySummaryDTO, error) {
	day, err := s.repo.FindDayByDate(ctx, branchID, date)
	if err != nil {
		return nil, err
	}
	s.recomputeClosing(day)
	var totalExpenses float64
	for _, e := range day.Expenses {
		totalExpenses += e.Amount
	}
	return &contracts.CustodySummaryDTO{
		DayID:          day.ID,
		BranchID:       day.BranchID,
		Date:           day.Date,
		OpeningBalance: day.OpeningBalance,
		AddedAmount:    day.AddedAmount,
		TotalExpenses:  totalExpenses,
		ClosingBalance: day.ClosingBalance,
	}, nil
}

func (s *custodyService) SetLedgerImage(ctx context.Context, branchID *uuid.UUID, req dto.SetLedgerImageRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
	day, err := s.repo.FindDayByDate(ctx, branchID, req.Date)
	if err != nil {
		day = &domain.CustodyDay{
			BranchID: branchID,
			Date:     req.Date,
		}
		if prev, errPrev := s.repo.FindPreviousDay(ctx, branchID, req.Date); errPrev == nil && prev != nil {
			day.OpeningBalance = prev.ClosingBalance
		}
		day.ClosingBalance = day.OpeningBalance
		day.LedgerImageURL = req.ImageURL
		if errCreate := s.repo.CreateDay(ctx, day); errCreate != nil {
			return nil, errCreate
		}
	} else {
		day.LedgerImageURL = req.ImageURL
		if errUpdate := s.repo.UpdateDay(ctx, day); errUpdate != nil {
			return nil, errUpdate
		}
	}

	updatedDay, err := s.repo.FindDayByID(ctx, day.ID)
	if err == nil && updatedDay != nil {
		day = updatedDay
	}
	s.recomputeClosing(day)
	return s.toResponse(day), nil
}
