package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/ticket/contracts"
	"delivery-backend/internal/modules/ticket/domain"
	"delivery-backend/internal/modules/ticket/dto"
	"delivery-backend/internal/modules/ticket/service"
)

// MockRepository implements repository.SupportTicketRepository in memory for pure unit testing
type mockTicketRepo struct {
	tickets map[uuid.UUID]*domain.SupportTicket
}

func newMockTicketRepo() *mockTicketRepo {
	return &mockTicketRepo{tickets: make(map[uuid.UUID]*domain.SupportTicket)}
}

func (m *mockTicketRepo) Create(ctx context.Context, ticket *domain.SupportTicket) error {
	m.tickets[ticket.ID] = ticket
	return nil
}

func (m *mockTicketRepo) Update(ctx context.Context, ticket *domain.SupportTicket) error {
	m.tickets[ticket.ID] = ticket
	return nil
}

func (m *mockTicketRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.tickets, id)
	return nil
}

func (m *mockTicketRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.SupportTicket, error) {
	t, ok := m.tickets[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

func (m *mockTicketRepo) FindAll(ctx context.Context, filter dto.SupportTicketFilter) ([]domain.SupportTicket, int64, error) {
	var list []domain.SupportTicket
	for _, t := range m.tickets {
		if filter.BranchID != nil && t.BranchID != nil && *t.BranchID != *filter.BranchID {
			continue
		}
		list = append(list, *t)
	}
	return list, int64(len(list)), nil
}

// MockEmployeeContract for isolated testing
type mockEmpContract struct {
	summary *contracts.EmployeeSummaryDTO
}

func (m *mockEmpContract) GetEmployeeSummary(ctx context.Context, employeeID uuid.UUID) (*contracts.EmployeeSummaryDTO, error) {
	return m.summary, nil
}

// MockNotificationContract for isolated testing
type mockNotifContract struct {
	sentCount int
}

func (m *mockNotifContract) SendAsyncNotification(ctx context.Context, notif contracts.PushNotificationDTO) {
	m.sentCount++
}

func TestSupportTicketService_Create(t *testing.T) {
	repo := newMockTicketRepo()
	branchID := uuid.New()
	empID := uuid.New()

	empContract := &mockEmpContract{
		summary: &contracts.EmployeeSummaryDTO{
			ID:       empID,
			Name:     "Ahmed Ali",
			Phone:    "0500000000",
			BranchID: &branchID,
		},
	}
	notifContract := &mockNotifContract{}

	svc := service.NewSupportTicketService(repo, empContract, notifContract)

	req := dto.CreateSupportTicketRequest{
		EmployeeID:  &empID,
		Subject:     "طلب صيانة عاجل",
		Category:    "VEHICLE",
		Priority:    "HIGH",
		Description: "مشكلة في فرامل الدباب",
	}

	created, err := svc.Create(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.Subject != req.Subject {
		t.Errorf("expected subject %s, got %s", req.Subject, created.Subject)
	}

	if created.BranchID == nil || *created.BranchID != branchID {
		t.Errorf("expected branchID to be derived from employee contract: %v, got %v", branchID, created.BranchID)
	}

	if created.Status != "OPEN" {
		t.Errorf("expected status OPEN, got %s", created.Status)
	}
}

func TestSupportTicketService_Update(t *testing.T) {
	repo := newMockTicketRepo()
	svc := service.NewSupportTicketService(repo, nil, nil)

	ticketID := uuid.New()
	repo.tickets[ticketID] = &domain.SupportTicket{
		ID:           ticketID,
		TicketNumber: "TCK-123456",
		Subject:      "Subject 1",
		Status:       "OPEN",
	}

	newStatus := "RESOLVED"
	resolution := "تم حل المشكلة وتغيير القطعة"
	req := dto.UpdateSupportTicketRequest{
		Status:     &newStatus,
		Resolution: &resolution,
	}

	updated, err := svc.Update(context.Background(), ticketID, req)
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if updated.Status != "RESOLVED" {
		t.Errorf("expected status RESOLVED, got %s", updated.Status)
	}

	if updated.Resolution != resolution {
		t.Errorf("expected resolution %s, got %s", resolution, updated.Resolution)
	}
}
