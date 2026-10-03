package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/attendance/domain"
	"delivery-backend/internal/modules/attendance/service"
)

type mockAttendanceRepo struct {
	records map[string]*domain.Attendance
}

func newMockAttendanceRepo() *mockAttendanceRepo {
	return &mockAttendanceRepo{records: make(map[string]*domain.Attendance)}
}

func (m *mockAttendanceRepo) Upsert(ctx context.Context, attendance *domain.Attendance) error {
	key := attendance.EmployeeID.String() + ":" + attendance.Date
	m.records[key] = attendance
	return nil
}

func (m *mockAttendanceRepo) FindByDate(ctx context.Context, date string) ([]domain.Attendance, error) {
	var list []domain.Attendance
	for _, rec := range m.records {
		if rec.Date == date {
			list = append(list, *rec)
		}
	}
	return list, nil
}

func (m *mockAttendanceRepo) FindByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date string) (*domain.Attendance, error) {
	key := employeeID.String() + ":" + date
	if rec, ok := m.records[key]; ok {
		return rec, nil
	}
	return nil, nil
}

func (m *mockAttendanceRepo) DeleteByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date string) error {
	key := employeeID.String() + ":" + date
	delete(m.records, key)
	return nil
}

type mockEmployeeFinder struct {
	employees []service.EmployeeInfo
}

func (m *mockEmployeeFinder) FindAllEmployees(ctx context.Context, branchID *uuid.UUID) ([]service.EmployeeInfo, error) {
	return m.employees, nil
}

func (m *mockEmployeeFinder) FindEmployeeByID(ctx context.Context, id uuid.UUID) (*service.EmployeeInfo, error) {
	for _, e := range m.employees {
		if e.ID == id {
			return &e, nil
		}
	}
	return nil, nil
}

func TestAttendanceService_ToggleAndGet(t *testing.T) {
	repo := newMockAttendanceRepo()
	empID := uuid.New()
	finder := &mockEmployeeFinder{
		employees: []service.EmployeeInfo{
			{
				ID:          empID,
				Name:        "Test Employee",
				NationalID:  "1234567890",
				BranchName:  "Main Branch",
				VehicleType: "motorcycle",
			},
		},
	}

	svc := service.NewAttendanceService(repo, finder)
	ctx := context.Background()
	date := "2026-10-03"

	// Initial Get should be absent
	list, err := svc.GetAttendance(ctx, date, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 || list[0].Status != "absent" {
		t.Errorf("expected absent status, got %v", list[0].Status)
	}

	// Toggle to present
	toggled, err := svc.ToggleAttendance(ctx, uuid.Nil, empID, date, "present", "on time")
	if err != nil {
		t.Fatalf("unexpected error on toggle: %v", err)
	}
	if toggled.Status != "present" || toggled.Note != "on time" {
		t.Errorf("expected present status, got %v", toggled.Status)
	}

	// Contract check
	status, err := svc.GetEmployeeAttendanceStatus(ctx, empID, date)
	if err != nil {
		t.Fatalf("unexpected error getting contract status: %v", err)
	}
	if status.Status != "present" {
		t.Errorf("contract expected present, got %v", status.Status)
	}
}
