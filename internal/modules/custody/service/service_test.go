package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/custody/domain"
	"delivery-backend/internal/modules/custody/dto"
	"delivery-backend/internal/modules/custody/service"
)

type mockCustodyRepo struct {
	days     map[uuid.UUID]*domain.CustodyDay
	expenses map[uuid.UUID]*domain.CustodyExpense
	logs     []domain.CustodyLog
}

func newMockCustodyRepo() *mockCustodyRepo {
	return &mockCustodyRepo{
		days:     make(map[uuid.UUID]*domain.CustodyDay),
		expenses: make(map[uuid.UUID]*domain.CustodyExpense),
		logs:     make([]domain.CustodyLog, 0),
	}
}

func (m *mockCustodyRepo) CreateDay(ctx context.Context, day *domain.CustodyDay) error {
	m.days[day.ID] = day
	return nil
}

func (m *mockCustodyRepo) UpdateDay(ctx context.Context, day *domain.CustodyDay) error {
	m.days[day.ID] = day
	return nil
}

func (m *mockCustodyRepo) FindDayByID(ctx context.Context, id uuid.UUID) (*domain.CustodyDay, error) {
	d, ok := m.days[id]
	if !ok {
		return nil, domain.ErrDayNotFound
	}
	return d, nil
}

func (m *mockCustodyRepo) FindDayByDate(ctx context.Context, branchID *uuid.UUID, date string) (*domain.CustodyDay, error) {
	for _, d := range m.days {
		if d.Date == date {
			return d, nil
		}
	}
	return nil, domain.ErrDayNotFound
}

func (m *mockCustodyRepo) FindLastDay(ctx context.Context, branchID *uuid.UUID) (*domain.CustodyDay, error) {
	for _, d := range m.days {
		return d, nil
	}
	return nil, domain.ErrDayNotFound
}

func (m *mockCustodyRepo) FindAll(ctx context.Context, branchID *uuid.UUID) ([]domain.CustodyDay, error) {
	var list []domain.CustodyDay
	for _, d := range m.days {
		list = append(list, *d)
	}
	return list, nil
}

func (m *mockCustodyRepo) CreateExpense(ctx context.Context, expense *domain.CustodyExpense) error {
	m.expenses[expense.ID] = expense
	return nil
}

func (m *mockCustodyRepo) DeleteExpense(ctx context.Context, id uuid.UUID) error {
	delete(m.expenses, id)
	return nil
}

func (m *mockCustodyRepo) FindExpenseByID(ctx context.Context, id uuid.UUID) (*domain.CustodyExpense, error) {
	e, ok := m.expenses[id]
	if !ok {
		return nil, domain.ErrExpenseNotFound
	}
	return e, nil
}

func (m *mockCustodyRepo) CreateLog(ctx context.Context, log *domain.CustodyLog) error {
	m.logs = append(m.logs, *log)
	return nil
}

func (m *mockCustodyRepo) FindLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error) {
	return m.logs, int64(len(m.logs)), nil
}

func (m *mockCustodyRepo) FindLogByID(ctx context.Context, id uuid.UUID) (*domain.CustodyLog, error) {
	for _, l := range m.logs {
		if l.ID == id {
			return &l, nil
		}
	}
	return nil, domain.ErrLogNotFound
}

func (m *mockCustodyRepo) DeleteLog(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestCustodyService_Lifecycle(t *testing.T) {
	repo := newMockCustodyRepo()
	svc := service.NewCustodyService(repo)
	ctx := context.Background()

	adminID := uuid.New()

	// 1. Create Custody Day
	day, err := svc.Create(ctx, dto.CreateCustodyDayRequest{
		Date:        "2026-10-03",
		AddedAmount: 500.0,
	}, &adminID, "Admin", "admin")
	if err != nil || day == nil {
		t.Fatalf("failed to create day: %v", err)
	}
	if day.CustodyValue != 500.0 {
		t.Fatalf("expected custody value 500, got %f", day.CustodyValue)
	}

	// 2. Add Amount
	day, err = svc.AddAmount(ctx, dto.AddCustodyAmountRequest{
		CustodyDayID: day.ID,
		AddedAmount:  200.0,
	}, &adminID, "Admin", "admin")
	if err != nil || day == nil {
		t.Fatalf("failed to add amount: %v", err)
	}
	if day.CustodyValue != 700.0 {
		t.Fatalf("expected custody value 700, got %f", day.CustodyValue)
	}

	// 3. Add Expense
	day, err = svc.AddExpense(ctx, day.ID, nil, dto.CreateCustodyExpenseRequest{
		Category:      "fuel",
		Amount:        100.0,
		RecipientName: "Driver 1",
	}, &adminID, "Admin", "admin")
	if err != nil || day == nil {
		t.Fatalf("failed to add expense: %v", err)
	}
	if day.TotalExpenses != 100.0 || day.ClosingBalance != 600.0 {
		t.Fatalf("expected closing balance 600, got %f", day.ClosingBalance)
	}

	// 4. Get Summary Contract
	summary, err := svc.GetDaySummary(ctx, nil, "2026-10-03")
	if err != nil || summary == nil {
		t.Fatalf("failed to get summary contract: %v", err)
	}
	if summary.ClosingBalance != 600.0 {
		t.Fatalf("expected contract summary 600, got %f", summary.ClosingBalance)
	}
}
