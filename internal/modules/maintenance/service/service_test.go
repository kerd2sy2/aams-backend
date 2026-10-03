package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/maintenance/domain"
	"delivery-backend/internal/modules/maintenance/dto"
	"delivery-backend/internal/modules/maintenance/service"
)

type mockMaintenanceRepo struct {
	logs     []domain.MaintenanceLog
	requests map[uuid.UUID]*domain.MaintenanceRequest
}

func newMockMaintenanceRepo() *mockMaintenanceRepo {
	return &mockMaintenanceRepo{
		logs:     make([]domain.MaintenanceLog, 0),
		requests: make(map[uuid.UUID]*domain.MaintenanceRequest),
	}
}

func (m *mockMaintenanceRepo) CreateLog(ctx context.Context, log *domain.MaintenanceLog) error {
	m.logs = append(m.logs, *log)
	return nil
}

func (m *mockMaintenanceRepo) FindByEmployeeID(ctx context.Context, empID uuid.UUID, limit int) ([]domain.MaintenanceLog, error) {
	var result []domain.MaintenanceLog
	for _, l := range m.logs {
		if l.EmployeeID != nil && *l.EmployeeID == empID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockMaintenanceRepo) FindAllLogs(ctx context.Context, page, limit int) ([]domain.MaintenanceLog, int64, error) {
	return m.logs, int64(len(m.logs)), nil
}

func (m *mockMaintenanceRepo) FindLastOilChange(ctx context.Context, empID uuid.UUID) (*domain.MaintenanceLog, error) {
	for i := len(m.logs) - 1; i >= 0; i-- {
		l := m.logs[i]
		if l.EmployeeID != nil && *l.EmployeeID == empID && l.Type == "oil_change" {
			return &l, nil
		}
	}
	return nil, domain.ErrLogNotFound
}

func (m *mockMaintenanceRepo) CreateRequest(ctx context.Context, req *domain.MaintenanceRequest) error {
	m.requests[req.ID] = req
	return nil
}

func (m *mockMaintenanceRepo) UpdateRequest(ctx context.Context, req *domain.MaintenanceRequest) error {
	m.requests[req.ID] = req
	return nil
}

func (m *mockMaintenanceRepo) DeleteRequest(ctx context.Context, id uuid.UUID) error {
	delete(m.requests, id)
	return nil
}

func (m *mockMaintenanceRepo) FindRequestByID(ctx context.Context, id uuid.UUID) (*domain.MaintenanceRequest, error) {
	req, ok := m.requests[id]
	if !ok {
		return nil, domain.ErrRequestNotFound
	}
	return req, nil
}

func (m *mockMaintenanceRepo) FindAllRequests(ctx context.Context, filter dto.MaintenanceRequestFilter) ([]domain.MaintenanceRequest, int64, error) {
	var list []domain.MaintenanceRequest
	for _, r := range m.requests {
		list = append(list, *r)
	}
	return list, int64(len(list)), nil
}

func TestMaintenanceService_LogsAndRequests(t *testing.T) {
	repo := newMockMaintenanceRepo()
	svc := service.NewMaintenanceService(repo)
	ctx := context.Background()

	empID := uuid.New()

	// 1. Record Maintenance Log
	err := svc.RecordMaintenanceLog(ctx, &empID, "oil_change", "Oil changed 10W40", 5000, 150, "Admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2. Get Employee Logs
	logs, err := svc.GetEmployeeLogs(ctx, empID, 10)
	if err != nil || len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d, err: %v", len(logs), err)
	}

	// 3. Get Last Oil Change
	lastOil, err := svc.GetLastOilChange(ctx, empID)
	if err != nil || lastOil == nil {
		t.Fatalf("expected last oil change log, got err: %v", err)
	}
	if lastOil.DistanceAt != 5000 {
		t.Fatalf("expected distance 5000, got %f", lastOil.DistanceAt)
	}

	// 4. Create Maintenance Request
	req, err := svc.CreateRequest(ctx, dto.CreateMaintenanceRequestRequest{
		VehiclePlate:     "ABC-1234",
		EmployeeID:       &empID,
		IssueDescription: "Brakes squeaking",
		Priority:         "HIGH",
	}, nil)
	if err != nil || req == nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// 5. Update Maintenance Request
	status := "IN_PROGRESS"
	cost := 350.0
	updated, err := svc.UpdateRequest(ctx, req.ID, dto.UpdateMaintenanceRequestRequest{
		Status:     &status,
		ActualCost: &cost,
	})
	if err != nil || updated.Status != "IN_PROGRESS" || updated.ActualCost != 350.0 {
		t.Fatalf("failed to update request: %v", err)
	}

	// 6. Delete Maintenance Request
	if err := svc.DeleteRequest(ctx, req.ID); err != nil {
		t.Fatalf("failed to delete request: %v", err)
	}
	_ = time.Second
}
