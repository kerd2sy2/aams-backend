package custody

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StorageProvider abstracts base64 image persistence to disk/object storage.
type StorageProvider interface {
	SaveBase64Image(base64Str string, category string) (string, error)
}

type CustodyService interface {
	ICustodyContract
	List(ctx context.Context, branchID *uuid.UUID) ([]CustodyDayResponse, error)
	Create(ctx context.Context, req CreateCustodyDayRequest, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error)
	AddAmount(ctx context.Context, req AddCustodyAmountRequest, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error)
	AddExpense(ctx context.Context, dayID uuid.UUID, branchID *uuid.UUID, req CreateCustodyExpenseRequest, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error)
	UpdateExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, req UpdateCustodyExpenseRequest, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error)
	DeleteExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error)
	GetLogs(ctx context.Context, filter CustodyLogFilter) ([]CustodyLog, int64, error)
	DeleteLog(ctx context.Context, id uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) error
	SetLedgerImage(ctx context.Context, branchID *uuid.UUID, req SetLedgerImageRequest, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error)
}

type custodyService struct {
	repo    CustodyRepository
	storage StorageProvider
}

func NewCustodyService(repo CustodyRepository) CustodyService {
	return &custodyService{repo: repo}
}

func NewCustodyServiceWithStorage(repo CustodyRepository, storage StorageProvider) CustodyService {
	return &custodyService{repo: repo, storage: storage}
}

// sanitizeImageURL converts raw data:image/... base64 blobs into stored /uploads/custody/... URLs.
func (s *custodyService) sanitizeImageURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || s.storage == nil {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "data:image") {
		if savedURL, err := s.storage.SaveBase64Image(trimmed, "custody"); err == nil && savedURL != "" {
			return savedURL
		}
	}
	return trimmed
}

func (s *custodyService) toResponse(day *CustodyDay) *CustodyDayResponse {
	resp := &CustodyDayResponse{
		ID:             day.ID,
		BranchID:       day.BranchID,
		Date:           day.Date,
		OpeningBalance: day.OpeningBalance,
		AddedAmount:    day.AddedAmount,
		CustodyValue:   day.OpeningBalance + day.AddedAmount,
		ClosingBalance: day.ClosingBalance,
		LedgerImageURL: day.LedgerImageURL,
		Expenses:       make([]CustodyExpenseResponse, 0, len(day.Expenses)),
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
		resp.Expenses = append(resp.Expenses, CustodyExpenseResponse{
			ID:                e.ID,
			CustodyDayID:      e.CustodyDayID,
			Category:          e.Category,
			Amount:            e.Amount,
			RecipientName:     e.RecipientName,
			ExportNumber:      e.ExportNumber,
			ImageURL:          e.ImageURL,
			CreatedByName:     e.CreatedByName,
			CreatedByUsername: e.CreatedByUsername,
			CreatedAt:         e.CreatedAt,
		})
	}
	resp.TotalExpenses = totalExp
	resp.ClosingBalance = resp.CustodyValue - totalExp
	resp.CreatedAt = day.CreatedAt

	return resp
}

func (s *custodyService) sumExpenses(day *CustodyDay) float64 {
	var total float64
	for _, e := range day.Expenses {
		total += e.Amount
	}
	return total
}

func (s *custodyService) recomputeClosing(day *CustodyDay) {
	day.ClosingBalance = (day.OpeningBalance + day.AddedAmount) - s.sumExpenses(day)
}

func (s *custodyService) cascadeRecalculateChain(ctx context.Context, branchID *uuid.UUID, fromDate string) (*CustodyDay, error) {
	days, err := s.repo.FindDaysFrom(ctx, branchID, fromDate)
	if err != nil {
		return nil, err
	}
	if len(days) == 0 {
		return nil, nil
	}

	var prevClosing float64
	hasPrev := false
	if prev, errPrev := s.repo.FindPreviousDay(ctx, branchID, days[0].Date); errPrev == nil && prev != nil {
		s.recomputeClosing(prev)
		prevClosing = prev.ClosingBalance
		hasPrev = true
	}

	for i := range days {
		day := &days[i]
		oldOpening := day.OpeningBalance
		oldClosing := day.ClosingBalance
		if hasPrev {
			day.OpeningBalance = prevClosing
		}
		s.recomputeClosing(day)
		if day.OpeningBalance != oldOpening || day.ClosingBalance != oldClosing {
			_ = s.repo.UpdateDay(ctx, day)
		}
		prevClosing = day.ClosingBalance
		hasPrev = true
	}
	return &days[0], nil
}

// List returns custody days in descending date order as a pure read-only O(N) in-memory pass.
func (s *custodyService) List(ctx context.Context, branchID *uuid.UUID) ([]CustodyDayResponse, error) {
	days, err := s.repo.FindAll(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if len(days) == 0 {
		return []CustodyDayResponse{}, nil
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Date < days[j].Date
	})

	branchLastClosing := make(map[string]float64)
	branchHasPrev := make(map[string]bool)

	for i := range days {
		day := &days[i]
		bKey := "default"
		if day.BranchID != nil {
			bKey = day.BranchID.String()
		}
		if branchHasPrev[bKey] {
			day.OpeningBalance = branchLastClosing[bKey]
		}
		s.recomputeClosing(day)
		branchLastClosing[bKey] = day.ClosingBalance
		branchHasPrev[bKey] = true
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Date > days[j].Date
	})

	result := make([]CustodyDayResponse, 0, len(days))
	for i := range days {
		result = append(result, *s.toResponse(&days[i]))
	}
	return result, nil
}

func (s *custodyService) Create(ctx context.Context, req CreateCustodyDayRequest, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error) {
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

	day := &CustodyDay{
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
		logEntry := &CustodyLog{
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

	if updatedDay, _ := s.cascadeRecalculateChain(ctx, req.BranchID, req.Date); updatedDay != nil && updatedDay.ID == day.ID {
		day = updatedDay
	}

	return s.toResponse(day), nil
}

func (s *custodyService) AddAmount(ctx context.Context, req AddCustodyAmountRequest, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error) {
	day, err := s.repo.FindDayByID(ctx, req.CustodyDayID)
	if err != nil {
		return nil, errors.New("سجل العهدة غير موجود")
	}

	day.AddedAmount += req.AddedAmount
	s.recomputeClosing(day)

	if err := s.repo.UpdateDay(ctx, day); err != nil {
		return nil, err
	}

	logEntry := &CustodyLog{
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

	if updatedDay, _ := s.cascadeRecalculateChain(ctx, day.BranchID, day.Date); updatedDay != nil && updatedDay.ID == day.ID {
		day = updatedDay
	}

	return s.toResponse(day), nil
}

func (s *custodyService) SetLedgerImage(ctx context.Context, branchID *uuid.UUID, req SetLedgerImageRequest, adminID *uuid.UUID, adminName, adminUsername string) (*CustodyDayResponse, error) {
	cleanImg := s.sanitizeImageURL(req.ImageURL)
	day, err := s.repo.FindDayByDate(ctx, branchID, req.Date)
	if err != nil {
		day = &CustodyDay{
			BranchID: branchID,
			Date:     req.Date,
		}
		if prev, errPrev := s.repo.FindPreviousDay(ctx, branchID, req.Date); errPrev == nil && prev != nil {
			s.recomputeClosing(prev)
			day.OpeningBalance = prev.ClosingBalance
		}
		day.ClosingBalance = day.OpeningBalance
		day.LedgerImageURL = cleanImg
		if errCreate := s.repo.CreateDay(ctx, day); errCreate != nil {
			return nil, errCreate
		}
	} else {
		day.LedgerImageURL = cleanImg
		if errUpdate := s.repo.UpdateDay(ctx, day); errUpdate != nil {
			return nil, errUpdate
		}
	}

	s.recomputeClosing(day)
	return s.toResponse(day), nil
}

func (s *custodyService) GetDaySummary(ctx context.Context, branchID *uuid.UUID, date string) (*CustodySummaryDTO, error) {
	day, err := s.repo.FindDayByDate(ctx, branchID, date)
	if err != nil {
		return nil, err
	}
	s.recomputeClosing(day)
	return &CustodySummaryDTO{
		DayID:          day.ID,
		BranchID:       day.BranchID,
		Date:           day.Date,
		OpeningBalance: day.OpeningBalance,
		AddedAmount:    day.AddedAmount,
		TotalExpenses:  s.sumExpenses(day),
		ClosingBalance: day.ClosingBalance,
	}, nil
}
