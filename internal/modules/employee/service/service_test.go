package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/employee/domain"
	"delivery-backend/internal/modules/employee/dto"
	"delivery-backend/internal/modules/employee/repository"
	"delivery-backend/internal/modules/employee/service"
)

type mockEmployeeRepo struct {
	repository.EmployeeRepository
	employees map[uuid.UUID]*domain.Employee
}

func newMockEmployeeRepo() *mockEmployeeRepo {
	return &mockEmployeeRepo{
		employees: make(map[uuid.UUID]*domain.Employee),
	}
}

func (m *mockEmployeeRepo) Create(ctx context.Context, emp *domain.Employee) error {
	m.employees[emp.ID] = emp
	return nil
}

func (m *mockEmployeeRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	if e, ok := m.employees[id]; ok {
		return e, nil
	}
	return nil, nil
}

func (m *mockEmployeeRepo) FindByNationalID(ctx context.Context, nationalID string) (*domain.Employee, error) {
	for _, e := range m.employees {
		if e.NationalID == nationalID {
			return e, nil
		}
	}
	return nil, nil
}

func (m *mockEmployeeRepo) Update(ctx context.Context, emp *domain.Employee) error {
	m.employees[emp.ID] = emp
	return nil
}

func (m *mockEmployeeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.employees, id)
	return nil
}

func (m *mockEmployeeRepo) FindAll(ctx context.Context, filter dto.EmployeeFilter) ([]domain.Employee, int64, error) {
	var list []domain.Employee
	for _, e := range m.employees {
		list = append(list, *e)
	}
	return list, int64(len(list)), nil
}

func TestEmployeeService_CreateAndFind(t *testing.T) {
	repo := newMockEmployeeRepo()
	svc := service.NewEmployeeService(repo)
	ctx := context.Background()

	// 1. Create
	req := dto.CreateEmployeeRequest{
		Name:        "Mohamed Ali",
		NationalID:  "1234567890",
		Phone:       "0501234567",
		VehicleType: "motorcycle",
	}
	emp, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error creating employee: %v", err)
	}
	if emp.Name != "Mohamed Ali" || emp.NationalID != "1234567890" {
		t.Errorf("unexpected employee fields: %+v", emp)
	}

	// 2. Cannot duplicate national ID
	_, err = svc.Create(ctx, req)
	if err == nil {
		t.Fatalf("expected error on duplicate nationalID, got nil")
	}

	// 3. Find By ID
	found, err := svc.FindByID(ctx, emp.ID)
	if err != nil || found == nil {
		t.Fatalf("expected employee to be found, got error: %v", err)
	}

	// 4. Contract FindByNationalID
	contractDTO, err := svc.FindByNationalID(ctx, "1234567890")
	if err != nil || contractDTO == nil {
		t.Fatalf("expected contract DTO, got: %v", err)
	}
	if contractDTO.Name != "Mohamed Ali" {
		t.Errorf("contract DTO name mismatch: %s", contractDTO.Name)
	}
}
