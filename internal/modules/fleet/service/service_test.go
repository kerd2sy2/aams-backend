package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/fleet/domain"
	"delivery-backend/internal/modules/fleet/dto"
	"delivery-backend/internal/modules/fleet/service"
)

type mockVehicleRepo struct {
	vehicles map[uuid.UUID]*domain.Vehicle
	plates   map[string]*domain.Vehicle
}

func newMockVehicleRepo() *mockVehicleRepo {
	return &mockVehicleRepo{
		vehicles: make(map[uuid.UUID]*domain.Vehicle),
		plates:   make(map[string]*domain.Vehicle),
	}
}

func (m *mockVehicleRepo) Create(ctx context.Context, v *domain.Vehicle) error {
	m.vehicles[v.ID] = v
	m.plates[v.PlateNumber] = v
	return nil
}

func (m *mockVehicleRepo) Update(ctx context.Context, v *domain.Vehicle) error {
	m.vehicles[v.ID] = v
	m.plates[v.PlateNumber] = v
	return nil
}

func (m *mockVehicleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if v, ok := m.vehicles[id]; ok {
		delete(m.plates, v.PlateNumber)
		delete(m.vehicles, id)
	}
	return nil
}

func (m *mockVehicleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Vehicle, error) {
	v, ok := m.vehicles[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return v, nil
}

func (m *mockVehicleRepo) FindByPlateNumber(ctx context.Context, plate string) (*domain.Vehicle, error) {
	v, ok := m.plates[plate]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return v, nil
}

func (m *mockVehicleRepo) FindByPlateNumberUnscoped(ctx context.Context, plate string) (*domain.Vehicle, error) {
	return m.FindByPlateNumber(ctx, plate)
}

func (m *mockVehicleRepo) RestoreVehicle(ctx context.Context, v *domain.Vehicle) error {
	return m.Update(ctx, v)
}

func (m *mockVehicleRepo) FindAll(ctx context.Context, filter dto.VehicleFilter) ([]domain.Vehicle, int64, error) {
	var list []domain.Vehicle
	for _, v := range m.vehicles {
		list = append(list, *v)
	}
	return list, int64(len(list)), nil
}

func (m *mockVehicleRepo) CountAll(ctx context.Context, branchID *uuid.UUID) (int64, error) {
	return int64(len(m.vehicles)), nil
}

func (m *mockVehicleRepo) UpdateOdometer(ctx context.Context, plate string, newKM float64, distance float64) error {
	v, ok := m.plates[plate]
	if ok {
		v.CurrentKM = newKM
		v.TotalDistance += distance
	}
	return nil
}

func (m *mockVehicleRepo) RecordOilChange(ctx context.Context, id uuid.UUID, currentKM float64) error {
	v, ok := m.vehicles[id]
	if ok {
		v.LastOilChangeKM = currentKM
	}
	return nil
}

func (m *mockVehicleRepo) FindLatestVehicleKM(ctx context.Context, plate string) (float64, error) {
	v, ok := m.plates[plate]
	if ok {
		return v.CurrentKM, nil
	}
	return 0, nil
}

func TestVehicleService_CreateAndGet(t *testing.T) {
	repo := newMockVehicleRepo()
	svc := service.NewVehicleService(repo, nil)

	req := dto.CreateVehicleRequest{
		PlateNumber:     "7572-ABC",
		VehicleType:     "motorcycle",
		Brand:           "Honda",
		CurrentKM:       1200,
		LastOilChangeKM: 1000,
	}

	created, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created.PlateNumber != "7572-ABC" {
		t.Errorf("expected plate 7572-ABC, got %s", created.PlateNumber)
	}

	resp, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected get by ID to succeed, got %v", err)
	}

	// 1200 - 1000 = 200 km driven. Interval for motorcycle is 950 km.
	// NeedsOilChange should be false, Remaining should be 750.
	if resp.NeedsOilChange {
		t.Errorf("expected NeedsOilChange to be false for 200 km driven")
	}
	if resp.RemainingOilKM != 750 {
		t.Errorf("expected RemainingOilKM to be 750, got %f", resp.RemainingOilKM)
	}
}

func TestVehicleService_OilChangeEligibility(t *testing.T) {
	repo := newMockVehicleRepo()
	svc := service.NewVehicleService(repo, nil)

	req := dto.CreateVehicleRequest{
		PlateNumber:     "9999-XYZ",
		VehicleType:     "motorcycle",
		CurrentKM:       2000,
		LastOilChangeKM: 1000, // Driven 1000 km >= 950 km threshold
	}

	created, _ := svc.Create(context.Background(), req)
	resp, _ := svc.GetByID(context.Background(), created.ID)

	if !resp.NeedsOilChange {
		t.Errorf("expected NeedsOilChange to be true when driven 1000 km >= 950 km interval")
	}
}
