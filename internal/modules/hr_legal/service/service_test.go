package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/hr_legal/domain"
	"delivery-backend/internal/modules/hr_legal/dto"
	"delivery-backend/internal/modules/hr_legal/service"
)

type mockHRLegalRepo struct {
	investigations map[uuid.UUID]*domain.Investigation
	documents      map[uuid.UUID]*domain.EmployeeDocument
	bankAccounts   map[uuid.UUID]*domain.EmployeeBankAccount
	leaves         map[uuid.UUID]*domain.LeaveRequest
	violations     map[uuid.UUID]*domain.TrafficViolation
	fuelLogs       map[uuid.UUID]*domain.FuelLog
}

func newMockHRLegalRepo() *mockHRLegalRepo {
	return &mockHRLegalRepo{
		investigations: make(map[uuid.UUID]*domain.Investigation),
		documents:      make(map[uuid.UUID]*domain.EmployeeDocument),
		bankAccounts:   make(map[uuid.UUID]*domain.EmployeeBankAccount),
		leaves:         make(map[uuid.UUID]*domain.LeaveRequest),
		violations:     make(map[uuid.UUID]*domain.TrafficViolation),
		fuelLogs:       make(map[uuid.UUID]*domain.FuelLog),
	}
}

// 1. Investigation
func (m *mockHRLegalRepo) CreateInvestigation(ctx context.Context, inv *domain.Investigation) error {
	m.investigations[inv.ID] = inv
	return nil
}

func (m *mockHRLegalRepo) UpdateInvestigation(ctx context.Context, inv *domain.Investigation) error {
	m.investigations[inv.ID] = inv
	return nil
}

func (m *mockHRLegalRepo) FindInvestigations(ctx context.Context, branchID *uuid.UUID) ([]domain.Investigation, error) {
	var list []domain.Investigation
	for _, inv := range m.investigations {
		list = append(list, *inv)
	}
	return list, nil
}

func (m *mockHRLegalRepo) FindInvestigationByID(ctx context.Context, id uuid.UUID) (*domain.Investigation, error) {
	inv, ok := m.investigations[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return inv, nil
}

func (m *mockHRLegalRepo) CountPendingInvestigations(ctx context.Context) (int64, error) {
	var count int64
	for _, inv := range m.investigations {
		if inv.Status == "pending" {
			count++
		}
	}
	return count, nil
}

// 2. Document
func (m *mockHRLegalRepo) CreateDocument(ctx context.Context, doc *domain.EmployeeDocument) error {
	m.documents[doc.ID] = doc
	return nil
}

func (m *mockHRLegalRepo) UpdateDocument(ctx context.Context, doc *domain.EmployeeDocument) error {
	m.documents[doc.ID] = doc
	return nil
}

func (m *mockHRLegalRepo) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	delete(m.documents, id)
	return nil
}

func (m *mockHRLegalRepo) FindDocumentByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDocument, error) {
	doc, ok := m.documents[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return doc, nil
}

func (m *mockHRLegalRepo) FindDocuments(ctx context.Context, filter dto.EmployeeDocumentFilter) ([]domain.EmployeeDocument, int64, error) {
	var list []domain.EmployeeDocument
	for _, doc := range m.documents {
		list = append(list, *doc)
	}
	return list, int64(len(list)), nil
}

func (m *mockHRLegalRepo) FindExpiringDocuments(ctx context.Context, days int) ([]domain.EmployeeDocument, error) {
	return nil, nil
}

// 3. Bank Account
func (m *mockHRLegalRepo) CreateBankAccount(ctx context.Context, acc *domain.EmployeeBankAccount) error {
	m.bankAccounts[acc.ID] = acc
	return nil
}

func (m *mockHRLegalRepo) UpdateBankAccount(ctx context.Context, acc *domain.EmployeeBankAccount) error {
	m.bankAccounts[acc.ID] = acc
	return nil
}

func (m *mockHRLegalRepo) DeleteBankAccount(ctx context.Context, id uuid.UUID) error {
	delete(m.bankAccounts, id)
	return nil
}

func (m *mockHRLegalRepo) FindBankAccountByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeBankAccount, error) {
	acc, ok := m.bankAccounts[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return acc, nil
}

func (m *mockHRLegalRepo) FindBankAccounts(ctx context.Context, filter dto.EmployeeBankAccountFilter) ([]domain.EmployeeBankAccount, int64, error) {
	var list []domain.EmployeeBankAccount
	for _, acc := range m.bankAccounts {
		list = append(list, *acc)
	}
	return list, int64(len(list)), nil
}

func (m *mockHRLegalRepo) ClearDefaultBankAccount(ctx context.Context, employeeID uuid.UUID) error {
	for _, acc := range m.bankAccounts {
		if acc.EmployeeID == employeeID {
			acc.IsDefault = false
		}
	}
	return nil
}

// 4. Leave Request
func (m *mockHRLegalRepo) CreateLeaveRequest(ctx context.Context, req *domain.LeaveRequest) error {
	m.leaves[req.ID] = req
	return nil
}

func (m *mockHRLegalRepo) UpdateLeaveRequest(ctx context.Context, req *domain.LeaveRequest) error {
	m.leaves[req.ID] = req
	return nil
}

func (m *mockHRLegalRepo) DeleteLeaveRequest(ctx context.Context, id uuid.UUID) error {
	delete(m.leaves, id)
	return nil
}

func (m *mockHRLegalRepo) FindLeaveRequestByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error) {
	l, ok := m.leaves[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return l, nil
}

func (m *mockHRLegalRepo) FindLeaveRequests(ctx context.Context, filter dto.LeaveRequestFilter) ([]domain.LeaveRequest, int64, error) {
	var list []domain.LeaveRequest
	for _, l := range m.leaves {
		list = append(list, *l)
	}
	return list, int64(len(list)), nil
}

func (m *mockHRLegalRepo) CountPendingLeaveRequests(ctx context.Context) (int64, error) {
	var count int64
	for _, l := range m.leaves {
		if l.Status == "PENDING" {
			count++
		}
	}
	return count, nil
}

func (m *mockHRLegalRepo) FindActiveLeave(ctx context.Context, empID uuid.UUID, date string) (*domain.LeaveRequest, error) {
	for _, l := range m.leaves {
		if l.EmployeeID == empID && l.Status == "APPROVED" && l.StartDate <= date && l.EndDate >= date {
			return l, nil
		}
	}
	return nil, domain.ErrNotFound
}

// 5. Traffic Violation
func (m *mockHRLegalRepo) CreateViolation(ctx context.Context, v *domain.TrafficViolation) error {
	m.violations[v.ID] = v
	return nil
}

func (m *mockHRLegalRepo) UpdateViolation(ctx context.Context, v *domain.TrafficViolation) error {
	m.violations[v.ID] = v
	return nil
}

func (m *mockHRLegalRepo) DeleteViolation(ctx context.Context, id uuid.UUID) error {
	delete(m.violations, id)
	return nil
}

func (m *mockHRLegalRepo) FindViolationByID(ctx context.Context, id uuid.UUID) (*domain.TrafficViolation, error) {
	v, ok := m.violations[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return v, nil
}

func (m *mockHRLegalRepo) FindViolations(ctx context.Context, filter dto.TrafficViolationFilter) ([]domain.TrafficViolation, int64, error) {
	var list []domain.TrafficViolation
	for _, v := range m.violations {
		list = append(list, *v)
	}
	return list, int64(len(list)), nil
}

// 6. Fuel Log
func (m *mockHRLegalRepo) CreateFuelLog(ctx context.Context, log *domain.FuelLog) error {
	m.fuelLogs[log.ID] = log
	return nil
}

func (m *mockHRLegalRepo) UpdateFuelLog(ctx context.Context, log *domain.FuelLog) error {
	m.fuelLogs[log.ID] = log
	return nil
}

func (m *mockHRLegalRepo) DeleteFuelLog(ctx context.Context, id uuid.UUID) error {
	delete(m.fuelLogs, id)
	return nil
}

func (m *mockHRLegalRepo) FindFuelLogByID(ctx context.Context, id uuid.UUID) (*domain.FuelLog, error) {
	f, ok := m.fuelLogs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return f, nil
}

func (m *mockHRLegalRepo) FindFuelLogs(ctx context.Context, filter dto.FuelLogFilter) ([]domain.FuelLog, int64, error) {
	var list []domain.FuelLog
	for _, f := range m.fuelLogs {
		list = append(list, *f)
	}
	return list, int64(len(list)), nil
}

func TestHRLegalService_AllEntities(t *testing.T) {
	repo := newMockHRLegalRepo()
	svc := service.NewHRLegalService(repo)
	ctx := context.Background()

	empID := uuid.New()
	supID := uuid.New()

	// 1. Investigation Test
	invResp, err := svc.CreateInvestigation(ctx, dto.CreateInvestigationRequest{
		EmployeeID: empID.String(),
		Type:       "absence",
		ReportText: "Absent without notice",
	}, supID)
	if err != nil || invResp == nil {
		t.Fatalf("failed to create investigation: %v", err)
	}

	appResp, err := svc.ApproveInvestigation(ctx, invResp.ID, supID, "Supervisor", "sup")
	if err != nil || appResp.Status != "approved" {
		t.Fatalf("failed to approve investigation: %v", err)
	}

	// 2. Document Test
	doc, err := svc.CreateDocument(ctx, dto.CreateEmployeeDocumentRequest{
		EmployeeID: empID,
		DocType:    "DRIVING_LICENSE",
		Title:      "Driver License",
	})
	if err != nil || doc == nil {
		t.Fatalf("failed to create document: %v", err)
	}

	// 3. Bank Account Test
	acc, err := svc.CreateBankAccount(ctx, dto.CreateEmployeeBankAccountRequest{
		EmployeeID:       empID,
		BankName:         "Al Rajhi",
		IBAN:             "SA0380000000000000000000",
		AccountOwnerName: "Driver Name",
		IsDefault:        true,
	})
	if err != nil || acc == nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	// 4. Leave Request Test
	leave, err := svc.CreateLeaveRequest(ctx, dto.CreateLeaveRequestRequest{
		EmployeeID: empID,
		LeaveType:  "ANNUAL",
		StartDate:  "2026-10-01",
		EndDate:    "2026-10-10",
		DaysCount:  10,
	})
	if err != nil || leave == nil {
		t.Fatalf("failed to create leave request: %v", err)
	}

	appLeave, err := svc.UpdateLeaveRequestStatus(ctx, leave.ID, dto.UpdateLeaveRequestStatusRequest{
		Status:         "APPROVED",
		ApprovedByName: "HR Admin",
	})
	if err != nil || appLeave.Status != "APPROVED" {
		t.Fatalf("failed to approve leave: %v", err)
	}

	// 5. Contract: Check Employee On Leave
	checkDate, _ := time.Parse("2006-01-02", "2026-10-05")
	leaveCheck, err := svc.CheckEmployeeOnLeave(ctx, empID, checkDate)
	if err != nil || !leaveCheck.IsOnLeave {
		t.Fatalf("expected employee to be on leave, got: %v", leaveCheck)
	}

	// 6. Traffic Violation Test
	violation, err := svc.CreateViolation(ctx, dto.CreateTrafficViolationRequest{
		ViolationNumber: "VIO-123456",
		EmployeeID:      &empID,
		VehiclePlate:    "ABC-9999",
		Amount:          300.0,
		Reason:          "Speeding",
	})
	if err != nil || violation == nil {
		t.Fatalf("failed to create violation: %v", err)
	}

	// 7. Fuel Log Test
	fuel, err := svc.CreateFuelLog(ctx, dto.CreateFuelLogRequest{
		EmployeeID:   &empID,
		VehiclePlate: "ABC-9999",
		Amount:       50.0,
		Liters:       20.0,
	}, nil)
	if err != nil || fuel == nil {
		t.Fatalf("failed to create fuel log: %v", err)
	}
}
