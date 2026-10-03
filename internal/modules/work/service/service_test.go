package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/work/domain"
	"delivery-backend/internal/modules/work/dto"
	"delivery-backend/internal/modules/work/repository"
	"delivery-backend/internal/modules/work/service"
)

type mockWorkRepo struct {
	sessions map[uuid.UUID]*domain.WorkSession
}

func newMockWorkRepo() *mockWorkRepo {
	return &mockWorkRepo{sessions: make(map[uuid.UUID]*domain.WorkSession)}
}

func (m *mockWorkRepo) ExecInTx(ctx context.Context, fn func(txRepo repository.WorkRepository) error) error {
	return fn(m)
}

func (m *mockWorkRepo) Create(ctx context.Context, session *domain.WorkSession) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *mockWorkRepo) Update(ctx context.Context, session *domain.WorkSession) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *mockWorkRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.WorkSession, error) {
	if s, ok := m.sessions[id]; ok {
		return s, nil
	}
	return nil, nil
}

func (m *mockWorkRepo) FindActiveSessionByEmployeeID(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error) {
	for _, s := range m.sessions {
		if s.EmployeeID != nil && *s.EmployeeID == empID && s.Status == "ACTIVE" {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockWorkRepo) FindLastCompletedSession(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error) {
	for _, s := range m.sessions {
		if s.EmployeeID != nil && *s.EmployeeID == empID && s.Status == "COMPLETED" {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockWorkRepo) FindLastSessionByMotorcycle(ctx context.Context, motorcycleNumber string) (*domain.WorkSession, error) {
	for _, s := range m.sessions {
		if s.MotorcycleNumber == motorcycleNumber && s.Status == "COMPLETED" {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockWorkRepo) CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error) {
	var count int64
	for _, s := range m.sessions {
		if s.EmployeeID != nil && *s.EmployeeID == empID {
			count++
		}
	}
	return count, nil
}

func (m *mockWorkRepo) GetActiveSessions(ctx context.Context, branchID *uuid.UUID) ([]domain.WorkSession, error) {
	var list []domain.WorkSession
	for _, s := range m.sessions {
		if s.Status == "ACTIVE" {
			list = append(list, *s)
		}
	}
	return list, nil
}

type mockEmployeeProvider struct {
	emp *service.EmployeeData
}

func (m *mockEmployeeProvider) GetEmployee(ctx context.Context, id uuid.UUID) (*service.EmployeeData, error) {
	return m.emp, nil
}

func (m *mockEmployeeProvider) UpdateEmployeeOnStartWork(ctx context.Context, id uuid.UUID, appID, appType, motorcycleNumber string) error {
	return nil
}

func (m *mockEmployeeProvider) UpdateEmployeeOnEndWork(ctx context.Context, id uuid.UUID, addedDistance float64, totalOrders int) error {
	m.emp.TotalDistance += addedDistance
	return nil
}

func (m *mockEmployeeProvider) SendShiftApprovalNotification(ctx context.Context, empID uuid.UUID, sessionID uuid.UUID, ordersCount int, fuelCost float64) error {
	return nil
}

type mockVehicleProvider struct{}

func (v *mockVehicleProvider) GetVehicleLastKM(ctx context.Context, plateNumber string) (float64, error) {
	return 1500, nil
}

func (v *mockVehicleProvider) GetVehicleInfo(ctx context.Context, plateNumber string) (*service.VehicleData, error) {
	return &service.VehicleData{
		PlateNumber:     plateNumber,
		VehicleType:     "motorcycle",
		CurrentKM:       1500,
		LastOilChangeKM: 1500,
		TotalDistance:   1500,
	}, nil
}

func (v *mockVehicleProvider) UpdateVehicleKM(ctx context.Context, plateNumber string, km float64) error {
	return nil
}

func (v *mockVehicleProvider) HasActiveMaintenance(ctx context.Context, motorcycleNumber string) (bool, error) {
	return false, nil
}

func TestWorkService_StartAndEndWork(t *testing.T) {
	empID := uuid.New()
	empProvider := &mockEmployeeProvider{
		emp: &service.EmployeeData{
			ID:                    empID,
			Name:                  "Ahmed",
			NationalID:            "1234567890",
			VehicleType:           "motorcycle",
			MotorcycleNumber:      "123-ABC",
			TotalDistance:         500,
			LastOilChangeDistance: 0,
		},
	}
	repo := newMockWorkRepo()
	svc := service.NewWorkService(repo, empProvider, &mockVehicleProvider{}, nil)
	ctx := context.Background()

	// 1. Start Work
	startReq := dto.StartWorkRequest{
		EmployeeID:       empID.String(),
		StartKM:          500,
		MotorcycleNumber: "123-ABC",
	}
	sess, err := svc.StartWork(ctx, startReq)
	if err != nil {
		t.Fatalf("unexpected error starting work: %v", err)
	}
	if sess.Status != "ACTIVE" || sess.StartKM != 500 {
		t.Errorf("expected ACTIVE status and 500 startKM, got %v / %v", sess.Status, sess.StartKM)
	}

	// 2. Cannot start duplicate
	_, err = svc.StartWork(ctx, startReq)
	if err == nil {
		t.Fatalf("expected error starting duplicate active session, got nil")
	}

	// 3. End Work
	endReq := dto.EndWorkRequest{
		EmployeeID:  empID.String(),
		EndKM:       550,
		OrdersCount: 15,
		FuelCost:    20,
	}
	endedSess, err := svc.EndWork(ctx, endReq, nil, "", false)
	if err != nil {
		t.Fatalf("unexpected error ending work: %v", err)
	}
	if endedSess.Status != "COMPLETED" || endedSess.Distance != 50 || endedSess.OrdersCount != 15 {
		t.Errorf("unexpected end session values: status=%v dist=%v orders=%v", endedSess.Status, endedSess.Distance, endedSess.OrdersCount)
	}
}
