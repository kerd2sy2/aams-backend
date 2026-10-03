package service

import (
	"context"
	"errors"
	"fmt"
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
	DeleteExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error)
	GetLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error)
	DeleteLog(ctx context.Context, id uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) error

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
		Expenses:       make([]dto.CustodyExpenseResponse, 0, len(day.Expenses)),
	}

	var totalExp float64
	for _, e := range day.Expenses {
		totalExp += e.Amount
		switch e.Category {
		case "fuel":
			resp.Totals.Fuel += e.Amount
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
			CreatedAt:     e.CreatedAt,
		})
	}
	resp.TotalExpenses = totalExp
	resp.ClosingBalance = resp.CustodyValue - totalExp
	resp.CreatedAt = day.CreatedAt

	return resp
}

func (s *custodyService) recomputeClosing(day *domain.CustodyDay) {
	var totalExpenses float64
	for _, e := range day.Expenses {
		totalExpenses += e.Amount
	}
	day.ClosingBalance = (day.OpeningBalance + day.AddedAmount) - totalExpenses
}

func (s *custodyService) List(ctx context.Context, branchID *uuid.UUID) ([]dto.CustodyDayResponse, error) {
	days, err := s.repo.FindAll(ctx, branchID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.CustodyDayResponse, 0, len(days))
	for i := range days {
		result = append(result, *s.toResponse(&days[i]))
	}
	return result, nil
}

func (s *custodyService) Create(ctx context.Context, req dto.CreateCustodyDayRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
	existing, err := s.repo.FindDayByDate(ctx, req.BranchID, req.Date)
	if err == nil && existing != nil {
		return nil, errors.New("يوجد سجل عهدة مسجل بالفعل لهذا اليوم")
	}

	var openingBalance float64
	lastDay, err := s.repo.FindLastDay(ctx, req.BranchID)
	if err == nil && lastDay != nil {
		s.recomputeClosing(lastDay)
		openingBalance = lastDay.ClosingBalance
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
		AdminID:       adminID,
		AdminName:     adminName,
		AdminUsername: adminUsername,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateLog(ctx, logEntry)

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
